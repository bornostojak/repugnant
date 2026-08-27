package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
	"strings"
	"time"
)

type Store struct {
	db       *sql.DB
	postgres bool
}
type Project struct{ Slug, Name, APIKey string }
type Article struct {
	ID, ShortID, ProjectSlug, Title, Body string
	Revision                              int
	CreatedAt                             time.Time
}

func Open(driver, dsn string) (*Store, error) {
	if driver == "" {
		driver = "sqlite"
	}
	if dsn == "" {
		dsn = "file:rpg.db?_pragma=busy_timeout(5000)"
	}
	name := driver
	pg := driver == "postgres"
	if pg {
		name = "pgx"
	}
	db, e := sql.Open(name, dsn)
	if e != nil {
		return nil, e
	}
	s := &Store{db, pg}
	return s, s.Migrate()
}
func (s *Store) Close() error { return s.db.Close() }
func (s *Store) q(query string) string {
	if !s.postgres {
		return query
	}
	n := 0
	var b strings.Builder
	for _, r := range query {
		if r == '?' {
			n++
			fmt.Fprintf(&b, "$%d", n)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
func (s *Store) Migrate() error {
	schema := `CREATE TABLE IF NOT EXISTS projects (slug TEXT PRIMARY KEY, name TEXT NOT NULL, api_key TEXT NOT NULL UNIQUE, created_at TIMESTAMP NOT NULL); CREATE TABLE IF NOT EXISTS articles (id TEXT PRIMARY KEY, short_id TEXT NOT NULL UNIQUE, project_slug TEXT NOT NULL, title TEXT NOT NULL, body TEXT NOT NULL, revision INTEGER NOT NULL, created_at TIMESTAMP NOT NULL);`
	if s.postgres {
		schema = `CREATE TABLE IF NOT EXISTS projects (slug TEXT PRIMARY KEY, name TEXT NOT NULL, api_key TEXT NOT NULL UNIQUE, created_at TIMESTAMPTZ NOT NULL); CREATE TABLE IF NOT EXISTS articles (id TEXT PRIMARY KEY, short_id TEXT NOT NULL UNIQUE, project_slug TEXT NOT NULL, title TEXT NOT NULL, body TEXT NOT NULL, revision INTEGER NOT NULL, created_at TIMESTAMPTZ NOT NULL);`
	}
	_, e := s.db.Exec(schema)
	return e
}
func token(n int) (string, error) {
	b := make([]byte, n)
	if _, e := rand.Read(b); e != nil {
		return "", e
	}
	return strings.TrimRight(base64.RawURLEncoding.EncodeToString(b), "=")[:n], nil
}
func (s *Store) CreateProject(name, slug string) (Project, error) {
	key, e := token(32)
	if e != nil {
		return Project{}, e
	}
	_, e = s.db.Exec(s.q(`INSERT INTO projects(slug,name,api_key,created_at) VALUES(?,?,?,?)`), slug, name, key, time.Now().UTC())
	return Project{slug, name, key}, e
}
func (s *Store) Authorize(slug, key string) (bool, error) {
	var n int
	e := s.db.QueryRow(s.q(`SELECT count(*) FROM projects WHERE slug=? AND api_key=?`), slug, key).Scan(&n)
	return n == 1, e
}
func (s *Store) AddArticle(a Article) (Article, error) {
	if a.ID == "" {
		a.ID, _ = token(12)
	}
	if a.ShortID == "" {
		a.ShortID, _ = token(9)
	}
	a.Revision = 1
	a.CreatedAt = time.Now().UTC()
	_, e := s.db.Exec(s.q(`INSERT INTO articles(id,short_id,project_slug,title,body,revision,created_at) VALUES(?,?,?,?,?,?,?)`), a.ID, a.ShortID, a.ProjectSlug, a.Title, a.Body, a.Revision, a.CreatedAt)
	return a, e
}
func (s *Store) FindShort(short string) (Article, error) {
	var a Article
	e := s.db.QueryRow(s.q(`SELECT id,short_id,project_slug,title,body,revision,created_at FROM articles WHERE short_id=?`), short).Scan(&a.ID, &a.ShortID, &a.ProjectSlug, &a.Title, &a.Body, &a.Revision, &a.CreatedAt)
	return a, e
}
func (s *Store) FindArticle(projectSlug, id string) (Article, error) {
	var a Article
	e := s.db.QueryRow(s.q(`SELECT id,short_id,project_slug,title,body,revision,created_at FROM articles WHERE project_slug=? AND id=?`), projectSlug, id).Scan(&a.ID, &a.ShortID, &a.ProjectSlug, &a.Title, &a.Body, &a.Revision, &a.CreatedAt)
	return a, e
}
func Slug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	dash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
		} else if !dash {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
func RequireSlug(s string) error {
	if Slug(s) != s || s == "" {
		return fmt.Errorf("invalid project slug")
	}
	return nil
}
