package gitlab

import (
	"context"
	"sync"
	"testing"

	gitlabmock "g.hz.netease.com/horizon/mock/pkg/gitlab"
	"g.hz.netease.com/horizon/pkg/gitlab/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	ctx         = context.Background()
	gitlabName  = "control"
	gitlabURL   = "https://gitlab.com"
	gitlabToken = "asdfghjk"
)

func Test(t *testing.T) {
	mockCtl := gomock.NewController(t)
	gitlabMgr := gitlabmock.NewMockManager(mockCtl)
	gitlabMgr.EXPECT().GetByName(ctx, gitlabName).Return(&models.Gitlab{
		Model: gorm.Model{
			ID: 1,
		},
		Name:  gitlabName,
		URL:   gitlabURL,
		Token: gitlabToken,
	}, nil)
	gitlabMgr.EXPECT().GetByName(ctx, "not-exists").Return(nil, nil)

	ctl := &controller{
		m:         &sync.Map{},
		gitlabMgr: gitlabMgr,
	}

	// 1. query db at first
	gitlabLib, err := ctl.GetByName(ctx, gitlabName)
	assert.Nil(t, err)
	assert.NotNil(t, gitlabLib)

	// 2. get from cache directly
	gitlabLib, err = ctl.GetByName(ctx, gitlabName)
	assert.Nil(t, err)
	assert.NotNil(t, gitlabLib)

	// 3. get for name not exists
	gitlabLib, err = ctl.GetByName(ctx, "not-exists")
	assert.NotNil(t, err)
	assert.Nil(t, gitlabLib)
}

func TestNewController(t *testing.T) {
	ctl := NewController()
	assert.NotNil(t, ctl)
}
