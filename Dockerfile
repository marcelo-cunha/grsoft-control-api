# Build stage
FROM golang:1.24-alpine AS builder

# Set necessary environment variables
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory /build
WORKDIR /build

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -a -installsuffix cgo -ldflags="-w -s" -o server ./cmd/server

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create appuser
RUN addgroup -g 1001 appgroup && adduser -u 1001 -G appgroup -s /bin/sh -D appuser

# Set working directory
WORKDIR /app

# Copy binary from build stage
COPY --from=builder /build/server .

# Change ownership to appuser
RUN chown -R appuser:appgroup /app

# Switch to appuser
USER appuser

EXPOSE 8080

# Command to run
ENTRYPOINT ["./server"]
