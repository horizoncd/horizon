package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Route is the information for every URI.
type Route struct {
	// Method is the string for the HTTP method. ex) GET, POST etc..
	Method string
	// Pattern is the pattern of the URI.
	Pattern string
	// HandlerFunc is the handler function of this route.
	HandlerFunc gin.HandlerFunc
}

// Routes is the list of the generated Route.
type Routes []Route

// RegisterRoutes register every route to routerGroup
func RegisterRoutes(api *gin.RouterGroup, routes Routes) {
	for _, route := range routes {
		switch route.Method {
		case http.MethodGet:
			api.GET(route.Pattern, route.HandlerFunc)
		case http.MethodPost:
			api.POST(route.Pattern, route.HandlerFunc)
		case http.MethodPut:
			api.PUT(route.Pattern, route.HandlerFunc)
		case http.MethodPatch:
			api.PATCH(route.Pattern, route.HandlerFunc)
		case http.MethodDelete:
			api.DELETE(route.Pattern, route.HandlerFunc)
		}
	}
}
