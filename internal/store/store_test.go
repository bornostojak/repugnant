package store

import "testing"

func TestSQLiteProjectAndArticle(t *testing.T) {
	s, err := Open("sqlite", "file:store-test?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	p, err := s.CreateProject("Test Project", "test-project")
	if err != nil {
		t.Fatal(err)
	}
	if ok, _ := s.Authorize(p.Slug, p.APIKey); !ok {
		t.Fatal("key should authorize")
	}
	a, err := s.AddArticle(Article{ProjectSlug: p.Slug, Title: "A", Body: "body"})
	if err != nil {
		t.Fatal(err)
	}
	short, err := s.FindShort(a.ShortID)
	if err != nil || short.ID != a.ID {
		t.Fatalf("short lookup: %+v %v", short, err)
	}
	article, err := s.FindArticle(p.Slug, a.ID)
	if err != nil || article.Title != "A" {
		t.Fatalf("article lookup: %+v %v", article, err)
	}
	updated, err := s.AddArticle(Article{ID: a.ID, ProjectSlug: p.Slug, Title: "A revised", Body: "next"})
	if err != nil || updated.Revision != 2 {
		t.Fatalf("revision = %+v, err=%v", updated, err)
	}
	unchanged, err := s.AddArticle(Article{ID: a.ID, ProjectSlug: p.Slug, Title: "A revised", Body: "next"})
	if err != nil || unchanged.Revision != 2 || unchanged.ShortID != a.ShortID { t.Fatalf("idempotent publish = %+v, err=%v", unchanged, err) }
	organized, err := s.UpdateOrganization(p.Slug, a.ID, "Platform/Caching", []string{"cache", "fast"})
	if err != nil || organized.Category != "Platform/Caching" || organized.Revision != 2 { t.Fatalf("organization = %+v, err=%v", organized, err) }
}

// Publishing the same content again after only its source location changed
// (e.g. lines shifted elsewhere in the file) must not bump the revision, but
// it must still refresh the stored source_range instead of leaving it stale.
func TestAddArticleRefreshesSourceRangeWithoutNewRevision(t *testing.T) {
	s, err := Open("sqlite", "file:store-range-test?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	p, err := s.CreateProject("Test Project", "test-project")
	if err != nil {
		t.Fatal(err)
	}
	a, err := s.AddArticle(Article{ProjectSlug: p.Slug, Title: "A", Body: "body", SourcePath: "main.go", SourceRange: "3-6"})
	if err != nil {
		t.Fatal(err)
	}
	moved, err := s.AddArticle(Article{ID: a.ID, ProjectSlug: p.Slug, Title: "A", Body: "body", SourcePath: "main.go", SourceRange: "9-12"})
	if err != nil {
		t.Fatal(err)
	}
	if moved.Revision != a.Revision {
		t.Fatalf("a pure location change must not create a new revision, got revision %d (was %d)", moved.Revision, a.Revision)
	}
	stored, err := s.FindArticle(p.Slug, a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.SourceRange != "9-12" {
		t.Fatalf("stored source_range was not refreshed, got %q", stored.SourceRange)
	}
}
