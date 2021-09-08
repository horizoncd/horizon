package group

import (
	"fmt"

	groupDao "g.hz.netease.com/horizon/pkg/group/dao"
	"g.hz.netease.com/horizon/pkg/group/models"
	"g.hz.netease.com/horizon/server/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	CreateGroupError = "CreateGroupError"
)

type Controller struct {
	db       *gorm.DB
	groupDao groupDao.DAO
}

func NewController(db *gorm.DB) *Controller {
	return &Controller{
		db:       db,
		groupDao: groupDao.New(),
	}
}

func (controller *Controller) CreateGroup(c *gin.Context) {
	var group *models.Group
	err := c.ShouldBindJSON(&group)
	if err != nil {
		response.AbortWithRequestError(c, CreateGroupError, fmt.Sprintf("create group failed: %v", err))
		return
	}

	create, err := controller.groupDao.Create(controller.db, group)
	if err != nil {
		response.AbortWithInternalError(c, CreateGroupError, fmt.Sprintf("create group failed: %v", err))
		return
	}

	response.NewResponseWithData(create)
}
