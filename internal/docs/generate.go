package docs

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/bornostojak/repugnant/internal/project"
)

const manifestName = "manifest.json"

type manifest struct {
	Articles map[string]manifestArticle `json:"articles"`
}
type manifestArticle struct {
	Path, SourceRange, QuoteHash, Title, Category string
	Tags                                          []string
	Revision                                      int
}

// GenerateOptions tunes a generation run.
type GenerateOptions struct {
	// Staged limits processing to files with staged (indexed) changes and
	// ignores rPg markers in files that are not staged. The pre-commit hook
	// uses this so a commit never absorbs markers from unrelated, unstaged
	// files, and so the source rewrites it does make match what is committed.
	Staged bool
}

// GenerateResult reports what a generation run produced.
type GenerateResult struct {
	// Count is the number of new articles plus appended revisions.
	Count int
	// ChangedFiles lists paths (relative to root) whose markers were rewritten
	// in place, so callers such as the pre-commit hook can stage them.
	ChangedFiles []string
}

// Generate is the backwards-compatible whole-working-tree entry point.
func Generate(root string) (int, error) {
	res, err := GenerateWith(root, GenerateOptions{})
	return res.Count, err
}

// GenerateWith only creates docs for new markers, appends explicit revisions,
// and rejects silent changes to tracked quote regions. Hand-edited Markdown is
// not regenerated or overwritten. With GenerateOptions.Staged it restricts work
// to files currently staged in git.
func GenerateWith(root string, opts GenerateOptions) (GenerateResult, error) {
	c, err := project.Load(root)
	if err != nil {
		return GenerateResult{}, err
	}
	if !c.Output.Docs.Enabled {
		return GenerateResult{}, nil
	}
	var staged map[string]bool
	if opts.Staged {
		staged, err = StagedFiles(root)
		if err != nil {
			return GenerateResult{}, err
		}
	}
	out := filepath.Join(root, c.Output.Docs.Dir)
	if err = os.MkdirAll(out, 0o755); err != nil {
		return GenerateResult{}, err
	}
	m, err := loadManifest(root)
	if err != nil {
		return GenerateResult{}, err
	}
	name, email := gitIdentity(root)
	enabled := c.EnabledLanguages(root)
	changed := 0
	manifestDirty := false
	var changedFiles []string
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != root && (path == out || entry.Name() == ".git" || entry.Name() == ".rpg" || entry.Name() == "node_modules") {
				return filepath.SkipDir
			}
			return nil
		}
		if len(enabled) > 0 && !enabled[Language(path)] {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if opts.Staged && !staged[rel] {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		findings, err := Parse(path, string(data))
		if err != nil {
			return err
		}
		if len(findings) == 0 {
			return nil
		}
		lines := strings.Split(string(data), "\n")
		var edits []lineEdit
		for _, f := range findings {
			switch f.Kind {
			case "article", "quote":
				id := newID()
				a := Article{ID: id, Title: f.Title, Category: f.Category, Tags: f.Tags, Markdown: f.Markdown, Quote: f.Quote, Revision: 1, Path: rel}
				if err := os.WriteFile(filepath.Join(out, id+".md"), []byte(Render(a, name, email, time.Now().UTC(), webArticleURL(c, id))), 0o644); err != nil {
					return err
				}
				m.Articles[id] = manifestArticle{Path: rel, SourceRange: sourceRange(f), QuoteHash: quoteHash(f.Quote), Title: f.Title, Category: f.Category, Tags: f.Tags, Revision: 1}
				edits = append(edits, markerEdits(lines, f, buildRefs(c, path, out, id), f.Title)...)
				changed++
			case "tracked":
				record, known := m.Articles[f.ID]
				if !known {
					return fmt.Errorf("%s:%d: rPg ID %q is not in .rpg/%s; restore the manifest or replace this marker with a new $rPg annotation", rel, f.Start+1, f.ID, manifestName)
				}
				if f.Quote != "" && record.QuoteHash != quoteHash(f.Quote) {
					return fmt.Errorf("%s:%d: documented quote changed for %s; add $rPg@%s: explain the change above the quote, followed by $~ Markdown and !rPg", rel, f.Start+1, f.ID, f.ID)
				}
				if newRange := sourceRange(f); record.Path != rel || record.SourceRange != newRange {
					record.Path, record.SourceRange = rel, newRange
					m.Articles[f.ID] = record
					manifestDirty = true
				}
			case "revision":
				record, known := m.Articles[f.ID]
				if !known {
					return fmt.Errorf("%s:%d: revision references unknown rPg ID %q", rel, f.Start+1, f.ID)
				}
				if err := appendRevision(filepath.Join(out, f.ID+".md"), f, record); err != nil {
					return fmt.Errorf("%s:%d: %w", rel, f.Start+1, err)
				}
				record.Revision++
				if f.Title != "" {
					record.Title = f.Title
				}
				if f.Quote != "" {
					record.QuoteHash = quoteHash(f.Quote)
				}
				record.SourceRange = sourceRange(f)
				m.Articles[f.ID] = record
				title := f.Title
				if title == "" {
					title = record.Title
				}
				edits = append(edits, markerEdits(lines, f, buildRefs(c, path, out, f.ID), title)...)
				changed++
			}
		}
		if len(edits) > 0 {
			lines = applyLineEdits(lines, edits)
			if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
				return err
			}
			changedFiles = append(changedFiles, rel)
		}
		return nil
	})
	if err != nil {
		return GenerateResult{Count: changed, ChangedFiles: changedFiles}, err
	}
	if changed > 0 || manifestDirty {
		err = saveManifest(root, m)
	}
	return GenerateResult{Count: changed, ChangedFiles: changedFiles}, err
}
func sourceRange(f Finding) string { return fmt.Sprintf("%d-%d", f.Start+1, f.End+1) }

func loadManifest(root string) (manifest, error) {
	m := manifest{Articles: map[string]manifestArticle{}}
	b, err := os.ReadFile(filepath.Join(root, ".rpg", manifestName))
	if os.IsNotExist(err) {
		return m, nil
	}
	if err != nil {
		return m, err
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return m, fmt.Errorf("read .rpg/%s: %w", manifestName, err)
	}
	if m.Articles == nil {
		m.Articles = map[string]manifestArticle{}
	}
	return m, nil
}
func saveManifest(root string, m manifest) error {
	if err := os.MkdirAll(filepath.Join(root, ".rpg"), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, ".rpg", manifestName), append(b, '\n'), 0o600)
}
func quoteHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
func newID() string {
	b := make([]byte, 9)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
// lineEdit replaces the inclusive line range [start, end] with replacement.
type lineEdit struct {
	start, end  int
	replacement []string
}

// markerEdits rewrites an authoring marker into the clean tracked form:
//
//	// rPg: <title>
//	// ~ <web link>        (only when web output is configured)
//	// ~ <relative docs path>
//
// The $~/?~ prose lines are dropped (already rendered into the article) and, for
// quoted findings, the closing !rPg line is removed while the quoted code stays.
// The returned edits are applied bottom-up so line shifts do not disturb one
// another.
func markerEdits(lines []string, f Finding, refs []string, title string) []lineEdit {
	prefix := markerPrefix(lines[f.Start])
	header := []string{prefix + "rPg: " + title}
	for _, r := range refs {
		header = append(header, prefix+"~ "+r)
	}
	edits := []lineEdit{{start: f.Start, end: f.BodyStart - 1, replacement: header}}
	if f.End > f.Start {
		edits = append(edits, lineEdit{start: f.End, end: f.End})
	}
	return edits
}

// markerPrefix returns the indentation and comment opener that precede a marker
// token, e.g. "    // " or "# ", so rewritten lines keep the file's style.
func markerPrefix(line string) string {
	for _, marker := range []string{"$rPg@", "$rPg", "?rPg", "rPg:"} {
		if i := strings.Index(line, marker); i >= 0 {
			return line[:i]
		}
	}
	return ""
}

// buildRefs assembles the backlinks a rewritten marker carries: the web article
// URL first (when web publishing is configured) and then the article's Markdown
// file relative to the source file, so `gf` opens it straight from an editor.
func buildRefs(c project.Config, sourcePath, docsDir, id string) []string {
	var refs []string
	if url := webArticleURL(c, id); url != "" {
		refs = append(refs, url)
	}
	if c.Output.Docs.Enabled {
		refs = append(refs, docsRelPath(sourcePath, docsDir, id))
	}
	return refs
}

func docsRelPath(sourcePath, docsDir, id string) string {
	target := filepath.Join(docsDir, id+".md")
	rel, err := filepath.Rel(filepath.Dir(sourcePath), target)
	if err != nil {
		return filepath.ToSlash(target)
	}
	return filepath.ToSlash(rel)
}

// applyLineEdits applies non-overlapping edits from the bottom of the file up so
// each edit's recorded line numbers stay valid as earlier ones change length.
func applyLineEdits(lines []string, edits []lineEdit) []string {
	sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
	for _, e := range edits {
		rebuilt := make([]string, 0, len(lines)-(e.end-e.start+1)+len(e.replacement))
		rebuilt = append(rebuilt, lines[:e.start]...)
		rebuilt = append(rebuilt, e.replacement...)
		rebuilt = append(rebuilt, lines[e.end+1:]...)
		lines = rebuilt
	}
	return lines
}
func gitIdentity(root string) (string, string) {
	get := func(key, fallback string) string {
		out, err := exec.Command("git", "-C", root, "config", "--get", key).Output()
		if err != nil || strings.TrimSpace(string(out)) == "" {
			return fallback
		}
		return strings.TrimSpace(string(out))
	}
	return get("user.name", "Unknown"), get("user.email", "unknown")
}
func webArticleURL(c project.Config, id string) string {
	if !c.Output.Web.Enabled {
		return ""
	}
	return strings.TrimRight(c.Output.Web.Endpoint, "/") + "/a/" + id
}

var revisionMetadata = regexp.MustCompile(`(?m)^\| Revision \| \d+ \|$`)

func appendRevision(path string, f Finding, record manifestArticle) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read existing article: %w", err)
	}
	revision := record.Revision + 1
	content := revisionMetadata.ReplaceAllString(string(b), fmt.Sprintf("| Revision | %d |", revision))
	var add strings.Builder
	fmt.Fprintf(&add, "\n# Revision %d\n\n", revision)
	if len(f.Markdown) > 0 {
		add.WriteString(strings.Join(f.Markdown, "\n"))
		add.WriteString("\n")
	}
	if f.Quote != "" {
		add.WriteString("\n## Documented code\n\n```\n")
		add.WriteString(f.Quote)
		add.WriteString("\n```\n")
	}
	if len(f.Markdown) == 0 && f.Quote == "" {
		add.WriteString("Revision recorded from source annotation.\n")
	}
	return os.WriteFile(path, []byte(strings.TrimRight(content, "\n")+"\n"+add.String()), 0o644)
}
