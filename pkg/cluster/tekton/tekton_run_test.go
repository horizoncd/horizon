package tekton

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/pkg/config/tekton"
	perror "g.hz.netease.com/horizon/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	fakedtekton "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakeddynamic "k8s.io/client-go/dynamic/fake"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

var tektonConfig = &tekton.Tekton{
	Server:    "",
	Namespace: "tekton",
}

// nolint
func TestTekton_StopPipelineRun(t1 *testing.T) {
	pr1 := &v1beta1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test1-1-1",
			Namespace: "tekton",
			Labels: map[string]string{
				common.TektonTriggersEventIDKey: "1",
			},
		},
		Status: v1beta1.PipelineRunStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{
					{
						Type:   apis.ConditionSucceeded,
						Status: corev1.ConditionTrue,
						Reason: string(v1beta1.PipelineRunReasonSuccessful),
					},
				},
			},
		},
	}
	pr2 := &v1beta1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test2-2-1",
			Namespace: "tekton",
			Labels: map[string]string{
				common.TektonTriggersEventIDKey: "2",
			},
		},
		Status: v1beta1.PipelineRunStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{
					{
						Type:   apis.ConditionSucceeded,
						Status: corev1.ConditionUnknown,
						Reason: string(v1beta1.PipelineRunReasonRunning),
					},
				},
			},
		},
	}
	type fields struct {
		TektonConfig *tekton.Tekton
	}
	type args struct {
		ctx       context.Context
		ciEventID string
	}
	tests := []struct {
		name         string
		fields       fields
		pipelineRuns []runtime.Object
		args         args
		wantErr      bool
	}{
		{
			name: "stop pipelineRun normal",
			fields: fields{
				TektonConfig: tektonConfig,
			},
			pipelineRuns: []runtime.Object{pr1},
			args: args{
				ctx: context.Background(),
				ciEventID: "1",
			},
			wantErr: false,
		},
		{
			name: "stop pipelineRun normal",
			fields: fields{
				TektonConfig: tektonConfig,
			},
			pipelineRuns: []runtime.Object{pr2},
			args: args{
				ctx: context.Background(),
				ciEventID: "2",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Tekton{
				namespace: tektonConfig.Namespace,
				client: &Client{
					Tekton:  fakedtekton.NewSimpleClientset(tt.pipelineRuns...),
					Dynamic: fakeddynamic.NewSimpleDynamicClient(runtime.NewScheme()),
				},
			}
			if err := t.StopPipelineRun(tt.args.ctx, tt.args.ciEventID); (err != nil) != tt.wantErr {
				t1.Errorf("StopPipelineRun() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type MockTektonController struct {
	R *mux.Router
}

func errResponse(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(err.Error()))
}

func NewMockTektonController() *MockTektonController {
	r := mux.NewRouter()
	c := &MockTektonController{
		R: r,
	}
	r.Methods(http.MethodPost).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			errResponse(err, w)
			return
		}
		var pr PipelineRun
		if err := json.Unmarshal(data, &pr); err != nil {
			errResponse(err, w)
			return
		}
		if pr.Application == "app" {
			resp := map[string]string{
				"eventID": "1234",
			}
			b, err := json.Marshal(resp)
			if err != nil {
				errResponse(err, w)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write(b)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	return c
}

func TestMain(m *testing.M) {
	c := NewMockTektonController()
	s := httptest.NewServer(http.HandlerFunc(c.R.ServeHTTP))
	tektonConfig.Server = s.Listener.Addr().String()
	os.Exit(m.Run())
}

func TestTekton_CreatePipelineRun(t1 *testing.T) {
	type fields struct {
		server    string
		namespace string
		client    *Client
	}
	type args struct {
		ctx context.Context
		pr  *PipelineRun
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantEventID string
		wantErr     bool
	}{
		{
			name: "create pipeline run",
			fields: fields{
				server: "http://" + tektonConfig.Server,
			},
			args: args{
				ctx: context.Background(),
				pr: &PipelineRun{
					Application: "app",
				},
			},
			wantErr:     false,
			wantEventID: "1234",
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Tekton{
				server:    tt.fields.server,
				namespace: tt.fields.namespace,
				client:    tt.fields.client,
			}
			gotEventID, err := t.CreatePipelineRun(tt.args.ctx, tt.args.pr)
			if (err != nil) != tt.wantErr {
				t1.Errorf("CreatePipelineRun() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotEventID != tt.wantEventID {
				t1.Errorf("CreatePipelineRun() gotEventID = %v, want %v", gotEventID, tt.wantEventID)
			}
		})
	}
}

// nolint
func TestTekton_getPipelineRunByID(t1 *testing.T) {
	var pr1, pr2, pr3 v1beta1.PipelineRun
	pr1.Name = "pr1"
	pr1.Labels = map[string]string{common.TektonTriggersEventIDKey: "1"}
	pr2.Name = "pr2"
	pr2.Labels = map[string]string{common.TektonTriggersEventIDKey: "1"}
	pr3.Name = "pr3"
	pr3.Labels = map[string]string{common.TektonTriggersEventIDKey: "2"}

	t := &Tekton{
		client: &Client{
			Tekton:  fakedtekton.NewSimpleClientset(&pr1, &pr2, &pr3),
			Dynamic: fakeddynamic.NewSimpleDynamicClient(runtime.NewScheme()),
		},
	}
	pr, err := t.getPipelineRunByID(context.Background(), "10")
	e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound)
	assert.True(t1, ok)
	assert.Equal(t1, e.Source, herrors.PipelinerunInTekton)

	pr, err = t.getPipelineRunByID(context.Background(), "2")
	assert.Nil(t1, err)
	assert.Equal(t1, pr3.Name, pr.Name)

	pr, err = t.getPipelineRunByID(context.Background(), "1")
	assert.Equal(t1, herrors.ErrTektonInternal, perror.Cause(err))
}
