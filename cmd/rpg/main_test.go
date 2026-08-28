package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitRun runs a git command in dir and fails the test on error.
func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
}

func writeConfig(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, ".rpg"), 0o755); err != nil {
		t.Fatal(err)
	}
	config := "version: 1\nlangs: [go]\noutput:\n  docs: {enabled: true, dir: docs}\n  web: {enabled: false}\nhooks: {on_publish_failure: block}\n"
	if err := os.WriteFile(filepath.Join(root, "rpg.conf.yaml"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestGenerateStagedFlag confirms `rpg generate --staged` only rewrites staged
// files and ignores markers in files that are not staged.
func TestGenerateStagedFlag(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root)
	gitRun(t, root, "init")
	gitRun(t, root, "config", "user.name", "Test")
	gitRun(t, root, "config", "user.email", "test@example.test")
	marker := "package main\n\n// $rPg: Documented, tag\nfunc f() {}\n"
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(marker), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "b.go"), []byte(marker), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, root, "add", "--", "a.go")
	chdir(t, root)

	var out bytes.Buffer
	if err := run([]string{"generate", "--staged"}, &out, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "generated 1 documentation article(s)\n" {
		t.Fatalf("unexpected output: %q", out.String())
	}
	a, _ := os.ReadFile(filepath.Join(root, "a.go"))
	if strings.Contains(string(a), "$rPg") {
		t.Fatalf("staged file marker not rewritten: %s", a)
	}
	b, _ := os.ReadFile(filepath.Join(root, "b.go"))
	if !strings.Contains(string(b), "$rPg") {
		t.Fatalf("unstaged file must be untouched: %s", b)
	}
}

// TestPreCommitHookStagesRewrittenSource is the core regression: after the
// pre-commit hook rewrites a marker in the staged source, that rewrite must be
// staged so the committed source matches the generated docs, and an unstaged
// marker file must stay out of it entirely.
func TestPreCommitHookStagesRewrittenSource(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root)
	gitRun(t, root, "init")
	gitRun(t, root, "config", "user.name", "Test")
	gitRun(t, root, "config", "user.email", "test@example.test")
	marker := "package main\n\n// $rPg: Documented, tag\nfunc f() {}\n"
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(marker), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "b.go"), []byte(marker), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, root, "add", "--", "a.go", "rpg.conf.yaml")
	chdir(t, root)

	if err := run([]string{"hook", "pre-commit"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	// a.go's rewritten marker is staged (no unstaged diff for it), and the
	// generated doc under docs/ is staged too.
	stagedNames := gitOutput(t, root, "diff", "--cached", "--name-only")
	if !strings.Contains(stagedNames, "a.go") {
		t.Fatalf("rewritten source a.go was not staged; staged=%q", stagedNames)
	}
	if !strings.Contains(stagedNames, "docs/") {
		t.Fatalf("generated docs were not staged; staged=%q", stagedNames)
	}
	// a.go has no remaining unstaged changes: what was rewritten is committed.
	unstaged := gitOutput(t, root, "diff", "--name-only")
	if strings.Contains(unstaged, "a.go") {
		t.Fatalf("a.go still has unstaged changes after the hook: %q", unstaged)
	}
	// b.go was never staged, so the hook must not have documented it.
	b, _ := os.ReadFile(filepath.Join(root, "b.go"))
	if !strings.Contains(string(b), "$rPg") {
		t.Fatalf("unstaged b.go must be untouched: %s", b)
	}
	if strings.Contains(stagedNames, "b.go") {
		t.Fatalf("unstaged b.go must not be staged by the hook: %q", stagedNames)
	}
}

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
