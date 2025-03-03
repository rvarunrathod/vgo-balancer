FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o vgo-balancer ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/vgo-balancer .
COPY config.yaml .

EXPOSE 8080 8081 9090
CMD ["./vgo-balancer", "--config", "config.yaml"]