package harbor

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"sync"
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
	HarborUploadPath   = "/api/chartrepo/%s/charts"
	HarborDeletePath   = "/api/chartrepo/%s/charts/%s/%s"
	HarborStatPath     = "/api/chartrepo/%s/charts/%s/%s"
	HarborDownloadPath = "/chartrepo/%s/%s"

	defaultCacheLifeTime = 1000 * time.Second
	cacheKeyFormat       = "%s-%s"
)

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

type TemplateRepo struct {
	url      *url.URL
	username string
	password string
	repoName string
	client   *http.Client
	cache    *cache
}

func NewTemplateRepo(config config.Repo) (*TemplateRepo, error) {
	host, err := url.Parse(config.Host)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("url is incorrect: %v", err))
	}

	transport := &http.Transport{}

	tlsConf, err := tlsutil.NewClientTLS(config.CertFile, config.KeyFile, config.CAFile)
	if err != nil {
		return nil, perror.Wrap(herrors.NewErrCreateFailed(herrors.TLS, err.Error()),
			"failed to create TLS: %v")
	}
	tlsConf.InsecureSkipVerify = config.Insecure

	transport.TLSClientConfig = tlsConf

	client := &http.Client{}
	client.Transport = transport

	return &TemplateRepo{
		repoName: config.RepoName,
		url:      host,
		username: config.Username,
		password: config.Password,
		client:   client,
		cache:    &cache{},
	}, nil
}

func (h *TemplateRepo) GetLoc() string {
	return fmt.Sprintf("%s://%s/chartrepo/%s", h.url.Scheme, h.url.Host, h.repoName)
}

func (h *TemplateRepo) UploadChart(chart *chart.Chart) error {
	var (
		err error
		req *http.Request
	)
	defer func() {
		if err == nil {
			key := fmt.Sprintf(cacheKeyFormat, chart.Metadata.Name, chart.Metadata.Version)
			h.cache.setWithExpiration(key, chart, defaultCacheLifeTime)
		}
	}()

	h.url.Path = fmt.Sprintf(HarborUploadPath, h.repoName)
	req, err = http.NewRequest("POST", h.url.String(), nil)
	if err != nil {
		return perror.Wrap(herrors.ErrHTTPRequestFailed,
			fmt.Sprintf("failed to create request: %v", err))
	}
	h.url.RawQuery = "force"

	if err = writeChartToBody(req, chart); err != nil {
		return err
	}
	req.SetBasicAuth(h.username, h.password)

	resp, err := h.client.Do(req)
	if err != nil {
		return perror.Wrap(herrors.ErrHTTPRequestFailed,
			fmt.Sprintf("failed to create send request: %v", err))
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

func (h *TemplateRepo) DeleteChart(name string, version string) error {
	var (
		err error
		req *http.Request
	)
	defer func() {
		if err == nil {
			key := fmt.Sprintf(cacheKeyFormat, name, version)
			h.cache.remove(key)
		}
	}()
	h.url.Path = fmt.Sprintf(HarborDeletePath, h.repoName, name, version)
	req, err = http.NewRequest("DELETE", h.url.String(), nil)
	if err != nil {
		return perror.Wrap(herrors.ErrHTTPRequestFailed,
			fmt.Sprintf("failed to create request: %v", err))
	}

	req.SetBasicAuth(h.username, h.password)

	resp, err := h.client.Do(req)
	if err != nil {
		return perror.Wrap(herrors.ErrHTTPRequestFailed,
			fmt.Sprintf("failed to create send request: %v", err))
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

func (h *TemplateRepo) ExistChart(name string, version string) (bool, error) {
	_, err := h.statChart(name, version)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (h *TemplateRepo) GetChart(name string, version string) (*chart.Chart, error) {
	key := fmt.Sprintf(cacheKeyFormat, name, version)
	value, getOK := h.cache.get(key)
	if c, ok := value.(*chart.Chart); getOK && ok {
		return c, nil
	}
	var (
		err          error
		chartPackage *chart.Chart
	)

	defer func() {
		if err == nil {
			key := fmt.Sprintf(cacheKeyFormat, name, version)
			h.cache.setWithExpiration(key, chartPackage, defaultCacheLifeTime)
		}
	}()

	meta, err := h.statChart(name, version)
	if err != nil {
		return nil, err
	}

	if len(meta.Urls) < 1 {
		return nil, perror.Wrap(herrors.NewErrGetFailed(herrors.HarborChartURL,
			"chart url is empty"), "chart url is empty")
	}

	h.url.Path = fmt.Sprintf(HarborDownloadPath, h.repoName, meta.Urls[0])
	req, err := http.NewRequest("GET", h.url.String(), nil)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrHTTPRequestFailed,
			fmt.Sprintf("failed to create request: %v", err))
	}

	req.SetBasicAuth(h.username, h.password)

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrHTTPRequestFailed,
			fmt.Sprintf("failed to create send request: %v", err))
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	b, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		if err != nil {
			return nil, perror.Wrap(herrors.ErrReadFailed,
				fmt.Sprintf("failed to read response: %v", err))
		}
		return nil, perror.Wrap(herrors.ErrHTTPRespNotAsExpected,
			fmt.Sprintf("%s: %s", resp.Status, string(b)))
	}
	chartPackage, err = loader.LoadArchive(bytes.NewReader(b))
	if err != nil {
		return nil, perror.Wrap(herrors.ErrLoadChartArchive,
			fmt.Sprintf("failed to load archive: %v", err))
	}
	return chartPackage, nil
}

func (h *TemplateRepo) statChart(name string, version string) (*Metadata, error) {
	h.url.Path = fmt.Sprintf(HarborStatPath, h.repoName, name, version)
	req, err := http.NewRequest("GET", h.url.String(), nil)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrHTTPRequestFailed,
			fmt.Sprintf("failed to create request: %v", err))
	}

	req.SetBasicAuth(h.username, h.password)

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrHTTPRequestFailed,
			fmt.Sprintf("failed to create send request: %v", err))
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

func writeChartToBody(req *http.Request, c *chart.Chart) error {
	var buf bytes.Buffer
	bodyWriter := multipart.NewWriter(&buf)
	chartWriter, err := bodyWriter.CreateFormFile("chart",
		fmt.Sprintf("%s-%s", c.Metadata.Name, c.Metadata.Version))
	if err != nil {
		return perror.Wrap(herrors.ErrHTTPRequestFailed,
			fmt.Sprintf("failed to create multipart writer: %v", err))
	}

	gzipper := gzip.NewWriter(chartWriter)
	twriter := tar.NewWriter(gzipper)
	defer func() {
		_ = twriter.Close()
		_ = gzipper.Close()
		_ = bodyWriter.Close()
	}()
	err = templaterepo.WriteTarContents(twriter, c, "")
	if err != nil {
		return perror.Wrap(herrors.ErrWriteFailed,
			fmt.Sprintf("writing chart to buffer failed: %s", err.Error()))
	}
	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())
	req.Body = ioutil.NopCloser(&buf)
	return nil
}

type cache struct {
	m sync.Map
}

func (c *cache) get(k interface{}) (interface{}, bool) {
	return c.m.Load(k)
}

func (c *cache) set(k, v interface{}) {
	c.m.Store(k, v)
}

func (c *cache) remove(k interface{}) {
	c.m.Delete(k)
}

func (c *cache) setWithExpiration(k, v interface{}, duration time.Duration) {
	c.set(k, v)
	time.AfterFunc(duration, func() {
		c.remove(k)
	})
}
