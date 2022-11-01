package harbor

import (
	"os"
	"strings"
	"testing"
	"time"

	config "g.hz.netease.com/horizon/pkg/config/templaterepo"
	"g.hz.netease.com/horizon/pkg/templaterepo"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"
)

const (
	EnvHarborHost     = "HARBOR_HOST"
	EnvHarborUser     = "HARBOR_USER"
	EnvHarborPasswd   = "HARBOR_PASSWD"
	EnvHarborRepoName = "HARBOR_REPO_NAME"
)

var (
	harborHost     string
	harborAdmin    string
	harborPasswd   string
	harborRepoName string

	templateName = "test"
	releaseName  = "v1.0.0"
)

func TestMain(m *testing.M) {
	harborHost = os.Getenv(EnvHarborHost)
	harborHost = strings.TrimPrefix(harborHost, "https://")
	harborHost = strings.TrimPrefix(harborHost, "http://")
	harborAdmin = os.Getenv(EnvHarborUser)
	harborPasswd = os.Getenv(EnvHarborPasswd)
	harborRepoName = os.Getenv(EnvHarborRepoName)

	os.Exit(m.Run())
}

func checkSkip(t *testing.T) {
	if harborHost == "" ||
		harborAdmin == "" ||
		harborPasswd == "" ||
		harborRepoName == "" {
		t.Skip()
	}
}

func createHarbor(t *testing.T) templaterepo.TemplateRepo {
	repo, err := NewRepo(config.Repo{
		Host:     harborHost,
		Scheme:   "https",
		Username: harborAdmin,
		Password: harborPasswd,
		Insecure: true,
		CertFile: "",
		KeyFile:  "",
		CAFile:   "",
		RepoName: harborRepoName,
	})
	assert.Nil(t, err)

	return repo
}

func TestHarbor(t *testing.T) {
	checkSkip(t)
	harbor := createHarbor(t)

	name := "test"
	data := []byte("hello, world")
	c := &chart.Chart{Metadata: &chart.Metadata{}, Files: []*chart.File{{Name: name, Data: data}}}
	c.Metadata.Name = templateName
	c.Metadata.Version = releaseName

	err := harbor.UploadChart(c)
	assert.Nil(t, err)

	tm := time.Now()
	c, err = harbor.GetChart(templateName, releaseName, tm)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	// use cache
	c, err = harbor.GetChart(templateName, releaseName, tm)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	res, err := harbor.ExistChart(templateName, releaseName)
	assert.Nil(t, err)
	assert.Equal(t, true, res)

	err = harbor.DeleteChart(templateName, releaseName)
	assert.Nil(t, err)

	_, err = harbor.GetChart(templateName, releaseName, time.Now())
	assert.NotNil(t, err)
}
