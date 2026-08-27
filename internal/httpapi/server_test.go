package httpapi

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	server := New(slog.New(slog.NewTextHandler(io.Discard, nil)))
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	if recorder.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("body = %q", recorder.Body.String())
	}
}
