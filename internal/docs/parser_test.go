package docs

import "testing"

func TestParseQuoteAndUnindent(t *testing.T) {
	src := "func f() {\n  // ?rPg: Cache\n  // ?~ Why this works\n  // quoted comment\n  if ok {\n    work()\n  }\n  // !rPg\n}"
	got, err := Parse("f.go", src)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Kind != "quote" || got[0].Quote != "// quoted comment\nif ok {\n  work()\n}" {
		t.Fatalf("%+v", got)
	}
}
func TestParseRejectsUnclosedQuote(t *testing.T) {
	_, err := Parse("f.py", "# ?rPg: no\nx=1")
	if err == nil {
		t.Fatal("wanted error")
	}
}
