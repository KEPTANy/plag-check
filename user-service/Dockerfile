FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY shared/ ./shared/
COPY user-service/go.mod user-service/go.sum ./user-service/

WORKDIR /app/shared
RUN go mod download

WORKDIR /app/user-service
RUN go mod download

WORKDIR /app
COPY user-service ./user-service/

WORKDIR /app/user-service
RUN CGO_ENABLED=0 GOOS=linux go build -o user-service ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/user-service/user-service .
COPY --from=builder /app/user-service/migrations ./migrations/

EXPOSE 8081

CMD ["./user-service"]
