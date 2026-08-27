package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
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
	ID, ShortID, ProjectSlug, Title, Body, Category, Tags, SourcePath, SourceRange string
	Revision                                                                       int
	CreatedAt                                                                      time.Time
}

func (s *Store) ListProjects() ([]Project, error) {
	rows, e := s.db.Query(`SELECT slug,name FROM projects ORDER BY name`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []Project
	for rows.Next() {
		var p Project
		if e = rows.Scan(&p.Slug, &p.Name); e != nil {
			return nil, e
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
func (s *Store) ListArticles(slug, query string) ([]Article, error) {
	q := `SELECT id,short_id,project_slug,title,body,category,tags,source_path,source_range,revision,created_at FROM articles WHERE project_slug=?`
	args := []any{slug}
	if query != "" {
		q += ` AND (title LIKE ? OR body LIKE ? OR category LIKE ? OR tags LIKE ?)`
		args = append(args, "%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%")
	}
	q += ` ORDER BY created_at DESC`
	rows, e := s.db.Query(s.q(q), args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []Article
	for rows.Next() {
		var a Article
		if e = rows.Scan(&a.ID, &a.ShortID, &a.ProjectSlug, &a.Title, &a.Body, &a.Category, &a.Tags, &a.SourcePath, &a.SourceRange, &a.Revision, &a.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, a)
	}
	return out, rows.Err()
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
	schema := `CREATE TABLE IF NOT EXISTS projects (slug TEXT PRIMARY KEY, name TEXT NOT NULL, api_key TEXT NOT NULL UNIQUE, created_at TIMESTAMP NOT NULL); CREATE TABLE IF NOT EXISTS articles (id TEXT PRIMARY KEY, short_id TEXT NOT NULL UNIQUE, project_slug TEXT NOT NULL, title TEXT NOT NULL, body TEXT NOT NULL, category TEXT NOT NULL DEFAULT '', tags TEXT NOT NULL DEFAULT '', source_path TEXT NOT NULL DEFAULT '', source_range TEXT NOT NULL DEFAULT '', revision INTEGER NOT NULL, created_at TIMESTAMP NOT NULL); CREATE TABLE IF NOT EXISTS article_revisions (article_id TEXT NOT NULL, revision INTEGER NOT NULL, title TEXT NOT NULL, body TEXT NOT NULL, created_at TIMESTAMP NOT NULL, PRIMARY KEY(article_id,revision));`
	if s.postgres {
		schema = `CREATE TABLE IF NOT EXISTS projects (slug TEXT PRIMARY KEY, name TEXT NOT NULL, api_key TEXT NOT NULL UNIQUE, created_at TIMESTAMPTZ NOT NULL); CREATE TABLE IF NOT EXISTS articles (id TEXT PRIMARY KEY, short_id TEXT NOT NULL UNIQUE, project_slug TEXT NOT NULL, title TEXT NOT NULL, body TEXT NOT NULL, category TEXT NOT NULL DEFAULT '', tags TEXT NOT NULL DEFAULT '', source_path TEXT NOT NULL DEFAULT '', source_range TEXT NOT NULL DEFAULT '', revision INTEGER NOT NULL, created_at TIMESTAMPTZ NOT NULL); CREATE TABLE IF NOT EXISTS article_revisions (article_id TEXT NOT NULL, revision INTEGER NOT NULL, title TEXT NOT NULL, body TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL, PRIMARY KEY(article_id,revision));`
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
	a.Tags = normalizeTags(a.Tags)
	var current int
	var oldShortID string
	var oldTitle, oldBody, oldCategory, oldTags, oldSource string
	err := s.db.QueryRow(s.q(`SELECT short_id,revision,title,body,category,tags,source_path FROM articles WHERE id=? AND project_slug=?`), a.ID, a.ProjectSlug).Scan(&oldShortID, &current, &oldTitle, &oldBody, &oldCategory, &oldTags, &oldSource)
	if err == nil {
		a.ShortID = oldShortID
		if oldTitle == a.Title && oldBody == a.Body && oldCategory == a.Category && oldTags == a.Tags && oldSource == a.SourcePath {
			a.Revision, a.CreatedAt = current, time.Now().UTC()
			return a, nil
		}
		a.Revision = current + 1
		a.CreatedAt = time.Now().UTC()
		_, err = s.db.Exec(s.q(`UPDATE articles SET title=?,body=?,category=?,tags=?,source_path=?,source_range=?,revision=?,created_at=? WHERE id=? AND project_slug=?`), a.Title, a.Body, a.Category, a.Tags, a.SourcePath, a.SourceRange, a.Revision, a.CreatedAt, a.ID, a.ProjectSlug)
		if err == nil {
			_, err = s.db.Exec(s.q(`INSERT INTO article_revisions(article_id,revision,title,body,created_at) VALUES(?,?,?,?,?)`), a.ID, a.Revision, a.Title, a.Body, a.CreatedAt)
		}
		return a, err
	}
	if err != sql.ErrNoRows {
		return a, err
	}
	a.Revision = 1
	a.CreatedAt = time.Now().UTC()
	_, e := s.db.Exec(s.q(`INSERT INTO articles(id,short_id,project_slug,title,body,category,tags,source_path,source_range,revision,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`), a.ID, a.ShortID, a.ProjectSlug, a.Title, a.Body, a.Category, a.Tags, a.SourcePath, a.SourceRange, a.Revision, a.CreatedAt)
	if e == nil {
		_, e = s.db.Exec(s.q(`INSERT INTO article_revisions(article_id,revision,title,body,created_at) VALUES(?,?,?,?,?)`), a.ID, a.Revision, a.Title, a.Body, a.CreatedAt)
	}
	return a, e
}
func (s *Store) FindShort(short string) (Article, error) {
	var a Article
	e := s.db.QueryRow(s.q(`SELECT id,short_id,project_slug,title,body,category,tags,source_path,source_range,revision,created_at FROM articles WHERE short_id=?`), short).Scan(&a.ID, &a.ShortID, &a.ProjectSlug, &a.Title, &a.Body, &a.Category, &a.Tags, &a.SourcePath, &a.SourceRange, &a.Revision, &a.CreatedAt)
	return a, e
}
func (s *Store) FindArticle(projectSlug, id string) (Article, error) {
	var a Article
	e := s.db.QueryRow(s.q(`SELECT id,short_id,project_slug,title,body,category,tags,source_path,source_range,revision,created_at FROM articles WHERE project_slug=? AND id=?`), projectSlug, id).Scan(&a.ID, &a.ShortID, &a.ProjectSlug, &a.Title, &a.Body, &a.Category, &a.Tags, &a.SourcePath, &a.SourceRange, &a.Revision, &a.CreatedAt)
	return a, e
}
func normalizeTags(tags string) string {
	var values []string
	if json.Unmarshal([]byte(tags), &values) == nil {
		b, _ := json.Marshal(values)
		return string(b)
	}
	return tags
}
func (s *Store) Revisions(projectSlug, id string) ([]Article, error) {
	a, e := s.FindArticle(projectSlug, id)
	if e != nil {
		return nil, e
	}
	rows, e := s.db.Query(s.q(`SELECT revision,title,body,created_at FROM article_revisions WHERE article_id=? ORDER BY revision DESC`), id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []Article
	for rows.Next() {
		r := a
		if e = rows.Scan(&r.Revision, &r.Title, &r.Body, &r.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// UpdateOrganization changes navigation metadata without modifying the article
// body or creating a content revision. It is intentionally separate from
// AddArticle so a drag-and-drop tree move cannot look like a documentation edit.
func (s *Store) UpdateOrganization(projectSlug, id, category string, tags []string) (Article, error) {
	a, err := s.FindArticle(projectSlug, id)
	if err != nil {
		return Article{}, err
	}
	encoded, _ := json.Marshal(tags)
	if _, err = s.db.Exec(s.q(`UPDATE articles SET category=?,tags=? WHERE project_slug=? AND id=?`), category, string(encoded), projectSlug, id); err != nil {
		return Article{}, err
	}
	a.Category, a.Tags = category, string(encoded)
	return a, nil
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
