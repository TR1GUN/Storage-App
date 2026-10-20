FROM golang:1.24-alpine AS builder

WORKDIR /app/bench

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o benchmark .

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/bench/benchmark .

# Run benchmark against specified URL
# Usage: docker run --env BENCHMARK_URL=http://host.docker.internal:8000 benchmark
CMD ["./benchmark"]
