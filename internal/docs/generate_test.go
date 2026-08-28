package docs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGenerateRewritesCleanMarkerAndAppendsRevision covers the full clean-output
// contract: authoring markers and their $~/?~ prose are moved into the doc, the
// closing !rPg is dropped while the quoted code stays, the marker keeps only the
// title plus a docs backlink, re-generation is a no-op, and an explicit
// revision keyed by "$rPg@{id}" still appends to the article.
func TestGenerateRewritesCleanMarkerAndAppendsRevision(t *testing.T) {
	root := t.TempDir()
	writeTestConfig(t, root)
	path := filepath.Join(root, "main.go")
	source := "package main\n\n// ?rPg(Platform/Caching/Resolve cache, performance, redis)\n// ?~ Resolves a cache entry before the database.\nif hit {\n  return value\n}\n// !rPg\n"
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	n, err := Generate(root)
	if err != nil || n != 1 {
		t.Fatalf("Generate = %d, %v", n, err)
	}
	updated, _ := os.ReadFile(path)
	// Authoring syntax and prose are gone; the quoted code remains in place.
	for _, gone := range []string{"?rPg", "$rPg", "?~", "$~", "!rPg", "Resolves a cache entry"} {
		if strings.Contains(string(updated), gone) {
			t.Fatalf("expected %q removed from source:\n%s", gone, updated)
		}
	}
	if !strings.Contains(string(updated), "// rPg: Resolve cache") {
		t.Fatalf("marker title not written:\n%s", updated)
	}
	if !strings.Contains(string(updated), "return value") {
		t.Fatalf("quoted code was not retained:\n%s", updated)
	}

	// Exactly one article was written; recover its id from the docs directory.
	found, _ := filepath.Glob(filepath.Join(root, "docs", "*.md"))
	if len(found) != 1 {
		t.Fatalf("expected 1 doc, got %v", found)
	}
	id := strings.TrimSuffix(filepath.Base(found[0]), ".md")
	if !strings.Contains(string(updated), "// ~ docs/"+id+".md") {
		t.Fatalf("marker missing docs backlink for %s:\n%s", id, updated)
	}
	doc, _ := os.ReadFile(found[0])
	if !strings.Contains(string(doc), "| Category | Platform/Caching |") || !strings.Contains(string(doc), "performance, redis") || !strings.Contains(string(doc), "Resolves a cache entry") {
		t.Fatalf("doc missing moved metadata/prose: %s", doc)
	}

	// Re-generation is a clean no-op.
	if n, err := Generate(root); err != nil || n != 0 {
		t.Fatalf("second Generate = %d, %v", n, err)
	}

	// An explicit revision keyed by the id still appends to the article.
	revision := "package main\n\n// $rPg@" + id + ": Cache invalidation\n// $~ Explain why the branch changed.\n"
	if err := os.WriteFile(path, []byte(revision), 0o644); err != nil {
		t.Fatal(err)
	}
	if n, err := Generate(root); err != nil || n != 1 {
		t.Fatalf("revision Generate = %d, %v", n, err)
	}
	doc, _ = os.ReadFile(found[0])
	if !strings.Contains(string(doc), "# Revision 2") || !strings.Contains(string(doc), "Explain why") {
		t.Fatalf("revision missing: %s", doc)
	}
}

// Regression test for a real bug hit while dogfooding the tool: a standalone
// article followed later in the same file by an unrelated quote used to make
// the second Generate run fail with a false "documented quote changed" error,
// because the first marker's forward scan for !rPg walked straight through
// the second marker and matched its closing tag.
func TestGenerateStandaloneArticleFollowedByLaterQuoteIsStable(t *testing.T) {
	root := t.TempDir()
	writeTestConfig(t, root)
	path := filepath.Join(root, "main.go")
	source := "package main\n\n// $rPg(Services/Cache/Resolve, cache)\n// $~ Reads hit cache first.\nfunc Resolve() {}\n\nfunc guarded() {\n  // ?rPg(Services/Cache/Guard, cache)\n  // ?~ Serializes repopulation.\n  work()\n  // !rPg\n}\n"
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	if n, err := Generate(root); err != nil || n != 2 {
		t.Fatalf("first Generate = %d, %v", n, err)
	}
	// Nothing changed: re-running generate must be a clean no-op, not a
	// false quote-drift error caused by the first marker's stale forward scan.
	if n, err := Generate(root); err != nil || n != 0 {
		t.Fatalf("second Generate = %d, %v", n, err)
	}
}

// TestGenerateInlineRevision drives the fast revision syntax: append $#/$~ lines
// under an existing clean marker and generate. The article keeps its title, the
// doc gains a subtitled Revision 2, and the source returns to the clean marker.
func TestGenerateInlineRevision(t *testing.T) {
	root := t.TempDir()
	writeTestConfig(t, root)
	path := filepath.Join(root, "main.go")
	if err := os.WriteFile(path, []byte("package main\n\n// $rPg: Cache resolution, cache\n// $~ Reads hit the cache first.\nfunc f() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if n, err := Generate(root); err != nil || n != 1 {
		t.Fatalf("initial Generate = %d, %v", n, err)
	}
	found, _ := filepath.Glob(filepath.Join(root, "docs", "*.md"))
	if len(found) != 1 {
		t.Fatalf("expected 1 doc, got %v", found)
	}
	id := strings.TrimSuffix(filepath.Base(found[0]), ".md")

	// Author a revision inline: keep the clean marker, append $#/$~ below it.
	revised := "package main\n\n// rPg: Cache resolution\n// ~ docs/" + id + ".md\n// $# striped locks\n// $~ Now uses 16 striped locks keyed by hash.\nfunc f() {}\n"
	if err := os.WriteFile(path, []byte(revised), 0o644); err != nil {
		t.Fatal(err)
	}
	if n, err := Generate(root); err != nil || n != 1 {
		t.Fatalf("revision Generate = %d, %v", n, err)
	}

	// The doc keeps its title, and gains a subtitled Revision 2 with the body.
	doc, _ := os.ReadFile(found[0])
	for _, want := range []string{"# Cache resolution", "# Revision 1", "# Revision 2 — striped locks", "Now uses 16 striped locks"} {
		if !strings.Contains(string(doc), want) {
			t.Fatalf("doc missing %q:\n%s", want, doc)
		}
	}

	// The source is back to the clean marker (no $#/$~ left) and re-runs cleanly.
	updated, _ := os.ReadFile(path)
	for _, gone := range []string{"$#", "$~"} {
		if strings.Contains(string(updated), gone) {
			t.Fatalf("expected %q consumed from source:\n%s", gone, updated)
		}
	}
	if !strings.Contains(string(updated), "// rPg: Cache resolution\n// ~ docs/"+id+".md") {
		t.Fatalf("marker not restored to clean form:\n%s", updated)
	}
	if n, err := Generate(root); err != nil || n != 0 {
		t.Fatalf("re-generate after revision = %d, %v", n, err)
	}
}

func TestGenerateHonorsConfiguredLanguages(t *testing.T) {
	root := t.TempDir()
	writeTestConfig(t, root)
	config, _ := os.ReadFile(filepath.Join(root, "rpg.conf.yaml"))
	if err := os.WriteFile(filepath.Join(root, "rpg.conf.yaml"), []byte(strings.Replace(string(config), "langs: [go]", "langs: [ruby]", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("// $rPg: ignored\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if n, err := Generate(root); err != nil || n != 0 {
		t.Fatalf("Generate = %d, %v", n, err)
	}
}

// TestGenerateEmbedsWebAndDocsBacklinks verifies that when both web and docs
// output are configured, the rewritten marker carries the web /a/{id} link first
// and the relative docs path second, and that re-generation is a clean no-op.
func TestGenerateEmbedsWebAndDocsBacklinks(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".rpg"), 0o755); err != nil {
		t.Fatal(err)
	}
	config := "version: 1\nlangs: [go]\noutput:\n  docs: {enabled: true, dir: docs}\n  web: {enabled: true, endpoint: 'http://127.0.0.1:8080'}\nhooks: {on_publish_failure: block}\nproject:\n  slug: demo\n  api_url: http://127.0.0.1:8080/api/projects/demo/articles\n  api_key: secret\n"
	if err := os.WriteFile(filepath.Join(root, "rpg.conf.yaml"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "main.go")
	if err := os.WriteFile(path, []byte("package main\n\n// $rPg: Cache resolution, cache\nfunc f() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if n, err := Generate(root); err != nil || n != 1 {
		t.Fatalf("Generate = %d, %v", n, err)
	}
	updated, _ := os.ReadFile(path)
	found, _ := filepath.Glob(filepath.Join(root, "docs", "*.md"))
	if len(found) != 1 {
		t.Fatalf("expected 1 doc, got %v", found)
	}
	id := strings.TrimSuffix(filepath.Base(found[0]), ".md")
	wantHeader := "// rPg: Cache resolution\n// ~ http://127.0.0.1:8080/a/" + id + "\n// ~ docs/" + id + ".md\n"
	if !strings.Contains(string(updated), wantHeader) {
		t.Fatalf("marker header = \n%s\nwant to contain:\n%s", updated, wantHeader)
	}
	// The generated doc's Web column points at the same permalink.
	doc, _ := os.ReadFile(found[0])
	if !strings.Contains(string(doc), "/a/"+id+")") {
		t.Fatalf("doc missing /a/ web link: %s", doc)
	}
	if n, err := Generate(root); err != nil || n != 0 {
		t.Fatalf("second Generate = %d, %v (marker: %s)", n, err, updated)
	}
}

func writeTestConfig(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, ".rpg"), 0o755); err != nil {
		t.Fatal(err)
	}
	config := "version: 1\nlangs: [go]\noutput:\n  docs: {enabled: true, dir: docs}\n  web: {enabled: false}\nhooks: {on_publish_failure: block}\n"
	if err := os.WriteFile(filepath.Join(root, "rpg.conf.yaml"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
}
