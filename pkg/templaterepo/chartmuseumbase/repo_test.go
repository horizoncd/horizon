package chartmuseumbase

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	config "g.hz.netease.com/horizon/pkg/config/templaterepo"
	"g.hz.netease.com/horizon/pkg/templaterepo"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"
)

const (
	EnvTemplateRepos = "TEMPLATE_REPOS"
)

type RepoConfig struct {
	Kind              string `json:"kind"`
	Host              string `json:"host"`
	Passwd            string `json:"passwd"`
	RepoName          string `json:"repoName"`
	Username          string `json:"username"`
	TemplateName      string `json:"templateName"`
	TemplateRepo      string `json:"templateRepo"`
	TemplateRepoToken string `json:"templateRepoToken"`
	TemplateTag       string `json:"templateTag"`
}

var repoConfig *RepoConfig

func Test(t *testing.T) {
	templateRepos := os.Getenv(EnvTemplateRepos)
	configs := make([]RepoConfig, 0)

	if err := json.Unmarshal([]byte(templateRepos), &configs); err != nil {
		panic(err)
	}

	for _, cfg := range configs {
		repoConfig = &cfg

		t.Run(fmt.Sprintf("TestRepo_%s", repoConfig.Kind), TestRepo)
	}
}

func createRepo(t *testing.T) templaterepo.TemplateRepo {
	repo, err := NewRepo(config.Repo{
		Kind:     repoConfig.Kind,
		Host:     repoConfig.Host,
		Username: repoConfig.Username,
		Password: repoConfig.Passwd,
		Insecure: true,
		CertFile: "",
		KeyFile:  "",
		CAFile:   "",
		RepoName: repoConfig.RepoName,
	})
	assert.Nil(t, err)

	return repo
}

func TestRepo(t *testing.T) {
	repo := createRepo(t)

	name := "test"
	data := []byte("hello, world")
	c := &chart.Chart{Metadata: &chart.Metadata{}, Files: []*chart.File{{Name: name, Data: data}}}
	c.Metadata.Name = repoConfig.TemplateName
	c.Metadata.Version = repoConfig.TemplateTag

	err := repo.UploadChart(c)
	assert.Nil(t, err)

	tm := time.Now()
	c, err = repo.GetChart(repoConfig.TemplateName, repoConfig.TemplateTag, tm)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	// use cache
	c, err = repo.GetChart(repoConfig.TemplateName, repoConfig.TemplateTag, tm)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	res, err := repo.ExistChart(repoConfig.TemplateName, repoConfig.TemplateTag)
	assert.Nil(t, err)
	assert.Equal(t, true, res)

	err = repo.DeleteChart(repoConfig.TemplateName, repoConfig.TemplateTag)
	assert.Nil(t, err)

	_, err = repo.GetChart(repoConfig.TemplateRepo, repoConfig.TemplateTag, time.Now())
	assert.NotNil(t, err)
}
