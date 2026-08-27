package docs

import (
	"fmt"
	"strings"
	"time"
)

func Render(a Article, name, email string, at time.Time, webURLs ...string) string {
	webURL := ""
	if len(webURLs) > 0 {
		webURL = webURLs[0]
	}
	web := "—"
	if webURL != "" {
		web = "[Open article](" + webURL + ")"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n| Metadata | Info |\n| :- | :- |\n| By | %s |\n| Email | %s |\n| Generated on | %s |\n| Revision | %d |\n| Web | %s |\n", a.Title, name, email, at.Format(time.RFC3339), a.Revision, web)
	if a.Category != "" {
		fmt.Fprintf(&b, "| Category | %s |\n", a.Category)
	}
	if len(a.Tags) > 0 {
		fmt.Fprintf(&b, "| Tags | %s |\n", strings.Join(a.Tags, ", "))
	}
	fmt.Fprintf(&b, "\n# Revision %d\n\n", a.Revision)
	b.WriteString(strings.Join(a.Markdown, "\n"))
	b.WriteString("\n")
	if a.Quote != "" {
		b.WriteString("\n## Documented code\n\n```\n")
		b.WriteString(a.Quote)
		b.WriteString("\n```\n")
	}
	return b.String()
}
