package tekton

import (
	"encoding/json"
	"os"
	"testing"

	tektonconfig "g.hz.netease.com/horizon/pkg/config/tekton"
	"github.com/stretchr/testify/assert"
)

func TestNewTekton(t *testing.T) {
	tektonConfig := &tektonconfig.Tekton{
		Kubeconfig: "/",
	}
	tekton, err := NewTekton(tektonConfig)
	assert.Nil(t, tekton)
	assert.NotNil(t, err)

	tektonConfig = &tektonconfig.Tekton{
		Kubeconfig: "",
	}
	tekton, err = NewTekton(tektonConfig)
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	// 通过这两个环境变量判断是否在k8s集群中运行
	if len(host) == 0 || len(port) == 0 {
		assert.Nil(t, tekton)
		assert.NotNil(t, err)
	} else {
		assert.NotNil(t, tekton)
		assert.Nil(t, err)
	}
}

func TestPipelineRun_UnmarshalJSON(t *testing.T) {
	var pipelineRun PipelineRun
	// case1: dockerfile is a object with dockerfile path
	data1 := `
	{
	    "application":"app",
	    "cluster":"cluster",
	    "environment":"test",
	    "git":{
	        "url":"ssh://git.url",
	        "branch":"master",
	        "subfolder":"/"
	    },
	    "buildxml":"buildxml",
	    "imageurl":"hub.c.163.com/xxx/xxx:latest",
	    "operator":"gjq",
	    "dockerfile":{
             "content": "",
             "path": "dockerfile path"
         }
	}`
	if err := json.Unmarshal([]byte(data1), &pipelineRun); err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("%s", pipelineRun)
	assert.Equal(t, "", pipelineRun.Dockerfile.Content)
	assert.Equal(t, "dockerfile path", pipelineRun.Dockerfile.Path)

	// case2: dockerfile is a object with all value
	data2 := `
	{
	    "application":"app",
	    "cluster":"cluster",
	    "environment":"test",
	    "git":{
	        "url":"ssh://git.url",
	        "branch":"master",
	        "subfolder":"/"
	    },
	    "buildxml":"buildxml",
	    "imageurl":"hub.c.163.com/xxx/xxx:latest",
	    "operator":"gjq",
	    "dockerfile":{
             "content": "dockerfile content",
             "path": "dockerfile path"
         }
	}`
	if err := json.Unmarshal([]byte(data2), &pipelineRun); err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("%s", pipelineRun)
	assert.Equal(t, "dockerfile content", pipelineRun.Dockerfile.Content)
	assert.Equal(t, "dockerfile path", pipelineRun.Dockerfile.Path)
}
