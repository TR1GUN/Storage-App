#!/bin/bash
# Start Storage-App (Go version)

set -e

echo "🚀 Starting Storage-App (Go) on :8000..."

# Create data directories
mkdir -p data/cold

# Build and run
go build -o storage_app.exe . && ./storage_app.exe
