# Start with a minimal base image that includes Go and tools
FROM golang:1.23.1-bookworm AS builder

# Install root certificates — required for HTTPS connections (OAuth, APIs, etc.)
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

# Set working directory inside the container
WORKDIR /app

# Copy go mod files first for better build caching
COPY go.mod go.sum ./

# Download dependencies (this layer is cached if go.mod/go.sum don't change)
RUN go mod download

# Copy the rest of your application
COPY . .

# Build the Go binary
RUN CGO_ENABLED=1 go build -o forum

# ---- Runtime container ----
FROM debian:bookworm-slim

# Install root certificates and SQLite in the final image
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates sqlite3 && rm -rf /var/lib/apt/lists/*

# Create a non-root user for security
RUN groupadd -r forum && useradd -r -g forum forum

# Set working directory
WORKDIR /app

# Copy the compiled binary from the builder
COPY --from=builder /app/forum .

# Copy static files and templates
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/sql ./sql
COPY --from=builder /app/populate ./populate

# Create uploads directory and set permissions
RUN mkdir -p uploads && \
    chown -R forum:forum /app && \
    chmod 755 /app && \
    chmod 755 /app/uploads

# Switch to non-root user
USER forum

# Expose port 8080
EXPOSE 8080

# Command to run your app
CMD ["./forum"] 