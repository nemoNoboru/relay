package actor

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// WebSocketGatewayActor handles WebSocket connections and translates messages.
type WebSocketGatewayActor struct {
	actor    *Actor
	upgrader websocket.Upgrader
	server   *http.Server
}

// NewWebSocketGatewayActor creates a new WebSocketGatewayActor.
func NewWebSocketGatewayActor(name string, addr string, router *Router) *WebSocketGatewayActor {
	g := &WebSocketGatewayActor{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
	}
	g.actor = NewActor(name, router, g.Receive)
	g.server = &http.Server{Addr: addr, Handler: g}
	return g
}

func (g *WebSocketGatewayActor) Start() {
	g.actor.Start()
	go func() {
		log.Printf("WebSocket Gateway %s starting on %s", g.actor.Name, g.server.Addr)
		if err := g.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("WebSocket Gateway %s failed: %v", g.actor.Name, err)
		}
	}()
}

func (g *WebSocketGatewayActor) Stop() {
	log.Printf("WebSocket Gateway %s stopping", g.actor.Name)
	if err := g.server.Shutdown(nil); err != nil {
		log.Printf("WebSocket Gateway %s shutdown error: %v", g.actor.Name, err)
	}
	g.actor.Stop()
	log.Printf("WebSocket Gateway %s stopped", g.actor.Name)
}

// Receive handles messages for the gateway actor.
func (g *WebSocketGatewayActor) Receive(msg ActorMsg) {}

// ServeHTTP handles incoming HTTP requests and upgrades them to WebSockets.
func (g *WebSocketGatewayActor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := g.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket connection: %v", err)
		return
	}
	defer conn.Close()
	g.handleConnection(conn)
}

// handleConnection reads and processes messages from a single WebSocket client.
func (g *WebSocketGatewayActor) handleConnection(conn *websocket.Conn) {
	for {
		var request JSONRPCRequest
		err := conn.ReadJSON(&request)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket client connection error: %v", err)
			}
			break
		}

		targetActorName := request.Method
		msg := ActorMsg{
			To:   targetActorName,
			From: g.actor.Name,
			Type: "websocket_request",
			Data: request,
		}
		g.actor.router.Send(msg)

		// Dummy response
		response := JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  "message sent",
			ID:      request.ID,
		}
		if err := conn.WriteJSON(response); err != nil {
			log.Printf("Failed to write JSON to WebSocket: %v", err)
			break
		}
	}
}
