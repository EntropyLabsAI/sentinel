package main

import (
	"container/list"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Upgrade HTTP connection to WebSocket
var upgrader = websocket.Upgrader{}

// ReviewRequest represents the review data structure
type ReviewRequest struct {
	ID             string `json:"id"`
	Context        string `json:"context"`
	ProposedAction string `json:"proposed_action"`
}

// ReviewerResponse represents the response from the reviewer
type ReviewerResponse struct {
	ID       string `json:"id"`
	Decision string `json:"decision"`
}

// Hub maintains active connections and broadcasts messages
type Hub struct {
	Clients              map[*Client]bool
	Review               chan ReviewRequest
	Register             chan *Client
	Unregister           chan *Client
	Queue                *list.List
	QueueMutex           sync.Mutex
	FreeClients          chan *Client // Channel to track available clients
	assignedReviews      map[string]*Client
	assignedReviewsMutex sync.Mutex
}

// NewHub initializes a new Hub
func NewHub() *Hub {
	return &Hub{
		Review:          make(chan ReviewRequest),
		Register:        make(chan *Client),
		Unregister:      make(chan *Client),
		Clients:         make(map[*Client]bool),
		Queue:           list.New(),
		FreeClients:     make(chan *Client, 100), // Buffered channel to hold available clients
		assignedReviews: make(map[string]*Client),
	}
}

// Run the hub to manage client connections and messages
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
			// Mark the client as available
			h.FreeClients <- client
			log.Println("Client registered and marked as available.")
			// Process the queue to assign any queued reviews to the new client
			h.processQueue()
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
				log.Println("Client unregistered.")

				// Re-queue any assigned reviews for this client
				h.requeueAssignedReviews(client)
			}
		case review := <-h.Review:
			h.assignReview(review)
		}
	}
}

// assignReview assigns a review to an available client or queues it
func (h *Hub) assignReview(review ReviewRequest) {
	for {
		select {
		case client := <-h.FreeClients:
			// Check if the client is still registered
			if _, exists := h.Clients[client]; !exists {
				log.Printf("Client was unregistered. Skipping review ID %s.", review.ID)
				continue // Skip to the next available client
			}

			// Attempt to send the review to the client
			select {
			case client.Send <- review:
				// Track the assigned review
				h.assignedReviewsMutex.Lock()
				h.assignedReviews[review.ID] = client
				h.assignedReviewsMutex.Unlock()

				log.Printf("Assigned review ID %s to a client.", review.ID)
				return
			default:
				// If client's send channel is full, enqueue the review and mark client as unavailable
				h.QueueMutex.Lock()
				h.Queue.PushBack(review)
				h.QueueMutex.Unlock()
				log.Printf("Client's send channel full. Queued review ID %s.", review.ID)
				return
			}
		default:
			// No available clients; enqueue the review
			h.QueueMutex.Lock()
			h.Queue.PushBack(review)
			h.QueueMutex.Unlock()
			log.Printf("No available clients. Queued review ID %s.", review.ID)
			return
		}
	}
}

// processQueue assigns queued reviews to available clients
func (h *Hub) processQueue() {
	h.QueueMutex.Lock()
	defer h.QueueMutex.Unlock()

	for h.Queue.Len() > 0 {
		select {
		case client := <-h.FreeClients:
			// Check if the client is still registered
			if _, exists := h.Clients[client]; !exists {
				log.Printf("Client was unregistered. Skipping queued reviews.")
				continue // Skip to the next available client
			}

			// Assign the next review in the queue to the client
			element := h.Queue.Front()
			review := element.Value.(ReviewRequest)
			select {
			case client.Send <- review:
				// Track the assigned review
				h.assignedReviewsMutex.Lock()
				h.assignedReviews[review.ID] = client
				h.assignedReviewsMutex.Unlock()

				h.Queue.Remove(element)
				log.Printf("Assigned queued review ID %s to a client.", review.ID)
			default:
				// If client's send channel is full, put the review back and mark client as unavailable
				h.Queue.MoveToBack(element)
				log.Printf("Client's send channel full. Keeping review ID %s in queue.", review.ID)
			}
		default:
			// No available clients
			break
		}
	}
}

// requeueAssignedReviews re-queues any reviews assigned to the disconnected client
func (h *Hub) requeueAssignedReviews(client *Client) {
	h.assignedReviewsMutex.Lock()
	defer h.assignedReviewsMutex.Unlock()

	for reviewID, assignedClient := range h.assignedReviews {
		if assignedClient == client {
			// Re-queue the review
			var review ReviewRequest
			// Assuming you have a way to retrieve the review details.
			// This might require storing the review details elsewhere.
			// For simplicity, we'll assume review details can be reconstructed or stored.
			// Here, we simply log the action.
			log.Printf("Re-queuing review ID %s as client disconnected.", reviewID)

			// In a real implementation, you'd retrieve the ReviewRequest details.
			// Here, we'll create a placeholder. Modify as needed.
			review = ReviewRequest{
				ID: reviewID,
				// Populate other fields as necessary
			}

			h.QueueMutex.Lock()
			h.Queue.PushBack(review)
			h.QueueMutex.Unlock()

			// Remove the mapping
			delete(h.assignedReviews, reviewID)
		}
	}
}

// Client represents a single WebSocket connection
type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan ReviewRequest
}

// ReadPump handles incoming messages from the WebSocket
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		// Handle responses from the frontend
		var response ReviewerResponse
		err = json.Unmarshal(message, &response)
		if err != nil {
			log.Println("Error unmarshaling reviewer response:", err)
			continue
		}

		// Retrieve the channel waiting for this response
		if ch, ok := reviewChannels.Load(response.ID); ok {
			// Send response to the waiting HTTP handler
			responseChan := ch.(chan ReviewerResponse)
			responseChan <- response
		} else {
			log.Printf("No pending review request found for ID: %s", response.ID)
		}

		// After handling the response, remove the assigned review and mark the client as available again
		c.Hub.assignedReviewsMutex.Lock()
		delete(c.Hub.assignedReviews, response.ID)
		c.Hub.assignedReviewsMutex.Unlock()

		c.Hub.FreeClients <- c
		log.Printf("Client marked as available after handling review ID %s.", response.ID)

		// Optional: send the response to an external API if needed
		// go sendResponseToAPI(response)
	}
}

// WritePump sends messages to the WebSocket client
func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
		// Ensure the client is unregistered if WritePump exits
		c.Hub.Unregister <- c
	}()

	for review := range c.Send {
		err := c.Conn.WriteJSON(review)
		if err != nil {
			log.Println("Error sending review to client:", err)
			break
		}
		log.Printf("Sent review ID %s to client.", review.ID)
		// Note: The client will be marked as available after sending the review and handling the response
	}
}

// serveWs handles WebSocket requests from the frontend
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	client := &Client{
		Hub:  hub,
		Conn: conn,
		Send: make(chan ReviewRequest, 1),
	}
	client.Hub.Register <- client

	// Start reading and writing pumps
	go client.WritePump()
	go client.ReadPump()
}
