// service1.go

package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Define a handler function for the endpoint
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from Backend Service 1!")
	})

	// Start the HTTP server on port 8001
	port := ":8001"
	log.Printf("Backend Service 1 listening on port %s\n", port)
	err := http.ListenAndServe(port, handler)
	if err != nil {
		log.Fatalf("Error starting Backend Service 1: %v", err)
	}
}
