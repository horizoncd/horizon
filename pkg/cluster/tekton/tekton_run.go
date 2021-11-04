package tekton

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"github.com/tektoncd/cli/pkg/options"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"

	"g.hz.netease.com/horizon/pkg/cluster/tekton/log"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

const (
	labelKeyPrefix      = "cloudnative.music.netease.com/"
	labelKeyApplication = labelKeyPrefix + "application"
	labelKeyCluster     = labelKeyPrefix + "cluster"
)

type PipelineRunList []v1beta1.PipelineRun

func (p PipelineRunList) Len() int { return len(p) }

func (p PipelineRunList) Less(i, j int) bool {
	return !p[i].CreationTimestamp.Before(&p[j].CreationTimestamp)
}

func (p PipelineRunList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (t *Tekton) GetLatestPipelineRun(ctx context.Context,
	application, cluster string) (pr *v1beta1.PipelineRun, err error) {
	const op = "tekton: get latest pipelineRun"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	pipelineRuns, err := t.client.Tekton.TektonV1beta1().PipelineRuns(t.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s,%s=%s",
			labelKeyApplication, application, labelKeyCluster, cluster),
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	if len(pipelineRuns.Items) == 0 {
		return nil, nil
	}
	sort.Sort(PipelineRunList(pipelineRuns.Items))
	return &pipelineRuns.Items[0], nil
}

func (t *Tekton) GetRunningPipelineRun(ctx context.Context,
	application, cluster string) (pr *v1beta1.PipelineRun, err error) {
	const op = "tekton: get running pipelineRun"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	pipelineRuns, err := t.client.Tekton.TektonV1beta1().PipelineRuns(t.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s,%s=%s",
			labelKeyApplication, application, labelKeyCluster, cluster),
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	for _, pipelineRun := range pipelineRuns.Items {
		if t.getPipelineRunCondition(&pipelineRun) == string(v1beta1.PipelineRunReasonRunning) {
			return &pipelineRun, nil
		}
	}

	return nil, nil
}

func (t *Tekton) CreatePipelineRun(ctx context.Context, pr *PipelineRun) (eventID string, err error) {
	const op = "tekton: create pipelineRun"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	bodyBytes, err := json.Marshal(pr)
	if err != nil {
		return "", errors.E(op, err)
	}

	resp, err := t.sendHTTPRequest(ctx, http.MethodPost, t.server, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", errors.E(op, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		message := wlog.Response(ctx, resp)
		return "", errors.E(op, resp.StatusCode, message)
	}
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.E(op, err)
	}
	var pipelineRunResp struct {
		EventID string `json:"eventID"`
	}
	err = json.Unmarshal(respData, &pipelineRunResp)
	if err != nil {
		return "", errors.E(op, err)
	}

	return pipelineRunResp.EventID, nil
}

type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func (t *Tekton) StopPipelineRun(ctx context.Context, application, cluster string) (err error) {
	const op = "tekton: stop pipelineRun"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	pr, err := t.GetRunningPipelineRun(ctx, application, cluster)
	if err != nil {
		return errors.E(op, err)
	}
	if pr == nil {
		// 如果没有处于Running状态的PipelineRun，则直接返回
		return nil
	}

	// 这个判断参考tekton/cli的源代码：https://github.com/tektoncd/cli/blob/master/pkg/cmd/pipelinerun/cancel.go#L69
	if len(pr.Status.Conditions) > 0 {
		if pr.Status.Conditions[0].Status != corev1.ConditionUnknown {
			return errors.E(op, fmt.Errorf(
				"failed to cancel PipelineRun %s: PipelineRun has already finished execution", pr.Name))
		}
	}

	payload := []patchStringValue{{
		Op:    "replace",
		Path:  "/spec/status",
		Value: v1beta1.PipelineRunSpecStatusCancelled,
	}}

	data, err := json.Marshal(payload)
	if err != nil {
		return errors.E(op, err)
	}
	if _, err := t.client.Tekton.TektonV1beta1().PipelineRuns(pr.Namespace).Patch(ctx, pr.Name,
		types.JSONPatchType, data, metav1.PatchOptions{}); err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (t *Tekton) GetLatestPipelineRunLog(ctx context.Context, application,
	cluster string) (logChan <-chan log.Log, errChan <-chan error, err error) {
	const op = "tekton: get latest pipelineRun log"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	var pr *v1beta1.PipelineRun
	// 返回最近一次执行的PipelineRun的日志
	pr, err = t.GetLatestPipelineRun(ctx, application, cluster)
	if err != nil {
		return nil, nil, errors.E(op, err)
	}
	if pr == nil {
		return nil, nil, errors.E(op, http.StatusNotFound, fmt.Errorf("no pipelineRun exists for %s", cluster))
	}

	return t.GetPipelineRunLog(ctx, pr)
}

func (t *Tekton) GetPipelineRunLog(ctx context.Context, pr *v1beta1.PipelineRun) (<-chan log.Log, <-chan error, error) {
	const op = "tekton: get pipelineRun log"
	condition := t.getPipelineRunCondition(pr)
	logOps := &options.LogOptions{
		Params:          log.NewTektonParams(t.client.Dynamic, t.client.Kube, t.client.Tekton, t.namespace),
		PipelineRunName: pr.Name,
	}
	// 如果不失败的情况下，只返回我们自己的task/step的日志，不返回tekton的一些init阶段的日志
	if condition != string(v1beta1.PipelineRunReasonFailed) {
		logOps.Tasks = strings.Split(t.filteredTasks, ",")
		logOps.Steps = strings.Split(t.filteredSteps, ",")
	}

	lr, err := log.NewReader(log.LogTypePipeline, logOps)
	if err != nil {
		return nil, nil, errors.E(op, err)
	}
	return lr.Read()
}

func (t *Tekton) DeletePipelineRun(ctx context.Context, pr *v1beta1.PipelineRun) error {
	const op = "tekton: deletePipelineRun"
	if pr == nil {
		return nil
	}
	err := t.client.Tekton.TektonV1beta1().PipelineRuns(t.namespace).
		Delete(ctx, pr.Name, metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return errors.E(op, http.StatusNotFound, err)
		}
		return errors.E(op, err)
	}

	return nil
}

func (t *Tekton) getPipelineRunCondition(pr *v1beta1.PipelineRun) string {
	for _, cond := range pr.Status.Conditions {
		if cond.Type == apis.ConditionSucceeded {
			return cond.Reason
		}
	}
	return ""
}

func (t *Tekton) sendHTTPRequest(ctx context.Context, method string,
	url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	// 添加X-Request-Id header，供tekton trigger使用, TODO(gjq) add requestID
	req.Header.Set("X-Request-Id", "")
	client := &http.Client{}
	return client.Do(req)
}
