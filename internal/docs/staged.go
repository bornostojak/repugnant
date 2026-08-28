package docs

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// StagedFiles returns the set of paths (relative to root, forward-slashed) that
// have staged additions or modifications in git. Deletions are excluded because
// a removed file has nothing left to document. Paths are made relative to root
// via git's --relative, which also scopes the result to root's subtree so it
// lines up with the paths GenerateWith walks.
func StagedFiles(root string) (map[string]bool, error) {
	cmd := exec.Command("git", "-C", root, "diff", "--cached", "--relative", "--name-only", "--diff-filter=ACMR", "-z")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list staged files: %w", err)
	}
	set := map[string]bool{}
	for _, name := range strings.Split(string(out), "\x00") {
		if name != "" {
			set[filepath.ToSlash(name)] = true
		}
	}
	return set, nil
}
