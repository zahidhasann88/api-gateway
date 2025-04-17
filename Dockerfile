FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-gateway ./cmd/gateway

# Create a minimal image
FROM alpine:3.18

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /api-gateway .
# Copy config files
COPY --from=builder /app/configs ./configs/

# Expose the port
EXPOSE 8080

# Command to run
CMD ["./api-gateway"]