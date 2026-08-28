package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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

func TestHelp(t *testing.T) {
	for _, arg := range [][]string{{}, {"help"}, {"-h"}, {"--help"}} {
		var output bytes.Buffer
		if err := run(arg, &output, &bytes.Buffer{}); err != nil {
			t.Fatalf("run(%v) returned error: %v", arg, err)
		}
		for _, want := range []string{"generate", "push", "status", "project create"} {
			if !strings.Contains(output.String(), want) {
				t.Fatalf("run(%v) help output missing %q: %q", arg, want, output.String())
			}
		}
	}
}

func TestUnknownCommand(t *testing.T) {
	err := run([]string{"bogus"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("expected unknown command error, got: %v", err)
	}
}

func TestProjectWithoutCreate(t *testing.T) {
	err := run([]string{"project"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "rpg project create") {
		t.Fatalf("expected usage error, got: %v", err)
	}
}

func TestStatusNoPending(t *testing.T) {
	root := t.TempDir()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := run([]string{"status"}, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if output.String() != "rpg status: no pending documentation\n" {
		t.Fatalf("unexpected status output: %q", output.String())
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
