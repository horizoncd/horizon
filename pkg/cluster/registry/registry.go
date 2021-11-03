package registry

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// Registry ...
type Registry interface {
	// CreateProject create a project, if the project is already exists, return true, or return project's ID
	CreateProject(ctx context.Context, project string) (int, error)
	// AddMembers add members for project
	AddMembers(ctx context.Context, projectID int) error
	// DeleteRepository delete repository
	DeleteRepository(ctx context.Context, project string, repository string) error
	// ListImage list images for a repository
	ListImage(ctx context.Context, project string, repository string) ([]string, error)
	PreheatProject(ctx context.Context, project string,
		projectID int) (err error)
	GetServer(ctx context.Context) string
}

type HarborMember struct {
	// harbor role 1:manager，2:developer，3:guest
	Role int `yaml:"role"`
	// harbor user name
	Username string `yaml:"username"`
}

// HarborRegistry implement Registry
type HarborRegistry struct {
	// harbor server address
	server string
	// harbor token
	token string
	// harbor preheat policy id
	preheatPolicyID int
	// the member to add to projects
	members []*HarborMember
	// http client
	client *http.Client
	// retryableClient retryable client
	retryableClient *retryablehttp.Client
}

// default params
const (
	_backoffDuration = 1 * time.Second
	_retry           = 3
	_timeout         = 4 * time.Second
)

// members harbor member to add to harbor project
// TODO(gjq): move this to config
var members = []*HarborMember{
	{
		Role:     3,
		Username: "musiccloudnative",
	},
}

type HarborConfig struct {
	Server          string
	Token           string
	PreheatPolicyID int
}

var harborRegistryCache *sync.Map

func init() {
	harborRegistryCache = &sync.Map{}
}

// NewHarborRegistry new a HarborRegistry
func NewHarborRegistry(harbor *HarborConfig) Registry {
	key := fmt.Sprintf("%v-%v", harbor.Server, harbor.Token)
	if ret, ok := harborRegistryCache.Load(key); ok {
		return ret.(Registry)
	}

	transport := http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	harborRegistry := &HarborRegistry{
		server:          harbor.Server,
		token:           harbor.Token,
		preheatPolicyID: harbor.PreheatPolicyID,
		members:         members,
		client: &http.Client{
			Transport: &transport,
		},
		retryableClient: &retryablehttp.Client{
			HTTPClient: &http.Client{
				Transport: &transport,
				Timeout:   _timeout,
			},
			RetryMax:   _retry,
			CheckRetry: retryablehttp.DefaultRetryPolicy,
			Backoff: func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
				// wait for this duration if failed
				return _backoffDuration
			},
		},
	}

	harborRegistryCache.Store(key, harborRegistry)

	return harborRegistry
}

var _ Registry = (*HarborRegistry)(nil)
