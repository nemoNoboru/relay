# Relay - Federated Web Programming Language

> **Version 0.3.0-dev "Cloudpunks Edition"**

A federated, minimal programming language designed to build modern distributed web services with the simplicity of early PHP, but with built-in federation capabilities.

## Overview

Relay is designed to democratize web development by moving away from centralized cloud oligarchies (AWS, Azure, Google Cloud) toward community-owned infrastructure. It combines the simplicity of the original LAMP stack with modern federation capabilities.

### Key Features

- **Federation-native**: Services automatically discover and communicate with each other
- **JSON-RPC 2.0 by default**: Every `.relay` file becomes a web server
- **Community hosting**: Deploy to community-owned infrastructure
- **Language interoperability**: Call from any programming language
- **Zero-config deployment**: Single binary distribution
- **Built-in state management**: Automatic persistence with embedded database
- **Template system**: Server-side rendering with simple syntax

## Quick Start

### Install

```bash
go install github.com/cloudpunks/relay/cmd/relay@latest
```

### Hello World

Create `hello.relay`:

```relay
protocol HelloService {
  greet(name: string) -> string
}

server implements HelloService {
  receive greet {name: string} -> string {
    return `Hello, ${name}!`
  }
}
```

Run it:

```bash
relay hello.relay
```

Your service is now running on `http://localhost:8080` with a JSON-RPC 2.0 API!

Test it:

```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"greet","params":{"name":"World"},"id":1}'
```

## Language Features

### Immutable Data Operations

Relay is immutable-by-default, making it safer and more predictable:

```relay
// Object updates - always create new instances
set user = User{name: "John", age: 25}
set updated_user = user.set("age", 26)                    // Single field
set full_user = user.merge({email: "john@test.com", active: true})  // Multiple fields

// Array operations - always return new arrays  
set users = [user1, user2]
set more_users = users.add(user3)                         // Add to end
set ordered_users = users.sort_by(u => u.name)           // Sort
set active_users = users.filter(u => u.active)          // Filter
```

### Structs and Protocols

```relay
struct User {
  name: string.min(1).max(50),
  email: string.email(),
  age: number
}

protocol UserService {
  create_user(user: User) -> User
  get_user(id: string) -> User
}
```

### Servers with State

```relay
server user_service implements UserService {
  state {
    users: [User] = [],
    next_id: number = 1
  }
  
  receive create_user {user: User} -> User {
    set new_user = user.set("id", state.next_id.toString())
                       .set("created_at", now())
                       .set("active", true)
    
    state.users = state.users.add(new_user)
    state.next_id = state.next_id + 1
    return new_user
  }
  
  receive update_user {id: string, updates: UserUpdates} -> User {
    set user = state.users.find(u => u.id == id)
    set updated_user = user.merge(updates).set("updated_at", now())
    
    state.users = state.users.replace(user, updated_user)
    return updated_user
  }
}
```

### Federation

```relay
server {
  receive get_all_blogs {} {
    // Automatically discover all services implementing BlogService
    set blog_services = discover(BlogService)
    set all_posts = []
    
    for service in blog_services {
      set posts = send service get_posts {}
      all_posts = all_posts.append(posts)
    }
    
    return all_posts
  }
}
```

### Web Templates

```relay
// Auto-create web routes
template "index.html" from get_posts {}
template "api/posts.json" from get_posts {}
```

## Project Structure

```
relay/
├── cmd/relay/           # Main CLI application
├── pkg/
│   ├── lexer/          # Tokenizer
│   ├── parser/         # Syntax parser
│   ├── ast/            # Abstract Syntax Tree
│   ├── typechecker/    # Type checking
│   ├── compiler/       # Code generation
│   ├── runtime/        # Runtime engine
│   ├── server/         # JSON-RPC server
│   ├── federation/     # Service discovery
│   ├── templates/      # Template engine
│   └── state/          # State management
├── internal/
│   ├── config/         # Configuration
│   └── utils/          # Utilities
├── examples/           # Example programs
├── spec.md            # Language specification
├── vision.md          # Project vision
└── quickstart.md      # Quick start guide
```

## Development

### Building

```bash
go build -o relay cmd/relay/main.go
```

### Testing

```bash
go test ./...
```

### Running Examples

```bash
./relay examples/simple_blog.relay
```

## Architecture

The Relay language implementation consists of several key components:

1. **Lexer**: Tokenizes Relay source code
2. **Parser**: Builds Abstract Syntax Tree (AST)
3. **Type Checker**: Validates types and structures
4. **Compiler**: Generates Go code or bytecode
5. **Runtime**: 
   - JSON-RPC 2.0 HTTP server
   - State management with embedded database
   - Federation client for service discovery
   - Template rendering engine

## Deployment Options

### Development
```bash
relay run app.relay --port 3000
```

### Production Build
```bash
relay build app.relay --output my-app
./my-app --port 8080
```

### Community Hosting
```bash
relay deploy app.relay --community my-neighborhood.relay
```

### Global Network
```bash
relay deploy app.relay --global
```

## Contributing

We welcome contributions! This is an ambitious project to reshape web development.

### Getting Started

1. Fork the repository
2. Create a feature branch
3. Implement your changes
4. Add tests
5. Submit a pull request

### Areas for Contribution

- [ ] Lexer improvements
- [ ] Parser implementation
- [ ] Type checker
- [ ] Code generation
- [ ] Runtime features
- [ ] Federation protocol
- [ ] Template engine
- [ ] State management
- [ ] Documentation
- [ ] Examples

## Vision

Our goal is to recreate the golden age of web development (LAMP stack era) with modern federation capabilities. We want to make building distributed web applications as simple as PHP was in 1995, while enabling communities to own their own infrastructure.

This is a **"cloudpunk"** project - fighting against the centralization of the internet and giving power back to communities and individual developers.

## License

MIT License - see LICENSE file for details.

## Community

- **Discord**: [Join our community](https://discord.gg/relay-lang)
- **GitHub Discussions**: [Participate in discussions](https://github.com/cloudpunks/relay/discussions)
- **Twitter**: [@RelayLang](https://twitter.com/RelayLang)

---

*Built with ❤️ by the Cloudpunks community* 