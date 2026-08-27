package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestVersion(t *testing.T) {
	var output bytes.Buffer
	if err := run([]string{"--version"}, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if output.String() != version+"\n" {
		t.Fatalf("unexpected version output: %q", output.String())
	}
}

func TestHookCommand(t *testing.T) {
	var output bytes.Buffer
	if err := run([]string{"hook", "pre-commit"}, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if output.String() != "rpg pre-commit: no pending documentation changes\n" {
		t.Fatalf("unexpected hook output: %q", output.String())
	}
}

func TestInitCommand(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git", "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"init"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "rpg.conf.yaml")); err != nil {
		t.Fatal(err)
	}
}
