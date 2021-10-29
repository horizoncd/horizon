package registry

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func Test_observe(t *testing.T) {
	observe("server1", "GET", "/uri1", "200", "createProject", 100*time.Millisecond)
	observe("server1", "GET", "/uri1", "300", "createProject", 400*time.Millisecond)
	observe("server1", "GET", "/uri1", "400", "createProject", 1000*time.Millisecond)

	observe("server2", "GET", "/uri1", "200", "addMembers", 100*time.Millisecond)
	observe("server2", "GET", "/uri1", "300", "addMembers", 400*time.Millisecond)
	observe("server2", "GET", "/uri1", "400", "addMembers", 1000*time.Millisecond)

	observe("server1", "GET", "/uri2", "200", "deleteRepository", 100*time.Millisecond)
	observe("server1", "GET", "/uri2", "300", "deleteRepository", 400*time.Millisecond)
	observe("server1", "GET", "/uri2", "400", "deleteRepository", 1000*time.Millisecond)

	observe("server1", "POST", "/uri3", "200", "listImage", 100*time.Millisecond)
	observe("server1", "POST", "/uri3", "300", "listImage", 400*time.Millisecond)
	observe("server1", "POST", "/uri3", "400", "listImage", 1000*time.Millisecond)

	metadata := `
		# HELP harbor_request_duration_milliseconds Harbor request duration in milliseconds
        # TYPE harbor_request_duration_milliseconds histogram
    `
	expect := `
		harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",uri="/uri1",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",uri="/uri1",le="100"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",uri="/uri1",le="200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",uri="/uri1",le="400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",uri="/uri1",le="800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",uri="/uri1",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",uri="/uri1",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",uri="/uri1",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",uri="/uri1",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",uri="/uri1",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="200",uri="/uri1",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="GET",operation="addMembers",server="server2",statuscode="200",uri="/uri1"} 100
        harbor_request_duration_milliseconds_count{method="GET",operation="addMembers",server="server2",statuscode="200",uri="/uri1"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",uri="/uri1",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",uri="/uri1",le="100"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",uri="/uri1",le="200"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",uri="/uri1",le="400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",uri="/uri1",le="800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",uri="/uri1",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",uri="/uri1",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",uri="/uri1",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",uri="/uri1",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",uri="/uri1",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="300",uri="/uri1",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="GET",operation="addMembers",server="server2",statuscode="300",uri="/uri1"} 400
        harbor_request_duration_milliseconds_count{method="GET",operation="addMembers",server="server2",statuscode="300",uri="/uri1"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",uri="/uri1",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",uri="/uri1",le="100"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",uri="/uri1",le="200"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",uri="/uri1",le="400"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",uri="/uri1",le="800"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",uri="/uri1",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",uri="/uri1",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",uri="/uri1",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",uri="/uri1",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",uri="/uri1",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="addMembers",server="server2",statuscode="400",uri="/uri1",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="GET",operation="addMembers",server="server2",statuscode="400",uri="/uri1"} 1000
        harbor_request_duration_milliseconds_count{method="GET",operation="addMembers",server="server2",statuscode="400",uri="/uri1"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",uri="/uri1",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",uri="/uri1",le="100"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",uri="/uri1",le="200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",uri="/uri1",le="400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",uri="/uri1",le="800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",uri="/uri1",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",uri="/uri1",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",uri="/uri1",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",uri="/uri1",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",uri="/uri1",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="200",uri="/uri1",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="GET",operation="createProject",server="server1",statuscode="200",uri="/uri1"} 100
        harbor_request_duration_milliseconds_count{method="GET",operation="createProject",server="server1",statuscode="200",uri="/uri1"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",uri="/uri1",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",uri="/uri1",le="100"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",uri="/uri1",le="200"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",uri="/uri1",le="400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",uri="/uri1",le="800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",uri="/uri1",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",uri="/uri1",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",uri="/uri1",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",uri="/uri1",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",uri="/uri1",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="300",uri="/uri1",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="GET",operation="createProject",server="server1",statuscode="300",uri="/uri1"} 400
        harbor_request_duration_milliseconds_count{method="GET",operation="createProject",server="server1",statuscode="300",uri="/uri1"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",uri="/uri1",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",uri="/uri1",le="100"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",uri="/uri1",le="200"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",uri="/uri1",le="400"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",uri="/uri1",le="800"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",uri="/uri1",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",uri="/uri1",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",uri="/uri1",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",uri="/uri1",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",uri="/uri1",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="createProject",server="server1",statuscode="400",uri="/uri1",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="GET",operation="createProject",server="server1",statuscode="400",uri="/uri1"} 1000
        harbor_request_duration_milliseconds_count{method="GET",operation="createProject",server="server1",statuscode="400",uri="/uri1"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",uri="/uri2",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",uri="/uri2",le="100"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",uri="/uri2",le="200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",uri="/uri2",le="400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",uri="/uri2",le="800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",uri="/uri2",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",uri="/uri2",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",uri="/uri2",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",uri="/uri2",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",uri="/uri2",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="200",uri="/uri2",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="GET",operation="deleteRepository",server="server1",statuscode="200",uri="/uri2"} 100
        harbor_request_duration_milliseconds_count{method="GET",operation="deleteRepository",server="server1",statuscode="200",uri="/uri2"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",uri="/uri2",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",uri="/uri2",le="100"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",uri="/uri2",le="200"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",uri="/uri2",le="400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",uri="/uri2",le="800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",uri="/uri2",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",uri="/uri2",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",uri="/uri2",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",uri="/uri2",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",uri="/uri2",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="300",uri="/uri2",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="GET",operation="deleteRepository",server="server1",statuscode="300",uri="/uri2"} 400
        harbor_request_duration_milliseconds_count{method="GET",operation="deleteRepository",server="server1",statuscode="300",uri="/uri2"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",uri="/uri2",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",uri="/uri2",le="100"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",uri="/uri2",le="200"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",uri="/uri2",le="400"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",uri="/uri2",le="800"} 0
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",uri="/uri2",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",uri="/uri2",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",uri="/uri2",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",uri="/uri2",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",uri="/uri2",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="GET",operation="deleteRepository",server="server1",statuscode="400",uri="/uri2",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="GET",operation="deleteRepository",server="server1",statuscode="400",uri="/uri2"} 1000
        harbor_request_duration_milliseconds_count{method="GET",operation="deleteRepository",server="server1",statuscode="400",uri="/uri2"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",uri="/uri3",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",uri="/uri3",le="100"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",uri="/uri3",le="200"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",uri="/uri3",le="400"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",uri="/uri3",le="800"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",uri="/uri3",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",uri="/uri3",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",uri="/uri3",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",uri="/uri3",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",uri="/uri3",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="200",uri="/uri3",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="POST",operation="listImage",server="server1",statuscode="200",uri="/uri3"} 100
        harbor_request_duration_milliseconds_count{method="POST",operation="listImage",server="server1",statuscode="200",uri="/uri3"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",uri="/uri3",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",uri="/uri3",le="100"} 0
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",uri="/uri3",le="200"} 0
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",uri="/uri3",le="400"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",uri="/uri3",le="800"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",uri="/uri3",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",uri="/uri3",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",uri="/uri3",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",uri="/uri3",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",uri="/uri3",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="300",uri="/uri3",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="POST",operation="listImage",server="server1",statuscode="300",uri="/uri3"} 400
        harbor_request_duration_milliseconds_count{method="POST",operation="listImage",server="server1",statuscode="300",uri="/uri3"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",uri="/uri3",le="50"} 0
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",uri="/uri3",le="100"} 0
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",uri="/uri3",le="200"} 0
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",uri="/uri3",le="400"} 0
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",uri="/uri3",le="800"} 0
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",uri="/uri3",le="1600"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",uri="/uri3",le="3200"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",uri="/uri3",le="6400"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",uri="/uri3",le="12800"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",uri="/uri3",le="25600"} 1
        harbor_request_duration_milliseconds_bucket{method="POST",operation="listImage",server="server1",statuscode="400",uri="/uri3",le="+Inf"} 1
        harbor_request_duration_milliseconds_sum{method="POST",operation="listImage",server="server1",statuscode="400",uri="/uri3"} 1000
        harbor_request_duration_milliseconds_count{method="POST",operation="listImage",server="server1",statuscode="400",uri="/uri3"} 1
	` // nolint

	err := testutil.CollectAndCompare(_harborDurationHistogram, strings.NewReader(metadata+expect))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
}
