package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/wlog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	"github.com/gorilla/mux"
	"k8s.io/kubernetes/pkg/apis/apps"
)

// Repository the credentials for ArgoCD to interact with Git Repository.
type Repository struct {
	// Type ssh
	Type string `json:"type"`
	// Name application name
	Name string `json:"name"`
	// Repo git repository url
	Repo string `json:"repo"`
	// SSHPrivateKey
	SSHPrivateKey string `json:"sshPrivateKey"`
}

type ArgoServer struct {
	R            *mux.Router
	Applications map[string]*Application
}

type Application struct {
	repository *Repository
	crd        *v1alpha1.Application
	synced     bool
}

func NewArgoServer() *ArgoServer {
	r := mux.NewRouter()
	c := &ArgoServer{
		R:            r,
		Applications: map[string]*Application{},
	}

	r.UseEncodedPath().Path("/api/v1/repositories/{repository}").Methods(http.MethodGet).
		HandlerFunc(c.GetRepository)
	r.Path("/api/v1/applications").Methods(http.MethodPost).
		HandlerFunc(c.CreateApplication)
	r.Path("/api/v1/applications/{application}/sync").Methods(http.MethodPost).
		HandlerFunc(c.DeployApplication)
	r.Path("/api/v1/applications/{application}").Methods(http.MethodDelete).
		Queries("cascade", "true").HandlerFunc(c.DeleteApplication)
	r.Path("/api/v1/applications/{application}").Methods(http.MethodGet).
		HandlerFunc(c.GetApplication)
	r.Path("/api/v1/applications/{application}/resource-tree").Methods(http.MethodGet).
		HandlerFunc(c.GetApplicationTree)
	r.Path("/api/v1/applications/{application}/resource").Methods(http.MethodGet).
		Queries("namespace", "{namespace}", "resourceName", "{resourceName}",
			"group", "{group}", "version", "{version}", "kind", "{kind}").
		HandlerFunc(c.GetApplicationResource)
	r.Path("/api/v1/applications/{application}/events").Methods(http.MethodGet).
		Queries("resourceUID", "{resourceUID}", "resourceNamespace", "{resourceNamespace}",
			"resourceName", "{resourceName}").
		HandlerFunc(c.ListResourceEvents)
	r.Path("/api/v1/applications/{application}/resource/actions").Methods(http.MethodPost).
		HandlerFunc(c.ResumeRollout)
	r.Path("/api/v1/applications/{application}/pods/{pod}/logs").Methods(http.MethodGet).
		Queries("container", "{container}", "namespace", "{namespace}").
		HandlerFunc(c.GetContainerLog)

	return c
}

func (argoServer *ArgoServer) GetRepository(w http.ResponseWriter, r *http.Request) {
	const op = "argo mock server: get repository "

	var err error
	ctx := log.WithContext(context.Background(), "GetRepository")
	defer wlog.Start(ctx, op).StopPrint()

	vars := mux.Vars(r)
	repository := vars["repository"]
	repository, err = url.QueryUnescape(repository)
	log.Infof(ctx, "repository=%v", repository)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for key := range argoServer.Applications {
		if argoServer.Applications[key].repository.Repo == repository {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`
{
    "repo":"ssh://git@cloudnative.com:22222/music-cloud-native-robot-dev/unit-test-repo.git",
    "connectionState":{
        "status":"Successful",
        "message":"",
        "attemptedAt":"2020-11-17T11:45:14Z"
    },
    "type":"git"
}
`))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`
{
  "repo": "ssh://git@cloudnative.com:22222/music-cloud-native-robot-dev/unit-test-repo.git",
  "connectionState": {
    "status": "Failed",
    "message": "Unable Get \"ssh://git@cloudnative.com:22222/xx.git/index.yaml\": unsupported protocol scheme \"ssh\"",
    "attemptedAt": "2020-11-17T11:45:56Z"
  },
  "type": "git"
}
`))
}

func (argoServer *ArgoServer) CreateApplication(w http.ResponseWriter, r *http.Request) {
	const op = "argo mock server: create application"

	var err error
	ctx := log.WithContext(context.Background(), "CreateApplication")
	defer wlog.Start(ctx, op).StopPrint()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var crd *v1alpha1.Application
	if err := json.Unmarshal(data, &crd); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if argoServer.Applications[crd.Name] == nil {
		argoServer.Applications[crd.Name] = &Application{crd: crd}
	} else {
		application := argoServer.Applications[crd.Name]
		application.crd = crd
	}
	w.WriteHeader(http.StatusOK)
}

func (argoServer *ArgoServer) DeployApplication(w http.ResponseWriter, r *http.Request) {
	const op = "argo mock server: deploy application"

	ctx := log.WithContext(context.Background(), "DeployApplication")
	defer wlog.Start(ctx, op).StopPrint()

	vars := mux.Vars(r)
	application := vars["application"]

	if argoServer.Applications[application] == nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		application := argoServer.Applications[application]
		application.synced = true
	}
	w.WriteHeader(http.StatusOK)
}

func (argoServer *ArgoServer) GetApplication(w http.ResponseWriter, r *http.Request) {
	const op = "argo mock server: get application"

	var err error
	ctx := log.WithContext(context.Background(), "GetApplication")
	defer wlog.Start(ctx, op).StopPrint()

	vars := mux.Vars(r)
	application := vars["application"]

	if argoServer.Applications[application] == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	cache := argoServer.Applications[application]
	if cache.synced {
		cache.crd.Status.Sync.Status = v1alpha1.SyncStatusCodeSynced
		cache.crd.Status.Health.Status = health.HealthStatusProgressing
	} else {
		cache.crd.Status.Sync.Status = v1alpha1.SyncStatusCodeOutOfSync
	}

	data, err := json.Marshal(cache.crd)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (argoServer *ArgoServer) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	const op = "argo mock server: delete application"

	ctx := log.WithContext(context.Background(), "DeleteApplication")
	defer wlog.Start(ctx, op).StopPrint()

	vars := mux.Vars(r)
	application := vars["application"]

	delete(argoServer.Applications, application)
	w.WriteHeader(http.StatusOK)
}

func (argoServer *ArgoServer) GetApplicationTree(w http.ResponseWriter, _ *http.Request) {
	d := []byte(`
{
  "nodes": [
    {
      "group": "apps",
      "version": "v1",
      "kind": "Deployment",
      "namespace": "test-guanggao",
      "name": "unit-test-repo-test-2",
      "uid": "bd743b14-db33-424e-903b-74f138b539c1",
      "info": [
        {
          "name": "Revision",
          "value": "Rev:1"
        }
      ],
      "resourceVersion": "23865",
      "health": {
        "status": "Progressing",
        "message": "Waiting for rollout to finish: 0 of 1 updated replicas are available..."
      },
      "createdAt": "2020-10-30T08:18:23Z"
    },
    {
      "group": "apps",
      "version": "v1",
      "kind": "ReplicaSet",
      "namespace": "test-guanggao",
      "name": "unit-test-repo-test-2-55546456c",
      "uid": "52ebb390-6c09-4481-824e-b33cb333ba4d",
      "parentRefs": [
        {
          "group": "apps",
          "kind": "Deployment",
          "namespace": "test-guanggao",
          "name": "unit-test-repo-test-2",
          "uid": "bd743b14-db33-424e-903b-74f138b539c1"
        }
      ],
      "info": [
        {
          "name": "Revision",
          "value": "Rev:1"
        }
      ],
      "resourceVersion": "23864",
      "health": {
        "status": "Progressing",
        "message": "Waiting for rollout to finish: 0 out of 1 new replicas are available..."
      },
      "createdAt": "2020-10-30T08:18:23Z"
    },
    {
      "version": "v1",
      "kind": "Pod",
      "namespace": "test-guanggao",
      "name": "unit-test-repo-test-2-55546456c-64brq",
      "uid": "27fc76ba-c497-4cc7-8c8a-510a92458081",
      "parentRefs": [
        {
          "group": "apps",
          "kind": "ReplicaSet",
          "namespace": "test-guanggao",
          "name": "unit-test-repo-test-2-55546456c",
          "uid": "52ebb390-6c09-4481-824e-b33cb333ba4d"
        }
      ],
      "info": [
        {
          "name": "Status Reason",
          "value": "ContainerCreating"
        },
        {
          "name": "Containers",
          "value": "0/1"
        }
      ],
      "networkingInfo": {
        "labels": {
          "app": "unit-test-repo-test-2",
          "pod-template-hash": "55546456c"
        }
      },
      "resourceVersion": "23871",
      "images": [
        "hub.c.163.com/commonwork/poc-template:tomcat7_jdk8_2"
      ],
      "health": {
        "status": "Progressing"
      },
      "createdAt": "2020-10-30T08:18:23Z"
    }
  ]
}
`)
	_, _ = w.Write(d)
}

func (argoServer *ArgoServer) GetApplicationResource(w http.ResponseWriter, _ *http.Request) {
	deployment := apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "deployment",
			Namespace:  "test-1",
			UID:        "uid",
			Generation: 1,
		},
	}
	data, _ := json.Marshal(&deployment)
	m := struct{ Manifest string }{}
	m.Manifest = string(data)
	data, _ = json.Marshal(m)
	_, _ = w.Write(data)
}

func (argoServer *ArgoServer) ListResourceEvents(w http.ResponseWriter, _ *http.Request) {
	d := []byte(`
{
  "metadata": {
    "selfLink": "/api/v1/namespaces/test-guanggao/events",
    "resourceVersion": "23182"
  },
  "items": [
    {
      "metadata": {
        "name": "unit-test-repo-test-2-55546456c-m7hlk.1642b6edc63882a4",
        "namespace": "test-guanggao",
        "selfLink": "/api/v1/namespaces/test-guanggao/events/unit-test-repo-test-2-55546456c-m7hlk.1642b6edc63882a4",
        "uid": "5f0f1fc0-1db0-4b65-b192-6169b46f39b4",
        "resourceVersion": "23130",
        "creationTimestamp": "2020-10-30T08:12:29Z"
      },
      "involvedObject": {
        "kind": "Pod",
        "namespace": "test-guanggao",
        "name": "unit-test-repo-test-2-55546456c-m7hlk",
        "uid": "3111e4b3-7d50-4e89-9186-db658f7058aa",
        "apiVersion": "v1",
        "resourceVersion": "23125"
      },
      "reason": "Scheduled",
      "message": "Successfully assigned test-guanggao/unit-test-repo-test-2-55546456c-m7hlk to docker-desktop",
      "source": {
        "component": "default-scheduler"
      },
      "firstTimestamp": "2020-10-30T08:12:29Z",
      "lastTimestamp": "2020-10-30T08:12:29Z",
      "count": 1,
      "type": "Normal",
      "eventTime": null,
      "reportingComponent": "",
      "reportingInstance": ""
    },
    {
      "metadata": {
        "name": "unit-test-repo-test-2-55546456c-m7hlk.1642b6ee2cda37c8",
        "namespace": "test-guanggao",
        "selfLink": "/api/v1/namespaces/test-guanggao/events/unit-test-repo-test-2-55546456c-m7hlk.1642b6ee2cda37c8",
        "uid": "2ce39c6c-d09c-4a11-a7a6-2ea4ca1b9b1b",
        "resourceVersion": "23181",
        "creationTimestamp": "2020-10-30T08:12:31Z"
      },
      "involvedObject": {
        "kind": "Pod",
        "namespace": "test-guanggao",
        "name": "unit-test-repo-test-2-55546456c-m7hlk",
        "uid": "3111e4b3-7d50-4e89-9186-db658f7058aa",
        "apiVersion": "v1",
        "resourceVersion": "23126",
        "fieldPath": "spec.containers{unit-test-repo-test-2}"
      },
      "reason": "Pulling",
      "message": "Pulling image \"hub.c.163.com/commonwork/poc-template:tomcat7_jdk8_2\"",
      "source": {
        "component": "kubelet",
        "host": "docker-desktop"
      },
      "firstTimestamp": "2020-10-30T08:12:31Z",
      "lastTimestamp": "2020-10-30T08:12:47Z",
      "count": 2,
      "type": "Normal",
      "eventTime": null,
      "reportingComponent": "",
      "reportingInstance": ""
    }
  ]
}
`)
	_, _ = w.Write(d)
}

func (argoServer *ArgoServer) ResumeRollout(w http.ResponseWriter, r *http.Request) {
	const op = "argo mock server: resume rollout"

	ctx := log.WithContext(context.Background(), "ResumeRollout")
	defer wlog.Start(ctx, op).StopPrint()

	vars := mux.Vars(r)
	application := vars["application"]

	if argoServer.Applications[application] == nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		application := argoServer.Applications[application]
		application.synced = true
	}
	w.WriteHeader(http.StatusOK)
}

func (argoServer *ArgoServer) GetContainerLog(w http.ResponseWriter, r *http.Request) {
	const op = "argo mock server: get container log"

	ctx := log.WithContext(context.Background(), "GetContainerLog")
	defer wlog.Start(ctx, op).StopPrint()

	vars := mux.Vars(r)
	application := vars["application"]

	if application == "PodInitializing" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"grpc_code":9,"http_code":400,
"message":"container \"gmiddle\" in pod \"gmiddle-567dcb47b4-v49bw\" is waiting to start: PodInitializing",
"http_status":"Bad Request"}}`))
		return
	}

	if argoServer.Applications[application] == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	line := "Hello, World!"
	maxLine := 500
	tailLines := maxLine
	tailLinesQueryStr := r.FormValue("tailLines")
	if tailLinesQueryStr != "" {
		tailLinesQueryInt, err := strconv.Atoi(tailLinesQueryStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("tailLines is not a valid number!"))
		}
		if tailLinesQueryInt <= maxLine {
			tailLines = tailLinesQueryInt
		}
	}
	for i := 0; i < tailLines; i++ {
		_, _ = w.Write([]byte(fmt.Sprintf(`{"result": {"content": "[%v] %v", "timestamp": "%v"}}`,
			i+1, line, time.Now().Local()) + "\n"))
	}
}
