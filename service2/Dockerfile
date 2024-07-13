FROM golang:1.22.1-alpine as builder

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

RUN go build -o service2 .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/service2 .

EXPOSE 8002

CMD ["./service2"]
