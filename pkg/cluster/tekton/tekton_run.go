package tekton

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/horizoncd/horizon/core/common"
	"k8s.io/apimachinery/pkg/fields"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/log"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/log/wlog"
	"github.com/tektoncd/cli/pkg/options"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (t *Tekton) GetPipelineRunByID(ctx context.Context, ciEventID string) (_ *v1beta1.PipelineRun, err error) {
	return t.getPipelineRunByID(ctx, ciEventID)
}

func (t *Tekton) CreatePipelineRun(ctx context.Context, pr *PipelineRun) (eventID string, err error) {
	const op = "tekton: create pipelineRun"
	defer wlog.Start(ctx, op).StopPrint()

	bodyBytes, err := json.Marshal(pr)
	if err != nil {
		return "", perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}

	resp, err := t.sendHTTPRequest(ctx, http.MethodPost, t.server, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", perror.Wrap(herrors.ErrHTTPRequestFailed, err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	// From Tekton Trigger v0.15.0, the EventListener will always respond with a 202 Accepted response code.
	// Please refer: https://github.com/tektoncd/triggers/releases/tag/v0.15.0.
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		message := common.Response(ctx, resp)
		return "", perror.Wrapf(herrors.ErrHTTPRespNotAsExpected, "statusCode = %d, message = %s", resp.StatusCode, message)
	}
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", perror.Wrap(herrors.ErrReadFailed, err.Error())
	}
	var pipelineRunResp struct {
		EventID string `json:"eventID"`
	}
	err = json.Unmarshal(respData, &pipelineRunResp)
	if err != nil {
		return "", perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}

	return pipelineRunResp.EventID, nil
}

type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func (t *Tekton) StopPipelineRun(ctx context.Context, ciEventID string) (err error) {
	const op = "tekton: stop pipelineRun"
	defer wlog.Start(ctx, op).StopPrint()

	pr, err := t.GetPipelineRunByID(ctx, ciEventID)
	if err != nil {
		return err
	}
	if pr == nil {
		return nil
	}

	// This judgment refers to the source code of tekton/cli：
	// https://github.com/tektoncd/cli/blob/master/pkg/cmd/pipelinerun/cancel.go#L69
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
		return perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}
	if _, err := t.client.Tekton.TektonV1beta1().PipelineRuns(pr.Namespace).Patch(ctx, pr.Name,
		types.JSONPatchType, data, metav1.PatchOptions{}); err != nil {
		return herrors.NewErrUpdateFailed(herrors.Pipelinerun, err.Error())
	}
	return nil
}

func (t *Tekton) GetPipelineRunLogByID(ctx context.Context, ciEventID string) (
	_ <-chan log.Log, _ <-chan error, err error) {
	pr, err := t.getPipelineRunByID(ctx, ciEventID)

	if err != nil {
		return nil, nil, perror.WithMessage(err, "failed to get pipeline run with labels")
	}

	return t.GetPipelineRunLog(ctx, pr)
}

func (t *Tekton) GetPipelineRunLog(ctx context.Context, pr *v1beta1.PipelineRun) (<-chan log.Log, <-chan error, error) {
	logOps := &options.LogOptions{
		Params:          log.NewTektonParams(t.client.Dynamic, t.client.Kube, t.client.Tekton, t.namespace),
		PipelineRunName: pr.Name,
	}

	lr, err := log.NewReader(log.LogTypePipeline, logOps)
	if err != nil {
		return nil, nil, err
	}
	return lr.Read()
}

func (t *Tekton) DeletePipelineRun(ctx context.Context, pr *v1beta1.PipelineRun) error {
	if pr == nil {
		return nil
	}
	err := t.client.Tekton.TektonV1beta1().PipelineRuns(t.namespace).
		Delete(ctx, pr.Name, metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return herrors.NewErrNotFound(herrors.Pipelinerun, err.Error())
		}
		return herrors.NewErrDeleteFailed(herrors.Pipelinerun, err.Error())
	}

	return nil
}

func (t *Tekton) getPipelineRunByID(ctx context.Context, ciEventID string) (_ *v1beta1.PipelineRun,
	err error) {
	selector := fields.ParseSelectorOrDie(fmt.Sprintf("%v=%v", common.TektonTriggersEventIDKey, ciEventID))
	prs, err := t.client.Tekton.TektonV1beta1().PipelineRuns(t.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, perror.Wrap(herrors.ErrTektonInternal, err.Error())
	}
	if prs == nil || len(prs.Items) == 0 {
		return nil, herrors.NewErrNotFound(herrors.PipelinerunInTekton,
			fmt.Sprintf("failed to list pipeline with selector %s", selector.String()))
	} else if len(prs.Items) > 1 {
		return nil, perror.Wrap(herrors.ErrTektonInternal,
			fmt.Sprintf("unexpected: selector %s match multi pipeline runs", selector.String()))
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
	req.Header.Set("X-Request-Id", "")
	client := &http.Client{}
	return client.Do(req)
}
