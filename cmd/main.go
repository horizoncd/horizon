package main

import (
	"log"

	"github.com/horizon/http/api/v1/login"
)

func main() {
	log.Printf("Server started")

	router := login.NewRouter()

	log.Fatal(router.Run(":8080"))
}
