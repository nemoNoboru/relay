# Relay Project Roadmap & Feature Status

**Generated:** @{current_date}

This document provides a detailed breakdown of the Relay project's current feature status, based on an analysis of the source code, documentation (`vision.md`, `spec.md`), and the extensive test suite.

---

## 🚀 Vision & Core Principles

| Principle | Status | Evidence & Notes |
| --- | --- | --- |
| **Community-first** |  Partially Implemented | The architecture supports federation, but community governance/hosting models are not yet defined. |
| **Federation-native** | Not Implemented | The P2P networking stack is currently broken. Unit tests (`TestHTTPServer_RemoteCall_Validation`) fail with a routing loop error, and E2E tests (`e2e_p2p_test_simple.sh`) confirm that all remote calls fail. The `MessageRouter` and `TransportAdapter` exist, but do not function correctly for internode communication. |
| **Developer-friendly** | Partially Implemented | The language is simple, but tooling (like the REPL) and documentation are still maturing. |
| **Deployment-simple** | Implemented | The project builds to a single binary (`go build`), fulfilling the core vision of simple deployment. |
| **Escape-hatch ready** | Not Implemented | The planned `Go Actor` for system-level tasks is not yet implemented. |

---

## 💎 Core Language Features

| Feature | Status | Evidence & Notes |
| --- | --- | --- |
| **Expression Evaluation** | Fully Implemented | The `pkg/runtime/core_test.go` and `evaluator_test.go` suites contain hundreds of tests covering all expression types. |
| **First-class Functions** | Fully Implemented | Verified in `pkg/runtime/functions_test.go`. Closures and lexical scoping are working as specified. |
| **Structs & Objects** | Fully Implemented | `pkg/runtime/structs_test.go` confirms definitions, instantiation, and field access work correctly. |
| **Immutable Data Ops** | Fully Implemented | `pkg/runtime/arrays_test.go` and `operations_test.go` show that methods like `.add()` and `.set()` return new instances. |
| **Control Flow** | Partially Implemented | `if`/`else` and `dispatch` are tested in `control_flow_test.go`. `for` loops are parsed but not yet implemented in the runtime. |
| **Error Handling** | Partially Implemented | `try`/`catch` is not implemented. `throw` is supported for basic error objects. |
| **Symbol Literals**| Fully Implemented | The parser handles symbol literals correctly, as seen in various test files. |

---

##  एक्ट Actor Model & Concurrency

| Feature | Status | Evidence & Notes |
| --- | --- | --- |
| **Server Actors** | Fully Implemented | Servers run in their own goroutines, as confirmed by `pkg/runtime/value.go` (`Server.Start`). |
| **Message Passing** | Fully Implemented | The `message()` function and `MessageRouter` provide the core message-passing functionality. |
| **State Management** | Partially Implemented | In-memory state with thread-safe access via message passing is fully functional. **Persistence is not yet implemented.** |
| **`send` Keyword** | Implemented (Local Only) | The `send` keyword is parsed, but the runtime implementation currently uses the `message()` built-in. Local server-to-server calls work. |

---

## 🌐 Networking & Federation

| Feature | Status | Evidence & Notes |
| --- | --- | --- |
| **Unified Message Router**| Fully Implemented | `pkg/runtime/message_router.go` and its tests confirm a centralized, transport-agnostic routing system. |
| **HTTP JSON-RPC Server** | Fully Implemented | `pkg/runtime/http_server_unified.go` provides a working JSON-RPC 2.0 endpoint. |
| **WebSocket P2P** | Partially Implemented | The `TransportAdapter` can establish WebSocket connections, but the routing logic is non-functional, as confirmed by E2E test failures. No data is successfully transmitted between nodes. |
| **Federation Protocol** | Not Implemented | With the core P2P communication broken, the higher-level federation protocol is not functional. Service discovery and multi-hop routing are not implemented. |
| **Load Balancing** | Not Implemented | No load balancing strategies are present in the current codebase. |

---

## 📄 Template & Configuration System

| Feature | Status | Evidence & Notes |
| --- | --- | --- |
| **Template Parsing** | Partially Implemented | The `template` keyword and syntax are recognized by the parser (`pkg/parser/definitions_test.go`), but there is no rendering engine. |
| **Template Rendering**| Not Implemented | No server-side rendering logic exists. |
| **`config` block** | Not Implemented | The `config` block is parsed but its values are not used by the runtime. All configuration is currently handled by CLI flags. |

---

## 🛠️ Tooling

| Feature | Status | Evidence & Notes |
| --- | --- | --- |
| **REPL** | Fully Implemented | `pkg/repl/repl.go` shows a working REPL with both execution and AST modes. |
| **CLI** | Partially Implemented | The CLI supports running files (`-run`), starting the server (`-server`), and the REPL (`-repl`). The `-build` command is not implemented. |

---

## 🏁 Summary & Next Steps

The core of the Relay language and its actor-based runtime is solid and well-tested. The unified networking architecture is a major success, providing a flexible foundation for federation.

**Immediate priorities based on this analysis should be:**
1.  **Implement State Persistence:** This is the most critical missing piece for building any real application.
2.  **Build the Template Rendering Engine:** To fulfill the "PHP-style" web hosting vision.
3.  **Implement the `config` block:** To move beyond CLI flags for configuration.
4.  **Flesh out the Federation Protocol:** Implement service discovery and more robust routing.
5.  **Implement the `Go Actor`:** To provide an "escape hatch" for system-level tasks. 