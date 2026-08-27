package httpapi

import (
	"encoding/json"
	"github.com/bornostojak/repugnant/internal/store"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestProjectArticleAndShortRedirect(t *testing.T) {
	db, err := store.Open("sqlite", "file:httpapi-test?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	server := NewWithStore(slog.New(slog.NewTextHandler(io.Discard, nil)), db).Handler()
	project := httptest.NewRecorder()
	server.ServeHTTP(project, httptest.NewRequest(http.MethodPost, "/api/projects", strings.NewReader(`{"name":"Test Project"}`)))
	if project.Code != http.StatusCreated {
		t.Fatalf("project %d %s", project.Code, project.Body.String())
	}
	var p map[string]string
	_ = json.Unmarshal(project.Body.Bytes(), &p)
	article := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/projects/test-project/articles", strings.NewReader(`{"title":"Hello","body":"world"}`))
	req.Header.Set("X-RPG-API-Key", p["api_key"])
	server.ServeHTTP(article, req)
	if article.Code != http.StatusCreated {
		t.Fatalf("article %d %s", article.Code, article.Body.String())
	}
	var a map[string]any
	_ = json.Unmarshal(article.Body.Bytes(), &a)
	redirect := httptest.NewRecorder()
	server.ServeHTTP(redirect, httptest.NewRequest(http.MethodGet, "/"+a["short_id"].(string), nil))
	if redirect.Code != http.StatusFound {
		t.Fatalf("redirect %d", redirect.Code)
	}
	if !strings.HasSuffix(redirect.Header().Get("Location"), "/1") { t.Fatalf("short redirect location = %q", redirect.Header().Get("Location")) }
	patch := httptest.NewRecorder()
	server.ServeHTTP(patch, httptest.NewRequest(http.MethodPatch, "/api/projects/test-project/articles/"+a["id"].(string), strings.NewReader(`{"category":"Platform/Caching","tags":["cache"]}`)))
	if patch.Code != http.StatusOK { t.Fatalf("organization patch %d %s", patch.Code, patch.Body.String()) }
}
