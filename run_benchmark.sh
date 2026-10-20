#!/bin/bash
# Run benchmark against Storage-App (Go version)

set -e

cd "$(dirname "$0")/benchmark"

echo "📊 Running benchmark against Storage-App (Go)..."
echo ""

go run .
