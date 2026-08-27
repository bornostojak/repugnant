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
)

func Push(root string) (int, error) {
	c, e := project.Load(root)
	if e != nil {
		return 0, e
	}
	if !c.Output.Web.Enabled {
		return 0, fmt.Errorf("web output is disabled")
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
		p, _ := json.Marshal(map[string]string{"id": strings.TrimSuffix(filepath.Base(f), ".md"), "title": title, "body": string(b)})
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
