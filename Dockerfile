# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o api cmd/api/main.go

# Runtime Stage
FROM alpine:latest

WORKDIR /app

# Install CA certificates for HTTPS calls
RUN apk add --no-cache ca-certificates tzdata

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Copy binary from builder
COPY --from=builder --chown=appuser:appgroup /app/api .
# Copy migration files if needed (assuming migrations are applied separately or via an entrypoint script)
# COPY --from=builder /app/internal/database/migrations ./migrations

# Expose port
EXPOSE 8080

# Run
CMD ["./api"]
