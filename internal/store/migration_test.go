package store

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMigrateUpgradesEarlierArticleSchema(t *testing.T) {
	dsn := "file:migrate-old?mode=memory&cache=shared"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`CREATE TABLE articles (id TEXT PRIMARY KEY, short_id TEXT NOT NULL UNIQUE, project_slug TEXT NOT NULL, title TEXT NOT NULL, body TEXT NOT NULL, revision INTEGER NOT NULL, created_at TIMESTAMP NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s, err := Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.db.Exec(`SELECT source_range FROM articles`); err != nil {
		t.Fatalf("source_range missing after migration: %v", err)
	}
}
