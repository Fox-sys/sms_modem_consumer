FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go mod download && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /sms-consumer ./cmd/sms-consumer

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /sms-consumer /sms-consumer
ENTRYPOINT ["/sms-consumer"]
