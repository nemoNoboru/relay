# Relay Language - Complete Technical Specification

**Version:** 0.3 "Cloudpunks Edition"  
**By:** Cloudpunks  
**Mission:** A federated, minimal language to build modern distributed web services with the simplicity of early PHP.

---

## 1. Language Philosophy

Relay is designed for federated, self-hostable, and functional applications. Each script is:
- A JSON-RPC server by default
- Stateless by design, but supports automatic persistence
- Able to send, receive, and template-render data
- Built for high interoperability and zero config deployment
- PHP-simple for hobbyists, enterprise-ready for production

---

## 2. Reserved Keywords

```
struct, protocol, server, state, receive, send, template, dispatch,
config, set, fn, if, else, for, in, try, catch, return, throw
```

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

### Validation Modifiers
```relay
string.min(n)         // Minimum length
string.max(n)         // Maximum length  
string.email()        // Email validation
string.url()          // URL validation
string.regex(pattern) // Regex validation
```

### Immutable Data Operations

Relay is immutable-by-default. All data operations return new instances rather than modifying existing ones.

#### Object Methods
```relay
// Single field updates
user.set("name", "John")           // Set single field
user.set("email", "john@test.com") // Set single field
user.update("age", 25)             // Alias for set

// Multiple field updates  
user.merge({name: "John", age: 25})      // Merge multiple fields
user.defaults({active: true, role: "user"}) // Set defaults for missing fields

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
users.filter(u => u.active)     // Remove by condition

// Transformation
users.map(u => u.set("active", true))  // Transform each item
users.sort_by(u => u.name)            // Sort by field
users.reverse()                       // Reverse order
users.replace(old_user, new_user)     // Replace specific item
```

---

## 4. Core Language Constructs

### 4.1 Struct Definitions

```relay
struct User {
  username: string.min(3).max(50),
  email: string.email(),
  bio: optional(string.max(500)),
  created_at: datetime
}

struct Post {
  id: string,
  title: string.max(200),
  content: string,
  author: string,
  tags: [string],
  created_at: datetime
}
```

### 4.2 Protocol Definitions

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

### 4.3 Server Definitions (Multi-Server Support)

Every Relay program can contain multiple servers. All defined servers are automatically started and mounted on startup.

```relay
server blog_service implements BlogService {
  state {
    posts: [Post] = [],
    next_id: number = 1
  }
  
  receive get_posts {} -> [Post] {
    return state.posts  // Automatically wrapped as {state, result}
  }
  
  receive create_post {title: string, content: string} -> Post {
    set post = Post{
      id: state.next_id.toString(),
      title: title,
      content: content,
      author: config.default_author,
      tags: [],
      created_at: now()
    }
    
    state.posts = state.posts.prepend(post)
    state.next_id = state.next_id + 1
    
    return post
  }
  
  receive get_post {id: string} -> Post {
    set post = state.posts.find(p => p.id == id)
    if post {
      return post
    } else {
      throw {error: "Post not found", id: id}
    }
  }
  
  receive delete_post {id: string} -> bool {
    set initial_length = state.posts.length
    state.posts = state.posts.filter(p => p.id != id)
    return state.posts.length < initial_length
  }
}

server user_service implements UserService {
  state {
    users: [User] = [],
    next_id: number = 1
  }
  
  receive create_user {user: User} -> User {
    set new_user = user.set("created_at", now())
    state.users = state.users.add(new_user)
    return new_user
  }
  
  receive get_user {username: string} -> User {
    set user = state.users.find(u => u.username == username)
    if user {
      return user
    } else {
      throw {error: "User not found", username: username}
    }
  }
}
```

### 4.4 State Management (Stateless with Persistence)

**Key Innovation:** Servers are stateless in syntax but stateful in behavior. State is automatically persisted to embedded NoSQL database and restored on startup.

```relay
server counter_service {
  state {
    count: number = 0,
    last_updated: datetime = now(),
    history: [object] = []
  }
  
  receive increment {amount: optional(number) = 1} -> number {
    set increment_amount = amount || 1
    state.count += increment_amount
    state.last_updated = now()
    state.history = state.history.add({
      action: "increment",
      amount: increment_amount,
      timestamp: state.last_updated
    })
    return state.count
  }
  
  receive get_stats {} -> object {
    return {
      count: state.count,
      last_updated: state.last_updated,
      total_operations: state.history.length
    }
  }
}
```

### 4.5 Functions and Lambdas

```relay
// Function definitions
fn calculate_total(items: [object]) -> number {
  set total = 0
  for item in items {
    total += item.price
  }
  return total
}

fn format_date(date: datetime, format: string) -> string {
  return date.format(format)
}

// Higher-order functions
fn apply_discount(calculator: fn(number) -> number, price: number) -> number {
  return calculator(price)
}

// Lambda expressions
set numbers = [1, 2, 3, 4, 5]
set doubled = numbers.map(x => x * 2)
set filtered = numbers.filter(x => x > 3)

// Multi-line lambdas
set complex_transform = numbers.map(x => {
  set doubled = x * 2
  set result = doubled + 10
  return result
})
```

### 4.6 Send/Receive Communication

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

### 4.7 Template System (PHP-style Web Hosting)

**Revolutionary Feature:** The `template` keyword performs automatic send + render in one step.

```relay
// Template declaration
template "homepage.html" from get_posts {}

// Equivalent to:
// set data = send get_posts {}
// return render("homepage.html", data)

// More complex example
template "post.html" from get_post {id: string}
template "user-profile.html" from get_user {username: string}

// JSON API endpoints
template "api/posts.json" from get_posts {}
template "api/user.json" from get_user {username: string}
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
                <h2>@{post.title}</h2>
                <p>@{post.content}</p>
                <small>By @{post.author} on @{format_date(post.created_at, "YYYY-MM-DD")}</small>
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

### 4.8 Dispatch Pattern Matching

```relay
receive handle_activity {activity: Activity} -> object {
  dispatch activity.type {
    "Create" -> {
      return send "blog_service" create_post {
        title: activity.object.title,
        content: activity.object.content
      }
    }
    
    "Delete" -> {
      return send "blog_service" delete_post {
        id: activity.object.id
      }
    }
    
    "Follow" -> {
      return send "user_service" add_follower {
        username: activity.object.username,
        follower: activity.actor
      }
    }
  }
}
```

### 4.9 Configuration

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
    "analytics": "https://analytics.company.com",
    "email": "https://email-service.com"
  }
}
```

---

## 5. Runtime Specification

### 5.1 JSON-RPC 2.0 HTTP Server

Every Relay program automatically exposes a JSON-RPC 2.0 HTTP endpoint:

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

### 5.2 State Management Runtime

- **Persistence**: Automatic persistence using embedded NoSQL database (similar to SQLite)
- **Concurrency**: State is automatically synchronized across concurrent requests
- **Durability**: All state changes are immediately written to disk
- **Recovery**: State is restored from database on server restart
- **Isolation**: Each server instance maintains separate state namespace
- **Threading**: State mutations are atomic and thread-safe
- **Tuple Handling**: All receive handlers internally return `{state, result}` tuples

### 5.3 Template Rendering Engine

- **Server-side rendering**: Templates are rendered on the server
- **State access**: Templates have access to data returned from send calls
- **HTTP routes**: Template declarations automatically create HTTP routes
- **Content-type detection**: Automatic content-type based on file extension
- **Template syntax**: `@{expression}`, `@if(condition) {...}`, `@for(item in items) {...}`
- **Error handling**: Failed sends result in configurable error pages

### 5.4 Load Balancing System

**Strategies:**
- `round_robin` (default): Distribute requests evenly
- `random`: Random selection
- `lowest_latency`: Route to fastest responding server
- `sticky`: Session-based routing using hashed keys

**Health Checking:**
- Automatic health checks for all registered servers
- Failed servers temporarily removed from pool
- Automatic recovery detection

### 5.5 PHP-style File Hosting

**File Structure:**
```
/
├── index.relay          # Main application
├── blog/
│   ├── index.relay     # Blog service
│   └── post.html       # Template file
├── static/
│   ├── style.css       # Static assets
│   └── script.js
└── templates/
    ├── layout.html     # Shared templates
    └── partials/
```

**Request Handling:**
- `.relay` files are executed as Relay programs
- `.html` files with template declarations are rendered
- Static files served directly
- Automatic routing based on file structure

### 5.6 Federation Protocol

**App Discovery:**
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

---

## 6. Complete Example Application

**main.relay:**
```relay
struct Post {
  id: string,
  title: string.max(200),
  content: string,
  author: string,
  created_at: datetime
}

protocol BlogService {
  get_posts() -> [Post]
  create_post(title: string, content: string) -> Post
}

server blog_service implements BlogService {
  state {
    posts: [Post] = [],
    next_id: number = 1
  }
  
  receive get_posts {} -> [Post] {
    return state.posts
  }
  
  receive create_post {title: string, content: string} -> Post {
    set post = Post{
      id: state.next_id.toString(),
      title: title,
      content: content,
      author: "Anonymous",
      created_at: now()
    }
    
    state.posts = state.posts.prepend(post)
    state.next_id = state.next_id + 1
    
    return post
  }
}

// Web routes
template "index.html" from get_posts {}
template "api/posts.json" from get_posts {}

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
                <h2>@{post.title}</h2>
                <p>@{post.content}</p>
                <small>@{format_date(post.created_at, "YYYY-MM-DD HH:mm")}</small>
            </article>
        }
    </div>
</body>
</html>
```

---

## 8. Technical Architecture

**Compiler Stack:**
- Lexer: Tokenize Relay source code
- Parser: Generate AST from tokens
- Type Checker: Validate types and structures
- Code Generator: Generate runtime bytecode/IR

**Runtime Stack:**
- HTTP Server: JSON-RPC 2.0 endpoint
- State Manager: Embedded database integration
- Template Engine: Server-side rendering
- Federation Client: Service discovery and communication
- Load Balancer: Request distribution

**Storage Layer:**
- Embedded NoSQL database (SQLite-like for key-value)
- State serialization/deserialization
- Automatic persistence and recovery
- Transaction support for state mutations

This specification provides everything needed to start building the Relay language from scratch. The focus is on simplicity, federation, and making distributed web development as easy as early PHP while being far more powerful and robust.