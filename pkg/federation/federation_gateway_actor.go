package federation

import (
	"log"
	"net"
	"net/http"
	"sync"

	"relay/pkg/actor"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type PeerConnection struct {
	Conn *websocket.Conn
	mu   sync.Mutex // Protects writes to the connection
}

func (pc *PeerConnection) WriteJSON(v interface{}) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	return pc.Conn.WriteJSON(v)
}

type FederationGatewayActor struct {
	*actor.Actor
	peers    map[string]*PeerConnection
	mu       sync.RWMutex
	listener net.Listener
	server   *http.Server
}

func NewFederationGatewayActor(name string, router *actor.Router) *FederationGatewayActor {
	fga := &FederationGatewayActor{}
	fga.Actor = actor.NewActor(name, router, fga.Receive)
	fga.peers = make(map[string]*PeerConnection)
	return fga
}

func (fga *FederationGatewayActor) Receive(msg actor.ActorMsg) {
	switch msg.Type {
	case "connect_to_peer":
		peerURL, ok := msg.Data.(string)
		if !ok {
			log.Printf("[%s] invalid data for 'connect_to_peer', expected URL string", fga.Name)
			return
		}
		go fga.connectToPeer(peerURL)
	case "forward_message":
		innerMsg, ok := msg.Data.(actor.ActorMsg)
		if !ok {
			// It might be a map from JSON deserialization, try to convert it
			msgMap, ok := msg.Data.(map[string]interface{})
			if !ok {
				log.Printf("[%s] invalid data for 'forward_message', expected ActorMsg or map", fga.Name)
				return
			}
			innerMsg = actor.ActorMsg{
				To:   msgMap["To"].(string),
				From: msgMap["From"].(string),
				Type: msgMap["Type"].(string),
				Data: msgMap["Data"],
			}
		}
		fga.broadcast(innerMsg)
	default:
		log.Printf("FederationGatewayActor %s received unhandled message: %v", fga.Name, msg)
	}
}

func (fga *FederationGatewayActor) connectToPeer(peerURL string) {
	log.Printf("[%s] Connecting to peer: %s", fga.Name, peerURL)
	c, _, err := websocket.DefaultDialer.Dial(peerURL, nil)
	if err != nil {
		log.Printf("[%s] Error connecting to peer %s: %v", fga.Name, peerURL, err)
		return
	}

	pc := &PeerConnection{Conn: c}
	fga.mu.Lock()
	fga.peers[peerURL] = pc
	fga.mu.Unlock()

	log.Printf("[%s] Successfully connected to peer: %s", fga.Name, peerURL)
	go fga.readLoop(pc)
}

func (fga *FederationGatewayActor) GetPeers() map[string]*PeerConnection {
	fga.mu.RLock()
	defer fga.mu.RUnlock()
	peersCopy := make(map[string]*PeerConnection, len(fga.peers))
	for k, v := range fga.peers {
		peersCopy[k] = v
	}
	return peersCopy
}

func (fga *FederationGatewayActor) StartListening(addr string) error {
	fga.mu.Lock()
	defer fga.mu.Unlock()

	if fga.listener != nil {
		log.Printf("[%s] Already listening", fga.Name)
		return nil
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	fga.listener = listener

	server := &http.Server{Handler: fga}
	fga.server = server

	go func() {
		log.Printf("[%s] Starting federation listener on %s", fga.Name, fga.listener.Addr().String())
		if err := server.Serve(fga.listener); err != http.ErrServerClosed {
			log.Printf("[%s] HTTP server error: %v", fga.Name, err)
		}
	}()

	return nil
}

func (fga *FederationGatewayActor) StopListening() {
	fga.mu.Lock()
	defer fga.mu.Unlock()

	if fga.server != nil {
		log.Printf("[%s] Stopping federation listener", fga.Name)
		fga.server.Close()
		fga.server = nil
		fga.listener = nil
	}
}

func (fga *FederationGatewayActor) GetListenURL() string {
	fga.mu.RLock()
	defer fga.mu.RUnlock()

	if fga.listener == nil {
		return ""
	}
	return "ws://" + fga.listener.Addr().String()
}

func (fga *FederationGatewayActor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[%s] Failed to upgrade connection: %v", fga.Name, err)
		return
	}

	peerAddr := conn.RemoteAddr().String()
	log.Printf("[%s] Accepted new peer connection from %s", fga.Name, peerAddr)

	pc := &PeerConnection{Conn: conn}
	fga.mu.Lock()
	fga.peers[peerAddr] = pc
	fga.mu.Unlock()
	go fga.readLoop(pc)
}

func (fga *FederationGatewayActor) broadcast(msg actor.ActorMsg) {
	fga.mu.RLock()
	defer fga.mu.RUnlock()
	log.Printf("[%s] Broadcasting message to %d peers", fga.Name, len(fga.peers))
	for peerURL, pc := range fga.peers {
		log.Printf("[%s] Forwarding message to peer %s", fga.Name, peerURL)
		if err := pc.WriteJSON(msg); err != nil {
			log.Printf("[%s] Error writing to peer %s: %v", fga.Name, peerURL, err)
			// TODO: Handle failed write (e.g., remove peer)
		}
	}
}

func (fga *FederationGatewayActor) readLoop(pc *PeerConnection) {
	defer func() {
		// Clean up connection
		pc.Conn.Close()
		fga.mu.Lock()
		for url, peer := range fga.peers {
			if peer == pc {
				delete(fga.peers, url)
				log.Printf("[%s] Removed peer connection: %s", fga.Name, url)
				break
			}
		}
		fga.mu.Unlock()
	}()

	for {
		var msg actor.ActorMsg
		if err := pc.Conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[%s] Peer connection read error: %v", fga.Name, err)
			} else {
				log.Printf("[%s] Peer connection closed", fga.Name)
			}
			break
		}
		log.Printf("[%s] Received message from peer: %+v", fga.Name, msg)
		fga.Router().Send(msg)
	}
}
