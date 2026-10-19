package http

import (
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewServerDefaultHandlers(t *testing.T) {
	srv, err := NewServer()
	if err != nil {
		t.Fatal(err)
	}
	srv.Route("/").GET("/known", func(ctx Context) error {
		return ctx.Result(nethttp.StatusOK, map[string]string{"message": "ok"})
	})
	for _, tc := range []struct {
		method string
		path   string
		status int
	}{
		{nethttp.MethodGet, "/missing", nethttp.StatusNotFound},
		{nethttp.MethodPost, "/known", nethttp.StatusMethodNotAllowed},
		{nethttp.MethodGet, "/known", nethttp.StatusOK},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
			if w.Code != tc.status {
				t.Fatalf("status = %d, want %d; body = %s", w.Code, tc.status, w.Body.String())
			}
			if w.Header().Get("Content-Type") != "application/json" || !json.Valid(w.Body.Bytes()) {
				t.Fatalf("expected JSON response, got headers %v, body %s", w.Header(), w.Body.String())
			}
		})
	}
}

func TestNewServerAccept(t *testing.T) {
	srv, err := NewServer(Accept("application/json"))
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(nethttp.MethodGet, "/missing", nil)
	r.Header.Set("Accept", "text/plain")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	if w.Code != nethttp.StatusNotAcceptable || !json.Valid(w.Body.Bytes()) {
		t.Fatalf("expected JSON 406, got %d: %s", w.Code, w.Body.String())
	}
}

func TestNewServerInvalidAcceptFails(t *testing.T) {
	for _, types := range [][]string{
		{"invalid"},
		{"*/*"},
		{"application/x-unknown"},
		{"application/json", "invalid"},
	} {
		t.Run(strings.Join(types, ","), func(t *testing.T) {
			srv, err := NewServer(Accept(types...))
			if srv != nil || err == nil {
				t.Fatalf("expected construction to fail, got server %v, error %v", srv, err)
			}
			if !strings.Contains(err.Error(), "http: invalid response media types") || !strings.Contains(err.Error(), types[len(types)-1]) {
				t.Fatalf("expected error to identify invalid response media types, got %v", err)
			}
		})
	}
}

func TestNewServerRequestBodyLimit(t *testing.T) {
	srv, err := NewServer(MaxBodyBytes(8))
	if err != nil {
		t.Fatal(err)
	}
	srv.Route("/").POST("/body", func(ctx Context) error {
		var body map[string]any
		if err := ctx.Bind(&body); err != nil {
			return err
		}
		return ctx.Result(nethttp.StatusOK, body)
	})
	for _, tc := range []struct {
		body   string
		status int
	}{
		{`{}`, nethttp.StatusOK},
		{`{"message":"too large"}`, nethttp.StatusRequestEntityTooLarge},
	} {
		r := httptest.NewRequest(nethttp.MethodPost, "/body", strings.NewReader(tc.body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, r)
		if w.Code != tc.status {
			t.Fatalf("status = %d, want %d; body = %s", w.Code, tc.status, w.Body.String())
		}
	}
}
