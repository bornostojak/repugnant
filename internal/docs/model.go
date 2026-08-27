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
	Start, End                int
	Markdown, Tags            []string
	Quote                     string
}
