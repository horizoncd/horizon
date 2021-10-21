package factory

import (
	"context"
	"sync"
	"testing"

	"g.hz.netease.com/horizon/pkg/config/gitlab"

	"github.com/stretchr/testify/assert"
)

var (
	ctx        = context.Background()
	gitlabName = "control"
)

func Test(t *testing.T) {
	mgr := &factory{
		m: &sync.Map{},
		gitlabMapper: gitlab.Mapper{
			gitlabName: {
				HTTPURL: "https://gitlab.com",
				SSHURL:  "ssh://gitlab.com",
				Token:   "asdfghjk",
			},
			"compute": {
				HTTPURL: "",
				SSHURL:  "",
				Token:   "",
			},
		},
	}

	// 1. query db at first
	gitlabLib, err := mgr.GetByName(ctx, gitlabName)
	assert.Nil(t, err)
	assert.NotNil(t, gitlabLib)

	// 2. get from cache directly
	gitlabLib, err = mgr.GetByName(ctx, gitlabName)
	assert.Nil(t, err)
	assert.NotNil(t, gitlabLib)

	// 3. get for name not exists
	gitlabLib, err = mgr.GetByName(ctx, "not-exists")
	assert.NotNil(t, err)
	assert.Nil(t, gitlabLib)
}

func TestNewController(t *testing.T) {
	ctl := NewFactory(nil)
	assert.NotNil(t, ctl)
}
