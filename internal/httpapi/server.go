package httpapi

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bornostojak/repugnant/internal/store"
)

// Server owns HTTP routing for the rePugnant backend.
type Server struct {
	logger *slog.Logger
	store  *store.Store
}

func New(logger *slog.Logger) *Server {
	return &Server{logger: logger}
}
func NewWithStore(logger *slog.Logger, database *store.Store) *Server {
	return &Server{logger: logger, store: database}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	if s.store != nil {
		mux.HandleFunc("POST /api/projects", s.createProject)
		mux.HandleFunc("GET /api/projects", s.listProjects)
		mux.HandleFunc("GET /api/projects/{slug}/articles", s.listArticles)
		mux.HandleFunc("GET /api/projects/{slug}/articles/{id}", s.getArticle)
		mux.HandleFunc("GET /api/projects/{slug}/articles/{id}/revisions", s.listRevisions)
		mux.HandleFunc("POST /api/projects/{slug}/articles", s.createArticle)
		mux.HandleFunc("GET /p/{slug}/article/{id}/{revision}", s.articlePage)
		mux.HandleFunc("GET /{shortID}", s.shortLink)
	}
	return s.logRequests(mux)
}
func (s *Server) listRevisions(w http.ResponseWriter, r *http.Request) {
	a, e := s.store.Revisions(r.PathValue("slug"), r.PathValue("id"))
	if e != nil {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, 200, a)
}
func (s *Server) getArticle(w http.ResponseWriter, r *http.Request) {
	a, e := s.store.FindArticle(r.PathValue("slug"), r.PathValue("id"))
	if e != nil {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, 200, a)
}
func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	p, e := s.store.ListProjects()
	if e != nil {
		http.Error(w, "list projects failed", 500)
		return
	}
	writeJSON(w, 200, p)
}
func (s *Server) listArticles(w http.ResponseWriter, r *http.Request) {
	a, e := s.store.ListArticles(r.PathValue("slug"), r.URL.Query().Get("q"))
	if e != nil {
		http.Error(w, "list articles failed", 500)
		return
	}
	writeJSON(w, 200, a)
}
func WithWeb(api http.Handler, dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, dir+"/index.html")
			return
		}
		candidate := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if _, e := os.Stat(candidate); e == nil {
			files.ServeHTTP(w, r)
			return
		}
		if !strings.HasPrefix(r.URL.Path, "/api/") && !strings.HasPrefix(r.URL.Path, "/p/") && r.URL.Path != "/healthz" {
			http.ServeFile(w, r, filepath.Join(dir, "index.html"))
			return
		}
		api.ServeHTTP(w, r)
	})
}

var articleTemplate = template.Must(template.New("article").Parse(`<!doctype html><html><head><meta name="viewport" content="width=device-width, initial-scale=1"><title>{{.Title}} · rePugnant</title><style>body{max-width:48rem;margin:3rem auto;padding:0 1rem;font:16px system-ui;color:#1d2835}header{border-bottom:1px solid #d8e0e8}pre{background:#f4f7f8;padding:1rem;overflow:auto}small{color:#52606d}</style></head><body><header><small>{{.ProjectSlug}} · revision {{.Revision}}</small><h1>{{.Title}}</h1><small>Short link: /{{.ShortID}}</small></header><main><pre>{{.Body}}</pre></main></body></html>`))

func (s *Server) articlePage(w http.ResponseWriter, r *http.Request) {
	a, e := s.store.FindArticle(r.PathValue("slug"), r.PathValue("id"))
	if e != nil {
		http.NotFound(w, r)
		return
	}
	if requested, err := strconv.Atoi(r.PathValue("revision")); err == nil && requested != a.Revision {
		for _, revision := range mustRevisions(s.store, a.ProjectSlug, a.ID) {
			if revision.Revision == requested {
				a = revision
				break
			}
		}
		if requested != a.Revision {
			http.NotFound(w, r)
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = articleTemplate.Execute(w, a)
}
func mustRevisions(database *store.Store, slug, id string) []store.Article {
	values, _ := database.Revisions(slug, id)
	return values
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.Name == "" {
		http.Error(w, "name is required", 400)
		return
	}
	if in.Slug == "" {
		in.Slug = store.Slug(in.Name)
	}
	if e := store.RequireSlug(in.Slug); e != nil {
		http.Error(w, e.Error(), 400)
		return
	}
	p, e := s.store.CreateProject(in.Name, in.Slug)
	if e != nil {
		http.Error(w, "project could not be created", 409)
		return
	}
	writeJSON(w, 201, map[string]string{"name": p.Name, "slug": p.Slug, "api_key": p.APIKey, "api_url": "/api/projects/" + p.Slug + "/articles"})
}
func (s *Server) createArticle(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	ok, e := s.store.Authorize(slug, r.Header.Get("X-RPG-API-Key"))
	if e != nil || !ok {
		http.Error(w, "invalid project API key", 401)
		return
	}
	var in struct {
		ID, Title, Body, Category, SourcePath string
		Tags                                  []string `json:"tags"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.Title == "" {
		http.Error(w, "title is required", 400)
		return
	}
	tags, _ := json.Marshal(in.Tags)
	a, e := s.store.AddArticle(store.Article{ID: in.ID, ProjectSlug: slug, Title: in.Title, Body: in.Body, Category: in.Category, Tags: string(tags), SourcePath: in.SourcePath})
	if e != nil {
		http.Error(w, "article could not be created", 409)
		return
	}
	writeJSON(w, 201, map[string]any{"id": a.ID, "short_id": a.ShortID, "revision": a.Revision, "url": "/p/" + slug + "/article/" + a.ID + "/" + strconv.Itoa(a.Revision), "short_url": "/" + a.ShortID})
}
func (s *Server) shortLink(w http.ResponseWriter, r *http.Request) {
	a, e := s.store.FindShort(r.PathValue("shortID"))
	if e != nil {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/p/"+a.ProjectSlug+"/article/"+a.ID+"/"+strconv.Itoa(a.Revision), http.StatusFound)
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
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
