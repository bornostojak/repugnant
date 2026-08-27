package docs

import (
	"strings"
	"testing"
	"time"
)

func TestRender(t *testing.T) {
	got := Render(Article{ID: "s9Aa3A3al", Title: "Cache", Revision: 1, Markdown: []string{"Body"}}, "A", "a@x", time.Unix(0, 0).UTC())
	if !strings.Contains(got, "# Cache") || !strings.Contains(got, "| Web | — |") {
		t.Fatal(got)
	}
}
