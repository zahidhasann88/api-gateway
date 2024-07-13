// service2.go

package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Define a handler function for the endpoint
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from Backend Service 2!")
	})

	// Start the HTTP server on port 8002
	port := ":8002"
	log.Printf("Backend Service 2 listening on port %s\n", port)
	err := http.ListenAndServe(port, handler)
	if err != nil {
		log.Fatalf("Error starting Backend Service 2: %v", err)
	}
}
