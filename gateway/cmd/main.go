package main

import (
	"log"

	"g.hz.netease.com/horizon/gateway/http/api/v1/login"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Printf("Server started")

	r := gin.Default()
	gin.ForceConsoleColor()
	login.RegisterRoutes(r)

	log.Fatal(r.Run(":8080"))
}
