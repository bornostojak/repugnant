package docs

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/bornostojak/repugnant/internal/project"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Generate(root string) (int, error) {
	c, e := project.Load(root)
	if e != nil {
		return 0, e
	}
	if !c.Output.Docs.Enabled {
		return 0, nil
	}
	out := filepath.Join(root, c.Output.Docs.Dir)
	if e = os.MkdirAll(out, 0755); e != nil {
		return 0, e
	}
	n := 0
	e = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".rpg", "docs", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		data, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		fs, e := Parse(path, string(data))
		if e != nil {
			return e
		}
		for _, f := range fs {
			if f.Kind != "article" && f.Kind != "quote" {
				continue
			}
			id := newID()
			title, tags := titleTags(f.Title)
			a := Article{ID: id, Title: title, Tags: tags, Markdown: f.Markdown, Quote: f.Quote, Revision: 1, Path: path}
			if e = os.WriteFile(filepath.Join(out, id+".md"), []byte(Render(a, "Unknown", "unknown", time.Now().UTC())), 0644); e != nil {
				return e
			}
			lines := strings.Split(string(data), "\n")
			line := lines[f.Start]
			idx := strings.Index(line, "$rPg")
			if idx < 0 {
				idx = strings.Index(line, "?rPg")
			}
			if idx >= 0 {
				lines[f.Start] = line[:idx] + "rPg: " + id
			}
			if e = os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644); e != nil {
				return e
			}
			n++
		}
		return nil
	})
	return n, e
}
func newID() string {
	b := make([]byte, 9)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
func titleTags(s string) (string, []string) {
	s = strings.TrimSpace(s)
	if strings.Contains(s, ",") {
		p := strings.Split(s, ",")
		return strings.TrimSpace(p[0]), trimAll(p[1:])
	}
	if strings.Contains(s, "/") {
		p := strings.Split(s, "/")
		return strings.TrimSpace(p[len(p)-1]), nil
	}
	return s, nil
}
func trimAll(v []string) []string {
	for i := range v {
		v[i] = strings.TrimSpace(v[i])
	}
	return v
}
