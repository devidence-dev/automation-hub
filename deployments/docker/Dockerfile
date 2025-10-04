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

RUN apk --no-cache add ca-certificates && \
    adduser -D -u 1000 appuser

WORKDIR /app

# Copy binary with correct ownership
COPY --from=builder --chown=appuser:appuser /app/automation-hub .
# Note: configs folder will be mounted as a volume from docker-compose.yml

USER appuser

EXPOSE 8080

CMD ["./automation-hub"]