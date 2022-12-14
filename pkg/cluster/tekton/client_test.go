package tekton

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
)

func TestInitClients(t *testing.T) {
	client, err := InitClient("/")
	assert.Nil(t, client)
	assert.NotNil(t, err)

	client2, err2 := InitClient("")
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		assert.Nil(t, client2)
		assert.NotNil(t, err2)
	} else {
		assert.NotNil(t, client2)
		assert.Nil(t, err2)
	}
}

func Test_tektonClient(t *testing.T) {
	c, err := tektonClient(&rest.Config{})
	assert.Nil(t, err)
	assert.NotNil(t, c)
}

func Test_kubeClient(t *testing.T) {
	c, err := kubeClient(&rest.Config{})
	assert.Nil(t, err)
	assert.NotNil(t, c)
}

func Test_dynamicClient(t *testing.T) {
	c, err := dynamicClient(&rest.Config{})
	assert.Nil(t, err)
	assert.NotNil(t, c)
}
