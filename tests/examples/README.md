# Example projects

Each directory is a small, intentionally different source fixture for a supported language. Integration testing copies every fixture to a temporary directory, initializes it as a Git worktree with `rpg init`, generates local Markdown through its pre-commit hook, and commits the resulting source-reference and `/docs` changes. The fixture sources remain versioned here; temporary `.git` directories never are.
