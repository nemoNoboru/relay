Relay Language - Complete Technical Specification
Version: 0.3 "Cloudpunks Edition"
By: Cloudpunks
Mission: A federated, minimal language to build modern distributed web services with the simplicity of early PHP.

Vision & Architecture
Executive Summary
A new web platform designed to democratize software development and hosting, moving away from centralized cloud oligarchies (AWS, Azure, Google Cloud) toward community-owned infrastructure. The platform combines the simplicity of the original LAMP stack with modern federation capabilities, allowing small communities to host their own web applications while seamlessly connecting to a larger federated network.

Vision Statement
Goal: Create an alternative way of building and deploying software where applications are built by small teams, hosted by communities, and administered by their hosts - completely independent of Silicon Valley cloud platforms.

Core Principles:

Community-first: Software owned and operated by its users
Federation-native: Seamless collaboration across independent nodes
Developer-friendly: As easy to learn as BASIC was in the 1980s
Deployment-simple: Single binary distribution, no complex configuration
Escape-hatch ready: Can drop down to lower-level code when needed
Language Philosophy
Relay is designed for federated, self-hostable, and functional applications. Each script is:

A JSON-RPC 2.0 server by default
Stateless by design, but supports automatic persistence
Able to send, receive, and template-render data
Built for high interoperability and zero config deployment
Callable from any programming language (Python, Go, Rust, PHP, etc.)
Core Concept
Relay = JavaScript + built-in server + federation

Every .relay file is automatically:

A JSON-RPC 2.0 server
A web server with HTML templates
Part of a federated network
Callable from any programming language (Python, Go, Rust, PHP, etc.)
The 30-Minute Guide
Learn Relay in 30 minutes if you know JavaScript

1. Basic Syntax (5 minutes)
Variables and Functions
// Exactly like JavaScript
set name = "Alice"
set age = 25
set items = [1, 2, 3]

fn greet(name) {
  return `Hello, ${name}!`
}
Data Types
// Just like JS, but with validation
struct User {
  name: string,
  email: string,
  age: number
}

// Define service contracts
protocol BlogService {
  get_posts() -> [Post]
  add_post(title: string, content: string) -> Post
}

set user = User{
  name: "Alice",
  email: "alice@example.com", 
  age: 25
}
2. Server Basics (10 minutes)
Your First Server

Here's how you create a simple blog server in Relay:

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
        set post = {
            id: id,
            title: title,
            content: content
        }
        
        set posts = state.get("posts")
        set updated_posts = posts.push(post)
        state.set("posts", updated_posts)
        state.set("next_id", id + 1)
        
        post
    }
}
```

To run this server, execute the following command:

```bash
./relay -run examples/blog_server.rl -server
```

That's it! You now have a JSON-RPC 2.0 server running on port 8080.

3. Language Interoperability (5 minutes)
Call Relay from Any Language

You can call your Relay server from any language that can make HTTP requests. Here's how you can call the `create_post` method using `curl`:

```bash
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "blog_server.create_post",
    "params": ["My First Post", "This is the content of my first blog post!"],
    "id": 1
  }'
```

And here's how you would get the list of posts:
```bash
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "blog_server.get_posts",
    "id": 2
  }'
```

Here are some examples of how to call the server from other languages:

Python:

```python
import requests
import json

response = requests.post('http://localhost:8080/rpc', json={
    "jsonrpc": "2.0",
    "method": "blog_server.create_post",
    "params": ["My First Post", "This is the content of my first blog post!"],
    "id": 1
})

post = response.json()['result']
print(post)
```

Go:
```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type JSONRPCRequest struct {
    Jsonrpc string        `json:"jsonrpc"`
    Method  string        `json:"method"`
    Params  []interface{} `json:"params"`
    ID      int           `json:"id"`
}

func main() {
    reqBody := &JSONRPCRequest{
        Jsonrpc: "2.0",
        Method:  "blog_server.create_post",
        Params:  []interface{}{"My First Post", "This is the content of my first blog post!"},
        ID:      1,
    }

    jsonValue, _ := json.Marshal(reqBody)
    resp, err := http.Post("http://localhost:8080/rpc", "application/json", bytes.NewBuffer(jsonValue))
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    fmt.Println(result)
}
```

JavaScript/Node.js:
```javascript
const fetch = require('node-fetch');

async function createPost() {
    const response = await fetch('http://localhost:8080/rpc', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({
            jsonrpc: '2.0',
            method: 'blog_server.create_post',
            params: ['My First Post', 'This is the content of my first blog post!'],
            id: 1
        })
    });

    const result = await response.json();
    console.log(result);
}

createPost();
```

Call Other Languages from Relay
// Define protocols for external services
protocol MLService {
  analyze(data: object, model: string) -> object
}

protocol DataProcessor {
  process(input: object) -> object  
}

protocol LegacySystem {
  get_user_data(user_id: string) -> object
}

server {
    receive process_data {data} {
        // Call services by protocol - language doesn't matter!
        set analysis = send MLService analyze {
            data: data,
            model: "sentiment"
        }
        
        set processed = send DataProcessor process {
            input: analysis.result
        }
        
        set legacy_data = send LegacySystem get_user_data {
            user_id: data.user_id
        }
        
        return {
            analysis: analysis,
            processed: processed, 
            legacy: legacy_data
        }
    }
}
4. Federation (10 minutes)
Calling Other Services
protocol UserService {
  get_user(username: string) -> User
}

protocol BlogService {
  get_posts_by(author: string) -> [Post]
}

server {
  receive get_user_posts {username} {
    // Call services by protocol - automatic discovery!
    set user = send UserService get_user {username}
    set posts = send BlogService get_posts_by {author: username}
    
    return {user, posts}
  }
}
Automatic Discovery
server {
  receive get_all_posts {} {
    // Find all services implementing BlogService protocol
    set blog_services = discover(BlogService)
    set all_posts = []
    
    for (service in blog_services) {
      set posts = send service get_posts {}
      all_posts.push(...posts)
    }
    
    return all_posts
  }
}
5. Authentication (5 minutes)
Add Auth to Any Service
auth {
  users: "local"  // or "federated" for cross-service login
}

server {
  set posts = []
  
  receive get_posts {} {
    // Anyone can read
    return posts
  }
  
  receive add_post {title, content} {
    require_login()  // Must be logged in
    
    posts.push({
      title, 
      content, 
      author: current_user.name,
      date: new Date()
    })
    
    return "Post added"
  }
}

// Login page auto-generated at /login
// User registration auto-generated at /register
6. Deployment Options
Community Relay Servers
config {
  // Deploy to your community
  community: "my-neighborhood.relay",
  
  // Or deploy to global network
  global_relay: "relay.world",
  
  // Control who can discover your service
  visibility: "community"  // "community", "global", "private"
}

server implements BlogService {
  // Your service is automatically:
  // 1. Hosted by the community relay server
  // 2. Discoverable by protocol within the community
  // 3. Load-balanced if multiple instances exist
}
Deployment Models
Community Hosting:

# Join a community
relay join my-neighborhood.relay

# Deploy your app to community infrastructure
relay deploy blog.relay --community my-neighborhood.relay

# Your app runs on community servers, costs shared by members
# Available at: blog.my-neighborhood.relay
Global Hosting:

# Deploy to the global relay network
relay deploy blog.relay --global

# Your app is discoverable worldwide
# Available at: blog.alice.relay.world
Self-Hosting:

# Run on your own server
relay build blog.relay --output my-blog-server
./my-blog-server --port 8080

# Register with community for discovery
relay register --service-url https://my-server.com:8080
Federation Discovery
// Services discover each other based on deployment model

server {
  receive get_community_blogs {} {
    // Find blogs in same community
    set local_blogs = discover(BlogService, {scope: "community"})
    return local_blogs
  }
  
  receive get_all_blogs {} {
    // Find blogs globally across all communities
    set global_blogs = discover(BlogService, {scope: "global"})
    return global_blogs
  }
  
  receive get_trusted_blogs {} {
    // Find blogs in trusted communities only
    set trusted_blogs = discover(BlogService, {
      scope: "trusted",
      communities: ["friends.relay", "family.relay"]
    })
    return trusted_blogs
  }
}
Complete Example: Federated Blog
blog.relay:

struct Post {
  title: string,
  content: string,
  author: string,
  date: Date
}

protocol BlogService {
  get_posts() -> [Post]
  add_post(title: string, content: string) -> Post
  new_post(post: Post) -> void  // For federation
}

auth {
  users: "local"
}

server implements BlogService {
  set posts = []
  
  receive get_posts {} {
    return posts
  }
  
  receive add_post {title, content} {
    require_login()
    
    set post = Post{
      title,
      content, 
      author: current_user.name,
      date: new Date()
    }
    
    posts.push(post)
    
    // Notify other blog services
    broadcast(BlogService) new_post {post}
    
    return post
  }
  
  receive new_post {post} {
    // Receive posts from other blogs
    posts.push({...post, federated: true})
  }
}

page "/" from get_posts
page "/write" template "write.html"
index.html:

<h1>Federated Blog</h1>

@if(current_user) {
  <p>Welcome, @{current_user.name}! <a href="/write">Write Post</a></p>
} else {
  <p><a href="/login">Login</a> to write posts</p>
}

@for(post in posts) {
  <article>
    <h2>@{post.title}</h2>
    <p>@{post.content}</p>
    <small>
      by @{post.author} 
      @if(post.federated) {
        (federated)
      }
    </small>
  </article>
}
write.html:

<h1>Write Post</h1>
<form action="/add_post" method="post">
  <input name="title" placeholder="Title" required>
  <textarea name="content" placeholder="Content" required></textarea>
  <button>Publish</button>
</form>
What You Get For Free:
✅ JSON-RPC 2.0 server
✅ Web server with HTML templates
✅ User authentication
✅ Data persistence
✅ Federation discovery
✅ Cross-service communication
✅ Interoperability with all programming languages
✅ Community hosting or global deployment
Deploy Anywhere:
relay run blog.relay                    # Development server

# Production options:
relay build blog.relay                  # Single binary for self-hosting
relay deploy community.relay            # Deploy to your community
relay deploy relay.world                # Deploy to global relay network

# Federation options:
relay register BlogService              # Make discoverable globally  
relay register BlogService --community  # Make discoverable in community only
Language Reference
Reserved Keywords
struct, protocol, server, auth, config, receive, send, discover, broadcast,
page, template, let, fn, if, else, for, in, try, catch, return, throw,
require_login, current_user
Data Types
// Primitive types
string, number, bool, Date

// Collection types  
[type]         // Arrays
{key: type}    // Objects

// Validation (for struct fields)
string.min(n).max(n)
string.email()
string.url()
Cheat Sheet
Core Syntax
// Variables
set x = 42

// Functions  
fn hello(name) { return `Hi ${name}` }

// Data structures
struct User { name: string, age: number }

// Service contracts
protocol UserService {
  get_user(id: string) -> User
  create_user(user: User) -> User
}

// Server (auto-exposes JSON-RPC 2.0)
server implements UserService { 
  receive method_name {param1, param2} { 
    return result 
  }
}

// Auth
auth { users: "local" }
require_login()
current_user.name

// Federation
send ProtocolName method_name {params}        // By protocol
send "service.com" method_name {params}       // By address
discover(ProtocolName)                        // Find services
broadcast(ProtocolName) method_name {params}  // Notify all

// Templates
page "/path" from method_name
@{variable}
@if(condition) { ... }
@for(item in items) { ... }

// JSON-RPC from other languages
POST /
{
  "jsonrpc": "2.0",
  "method": "method_name",
  "params": {...},
  "id": 1
}
Total concepts to learn: ~20
Time to productivity: 30 minutes
Lines of code for full-featured app: ~50

This is as simple as PHP was in 1995, but federated by default.

Technical Implementation Details
Runtime Specification
JSON-RPC 2.0 HTTP Server
Every Relay program automatically exposes a JSON-RPC 2.0 HTTP endpoint:

Request:

{
  "jsonrpc": "2.0",
  "method": "method_name",
  "params": {...},
  "id": 1
}
Response:

{
  "jsonrpc": "2.0", 
  "result": {...},
  "id": 1
}
Error:

{
  "jsonrpc": "2.0",
  "error": {
    "code": -32000,
    "message": "Application error",
    "data": {...}
  },
  "id": 1
}
State Management Runtime
Persistence: Automatic persistence using embedded NoSQL database (similar to SQLite)
Concurrency: State is automatically synchronized across concurrent requests
Durability: All state changes are immediately written to disk
Recovery: State is restored from database on server restart
Threading: State mutations are atomic and thread-safe
Template Rendering Engine
Server-side rendering: Templates are rendered on the server
State access: Templates have access to data returned from send calls
HTTP routes: Page declarations automatically create HTTP routes
Content-type detection: Automatic content-type based on file extension
Template syntax: @{expression}, @if(condition) {...}, @for(item in items) {...}
Federation Protocol
Service Discovery: Automatic discovery of services implementing protocols
Load Balancing: Automatic distribution across multiple instances
Health Checking: Failed services temporarily removed from pool
Cross-Language: Any language can participate via JSON-RPC