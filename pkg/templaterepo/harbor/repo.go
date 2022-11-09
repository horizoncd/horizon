package harbor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"time"

	herrors "g.hz.netease.com/horizon/core/errors"
	config "g.hz.netease.com/horizon/pkg/config/templaterepo"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/templaterepo"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/helm/pkg/tlsutil"
)

const (
	kindHarbor      = "harbor"
	kindChartMuseum = "chartmuseum"
)

func init() {
	templaterepo.Register(kindHarbor, NewRepo)
	templaterepo.Register(kindChartMuseum, NewRepo)
}

type Metadata struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	APIVersion  string    `json:"apiVersion"`
	Type        string    `json:"type"`
	Urls        []string  `json:"urls"`
	Created     time.Time `json:"created"`
	Digest      string    `json:"digest"`
}

type Stat struct {
	Metadata Metadata `json:"metadata"`
}

type Repo struct {
	prefix   string
	host     *url.URL
	token    string
	username string
	password string
	repoName string
	client   *http.Client
}

func NewRepo(config config.Repo) (templaterepo.TemplateRepo, error) {
	host, err := url.Parse(config.Host)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("url is incorrect: %v", err))
	}

	tlsConf, err := tlsutil.NewClientTLS(config.CertFile, config.KeyFile, config.CAFile)
	if err != nil {
		return nil, perror.Wrap(herrors.NewErrCreateFailed(herrors.TLS, err.Error()),
			"failed to create TLS: %v")
	}
	tlsConf.InsecureSkipVerify = config.Insecure

	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
			TLSClientConfig:    tlsConf,
		},
	}
	prefix := ""
	if config.Kind == kindHarbor {
		prefix = fmt.Sprintf("chartrepo/%s", config.RepoName)
	}

	return &Repo{
		repoName: config.RepoName,
		prefix:   prefix,
		host:     host,
		username: config.Username,
		password: config.Password,
		token:    config.Token,
		client:   client,
	}, nil
}

func (h *Repo) GetLoc() string {
	return fmt.Sprintf("%s://%s/chartrepo/%s", h.host.Scheme, h.host.Host, url.PathEscape(h.repoName))
}

func (h *Repo) UploadChart(chartPkg *chart.Chart) error {
	var buf bytes.Buffer
	bodyWriter := multipart.NewWriter(&buf)
	chartWriter, err := bodyWriter.CreateFormFile("chart",
		fmt.Sprintf("%s-%s", chartPkg.Metadata.Name, chartPkg.Metadata.Version))
	if err != nil {
		return perror.Wrap(herrors.ErrHTTPRequestFailed,
			fmt.Sprintf("failed to create multipart writer: %v", err))
	}

	err = templaterepo.ChartSerialize(chartPkg, chartWriter)
	if err != nil {
		return err
	}

	contentType := bodyWriter.FormDataContentType()
	err = bodyWriter.Close()
	if err != nil {
		return perror.Wrap(herrors.ErrHTTPRequestFailed,
			fmt.Sprintf("failed to create multipart writer: %v", err))
	}

	resp, err := h.do(http.MethodPost, h.uploadLink(),
		ioutil.NopCloser(&buf), http.Header{"Content-Type": []string{contentType}})
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 201 && resp.StatusCode != 202 {
		var b []byte
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return perror.Wrap(herrors.ErrReadFailed,
				fmt.Sprintf("failed to read response: %v", err))
		}
		return perror.Wrap(herrors.ErrHTTPRespNotAsExpected,
			fmt.Sprintf("%s: %s", resp.Status, string(b)))
	}
	return nil
}

func (h *Repo) DeleteChart(name string, version string) error {
	resp, err := h.do(http.MethodDelete, h.deleteLink(name, version), nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		var b []byte
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return perror.Wrap(herrors.ErrReadFailed,
				fmt.Sprintf("failed to read response: %v", err))
		}
		return perror.Wrap(herrors.ErrHTTPRespNotAsExpected,
			fmt.Sprintf("%s: %s", resp.Status, string(b)))
	}
	return nil
}

func (h *Repo) ExistChart(name string, version string) (bool, error) {
	_, err := h.statChart(name, version)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (h *Repo) GetChart(name string, version string, lastSyncAt time.Time) (*chart.Chart, error) {
	resp, err := h.do(http.MethodGet, h.downloadLink(name, version), nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	b, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		if err != nil {
			return nil, perror.Wrap(herrors.ErrReadFailed,
				fmt.Sprintf("failed to read response: %v", err))
		}
		if resp.StatusCode == http.StatusNotFound {
			return nil, perror.Wrap(herrors.NewErrNotFound(herrors.TemplateReleaseInRepo,
				fmt.Sprintf("%s: %s", resp.Status, string(b))),
				"not found")
		}
		return nil, perror.Wrap(herrors.ErrHTTPRespNotAsExpected,
			fmt.Sprintf("%s: %s", resp.Status, string(b)))
	}
	chartPackage, err := loader.LoadArchive(bytes.NewReader(b))
	if err != nil {
		return nil, perror.Wrap(herrors.ErrLoadChartArchive,
			fmt.Sprintf("failed to load archive: %v", err))
	}
	return chartPackage, nil
}

func (h *Repo) statChart(name string, version string) (*Metadata, error) {
	resp, err := h.do(http.MethodGet, h.statLink(name, version), nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var stat Stat

	b, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		if err != nil {
			return nil, perror.Wrap(herrors.ErrReadFailed,
				fmt.Sprintf("failed to read response: %v", err))
		}
		return nil, perror.Wrap(herrors.ErrHTTPRespNotAsExpected,
			fmt.Sprintf("%s: %s, chart name = %s version = %s", resp.Status, string(b), name, version))
	}
	err = json.Unmarshal(b, &stat)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("could not unmarshal stat: %v", err))
	}
	return &stat.Metadata, nil
}

func (h *Repo) do(method, url string, body io.Reader, headers ...http.Header) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrHTTPRequestFailed,
			fmt.Sprintf("failed to create request: %v", err))
	}

	for _, header := range headers {
		for k, values := range header {
			for _, v := range values {
				req.Header.Add(k, v)
			}
		}
	}

	req.SetBasicAuth(h.username, h.password)

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrHTTPRequestFailed,
			fmt.Sprintf("failed to create send request: %v", err))
	}

	return resp, nil
}

func (h *Repo) linkWithSchemeAndHost() string {
	return fmt.Sprintf("%s://%s", h.host.Scheme, h.host.Host)
}

func (h *Repo) uploadLink() string {
	return fmt.Sprintf("%s/%s?force",
		h.linkWithSchemeAndHost(), path.Join("api", h.prefix, "charts"))
}

func (h *Repo) deleteLink(name, version string) string {
	return fmt.Sprintf("%s/%s",
		h.linkWithSchemeAndHost(), path.Join("api", h.prefix, "charts",
			url.PathEscape(name), url.PathEscape(version)))
}

func (h *Repo) statLink(name, version string) string {
	return fmt.Sprintf("%s/%s",
		h.linkWithSchemeAndHost(), path.Join("api", h.prefix, "charts",
			url.PathEscape(name), url.PathEscape(version)))
}

func (h *Repo) downloadLink(name, version string) string {
	return fmt.Sprintf("%s/%s/%s-%s.tgz",
		h.linkWithSchemeAndHost(), path.Join(h.prefix, "charts"),
		url.PathEscape(name), url.PathEscape(version))
}
