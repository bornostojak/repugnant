package docs

type Article struct {
	ID, Title, Path string
	Tags, Markdown  []string
	Category        string
	Quote           string
	Revision        int
}

type Finding struct {
	Kind, ID, Title, Category string
	// Subtitle is a revision's optional heading, from a "$# ..." line (inline
	// revision) or the text after "$rPg@{id}: ...". It labels Revision N in the
	// article without changing the article's permanent title.
	Subtitle string
	// Start is the marker line; End is the closing !rPg line for quoted
	// findings (otherwise equal to Start). BodyStart is the first line after
	// the marker and its $~/?~/$#/~ continuation — i.e. where quoted code
	// begins. Generation uses [Start, BodyStart) to delete the marker header
	// and prose when rewriting to the clean tracked form.
	Start, End, BodyStart int
	Markdown, Tags        []string
	Quote                 string
}
