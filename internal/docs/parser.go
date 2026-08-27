package docs

import (
	"fmt"
	"path/filepath"
	"strings"
)

var extensions = map[string]string{
	".go": "//", ".ts": "//", ".tsx": "//", ".js": "//", ".jsx": "//", ".svelte": "//", ".rs": "//", ".c": "//", ".h": "//", ".cc": "//", ".cpp": "//", ".dart": "//", ".groovy": "//", ".qml": "//", ".py": "#", ".rb": "#", ".sh": "#", ".bash": "#", ".yaml": "#", ".yml": "#",
}

func Parse(path, source string) ([]Finding, error) {
	prefix, ok := extensions[filepath.Ext(path)]
	if !ok {
		return nil, nil
	}
	lines := strings.Split(source, "\n")
	var results []Finding
	for i := 0; i < len(lines); i++ {
		body, ok := commentBody(lines[i], prefix)
		if !ok {
			continue
		}
		kind, payload := marker(body)
		if kind == "" {
			continue
		}
		if kind == "end" {
			return nil, fmt.Errorf("%s:%d: !rPg has no matching ?rPg", path, i+1)
		}
		f := Finding{Kind: kind, Start: i, End: i, Title: payload}
		if kind == "stable" {
			f.ID = payload
		}
		if kind == "revision" {
			f.ID, f.Title = splitRevision(payload)
		}
		j := i + 1
		for j < len(lines) {
			b, ok := commentBody(lines[j], prefix)
			if !ok {
				break
			}
			if strings.HasPrefix(b, "$~") || strings.HasPrefix(b, "?~") {
				f.Markdown = append(f.Markdown, strings.TrimSpace(b[2:]))
				j++
				continue
			}
			break
		}
		if kind == "quote" || kind == "stable" {
			end := -1
			for k := j; k < len(lines); k++ {
				b, ok := commentBody(lines[k], prefix)
				if ok && strings.TrimSpace(b) == "!rPg" {
					end = k
					break
				}
			}
			if end >= 0 {
				f.Quote = unindent(strings.Join(lines[j:end], "\n"))
				f.End = end
				if kind == "stable" {
					f.Kind = "tracked"
				} else {
					f.Kind = "quote"
				}
			} else if kind == "quote" {
				return nil, fmt.Errorf("%s:%d: ?rPg requires a closing !rPg", path, i+1)
			}
		}
		results = append(results, f)
		if f.Kind == "quote" || f.Kind == "tracked" {
			i = f.End
		}
	}
	return results, nil
}
func commentBody(line, prefix string) (string, bool) {
	t := strings.TrimLeft(line, " \t")
	if !strings.HasPrefix(t, prefix) {
		return "", false
	}
	return strings.TrimSpace(strings.TrimPrefix(t, prefix)), true
}
func marker(s string) (string, string) {
	s = strings.TrimSpace(s)
	switch {
	case strings.HasPrefix(s, "$rPg@"):
		return "revision", strings.TrimSpace(strings.TrimPrefix(s, "$rPg@"))
	case strings.HasPrefix(s, "$rPg:"):
		return "article", strings.TrimSpace(strings.TrimPrefix(s, "$rPg:"))
	case strings.HasPrefix(s, "$rPg(") && strings.HasSuffix(s, ")"):
		return "article", strings.TrimSuffix(strings.TrimPrefix(s, "$rPg("), ")")
	case strings.HasPrefix(s, "?rPg:"):
		return "quote", strings.TrimSpace(strings.TrimPrefix(s, "?rPg:"))
	case strings.HasPrefix(s, "?rPg(") && strings.HasSuffix(s, ")"):
		return "quote", strings.TrimSuffix(strings.TrimPrefix(s, "?rPg("), ")")
	case strings.HasPrefix(s, "rPg:"):
		return "stable", strings.TrimSpace(strings.TrimPrefix(s, "rPg:"))
	case s == "!rPg":
		return "end", ""
	}
	return "", ""
}
func splitRevision(s string) (string, string) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(s), ""
}
func unindent(s string) string {
	lines := strings.Split(s, "\n")
	min := -1
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		n := len(l) - len(strings.TrimLeft(l, " \t"))
		if min < 0 || n < min {
			min = n
		}
	}
	if min > 0 {
		for i, l := range lines {
			if len(l) >= min {
				lines[i] = l[min:]
			}
		}
	}
	return strings.Join(lines, "\n")
}
