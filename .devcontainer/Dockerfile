# Development container for automation-hub
FROM golang:1.24.5-bookworm

# Install development tools and dependencies
RUN apt-get update && apt-get install -y \
    git \
    curl \
    wget \
    bash \
    ca-certificates \
    build-essential \
    sudo \
    && rm -rf /var/lib/apt/lists/*

# Create a non-root user for development
RUN useradd -m -s /bin/bash vscode && \
    echo "vscode ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Install Go tools for development
RUN go install golang.org/x/tools/gopls@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install golang.org/x/tools/cmd/goimports@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Set working directory
WORKDIR /workspace

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Change ownership to vscode user
RUN chown -R vscode:vscode /workspace /go

# Switch to vscode user
USER vscode

# Expose the application port
EXPOSE 8080

# Keep container running
CMD ["sleep", "infinity"]
