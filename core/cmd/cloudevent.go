package cmd

import (
	"fmt"
	"log"

	cloudeventctl "g.hz.netease.com/horizon/core/controller/cloudevent"
	"g.hz.netease.com/horizon/core/http/cloudevent"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/factory"
	"g.hz.netease.com/horizon/pkg/config/server"

	"github.com/gin-gonic/gin"
)

func runCloudEventServer(tektonFty factory.Factory, config server.Config) {
	r := gin.Default()

	cloudEventCtl := cloudeventctl.NewController(tektonFty)

	cloudevent.RegisterRoutes(r, cloudevent.NewAPI(cloudEventCtl))

	log.Fatal(r.Run(fmt.Sprintf(":%d", config.Port)))
}
