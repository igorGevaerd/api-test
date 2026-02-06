# Stage 1: Build the application
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY cmd ./cmd
COPY internal ./internal

# Build the application
RUN go build -o api-server ./cmd/api

# Stage 2: Create the runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS requests (if needed)
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/api-server .

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./api-server"]
