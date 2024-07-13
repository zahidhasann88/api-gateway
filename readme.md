# API Gateway

## Purpose
#### The API Gateway serves as a centralized entry point that routes requests to multiple backend services. It includes middleware for logging, authentication, and rate limiting to enhance security and performance.

## How to Run
### Prerequisites
1. Go installed on your local machine
2. Docker and Docker Compose (optional for containerized deployment)
## Local Installation
1. Clone the repository:
```bash
git clone <repository-url>
cd api-gateway
```
2. Initialize Go modules and run services:
```bash
go mod tidy
go run main.go
go run service1.go
go run service2.go
```

## Docker Installation
### Build and run with Docker Compose:
```bash
docker-compose up --build
```

## Middleware and Functionality
### Middleware
1. Logging: Records incoming requests for debugging and analysis.
1. Authentication: Verifies tokens to secure access to backend services.
1. Rate Limiting: Controls the rate of requests to prevent abuse and ensure service availability.

## Usage Example
### Assume you have two backend services running locally:

1. Service 1 on http://localhost:8001
2. Service 2 on http://localhost:8002

### The API Gateway routes requests as follows:

1. /service1 proxies to http://localhost:8001
2. /service2 proxies to http://localhost:8002