package group

import (
	"fmt"
	"g.hz.netease.com/horizon/pkg/group/dao"
	"g.hz.netease.com/horizon/pkg/group/models"
	"g.hz.netease.com/horizon/server/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	CreateGroupError = "CreateGroupError"
)

type Controller struct {
	db  *gorm.DB
	dao dao.DAO
}

func NewController(db *gorm.DB) *Controller {
	return &Controller{
		db:  db,
		dao: dao.New(),
	}
}

func (controller *Controller) CreateGroup(c *gin.Context) {
	var group *models.Group
	err := c.ShouldBindJSON(&group)
	if err != nil {
		response.AbortWithRequestError(c, CreateGroupError, fmt.Sprintf("create group failed: %v", err))
		return
	}

	create, err := controller.dao.Create(controller.db, group)

	response.NewResponseWithData(create)
}
