package  authorization

type Role struct {
	Name  string
	PolicyRules []PolicyRule
}

type PolicyRule struct {
	Verbs 	  		[]string
	APIGroups 		[]string
	Resources 		[]string
	Scopes    		[]string
	NonResourceURLs []string
}