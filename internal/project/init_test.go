package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCreatesProjectFilesAndHooks(t *testing.T) {
	root := newRepository(t)
	if err := Init(root, false); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(root); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{".rpg", ".git/hooks/pre-commit", ".git/hooks/pre-push"} {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Fatalf("expected %s: %v", path, err)
		}
	}
	gitignore, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil || !strings.Contains(string(gitignore), ".rpg/") {
		t.Fatalf(".gitignore did not contain staging path: %v", err)
	}
}

func TestInitIsIdempotent(t *testing.T) {
	root := newRepository(t)
	if err := Init(root, false); err != nil {
		t.Fatal(err)
	}
	if err := Init(root, false); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(string(data), ".rpg/"); got != 1 {
		t.Fatalf(".rpg entries = %d, want 1", got)
	}
}

func TestInitBacksUpExistingHook(t *testing.T) {
	root := newRepository(t)
	hook := filepath.Join(root, ".git", "hooks", "pre-commit")
	if err := os.WriteFile(hook, []byte("#!/bin/sh\necho custom\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Init(root, false); err != nil {
		t.Fatal(err)
	}
	backup, err := os.ReadFile(hook + ".rpg-backup")
	if err != nil || !strings.Contains(string(backup), "custom") {
		t.Fatalf("backup = %q, err = %v", backup, err)
	}
	installed, err := os.ReadFile(hook)
	if err != nil || !strings.Contains(string(installed), "rpg-backup") {
		t.Fatalf("installed hook does not invoke backup: %q, err = %v", installed, err)
	}
}

func TestConfigValidation(t *testing.T) {
	config := DefaultConfig()
	config.Output.Web.Enabled = true
	if err := config.Validate(); err == nil {
		t.Fatal("expected missing endpoint error")
	}
}

func newRepository(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git", "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}
