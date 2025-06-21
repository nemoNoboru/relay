# Servers and Actors in Relay

## Table of Contents
1. [Introduction to the Actor Model](#introduction-to-the-actor-model)
2. [Server Basics](#server-basics)
3. [Server Definition Syntax](#server-definition-syntax)
4. [State Management](#state-management)
5. [Receive Functions](#receive-functions)
6. [Message Passing](#message-passing)
7. [Concurrency and Thread Safety](#concurrency-and-thread-safety)
8. [Advanced Examples](#advanced-examples)
9. [Best Practices](#best-practices)
10. [Implementation Details](#implementation-details)
11. [Error Handling](#error-handling)
12. [Performance Considerations](#performance-considerations)

## Introduction to the Actor Model

The **Actor Model** is a mathematical model of concurrent computation that treats "actors" as the universal primitives of concurrent computation. In Relay, servers implement this actor model, providing a robust foundation for building concurrent, distributed applications.

### Key Principles

1. **Encapsulation**: Each actor (server) has its own private state
2. **Message Passing**: Actors communicate only through asynchronous messages
3. **No Shared State**: Actors don't share mutable state, eliminating race conditions
4. **Fault Isolation**: If one actor fails, it doesn't affect others
5. **Location Transparency**: Actors can be local or remote without changing the interface

### Benefits

- **Concurrency**: Natural support for parallel processing
- **Scalability**: Easy to distribute across multiple cores or machines
- **Fault Tolerance**: Isolated failure domains
- **Reasoning**: Easier to reason about concurrent programs
- **Composability**: Actors can be composed to build complex systems

## Server Basics

In Relay, a **server** is an implementation of an actor. Each server:

- Runs in its own goroutine
- Has private state accessible only to itself
- Processes messages sequentially (eliminating race conditions)
- Can respond to messages with return values
- Can send messages to other servers

### Server Lifecycle

1. **Definition**: Server structure is defined in code
2. **Instantiation**: Server is created and registered
3. **Startup**: Server goroutine starts and begins processing messages
4. **Message Processing**: Server processes incoming messages sequentially
5. **Shutdown**: Server can be stopped and cleaned up

## Server Definition Syntax

### Basic Structure

```relay
server server_name {
    state {
        field_name: type = default_value,
        // ... more fields
    }
    
    receive fn method_name(param1: type, param2: type) -> return_type {
        // Method implementation
        // Can access and modify state
        // Return value sent back to caller
    }
    
    // ... more receive functions
}
```

### Example: Counter Server

```relay
server counter {
    state {
        count: number = 0,
        name: string = "Counter"
    }
    
    receive fn increment() -> number {
        state.set("count", state.get("count") + 1)
        state.get("count")
    }
    
    receive fn decrement() -> number {
        state.set("count", state.get("count") - 1)
        state.get("count")
    }
    
    receive fn get_count() -> number {
        state.get("count")
    }
    
    receive fn reset() -> number {
        state.set("count", 0)
        0
    }
}
```

## State Management

### State Declaration

Server state is declared in the `state` block and consists of named fields with types and optional default values:

```relay
state {
    // Simple types
    counter: number = 0,
    name: string = "MyServer",
    active: bool = true,
    
    // Complex types
    users: [User] = [],
    config: Config = Config{host: "localhost", port: 8080},
    
    // Optional fields (default to nil)
    last_update: datetime
}
```

### State Access

State is accessed through the special `state` object using `get()` and `set()` methods:

```relay
receive fn update_counter(increment: number) -> number {
    // Get current value
    set current = state.get("counter")
    
    // Calculate new value
    set new_value = current + increment
    
    // Update state
    state.set("counter", new_value)
    
    // Return new value
    new_value
}
```

### State Immutability

- State can only be accessed from within the server's receive functions
- Each server has exclusive access to its own state
- State modifications are atomic within a single message handler
- No external code can directly access or modify server state

### State Persistence

```relay
server user_manager {
    state {
        users: [User] = [],
        total_count: number = 0,
        last_save: datetime = nil
    }
    
    receive fn add_user(user: User) -> bool {
        // Get current users
        set current_users = state.get("users")
        
        // Add new user
        set updated_users = current_users.add(user)
        
        // Update state atomically
        state.set("users", updated_users)
        state.set("total_count", state.get("total_count") + 1)
        state.set("last_save", now())
        
        true
    }
}
```

## Receive Functions

Receive functions are the only way to interact with a server. They define the server's public interface and handle incoming messages.

### Function Signature

```relay
receive fn function_name(param1: type, param2: type) -> return_type {
    // Function body
    // Can access state
    // Can call other functions
    // Must return a value of return_type
}
```

### Parameters

- Functions can have zero or more parameters
- Parameters are passed by value
- Complex types (structs, arrays) are copied

### Return Values

- Every receive function must declare a return type
- Functions must return a value of the declared type
- Return values are sent back to the message sender
- Use `nil` return type for functions that don't return meaningful data

### Example: Banking Server

```relay
server bank_account {
    state {
        balance: number = 0.0,
        account_number: string = "",
        is_active: bool = true
    }
    
    receive fn get_balance() -> number {
        state.get("balance")
    }
    
    receive fn deposit(amount: number) -> bool {
        if amount <= 0 {
            return false
        }
        
        set current_balance = state.get("balance")
        state.set("balance", current_balance + amount)
        true
    }
    
    receive fn withdraw(amount: number) -> bool {
        if amount <= 0 {
            return false
        }
        
        set current_balance = state.get("balance")
        if current_balance < amount {
            return false
        }
        
        state.set("balance", current_balance - amount)
        true
    }
    
    receive fn transfer(to_account: string, amount: number) -> bool {
        // Withdraw from this account by sending a message to self
        set withdraw_success = message("bank_account", "withdraw", amount)
        if !withdraw_success {
            return false
        }
        
        // TODO: Implement sending to another account
        // set transfer_success = message(to_account, "deposit", amount)
        // if !transfer_success {
        //     // Rollback logic needed here
        //     deposit(amount) 
        //     return false
        // }
        
        true
    }
}
```

## Message Passing

### The `message()` Function

The `message()` function is used to send messages to servers:

```relay
message(server_name: string, method_name: string, ...args)
```

- **server_name**: Name of the target server
- **method_name**: Name of the receive function to call
- **args**: Arguments to pass to the function

### Synchronous Communication

By default, message passing in Relay is synchronous:

```relay
// This blocks until the server responds
set result = message("counter", "increment")
print("New count: " + result)
```

### Message Examples

```relay
// Simple message with no arguments
set count = message("counter", "get_count")

// Message with arguments
set success = message("bank_account", "deposit", 100.50)

// Complex message with multiple arguments
set user = User{name: "Alice", email: "alice@example.com"}
set result = message("user_manager", "add_user", user, "admin")
```

### Error Handling

```relay
set {err, result} = message("server", "method", arg1, arg2)
if err != nil {
    print("Message failed: " + err)
} else {
    print("Success: " + result)
}
```

## Concurrency and Thread Safety

### Goroutine per Server

Each server runs in its own goroutine:

```go
// Implementation detail (Go code)
func (s *Server) Start(evaluator *Evaluator) {
    s.Running = true
    go s.messageLoop(evaluator)
}
```

### Message Queue

Servers process messages from a buffered channel:

- Messages are queued and processed sequentially
- No race conditions within a single server
- Multiple servers can run concurrently

### Thread Safety Guarantees

1. **No Data Races**: Each server's state is accessed by only one goroutine
2. **Sequential Processing**: Messages are processed one at a time per server
3. **Atomic State Updates**: State changes within a message handler are atomic
4. **Isolation**: Servers cannot interfere with each other's state

### Example: Producer-Consumer Pattern

```relay
server producer {
    state {
        items_produced: number = 0
    }
    
    receive fn produce_item() -> string {
        state.set("items_produced", state.get("items_produced") + 1)
        set item_id = "item_" + state.get("items_produced")
        
        // Send to consumer
        set {err, _} = message("consumer", "consume", item_id)
        if err != nil {
            print("Failed to send to consumer: " + err)
        }
        
        item_id
    }
}

server consumer {
    state {
        items_consumed: number = 0,
        items: [string] = []
    }
    
    receive fn consume(item: string) -> bool {
        set current_items = state.get("items")
        set updated_items = current_items.add(item)
        
        state.set("items", updated_items)
        state.set("items_consumed", state.get("items_consumed") + 1)
        
        true
    }
    
    receive fn get_stats() -> object {
        {
            consumed: state.get("items_consumed"),
            items: state.get("items")
        }
    }
}
```

## Advanced Examples

### Chat Server

```relay
server chat_room {
    state {
        room_name: string = "General",
        users: [User] = [],
        messages: [Message] = [],
        max_messages: number = 100
    }
    
    receive fn join(user: User) -> bool {
        set current_users = state.get("users")
        
        // Check if user already in room
        for existing_user in current_users {
            if existing_user.get("id") == user.get("id") {
                return false
            }
        }
        
        // Add user
        set updated_users = current_users.add(user)
        state.set("users", updated_users)
        
        // Broadcast join message
        set join_msg = Message{
            user: user,
            content: user.get("name") + " joined the room",
            timestamp: now(),
            type: "system"
        }
        
        add_message(join_msg)
        true
    }
    
    receive fn leave(user_id: string) -> bool {
        set current_users = state.get("users")
        set updated_users = current_users.filter(fn (u) { 
            u.get("id") != user_id 
        })
        
        state.set("users", updated_users)
        true
    }
    
    receive fn send_message(user_id: string, content: string) -> bool {
        set users = state.get("users")
        set user = users.find(fn (u) { u.get("id") == user_id })
        
        if user == nil {
            return false
        }
        
        set message = Message{
            user: user,
            content: content,
            timestamp: now(),
            type: "user"
        }
        
        add_message(message)
        true
    }
    
    receive fn add_message(message: Message) -> nil {
        set current_messages = state.get("messages")
        set updated_messages = current_messages.add(message)
        
        // Keep only last max_messages
        set max = state.get("max_messages")
        if updated_messages.length() > max {
            set start_index = updated_messages.length() - max
            updated_messages = updated_messages.slice(start_index)
        }
        
        state.set("messages", updated_messages)
        nil
    }
    
    receive fn get_messages() -> [Message] {
        state.get("messages")
    }
    
    receive fn get_users() -> [User] {
        state.get("users")
    }
}
```

### Load Balancer Server

```relay
server load_balancer {
    state {
        servers: [string] = [],
        current_index: number = 0,
        health_status: object = {},
        request_count: number = 0
    }
    
    receive fn register_server(server_name: string) -> bool {
        set current_servers = state.get("servers")
        set updated_servers = current_servers.add(server_name)
        state.set("servers", updated_servers)
        
        // Initialize health status
        set health = state.get("health_status")
        health.set(server_name, true)
        state.set("health_status", health)
        
        true
    }
    
    receive fn route_request(request: Request) -> Response {
        set servers = state.get("servers")
        if servers.length() == 0 {
            return ErrorResponse{message: "No servers available"}
        }
        
        // Round-robin selection
        set index = state.get("current_index")
        set server_name = servers.get(index)
        
        // Update index for next request
        set next_index = (index + 1) % servers.length()
        state.set("current_index", next_index)
        state.set("request_count", state.get("request_count") + 1)
        
        // Forward request
        set {err, response} = message(server_name, "handle_request", request)
        if err != nil {
            mark_server_unhealthy(server_name)
            // Retry with next server
            return route_request(request)
        }
        
        mark_server_healthy(server_name)
        response
    }
    
    receive fn mark_server_healthy(server_name: string) -> nil {
        set health = state.get("health_status")
        health.set(server_name, true)
        state.set("health_status", health)
        nil
    }
    
    receive fn mark_server_unhealthy(server_name: string) -> nil {
        set health = state.get("health_status")
        health.set(server_name, false)
        state.set("health_status", health)
        nil
    }
    
    receive fn get_stats() -> object {
        {
            total_requests: state.get("request_count"),
            active_servers: state.get("servers").length(),
            health_status: state.get("health_status")
        }
    }
}
```

## Best Practices

### 1. Single Responsibility

Each server should have a single, well-defined responsibility:

```relay
// Good: Focused responsibility
server user_authenticator {
    // Only handles authentication
}

// Good: Focused responsibility  
server session_manager {
    // Only handles sessions
}

// Bad: Too many responsibilities
server user_session_auth_manager {
    // Handles authentication, sessions, and user management
}
```

### 2. Immutable Messages

Use immutable data structures for message parameters:

```relay
// Good: Immutable struct
set user = User{id: "123", name: "Alice"}
message("user_manager", "add_user", user)

// Avoid: Mutable objects that might change
set mutable_data = {}
mutable_data.set("name", "Alice")
message("server", "process", mutable_data) // Risky
```

### 3. Error Handling

Always handle potential errors in message passing:

```relay
receive fn transfer_funds(to_account: string, amount: number) -> TransferResult {
    set {err, result} = message("bank", "transfer", to_account, amount)
    if err != nil {
        return TransferResult{success: false, error: err}
    }
    TransferResult{success: true, result: result}
}
```

### 4. State Validation

Validate state before and after modifications:

```relay
receive fn update_balance(new_balance: number) -> bool {
    // Validate input
    if new_balance < 0 {
        return false
    }
    
    // Update state
    state.set("balance", new_balance)
    
    // Validate state consistency
    if state.get("balance") != new_balance {
        // Log error, rollback, etc.
        return false
    }
    
    true
}
```

### 5. Timeouts

Consider implementing timeouts for long-running operations:

```relay
receive fn process_with_timeout(data: object) -> ProcessResult {
    set start_time = now()
    set timeout = 30 // seconds
    
    // Process data...
    
    if (now() - start_time) > timeout {
        return ProcessResult{success: false, error: "Timeout"}
    }
    
    ProcessResult{success: true, data: processed_data}
}
```

### 6. Graceful Degradation

Design servers to handle partial failures:

```relay
receive fn get_user_profile(user_id: string) -> UserProfile {
    set basic_info = get_basic_user_info(user_id)
    
    // Try to get additional info, but don't fail if unavailable
    set preferences = nil
    set {err, prefs} = message("preference_service", "get_preferences", user_id)
    if err != nil {
        // Log warning but continue
        log("Warning: Could not load preferences for " + user_id + ": " + err)
    } else {
        preferences = prefs
    }
    
    UserProfile{
        basic_info: basic_info,
        preferences: preferences
    }
}
```

## Implementation Details

### Message Structure

```go
type Message struct {
    Method string
    Args   []*Value
    Reply  chan *Value
}
```

### Server Structure

```go
type Server struct {
    Name        string
    State       map[string]*Value
    Receivers   map[string]*Function
    MessageChan chan *Message
    StateMutex  sync.RWMutex
    Running     bool
    Environment *Environment
}
```

### Message Processing Loop

```go
func (s *Server) messageLoop(evaluator *Evaluator) {
    for s.Running {
        select {
        case msg := <-s.MessageChan:
            s.handleMessage(msg, evaluator)
        }
    }
}
```

### State Synchronization

- Server state is protected by a mutex
- State object in message handlers is synchronized with server state
- Atomic updates ensure consistency

## Error Handling

### Common Error Scenarios

1. **Server Not Found**
   ```relay
   // Will throw error if "nonexistent_server" doesn't exist
   message("nonexistent_server", "method")
   ```

2. **Method Not Found**
   ```relay
   // Will throw error if server doesn't have "nonexistent_method"
   message("server", "nonexistent_method")
   ```

3. **Invalid Arguments**
   ```relay
   // Will throw error if method expects different argument types
   message("server", "method", wrong_type_arg)
   ```

4. **Timeout**
   ```relay
   // Will throw error if server doesn't respond within timeout
   message("slow_server", "slow_method")
   ```

### Error Recovery Patterns

```relay
server resilient_service {
    state {
        retry_count: number = 0,
        max_retries: number = 3
    }
    
    receive fn process_with_retry(data: object) -> ProcessResult {
        set attempts = 0
        
        while attempts < state.get("max_retries") {
            set {err, result} = risky_operation(data)
            if err == nil {
                state.set("retry_count", 0) // Reset on success
                return ProcessResult{success: true, result: result}
            }
            
            attempts = attempts + 1
            state.set("retry_count", attempts)
            
            if attempts >= state.get("max_retries") {
                return ProcessResult{
                    success: false, 
                    error: "Max retries exceeded: " + err
                }
            }
            
            // Wait before retry (exponential backoff)
            sleep(2^attempts * 1000) // milliseconds
        }
        
        ProcessResult{success: false, error: "Unexpected error"}
    }
}
```

## Performance Considerations

### 1. Message Overhead

- Each message creates goroutine communication overhead
- Batch operations when possible:

```relay
// Better: Batch operation
receive fn add_multiple_users(users: [User]) -> number {
    set added = 0
    for user in users {
        if add_single_user(user) {
            added = added + 1
        }
    }
    added
}

// Worse: Multiple messages
// for user in users {
//     message("user_manager", "add_user", user)
// }
```

### 2. State Size

- Keep state minimal and relevant
- Consider state cleanup for long-running servers:

```relay
receive fn cleanup_old_data() -> nil {
    set cutoff_time = now() - (24 * 60 * 60) // 24 hours ago
    set data = state.get("time_series_data")
    
    set cleaned_data = data.filter(fn (entry) {
        entry.get("timestamp") > cutoff_time
    })
    
    state.set("time_series_data", cleaned_data)
    nil
}
```

### 3. Memory Management

- Large data structures should be processed in chunks
- Consider using pagination for large datasets:

```relay
receive fn get_users_page(page: number, page_size: number) -> UserPage {
    set all_users = state.get("users")
    set start_index = page * page_size
    set end_index = start_index + page_size
    
    set page_users = all_users.slice(start_index, end_index)
    
    UserPage{
        users: page_users,
        page: page,
        total_pages: (all_users.length() + page_size - 1) / page_size,
        total_users: all_users.length()
    }
}
```

### 4. Hot Paths

Optimize frequently called operations:

```relay
server metrics_collector {
    state {
        counter_cache: object = {},
        last_cache_update: datetime = nil
    }
    
    receive fn increment_counter(name: string) -> nil {
        // Fast path: update cache
        set cache = state.get("counter_cache")
        set current = cache.get(name) ?? 0
        cache.set(name, current + 1)
        
        // Update timestamp
        state.set("last_cache_update", now())
        nil
    }
    
    receive fn get_counter(name: string) -> number {
        set cache = state.get("counter_cache")
        cache.get(name) ?? 0
    }
}
```

---

## Conclusion

The server and actor system in Relay provides a powerful foundation for building concurrent, scalable applications. By following the actor model principles and best practices outlined in this document, you can create robust systems that are:

- **Concurrent**: Natural parallelism without locks
- **Fault-tolerant**: Isolated failure domains
- **Scalable**: Easy to distribute and scale
- **Maintainable**: Clear separation of concerns
- **Reliable**: Strong consistency guarantees

The combination of immutable messages, isolated state, and sequential processing creates a programming model that's both powerful and safe, making it easier to build correct concurrent programs. 