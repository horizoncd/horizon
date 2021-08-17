package login

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Route is the information for every URI.
type Route struct {
	// Name is the name of this Route.
	Name string
	// Method is the string for the HTTP method. ex) GET, POST etc..
	Method string
	// Pattern is the pattern of the URI.
	Pattern string
	// HandlerFunc is the handler function of this route.
	HandlerFunc gin.HandlerFunc
}

// Routes is the list of the generated Route.
type Routes []Route

// NewRouter returns a new router.
func NewRouter() *gin.Engine {
	router := gin.Default()
	for _, route := range routes {
		switch route.Method {
		case http.MethodGet:
			router.GET(route.Pattern, route.HandlerFunc)
		case http.MethodPost:
			router.POST(route.Pattern, route.HandlerFunc)
		case http.MethodPut:
			router.PUT(route.Pattern, route.HandlerFunc)
		case http.MethodPatch:
			router.PATCH(route.Pattern, route.HandlerFunc)
		case http.MethodDelete:
			router.DELETE(route.Pattern, route.HandlerFunc)
		}
	}
	return router
}

var routes = Routes{
	{
		"Login",
		http.MethodGet,
		"/api/v1/login",
		Login,
	},

	{
		"LoginCallback",
		http.MethodGet,
		"/api/v1/login/callback",
		Callback,
	},

	{
		"Logout",
		http.MethodGet,
		"/api/v1/logout",
		Logout,
	},

	{
		"UserStatus",
		http.MethodGet,
		"/api/v1/status",
		UserStatus,
	},
}
