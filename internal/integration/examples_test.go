package integration_test

import (
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/bornostojak/repugnant/internal/httpapi"
	"github.com/bornostojak/repugnant/internal/store"
)

// TestExamplesExerciseRealHooks copies every versioned fixture into its own
// temporary Git worktree, lets the installed pre-commit hook generate and
// stage documentation, then publishes it through the real rpg CLI to an HTTP
// server. Fixtures stay ordinary source code in tests/examples; no .git data
// is kept in the repository.
func TestExamplesExerciseRealHooks(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	repo := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "../.."))
	binDir := t.TempDir()
	bin := filepath.Join(binDir, "rpg")
	command(t, repo, nil, "go", "build", "-o", bin, "./cmd/rpg")
	database, err := store.Open("sqlite", "file:examples-integration?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	server := httptest.NewServer(httpapi.NewWithStore(slog.New(slog.NewTextHandler(io.Discard, nil)), database).Handler())
	defer server.Close()
	examples, err := os.ReadDir(filepath.Join(repo, "tests", "examples"))
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, example := range examples {
		if example.IsDir() {
			count++
		}
	}
	if count != 10 {
		t.Fatalf("expected 10 examples, found %d", count)
	}
	for _, example := range examples {
		if !example.IsDir() {
			continue
		}
		t.Run(example.Name(), func(t *testing.T) {
			root := t.TempDir()
			copyTree(t, filepath.Join(repo, "tests", "examples", example.Name()), root)
			env := append(os.Environ(), "PATH="+binDir+":"+os.Getenv("PATH"))
			command(t, root, env, "git", "init")
			command(t, root, env, "git", "config", "user.name", "Fixture Author")
			command(t, root, env, "git", "config", "user.email", "fixture@example.test")
			command(t, root, env, bin, "init")
			p, err := database.CreateProject(example.Name(), example.Name())
			if err != nil {
				t.Fatal(err)
			}
			config := filepath.Join(root, "rpg.conf.yaml")
			data, err := os.ReadFile(config)
			if err != nil {
				t.Fatal(err)
			}
			configured := strings.Replace(string(data), "enabled: false", "enabled: true", 1)
			configured = strings.Replace(configured, "# endpoint: http://127.0.0.1:8080", "endpoint: "+server.URL, 1)
			data = []byte(configured + "\nproject:\n  slug: " + p.Slug + "\n  api_url: " + server.URL + "/api/projects/" + p.Slug + "/articles\n  api_key: " + p.APIKey + "\n")
			if err := os.WriteFile(config, data, 0o644); err != nil {
				t.Fatal(err)
			}
			command(t, root, env, "git", "add", ".")
			command(t, root, env, "git", "commit", "-m", "document fixture")
			docs, err := filepath.Glob(filepath.Join(root, "docs", "*.md"))
			if err != nil || len(docs) != 1 {
				t.Fatalf("generated docs = %v, %v", docs, err)
			}
			command(t, root, env, bin, "push")
			articles, err := database.ListArticles(p.Slug, "")
			if err != nil || len(articles) != 1 {
				t.Fatalf("published articles = %+v, %v", articles, err)
			}
		})
	}
}

func command(t *testing.T, dir string, env []string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir, cmd.Env = dir, env
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s failed: %v\n%s", name, strings.Join(args, " "), err, output)
	}
}
func copyTree(t *testing.T, source, destination string) {
	t.Helper()
	if err := filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(source, path)
		target := filepath.Join(destination, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	}); err != nil {
		t.Fatal(err)
	}
}
