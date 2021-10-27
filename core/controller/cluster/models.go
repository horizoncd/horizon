package cluster

type Base struct {
	Description string `json:"description"`
	Git         *Git   `json:"git"`
}

// Git struct about git
type Git struct {
	Branch string `json:"branch"`
}
