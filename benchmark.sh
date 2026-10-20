#!/bin/bash
# Quick curl tests for Storage-App (Go version)

SERVER="${1:-http://127.0.0.1:8000}"

echo "========================================"
echo "  Storage-App Quick Tests"
echo "  Server: $SERVER"
echo "========================================"
echo ""

# 1. POST - Create a record
echo "[1/3] POST /records (create record)"
curl -s -X POST "$SERVER/records" \
  -H "Content-Type: application/json" \
  -d '{"idx":"test-001","record":{"data":"Hello World","value":42}}' | python -m json.tool 2>/dev/null || true
echo ""

# 2. GET - Get record by ID
echo "[2/3] GET /records/test-001 (get by ID)"
curl -s "$SERVER/records/test-001" | python -m json.tool 2>/dev/null || true
echo ""

# 3. GET - List all records
echo "[3/3] GET /records (list all)"
curl -s "$SERVER/records" | python -m json.tool 2>/dev/null || true
echo ""

echo "========================================"
echo "  Tests complete."
echo "========================================"
