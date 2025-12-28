FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY shared/ ./shared/
COPY analysis-service/go.mod analysis-service/go.sum ./analysis-service/

WORKDIR /app/shared
RUN go mod download

WORKDIR /app/analysis-service
RUN go mod download

WORKDIR /app
COPY analysis-service ./analysis-service/

WORKDIR /app/analysis-service
RUN CGO_ENABLED=0 GOOS=linux go build -o analysis-service ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/analysis-service/analysis-service .

EXPOSE 8083

CMD ["./analysis-service"]
