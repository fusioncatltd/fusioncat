# Multi-stage Dockerfile for building and running the Fusioncat application
# Stage 1: Build the Go binary
# Stage 2: Create minimal runtime environment

# -- Stage 1: Builder -- #
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    bash \
    git \
    nodejs \
    npm \
    build-base

# Install quicktype globally
RUN npm install -g quicktype

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Generate Swagger documentation
RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init -g main.go -o ./docs

# Build the application
RUN go build -o bin/fusioncat main.go

# -- Stage 2: Runtime -- #
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    nodejs \
    npm \
    bash

# Install quicktype in runtime environment
RUN npm install -g quicktype

# Create app directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/bin/fusioncat /app/fusioncat

# Copy templates directory
COPY --from=builder /app/templates /app/templates

# Copy migrations directory
COPY --from=builder /app/migrations /app/migrations

# Note: manual_migrations and debug directories are optional
# They will be created but may be empty if not present in source

# Set environment variables for paths
ENV PATH_TO_STUBS_TEMPLATES_FOLDER=/app/templates
ENV PATH_TO_MIGRATIONS=/app/migrations
ENV JSON_SCHEMA_CONVERTOR_CMD=/usr/bin/quicktype

# Expose the application port (adjust if needed)
EXPOSE 8080

# Health check (adjust the endpoint if needed)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["/app/fusioncat"]