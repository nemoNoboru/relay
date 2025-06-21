# Relay REPL Guide

## Overview

The Relay REPL (Read-Eval-Print Loop) is a powerful interactive development tool for the Relay language. It operates in two modes: **execution** for running code and **AST inspection** for analyzing syntax. This dual functionality makes it an essential tool for everything from simple expression testing to complex parser development.

## Getting Started

### Starting the REPL

```bash
# Build the Relay CLI
go build -o relay cmd/relay/main.go

# Start the REPL in execution mode
./relay -repl
```

## REPL Commands

The REPL supports several special commands (all start with `:`):

- `:help`, `:h` - Show available commands
- `:quit`, `:q`, `:exit` - Exit the REPL
- `:clear`, `:c` - Clear the screen
- `:mode` - Toggle between execution mode and AST inspection mode
- `:load <filename>` - Load and evaluate a `.rl` file into the current session
- `:examples`, `:ex` - Show example Relay code
- `:test` - Run quick syntax validation tests
- `:version`, `:v` - Show version and parser info

## Features

### 1. Dual Mode: Execution and AST Inspection

The REPL's core feature is its ability to switch between two modes using the `:mode` command.

- **Execution Mode (Default)**: Input is parsed and immediately evaluated by the runtime. The result of the expression is printed. This is useful for testing logic, server interactions, and language semantics.
- **AST Mode**: Input is parsed, and a detailed Abstract Syntax Tree (AST) is visualized. This is invaluable for language development, debugging parser rules, and understanding how code is structured.

### 2. Multi-line Input Support

The REPL automatically detects incomplete expressions and prompts for continuation:

```
relay> struct User {
    |     name: string,
    |     email: string
    | }
User = struct{name: string, email: string}
```

### 3. Loading Files

You can load a `.rl` file directly into your REPL session, which is perfect for setting up a development or debugging environment.

```bash
relay> :load examples/blog_server.rl
File loaded successfully!
```

### 4. Detailed Error Reporting

Parse and runtime errors are clearly displayed with context.

```
relay> 1 + "hello"
Runtime error: unsupported operand types for binary operation: number + string
```

## Example Session

```
relay> :version
=== Relay Language Info ===
Version: 0.3.0-dev
Parser: Participle v2
Status: Development
===========================

relay> set x = 10
x = 10

relay> set y = 20
y = 20

relay> x + y
30

relay> :mode
Switched to AST mode

relay> x + y
=== Complete AST Tree ===
Program
├─ [0] BinaryExpr: +
│  ├─ Left: PrimaryExpr
│  │  └─ BaseExpr
│  │     └─ Identifier: 'x'
│  └─ Right: UnaryExpr
│     └─ PrimaryExpr
│        └─ BaseExpr
│           └─ Identifier: 'y'
=========================

relay> :mode
Switched to execution mode

relay> :load examples/simple_struct.rl
File loaded successfully!

relay> user1
{name: "John Doe", email: "john.doe@example.com", age: 30}

relay> :quit
Goodbye!
```

## Known Limitations

1. **Simple Error Recovery**: The parser stops at the first error it encounters.
2. **State Persistence**: REPL session state is lost on exit.

## Future Enhancements

- Syntax highlighting
- Auto-completion
- History navigation (up/down arrows)
- Load/save complete sessions
---
The Relay REPL is a powerful tool for both using and developing the language. Use it to experiment, test, and refine your code and the language itself! 