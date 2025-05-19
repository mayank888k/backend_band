FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go module files first for better caching
COPY go.* ./
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o server ./cmd/api

# Use a smaller image for runtime
FROM alpine:latest

# Install dependencies required for SQLite
RUN apk add --no-cache ca-certificates libc6-compat tzdata

WORKDIR /app

# Create data directory for SQLite
RUN mkdir -p /app/data

# Copy the binary from the builder stage
COPY --from=builder /app/server .

# Set reasonable permissions
RUN chmod +x /app/server

# Set timezone
ENV TZ=Asia/Kolkata

# Expose the port the app runs on
EXPOSE 8081

# Command to run the application
CMD ["/app/server"] 