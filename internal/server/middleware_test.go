package server_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/grpc-health-proxy/internal/server"
)

func newLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewTextHandler(buf, nil))
}

func TestLoggingMiddleware(t *testing.T) {
	var buf bytes.Buffer
	logger := newLogger(&buf)

	handler := server.Logging(logger, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(buf.String(), "/healthz") {
		t.Errorf("expected log to contain path, got: %s", buf.String())
	}
}

func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := newLogger(&buf)

	handler := server.Recovery(logger, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRecoveryMiddleware_Panic(t *testing.T) {
	var buf bytes.Buffer
	logger := newLogger(&buf)

	handler := server.Recovery(logger, http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("something went wrong")
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
	if !strings.Contains(buf.String(), "panic recovered") {
		t.Errorf("expected log to contain panic recovered, got: %s", buf.String())
	}
}
