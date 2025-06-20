# Relay HTTP Server & JSON-RPC 2.0 API

## 🚀 Overview

Relay now includes a built-in HTTP server that exposes your Relay server actors as JSON-RPC 2.0 endpoints. This allows you to build web APIs and microservices using Relay's actor model with full HTTP/JSON interoperability.

## ✨ Features

- **JSON-RPC 2.0 Protocol**: Standards-compliant JSON-RPC 2.0 implementation
- **Actor Integration**: Direct exposure of Relay server actors over HTTP
- **Automatic Type Conversion**: Seamless conversion between JSON and Relay values
- **CORS Support**: Built-in CORS headers for web browser compatibility  
- **Health & Info Endpoints**: Built-in endpoints for monitoring and discovery
- **Error Handling**: Proper HTTP status codes and JSON-RPC error responses
- **Configurable**: Host, port, timeouts, and headers configuration

## 🎯 Quick Start

### 1. Create a Relay Server

```relay
// blog_server.rl
server blog_server {
    state {
        posts: [object] = [],
        next_id: number = 1
    }
    
    receive fn get_posts() -> [object] {
        state.get("posts")
    }
    
    receive fn create_post(title: string, content: string) -> object {
        set id = state.get("next_id")
        set post = {id: id, title: title, content: content}
        
        set posts = state.get("posts")
        state.set("posts", posts.push(post))
        state.set("next_id", id + 1)
        
        post
    }
}
```

### 2. Start HTTP Server

```bash
# Start server on default port (8080)
./relay -run blog_server.rl -server

# Start server on custom port and host
./relay -run blog_server.rl -server -port 9090 -host 127.0.0.1

# Start standalone HTTP server (no Relay file)
./relay -server
```

### 3. Make JSON-RPC Calls

```bash
# Get all posts
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "blog_server.get_posts", 
    "id": 1
  }'

# Create a new post
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "blog_server.create_post",
    "params": ["My Post", "Post content here"],
    "id": 2
  }'
```

## 📡 JSON-RPC 2.0 Protocol

### Request Format

```json
{
  "jsonrpc": "2.0",
  "method": "server_name.method_name",
  "params": [...],
  "id": 1
}
```

### Response Format

**Success:**
```json
{
  "jsonrpc": "2.0",
  "result": {...},
  "id": 1
}
```

**Error:**
```json
{
  "jsonrpc": "2.0", 
  "error": {
    "code": -32601,
    "message": "Method not found",
    "data": "Server 'blog_server' not found"
  },
  "id": 1
}
```

### Parameter Formats

**Positional Parameters:**
```json
{
  "method": "server.method",
  "params": ["arg1", "arg2", 42]
}
```

**Named Parameters:**
```json
{
  "method": "server.method", 
  "params": {"title": "My Post", "content": "Content"}
}
```

## 🔧 Endpoints

### JSON-RPC Endpoint

- **URL**: `POST /rpc`  
- **Content-Type**: `application/json`
- **Purpose**: Execute Relay server methods via JSON-RPC 2.0

### Health Check

- **URL**: `GET /health`
- **Response**: 
  ```json
  {
    "status": "healthy",
    "timestamp": 1640995200,
    "servers": 2
  }
  ```

### Server Info

- **URL**: `GET /info`
- **Response**:
  ```json
  {
    "relay_version": "0.3.0-dev",
    "servers": ["blog_server", "user_server"],
    "endpoints": {
      "rpc": "/rpc",
      "health": "/health", 
      "info": "/info"
    },
    "jsonrpc_version": "2.0"
  }
  ```

## 🔄 Type Conversion

### JSON to Relay

| JSON Type | Relay Type |
|-----------|------------|
| `null` | `nil` |
| `boolean` | `bool` |  
| `number` | `number` |
| `string` | `string` |
| `array` | `[type]` |
| `object` | `object` |

### Relay to JSON

| Relay Type | JSON Type |
|------------|-----------|
| `nil` | `null` |
| `bool` | `boolean` |
| `number` | `number` |
| `string` | `string` |
| `[type]` | `array` |
| `object` | `object` |
| `struct` | `object` (with `_type` field) |
| `function` | `"<function: name>"` |
| `server` | `"<server: name>"` |

## ⚙️ Configuration

### Command Line Options

```bash
./relay -server [options]

Options:
  -host string    Host to bind to (default "0.0.0.0")
  -port int       Port to listen on (default 8080)
```

### Programmatic Configuration

```go
config := &runtime.HTTPServerConfig{
    Host:         "127.0.0.1",
    Port:         9090,
    EnableCORS:   true,
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 15 * time.Second,
    Headers: map[string]string{
        "X-API-Version": "1.0",
    },
}

server := runtime.NewHTTPServer(evaluator, config)
server.Start()
```

## 🧪 Testing

### Run Example

1. **Start the blog server:**
   ```bash
   ./relay -run examples/blog_server.rl -server
   ```

2. **Run the test script:**  
   ```bash
   ./examples/test_http_api.sh
   ```

### Manual Testing

```bash
# Health check
curl http://localhost:8080/health

# Server info  
curl http://localhost:8080/info

# Create post
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "blog_server.create_post", 
    "params": ["Test Post", "This is a test"],
    "id": 1
  }'

# Get posts
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "blog_server.get_posts",
    "id": 2  
  }'
```

## 🚨 Error Codes

| Code | Message | Description |
|------|---------|-------------|
| `-32700` | Parse error | Invalid JSON received |
| `-32600` | Invalid Request | Invalid JSON-RPC format |
| `-32601` | Method not found | Server or method doesn't exist |
| `-32602` | Invalid params | Parameter validation failed |
| `-32603` | Internal error | Server execution error |

## 🔒 Security Considerations

- **CORS**: CORS is enabled by default (`*` origin)
- **Input Validation**: All JSON-RPC parameters are validated
- **Error Isolation**: Server errors don't crash the HTTP server
- **Timeouts**: Built-in request/response timeouts
- **Production**: Consider adding authentication, rate limiting, etc.

## 🎯 Use Cases

### Microservices
```relay
server user_service {
    receive fn get_user(id: string) -> object { ... }
    receive fn create_user(data: object) -> object { ... }
}

server auth_service {  
    receive fn login(username: string, password: string) -> object { ... }
    receive fn validate_token(token: string) -> bool { ... }
}
```

### Web APIs
```relay
server api_server {
    receive fn handle_webhook(payload: object) -> object { ... }
    receive fn process_payment(data: object) -> object { ... }
}
```

### Real-time Services
```relay
server notification_service {
    receive fn send_notification(user_id: string, message: string) -> bool { ... }
    receive fn get_notifications(user_id: string) -> [object] { ... }
}
```

## 🧬 Architecture

```
HTTP Request → JSON-RPC Parser → Method Router → Actor System → JSON Response
     ↓               ↓                ↓             ↓            ↑
  JSON/HTTP     JSON-RPC 2.0    Server.Method   Message     Value → JSON
```

## 📚 Advanced Examples

### Complex Parameter Handling
```relay
server complex_server {
    receive fn process_data(config: object, items: [object]) -> object {
        // Process complex nested data structures
        set results = items.map(fn(item) {
            // Transform each item based on config
            transform_item(item, config)
        })
        {status: "success", results: results}
    }
}
```

### Error Handling
```relay
server error_server {
    receive fn risky_operation(data: object) -> object {
        if data.get("invalid") {
            // This will return nil, converted to JSON null
            nil
        } else {
            {success: true, data: data}
        }
    }
}
```

## 🔮 Future Enhancements

- **WebSocket Support**: Real-time bidirectional communication
- **Authentication**: Built-in auth middleware  
- **Rate Limiting**: Request rate limiting
- **Metrics**: Built-in metrics and monitoring
- **OpenAPI**: Automatic OpenAPI/Swagger generation
- **Batch Requests**: JSON-RPC batch request support

---

**🎉 Ready to build web APIs with Relay's actor model!** 