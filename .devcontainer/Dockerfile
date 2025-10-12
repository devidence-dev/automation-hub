# Development container for automation-hub
FROM golang:1.24.5-bookworm

# Create a non-root user for development
RUN useradd -m -s /bin/bash vscode && \
    echo "vscode ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Install development tools and dependencies with non-interactive frontend
RUN DEBIAN_FRONTEND=noninteractive apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
    --no-install-recommends \
    -o Dpkg::Options::="--force-confdef" \
    -o Dpkg::Options::="--force-confold" \
    bash \
    build-essential \
    ca-certificates \
    curl \
    git \
    sudo \
    wget \
    && rm -rf /var/lib/apt/lists/*

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
