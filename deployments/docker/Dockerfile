# Build stage
FROM golang:1.26.0-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with explicit architecture support
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -o automation-hub cmd/automation-hub/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates && \
    adduser -D -u 1000 appuser

WORKDIR /app

# Copy binary with correct ownership
COPY --from=builder --chown=appuser:appuser /app/automation-hub .
# Copy config file
COPY --chown=appuser:appuser configs/config.yaml .

USER appuser

EXPOSE 8080

CMD ["./automation-hub"]