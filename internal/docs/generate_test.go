package docs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateTracksQuoteAndAppendsExplicitRevision(t *testing.T) {
	root := t.TempDir()
	writeTestConfig(t, root)
	path := filepath.Join(root, "main.go")
	source := "package main\n\n// ?rPg(Platform/Caching/Resolve cache, performance, redis)\n// ?~ Resolves a cache entry before the database.\n// context line\nif hit {\n  return value\n}\n// !rPg\n"
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	n, err := Generate(root)
	if err != nil || n != 1 {
		t.Fatalf("Generate = %d, %v", n, err)
	}
	updated, _ := os.ReadFile(path)
	parts := strings.Fields(string(updated))
	var id string
	for i, value := range parts {
		if value == "rPg:" && i+1 < len(parts) {
			id = parts[i+1]
			break
		}
	}
	if id == "" {
		t.Fatalf("marker not replaced: %s", updated)
	}
	docPath := filepath.Join(root, "docs", id+".md")
	doc, err := os.ReadFile(docPath)
	if err != nil || !strings.Contains(string(doc), "| Category | Platform/Caching |") || !strings.Contains(string(doc), "performance, redis") {
		t.Fatalf("doc = %s, %v", doc, err)
	}
	if err := os.WriteFile(path, []byte(strings.Replace(string(updated), "return value", "return refreshed", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(root); err == nil || !strings.Contains(err.Error(), "$rPg@"+id) {
		t.Fatalf("wanted drift error, got %v", err)
	}
	revision := "// $rPg@" + id + ": Cache invalidation\n// $~ Explain why the branch changed.\n// context line\nif hit {\n  return refreshed\n}\n// !rPg\n"
	if err := os.WriteFile(path, []byte("package main\n\n"+revision), 0o644); err != nil {
		t.Fatal(err)
	}
	if n, err = Generate(root); err != nil || n != 1 {
		t.Fatalf("revision Generate = %d, %v", n, err)
	}
	doc, _ = os.ReadFile(docPath)
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

// Regression test: when a tracked quote's own content is unchanged but its
// location in the file shifts (e.g. lines were added above it), the manifest
// must refresh the stored source range so later publishes don't carry stale
// line numbers forever.
func TestGenerateRefreshesSourceRangeWhenQuoteMoves(t *testing.T) {
	root := t.TempDir()
	writeTestConfig(t, root)
	path := filepath.Join(root, "main.go")
	source := "package main\n\n// ?rPg(Platform/Caching/Resolve, cache)\n// ?~ Explains it.\nwork()\n// !rPg\n"
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(root); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(root, ".rpg", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(before), `"SourceRange": "3-6"`) {
		t.Fatalf("unexpected initial source range: %s", before)
	}
	updated, _ := os.ReadFile(path)
	shifted := "// unrelated leading comment\n// another one\n" + string(updated)
	if err := os.WriteFile(path, []byte(shifted), 0o644); err != nil {
		t.Fatal(err)
	}
	if n, err := Generate(root); err != nil || n != 0 {
		t.Fatalf("shifted Generate = %d, %v", n, err)
	}
	after, err := os.ReadFile(filepath.Join(root, ".rpg", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(after), `"SourceRange": "3-6"`) {
		t.Fatalf("source range was not refreshed after the quote moved: %s", after)
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
