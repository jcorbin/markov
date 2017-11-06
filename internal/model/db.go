package model

// DocDB represents a database of extracted documents. It contains a markov
// language for generated a plausible document title, and an inverted index to
// map back to supporting document info from such a title.
type DocDB struct {
	Docs      map[string]DocInfo  `json:"docs"`
	TitleLang Lang                `json:"titleLang"`
	InvTW     map[string][]string `json:"invertedTitleWords"`
}

// DocInfo contains meta data for a documnet in a DocDB.
type DocInfo struct {
	SourceFile string            `json:"sourceFile"`
	TransFile  string            `json:"transFile"`
	Title      string            `json:"title"`
	Info       map[string]string `json:"info"`
}

// Doc represents an extracted document loaded from a DocInfo in a DocDB.
type Doc struct {
	Title string            `json:"title"`
	Info  map[string]string `json:"info"`
	Lang  Lang              `json:"language"`
}
