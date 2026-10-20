# =====================
# Go Version (Pure stdlib)
# =====================

FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod ./
RUN go mod download

# Copy application source
COPY . .

# Build
RUN go build -o storage_app .

# Create production image
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create data directories
RUN mkdir -p data/cold

# Copy binary
COPY --from=builder /app/storage_app .

# Expose port
EXPOSE 8000

# Run
CMD ["./storage_app"]
