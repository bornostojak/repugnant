package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// Server owns HTTP routing for the rePugnant backend.
type Server struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) *Server {
	return &Server{logger: logger}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	return s.logRequests(mux)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Debug("http request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
