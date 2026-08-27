# Test strategy

The following high-value cases must be automated where applicable and exercised manually for end-to-end features.

1. `rpg init` is idempotent and creates a valid commented config.
2. Init safely adds `.rpg/` to an existing `.gitignore` without duplication.
3. Hooks install and repair without overwriting unrelated user hook behavior.
4. Auto-detection recognizes each supported language manifest.
5. Explicit `langs` overrides auto-detection.
6. Unsupported/invalid config gives a precise diagnostic.
7. Standalone annotations produce Markdown with correct title/body/metadata.
8. Quote annotations exclude rpg marker lines and preserve safe indentation.
9. Heredoc content and terminators are retained exactly.
10. YAML and indentation-sensitive examples are not invalidated by quote formatting.
11. Missing or mismatched `!rPg` fails with file and line details.
12. Opaque IDs are URL-safe, unique, and stable after a title change.
13. Revisions append in order and preserve earlier content/title metadata.
14. A historical API URL returns the requested revision; latest returns current.
15. Changed quoted code without a revision annotation blocks pre-commit clearly.
16. Generated local docs are staged only when configured for local output.
17. Manual local doc edits survive a subsequent appended revision.
18. Pending web payloads survive failed publishing and are retried later.
19. Interactive pre-push continuation retains pending/pre-approved data.
20. Noninteractive publish failure blocks by default and obeys configured override.
21. SQLite persistence survives a container restart.
22. PostgreSQL implementation passes the same storage contract tests.
23. API rejects malformed revision payloads without partial writes.
24. API search covers title, body, and tags.
25. Organization tree mutations preserve article identity and revision history.
26. Source browser returns only documented files and safe source ranges.
27. Backend logs include useful correlation/context but no secrets/source dumps.
28. Web UI keyboard navigation and focus states work.
29. Web UI remains usable at common mobile and desktop Chromium viewports.
30. Revision history, search, category tree, and source browser work together end-to-end.
