# Relay Language - Complete Technical Specification

🎉 **MAJOR MILESTONES**: Full closure support, first-class functions, struct system, and unified runtime architecture!

**Version:** 0.3 "Cloudpunks Edition"  
**By:** Cloudpunks  
**Mission:** A federated, minimal language to build modern distributed web services with the simplicity of early PHP.

## 🔧 Implementation Status

**✅ FULLY IMPLEMENTED:**
- ✅ Complete expression evaluation system (739 tests passing)
- ✅ First-class functions with closure support
- ✅ Struct definitions and instantiation
- ✅ In-memory server actors with message passing
- ✅ Thread-safe server state management via sequential message processing (memory-only)
- ✅ Immutable-by-default semantics with method-based operations
- ✅ Array and object operations with method chaining
- ✅ Binary operations (arithmetic, logical, comparisons)
- ✅ Variable scoping and environment management
- ✅ Built-in functions and type system
- ✅ Unified runtime architecture optimized for simplicity

**🚧 PARTIALLY IMPLEMENTED:**
- 🚧 Command-line interface (REPL and file execution work, build mode planned)
- 🚧 Template parsing (declarations parsed, rendering not implemented)

**📋 PLANNED/TODO:**
- 📋 HTTP server infrastructure with JSON-RPC 2.0 endpoints
- 📋 Server state persistence (embedded NoSQL database)
- 📋 Template rendering engine with server-side rendering
- 📋 Federation protocol for service discovery and communication
- 📋 Load balancing system with health checking
- 📋 Go Actor runtime for I/O and system operations

---

## 1. Language Philosophy

Relay is designed for federated, self-hostable, and functional applications. Each script is:
- **Currently**: In-memory actor system with message passing
- **Planned**: JSON-RPC server by default
- **Currently**: Stateless by design with in-memory state management
- **Planned**: Automatic persistence with embedded database
- **Currently**: Able to send and receive messages between actors
- **Planned**: Template rendering and HTTP endpoints
- Built for high interoperability and zero config deployment
- BASIC-simple for hobbyists, enterprise-ready for production
- ✨ **IMPLEMENTED**: Fully functional with closures and first-class functions for advanced programming patterns
- ✨ **IMPLEMENTED**: Complete struct system with definitions, instantiation, and field access

---

## 2. Reserved Keywords

```
struct, protocol, server, state, receive, send, template, dispatch,
config, fn, if, else, for, in, return, throw
```

**Note:** All keywords are strictly reserved and cannot be used as identifiers.

---

## 3. Data Types

### Primitive Types
```relay
string         // UTF-8 strings
number         // 64-bit floating point
bool           // true/false
datetime       // ISO 8601 timestamps
```

### Collection Types
```relay
[type]         // Arrays
{key: type}    // Objects/Maps
optional(type) // Nullable types
```

### Symbol Literals

Relay supports symbol literals as syntactic sugar for strings:

```relay
:hello         // Equivalent to "hello"
:get_posts     // Equivalent to "get_posts"
:default       // Equivalent to "default"

// Symbols follow identifier rules: letters, numbers, underscores only
:valid_symbol_123  // Valid
:invalid-symbol    // Invalid (contains hyphen)
```

### Immutable Data Operations

Relay is immutable-by-default. All data operations return new instances rather than modifying existing ones.

#### Object Methods
```relay
// Field access is always a method call
user.get("name")                    // Get field value
user.set("name", "John")           // Set single field
user.merge({name: "John", age: 25}) // Merge multiple fields

// Method chaining
user.set("name", "John")
    .set("email", "john@test.com") 
    .set("active", true)
```

#### Array Methods
```relay
// Addition
users.add(new_user)              // Add single item to end
users.append(more_users)         // Add multiple items to end  
users.prepend(first_user)        // Add single item to beginning
users.insert(2, user)            // Insert at specific index

// Removal
users.remove(user)               // Remove by value
users.remove_at(0)               // Remove by index
users.filter(fn (u) { u.get("active") }) // Remove by condition

// Transformation
users.map(fn (u) { u.set("active", true) })  // Transform each item
users.sort_by(fn (u) { u.get("name") })      // Sort by field
users.reverse()                              // Reverse order
users.replace(old_user, new_user)            // Replace specific item
```

---

## 4. Core Language Constructs

### 4.1 Expression-Based Design

**Everything is an expression** - including blocks, control flow, and statements. Blocks evaluate to their last expression.

```relay
// Block as expression
set result = {
  set x = 10
  set y = 20
  x + y  // Block evaluates to 30
}

// If as expression
set message = if count > 0 { "items found" } else { "no items" }

// Early return from blocks
set result = {
  if error_condition {
    return "error"  // Returns from this block only
  }
  "success"
}
```

### 4.2 Function Definitions

🎉 **MAJOR MILESTONE**: Functions in Relay are **closures** and **first-class citizens** with full lexical scoping!

All functions use the `fn` keyword with consistent syntax:

```relay
// Named functions
fn calculate_total(items: [object]) -> number {
  items.map(fn (item) { item.get("price") })
       .reduce(fn (acc, price) { acc + price }, 0)
}

fn format_date(date: datetime, format: string) -> string {
  date.format(format)
}

// Higher-order functions
fn apply_discount(calculator: fn(number) -> number, price: number) -> number {
  calculator(price)
}

// Lambda expressions (anonymous functions)
set numbers = [1, 2, 3, 4, 5]
set doubled = numbers.map(fn (x) { x * 2 })
set filtered = numbers.filter(fn (x) { x > 3 })

// Multi-line lambdas
set complex_transform = numbers.map(fn (x) {
  set doubled = x * 2
  doubled + 10
})
```

#### Closure Support (NEW!)

**Functions are true closures** that capture their lexical environment:

```relay
set captured_value = 42

fn get_captured() {
  captured_value  // ✅ Accesses outer scope variable
}

get_captured()  // Returns 42

// Closure with state preservation
fn make_counter(start) {
  set count = start
  fn() { 
    set count = count + 1
    count 
  }
}

set counter = make_counter(10)
counter()  // Returns 11
counter()  // Returns 12 - state preserved between calls!
```

#### First-Class Functions (NEW!)

**Functions can be stored, passed, and returned like any other value:**

```relay
// Functions as variables
fn double(x) { x * 2 }
set my_operation = double
my_operation(5)  // Returns 10

// Functions as arguments
fn apply_twice(operation, value) {
  operation(operation(value))
}

apply_twice(double, 3)  // Returns 12

// Functions as return values
fn make_multiplier(factor) {
  fn(x) { x * factor }  // Returns a new function
}

set times_5 = make_multiplier(5)
times_5(3)  // Returns 15

// Complex function composition
fn compose(f, g) {
  fn(x) { f(g(x)) }
}

fn square(x) { x * x }
fn increment(x) { x + 1 }
set square_then_increment = compose(increment, square)
square_then_increment(3)  // Returns 10 (3² + 1)
```

#### Functional Programming Patterns

With first-class functions and closures, Relay supports advanced functional programming:

```relay
// Map/filter/reduce with custom functions
fn is_active(user) { user.get("active") }
fn get_name(user) { user.get("name") }

set active_names = users
  .filter(is_active)
  .map(get_name)

// Partial application through closures
fn make_filter(predicate) {
  fn(collection) { collection.filter(predicate) }
}

set filter_active = make_filter(is_active)
set active_users = filter_active(users)

// Strategy pattern with functions
fn process_with_strategy(data, strategy) {
  strategy(data)
}

process_with_strategy(users, filter_active)
```

### 4.3 Struct Definitions

✨ **IMPLEMENTED**: Struct definitions and instantiation now fully supported in the runtime!

```relay
struct User {
  username: string,
  email: string,
  bio: optional(string),
  created_at: datetime
}

struct Post {
  id: string,
  title: string,
  content: string,
  author: string,
  tags: [string],
  created_at: datetime
}
```

#### Struct Instantiation and Usage

**Creating struct instances:**
```relay
// Create a User instance
set user = User{
  username: "john_doe",
  email: "john@example.com",
  bio: "Software developer",
  created_at: now()
}

// Field order doesn't matter
set user2 = User{
  email: "jane@example.com",
  username: "jane_smith",
  created_at: now(),
  bio: "Designer"
}
```

**Accessing struct fields:**
```relay
// Access fields using .get() method
set name = user.get("username")     // Returns "john_doe"
set email = user.get("email")       // Returns "john@example.com"

// Use in expressions
set greeting = "Hello, " + user.get("username")
set isRecent = user.get("created_at") > yesterday()
```

**Struct equality and comparison:**
```relay
set user1 = User{ username: "john", email: "john@test.com" }
set user2 = User{ username: "john", email: "john@test.com" }
set user3 = User{ email: "john@test.com", username: "john" }

user1 == user2  // Returns true (same values)
user1 == user3  // Returns true (field order doesn't matter)
```

**Complex struct operations:**
```relay
// Create a post with user information
set post = Post{
  id: generateId(),
  title: "My First Post",
  content: "Hello, world!",
  author: user.get("username"),
  tags: ["intro", "hello"],
  created_at: now()
}

// Use struct fields in functions
fn getUserSummary(u: User) -> string {
  u.get("username") + " (" + u.get("email") + ")"
}

set summary = getUserSummary(user)
```

### 4.4 Protocol Definitions

```relay
protocol BlogService {
  get_posts() -> [Post]
  create_post(title: string, content: string) -> Post
  get_post(id: string) -> Post
  delete_post(id: string) -> bool
}

protocol UserService {
  create_user(user: User) -> User
  get_user(username: string) -> User
  update_user(username: string, updates: object) -> User
}
```

### 4.5 Server Definitions

Every Relay program can contain multiple servers. All defined servers are automatically started and mounted on startup.

```relay
server blog_service {
  state {
    posts: [Post] = [],
    next_id: number = 1
  }
  
  receive fn get_posts() -> [Post] {
    state.get("posts")
  }
  
  receive fn create_post(title: string, content: string) -> Post {
    set post = Post{
      id: state.get("next_id").toString(),
      title: title,
      content: content,
      author: "Anonymous",
      tags: [],
      created_at: now()
    }
    
    state.set("posts", state.get("posts").add(post))
    state.set("next_id", state.get("next_id") + 1)
    
    post
  }
  
  receive fn get_post(id: string) -> Post {
    set post = state.get("posts").find(fn (p) { p.get("id") == id })
    if post {
      post
    } else {
      throw {error: "Post not found", id: id}
    }
  }
  
  receive fn delete_post(id: string) -> bool {
    set initial_length = state.get("posts").length
    state.set("posts", state.get("posts").filter(fn (p) { p.get("id") != id }))
    state.get("posts").length < initial_length
  }
}

server user_service {
  state {
    users: [User] = [],
    next_id: number = 1
  }
  
  receive fn create_user(user: User) -> User {
    set new_user = user.set("created_at", now())
    state.set("users", state.get("users").add(new_user))
    new_user
  }
  
  receive fn get_user(username: string) -> User {
    set user = state.get("users").find(fn (u) { u.get("username") == username })
    if user {
      user
    } else {
      throw {error: "User not found", username: username}
    }
  }
}
```

### 4.6 State Management

**Current Implementation:** Servers are in-memory actors that maintain thread-safe state during runtime. State is lost on restart (persistence planned for future implementation).

State is modified using method calls for consistent immutable semantics:

```relay
server counter_service {
  state {
    count: number = 0,
    last_updated: datetime = now(),
    history: [object] = []
  }
  
  receive fn increment(amount: optional(number)) -> number {
    set increment_amount = amount ?? 1
    state.set("count", state.get("count") + increment_amount)
    state.set("last_updated", now())
    state.set("history", state.get("history").add({
      action: "increment",
      amount: increment_amount,
      timestamp: state.get("last_updated")
    }))
    state.get("count")
  }
  
  receive fn get_stats() -> object {
    {
      count: state.get("count"),
      last_updated: state.get("last_updated"),
      total_operations: state.get("history").length
    }
  }
}
```

### 4.7 Send/Receive Communication

```relay
// Local server call
set posts = send "blog_service" get_posts {}

// Remote server call  
set posts = send "blog.example.com" get_posts {}

// Federated call through relay
set posts = send "relay.community.com" proxy {
  app: "blog",
  method: "get_posts", 
  params: {}
}

// With built-in load balancing
// If multiple blog_service instances exist, automatically load-balanced
set new_post = send "blog_service" create_post {
  title: "Hello World",
  content: "My first post"
}
```

### 4.8 Template System (PHP-style Web Hosting)

**Revolutionary Feature:** The `template` keyword performs automatic send + render in one step.

```relay
// Template declarations - simplified function call syntax
template "homepage.html" from get_posts()
template "post.html" from get_post(id: string)
template "user-profile.html" from get_user(username: string)

// JSON API endpoints
template "api/posts.json" from get_posts()
template "api/user.json" from get_user(username: string)
```

**Template Files:**

**homepage.html:**
```html
<!DOCTYPE html>
<html>
<head>
    <title>My Blog</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <h1>Recent Posts</h1>
    
    @if(posts.length > 0) {
        @for(post in posts) {
            <article>
                <h2>@{post.get("title")}</h2>
                <p>@{post.get("content")}</p>
                <small>By @{post.get("author")} on @{format_date(post.get("created_at"), "YYYY-MM-DD")}</small>
            </article>
        }
    } else {
        <p>No posts yet.</p>
    }
    
    <a href="/create-post.html">Create New Post</a>
</body>
</html>
```

**api/posts.json:**
```json
{
    "posts": @{posts},
    "count": @{posts.length},
    "generated_at": "@{now().toISOString()}"
}
```

### 4.9 Dispatch Pattern Matching

Dispatch uses lambda functions for cleaner, more consistent syntax:

```relay
receive fn handle_activity(activity: Activity) -> object {
  dispatch activity.get("type") {
    :create: fn (data) {
      send "blog_service" create_post {
        title: activity.get("object").get("title"),
        content: activity.get("object").get("content")
      }
    },
    
    :delete: fn (data) {
      send "blog_service" delete_post {
        id: activity.get("object").get("id")
      }
    },
    
    :follow: fn (data) {
      send "user_service" add_follower {
        username: activity.get("object").get("username"),
        follower: activity.get("actor")
      }
    },
    
    :default: fn (_) {
      throw {error: "Unknown activity type", type: activity.get("type")}
    }
  }
}
```

### 4.10 Configuration

```relay
config {
  app_name: "myblog",
  relay: "community.relay.com", 
  port: 8080,
  
  federation: {
    auto_discover: true,
    known_relays: ["relay1.com", "relay2.com"],
    cache_duration: "1h"
  },
  
  load_balancing: {
    strategy: "round_robin",  // round_robin, random, lowest_latency, sticky
    health_check_interval: "30s"
  },
  
  persistence: {
    backend: "embedded",  // embedded, redis, postgres
    sync_interval: "1s"
  },
  
  templates: {
    static_dir: "./static",
    template_dir: "./templates",
    auto_reload: true
  },
  
  external_services: {
    analytics: "https://analytics.company.com",
    email: "https://email-service.com"
  }
}
```

---

## 5. Simplified Grammar Rules

### 5.1 Consistent Syntax Patterns

- **All parameters:** Use `{}` for arguments in calls
- **All field access:** Use `.get("field")` method calls
- **All functions:** Use `fn name(params) { block }` format
- **All lambdas:** Use `fn (params) { block }` format
- **All symbols:** Use `:identifier` as shorthand for `"identifier"`
- **All blocks:** Are expressions that evaluate to their last expression
- **All state updates:** Use `state.set("field", expression)` method calls

### 5.2 Expression Precedence (Simplified)

```
primary: literal | identifier | (expression) | {object} | [array] | fn(params){block}
postfix: primary (.method{args} | [index])*
unary: (!|-) postfix
multiplicative: unary ((*|/) unary)*
additive: multiplicative ((+|-) multiplicative)*
relational: additive ((<|>|<=|>=) additive)*
equality: relational ((==|!=) relational)*
logical_and: equality (&& equality)*
logical_or: logical_and (|| logical_and)*
```

### 5.3 Object Creation Rules

```relay
// Struct constructors (type-checked)
set user = User{
  username: "john",
  email: "john@example.com",
  created_at: now()
}

// Object literals (any values)
set config = {
  debug: true,
  timeout: 30,
  endpoints: ["api1", "api2"]
}
```

---

## 6. Complete Example Application

**main.relay:**
```relay
struct Post {
  id: string,
  title: string,
  content: string,
  author: string,
  created_at: datetime
}

protocol BlogService {
  get_posts() -> [Post]
  create_post(title: string, content: string) -> Post
}

server blog_service {
  state {
    posts: [Post] = [],
    next_id: number = 1
  }
  
  receive fn get_posts() -> [Post] {
    state.get("posts")
  }
  
  receive fn create_post(title: string, content: string) -> Post {
    set post = Post{
      id: state.get("next_id").toString(),
      title: title,
      content: content,
      author: "Anonymous",
      created_at: now()
    }
    
    state.set("posts", state.get("posts").add(post))
    state.set("next_id", state.get("next_id") + 1)
    
    post
  }
}

// Web routes
template "index.html" from get_posts()
template "api/posts.json" from get_posts()

config {
  app_name: "simple_blog",
  port: 8080,
  load_balancing: {
    strategy: "round_robin"
  }
}
```

**index.html:**
```html
<!DOCTYPE html>
<html>
<head>
    <title>Simple Blog</title>
</head>
<body>
    <h1>My Blog</h1>
    
    <form action="/api/create_post" method="post">
        <input type="text" name="title" placeholder="Post title" required>
        <textarea name="content" placeholder="Post content" required></textarea>
        <button type="submit">Create Post</button>
    </form>
    
    <div id="posts">
        @for(post in posts) {
            <article>
                <h2>@{post.get("title")}</h2>
                <p>@{post.get("content")}</p>
                <small>@{format_date(post.get("created_at"), "YYYY-MM-DD HH:mm")}</small>
            </article>
        }
    </div>
</body>
</html>
```

---

## 7. Runtime Specification

### 7.1 In-Memory Actor System (Current Implementation)

Relay currently implements an in-memory actor system where servers are concurrent actors that communicate through message passing:

**Server Architecture:**
- Each server runs in its own goroutine
- Thread-safe state management using `sync.RWMutex`
- Message-based communication between actors
- Automatic lifecycle management (start/stop)

**Message Format:**
```go
type Message struct {
    Method string      // The receive function to call
    Args   []*Value    // Arguments for the function  
    Reply  chan *Value // Channel for reply (optional)
}
```

**Actor Communication:**
```relay
// Send message to server actor
set result = send(server_name, method_name, arg1, arg2)

// Built-in message function (alternative syntax)
set result = message(server_name, method_name, arg1, arg2)
```

### 7.2 JSON-RPC 2.0 HTTP Server (Planned)

**Future Implementation:** HTTP endpoints with JSON-RPC 2.0 protocol:

**Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "method_name",
  "params": {...},
  "id": 1
}
```

**Response:**
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
    "code": -32000,
    "message": "Application error",
    "data": {...}
  },
  "id": 1
}
```

### 7.3 State Management Runtime (Current Implementation)

- **Memory-Only**: State is stored in memory using Go maps with thread-safe access
- **Threading**: State mutations are atomic due to single-threaded message processing per actor
- **Immutability**: All state updates create new instances via method calls
- **Isolation**: Each server instance maintains separate state namespace
- **Lifecycle**: State exists only during program execution (lost on restart)

**Planned State Persistence:**
- **Persistence**: Automatic persistence using embedded NoSQL database (similar to SQLite)
- **Durability**: All state changes will be immediately written to disk
- **Recovery**: State will be restored from database on server restart

### 7.4 Template System

**Current Implementation (Parsing Only):**
- **Declaration Parsing**: Template declarations are parsed and stored in AST
- **Function Binding**: Templates are linked to data-providing functions
- **Syntax Support**: Template syntax is recognized in parser

**Planned Template Rendering Engine:**
- **Server-side rendering**: Templates will be rendered on the server
- **State access**: Templates will have access to data returned from send calls
- **HTTP routes**: Template declarations will automatically create HTTP routes
- **Content-type detection**: Automatic content-type based on file extension
- **Template syntax**: `@{expression}`, `@if(condition) {...}`, `@for(item in items) {...}`
- **Error handling**: Failed sends will result in configurable error pages
- **URL Parameters**: Template function parameters will be extracted from URL paths

### 7.5 Load Balancing System (Planned)

**Future Implementation - Load Balancing Strategies:**
- `round_robin` (default): Distribute requests evenly
- `random`: Random selection
- `lowest_latency`: Route to fastest responding server
- `sticky`: Session-based routing using hashed keys

**Planned Health Checking:**
- Automatic health checks for all registered servers
- Failed servers temporarily removed from pool
- Automatic recovery detection

### 7.6 Federation Protocol (Planned)

**Future Implementation - App Discovery:**
```json
{
  "method": "find_app", 
  "params": {"name": "myblog"}
}
```

**App Proxying:**
```json
{
  "method": "proxy",
  "params": {
    "app": "myblog",
    "method": "get_posts",
    "params": {}
  }
}
```

### 7.7 Federation and Mesh Networking (New)

**Core Architecture - Relay Proxy Actor:**

Relay instances include a **Federation Proxy Actor** written in pure Go that enables distributed communication, mesh networking, and interoperability between Relay instances and external services.

**Purpose:**
- Handle JSON-RPC calls from external Relay instances
- Route method calls to correct target servers within the instance
- Proxy calls to external Relay instances and services
- Manage WebSocket connections for peer-to-peer communication
- Handle HTTP requests and responses for local service interop
- Maintain peer lists for mesh routing and federation

**Federation Message Flow:**

When a Relay server sends a message to an external target:

```relay
// Server code - sends to external instance
send "blog.world.relay" hello_world {}
```

**Proxy Actor Processing:**
1. **Target Resolution**: Checks if `blog.world.relay` is outside current instance
2. **External Request**: Sends JSON-RPC request to `blog.world.relay`
3. **Response Handling**: Waits for reply from external instance
4. **Result Proxying**: Returns response to originating server

**JSON-RPC Protocol:**

External communication uses JSON-RPC 2.0 format:

```json
// Request to external instance
{
  "jsonrpc": "2.0",
  "method": "hello_world",
  "params": {},
  "id": "relay-123",
  "meta": {
    "source_instance": "local.relay",
    "target_server": "blog.world.relay",
    "timestamp": "2024-01-01T12:00:00Z"
  }
}

// Response from external instance
{
  "jsonrpc": "2.0",
  "result": {
    "message": "Hello from blog server!",
    "server_info": "blog.world.relay v1.2.3"
  },
  "id": "relay-123"
}
```

**Mesh Networking with Peer Routing:**

The Federation Proxy Actor supports multi-hop routing through peer networks:

**Configuration:**
```relay
config {
  federation: {
    instance_name: "kaz.local.relay",
    peers: [
      {
        name: "world.public.relay",
        endpoint: "https://world.relay.net:8443/jsonrpc",
        type: "websocket",
        trust_level: "public"
      },
      {
        name: "team.private.relay", 
        endpoint: "wss://team.internal:9443/ws",
        type: "websocket",
        trust_level: "private"
      }
    ],
    routing: {
      default_peer: "world.public.relay",
      timeout: 30,
      retry_attempts: 3
    }
  }
}
```

**Multi-hop Routing Example:**

```relay
// Instance 3 wants to call kaz.potemkin.relay
send "kaz.potemkin.relay" run {command: "execute"}
```

**Routing Process:**
1. **Instance 3 Proxy**: Checks local routing table
2. **Peer Lookup**: Finds `world.public.relay` as route to `kaz.*`
3. **First Hop**: WebSocket request to `world.public.relay`
4. **World Proxy**: Routes to `kaz.local.relay` via direct connection
5. **Target Execution**: `kaz` executes method on `potemkin` server
6. **Response Chain**: Result flows back through WebSocket chain

**WebSocket Federation Protocol:**

Long-lived connections between instances use WebSocket with structured messaging:

```json
// Peer connection handshake
{
  "type": "federation_handshake",
  "instance_name": "kaz.local.relay",
  "protocol_version": "1.0",
  "capabilities": ["method_proxy", "service_discovery", "health_check"],
  "auth_token": "signed-jwt-token"
}

// Method proxy request
{
  "type": "method_proxy",
  "request_id": "req-456",
  "target": "potemkin.relay",
  "method": "run",
  "params": {"command": "execute"},
  "source": "instance3.relay",
  "hop_count": 1,
  "max_hops": 5
}

// Health check ping
{
  "type": "health_check",
  "timestamp": "2024-01-01T12:00:00Z",
  "instance_status": "healthy",
  "load_average": 0.2
}
```

**HTTP Interoperability:**

The Proxy Actor can interface with local HTTP services and reverse proxy to other Relay instances:

```relay
// Local HTTP service integration
config {
  http_services: {
    "database": {
      endpoint: "http://localhost:5432/api",
      auth: "bearer",
      token_env: "DB_API_TOKEN"
    },
    "cache": {
      endpoint: "http://redis.local:6379/api",
      auth: "none"
    }
  }
}

// Call local service and proxy result
fn get_user_data(user_id: string) -> object {
  set db_result = send(:http_proxy, {
    service: "database",
    method: "GET",
    path: "/users/" + user_id
  })
  
  // Forward to another Relay instance
  send "analytics.remote.relay" track_user_access {
    user_id: user_id,
    timestamp: now(),
    source: "local"
  }
  
  db_result.data
}
```

**Service Discovery:**

Instances can discover available services through the federation network:

```relay
// Query available services
set services = send(:federation_proxy, {
  action: "discover_services",
  type: "blog",
  region: "us-west"
})

// Result format
{
  services: [
    {
      instance: "blog1.world.relay",
      endpoint: "https://blog1.world.relay.net:8443",
      capabilities: ["read_posts", "write_posts", "admin"],
      load: 0.3,
      latency: 45
    },
    {
      instance: "blog2.world.relay", 
      endpoint: "https://blog2.world.relay.net:8443",
      capabilities: ["read_posts"],
      load: 0.1,
      latency: 32
    }
  ]
}
```

**Security and Trust:**

Federation supports multiple trust levels and authentication:

```relay
config {
  federation: {
    security: {
      default_trust: "none",
      trusted_instances: [
        "*.internal.relay",
        "team.company.relay"
      ],
      auth_methods: ["jwt", "mutual_tls"],
      rate_limiting: {
        requests_per_minute: 1000,
        burst_size: 100
      }
    }
  }
}
```

**Implementation Details:**

The Federation Proxy Actor runs as a separate goroutine within each Relay instance:

- **Concurrent Processing**: Handles multiple federation requests simultaneously
- **Connection Pooling**: Maintains persistent WebSocket connections to peers
- **Circuit Breaker**: Automatically handles failed peer connections
- **Load Balancing**: Routes requests to least-loaded available instances
- **Caching**: Caches routing information and service discovery results
- **Monitoring**: Provides metrics on federation health and performance

This federation system enables Relay to function as a true distributed mesh network where instances can seamlessly communicate, route requests through peers, and integrate with existing HTTP services while maintaining security and performance.

### 7.8 Go Actor Runtime (Planned)

**Future System Architecture:**

Relay will include a dedicated **Go Actor** that runs alongside all server actors to handle I/O operations and complex system tasks that require direct access to the underlying Go runtime capabilities.

**Purpose:**
- Handle file system operations (read, write, delete files)
- Manage network requests (HTTP clients, API calls)
- Execute system commands and shell operations
- Perform complex computations that benefit from Go's performance
- Manage external service integrations (databases, message queues, etc.)
- Handle low-level operations not suitable for Relay's high-level abstractions

**Message Passing Protocol:**

The Go Actor communicates with Relay servers through a structured message-passing system:

```relay
// Sending messages to Go Actor
set file_content = send(:go_actor, {
  action: "read_file",
  path: "/data/config.json"
})

set api_response = send(:go_actor, {
  action: "http_request",
  method: "GET",
  url: "https://api.example.com/data",
  headers: {
    "Authorization": "Bearer " + auth_token
  }
})

// File operations
send(:go_actor, {
  action: "write_file",
  path: "/data/output.txt",
  content: "Hello, world!"
})

// System commands
set command_result = send(:go_actor, {
  action: "exec_command",
  command: "ls",
  args: ["-la", "/tmp"],
  timeout: 30
})
```

**Response Format:**

The Go Actor returns standardized responses:

```relay
// Successful operation
{
  status: "success",
  data: { ... },           // Operation-specific data
  timestamp: "2024-01-01T12:00:00Z"
}

// Error response
{
  status: "error",
  error_code: "FILE_NOT_FOUND",
  message: "File /data/config.json does not exist",
  timestamp: "2024-01-01T12:00:00Z"
}
```

**Concurrency and Safety:**

- The Go Actor runs as a separate goroutine with its own message queue
- All operations are thread-safe and can be called concurrently from multiple Relay servers
- Long-running operations support timeout mechanisms
- The actor maintains operation logging for debugging and monitoring
- Automatic retry logic for transient failures (configurable per operation type)

**Built-in Operations:**

The Go Actor provides a standard library of operations:

- **File System**: `read_file`, `write_file`, `delete_file`, `list_directory`, `create_directory`
- **Network**: `http_request`, `tcp_connect`, `udp_send`
- **System**: `exec_command`, `get_env_var`, `set_env_var`
- **Crypto**: `hash_data`, `encrypt`, `decrypt`, `generate_uuid`
- **Database**: `sql_query`, `nosql_operation` (driver-dependent)

**Configuration:**

```relay
config {
  go_actor: {
    enabled: true,
    max_concurrent_operations: 100,
    default_timeout: 30,
    allowed_operations: ["read_file", "write_file", "http_request"],
    restricted_paths: ["/etc", "/usr", "/bin"],
    max_file_size: "10MB"
  }
}
```

This architecture allows Relay to maintain its simple, high-level abstractions while providing access to powerful system-level capabilities when needed, all through a consistent message-passing interface.

---

## 8. Technical Architecture

**Compiler Stack:**
- ✅ **Lexer**: Tokenize Relay source code
- ✅ **Parser**: Generate AST from tokens (simplified grammar)
- 🚧 **Type Checker**: Validate types and structures (partial)
- 🚧 **Code Generator**: Generate runtime bytecode/IR (future)

**Runtime Stack:**
- ✅ **Core Evaluator**: Unified expression evaluation engine (pkg/runtime/core.go)
- ✅ **Value System**: Complete type system with immutable semantics (pkg/runtime/value.go)
- ✅ **Environment Management**: Lexical scoping with closure support (pkg/runtime/environment.go)
- ✅ **Server Infrastructure**: In-memory actor system with message handling (pkg/runtime/servers.go)
- ✅ **Built-in Functions**: Standard library functions (pkg/runtime/builtins.go)
- ✅ **Method Dispatch**: Object and array method calls (pkg/runtime/methods.go)
- 📋 **HTTP Server**: JSON-RPC 2.0 endpoint (planned)
- 📋 **Template Engine**: Server-side rendering (planned)
- 📋 **Federation Client**: Service discovery and communication (planned)
- 📋 **Load Balancer**: Request distribution (planned)
- 📋 **Go Actor**: I/O and system operations handler (planned)

**Storage Layer (Planned):**
- 📋 **Embedded NoSQL database** (SQLite-like for key-value)
- 📋 **State serialization/deserialization**
- 📋 **Automatic persistence and recovery**
- 📋 **Transaction support for state mutations**

**Current Implementation:**
The Relay runtime is built around a unified evaluation architecture that consolidates all expression handling in a single, well-documented system. The codebase has been optimized for simplicity while maintaining 100% test compatibility (739 tests passing). 

**Currently Working Features:**
- First-class functions with closures and lexical scoping
- Complete struct system with definitions, instantiation, and field access
- In-memory server actors with thread-safe message passing
- Comprehensive expression evaluation with immutable-by-default semantics
- Array and object operations with method chaining
- Built-in functions and comprehensive error handling

**Architecture Status:**
The core language features are fully implemented and tested. The runtime uses an in-memory actor model for server communication. Persistence, HTTP endpoints, template rendering, and federation features are planned for future implementation.

This specification provides a clean, consistent grammar that is much easier to parse while maintaining all the power and expressiveness of the Relay language. The focus is on simplicity, federation, and making distributed web development as easy as early PHP while being far more powerful and robust.