package docindex

// Section represents a documentation section from HTML
type Section struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Keywords []string `json:"keywords"`
	Preview  string   `json:"preview"`
	URL      string   `json:"url"`
}

// FileIndex contains all sections from a single HTML file
type FileIndex struct {
	Sections []Section `json:"sections"`
}

// Index maps HTML filenames to their section data
type Index map[string]FileIndex
