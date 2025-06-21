# Manual Guide: Migrating from P2P to Federation (Phase 1, Task 1)

This guide provides a step-by-step manual process for the first task of the networking stack migration: **Refactoring `WebSocketP2P` to `FederationRouter`**.

## 1. Objective

The goal of this task is to rename and restructure the core P2P networking component without altering its existing functionality. You will:
1.  Rename the `WebSocketP2P` struct to `FederationRouter`.
2.  Introduce the concept of `NodeType` (`main` vs. `home`) into the struct.
3.  Ensure all existing code compiles and all tests run (and fail in the same way) after the refactor.

This is a **purely structural refactor**. It will **not** fix the "routing loop" bug but is the essential first step required to implement the new hub-and-spoke architecture that will.

## 2. Core Files to Modify

Based on the codebase, the primary files you will be working with are:
-   `pkg/runtime/websocket_p2p.go`: The definition of the `WebSocketP2P` struct and its associated methods.
-   `pkg/runtime/server.go`: Where the `WebSocketP2P` component is initialized and used.
-   Other files that may reference `WebSocketP2P` directly.

## 3. Step-by-Step Instructions

### Step 3.1: Rename the Core File

To reflect the new naming scheme, rename the file itself.
**Action:**
In your terminal, run:
```sh
mv pkg/runtime/websocket_p2p.go pkg/runtime/federation_router.go
```

### Step 3.2: Introduce `NodeType`

This is the first architectural change. You will add a new type to distinguish between a "main" relay (a public server) and a "home" relay (a client behind a NAT).

**Action:**
In the newly renamed file `pkg/runtime/federation_router.go`, add the following code near the top, before the struct definition:

```go
// pkg/runtime/federation_router.go

// ... (imports)

// NodeType defines the role of a node in the federation architecture.
type NodeType string

const (
    // NodeTypeMain indicates a publicly accessible server node.
	NodeTypeMain NodeType = "main"
    // NodeTypeHome indicates a client node, typically behind a NAT.
	NodeTypeHome NodeType = "home"
)

// ...
```

### Step 3.3: Refactor the `WebSocketP2P` Struct

Now, perform the core rename and add the new `nodeType` field.

**Action:**
In `pkg/runtime/federation_router.go`, change the `WebSocketP2P` struct as follows:

**Before:**
```go
// pkg/runtime/federation_router.go

type WebSocketP2P struct {
	nodeID           string
	peers            map[string]*websocket.Conn
	peerMutex        sync.RWMutex
	messageQueue     chan *P2PMessage
	handlers         map[string]P2PMessageHandler
	registry         *ExposableServerRegistry
	responseChannels map[string]chan *P2PMessage
	running          bool
	shutdownChan     chan bool
}
```

**After:**
```go
// pkg/runtime/federation_router.go

type FederationRouter struct {
    nodeType         NodeType
	nodeID           string
	peers            map[string]*websocket.Conn
	peerMutex        sync.RWMutex
	messageQueue     chan *P2PMessage
	handlers         map[string]P2PMessageHandler
	registry         *ExposableServerRegistry
	responseChannels map[string]chan *P2PMessage
	running          bool
	shutdownChan     chan bool
}
```
*Note: The only changes are renaming the struct to `FederationRouter` and adding the `nodeType` field.*

### Step 3.4: Update the Constructor and Dependent Functions

The function that creates the P2P instance needs to be updated. It's likely called `NewWebSocketP2P`. You will rename it and modify it to accept the new `nodeType`.

**Action:**
1.  Find the `NewWebSocketP2P` function in `pkg/runtime/federation_router.go`.
2.  Rename it to `NewFederationRouter`.
3.  Add `nodeType NodeType` as the first parameter.
4.  Assign the `nodeType` to the struct field.

**Before:**
```go
// pkg/runtime/federation_router.go

func NewWebSocketP2P(nodeID string, registry *ExposableServerRegistry) *WebSocketP2P {
	// ...
}
```

**After:**
```go
// pkg/runtime/federation_router.go

func NewFederationRouter(nodeType NodeType, nodeID string, registry *ExposableServerRegistry) *FederationRouter {
	router := &FederationRouter{
		nodeType:         nodeType,
		nodeID:           nodeID,
		peers:            make(map[string]*websocket.Conn),
		messageQueue:     make(chan *P2PMessage, 100),
		handlers:         make(map[string]P2PMessageHandler),
		registry:         registry,
		responseChannels: make(map[string]chan *P2PMessage),
		shutdownChan:     make(chan bool),
	}
    // ... rest of the function
	return router
}
```

### Step 3.5: Update All Other Method Receivers

You must now change the receiver for all methods in this file from `(p *WebSocketP2P)` to `(fr *FederationRouter)`.

**Action:**
Perform a find-and-replace within `pkg/runtime/federation_router.go`:
-   **Find:** `*WebSocketP2P`
-   **Replace:** `*FederationRouter`

This will update method definitions like `func (p *WebSocketP2P) Start() { ... }` to `func (fr *FederationRouter) Start() { ... }`.

### Step 3.6: Update Usage in `server.go`

The `Server` struct in `pkg/runtime/server.go` holds an instance of the P2P component. You must update it.

**Action:**
In `pkg/runtime/server.go`:
1.  Find the `Server` struct definition.
2.  Change the type of the P2P-related field.

**Before:**
```go
// pkg/runtime/server.go
type Server struct {
    // ... other fields
    p2p *WebSocketP2P
}
```

**After:**
```go
// pkg/runtime/server.go
type Server struct {
    // ... other fields
    p2p *FederationRouter
}
```

3.  Find where `NewWebSocketP2P` is called (likely in `NewServer` or a similar constructor).
4.  Update the call to `NewFederationRouter` and pass in a default `NodeType`. For now, we will hardcode `NodeTypeMain` to preserve existing behavior. The logic for auto-detection will be added in a later task.

**Before:**
```go
// pkg/runtime/server.go
s.p2p = NewWebSocketP2P(nodeID, s.Registry)
```

**After:**
```go
// pkg/runtime/server.go
s.p2p = NewFederationRouter(NodeTypeMain, nodeID, s.Registry)
```

## 4. Verification

After completing these steps, you must verify that you have only changed the structure, not the behavior.

1.  **Compile the Project:** Run a full build to catch any syntax errors or type mismatches you may have missed.
    ```sh
    go build ./...
    ```
    This command should complete without any errors.

2.  **Run Unit Tests:** Execute the project's unit tests. They should all pass, just as they did before.
    ```sh
    go test ./...
    ```
    The test for the routing loop will still fail, which is expected.

3.  **Run the E2E Test:** Execute the primary end-to-end test.
    ```sh
    ./examples/e2e_p2p_test.sh
    ```
    **Expected Outcome:** The script will run, and the remote calls will still fail with the "routing loop detected" error in the logs. This is the **correct** outcome for this task. It confirms that your refactoring did not introduce new, unrelated problems.

You have successfully completed this task when the code is using the new `FederationRouter` naming and structure, but the system's behavior—including the existing bug—remains identical. 