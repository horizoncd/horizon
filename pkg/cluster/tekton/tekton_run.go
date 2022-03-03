package tekton

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/fields"
	"net/http"

	clustercommon "g.hz.netease.com/horizon/pkg/cluster/common"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/log"
	perrors "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
	"github.com/tektoncd/cli/pkg/options"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	ErrPipelineRunNotFound = perrors.New("tekton: pipeline run not found")
	ErrTektonInternal      = perrors.New("tekton: internal error")
)

func (t *Tekton) GetPipelineRunByID(ctx context.Context, cluster string,
	clusterID, pipelinerunID uint) (_ *v1beta1.PipelineRun, err error) {
	return t.getPipelineRunByID(ctx, clusterID, pipelinerunID)
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

func (t *Tekton) StopPipelineRun(ctx context.Context, cluster string, clusterID, pipelinerunID uint) (err error) {
	const op = "tekton: stop pipelineRun"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	pr, err := t.GetPipelineRunByID(ctx, cluster, clusterID, pipelinerunID)
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
			// PipelineRun has already finished execution
			return nil
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

func (t *Tekton) GetPipelineRunLogByID(ctx context.Context,
	cluster string, clusterID, pipelinerunID uint) (_ <-chan log.Log, _ <-chan error, err error) {
	pr, err := t.getPipelineRunByID(ctx, clusterID, pipelinerunID)

	if err != nil {
		return nil, nil, perrors.WithMessage(err, "failed to get pipeline run with labels")
	}

	return t.GetPipelineRunLog(ctx, pr)
}

func (t *Tekton) GetPipelineRunLog(ctx context.Context, pr *v1beta1.PipelineRun) (<-chan log.Log, <-chan error, error) {
	const op = "tekton: get pipelineRun log"
	logOps := &options.LogOptions{
		Params:          log.NewTektonParams(t.client.Dynamic, t.client.Kube, t.client.Tekton, t.namespace),
		PipelineRunName: pr.Name,
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

func (t *Tekton) getPipelineRunByID(ctx context.Context, clusterID, pipelinerunID uint) (_ *v1beta1.PipelineRun,
	err error) {
	selector := fields.ParseSelectorOrDie(fmt.Sprintf("%v=%v,%v=%v",
		clustercommon.PipelinerunIDLabelKey, pipelinerunID, clustercommon.ClusterIDLabelKey, clusterID))
	prs, err := t.client.Tekton.TektonV1beta1().PipelineRuns(t.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, perrors.WithMessage(ErrTektonInternal, err.Error())
	}
	if prs == nil || len(prs.Items) == 0 {
		return nil, perrors.WithMessagef(ErrPipelineRunNotFound, "failed to list pipeline with selector %s",
			selector.String())
	} else if len(prs.Items) > 1 {
		return nil, perrors.WithMessagef(ErrTektonInternal, "unexpected: selector %s match multi pipeline runs",
			selector.String())
	}
	return &(prs.Items[0]), nil
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
