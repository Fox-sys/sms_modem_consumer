FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /sms-consumer ./cmd/sms-consumer

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /sms-consumer /sms-consumer
ENTRYPOINT ["/sms-consumer"]
