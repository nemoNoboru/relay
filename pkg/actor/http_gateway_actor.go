package actor

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"relay/pkg/runtime"
	"time"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      interface{} `json:"id,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	ID      interface{} `json:"id,omitempty"`
}

// HTTPGatewayActor is an actor that acts as an HTTP gateway to the actor system.
// It is an http.Handler that creates a temporary RelayServerActor for each request.
type HTTPGatewayActor struct {
	actor        *Actor
	supervisorID string
	server       *http.Server
}

// NewHTTPGatewayActor creates a new HTTPGatewayActor.
func NewHTTPGatewayActor(name, supervisorID string, router *Router, port int) *HTTPGatewayActor {
	g := &HTTPGatewayActor{
		supervisorID: supervisorID,
	}
	g.actor = NewActor(name, router, g.Receive)
	g.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: g,
	}
	return g
}

func (g *HTTPGatewayActor) Start() {
	g.actor.Start()
	go func() {
		if err := g.server.ListenAndServe(); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()
	log.Printf("HTTP gateway actor %s started on port %d", g.actor.Name, g.server.Addr)
}

func (g *HTTPGatewayActor) Stop() {
	g.actor.Stop()
	g.server.Shutdown(context.Background())
}

// Receive handles messages for the gateway actor.
func (g *HTTPGatewayActor) Receive(msg ActorMsg) {
	// The gateway itself doesn't handle incoming actor messages directly for now.
	log.Printf("HTTPGatewayActor %s received unhandled message: %+v", g.actor.Name, msg)
}

// ServeHTTP makes HTTPGatewayActor an http.Handler.
func (g *HTTPGatewayActor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/eval" {
		http.NotFound(w, r)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	code := string(body)

	// 1. Ask the supervisor to create a temporary worker actor for this request.
	createReplyChan := make(chan ActorMsg, 1)
	createMsg := ActorMsg{
		To:        g.supervisorID,
		From:      g.actor.Name,
		Type:      "create_child: RelayServerActor",
		ReplyChan: createReplyChan,
	}
	g.actor.router.Send(createMsg)

	var workerName string
	select {
	case reply := <-createReplyChan:
		workerName = reply.Data.(string)
	case <-time.After(2 * time.Second):
		http.Error(w, "Timeout creating worker actor", http.StatusInternalServerError)
		return
	}
	defer func() {
		// 5. Tell the supervisor to stop the temporary worker.
		stopMsg := ActorMsg{
			To:   g.supervisorID,
			From: g.actor.Name,
			Type: "stop_child",
			Data: workerName,
		}
		g.actor.router.Send(stopMsg)
	}()

	// 2. Send code to the new worker for evaluation.
	evalReplyChan := make(chan ActorMsg, 1)
	evalMsg := ActorMsg{
		To:        workerName,
		From:      g.actor.Name,
		Type:      "eval",
		Data:      code,
		ReplyChan: evalReplyChan,
	}
	g.actor.router.Send(evalMsg)

	// 3. Wait for the result and write it to the response.
	select {
	case resultMsg := <-evalReplyChan:
		result, ok := resultMsg.Data.(*runtime.Value)
		if !ok {
			http.Error(w, "Invalid eval result from actor", http.StatusInternalServerError)
			return
		}
		// 4. Write the successful response.
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(result.String()))

	case <-time.After(2 * time.Second):
		http.Error(w, "Timeout waiting for eval result", http.StatusInternalServerError)
	}
}
