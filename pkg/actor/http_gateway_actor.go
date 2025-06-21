package actor

import (
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
type HTTPGatewayActor struct {
	actor          *Actor
	server         *http.Server
	supervisorName string
}

// NewHTTPGatewayActor creates a new HTTPGatewayActor.
func NewHTTPGatewayActor(name, addr, supervisorName string, router *Router) *HTTPGatewayActor {
	g := &HTTPGatewayActor{
		supervisorName: supervisorName,
	}
	g.actor = NewActor(name, router, g.Receive)
	g.server = &http.Server{Addr: addr, Handler: g}
	return g
}

func (g *HTTPGatewayActor) Start() {
	g.actor.Start()
	go func() {
		log.Printf("HTTP Gateway %s starting on %s", g.actor.Name, g.server.Addr)
		if err := g.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP Gateway %s failed: %v", g.actor.Name, err)
		}
	}()
}

func (g *HTTPGatewayActor) Stop() {
	log.Printf("HTTP Gateway %s stopping", g.actor.Name)
	if err := g.server.Shutdown(nil); err != nil {
		log.Printf("HTTP Gateway %s shutdown error: %v", g.actor.Name, err)
	}
	g.actor.Stop()
	log.Printf("HTTP Gateway %s stopped", g.actor.Name)
}

// Receive handles messages for the gateway actor.
// For now, it doesn't need to do anything.
func (g *HTTPGatewayActor) Receive(msg ActorMsg) {}

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

	// 1. Ask supervisor to create a child, waiting for a direct reply.
	replyChan := make(chan ActorMsg, 1)
	createMsg := ActorMsg{
		To:        g.supervisorName,
		From:      g.actor.Name,
		Type:      "create_child",
		Data:      "RelayServerActor",
		ReplyChan: replyChan,
	}
	g.actor.router.Send(createMsg)

	var childName string
	select {
	case reply := <-replyChan:
		var ok bool
		childName, ok = reply.Data.(string)
		if !ok || childName == "" {
			http.Error(w, "Failed to create child actor", http.StatusInternalServerError)
			return
		}
	case <-time.After(2 * time.Second):
		http.Error(w, "Timeout waiting for child actor creation", http.StatusInternalServerError)
		return
	}
	defer func() {
		// 4. Tell supervisor to stop the child once we're done.
		stopMsg := ActorMsg{
			To:   g.supervisorName,
			From: g.actor.Name,
			Type: "stop_child",
			Data: childName,
		}
		g.actor.router.Send(stopMsg)
	}()

	// 2. Send code to the new child for evaluation, waiting for a direct reply.
	evalReplyChan := make(chan ActorMsg, 1)
	evalMsg := ActorMsg{
		To:        childName,
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
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(result.String()))

	case <-time.After(2 * time.Second):
		http.Error(w, "Timeout waiting for eval result", http.StatusInternalServerError)
		return
	}
}
