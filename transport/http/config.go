package http

import (
	"bytes"
	"fmt"
	"io"
	nethttp "net/http"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/nexuer/corex/errors"
	"github.com/nexuer/corex/log"
	"github.com/nexuer/ghttp"
	"github.com/nexuer/ghttp/encoding"
)

const (
	NotAcceptableReason    = "NOT_ACCEPTABLE"
	MethodNotAllowedReason = "METHOD_NOT_ALLOWED"
	NotFoundReason         = "NOT_FOUND"
	CODECReason            = "CODEC"
)

type Config struct {
	Logger  *log.Logger
	Address string
	Timeout time.Duration
}

func (c *Config) log() *log.Logger {
	if c.Logger == nil {
		return log.GetLogger()
	}
	return c.Logger
}

func (c *Config) Build(m ...middleware.Middleware) []ServerOption {
	return []ServerOption{
		http.Address(c.Address),
		http.Timeout(c.Timeout),
		http.NotFoundHandler(c.NotFoundHandler()),
		http.MethodNotAllowedHandler(c.MethodNotAllowedHandler()),
		http.ErrorEncoder(c.ErrorEncoder),
		http.RequestDecoder(c.RequestDecoder),
		http.ResponseEncoder(c.ResponseEncoder),
		http.Middleware(m...),
	}
}

func (c *Config) NotFoundHandler() nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		codec, isAccept := c.setContentType(w, r)
		if !isAccept {
			return
		}
		resp := errors.NotFound(NotFoundReason, "404 Not Found")
		c.writeResponse(w, codec, nethttp.StatusNotFound, resp)
	})
}

func (c *Config) MethodNotAllowedHandler() nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		codec, isAccept := c.setContentType(w, r)
		if !isAccept {
			return
		}
		resp := errors.MethodNotAllowed(MethodNotAllowedReason, "405 Method Not Allowed")
		c.writeResponse(w, codec, nethttp.StatusMethodNotAllowed, resp)
	})
}

func (c *Config) ErrorEncoder(w http.ResponseWriter, r *http.Request, err error) {
	codec, isAccept := c.setContentType(w, r)
	if !isAccept {
		return
	}

	se := errors.FromError(err)

	if se.Message == "" {
		se.Message = fmt.Sprintf("%d %s", se.Code, nethttp.StatusText(int(se.Code)))
	}

	c.writeResponse(w, codec, int(se.Code), se)
}

func (c *Config) RequestDecoder(r *http.Request, v any) error {
	codec, _ := ghttp.CodecForRequest(r)
	if codec == nil {
		return errors.BadRequest(CODECReason, fmt.Sprintf("unsupported Content-Type: %s",
			r.Header.Get("Content-Type")))
	}

	data, err := io.ReadAll(r.Body)

	if err != nil {
		return errors.BadRequest("CODEC", err.Error())
	}

	// reset body.
	r.Body = io.NopCloser(bytes.NewBuffer(data))

	if len(data) == 0 {
		return nil
	}
	if err = codec.Unmarshal(data, v); err != nil {
		return errors.BadRequest("CODEC", fmt.Sprintf("request body unmarshal: %s", err.Error()))
	}
	return nil
}

func (c *Config) ResponseEncoder(w http.ResponseWriter, r *http.Request, v any) error {
	if v == nil {
		return nil
	}

	if rd, ok := v.(http.Redirector); ok {
		url, code := rd.Redirect()
		nethttp.Redirect(w, r, url, code)
		return nil
	}

	codec, isAccept := c.setContentType(w, r)
	if !isAccept {
		return nil
	}

	data, err := codec.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) setContentType(w http.ResponseWriter, r *http.Request) (encoding.Codec, bool) {
	accept := r.Header.Get("Accept")
	codec, ok := ghttp.CodecForRequest(r, "Accept")
	if len(accept) > 0 && !ok {
		resp := errors.NotAcceptable(NotAcceptableReason, "406 Not Acceptable")
		c.writeResponse(w, codec, nethttp.StatusNotAcceptable, resp)
		return codec, false
	}
	w.Header().Set("Content-Type", accept)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	return codec, true
}

func (c *Config) writeResponse(w http.ResponseWriter, codec encoding.Codec, code int, resp any) {
	w.WriteHeader(code)
	b, err := codec.Marshal(resp)
	if err != nil {
		w.WriteHeader(nethttp.StatusInternalServerError)
		c.log().Error(err, "http: write response failed")
		return
	}

	_, err = w.Write(b)
	if err != nil {
		w.WriteHeader(nethttp.StatusInternalServerError)
		c.log().Error(err, "http: write response failed")
	}
}
