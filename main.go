// main.go

package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	// Define routes to backend services
	routes := map[string]string{
		"/service1": "http://localhost:8001", // Example backend service 1
		"/service2": "http://localhost:8002", // Example backend service 2
	}

	// Initialize a new HTTP multiplexer (router)
	mux := http.NewServeMux()

	// Add middleware for logging
	mux.HandleFunc("/", loggingMiddleware)

	// Add reverse proxy handler for each route
	for path, target := range routes {
		mux.Handle(path, newReverseProxy(target))
	}

	// Start the API Gateway server
	port := ":8080"
	log.Printf("Starting API Gateway on port %s\n", port)
	err := http.ListenAndServe(port, mux)
	if err != nil {
		log.Fatalf("Error starting API Gateway: %v", err)
	}
}

// loggingMiddleware logs the incoming requests
func loggingMiddleware(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: %s %s", r.Method, r.URL.Path)
}

// newReverseProxy creates a new reverse proxy handler for the given target URL
func newReverseProxy(target string) http.Handler {
	url, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(url)

	// Modify the response from the backend if needed
	proxy.ModifyResponse = func(response *http.Response) error {
		// Add custom logic here if required
		return nil
	}

	return proxy
}
