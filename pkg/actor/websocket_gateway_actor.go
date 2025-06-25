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
}

// NewWebSocketGatewayActor creates a new WebSocketGatewayActor.
func NewWebSocketGatewayActor(name string, router *Router) *WebSocketGatewayActor {
	g := &WebSocketGatewayActor{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
	}
	g.actor = NewActor(name, router, g.Receive)
	return g
}

// Receive handles messages for the gateway actor.
func (g *WebSocketGatewayActor) Receive(msg ActorMsg) {}

// ServeHTTP handles incoming HTTP requests and upgrades them to WebSockets.
func (g *WebSocketGatewayActor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/ws" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

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
		msg := NewWebSocketRequestMsg(targetActorName, g.actor.Name, request)
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
