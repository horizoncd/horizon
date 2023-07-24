// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	rolloutv1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/stretchr/testify/assert"
	"k8s.io/kubernetes/pkg/apis/apps"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/argocd/mock"
	perror "github.com/horizoncd/horizon/pkg/errors"

	"github.com/horizoncd/horizon/pkg/util/log"
)

var (
	_master            = "master"
	_application       = "subgroup-for-unit-test"
	_cluster2          = "unit-test-repo-test-2"
	_cluster2Namespace = "test-guanggao"
	_cluster7          = "unit-test-repo-test-7"
	_argoURL           = "https://localhost:8080"
	_argoToken         = "" +
		""
	_argoClient          = &helper{URL: _argoURL, Token: _argoToken}
	_mockArgoServer      string
	_applicationManifest []byte
)

func TestMain(m *testing.M) {
	c := mock.NewArgoServer()
	s := httptest.NewServer(http.HandlerFunc(c.R.ServeHTTP))
	_mockArgoServer = s.Listener.Addr().String()
	_argoClient.URL = fmt.Sprintf("http://%v", _mockArgoServer)

	var crd Application
	crd.Metadata.Name = _application
	if data, err := json.Marshal(crd); err != nil {
		panic(err)
	} else {
		_applicationManifest = data
	}
	os.Exit(m.Run())
}

func TestApplication(t *testing.T) {
	ctx := log.WithContext(context.Background(), "TestApplication")
	SharedTestApplication(ctx, _argoClient, t)
}

func TestCreateApplication_Duplicate(t *testing.T) {
	ctx := log.WithContext(context.Background(), "TestCreateApplication_Duplicate")
	err := _argoClient.CreateApplication(ctx, _applicationManifest)
	if err != nil {
		t.Fatal(err)
	}

	err = _argoClient.CreateApplication(ctx, _applicationManifest)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeployApplication_Duplicate(t *testing.T) {
	ctx := log.WithContext(context.Background(), "TestApplication")

	defer func() {
		if err := _argoClient.DeleteApplication(ctx, _application); err != nil {
			t.Fatal(err)
		}
	}()

	if err := _argoClient.CreateApplication(ctx, _applicationManifest); err != nil {
		t.Fatal(err)
	}

	if err := _argoClient.DeployApplication(ctx, _application, _master); err != nil {
		t.Fatal(err)
	}

	if err := _argoClient.DeployApplication(ctx, _application, _master); err != nil {
		t.Fatal(err)
	}

	if err := _argoClient.WaitApplication(ctx, _application, "", http.StatusOK); err != nil {
		t.Fatal()
	}
}

func TestDeployApplication_NotExist(t *testing.T) {
	ctx := log.WithContext(context.Background(), "TestApplication")

	if err := _argoClient.DeployApplication(ctx, "TestDeployApplication_NotExist", _master); err == nil {
		t.Fatal(err)
	}
}

func TestGetApplication(t *testing.T) {
	ctx := log.WithContext(context.Background(), "TestApplication")

	defer func() {
		if err := _argoClient.DeleteApplication(ctx, _application); err != nil {
			t.Fatal(err)
		}
	}()

	if err := _argoClient.CreateApplication(ctx, _applicationManifest); err != nil {
		t.Fatal(err)
	}

	crd, err := _argoClient.GetApplication(ctx, _application)
	if err != nil || crd.Status.Sync.Status == v1alpha1.SyncStatusCodeSynced {
		t.Fatalf("crd.Status.Sync.Status != v1alpha1.SyncStatusCodeOutOfSync, value is %v", crd.Status.Sync.Status)
	}

	if err = _argoClient.DeployApplication(ctx, _application, _master); err != nil {
		t.Fatal(err)
	}

	if err = _argoClient.WaitApplication(ctx, _application, "", http.StatusOK); err != nil {
		t.Fatal()
	}

	crd, err = _argoClient.GetApplication(ctx, _application)
	if err != nil || crd.Status.Sync.Status != v1alpha1.SyncStatusCodeSynced {
		t.Fatalf("crd.Status.Sync.Status != v1alpha1.SyncStatusCodeSynced, value is %v", crd.Status.Sync.Status)
	}
}

func TestDeleteApplication_Duplicate(t *testing.T) {
	ctx := log.WithContext(context.Background(), "TestApplication")

	if err := _argoClient.CreateApplication(ctx, _applicationManifest); err != nil {
		t.Fatal(err)
	}

	if err := _argoClient.DeleteApplication(ctx, _application); err != nil {
		t.Fatal(err)
	}

	if err := _argoClient.DeleteApplication(ctx, _application); err != nil {
		t.Fatal(err)
	}
}

func TestWaitApplication_Duplicate(t *testing.T) {
	ctx := log.WithContext(context.Background(), "TestApplication")

	defer func() {
		if err := _argoClient.DeleteApplication(ctx, _application); err != nil {
			t.Fatal(err)
		}
	}()

	if err := _argoClient.CreateApplication(ctx, _applicationManifest); err != nil {
		t.Fatal(err)
	}

	if err := _argoClient.DeployApplication(ctx, _application, _master); err != nil {
		t.Fatal(err)
	}

	if err := _argoClient.WaitApplication(ctx, _application, "", http.StatusOK); err != nil {
		t.Fatal()
	}

	if err := _argoClient.WaitApplication(ctx, _application, "", http.StatusOK); err != nil {
		t.Fatal()
	}

	if err := _argoClient.WaitApplication(ctx, _application, "", http.StatusOK); err != nil {
		t.Fatal()
	}
}

func SharedTestApplication(ctx context.Context, argoClient *helper, t *testing.T) {
	if err := argoClient.DeleteApplication(ctx, _application); err != nil {
		t.Fatal(err)
	}

	if err := argoClient.WaitApplication(ctx, _application, "", http.StatusNotFound); err != nil {
		t.Fatal(err)
	}

	if err := argoClient.CreateApplication(ctx, _applicationManifest); err != nil {
		t.Fatal(err)
	}

	if err := argoClient.DeployApplication(ctx, _application, _master); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(argoClient.URL, _mockArgoServer) {
		var crd Application
		var manifest []byte
		crd.Metadata.Name = _cluster2
		if data, err := json.Marshal(crd); err != nil {
			panic(err)
		} else {
			manifest = data
		}
		if err := argoClient.CreateApplication(ctx, manifest); err != nil {
			t.Fatal(err)
		}
	}

	if err := argoClient.WaitApplication(ctx, _application, "", http.StatusOK); err != nil {
		t.Fatal()
	}

	if err := argoClient.DeployApplication(ctx, _cluster2, _master); err != nil {
		t.Fatal(err)
	}

	if err := argoClient.WaitApplication(ctx, _cluster2, "", http.StatusOK); err != nil {
		t.Fatal()
	}

	// clientset := fakek8s.NewSimpleClientset()
	// _, _, err := argoClient.GetCluster(ctx, _cluster2, clientset)
	// if err != nil && errors.Status(err) != http.StatusNotFound {
	// 	t.Fatal(err)
	// }

	if _, err := argoClient.GetApplicationTree(ctx, "notfound"); err == nil {
		assert.NotNil(t, err)
		_, ok := perror.Cause(err).(*herrors.HorizonErrNotFound)
		assert.True(t, ok)
	}

	if tree, err := argoClient.GetApplicationTree(ctx, _cluster2); err != nil {
		t.Fatal(err)
	} else {
		data, err := json.Marshal(tree)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("application tree:", string(data))
	}

	var deployment *apps.Deployment
	err := argoClient.GetApplicationResource(ctx, _cluster2, ResourceParams{
		Group:        "apps",
		Version:      "v1",
		Kind:         "Deployment",
		Namespace:    _cluster2Namespace,
		ResourceName: _cluster2,
	}, &deployment)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Logf("got: %v", deployment)
	}

	_, err = argoClient.ListResourceEvents(ctx, _cluster2, EventParam{
		ResourceNamespace: _cluster2Namespace,
		ResourceUID:       string(deployment.UID),
		ResourceName:      _cluster2,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := argoClient.DeleteApplication(ctx, _cluster2); err != nil {
		t.Fatal(err)
	}

	if err := argoClient.DeleteApplication(ctx, _cluster7); err != nil {
		t.Fatal(err)
	}

	if err := argoClient.DeleteApplication(ctx, _application); err != nil {
		t.Fatal(err)
	}
}

func TestResumeRollout(t *testing.T) {
	ctx := log.WithContext(context.Background(), "TestByMock")
	argoClient := &helper{URL: fmt.Sprintf("http://%v", _mockArgoServer)}

	if err := argoClient.CreateApplication(ctx, _applicationManifest); err != nil {
		t.Fatal(err)
	}

	if err := argoClient.ResumeRollout(ctx, _application); err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshal(t *testing.T) {
	var rollout *rolloutv1.Rollout
	err := json.Unmarshal([]byte("{}"), &rollout)
	assert.Nil(t, err)
	assert.NotNil(t, rollout)
}

func TestGetContainerLog(t *testing.T) {
	ctx := log.WithContext(context.Background(), "TestByMock")
	argoClient := &helper{URL: fmt.Sprintf("http://%v", _mockArgoServer)}
	if err := argoClient.CreateApplication(ctx, _applicationManifest); err != nil {
		t.Fatal(err)
	}
	logC, errC, err := argoClient.GetContainerLog(ctx, _application, ContainerLogParams{
		Namespace:     "test-guanggao",
		PodName:       "podName",
		ContainerName: "containerName",
		TailLines:     20,
	})
	if err != nil {
		t.Fatal(err)
	}
	for logC != nil || errC != nil {
		select {
		case l, ok := <-logC:
			if !ok {
				logC = nil
				continue
			}
			t.Logf("[%s] %s\n", l.Result.Timestamp, l.Result.Content)
		case e, ok := <-errC:
			if !ok {
				errC = nil
				continue
			}
			t.Logf("%s\n", e)
		}
	}
}
func TestGetContainerLog_PodInitializing(t *testing.T) {
	ctx := log.WithContext(context.Background(), "TestByMock")
	argoClient := &helper{URL: fmt.Sprintf("http://%v", _mockArgoServer)}
	if err := argoClient.CreateApplication(ctx, _applicationManifest); err != nil {
		t.Fatal(err)
	}

	_, _, err := argoClient.GetContainerLog(ctx, "PodInitializing", ContainerLogParams{
		Namespace:     "test-guanggao",
		PodName:       "podName",
		ContainerName: "containerName",
	})
	assert.NotNil(t, err)
	assert.Equal(t, perror.Cause(err), herrors.ErrHTTPRespNotAsExpected)
	assert.True(t, strings.Contains(err.Error(), "is waiting to start: PodInitializing"))
}
