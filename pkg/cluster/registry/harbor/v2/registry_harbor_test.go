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

package v2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/horizoncd/horizon/pkg/cluster/registry"
	"github.com/horizoncd/horizon/pkg/cluster/registry/harbor/v2/mockserver"
	"github.com/stretchr/testify/assert"
)

var config = &registry.Config{}
var server = mockserver.NewHarborServer()

func TestMain(m *testing.M) {
	s := httptest.NewServer(http.HandlerFunc(server.R.ServeHTTP))
	config.Server = "http://" + s.Listener.Addr().String()
	os.Exit(m.Run())
}

func TestByMock(t *testing.T) {
	config.Path = "project1"
	registry, _ := NewHarborRegistry(config)
	h := registry.(*Registry)
	ctx := context.Background()

	server.CreateProject("project1", nil)
	server.PushImage("project1", "horizon-demo/horizon-demo-dev", "v1")

	err := h.DeleteImage(ctx, "horizon-demo", "horizon-demo-dev")
	assert.Nil(t, err)
	err = h.DeleteImage(ctx, "horizon-demo", "horizon-demo-dev")
	assert.Nil(t, err)
}
