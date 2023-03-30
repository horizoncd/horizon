package code

// Git struct about git
type Git struct {
	URL       string `json:"url"`
	Subfolder string `json:"subfolder"`
	Branch    string `json:"branch"`
	Tag       string `json:"tag"`
	Commit    string `json:"commit"`
}
