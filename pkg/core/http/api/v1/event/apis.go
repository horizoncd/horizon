package event

import (
	"github.com/horizoncd/horizon/pkg/core/controller/event"
	"github.com/horizoncd/horizon/pkg/server/response"

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

// ListSupportEvents list actions categorized based on resources
func (a *API) ListSupportEvents(c *gin.Context) {
	response.SuccessWithData(c, a.eventCtl.ListSupportEvents(c))
}
