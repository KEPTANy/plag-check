FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY shared/ ./shared/
COPY gateway-api/go.mod ./gateway-api/

WORKDIR /app/shared
RUN go mod download

WORKDIR /app/gateway-api
RUN go mod download

WORKDIR /app
COPY gateway-api ./gateway-api/

WORKDIR /app/gateway-api
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway-api ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/gateway-api/gateway-api .

EXPOSE 8080

CMD ["./gateway-api"]
