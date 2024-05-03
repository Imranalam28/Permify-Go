package main

import (
	"log"
	"net/http"
)

var sessions = map[string]string{}

func main() {
	setupPermifyClient()
	setupRoutes()

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
