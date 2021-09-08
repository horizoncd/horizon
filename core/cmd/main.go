package main

import (
	"log"

	"g.hz.netease.com/horizon/core/http/api/v1"
	"github.com/gin-gonic/gin"
)

func main(){
	log.Printf("Server started")

	r := gin.Default()
	gin.ForceConsoleColor()
	v1.RegisterRoutes(r)

	log.Fatal(r.Run(":8080"))
}