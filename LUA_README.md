# Relay Lua Runtime

This branch implements Lua support for Relay using Wasmoon, a Lua WASM runtime.

## Overview

The Relay Lua Runtime provides a sandboxed Lua environment where users can write Lua code to create web applications. This approach offers several benefits:

- **Familiar Syntax**: Lua is easy to learn and widely used in game development
- **Sandboxed Execution**: WASM provides security isolation
- **Performance**: Lua is fast and lightweight
- **Rich Ecosystem**: Leverages existing Lua knowledge and patterns

## Architecture

### Core Components

1. **RelayLuaRuntime** (`src/core/lua-runtime.ts`)
   - Manages the Wasmoon Lua engine
   - Provides Relay API to Lua code
   - Handles component collection and state management

2. **Relay API Surface**
   - `relay.show()` - Render UI components
   - `relay.state()` - Manage application state
   - `relay.fetch()` - Make HTTP requests
   - `relay.form()` - Create forms
   - `relay.when()` - Handle events
   - `relay.storage` - File operations
   - `relay.json` - JSON utilities
   - `relay.math` - Math operations
   - `relay.string` - String utilities

## Usage Examples

### Basic Component Rendering

```lua
-- Show a simple heading
relay.show("heading", {text = "Hello World", level = 1})

-- Show a card with content
relay.show("card", {
    title = "My Card",
    content = "This is card content"
})
```

### State Management

```lua
-- Initialize state
local books = relay.state("books", {
    {title = "The Great Gatsby", rating = 4},
    {title = "To Kill a Mockingbird", rating = 5}
})

-- Access state
local current_books = relay.state("books")
```

### Forms and Events

```lua
-- Create a form
relay.form("add-book")

-- Add form inputs
relay.show("input", {name = "title", placeholder = "Book title"})
relay.show("button", {text = "Add Book"})

-- Handle events
relay.when("click", function(data)
    -- Handle button click
    return "clicked"
end)
```

### Data Processing

```lua
-- Process data with loops
local users = {
    {name = "Alice", age = 25},
    {name = "Bob", age = 30}
}

local young_users = {}
for i, user in ipairs(users) do
    if user.age < 30 then
        table.insert(young_users, user.name)
    end
end

-- Use JSON utilities
local json_data = relay.json.stringify(users)
local parsed_data = relay.json.parse(json_data)
```

### Math and String Operations

```lua
-- Math operations
local sum = relay.math.add(10, 5)
local product = relay.math.mul(3, 4)

-- String operations
local combined = relay.string.concat("Hello", " ", "World")
local upper = relay.string.upper("hello")
```

## API Reference

### UI Components

#### `relay.show(component, props)`
Renders a UI component with the given properties.

```lua
relay.show("card", {title = "My Card", content = "Content"})
relay.show("button", {text = "Click Me"})
relay.show("input", {name = "username", placeholder = "Enter username"})
```

#### `relay.form(name)`
Creates a form component.

```lua
relay.form("login-form")
```

### State Management

#### `relay.state(name, initialValue?)`
Manages application state. If called with an initial value, it sets the state. Otherwise, it returns the current value.

```lua
-- Initialize state
relay.state("counter", 0)

-- Get state value
local count = relay.state("counter")
```

### HTTP Requests

#### `relay.fetch(url, options?)`
Makes HTTP requests.

```lua
local data = relay.fetch("https://api.example.com/users")
```

### Events

#### `relay.when(eventName, handler)`
Registers event handlers.

```lua
relay.when("click", function(data)
    return "handled"
end)
```

### Storage

#### `relay.storage.read(path)`
Reads data from storage.

#### `relay.storage.write(path, data)`
Writes data to storage.

### Utilities

#### `relay.json.parse(text)`
Parses JSON string to Lua table.

#### `relay.json.stringify(obj)`
Converts Lua table to JSON string.

#### `relay.math.add(a, b)`
Adds two numbers.

#### `relay.math.sub(a, b)`
Subtracts two numbers.

#### `relay.math.mul(a, b)`
Multiplies two numbers.

#### `relay.math.div(a, b)`
Divides two numbers.

#### `relay.math.mod(a, b)`
Returns remainder of division.

#### `relay.string.concat(...)`
Concatenates strings.

#### `relay.string.length(str)`
Returns string length.

#### `relay.string.upper(str)`
Converts string to uppercase.

#### `relay.string.lower(str)`
Converts string to lowercase.

## Testing

The Lua runtime includes comprehensive tests in `src/core/__tests__/lua-runtime.test.ts`. Note that these tests require Node.js to be run with the `--experimental-vm-modules` flag due to WASM module loading requirements.

## Browser Integration

The Lua runtime is designed to work in browser environments. Example usage:

```typescript
import { runLuaExample } from './src/core/lua-example';

// Run a book club app
const result = await runLuaExample('book-club');
console.log(result);
```

## Security Considerations

- **Sandboxed Execution**: All Lua code runs in a WASM sandbox
- **Controlled API**: Only Relay-approved functions are available
- **No File System Access**: Direct file system access is blocked
- **No Network Access**: Network requests go through Relay's controlled `fetch` API

## Performance

- **Fast Startup**: Lua engine initializes quickly
- **Efficient Execution**: Lua is designed for performance
- **Small Bundle**: Wasmoon adds minimal overhead
- **Memory Efficient**: Lua has low memory footprint

## Future Enhancements

1. **Python Support**: Add Pyodide for Python runtime
2. **Language Detection**: Auto-detect language from code blocks
3. **Library Marketplace**: Community-curated Lua libraries
4. **Advanced UI Components**: More sophisticated component library
5. **Real-time Features**: WebSocket and real-time capabilities

## Migration from Original Relay Language

The original Relay language parser and interpreter have been removed in favor of the Lua runtime approach. This provides:

- **Lower Learning Barrier**: Lua is more familiar than custom syntax
- **Better Ecosystem**: Leverages existing Lua community and resources
- **Improved Security**: WASM sandbox is more robust
- **Faster Development**: No need to learn custom language

## Contributing

To contribute to the Lua runtime:

1. Install dependencies: `npm install`
2. Run tests: `npm test` (requires `--experimental-vm-modules` flag)
3. Add new API functions to `setupRelayAPI()` in `lua-runtime.ts`
4. Update tests and documentation

## License

This implementation follows the same license as the main Relay project. 