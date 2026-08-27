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
