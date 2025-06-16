# Examples

This directory contains example projects and language implementations that demonstrate various parsing and language design concepts.

## basic-interpreter/

A complete BASIC language interpreter implemented using the Participle parser generator. This serves as:

- An example of how to build a complete language interpreter
- A reference implementation for parser design patterns
- A separate Go module demonstrating project organization

### Running the BASIC Interpreter

```bash
cd examples/basic-interpreter
go run . example.bas
```

### BASIC Language Features

- Variables and assignments
- Arithmetic operations
- Control flow (IF/THEN, GOTO)
- INPUT/PRINT statements
- REM comments
- Line numbers

This example is independent of the main Relay language implementation and has its own `go.mod` file. 