# Build stage
FROM golang:1.24.5-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o automation-hub cmd/automation-hub/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary
COPY --from=builder /app/automation-hub .
COPY --from=builder /app/configs ./configs

EXPOSE 8080

CMD ["./automation-hub"]