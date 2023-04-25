// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
