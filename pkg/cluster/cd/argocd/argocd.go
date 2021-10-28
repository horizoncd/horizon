package argocd

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/log"
	"g.hz.netease.com/horizon/pkg/util/wlog"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/hashicorp/go-retryablehttp"
	corev1 "k8s.io/api/core/v1"
)

type (
	// ArgoCD interact with ArgoCD Server
	ArgoCD interface {
		// CreateApplication create an application in argoCD
		CreateApplication(ctx context.Context, manifest []byte) error

		// DeployApplication deploy an application in argoCD
		DeployApplication(ctx context.Context, application string, revision string) error

		// DeleteApplication delete an application in argoCD
		DeleteApplication(ctx context.Context, application string) error

		// WaitApplication 等待应用同步完成。在删除应用集群时，需要先删除Argo Application，再删除gitlab repo，否则，
		// Argo Application 永远无法删除。
		// 参见：https://argoproj.github.io/argo-cd/faq/#ive-deletedcorrupted-my-repo-and-cant-delete-my-app
		// status == 200: 表示期望应用同步成功
		// status == 404: 表示期望应用删除成功
		WaitApplication(ctx context.Context, application string, uid string, status int) error

		// GetApplication get an application in argoCD
		GetApplication(ctx context.Context, application string) (*v1alpha1.Application, error)

		// RefreshApplication ...
		RefreshApplication(ctx context.Context, application string) (app *v1alpha1.Application, err error)

		// GetApplicationTree get resource-tree of an application in argoCD
		GetApplicationTree(ctx context.Context, application string) (*v1alpha1.ApplicationTree, error)

		// GetApplicationResource get a resource under an application in argoCD
		GetApplicationResource(ctx context.Context, application string, param ResourceParams) (string, error)

		// ListResourceEvents get resource's events of an application in argoCD
		ListResourceEvents(ctx context.Context, application string, param EventParam) (*corev1.EventList, error)

		// ResumeRollout ...
		ResumeRollout(ctx context.Context, application string) error

		// GetContainerLog get standard output of container of an application in argoCD
		GetContainerLog(ctx context.Context, application string,
			param ContainerLogParams) (<-chan ContainerLog, <-chan error, error)
	}

	// EventParam the params for ListResourceEvents
	EventParam struct {
		ResourceNamespace string `json:"resourceNamespace"`
		ResourceUID       string `json:"resourceUID"`
		ResourceName      string `json:"resourceName"`
	}

	// ResourceParams the params for GetApplicationResource
	ResourceParams struct {
		// Group name in k8s, for example, Deployment resource is in 'apps' group
		Group string `json:"group,omitempty"`
		// Version in k8s, for example, Deployment resource has a 'v1' version
		Version string `json:"version,omitempty"`
		// the Kind of resource in k8s, for example, the kind of Deployment resource is 'Deployment'
		Kind string `json:"kind,omitempty"`
		// the namespace of a resource in k8s
		Namespace string `json:"namespace,omitempty"`
		// the resource name
		ResourceName string `json:"resourceName,omitempty"`
	}

	// ErrorResponse 由于无法复用 argocd 的 internal.StreamError, 故新建结构体
	ErrorResponse struct {
		StreamError struct {
			GrpcCode   int    `json:"grpc_code"`
			HTTPCode   int    `json:"http_code"`
			Message    string `json:"message"`
			HTTPStatus string `json:"http_status"`
		} `json:"error"`
	}

	// ContainerLogParams the params for GetContainerLog
	ContainerLogParams struct {
		Namespace     string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
		PodName       string `json:"podName,omitempty" yaml:"podName,omitempty"`
		ContainerName string `json:"containerName,omitempty" yaml:"containerName,omitempty"`
		TailLines     int    `json:"tailLines,omitempty" yaml:"tailLines,omitempty"`
	}

	ContainerLog struct {
		Result struct {
			Content   string `json:"content,omitempty" yaml:"content,omitempty"`
			Timestamp string `json:"timestamp,omitempty" yaml:"timestamp,omitempty"`
		} `json:"result"`
	}
)

type (
	// argo holding the info for ArgoCD Server
	helper struct {
		// URL for argoCD server
		URL string `json:"url"`
		// Token the token to be used for argoCD server
		Token string `json:"token"`
	}

	Hook struct{}

	Strategy struct {
		Hook Hook `json:"hook"`
	}

	DeployApplicationRequest struct {
		Revision string   `json:"revision"`
		Prune    bool     `json:"prune"`
		DryRun   bool     `json:"dryRun"`
		Strategy Strategy `json:"strategy"`
	}
)

func NewArgoCD(URL, token string) ArgoCD {
	return &helper{URL: URL, Token: token}
}

var _ ArgoCD = (*helper)(nil)

const (
	// http retry count
	_retry = 3
	// http timeout
	_timeout = 10 * time.Second
	// retry backoff duration
	_backoff = 1 * time.Second
)

var (
	_client = &retryablehttp.Client{
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Timeout: _timeout,
		},
		RetryMax:   _retry,
		CheckRetry: retryablehttp.DefaultRetryPolicy,
		Backoff: func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
			// 每次失败，等待 _backoff 时间后，进行下次重试
			return _backoff
		},
	}
)

func (h *helper) CreateApplication(ctx context.Context, manifest []byte) (err error) {
	const op = "argo: create application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	url := h.URL + "/api/v1/applications?validate=false&upsert=false"
	resp, err := h.sendHTTPRequest(ctx, http.MethodPost, url, bytes.NewReader(manifest))
	if err != nil {
		return errors.E(op, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		message := wlog.Response(ctx, resp)
		return errors.E(op, resp.StatusCode, message)
	}

	return nil
}

func (h *helper) DeployApplication(ctx context.Context, application string, revision string) (err error) {
	const op = "argo: deploy application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	url := fmt.Sprintf("%v/api/v1/applications/%v/sync", h.URL, application)
	req := DeployApplicationRequest{
		Revision: revision,
		Prune:    true,
		DryRun:   false,
		Strategy: Strategy{Hook: Hook{}},
	}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return errors.E(op, err)
	}
	resp, err := h.sendHTTPRequest(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return errors.E(op, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		message := wlog.Response(ctx, resp)
		return errors.E(op, resp.StatusCode, message)
	}

	return nil
}

func (h *helper) DeleteApplication(ctx context.Context, application string) (err error) {
	const op = "argo: delete application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	url := fmt.Sprintf("%v/api/v1/applications/%v?cascade=true", h.URL, application)
	resp, err := h.sendHTTPRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return errors.E(op, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		message := wlog.Response(ctx, resp)
		return errors.E(op, resp.StatusCode, message)
	}

	return nil
}

func (h *helper) WaitApplication(ctx context.Context, cluster string, uid string, status int) (err error) {
	const op = "argo: wait application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	const WaitErrCode = 1000
	waitError := errors.E(op, WaitErrCode, "continue to wait")

	waitFunc := func(i int) error {
		ctx, cancel := context.WithTimeout(ctx, time.Second*2)
		defer cancel()

		log.Infof(ctx, "wait for cluster<%v> to be status of %v, count=%v", cluster, status, i+1)
		applicationCR, err := h.RefreshApplication(ctx, cluster)
		if err != nil && stderrors.Is(err, context.DeadlineExceeded) {
			return waitError
		}

		// 如果使用switch，则需要两层break
		if err == nil {
			if uid != "" && uid != string(applicationCR.UID) {
				return errors.E(op, "the cluster has been recreated with the same name")
			}
			if status == http.StatusOK && applicationCR.Status.Sync.Status == v1alpha1.SyncStatusCodeSynced {
				return nil
			}
		} else if errors.Status(err) == http.StatusNotFound {
			if status == http.StatusNotFound {
				return nil
			}
		} else {
			return errors.E(op, err)
		}

		return waitError
	}

	for i := 0; i < 200; i++ {
		err := waitFunc(i)
		if err == nil {
			return nil
		}
		if errors.Status(err) != WaitErrCode {
			return err
		}
		time.Sleep(time.Second)
	}

	return errors.E(op, "time out")
}

func (h *helper) GetApplication(ctx context.Context,
	application string) (applicationCRD *v1alpha1.Application, err error) {
	const op = "argo: get application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	url := fmt.Sprintf("%v/api/v1/applications/%v", h.URL, application)
	return h.getOrRefreshApplication(ctx, url)
}

func (h *helper) RefreshApplication(ctx context.Context,
	application string) (applicationCRD *v1alpha1.Application, err error) {
	const op = "argo: refresh application "
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	url := fmt.Sprintf("%v/api/v1/applications/%v?refresh=normal", h.URL, application)
	return h.getOrRefreshApplication(ctx, url)
}

// getOrRefreshApplication 具体是 get 还是 refresh 操作，由调用者的 url 中的 parameter 决定
func (h *helper) getOrRefreshApplication(ctx context.Context,
	url string) (applicationCRD *v1alpha1.Application, err error) {
	const op = "argo: getApplication"
	resp, err := h.sendHTTPRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		var message string
		if resp.StatusCode != http.StatusNotFound {
			message = wlog.Response(ctx, resp)
		}
		return nil, errors.E(op, resp.StatusCode, message)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if err := json.Unmarshal(data, &applicationCRD); err != nil {
		return nil, errors.E(op, err)
	}
	return applicationCRD, nil
}

func (h *helper) GetApplicationTree(ctx context.Context, application string) (
	tree *v1alpha1.ApplicationTree, err error) {
	const op = "argo: get application tree"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	url := fmt.Sprintf("%v/api/v1/applications/%v/resource-tree", h.URL, application)
	resp, err := h.sendHTTPRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		message := wlog.Response(ctx, resp)
		return nil, errors.E(op, resp.StatusCode, message)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if err = json.Unmarshal(data, &tree); err != nil {
		return nil, errors.E(op, err)
	}

	return tree, nil
}

func (h *helper) GetApplicationResource(ctx context.Context, application string, gvk ResourceParams) (
	output string, err error) {
	const op = "argo: get application resource"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	url := fmt.Sprintf("%v/api/v1/applications/%v/resource?namespace=%v&resourceName=%v&group=%v&version=%v&kind=%v",
		h.URL, application, gvk.Namespace, gvk.ResourceName, gvk.Group, gvk.Version, gvk.Kind)
	resp, err := h.sendHTTPRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", errors.E(op, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		message := wlog.Response(ctx, resp)
		return "", errors.E(op, resp.StatusCode, message)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.E(op, err)
	}

	m := struct{ Manifest string }{}
	if err = json.Unmarshal(data, &m); err != nil {
		return "", errors.E(op, err)
	}

	if m.Manifest == "{}" {
		return "", errors.E(op, http.StatusNotFound, "manifest is empty")
	}

	return m.Manifest, nil
}

func (h *helper) ListResourceEvents(ctx context.Context, application string, param EventParam) (
	eventList *corev1.EventList, err error) {
	const op = "argo: list resource events"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	url := fmt.Sprintf("%v/api/v1/applications/%v/events?resourceUID=%v&resourceNamespace=%v&resourceName=%v",
		h.URL, application, param.ResourceUID, param.ResourceNamespace, param.ResourceName)
	resp, err := h.sendHTTPRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		message := wlog.Response(ctx, resp)
		return nil, errors.E(op, resp.StatusCode, message)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if err := json.Unmarshal(data, &eventList); err != nil {
		return nil, errors.E(op, err)
	}

	return eventList, nil
}

func (h *helper) ResumeRollout(ctx context.Context, application string) (err error) {
	const op = "argo: resume rollout"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	app, err := h.GetApplication(ctx, application)
	if err != nil {
		return errors.E(op, err)
	}
	rolloutVersion := "v1alpha1"
	rolloutGroup := "argoproj.io"
	namespace := app.Spec.Destination.Namespace
	// rollout名字和所属application的名字一致
	format := "%v/api/v1/applications/%v/resource/actions?namespace=%v&resourceName=%v&version=%s&kind=Rollout&group=%s"
	url := fmt.Sprintf(format, h.URL, application, namespace, application, rolloutVersion, rolloutGroup)
	requestBodyStr := `"resume"`
	resp, err := h.sendHTTPRequest(ctx, http.MethodPost, url, bytes.NewReader([]byte(requestBodyStr)))
	if err != nil {
		return errors.E(op, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		message := wlog.Response(ctx, resp)
		return errors.E(op, resp.StatusCode, message)
	}
	return nil
}

func (h *helper) GetContainerLog(ctx context.Context, application string,
	param ContainerLogParams) (lc <-chan ContainerLog, ec <-chan error, err error) {
	const op = "argo: get container log"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	format := "%v/api/v1/applications/%v/pods/%v/logs?container=%v&follow=false&namespace=%v&tailLines=%v"
	url := fmt.Sprintf(format, h.URL, application, param.PodName, param.ContainerName, param.Namespace, param.TailLines)
	resp, err := h.sendHTTPRequest(ctx, http.MethodGet, url, nil) // nolint:bodyclose
	if err != nil {
		return nil, nil, errors.E(op, err)
	}
	// NOTE: 不需要在此进行close操作，否则会引起 read on closed response body 错误

	if resp.StatusCode != http.StatusOK {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, err
		}
		_ = resp.Body.Close()
		wlog.ResponseContent(ctx, data)

		var errorResponse *ErrorResponse
		err = json.Unmarshal(data, &errorResponse)
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, errors.E(op, resp.StatusCode, errorResponse.StreamError.Message)
	}

	logC := make(chan ContainerLog)
	errC := make(chan error)

	go func() {
		defer close(logC)
		defer close(errC)
		defer func() { _ = resp.Body.Close() }()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			var containerLog ContainerLog
			if err := json.Unmarshal(scanner.Bytes(), &containerLog); err != nil {
				errC <- err
				return
			}
			logC <- containerLog
		}
		if err := scanner.Err(); err != nil {
			errC <- err
			return
		}
	}()

	return logC, errC, nil
}

func (h *helper) sendHTTPRequest(ctx context.Context, method string, url string,
	body io.Reader) (*http.Response, error) {
	log.Infof(ctx, "method: %v, url: %v", method, url)
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", h.Token))
	req.Header.Add("Content-Type", "application/json")

	r, err := retryablehttp.FromRequest(req)
	if err != nil {
		return nil, err
	}
	return _client.Do(r)
}
