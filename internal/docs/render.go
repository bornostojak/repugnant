package docs

import (
	"fmt"
	"strings"
	"time"
)

func Render(a Article, name, email string, at time.Time) string {
	web := ""
	if a.ID != "" {
		web = "`/d/" + a.ID + "`"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n| Metadata | Info |\n| :- | :- |\n| By | %s |\n| Email | %s |\n| Generated on | %s |\n| Revision | %d |\n| Web | %s |\n\n# Revision %d\n\n", a.Title, name, email, at.Format(time.RFC3339), a.Revision, web, a.Revision)
	b.WriteString(strings.Join(a.Markdown, "\n"))
	b.WriteString("\n")
	if a.Quote != "" {
		b.WriteString("\n## Documented code\n\n```\n")
		b.WriteString(a.Quote)
		b.WriteString("\n```\n")
	}
	return b.String()
}
