package event

import (
	"g.hz.netease.com/horizon/core/controller/event"
	"g.hz.netease.com/horizon/pkg/server/response"

	"github.com/gin-gonic/gin"
)

type API struct {
	eventCtl event.Controller
}

// NewAPI initializes a new event api
func NewAPI(controller event.Controller) *API {
	return &API{
		eventCtl: controller,
	}
}

// ListEventActions list actions categorized based on resources
func (a *API) ListEventActions(c *gin.Context) {
	response.SuccessWithData(c, a.eventCtl.ListEventActions(c))
}
