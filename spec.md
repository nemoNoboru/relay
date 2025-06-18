# Relay Language - Complete Technical Specification

🎉 **MAJOR MILESTONES**: Full closure support, first-class functions, and struct system implemented!

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
- BASIC-simple for hobbyists, enterprise-ready for production
- ✨ **NEW**: Fully functional with closures and first-class functions for advanced programming patterns
- ✨ **NEW**: Complete struct system with definitions, instantiation, and field access

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

**Key Innovation:** Servers are stateless in syntax but stateful in behavior. State is automatically persisted to embedded NoSQL database and restored on startup.

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

### 7.1 JSON-RPC 2.0 HTTP Server

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

### 7.2 State Management Runtime

- **Persistence**: Automatic persistence using embedded NoSQL database (similar to SQLite)
- **Concurrency**: State is automatically synchronized across concurrent requests
- **Durability**: All state changes are immediately written to disk
- **Recovery**: State is restored from database on server restart
- **Isolation**: Each server instance maintains separate state namespace
- **Threading**: State mutations are atomic and thread-safe
- **Immutability**: All state updates create new instances via method calls

### 7.3 Template Rendering Engine

- **Server-side rendering**: Templates are rendered on the server
- **State access**: Templates have access to data returned from send calls
- **HTTP routes**: Template declarations automatically create HTTP routes
- **Content-type detection**: Automatic content-type based on file extension
- **Template syntax**: `@{expression}`, `@if(condition) {...}`, `@for(item in items) {...}`
- **Error handling**: Failed sends result in configurable error pages
- **URL Parameters**: Template function parameters are extracted from URL paths

### 7.4 Load Balancing System

**Strategies:**
- `round_robin` (default): Distribute requests evenly
- `random`: Random selection
- `lowest_latency`: Route to fastest responding server
- `sticky`: Session-based routing using hashed keys

**Health Checking:**
- Automatic health checks for all registered servers
- Failed servers temporarily removed from pool
- Automatic recovery detection

### 7.5 Federation Protocol

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

## 8. Technical Architecture

**Compiler Stack:**
- Lexer: Tokenize Relay source code
- Parser: Generate AST from tokens (simplified grammar)
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

This specification provides a clean, consistent grammar that is much easier to parse while maintaining all the power and expressiveness of the Relay language. The focus is on simplicity, federation, and making distributed web development as easy as early PHP while being far more powerful and robust.