package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnabledLanguagesExplicitAndInferred(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := (Config{Langs: []string{"ruby", "go"}}).EnabledLanguages(root); !got["ruby"] || !got["go"] || got["typescript"] {
		t.Fatalf("explicit: %#v", got)
	}
	if got := (Config{}).EnabledLanguages(root); !got["go"] || got["ruby"] {
		t.Fatalf("inferred: %#v", got)
	}
	if got := (Config{}).EnabledLanguages(t.TempDir()); got != nil {
		t.Fatalf("unhinted should be permissive: %#v", got)
	}
}
