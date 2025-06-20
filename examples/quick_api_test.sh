#!/bin/bash

# Quick test script for Relay HTTP API
# Run this after starting: ./relay -run examples/simple_test.rl -server

echo "🧪 Quick Relay HTTP API Test"
echo "============================="

BASE_URL="http://localhost:8080"

# Test 1: Simple hello
echo ""
echo "1️⃣ Hello World"
curl -s -X POST "$BASE_URL/rpc" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "test_server.hello",
    "id": 1
  }' | jq -r '.result // .error'

# Test 2: Get initial counter
echo ""
echo "2️⃣ Get Counter (initial)"
curl -s -X POST "$BASE_URL/rpc" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "test_server.get_counter",
    "id": 2
  }' | jq -r '.result // .error'

# Test 3: Increment counter
echo ""
echo "3️⃣ Increment Counter"
curl -s -X POST "$BASE_URL/rpc" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "test_server.increment",
    "id": 3
  }' | jq -r '.result // .error'

# Test 4: Get counter again
echo ""
echo "4️⃣ Get Counter (after increment)"
curl -s -X POST "$BASE_URL/rpc" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "test_server.get_counter",
    "id": 4
  }' | jq -r '.result // .error'

# Test 5: Echo message
echo ""
echo "5️⃣ Echo Message"
curl -s -X POST "$BASE_URL/rpc" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "test_server.echo",
    "params": ["Hello from HTTP!"],
    "id": 5
  }' | jq -r '.result // .error'

# Test 6: Add numbers
echo ""
echo "6️⃣ Add Numbers (5 + 3)"
curl -s -X POST "$BASE_URL/rpc" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "test_server.add",
    "params": [5, 3],
    "id": 6
  }' | jq -r '.result // .error'

echo ""
echo "✅ Test complete!" 