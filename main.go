package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"golang.org/x/time/rate"
)

func main() {
	routes := map[string]string{
		"/service1": "http://localhost:8001",
		"/service2": "http://localhost:8002",
	}

	mux := http.NewServeMux()

	for path, target := range routes {
		handler := newReverseProxy(target)
		handler = loggingMiddleware(handler)
		handler = authMiddleware(handler)
		handler = rateLimitMiddleware(handler)
		mux.Handle(path, handler)
	}

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))

	port := ":8080"
	log.Printf("Starting API Gateway on port %s\n", port)
	err := http.ListenAndServe(port, mux)
	if err != nil {
		log.Fatalf("Error starting API Gateway: %v", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		expectedToken := "Bearer secret-token" // Replace with your actual expected token

		if token != expectedToken {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func rateLimitMiddleware(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(1, 5) // 1 request per second with burst of 5

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func newReverseProxy(target string) http.Handler {
	url, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(url)

	proxy.ModifyResponse = func(response *http.Response) error {
		return nil
	}

	return proxy
}
