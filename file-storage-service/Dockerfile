FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY shared/ ./shared/
COPY file-storage-service/go.mod file-storage-service/go.sum ./file-storage-service/

WORKDIR /app/shared
RUN go mod download

WORKDIR /app/file-storage-service
RUN go mod download

WORKDIR /app
COPY file-storage-service ./file-storage-service/

WORKDIR /app/file-storage-service
RUN CGO_ENABLED=0 GOOS=linux go build -o file-storage-service ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

RUN mkdir -p /uploads

COPY --from=builder /app/file-storage-service/file-storage-service .
COPY --from=builder /app/file-storage-service/migrations ./migrations/

EXPOSE 8082

CMD ["./file-storage-service"]
