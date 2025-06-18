# Relay REPL Guide

## Overview

The Relay REPL (Read-Eval-Print Loop) is an interactive development tool that helps you experiment with the Relay language syntax, test parser behavior, and develop language features.

## Getting Started

### Starting the REPL

```bash
# Build the Relay CLI
go build -o relay cmd/relay/main.go

# Start the REPL
./relay -repl
```

### Basic Usage

```
relay> struct User { name: string, email: string }
=== Parsed AST ===
[0] Struct 'User' with 2 fields
==================

relay> set greeting = "Hello, Relay!"
=== Parsed AST ===
[0] Set variable 'greeting'
==================
```

## REPL Commands

The REPL supports several special commands (all start with `:`):

- `:help`, `:h` - Show available commands
- `:quit`, `:q`, `:exit` - Exit the REPL
- `:clear`, `:c` - Clear the screen
- `:examples`, `:ex` - Show example Relay code
- `:ast` - Show AST inspection help
- `:test` - Run quick syntax validation tests
- `:version`, `:v` - Show version and parser info

## Features

### 1. Multi-line Input Support

The REPL automatically detects incomplete expressions and prompts for continuation:

```
relay> struct User {
    |     name: string,
    |     email: string
    | }
=== Parsed AST ===
[0] Struct 'User' with 2 fields
==================
```

### 2. Complete AST Tree Visualization

Every valid input shows its complete parsed AST structure recursively, displaying the full tree hierarchy:

```
=== Complete AST Tree ===
Program
├─ [0] StructExpr: 'User'
│  ├─ Field: 'name'
│  │  └─ Type: string
│  ├─ Field: 'email'
│  │  └─ Type: string
│  └─ Field: 'age'
│     └─ Type: number
=========================
```

This shows:
- Complete nested structure of all AST nodes
- Field types, parameters, return types
- Expression hierarchies and operators
- Method calls with arguments
- Block structures and control flow

### 3. Error Reporting

Parse errors are clearly displayed with line and column information:

```
relay> struct User { name string }
Parse error: repl:1:19: unexpected token "string" (expected ":")
```

### 4. Syntax Testing

Use `:test` to run quick validation tests on core language constructs:

```
relay> :test
=== Running Quick Syntax Tests ===
✅ Simple struct: OK
✅ Protocol: OK
✅ Function: OK
✅ Variable: OK
✅ Lambda: OK

Passed: 5/5 tests
===============================
```

## Development Workflow

### 1. Testing New Syntax

Use the REPL to quickly test if new syntax parses correctly:

```
relay> fn greet(name: string) -> string { "Hello, " + name }
=== Parsed AST ===
[0] Function 'greet' with 1 parameters
==================
```

### 2. Debugging Parser Issues

When you encounter parsing errors, the REPL shows exactly where the issue occurs:

```
relay> protocol Service { get_data() -> [Data], }
Parse error: repl:1:40: unexpected token "}" (expected method signature)
```

### 3. Iterative Development

Perfect for iterative language development:
1. Write new grammar rules
2. Test in REPL
3. Refine based on results
4. Repeat

## Example Session

```
relay> :version
=== Relay Language Info ===
Version: 0.3.0-dev
Parser: Participle v2.1.4
Status: Development
===========================

relay> :examples
=== Example Relay Code ===

1. Simple struct:
   struct User {
       name: string,
       email: string
   }
...

relay> struct Post {
    |     id: string,
    |     title: string,
    |     content: string
    | }
=== Parsed AST ===
[0] Struct 'Post' with 3 fields
==================

relay> protocol BlogService {
    |     get_posts() -> [Post]
    | }
=== Parsed AST ===
[0] Protocol 'BlogService' with 1 methods
==================

relay> :quit
Goodbye!
```

## Tips for Language Development

1. **Start Simple**: Test basic constructs before complex ones
2. **Use Multi-line**: Don't try to fit complex structures on one line
3. **Check AST Output**: Verify the parser interprets your code as expected
4. **Test Edge Cases**: Try malformed syntax to see error handling
5. **Use `:test`**: Regularly run syntax tests to catch regressions

## Integration with Development

The REPL is designed to complement your language development workflow:

- **Parser Development**: Test grammar rules immediately
- **Error Message Tuning**: See exactly what users will encounter
- **Documentation**: Generate examples from working REPL sessions
- **Testing**: Quick validation of language features

## Known Limitations

1. **No Execution**: Currently only parses and shows AST (evaluation not implemented)
2. **Simple Error Recovery**: Parser errors stop at first issue
3. **No State Persistence**: Each input is parsed independently

## Future Enhancements

- Syntax highlighting
- Auto-completion
- Variable state tracking
- Simple expression evaluation
- History navigation
- Load/save sessions

---

The Relay REPL is a powerful tool for language development. Use it to experiment, test, and refine your language implementation! 