package code

const (
	GitRefTypeBranch = "branch"
	GitRefTypeTag    = "tag"
	GitRefTypeCommit = "commit"
)

// Git struct about git.
type Git struct {
	URL       string `json:"url"`
	Subfolder string `json:"subfolder"`
	Branch    string `json:"branch,omitempty"`
	Tag       string `json:"tag,omitempty"`
	Commit    string `json:"commit,omitempty"`
}

type TemplateInfo struct {
	Name    string `json:"name"`
	Release string `json:"release"`
}

func NewGit(url, subfolder, refType, ref string) *Git {
	g := &Git{
		URL:       url,
		Subfolder: subfolder,
	}
	switch refType {
	case GitRefTypeCommit:
		g.Commit = ref
	case GitRefTypeTag:
		g.Tag = ref
	case GitRefTypeBranch:
		g.Branch = ref
	}
	return g
}

func (g *Git) RefType() (refType string) {
	if g.Commit != "" {
		refType = GitRefTypeCommit
	} else if g.Tag != "" {
		refType = GitRefTypeTag
	} else if g.Branch != "" {
		refType = GitRefTypeBranch
	}
	return refType
}

func (g *Git) Ref() (ref string) {
	if g.Commit != "" {
		ref = g.Commit
	} else if g.Tag != "" {
		ref = g.Tag
	} else if g.Branch != "" {
		ref = g.Branch
	}
	return ref
}
