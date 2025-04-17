# Go API Gateway

A high-performance, feature-rich API Gateway implemented in Go, designed to provide a unified entry point for client applications to access various microservices.

## Features

- **Routing & Proxying**: Forward requests to appropriate backend services
- **Load Balancing**: Distribute traffic across multiple service instances
- **Authentication & Authorization**: JWT-based authentication and role-based access control
- **Rate Limiting**: Protect backend services from overload
- **Circuit Breaker**: Prevent cascading failures in microservice environments
- **Observability**: Comprehensive logging, metrics, and tracing
- **WebSocket Support**: Bi-directional communication proxying
- **GraphQL Support**: Forward GraphQL requests to backend services
- **Request/Response Transformation**: Modify requests and responses as they pass through the gateway
- **Cross-Origin Resource Sharing (CORS)**: Built-in CORS support
- **Health Checks**: Ensure backend services are healthy

## Architecture

The API Gateway uses a clean and modular architecture:

- `cmd/`: Application entry points
- `internal/`: Private application code
  - `config/`: Configuration management
  - `server/`: HTTP server implementation
  - `handlers/`: Request handlers
  - `middleware/`: HTTP middleware
- `pkg/`: Shared packages
  - `logger/`: Logging utilities
  - `circuitbreaker/`: Circuit breaker implementation
- `configs/`: Configuration files
- `deploy/`: Deployment configurations

## Getting Started

### Prerequisites

- Go 1.19 or newer
- Docker (optional)
- Kubernetes (optional)

### Installation

1. Clone the repository

```bash
git clone https://github.com/yourusername/api-gateway.git
cd api-gateway
```

2. Install dependencies

```bash
go mod download
```

3. Configure the API Gateway

Edit the configuration file in `configs/config.yaml` to match your environment.

4. Build and run the application

```bash
go build -o api-gateway ./cmd/gateway
./api-gateway
```

### Docker

Build a Docker image:

```bash
docker build -t api-gateway .
```

Run the container:

```bash
docker run -p 8080:8080 api-gateway
```

### Kubernetes

Deploy to Kubernetes:

```bash
kubectl apply -f deploy/kubernetes/
```

## Configuration

The gateway is configured through a YAML file. Here's an example:

```yaml
logLevel: info

server:
  address: :8080
  timeout: 30

cors:
  allowedOrigins:
    - "*"
  allowedMethods:
    - GET
    - POST
    - PUT
    - DELETE
    - OPTIONS
  allowedHeaders:
    - Authorization
    - Content-Type

proxy:
  readTimeout: 5
  writeTimeout: 10
  idleTimeout: 120

auth:
  enabled: true
  jwtSecret: "your-jwt-secret"
  expiration: 24h
  issuer: "api-gateway"

services:
  users:
    url: http://users-service:8081
    timeout: 5
    retryCount: 3
    rateLimit: 100
    authentication: true
    authorization:
      roles:
        - admin
        - user
    circuitBreaker:
      enabled: true
      failureThreshold: 5
      resetTimeout: "10s"
      halfOpenSuccessThreshold: 2
    transformations:
      request:
        fieldMapping:
          "username": "user_name"
        headerToBody:
          "X-User-ID": "userId"
      response:
        fieldMapping:
          "user_id": "userId"
        bodyToHeader:
          "token": "X-Auth-Token"
          
  payments:
    url: http://payments-service:8082
    timeout: 10
    retryCount: 2
    rateLimit: 50
    authentication: true
    authorization:
      roles:
        - admin
```

## API Endpoints

By default, the API Gateway exposes the following endpoints:

- `GET /health`: Health check endpoint
- `GET /metrics`: Prometheus metrics
- `POST /api/login`: Authentication endpoint to get JWT tokens
- `/api/{service-name}/{path}`: Proxy requests to backend services
- `/ws/{service-name}/{path}`: WebSocket proxy
- `POST /graphql/{service-name}`: GraphQL proxy

## Security

The gateway implements several security measures:

- JWT Authentication
- Role-based Authorization
- Rate Limiting
- CORS Configuration
- Secure Headers

## Observability

The gateway provides observability through:

- Structured JSON logging with Zap
- Prometheus metrics
- Distributed tracing (when configured)
- Request ID tracking

## Performance

The API Gateway is designed for high performance:

- Low memory footprint
- Connection pooling
- Efficient routing
- Configurable timeouts
- Graceful shutdown