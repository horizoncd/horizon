package argocd

type Mapper map[string]*ArgoCD

type ArgoCD struct {
	URL      string `yaml:"url"`
	Token    string `yaml:"token"`
	HelmRepo string `yaml:"helmRepo"`
}
