package docs

type Article struct {
	ID, Title, Path          string
	Tags, Category, Markdown []string
	Quote                    string
	Revision                 int
}

type Finding struct {
	Kind, ID, Title string
	Start, End      int
	Markdown        []string
	Quote           string
}
