package network

import (
	"BkC/blockchain"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// NodeID represents a node's unique identifier in the DHT
type NodeID [20]byte

// KBucketSize is the maximum size of a k-bucket in the Kademlia DHT
const KBucketSize = 20

// KademliaNode represents a node in the Kademlia DHT
type KademliaNode struct {
	ID       NodeID    `json:"id"`
	URL      string    `json:"url"`
	LastSeen time.Time `json:"lastSeen"`
	Status   string    `json:"status"`
}

// EnhancedNetworkManager extends the basic NetworkManager with Kademlia DHT
type EnhancedNetworkManager struct {
	*NetworkManager
	nodeID         NodeID
	kBuckets       [160][]*KademliaNode // 160 bits = 20 bytes
	routingMutex   sync.RWMutex
	storageData    map[string][]byte
	storageMutex   sync.RWMutex
	bootstrapNodes []string
}

// NewEnhancedNetworkManager creates a new EnhancedNetworkManager with Kademlia DHT
func NewEnhancedNetworkManager(nodeURL string, bc *blockchain.Blockchain, isValidator bool) *EnhancedNetworkManager {
	// Create base network manager
	baseManager := NewNetworkManager(nodeURL, bc, isValidator)

	// Generate node ID from URL
	hash := sha256.Sum256([]byte(nodeURL))
	var nodeID NodeID
	copy(nodeID[:], hash[:20])

	// Create enhanced manager
	enhancedManager := &EnhancedNetworkManager{
		NetworkManager: baseManager,
		nodeID:         nodeID,
		kBuckets:       [160][]*KademliaNode{},
		storageData:    make(map[string][]byte),
		bootstrapNodes: []string{},
	}

	return enhancedManager
}

// ComputeDistance calculates the XOR distance between two node IDs
func ComputeDistance(a, b NodeID) *big.Int {
	distance := new(big.Int)

	// Convert node IDs to big.Int
	aInt := new(big.Int).SetBytes(a[:])
	bInt := new(big.Int).SetBytes(b[:])

	// XOR the values
	distance.Xor(aInt, bInt)

	return distance
}

// FindBucketIndex determines which k-bucket a node belongs to
func (enm *EnhancedNetworkManager) FindBucketIndex(id NodeID) int {
	distance := ComputeDistance(enm.nodeID, id)

	// If the distance is zero (same node), return -1
	if distance.BitLen() == 0 {
		return -1
	}

	// Find the position of the first set bit (the leftmost 1)
	// This determines the bucket index
	return 159 - distance.BitLen()
}

// AddToBucket adds a node to the appropriate k-bucket
func (enm *EnhancedNetworkManager) AddToBucket(node *KademliaNode) {
	bucketIndex := enm.FindBucketIndex(node.ID)
	if bucketIndex == -1 {
		return // Don't add self to bucket
	}

	enm.routingMutex.Lock()
	defer enm.routingMutex.Unlock()

	// Check if node already exists in bucket
	for i, n := range enm.kBuckets[bucketIndex] {
		if n.ID == node.ID {
			// Move to the end of the bucket (most recently seen)
			enm.kBuckets[bucketIndex] = append(enm.kBuckets[bucketIndex][:i], enm.kBuckets[bucketIndex][i+1:]...)
			enm.kBuckets[bucketIndex] = append(enm.kBuckets[bucketIndex], node)
			return
		}
	}

	// If bucket is not full, add the node
	if len(enm.kBuckets[bucketIndex]) < KBucketSize {
		enm.kBuckets[bucketIndex] = append(enm.kBuckets[bucketIndex], node)
		return
	}

	// Bucket is full, ping the least recently seen node
	leastRecentNode := enm.kBuckets[bucketIndex][0]

	// If node responds, move it to the end and discard the new node
	// Otherwise, remove it and add the new node
	if enm.pingNode(leastRecentNode.URL) {
		// Move to the end
		enm.kBuckets[bucketIndex] = enm.kBuckets[bucketIndex][1:]
		enm.kBuckets[bucketIndex] = append(enm.kBuckets[bucketIndex], leastRecentNode)
	} else {
		// Remove and add new node
		enm.kBuckets[bucketIndex] = enm.kBuckets[bucketIndex][1:]
		enm.kBuckets[bucketIndex] = append(enm.kBuckets[bucketIndex], node)
	}
}

// FindClosestNodes finds the k closest nodes to a target ID
func (enm *EnhancedNetworkManager) FindClosestNodes(targetID NodeID, k int) []*KademliaNode {
	enm.routingMutex.RLock()
	defer enm.routingMutex.RUnlock()

	// Gather all nodes from all buckets
	var allNodes []*KademliaNode
	for _, bucket := range enm.kBuckets {
		allNodes = append(allNodes, bucket...)
	}

	// Sort by distance to target
	sort.Slice(allNodes, func(i, j int) bool {
		distI := ComputeDistance(allNodes[i].ID, targetID)
		distJ := ComputeDistance(allNodes[j].ID, targetID)
		return distI.Cmp(distJ) < 0
	})

	// Return the k closest
	if len(allNodes) > k {
		return allNodes[:k]
	}
	return allNodes
}

// NodeLookup performs a Kademlia node lookup procedure
func (enm *EnhancedNetworkManager) NodeLookup(targetID NodeID) []*KademliaNode {
	// Start with alpha closest nodes from local k-buckets
	alpha := 3 // Parallelism parameter
	closestNodes := enm.FindClosestNodes(targetID, alpha)

	// Keep track of queried nodes
	queriedNodes := make(map[string]bool)

	// Create a shortlist of nodes to query
	shortlist := make([]*KademliaNode, len(closestNodes))
	copy(shortlist, closestNodes)

	// Keep iterating until we can't find any closer nodes
	for {
		if len(shortlist) == 0 {
			break
		}

		// Select alpha nodes from the shortlist that haven't been queried
		var nodesToQuery []*KademliaNode
		for _, node := range shortlist {
			if !queriedNodes[node.URL] {
				nodesToQuery = append(nodesToQuery, node)
				queriedNodes[node.URL] = true

				if len(nodesToQuery) >= alpha {
					break
				}
			}
		}

		if len(nodesToQuery) == 0 {
			break
		}

		// Query each selected node in parallel
		var wg sync.WaitGroup
		var resultMutex sync.Mutex
		var newNodes []*KademliaNode

		for _, node := range nodesToQuery {
			wg.Add(1)
			go func(n *KademliaNode) {
				defer wg.Done()

				// Find nodes via API
				nodes := enm.findNodesFromPeer(n.URL, targetID)

				resultMutex.Lock()
				newNodes = append(newNodes, nodes...)
				resultMutex.Unlock()
			}(node)
		}

		wg.Wait()

		// Add new nodes to the shortlist
		for _, node := range newNodes {
			// Skip if we've already queried this node
			if queriedNodes[node.URL] {
				continue
			}

			// Add to shortlist
			shortlist = append(shortlist, node)
		}

		// Re-sort the shortlist by distance
		sort.Slice(shortlist, func(i, j int) bool {
			distI := ComputeDistance(shortlist[i].ID, targetID)
			distJ := ComputeDistance(shortlist[j].ID, targetID)
			return distI.Cmp(distJ) < 0
		})

		// Trim to k closest nodes
		if len(shortlist) > KBucketSize {
			shortlist = shortlist[:KBucketSize]
		}
	}

	return shortlist
}

// findNodesFromPeer asks a peer for nodes close to targetID
func (enm *EnhancedNetworkManager) findNodesFromPeer(peerURL string, targetID NodeID) []*KademliaNode {
	// Create the request
	targetIDHex := hex.EncodeToString(targetID[:])
	requestURL := fmt.Sprintf("%s/p2p/find_node?target=%s", peerURL, targetIDHex)

	// Send the request
	resp, err := http.Get(requestURL)
	if err != nil {
		log.Printf("[P2P] Error requesting nodes from %s: %v", peerURL, err)
		return nil
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[P2P] Error reading response from %s: %v", peerURL, err)
		return nil
	}

	// Parse the response
	var response struct {
		Success bool            `json:"success"`
		Nodes   []*KademliaNode `json:"nodes"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("[P2P] Error parsing response from %s: %v", peerURL, err)
		return nil
	}

	if !response.Success {
		log.Printf("[P2P] Unsuccessful response from %s", peerURL)
		return nil
	}

	return response.Nodes
}

// HandleFindNodeRequest handles a find_node request from another peer
func (enm *EnhancedNetworkManager) HandleFindNodeRequest(w http.ResponseWriter, r *http.Request) {
	// Get the target ID from the query string
	targetIDHex := r.URL.Query().Get("target")
	if targetIDHex == "" {
		http.Error(w, "Target ID is required", http.StatusBadRequest)
		return
	}

	// Decode the target ID
	targetIDBytes, err := hex.DecodeString(targetIDHex)
	if err != nil {
		http.Error(w, "Invalid target ID", http.StatusBadRequest)
		return
	}

	var targetID NodeID
	copy(targetID[:], targetIDBytes)

	// Find closest nodes
	closestNodes := enm.FindClosestNodes(targetID, KBucketSize)

	// Respond with the nodes
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"nodes":   closestNodes,
	})
}

// Bootstrap initializes the routing table with the bootstrap nodes
func (enm *EnhancedNetworkManager) Bootstrap(bootstrapNodes []string) {
	enm.bootstrapNodes = bootstrapNodes

	// Add bootstrap nodes to routing table
	for _, url := range bootstrapNodes {
		if url == enm.NodeURL {
			continue // Skip self
		}

		// Generate node ID from URL
		hash := sha256.Sum256([]byte(url))
		var nodeID NodeID
		copy(nodeID[:], hash[:20])

		node := &KademliaNode{
			ID:       nodeID,
			URL:      url,
			LastSeen: time.Time{},
			Status:   NodeStatusInactive,
		}

		enm.AddToBucket(node)

		// Ping the node to update its status
		enm.pingNode(url)
	}

	// Perform node lookup for self to populate routing table
	enm.NodeLookup(enm.nodeID)
}

// Store stores a value in the DHT
func (enm *EnhancedNetworkManager) Store(key string, value []byte) {
	// Hash the key to get a node ID
	hash := sha256.Sum256([]byte(key))
	var targetID NodeID
	copy(targetID[:], hash[:20])

	// Find k closest nodes
	closestNodes := enm.NodeLookup(targetID)

	// Store locally
	enm.storageMutex.Lock()
	enm.storageData[key] = value
	enm.storageMutex.Unlock()

	// Store on the k closest nodes
	for _, node := range closestNodes {
		enm.storeOnNode(node.URL, key, value)
	}
}

// storeOnNode stores a key-value pair on a remote node
func (enm *EnhancedNetworkManager) storeOnNode(nodeURL, key string, value []byte) {
	// Create the request
	requestURL := fmt.Sprintf("%s/p2p/store", nodeURL)

	// Prepare the data
	data := map[string]interface{}{
		"key":   key,
		"value": hex.EncodeToString(value),
	}

	// Marshal the data
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("[P2P] Error marshaling store data: %v", err)
		return
	}

	// Send the request
	resp, err := http.Post(requestURL, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		log.Printf("[P2P] Error storing data on %s: %v", nodeURL, err)
		return
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		log.Printf("[P2P] Storage request to %s failed with status code %d", nodeURL, resp.StatusCode)
	}
}

// HandleStoreRequest handles a store request from another peer
func (enm *EnhancedNetworkManager) HandleStoreRequest(w http.ResponseWriter, r *http.Request) {
	// Decode the request
	var request struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Decode the value
	valueBytes, err := hex.DecodeString(request.Value)
	if err != nil {
		http.Error(w, "Invalid value", http.StatusBadRequest)
		return
	}

	// Store the value
	enm.storageMutex.Lock()
	enm.storageData[request.Key] = valueBytes
	enm.storageMutex.Unlock()

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// FindValue searches for a value in the DHT
func (enm *EnhancedNetworkManager) FindValue(key string) ([]byte, bool) {
	// Check local storage first
	enm.storageMutex.RLock()
	value, exists := enm.storageData[key]
	enm.storageMutex.RUnlock()

	if exists {
		return value, true
	}

	// Hash the key to get a node ID
	hash := sha256.Sum256([]byte(key))
	var targetID NodeID
	copy(targetID[:], hash[:20])

	// Find k closest nodes
	closestNodes := enm.NodeLookup(targetID)

	// Query each node for the value
	for _, node := range closestNodes {
		value, found := enm.findValueFromNode(node.URL, key)
		if found {
			// Store the value locally for caching
			enm.storageMutex.Lock()
			enm.storageData[key] = value
			enm.storageMutex.Unlock()

			return value, true
		}
	}

	return nil, false
}

// findValueFromNode queries a remote node for a value
func (enm *EnhancedNetworkManager) findValueFromNode(nodeURL, key string) ([]byte, bool) {
	// Create the request
	requestURL := fmt.Sprintf("%s/p2p/find_value?key=%s", nodeURL, key)

	// Send the request
	resp, err := http.Get(requestURL)
	if err != nil {
		log.Printf("[P2P] Error querying value from %s: %v", nodeURL, err)
		return nil, false
	}
	defer resp.Body.Close()

	// Parse the response
	var response struct {
		Success bool   `json:"success"`
		Found   bool   `json:"found"`
		Value   string `json:"value,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("[P2P] Error parsing response from %s: %v", nodeURL, err)
		return nil, false
	}

	if !response.Success || !response.Found {
		return nil, false
	}

	// Decode the value
	valueBytes, err := hex.DecodeString(response.Value)
	if err != nil {
		log.Printf("[P2P] Error decoding value from %s: %v", nodeURL, err)
		return nil, false
	}

	return valueBytes, true
}

// HandleFindValueRequest handles a find_value request from another peer
func (enm *EnhancedNetworkManager) HandleFindValueRequest(w http.ResponseWriter, r *http.Request) {
	// Get the key from the query string
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	// Check if we have the value
	enm.storageMutex.RLock()
	value, exists := enm.storageData[key]
	enm.storageMutex.RUnlock()

	// Prepare the response
	response := map[string]interface{}{
		"success": true,
		"found":   exists,
	}

	if exists {
		response["value"] = hex.EncodeToString(value)
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ExtendHandlers adds the Kademlia DHT handlers to the HTTP server
func (enm *EnhancedNetworkManager) ExtendHandlers(mux *http.ServeMux) {
	// Add handlers for DHT operations
	mux.HandleFunc("/p2p/find_node", enm.HandleFindNodeRequest)
	mux.HandleFunc("/p2p/store", enm.HandleStoreRequest)
	mux.HandleFunc("/p2p/find_value", enm.HandleFindValueRequest)
}

// String represents a NodeID as a hex string
func (id NodeID) String() string {
	return hex.EncodeToString(id[:])
}

// StringToNodeID converts a hex string to a NodeID
func StringToNodeID(s string) (NodeID, error) {
	var id NodeID
	bytes, err := hex.DecodeString(s)
	if err != nil {
		return id, err
	}

	if len(bytes) != 20 {
		return id, fmt.Errorf("invalid node ID length: got %d, want 20", len(bytes))
	}

	copy(id[:], bytes)
	return id, nil
}

// AutoRefresh periodically refreshes the routing table
func (enm *EnhancedNetworkManager) AutoRefresh(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Generate a random node ID
				var randomID NodeID
				binary.BigEndian.PutUint64(randomID[:8], uint64(time.Now().UnixNano()))

				// Perform a lookup to refresh routing table
				enm.NodeLookup(randomID)

				// Log the refresh
				log.Printf("[P2P] Refreshed routing table with %d nodes", enm.CountActiveNodes())
			}
		}
	}()
}

// CountActiveNodes counts the number of active nodes in the routing table
func (enm *EnhancedNetworkManager) CountActiveNodes() int {
	enm.routingMutex.RLock()
	defer enm.routingMutex.RUnlock()

	count := 0
	for _, bucket := range enm.kBuckets {
		for _, node := range bucket {
			if node.Status == NodeStatusActive || node.Status == NodeStatusValidator {
				count++
			}
		}
	}

	return count
}

// BroadcastBlockDHT broadcasts a block using the DHT for efficient propagation
func (enm *EnhancedNetworkManager) BroadcastBlockDHT(block *blockchain.Block) {
	// Serialize the block
	blockData, err := json.Marshal(block)
	if err != nil {
		log.Printf("[P2P] Error serializing block for broadcast: %v", err)
		return
	}

	// Store the block in the DHT using its hash as key
	key := fmt.Sprintf("block:%s", block.Hash)
	enm.Store(key, blockData)

	// Also store a reference to this block in an index by height
	heightKey := fmt.Sprintf("height:%d", block.Index)
	enm.Store(heightKey, []byte(block.Hash))

	// Broadcast to a subset of nodes directly for faster propagation
	activeNodes := enm.GetActiveNodes(10) // Get 10 active nodes

	for _, node := range activeNodes {
		enm.sendBlockToNode(node.URL, block)
	}
}

// GetActiveNodes returns a list of active nodes
func (enm *EnhancedNetworkManager) GetActiveNodes(limit int) []*KademliaNode {
	enm.routingMutex.RLock()
	defer enm.routingMutex.RUnlock()

	var activeNodes []*KademliaNode

	for _, bucket := range enm.kBuckets {
		for _, node := range bucket {
			if node.Status == NodeStatusActive || node.Status == NodeStatusValidator {
				activeNodes = append(activeNodes, node)

				if len(activeNodes) >= limit {
					return activeNodes
				}
			}
		}
	}

	return activeNodes
}

// sendBlockToNode sends a block to a specific node
func (enm *EnhancedNetworkManager) sendBlockToNode(nodeURL string, block *blockchain.Block) {
	// Create a message for the block
	blockData, _ := json.Marshal(block)
	message := Message{
		Type:    MessageTypeBlock,
		Payload: blockData,
		Sender:  enm.NodeURL,
		Time:    time.Now(),
	}

	// Send the message
	messageData, _ := json.Marshal(message)
	resp, err := http.Post(nodeURL+"/p2p/message", "application/json", strings.NewReader(string(messageData)))

	if err != nil {
		log.Printf("[P2P] Error sending block to %s: %v", nodeURL, err)
		return
	}
	defer resp.Body.Close()
}
