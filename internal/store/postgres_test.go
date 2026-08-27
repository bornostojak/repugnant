package store

import (
	"os"
	"testing"
	"time"
)

func TestPostgresStorageContract(t *testing.T) {
	dsn := os.Getenv("RPG_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("RPG_TEST_POSTGRES_DSN is not configured")
	}
	s, err := Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	slug := "postgres-contract-" + randomSuffix(t)
	p, err := s.CreateProject("PostgreSQL contract", slug)
	if err != nil {
		t.Fatal(err)
	}
	first, err := s.AddArticle(Article{ID: "postgres-article-" + randomSuffix(t), ProjectSlug: p.Slug, Title: "First", Body: "body", Category: "Platform", Tags: `["postgres"]`, SourcePath: "main.go", SourceRange: "1-3"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.AddArticle(first); err != nil {
		t.Fatal(err)
	}
	current, err := s.AddArticle(Article{ID: first.ID, ProjectSlug: p.Slug, Title: "Second", Body: "revised", Category: "Platform", Tags: `["postgres"]`, SourcePath: "main.go", SourceRange: "1-4"})
	if err != nil || current.Revision != 2 {
		t.Fatalf("revision = %+v, %v", current, err)
	}
	rows, err := s.Revisions(p.Slug, first.ID)
	if err != nil || len(rows) != 2 {
		t.Fatalf("revisions = %+v, %v", rows, err)
	}
}
func randomSuffix(t *testing.T) string {
	t.Helper()
	return time.Now().UTC().Format("150405.000000000")
}
