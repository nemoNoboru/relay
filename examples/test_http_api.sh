#!/bin/bash

# Test script for Relay HTTP JSON-RPC API
# Requires the blog server to be running: ./relay -run examples/blog_server.rl -server

BASE_URL="http://localhost:8080"
CONTENT_TYPE="application/json"

echo "🚀 Testing Relay HTTP JSON-RPC API"
echo "=================================="

# Test 1: Health check
echo ""
echo "📋 1. Health Check"
curl -s "$BASE_URL/health" | jq .

# Test 2: Server info
echo ""
echo "📋 2. Server Info"
curl -s "$BASE_URL/info" | jq .

# Test 3: Get initial posts (should be empty)
echo ""
echo "📋 3. Get Initial Posts"
curl -s -X POST "$BASE_URL/rpc" \
  -H "Content-Type: $CONTENT_TYPE" \
  -d '{
    "jsonrpc": "2.0",
    "method": "blog_server.get_posts",
    "id": 1
  }' | jq .

# Test 4: Create a new post
echo ""
echo "📋 4. Create New Post"
curl -s -X POST "$BASE_URL/rpc" \
  -H "Content-Type: $CONTENT_TYPE" \
  -d '{
    "jsonrpc": "2.0",
    "method": "blog_server.create_post",
    "params": ["My First Post", "This is the content of my first blog post!"],
    "id": 2
  }' | jq .

# Test 5: Create another post
echo ""
echo "📋 5. Create Second Post"
curl -s -X POST "$BASE_URL/rpc" \
  -H "Content-Type: $CONTENT_TYPE" \
  -d '{
    "jsonrpc": "2.0", 
    "method": "blog_server.create_post",
    "params": ["Second Post", "This is my second post with more content."],
    "id": 3
  }' | jq .

# Test 6: Get all posts
echo ""
echo "📋 6. Get All Posts"
curl -s -X POST "$BASE_URL/rpc" \
  -H "Content-Type: $CONTENT_TYPE" \
  -d '{
    "jsonrpc": "2.0",
    "method": "blog_server.get_posts",
    "id": 4
  }' | jq .

# Test 7: Get specific post
echo ""
echo "📋 7. Get Specific Post (ID: 1)"
curl -s -X POST "$BASE_URL/rpc" \
  -H "Content-Type: $CONTENT_TYPE" \
  -d '{
    "jsonrpc": "2.0",
    "method": "blog_server.get_post",
    "params": [1],
    "id": 5
  }' | jq .

# Test 8: Update post
echo ""
echo "📋 8. Update Post (ID: 1)"
curl -s -X POST "$BASE_URL/rpc" \
  -H "Content-Type: $CONTENT_TYPE" \
  -d '{
    "jsonrpc": "2.0",
    "method": "blog_server.update_post",
    "params": [1, "Updated First Post", "This is the updated content!"],
    "id": 6
  }' | jq .

# Test 9: Get stats
echo ""
echo "📋 9. Get Server Stats"
curl -s -X POST "$BASE_URL/rpc" \
  -H "Content-Type: $CONTENT_TYPE" \
  -d '{
    "jsonrpc": "2.0",
    "method": "blog_server.get_stats",
    "id": 7
  }' | jq .

# Test 10: Delete post
echo ""
echo "📋 10. Delete Post (ID: 2)"
curl -s -X POST "$BASE_URL/rpc" \
  -H "Content-Type: $CONTENT_TYPE" \
  -d '{
    "jsonrpc": "2.0",
    "method": "blog_server.delete_post",
    "params": [2],
    "id": 8
  }' | jq .

# Test 11: Get posts after deletion
echo ""
echo "📋 11. Get Posts After Deletion"
curl -s -X POST "$BASE_URL/rpc" \
  -H "Content-Type: $CONTENT_TYPE" \
  -d '{
    "jsonrpc": "2.0",
    "method": "blog_server.get_posts",
    "id": 9
  }' | jq .

# Test 12: Error handling - invalid method
echo ""
echo "📋 12. Error Handling - Invalid Method"
curl -s -X POST "$BASE_URL/rpc" \
  -H "Content-Type: $CONTENT_TYPE" \
  -d '{
    "jsonrpc": "2.0",
    "method": "blog_server.nonexistent_method",
    "id": 10
  }' | jq .

echo ""
echo "✅ API Testing Complete!" 