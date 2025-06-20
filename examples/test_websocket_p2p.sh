#!/bin/bash

# WebSocket P2P Test Script
# This script demonstrates WebSocket peer-to-peer communication between Relay nodes

echo "=== WebSocket P2P Communication Test ==="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test configuration
NODE1_PORT=8080
NODE2_PORT=8081
NODE3_PORT=8082

echo -e "${BLUE}Starting WebSocket P2P test with 3 nodes...${NC}"
echo "Node 1: localhost:$NODE1_PORT"
echo "Node 2: localhost:$NODE2_PORT" 
echo "Node 3: localhost:$NODE3_PORT"
echo

# Function to test HTTP endpoint
test_endpoint() {
    local url=$1
    local description=$2
    echo -e "${YELLOW}Testing: $description${NC}"
    echo "URL: $url"
    
    response=$(curl -s -w "HTTP_CODE:%{http_code}" "$url")
    http_code=$(echo "$response" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
    body=$(echo "$response" | sed 's/HTTP_CODE:[0-9]*$//')
    
    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✓ Success${NC}"
        echo "$body" | jq . 2>/dev/null || echo "$body"
    else
        echo -e "${RED}✗ Failed (HTTP $http_code)${NC}"
        echo "$body"
    fi
    echo
}

# Function to test JSON-RPC call
test_jsonrpc() {
    local url=$1
    local method=$2
    local params=$3
    local description=$4
    
    echo -e "${YELLOW}Testing JSON-RPC: $description${NC}"
    echo "Method: $method"
    
    if [ -n "$params" ]; then
        echo "Params: $params"
        payload=$(jq -n --arg method "$method" --argjson params "$params" '{
            jsonrpc: "2.0",
            method: $method,
            params: $params,
            id: 1
        }')
    else
        payload=$(jq -n --arg method "$method" '{
            jsonrpc: "2.0", 
            method: $method,
            id: 1
        }')
    fi
    
    response=$(curl -s -X POST "$url" \
        -H "Content-Type: application/json" \
        -d "$payload" \
        -w "HTTP_CODE:%{http_code}")
    
    http_code=$(echo "$response" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
    body=$(echo "$response" | sed 's/HTTP_CODE:[0-9]*$//')
    
    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✓ Success${NC}"
        echo "$body" | jq . 2>/dev/null || echo "$body"
    else
        echo -e "${RED}✗ Failed (HTTP $http_code)${NC}"
        echo "$body"
    fi
    echo
}

# Function to test remote server call
test_remote_call() {
    local url=$1
    local node_id=$2
    local server_name=$3
    local method=$4
    local args=$5
    local description=$6
    
    echo -e "${YELLOW}Testing Remote Call: $description${NC}"
    echo "Target Node: $node_id"
    echo "Server: $server_name"
    echo "Method: $method"
    
    if [ -n "$args" ]; then
        params=$(jq -n --arg node_id "$node_id" --arg server_name "$server_name" --arg method "$method" --argjson args "$args" '{
            node_id: $node_id,
            server_name: $server_name,
            method: $method,
            args: $args,
            timeout: 10
        }')
    else
        params=$(jq -n --arg node_id "$node_id" --arg server_name "$server_name" --arg method "$method" '{
            node_id: $node_id,
            server_name: $server_name,
            method: $method,
            timeout: 10
        }')
    fi
    
    payload=$(jq -n --argjson params "$params" '{
        jsonrpc: "2.0",
        method: "remote_call",
        params: $params,
        id: 1
    }')
    
    response=$(curl -s -X POST "$url" \
        -H "Content-Type: application/json" \
        -d "$payload" \
        -w "HTTP_CODE:%{http_code}")
    
    http_code=$(echo "$response" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
    body=$(echo "$response" | sed 's/HTTP_CODE:[0-9]*$//')
    
    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✓ Success${NC}"
        echo "$body" | jq . 2>/dev/null || echo "$body"
    else
        echo -e "${RED}✗ Failed (HTTP $http_code)${NC}"
        echo "$body"
    fi
    echo
}

# Wait for user to start servers
echo -e "${BLUE}Please start the Relay nodes in separate terminals:${NC}"
echo
echo "Terminal 1: ./relay examples/p2p_websocket_demo.rl -server -port $NODE1_PORT -node-id node1"
echo "Terminal 2: ./relay examples/p2p_websocket_demo.rl -server -port $NODE2_PORT -node-id node2"
echo "Terminal 3: ./relay examples/p2p_websocket_demo.rl -server -port $NODE3_PORT -node-id node3"
echo
echo "Press Enter when all nodes are running..."
read

echo -e "${BLUE}Step 1: Testing basic health and info endpoints${NC}"
echo

test_endpoint "http://localhost:$NODE1_PORT/health" "Node 1 Health Check"
test_endpoint "http://localhost:$NODE2_PORT/health" "Node 2 Health Check"  
test_endpoint "http://localhost:$NODE3_PORT/health" "Node 3 Health Check"

test_endpoint "http://localhost:$NODE1_PORT/info" "Node 1 Server Info"
test_endpoint "http://localhost:$NODE2_PORT/registry" "Node 2 Registry Info"

echo -e "${BLUE}Step 2: Testing local server calls${NC}"
echo

test_jsonrpc "http://localhost:$NODE1_PORT/rpc" "counter_server.get_count" "" "Get initial counter value"
test_jsonrpc "http://localhost:$NODE1_PORT/rpc" "counter_server.increment" "[5]" "Increment counter by 5"
test_jsonrpc "http://localhost:$NODE1_PORT/rpc" "counter_server.get_count" "" "Get updated counter value"

test_jsonrpc "http://localhost:$NODE2_PORT/rpc" "discovery_server.get_peers" "" "Get peer list"
test_jsonrpc "http://localhost:$NODE2_PORT/rpc" "discovery_server.health_check" "" "Discovery health check"

test_jsonrpc "http://localhost:$NODE3_PORT/rpc" "task_server.get_stats" "" "Get task server stats"
test_jsonrpc "http://localhost:$NODE3_PORT/rpc" "task_server.add_task" "[\"test_task\", {\"data\": \"test\"}, 1]" "Add a test task"
test_jsonrpc "http://localhost:$NODE3_PORT/rpc" "task_server.get_pending_tasks" "" "Get pending tasks"

echo -e "${BLUE}Step 3: Setting up peer connections${NC}"
echo

# Add peers via HTTP API
echo -e "${YELLOW}Adding peer relationships...${NC}"

# Node 1 connects to Node 2
curl -s -X POST "http://localhost:$NODE1_PORT/registry/peers/add" \
    -H "Content-Type: application/json" \
    -d '{"node_id": "node2", "address": "localhost:8081"}' | jq .

# Node 2 connects to Node 3  
curl -s -X POST "http://localhost:$NODE2_PORT/registry/peers/add" \
    -H "Content-Type: application/json" \
    -d '{"node_id": "node3", "address": "localhost:8082"}' | jq .

# Node 3 connects to Node 1
curl -s -X POST "http://localhost:$NODE3_PORT/registry/peers/add" \
    -H "Content-Type: application/json" \
    -d '{"node_id": "node1", "address": "localhost:8080"}' | jq .

echo
echo "Waiting 5 seconds for WebSocket connections to establish..."
sleep 5

echo -e "${BLUE}Step 4: Testing remote server calls via WebSocket P2P${NC}"
echo

# Note: These calls might fail initially as WebSocket P2P response handling needs to be fully implemented
# The calls will be sent but responses may not be properly routed back yet

test_remote_call "http://localhost:$NODE1_PORT/rpc" "node2" "discovery_server" "get_peers" "" "Call Node 2 discovery from Node 1"

test_remote_call "http://localhost:$NODE2_PORT/rpc" "node3" "task_server" "get_stats" "" "Call Node 3 task server from Node 2"

test_remote_call "http://localhost:$NODE3_PORT/rpc" "node1" "counter_server" "get_count" "" "Call Node 1 counter from Node 3"

test_remote_call "http://localhost:$NODE1_PORT/rpc" "node2" "relay_server" "get_history" "[5]" "Get message history from Node 2"

echo -e "${BLUE}Step 5: Testing multistep routing${NC}"
echo

# Test routing through intermediate nodes
echo -e "${YELLOW}Testing multistep routing (Node 1 -> Node 2 -> Node 3)...${NC}"

# This would require Node 2 to route messages from Node 1 to Node 3
test_remote_call "http://localhost:$NODE1_PORT/rpc" "node3" "task_server" "add_task" "[\"remote_task\", {\"routed\": true}, 2]" "Add task to Node 3 via routing"

echo -e "${BLUE}Step 6: Testing peer discovery and registry${NC}"
echo

test_endpoint "http://localhost:$NODE1_PORT/registry" "Node 1 Full Registry"
test_endpoint "http://localhost:$NODE2_PORT/registry/peers" "Node 2 Peer List"
test_endpoint "http://localhost:$NODE3_PORT/registry/servers" "Node 3 Server List"

echo -e "${BLUE}Step 7: Performance and stress testing${NC}"
echo

echo -e "${YELLOW}Running performance test - multiple rapid calls...${NC}"

for i in {1..5}; do
    echo "Batch $i:"
    test_jsonrpc "http://localhost:$NODE1_PORT/rpc" "counter_server.increment" "[1]" "Increment batch $i" &
    test_jsonrpc "http://localhost:$NODE2_PORT/rpc" "discovery_server.announce" "[\"node2\"]" "Announce batch $i" &
    test_jsonrpc "http://localhost:$NODE3_PORT/rpc" "task_server.add_task" "[\"batch_$i\", {\"batch\": $i}, 1]" "Task batch $i" &
    wait
done

echo -e "${GREEN}WebSocket P2P test completed!${NC}"
echo
echo -e "${BLUE}Summary:${NC}"
echo "✓ Basic HTTP endpoints tested"
echo "✓ Local JSON-RPC calls tested"
echo "✓ Peer connections established"
echo "✓ Remote server calls attempted"
echo "✓ Registry and discovery tested"
echo "✓ Performance testing completed"
echo
echo -e "${YELLOW}Note: Remote server calls may show errors as the WebSocket P2P response${NC}"
echo -e "${YELLOW}handling is still being developed. The infrastructure is in place.${NC}"
echo
echo "Check the server logs for WebSocket connection and message details." 