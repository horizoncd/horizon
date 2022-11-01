package harbor

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/pkg/cluster/registry"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"github.com/hashicorp/go-retryablehttp"

	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

const kind = "harbor"

// default params
const (
	_backoffDuration = 1 * time.Second
	_retry           = 3
	_timeout         = 4 * time.Second
)

func init() {
	registry.Register(kind, NewHarborRegistry)
}

// Registry implement Registry
type Registry struct {
	// harbor server address
	server string
	// harbor token
	token string
	// path prefix
	path string
	// http client
	client *http.Client
	// retryableClient retryable client
	retryableClient *retryablehttp.Client
}

func NewHarborRegistry(config *registry.Config) (registry.Registry, error) {
	transport := http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.InsecureSkipVerify,
		},
	}
	harborRegistry := &Registry{
		server: config.Server,
		token:  config.Token,
		path:   config.Path,
		client: &http.Client{
			Transport: &transport,
		},
		retryableClient: &retryablehttp.Client{
			HTTPClient: &http.Client{
				Transport: &transport,
				Timeout:   _timeout,
			},
			RetryMax:   _retry,
			CheckRetry: retryablehttp.DefaultRetryPolicy,
			Backoff: func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
				// wait for this duration if failed
				return _backoffDuration
			},
		},
	}

	return harborRegistry, nil
}

// for test only
func (h *Registry) createProject(ctx context.Context, project string) (_ int, err error) {
	const op = "registry: create project"
	defer wlog.Start(ctx, op).StopPrint()

	url := fmt.Sprintf("%s/api/v2.0/projects", h.server)
	body := map[string]interface{}{
		"project_name": project,
		"metadata": map[string]string{
			"public": "false",
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return -1, perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}
	resp, err := h.sendHTTPRequest(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes), false, "createProject")
	if err != nil {
		return -1, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusCreated {
		// 创建成功
		location := resp.Header.Get("Location")
		projectIDStr := location[strings.LastIndex(location, "/")+1:]
		projectID, err := strconv.Atoi(projectIDStr)
		if err != nil {
			return -1, perror.Wrap(herrors.ErrParamInvalid, err.Error())
		}
		return projectID, nil
	} else if resp.StatusCode == http.StatusConflict {
		// 已经存在
		return -1, nil
	}

	message := common.Response(ctx, resp)
	return -1, errors.E(op, resp.StatusCode, message)
}

func (h *Registry) DeleteRepository(ctx context.Context, repository string) (err error) {
	const op = "registry: delete repository"
	defer wlog.Start(ctx, op).StopPrint()

	link := path.Join("/api/v2.0/projects", h.path, "repositories", repository)
	link = fmt.Sprintf("%s%s", strings.TrimSuffix(h.server, "/"), link)

	resp, err := h.sendHTTPRequest(ctx, http.MethodDelete, link, nil, true, "deleteRepository")
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return perror.Wrap(herrors.ErrHTTPRespNotAsExpected, common.Response(ctx, resp))
}

func (h *Registry) sendHTTPRequest(ctx context.Context, method string,
	url string, body io.Reader, retry bool, operation string) (*http.Response, error) {
	begin := time.Now()
	var rsp *http.Response
	var err error
	defer func() {
		if rsp != nil {
			duration := time.Since(begin)
			server := strings.TrimPrefix(strings.TrimPrefix(h.server, "http://"), "https://")
			statuscode := strconv.Itoa(rsp.StatusCode)
			observe(server, method, statuscode, operation, duration)
		}
	}()
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrHTTPRequestFailed, err.Error())
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", h.token))
	if !retry {
		rsp, err = h.client.Do(req)
		if err != nil {
			return nil, perror.Wrap(herrors.ErrHTTPRequestFailed, err.Error())
		}
		return rsp, nil
	}
	r, err := retryablehttp.FromRequest(req)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrHTTPRequestFailed, err.Error())
	}
	rsp, err = h.retryableClient.Do(r)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrHTTPRequestFailed, err.Error())
	}
	return rsp, nil
}
