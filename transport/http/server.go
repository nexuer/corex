package http

import (
	"bytes"
	"fmt"
	"io"
	nethttp "net/http"

	kratoshttp "github.com/go-kratos/kratos/v3/transport/http"
	"github.com/nexuer/errors"
	"github.com/nexuer/errors/reasons"

	"github.com/nexuer/ghttp"
	"github.com/nexuer/ghttp/encoding"
	"github.com/nexuer/log"
)

type serverRuntime struct {
	logger       log.Printer
	negotiator   *ghttp.AcceptNegotiator
	maxBodyBytes int64
}

func newServerRuntime(c options) (*serverRuntime, error) {
	logger := c.logger
	if logger == nil {
		logger = log.NewPrinter(log.Default())
	}

	var negotiator *ghttp.AcceptNegotiator
	if len(c.accept) > 0 {
		var err error
		negotiator, err = ghttp.NewAcceptNegotiator(c.accept...)
		if err != nil {
			return nil, fmt.Errorf("http: invalid response media types %q: %w", c.accept, err)
		}
	}

	return &serverRuntime{
		logger:       logger,
		negotiator:   negotiator,
		maxBodyBytes: c.maxBodyBytes,
	}, nil
}

func (s *serverRuntime) writeResponse(w nethttp.ResponseWriter, codec encoding.Codec, code int, resp any) {
	data, err := codec.Marshal(resp)
	if err != nil {
		s.logger.Errorf("http: codec marshal failed: %v", err)
		nethttp.Error(w, nethttp.StatusText(nethttp.StatusInternalServerError), nethttp.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	if _, err = w.Write(data); err != nil {
		s.logger.Errorf("http: write response failed: %v", err)
	}
}

func (s *serverRuntime) setResponseHeaders(w nethttp.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

func (s *serverRuntime) setContentType(w nethttp.ResponseWriter, r *nethttp.Request) (encoding.Codec, bool) {
	codec := encoding.GetCodec("json")
	contentType := "application/json"

	if s.negotiator == nil {
		s.setResponseHeaders(w, contentType)
		return codec, true
	}

	negotiatedCodec, negotiatedType, ok := s.negotiator.Negotiate(r)
	if ok {
		s.setResponseHeaders(w, negotiatedType)
		return negotiatedCodec, true
	}
	s.setResponseHeaders(w, contentType)
	resp := reasons.ErrorNotAcceptable("406 Not Acceptable")
	s.writeResponse(w, codec, int(resp.Code), resp)
	return codec, false
}

func (s *serverRuntime) notFoundHandler() nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		codec, ok := s.setContentType(w, r)
		if !ok {
			return
		}
		resp := reasons.ErrorNotFound("404 Not Found")
		s.writeResponse(w, codec, int(resp.Code), resp)
	})
}

func (s *serverRuntime) methodNotAllowedHandler() nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		codec, ok := s.setContentType(w, r)
		if !ok {
			return
		}
		resp := reasons.ErrorMethodNotAllowed("405 Method Not Allowed")
		s.writeResponse(w, codec, int(resp.Code), resp)
	})
}

func (s *serverRuntime) errorEncoder(w kratoshttp.ResponseWriter, r *kratoshttp.Request, err error) {
	codec, ok := s.setContentType(w, r)
	if !ok {
		return
	}

	se := errors.FromError(err)
	if se.Message == "" {
		se.Message = fmt.Sprintf("%d %s", se.Code, nethttp.StatusText(int(se.Code)))
	}
	s.writeResponse(w, codec, int(se.Code), se)
}

func (s *serverRuntime) responseEncoder(w kratoshttp.ResponseWriter, r *kratoshttp.Request, v any) error {
	if v == nil {
		return nil
	}
	if rd, ok := v.(kratoshttp.Redirector); ok {
		url, code := rd.Redirect()
		nethttp.Redirect(w, r, url, code)
		return nil
	}

	codec, ok := s.setContentType(w, r)
	if !ok {
		return nil
	}
	data, err := codec.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func (s *serverRuntime) requestDecoder(r *kratoshttp.Request, v any) error {
	codec, ok := ghttp.CodecForRequest(r)
	if !ok {
		return reasons.ErrorUnsupportedMediaType(
			"Unsupported Content-Type Type: %s",
			r.Header.Get("Content-Type"))
	}

	data, err := ghttp.ReadBody(r.Body, s.maxBodyBytes)
	if err != nil {
		if ghttp.IsBodyTooLarge(err) {
			return reasons.ErrorRequestEntityTooLarge("%s", err.Error())
		}
		return reasons.ErrorBadRequest("%s", err.Error())
	}

	r.Body = io.NopCloser(bytes.NewReader(data))

	if len(data) == 0 {
		return nil
	}

	if err = codec.Unmarshal(data, v); err != nil {
		return reasons.ErrorCodec("request body: %s", err.Error())
	}
	return nil
}

// NewServer creates an HTTP server with the corex codecs and error handlers.
// Invalid response media types cause construction to fail.
func NewServer(opts ...ServerOption) (*Server, error) {
	var c options
	for _, opt := range opts {
		opt(&c)
	}
	runtime, err := newServerRuntime(c)
	if err != nil {
		return nil, err
	}
	serverOpts := []kratoshttp.ServerOption{
		kratoshttp.NotFoundHandler(runtime.notFoundHandler()),
		kratoshttp.MethodNotAllowedHandler(runtime.methodNotAllowedHandler()),
		kratoshttp.ErrorEncoder(runtime.errorEncoder),
		kratoshttp.RequestDecoder(runtime.requestDecoder),
		kratoshttp.ResponseEncoder(runtime.responseEncoder),
		kratoshttp.Filter(c.filters...),
	}
	serverOpts = append(serverOpts, c.serverOptions...)
	return kratoshttp.NewServer(serverOpts...), nil
}
