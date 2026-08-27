package httpapi

import (
	"encoding/json"
	"github.com/bornostojak/repugnant/internal/store"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
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

	// SPEC.md requires /d/{article-id} (latest) and /d/{article-id}/{revision}
	// (historical) permalinks, with no project slug in the URL. Every
	// generated local doc embeds exactly this link as its "Web" column.
	doc := httptest.NewRecorder()
	server.ServeHTTP(doc, httptest.NewRequest(http.MethodGet, "/d/"+a["id"].(string), nil))
	if doc.Code != http.StatusOK || !strings.Contains(doc.Body.String(), "Hello") {
		t.Fatalf("/d/{id} = %d %s", doc.Code, doc.Body.String())
	}
	revision := httptest.NewRecorder()
	server.ServeHTTP(revision, httptest.NewRequest(http.MethodGet, "/d/"+a["id"].(string)+"/1", nil))
	if revision.Code != http.StatusOK || !strings.Contains(revision.Body.String(), "Hello") {
		t.Fatalf("/d/{id}/{revision} = %d %s", revision.Code, revision.Body.String())
	}
	missing := httptest.NewRecorder()
	server.ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/d/"+a["id"].(string)+"/99", nil))
	if missing.Code != http.StatusNotFound {
		t.Fatalf("/d/{id}/{bad revision} = %d", missing.Code)
	}
	unknown := httptest.NewRecorder()
	server.ServeHTTP(unknown, httptest.NewRequest(http.MethodGet, "/d/does-not-exist", nil))
	if unknown.Code != http.StatusNotFound {
		t.Fatalf("/d/{unknown id} = %d", unknown.Code)
	}

	// The web UI's TypeScript types read lowercase/snake_case JSON keys
	// (project.slug, article.short_id, ...). Project/Article previously had
	// no json tags at all, so Go's default encoder emitted PascalCase field
	// names instead and every list/tree/reader view silently rendered blank.
	projects := httptest.NewRecorder()
	server.ServeHTTP(projects, httptest.NewRequest(http.MethodGet, "/api/projects", nil))
	if !strings.Contains(projects.Body.String(), `"slug":"test-project"`) {
		t.Fatalf("project JSON is missing the lowercase \"slug\" key the frontend reads: %s", projects.Body.String())
	}
	if strings.Contains(projects.Body.String(), "api_key") || strings.Contains(projects.Body.String(), "APIKey") {
		t.Fatalf("project listing must never expose the API key: %s", projects.Body.String())
	}
	list := httptest.NewRecorder()
	server.ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/api/projects/test-project/articles", nil))
	if !strings.Contains(list.Body.String(), `"title":"Hello"`) || !strings.Contains(list.Body.String(), `"short_id"`) {
		t.Fatalf("article JSON is missing lowercase keys the frontend reads: %s", list.Body.String())
	}
}

// A built web/dist is not available in this test, but WithWeb must still let
// /d/ requests fall through to the API handler instead of swallowing them
// into the SPA's index.html fallback the way /p/ intentionally does.
func TestWithWebPassesDocPermalinksToAPI(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/index.html", []byte("<html>spa</html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	reached := false
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached = true })
	handler := WithWeb(api, dir)
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/d/some-id", nil))
	if !reached {
		t.Fatal("/d/ request did not reach the API handler")
	}
}
