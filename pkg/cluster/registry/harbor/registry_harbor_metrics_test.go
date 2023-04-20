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

package harbor

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

// nolint
func Test_Observe(t *testing.T) {
	Observe("server1", "GET", "200", "createProject", 100*time.Millisecond)
	Observe("server1", "GET", "300", "createProject", 400*time.Millisecond)
	Observe("server1", "GET", "400", "createProject", 1000*time.Millisecond)

	Observe("server2", "GET", "200", "addMembers", 100*time.Millisecond)
	Observe("server2", "GET", "300", "addMembers", 400*time.Millisecond)
	Observe("server2", "GET", "400", "addMembers", 1000*time.Millisecond)

	Observe("server1", "GET", "200", "deleteRepository", 100*time.Millisecond)
	Observe("server1", "GET", "300", "deleteRepository", 400*time.Millisecond)
	Observe("server1", "GET", "400", "deleteRepository", 1000*time.Millisecond)

	Observe("server1", "POST", "200", "listImage", 100*time.Millisecond)
	Observe("server1", "POST", "300", "listImage", 400*time.Millisecond)
	Observe("server1", "POST", "400", "listImage", 1000*time.Millisecond)

	metadata := `
		# HELP harbor_request_duration_milliseconds Harbor request duration in milliseconds
        # TYPE harbor_request_duration_milliseconds histogram
    `
	expect := `
		harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",le="100"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",le="200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",le="400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",le="800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="GET",operation="addMembers",server="server2",statuscode="200"} 100
        harbor_request_duration_milliseconds_count{method="GET",operation="addMembers",server="server2",statuscode="200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",le="100"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",le="200"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",le="400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",le="800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="GET",operation="addMembers",server="server2",statuscode="300"} 400
        harbor_request_duration_milliseconds_count{method="GET",operation="addMembers",server="server2",statuscode="300"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",le="100"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",le="200"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",le="400"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",le="800"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="GET",operation="addMembers",server="server2",statuscode="400"} 1000
        harbor_request_duration_milliseconds_count{method="GET",operation="addMembers",server="server2",statuscode="400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",le="100"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",le="200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",le="400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",le="800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="GET",operation="createProject",server="server1",statuscode="200"} 100
        harbor_request_duration_milliseconds_count{method="GET",operation="createProject",server="server1",statuscode="200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",le="100"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",le="200"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",le="400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",le="800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="GET",operation="createProject",server="server1",statuscode="300"} 400
        harbor_request_duration_milliseconds_count{method="GET",operation="createProject",server="server1",statuscode="300"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",le="100"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",le="200"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",le="400"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",le="800"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="GET",operation="createProject",server="server1",statuscode="400"} 1000
        harbor_request_duration_milliseconds_count{method="GET",operation="createProject",server="server1",statuscode="400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",le="100"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",le="200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",le="400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",le="800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="GET",operation="deleteRepository",server="server1",statuscode="200"} 100
        harbor_request_duration_milliseconds_count{method="GET",operation="deleteRepository",server="server1",statuscode="200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",le="100"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",le="200"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",le="400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",le="800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="GET",operation="deleteRepository",server="server1",statuscode="300"} 400
        harbor_request_duration_milliseconds_count{method="GET",operation="deleteRepository",server="server1",statuscode="300"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",le="100"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",le="200"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",le="400"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",le="800"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="GET",operation="deleteRepository",server="server1",statuscode="400"} 1000
        harbor_request_duration_milliseconds_count{method="GET",operation="deleteRepository",server="server1",statuscode="400"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",le="100"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",le="200"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",le="400"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",le="800"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="POST",operation="listImage",server="server1",statuscode="200"} 100
        harbor_request_duration_milliseconds_count{method="POST",operation="listImage",server="server1",statuscode="200"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",le="100"} 0
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",le="200"} 0
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",le="400"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",le="800"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="POST",operation="listImage",server="server1",statuscode="300"} 400
        harbor_request_duration_milliseconds_count{method="POST",operation="listImage",server="server1",statuscode="300"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",le="100"} 0
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",le="200"} 0
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",le="400"} 0
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",le="800"} 0
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="POST",operation="listImage",server="server1",statuscode="400"} 1000
        harbor_request_duration_milliseconds_count{method="POST",operation="listImage",server="server1",statuscode="400"} 1
	` // nolint

	err := testutil.CollectAndCompare(_harborDurationHistogram, strings.NewReader(metadata+expect))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
}
