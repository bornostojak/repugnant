package docs

import "testing"

func TestParseQuoteAndUnindent(t *testing.T) {
	src := "func f() {\n  // ?rPg: Cache\n  // ?~ Why this works\n  // quoted comment\n  if ok {\n    work()\n  }\n  // !rPg\n}"
	got, err := Parse("f.go", src)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Kind != "quote" || got[0].Quote != "// quoted comment\nif ok {\n  work()\n}" {
		t.Fatalf("%+v", got)
	}
}
func TestParseRejectsUnclosedQuote(t *testing.T) {
	_, err := Parse("f.py", "# ?rPg: no\nx=1")
	if err == nil {
		t.Fatal("wanted error")
	}
}

// After generation a quoted article is left as a title-only marker with inert
// "~ <backlink>" lines and no closing !rPg. Re-parsing must skip those backlink
// lines (never folding them into the retained code) and yield a single no-op
// stable finding whose BodyStart points past them.
func TestParseSkipsBacklinkLines(t *testing.T) {
	src := "func f() {\n  // rPg: Cache guard\n  // ~ http://host/a/abc123\n  // ~ ../docs/abc123.md\n  work()\n}\n"
	got, err := Parse("f.go", src)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Kind != "stable" {
		t.Fatalf("want one stable finding, got %+v", got)
	}
	if len(got[0].Markdown) != 0 {
		t.Fatalf("backlink lines must not be captured as markdown: %+v", got[0].Markdown)
	}
	if got[0].BodyStart != 4 {
		t.Fatalf("BodyStart should skip the two ~ lines to the code at line index 4, got %d", got[0].BodyStart)
	}
}

// A plain stable reference (a standalone article already generated once)
// must never absorb an unrelated !rPg that belongs to a later, independent
// quote in the same file: it has to stay a no-op "stable" finding instead of
// being misclassified as "tracked" with a quote spanning everything in
// between, which previously produced a false quote-drift error and a
// corrupted source range.
func TestParseStableArticleDoesNotAbsorbLaterQuote(t *testing.T) {
	src := "package p\n\n// rPg: abc123\n// $~ already generated\nfunc A() {}\n\nfunc B() {\n  // ?rPg: Quoted\n  // ?~ explain\n  work()\n  // !rPg\n}\n"
	got, err := Parse("f.go", src)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 findings, got %d: %+v", len(got), got)
	}
	if got[0].Kind != "stable" || got[0].Quote != "" {
		t.Fatalf("standalone stable reference must stay untracked, got %+v", got[0])
	}
	if got[1].Kind != "quote" || got[1].Quote != "work()" {
		t.Fatalf("independent quote must be parsed on its own, got %+v", got[1])
	}
}
