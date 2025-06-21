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
	// We are only handling /eval for this integration
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

	// This is a blocking request-reply, which requires a temporary listener.
	replyChan := make(chan ActorMsg, 1)
	tempActorName := g.actor.Name + "-reply-proxy"

	// A temporary actor to receive the reply from the RelayServerActor
	replyProxy := NewActor(tempActorName, g.actor.router, func(msg ActorMsg) {
		replyChan <- msg
	})
	replyProxy.Start()
	defer replyProxy.Stop()

	// 1. Ask supervisor to create a child
	createMsg := ActorMsg{
		To:   g.supervisorName,
		From: tempActorName,
		Type: "create_child",
		Data: "RelayServerActor",
	}
	g.actor.router.Send(createMsg)

	// Wait for the supervisor's reply with the new child's name
	var childName string
	select {
	case reply := <-replyChan:
		if reply.Type == "child_created" {
			childName, _ = reply.Data.(string)
		} else {
			http.Error(w, "Failed to create child actor", http.StatusInternalServerError)
			return
		}
	case <-time.After(2 * time.Second):
		http.Error(w, "Timeout waiting for child actor creation", http.StatusInternalServerError)
		return
	}

	// 2. Send code to child for evaluation
	evalMsg := ActorMsg{
		To:   childName,
		From: tempActorName,
		Type: "eval",
		Data: code,
	}
	g.actor.router.Send(evalMsg)

	// 3. Wait for the result
	select {
	case resultMsg := <-replyChan:
		if resultMsg.Type == "eval_result" {
			result, ok := resultMsg.Data.(*runtime.Value)
			if !ok {
				http.Error(w, "Invalid eval result from actor", http.StatusInternalServerError)
				return
			}
			// Write result to HTTP response
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(result.String()))
		} else {
			http.Error(w, "Received unexpected message from actor", http.StatusInternalServerError)
			return
		}
	case <-time.After(2 * time.Second):
		http.Error(w, "Timeout waiting for eval result", http.StatusInternalServerError)
		return
	}

	// 4. Tell supervisor to stop the child
	stopMsg := ActorMsg{
		To:   g.supervisorName,
		From: g.actor.Name,
		Type: "stop_child",
		Data: childName,
	}
	g.actor.router.Send(stopMsg)
}
