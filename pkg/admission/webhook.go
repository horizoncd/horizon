package admission

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/admission/models"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/util/log"
)

const DefaultTimeout = 5 * time.Second

var (
	validatingWebhooks []Webhook
)

func Register(kind models.Kind, webhook Webhook) {
	switch kind {
	case models.KindValidating:
		validatingWebhooks = append(validatingWebhooks, webhook)
	}
}

func Clear() {
	validatingWebhooks = nil
}

type validateResult struct {
	err  error
	resp *Response
}

type Request struct {
	Operation    models.Operation       `json:"operation"`
	Resource     string                 `json:"resource"`
	ResourceName string                 `json:"resourceName"`
	SubResource  string                 `json:"subResource"`
	Version      string                 `json:"version"`
	Object       interface{}            `json:"object"`
	OldObject    interface{}            `json:"oldObject"`
	Options      map[string]interface{} `json:"options,omitempty"`
}

type Response struct {
	Allowed   *bool  `json:"allowed"`
	Result    string `json:"result,omitempty"`
	Patch     []byte `json:"patch,omitempty"`
	PatchType string `json:"patchType,omitempty"`
}

type Webhook interface {
	Handle(context.Context, *Request) (*Response, error)
	IgnoreError() bool
	Interest(*Request) bool
}

func Validating(ctx context.Context, request *Request) error {
	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()
	finishedCount := 0
	resCh := make(chan validateResult)
	for _, webhook := range validatingWebhooks {
		go func(webhook Webhook) {
			defer runtime.HandleCrash()
			if !webhook.Interest(request) {
				resCh <- validateResult{nil, nil}
				return
			}
			response, err := webhook.Handle(ctx, request)
			if err != nil {
				if webhook.IgnoreError() {
					log.Errorf(ctx, "failed to admit request: %v", err)
					resCh <- validateResult{nil, nil}
					return
				}
				resCh <- validateResult{err, nil}
				return
			}
			if response == nil || response.Allowed == nil {
				if webhook.IgnoreError() {
					log.Errorf(ctx, "failed to admit request: response is nil or allowed is nil")
					resCh <- validateResult{nil, nil}
					return
				}
				resCh <- validateResult{perror.New("response is nil or allowed is nil"), nil}
				return
			}
			resCh <- validateResult{nil, response}
		}(webhook)
	}

	for res := range resCh {
		finishedCount++
		if res.err != nil {
			return res.err
		}
		if res.resp != nil && res.resp.Allowed != nil && !*res.resp.Allowed {
			log.Infof(ctx, "request denied by webhook: %s", res.resp.Result)
			return perror.Wrapf(herrors.ErrForbidden, "request denied by webhook: %s", res.resp.Result)
		}
		if finishedCount >= len(validatingWebhooks) {
			close(resCh)
			break
		}
	}

	return nil
}
