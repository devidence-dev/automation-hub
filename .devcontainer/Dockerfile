# Development container for automation-hub
FROM golang:1.26.0

# Create vscode user and set up directories
RUN useradd -m -s /bin/bash vscode && \
    mkdir -p /go/pkg/mod /go/bin && \
    chown -R vscode:vscode /go

# Set working directory
WORKDIR /workspace

# Switch to vscode user for Go operations
USER vscode

# Copy go mod files first for better caching
COPY --chown=vscode:vscode go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install Go development tools
RUN go install golang.org/x/tools/gopls@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install golang.org/x/tools/cmd/goimports@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Expose the application port
EXPOSE 8080

# Keep container running
CMD ["sleep", "infinity"]
