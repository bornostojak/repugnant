package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultConfig = `# rePugnant project configuration
# Add an explicit list (for example: [go, typescript]) to override language auto-detection.
version: 1
langs: []

output:
  # Commit this directory when local documentation is the official output.
  docs:
    enabled: true
    dir: docs
  # Web publishing is staged locally and sent by the pre-push hook.
  web:
    enabled: false
    # endpoint: http://127.0.0.1:8080

hooks:
  # block is the safe non-interactive default; allow_pending lets a push proceed after failure.
  on_publish_failure: block
`

const preCommitHook = `#!/bin/sh
# Installed by rePugnant. An existing hook is preserved as .rpg-backup.
if [ -x "$0.rpg-backup" ]; then
  "$0.rpg-backup" "$@" || exit $?
fi
rpg hook pre-commit || exit $?
`

const prePushHook = `#!/bin/sh
# Installed by rePugnant. An existing hook is preserved as .rpg-backup.
if [ -x "$0.rpg-backup" ]; then
  "$0.rpg-backup" "$@" || exit $?
fi
exec rpg hook pre-push
`

// Init prepares a repository for rpg. HookOnly leaves configuration and .rpg untouched.
func Init(root string, hookOnly bool) error {
	if err := ensureGitRepository(root); err != nil {
		return err
	}
	if !hookOnly {
		if err := writeConfig(root); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Join(root, ".rpg"), 0o755); err != nil {
			return fmt.Errorf("create .rpg: %w", err)
		}
		if err := ensureGitignore(root); err != nil {
			return err
		}
	}
	return installHooks(root)
}

func ensureGitRepository(root string) error {
	info, err := os.Stat(filepath.Join(root, ".git"))
	if err != nil || !info.IsDir() {
		return fmt.Errorf("%s is not a Git worktree; run rpg init inside a repository", root)
	}
	return nil
}

func writeConfig(root string) error {
	path := filepath.Join(root, ConfigFileName)
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect %s: %w", ConfigFileName, err)
	}
	if err := os.WriteFile(path, []byte(defaultConfig), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", ConfigFileName, err)
	}
	return nil
}

func ensureGitignore(root string) error {
	path := filepath.Join(root, ".gitignore")
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read .gitignore: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == ".rpg/" {
			return nil
		}
	}
	if len(data) > 0 && !strings.HasSuffix(string(data), "\n") {
		data = append(data, '\n')
	}
	data = append(data, []byte("# rePugnant hook staging\n.rpg/\n")...)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("update .gitignore: %w", err)
	}
	return nil
}

func installHooks(root string) error {
	hooks := map[string]string{"pre-commit": preCommitHook, "pre-push": prePushHook}
	for name, contents := range hooks {
		path := filepath.Join(root, ".git", "hooks", name)
		if err := writeHook(path, contents); err != nil {
			return err
		}
	}
	return nil
}

func writeHook(path, contents string) error {
	existing, err := os.ReadFile(path)
	if err == nil && string(existing) != contents {
		backup := path + ".rpg-backup"
		if _, backupErr := os.Stat(backup); errors.Is(backupErr, os.ErrNotExist) {
			if err := os.WriteFile(backup, existing, 0o755); err != nil {
				return fmt.Errorf("back up existing hook %s: %w", filepath.Base(path), err)
			}
		}
	}
	if err := os.WriteFile(path, []byte(contents), 0o755); err != nil {
		return fmt.Errorf("write hook %s: %w", filepath.Base(path), err)
	}
	return nil
}
