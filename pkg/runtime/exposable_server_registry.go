package runtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// ExposableServerRegistry provides HTTP endpoints for server discovery and management
// This enables peer-to-peer functionality by allowing Relay nodes to discover and
// communicate with servers on other nodes
type ExposableServerRegistry struct {
	registry     ServerRegistry
	nodeID       string
	nodeAddress  string
	peers        map[string]*PeerNode
	peersMutex   sync.RWMutex
	localServers map[string]*ServerInfo
	serversMutex sync.RWMutex
}

// ServerInfo contains metadata about a server for discovery
type ServerInfo struct {
	Name        string            `json:"name"`
	NodeID      string            `json:"node_id"`
	NodeAddress string            `json:"node_address"`
	Methods     []string          `json:"methods"`
	State       map[string]string `json:"state,omitempty"` // State types, not values
	LastSeen    time.Time         `json:"last_seen"`
	IsLocal     bool              `json:"is_local"`
}

// PeerNode represents a peer Relay node
type PeerNode struct {
	NodeID      string    `json:"node_id"`
	Address     string    `json:"address"`
	LastSeen    time.Time `json:"last_seen"`
	IsHealthy   bool      `json:"is_healthy"`
	ServerCount int       `json:"server_count"`
}

// ServerRegistryResponse represents the response from /registry endpoint
type ServerRegistryResponse struct {
	NodeID      string                 `json:"node_id"`
	NodeAddress string                 `json:"node_address"`
	Servers     map[string]*ServerInfo `json:"servers"`
	Peers       map[string]*PeerNode   `json:"peers"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NewExposableServerRegistry creates a new exposable server registry
func NewExposableServerRegistry(registry ServerRegistry, nodeID, nodeAddress string) *ExposableServerRegistry {
	return &ExposableServerRegistry{
		registry:     registry,
		nodeID:       nodeID,
		nodeAddress:  nodeAddress,
		peers:        make(map[string]*PeerNode),
		localServers: make(map[string]*ServerInfo),
	}
}

// RegisterServer registers a server locally and updates discovery info
func (r *ExposableServerRegistry) RegisterServer(name string, server *Value) {
	// Register with underlying registry
	r.registry.RegisterServer(name, server)

	// Update local server info for discovery
	r.serversMutex.Lock()
	defer r.serversMutex.Unlock()

	methods := []string{}
	if server.Type == ValueTypeServer && server.Server != nil {
		for methodName := range server.Server.Receivers {
			methods = append(methods, methodName)
		}
	}

	r.localServers[name] = &ServerInfo{
		Name:        name,
		NodeID:      r.nodeID,
		NodeAddress: r.nodeAddress,
		Methods:     methods,
		LastSeen:    time.Now(),
		IsLocal:     true,
	}
}

// GetServer retrieves a server by name (local or remote)
func (r *ExposableServerRegistry) GetServer(name string) (*Value, bool) {
	// First try local registry
	if server, exists := r.registry.GetServer(name); exists {
		return server, true
	}

	// TODO: In future, check remote peers for the server
	// For now, just return local result
	return nil, false
}

// StopAllServers stops all local servers
func (r *ExposableServerRegistry) StopAllServers() {
	r.registry.StopAllServers()

	// Clear local server info
	r.serversMutex.Lock()
	defer r.serversMutex.Unlock()
	r.localServers = make(map[string]*ServerInfo)
}

// GetAllServers returns information about all known servers (local and remote)
func (r *ExposableServerRegistry) GetAllServers() map[string]*ServerInfo {
	r.serversMutex.RLock()
	defer r.serversMutex.RUnlock()

	result := make(map[string]*ServerInfo)
	for name, info := range r.localServers {
		result[name] = info
	}

	// TODO: Add remote servers from peers

	return result
}

// AddPeer adds a peer node to the registry
func (r *ExposableServerRegistry) AddPeer(nodeID, address string) {
	r.peersMutex.Lock()
	defer r.peersMutex.Unlock()

	r.peers[nodeID] = &PeerNode{
		NodeID:    nodeID,
		Address:   address,
		LastSeen:  time.Now(),
		IsHealthy: true,
	}
}

// RemovePeer removes a peer node from the registry
func (r *ExposableServerRegistry) RemovePeer(nodeID string) {
	r.peersMutex.Lock()
	defer r.peersMutex.Unlock()

	delete(r.peers, nodeID)
}

// GetPeers returns all known peer nodes
func (r *ExposableServerRegistry) GetPeers() map[string]*PeerNode {
	r.peersMutex.RLock()
	defer r.peersMutex.RUnlock()

	result := make(map[string]*PeerNode)
	for nodeID, peer := range r.peers {
		result[nodeID] = peer
	}

	return result
}

// SetupHTTPEndpoints adds server registry endpoints to an HTTP mux
func (r *ExposableServerRegistry) SetupHTTPEndpoints(mux *http.ServeMux) {
	// GET /registry - Returns full server registry
	mux.HandleFunc("/registry", r.handleRegistry)

	// GET /registry/servers - Returns just server list
	mux.HandleFunc("/registry/servers", r.handleServers)

	// GET /registry/peers - Returns peer nodes
	mux.HandleFunc("/registry/peers", r.handlePeers)

	// POST /registry/peers - Add a peer node
	mux.HandleFunc("/registry/peers/add", r.handleAddPeer)

	// DELETE /registry/peers/{nodeID} - Remove a peer node
	mux.HandleFunc("/registry/peers/remove", r.handleRemovePeer)

	// GET /registry/health - Health check for registry
	mux.HandleFunc("/registry/health", r.handleRegistryHealth)
}

// handleRegistry returns the complete server registry
func (r *ExposableServerRegistry) handleRegistry(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := ServerRegistryResponse{
		NodeID:      r.nodeID,
		NodeAddress: r.nodeAddress,
		Servers:     r.GetAllServers(),
		Peers:       r.GetPeers(),
		Timestamp:   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleServers returns just the server list
func (r *ExposableServerRegistry) handleServers(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	servers := r.GetAllServers()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(servers)
}

// handlePeers returns the peer nodes
func (r *ExposableServerRegistry) handlePeers(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	peers := r.GetPeers()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peers)
}

// handleAddPeer adds a new peer node
func (r *ExposableServerRegistry) handleAddPeer(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var peerRequest struct {
		NodeID  string `json:"node_id"`
		Address string `json:"address"`
	}

	if err := json.NewDecoder(req.Body).Decode(&peerRequest); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if peerRequest.NodeID == "" || peerRequest.Address == "" {
		http.Error(w, "node_id and address are required", http.StatusBadRequest)
		return
	}

	r.AddPeer(peerRequest.NodeID, peerRequest.Address)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Peer %s added successfully", peerRequest.NodeID),
	})
}

// handleRemovePeer removes a peer node
func (r *ExposableServerRegistry) handleRemovePeer(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var removeRequest struct {
		NodeID string `json:"node_id"`
	}

	if err := json.NewDecoder(req.Body).Decode(&removeRequest); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if removeRequest.NodeID == "" {
		http.Error(w, "node_id is required", http.StatusBadRequest)
		return
	}

	r.RemovePeer(removeRequest.NodeID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Peer %s removed successfully", removeRequest.NodeID),
	})
}

// handleRegistryHealth returns health status of the registry
func (r *ExposableServerRegistry) handleRegistryHealth(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	servers := r.GetAllServers()
	peers := r.GetPeers()

	healthResponse := map[string]interface{}{
		"status":       "healthy",
		"node_id":      r.nodeID,
		"node_address": r.nodeAddress,
		"server_count": len(servers),
		"peer_count":   len(peers),
		"timestamp":    time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(healthResponse)
}

// DiscoverPeers attempts to discover servers on peer nodes
func (r *ExposableServerRegistry) DiscoverPeers() error {
	r.peersMutex.RLock()
	peers := make(map[string]*PeerNode)
	for nodeID, peer := range r.peers {
		peers[nodeID] = peer
	}
	r.peersMutex.RUnlock()

	for nodeID, peer := range peers {
		err := r.fetchPeerRegistry(nodeID, peer)
		if err != nil {
			// Mark peer as unhealthy but don't fail the whole operation
			r.peersMutex.Lock()
			if p, exists := r.peers[nodeID]; exists {
				p.IsHealthy = false
				p.LastSeen = time.Now()
			}
			r.peersMutex.Unlock()
		}
	}

	return nil
}

// fetchPeerRegistry fetches the server registry from a peer node
func (r *ExposableServerRegistry) fetchPeerRegistry(nodeID string, peer *PeerNode) error {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(fmt.Sprintf("http://%s/registry/servers", peer.Address))
	if err != nil {
		return fmt.Errorf("failed to fetch registry from peer %s: %v", nodeID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("peer %s returned status %d", nodeID, resp.StatusCode)
	}

	var peerServers map[string]*ServerInfo
	if err := json.NewDecoder(resp.Body).Decode(&peerServers); err != nil {
		return fmt.Errorf("failed to decode peer registry: %v", err)
	}

	// Update peer health and server count
	r.peersMutex.Lock()
	if p, exists := r.peers[nodeID]; exists {
		p.IsHealthy = true
		p.LastSeen = time.Now()
		p.ServerCount = len(peerServers)
	}
	r.peersMutex.Unlock()

	// TODO: Store remote server info for future routing
	// For now, just log that we discovered them
	fmt.Printf("Discovered %d servers on peer %s\n", len(peerServers), nodeID)

	return nil
}

// StartPeriodicDiscovery starts a background goroutine that periodically discovers peers
func (r *ExposableServerRegistry) StartPeriodicDiscovery(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := r.DiscoverPeers(); err != nil {
				fmt.Printf("Error during peer discovery: %v\n", err)
			}
		}
	}()
}

// GetNodeInfo returns information about this node
func (r *ExposableServerRegistry) GetNodeInfo() map[string]interface{} {
	servers := r.GetAllServers()
	peers := r.GetPeers()

	return map[string]interface{}{
		"node_id":      r.nodeID,
		"node_address": r.nodeAddress,
		"server_count": len(servers),
		"peer_count":   len(peers),
		"uptime":       time.Now(), // Could track actual uptime
	}
}

// GetRegistryInfo returns the complete registry information
func (r *ExposableServerRegistry) GetRegistryInfo() *ServerRegistryResponse {
	return &ServerRegistryResponse{
		NodeID:      r.nodeID,
		NodeAddress: r.nodeAddress,
		Servers:     r.GetAllServers(),
		Peers:       r.GetPeers(),
		Timestamp:   time.Now(),
	}
}

// AddPeerFromURL adds a peer by discovering it from a URL
func (r *ExposableServerRegistry) AddPeerFromURL(peerURL string) error {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(fmt.Sprintf("%s/registry/health", peerURL))
	if err != nil {
		return fmt.Errorf("failed to connect to peer at %s: %v", peerURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("peer at %s returned status %d", peerURL, resp.StatusCode)
	}

	var healthResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&healthResponse); err != nil {
		return fmt.Errorf("failed to decode peer health response: %v", err)
	}

	nodeID, ok := healthResponse["node_id"].(string)
	if !ok {
		return fmt.Errorf("peer did not provide node_id")
	}

	nodeAddress, ok := healthResponse["node_address"].(string)
	if !ok {
		return fmt.Errorf("peer did not provide node_address")
	}

	r.AddPeer(nodeID, nodeAddress)
	return nil
}

// IsLocalServer checks if a server is local to this node
func (r *ExposableServerRegistry) IsLocalServer(serverName string) bool {
	r.serversMutex.RLock()
	defer r.serversMutex.RUnlock()

	if info, exists := r.localServers[serverName]; exists {
		return info.IsLocal
	}

	return false
}

// GetServerInfo returns detailed information about a specific server
func (r *ExposableServerRegistry) GetServerInfo(serverName string) (*ServerInfo, bool) {
	r.serversMutex.RLock()
	defer r.serversMutex.RUnlock()

	if info, exists := r.localServers[serverName]; exists {
		return info, true
	}

	// TODO: Check remote servers from peers
	return nil, false
}
