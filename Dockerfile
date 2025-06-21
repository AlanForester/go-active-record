# Use official Go image
FROM golang:1.21-alpine AS builder

# Install necessary packages
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final image
FROM alpine:latest

# Install necessary packages
RUN apk --no-cache add ca-certificates tzdata

# Create user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .

# Change file ownership
RUN chown -R appuser:appgroup /root/

# Switch to non-privileged user
USER appuser

# Expose port (if application uses web server)
EXPOSE 8080

# Default command
CMD ["./main"] 