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

package cmd

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	cloudeventctl "github.com/horizoncd/horizon/core/controller/cloudevent"
	"github.com/horizoncd/horizon/core/http/cloudevent"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/factory"
	"github.com/horizoncd/horizon/pkg/config/server"
	"github.com/horizoncd/horizon/pkg/param"
)

func runCloudEventServer(tektonFty factory.Factory, config server.Config,
	parameter *param.Param, middlewares ...gin.HandlerFunc) {
	r := gin.Default()
	r.Use(middlewares...)

	cloudEventCtl := cloudeventctl.NewController(tektonFty, parameter)

	cloudevent.RegisterRoutes(r, cloudevent.NewAPI(cloudEventCtl))

	log.Fatal(r.Run(fmt.Sprintf(":%d", config.Port)))
}
