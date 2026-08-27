package docs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bornostojak/repugnant/internal/project"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Push(root string) (int, error) {
	c, e := project.Load(root)
	if e != nil {
		return 0, e
	}
	if !c.Output.Web.Enabled {
		return 0, fmt.Errorf("web output is disabled")
	}
	manifest, e := loadManifest(root)
	if e != nil {
		return 0, e
	}
	fs, e := filepath.Glob(filepath.Join(root, c.Output.Docs.Dir, "*.md"))
	if e != nil {
		return 0, e
	}
	n := 0
	for _, f := range fs {
		b, e := os.ReadFile(f)
		if e != nil {
			return n, e
		}
		title := "Untitled"
		for _, l := range strings.Split(string(b), "\n") {
			if strings.HasPrefix(l, "# ") {
				title = strings.TrimPrefix(l, "# ")
				break
			}
		}
		id := strings.TrimSuffix(filepath.Base(f), ".md")
		record := manifest.Articles[id]
		p, _ := json.Marshal(map[string]any{"id": id, "title": title, "body": string(b), "category": record.Category, "tags": record.Tags, "source_path": record.Path})
		r, e := http.NewRequest(http.MethodPost, c.Project.APIURL, bytes.NewReader(p))
		if e != nil {
			return n, e
		}
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-RPG-API-Key", c.Project.APIKey)
		resp, e := http.DefaultClient.Do(r)
		if e != nil {
			return n, e
		}
		resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return n, fmt.Errorf("publish %s: %s", f, resp.Status)
		}
		n++
	}
	return n, nil
}

// RecordPending leaves a compact, ignored diagnostic for a later successful
// rpg push. The Markdown itself remains authoritative in the configured docs
// directory, so this cache cannot lose user edits.
func RecordPending(root string, cause error) error {
	if err := os.MkdirAll(filepath.Join(root, ".rpg"), 0o755); err != nil {
		return err
	}
	payload, _ := json.MarshalIndent(map[string]string{"failed_at": time.Now().UTC().Format(time.RFC3339), "error": cause.Error()}, "", "  ")
	if err := os.WriteFile(filepath.Join(root, ".rpg", "pending.json"), append(payload, '\n'), 0o600); err != nil {
		return fmt.Errorf("save pending publish state: %w", err)
	}
	return nil
}
