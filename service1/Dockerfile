FROM golang:1.22.1-alpine as builder

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

RUN go build -o service1 .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/service1 .

EXPOSE 8001

CMD ["./service1"]
