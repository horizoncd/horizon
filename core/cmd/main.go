package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"g.hz.netease.com/horizon/core/http/api/v1/login"
)

func main() {
	log.Printf("Server started")

	r := gin.Default()
	login.RegisterRoutes(r)

	log.Fatal(r.Run(":8080"))
}
