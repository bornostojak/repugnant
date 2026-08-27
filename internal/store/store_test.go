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
}
