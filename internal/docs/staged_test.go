package docs

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitInit turns root into a real git repo with a committed baseline so that
// staging state is meaningful.
func gitInit(t *testing.T, root string) {
	t.Helper()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.name", "Test"},
		{"config", "user.email", "test@example.test"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
}

func gitAdd(t *testing.T, root string, paths ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"add", "--"}, paths...)...)
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
}

// TestGenerateStagedIgnoresUnstagedFiles proves --staged only rewrites markers
// in files that are staged, and reports exactly those files as changed.
func TestGenerateStagedIgnoresUnstagedFiles(t *testing.T) {
	root := t.TempDir()
	writeTestConfig(t, root)
	gitInit(t, root)

	staged := filepath.Join(root, "staged.go")
	unstaged := filepath.Join(root, "unstaged.go")
	marker := "package main\n\n// $rPg: Documented thing, cache\nfunc f() {}\n"
	if err := os.WriteFile(staged, []byte(marker), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(unstaged, []byte(marker), 0o644); err != nil {
		t.Fatal(err)
	}
	// Only stage one of the two marker files.
	gitAdd(t, root, "staged.go")

	res, err := GenerateWith(root, GenerateOptions{Staged: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Count != 1 {
		t.Fatalf("expected 1 article from the staged file, got %d", res.Count)
	}
	if len(res.ChangedFiles) != 1 || res.ChangedFiles[0] != "staged.go" {
		t.Fatalf("expected ChangedFiles=[staged.go], got %v", res.ChangedFiles)
	}

	stagedContent, _ := os.ReadFile(staged)
	if !strings.Contains(string(stagedContent), "rPg: ") || strings.Contains(string(stagedContent), "$rPg") {
		t.Fatalf("staged file marker was not rewritten: %s", stagedContent)
	}
	unstagedContent, _ := os.ReadFile(unstaged)
	if !strings.Contains(string(unstagedContent), "$rPg") {
		t.Fatalf("unstaged file must be left untouched: %s", unstagedContent)
	}
}

// TestGenerateStagedNoStagedFilesIsNoop confirms that with nothing staged,
// staged generation does nothing even when markers exist in the tree.
func TestGenerateStagedNoStagedFilesIsNoop(t *testing.T) {
	root := t.TempDir()
	writeTestConfig(t, root)
	gitInit(t, root)
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("// $rPg: Thing, tag\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := GenerateWith(root, GenerateOptions{Staged: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Count != 0 || len(res.ChangedFiles) != 0 {
		t.Fatalf("expected no-op, got count=%d changed=%v", res.Count, res.ChangedFiles)
	}
}

// TestStagedFilesExcludesDeletions verifies deleted files never appear in the
// staged set.
func TestStagedFilesExcludesDeletions(t *testing.T) {
	root := t.TempDir()
	gitInit(t, root)
	path := filepath.Join(root, "gone.go")
	if err := os.WriteFile(path, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitAdd(t, root, "gone.go")
	commit := exec.Command("git", "commit", "-m", "seed")
	commit.Dir = root
	if out, err := commit.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	gitAdd(t, root, "gone.go") // stage the deletion

	set, err := StagedFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	if set["gone.go"] {
		t.Fatalf("deleted file should not be staged for documentation: %v", set)
	}
}
