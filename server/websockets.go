package sentinel

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const MAX_REVIEWS_PER_CLIENT = 1

// Upgrade HTTP connection to WebSocket with proper settings
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Adjust as needed for security
}

// Hub maintains active connections and broadcasts messages
type Hub struct {
	// Clients is a map of clients to their connection status
	Clients      map[*Client]bool
	ClientsMutex sync.RWMutex
	// ReviewChan is a channel that receives new reviews, then assigns them to a connected client
	ReviewChan chan ReviewRequest
	// Register and Unregister are used when a new client connects and disconnects
	Register   chan *Client
	Unregister chan *Client
	// AssignedReviews is a map of clients to the reviews they are currently processing
	AssignedReviews      map[*Client]map[string]bool
	AssignedReviewsMutex sync.RWMutex

	// CompletedReviewCount is used to count the number of reviews that have been completed
	CompletedReviewCount int
	Store                Store
}

func NewHub(store Store, humanReviewChan chan ReviewRequest) *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		ReviewChan: humanReviewChan,
		Register:   make(chan *Client),
		Unregister: make(chan *Client),

		AssignedReviews: make(map[*Client]map[string]bool),

		Store: store,
	}
}

// serveWs upgrades the HTTP connection to a WebSocket connection and registers the client with the hub
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	client := &Client{
		Hub:  hub,
		Conn: conn,
		Send: make(chan ReviewRequest),
	}
	hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}

// Run starts the hub and handles client connections/disconnections and review assignments
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)
		case client := <-h.Unregister:
			h.unregisterClient(client)
		case review := <-h.ReviewChan:
			fmt.Printf("Received review from ReviewChan: %v\n", review.Id)
			h.assignReview(review)
		}
	}
}

// registerClient adds a new client to the hub and initializes their assigned reviews map
func (h *Hub) registerClient(client *Client) {
	h.ClientsMutex.Lock()
	h.AssignedReviewsMutex.Lock()
	h.Clients[client] = true
	h.AssignedReviews[client] = make(map[string]bool)
	h.AssignedReviewsMutex.Unlock()
	h.ClientsMutex.Unlock()

	log.Println("Client registered.")
}

// unregisterClient removes a client from the hub and handles the cleanup of their assigned reviews
func (h *Hub) unregisterClient(client *Client) {
	h.ClientsMutex.Lock()
	if _, exists := h.Clients[client]; exists {
		delete(h.Clients, client)
		h.ClientsMutex.Unlock()

		// Remove the client from the AssignedReviews map and requeue their reviews
		h.AssignedReviewsMutex.Lock()
		h.requeueAssignedReviews(client)
		delete(h.AssignedReviews, client)
		h.AssignedReviewsMutex.Unlock()

		close(client.Send)
		log.Println("Client unregistered.")
	} else {
		h.ClientsMutex.Unlock()
	}
}

// assignReview assigns a review to a client if they have capacity, otherwise it queues the review
func (h *Hub) assignReview(review ReviewRequest) {
	if review.Id == nil {
		log.Fatalf("can't assign review with nil ID")
	}

	h.ClientsMutex.RLock()
	defer h.ClientsMutex.RUnlock()

	// Attempt to assign the review to a client
	if !h.assignReviewToClient(review) {
		// If no client is available, do nothing.
		log.Printf("No available clients with capacity. Queued review.RequestId %s.", review.Id)
	}
}

// assignReviewToClient attempts to assign a review to a client if they have capacity
func (h *Hub) assignReviewToClient(review ReviewRequest) bool {
	h.AssignedReviewsMutex.Lock()
	defer h.AssignedReviewsMutex.Unlock()

	// Iterate over all clients and assign the review if they have capacity
	for client := range h.Clients {
		assignedReviewsCount := len(h.AssignedReviews[client])

		if assignedReviewsCount < MAX_REVIEWS_PER_CLIENT {
			client.Send <- review

			h.AssignedReviews[client][review.Id.String()] = true
			log.Printf("Assigned review.RequestId %s to client.", review.Id)

			// Update the review status to assigned
			err := h.Store.CreateReviewStatus(context.Background(), *review.Id, ReviewStatus{
				Status: Assigned,
			})
			if err != nil {
				fmt.Printf("Error creating review status: %v\n", err)
			}

			return true // Review assigned
		}
	}

	return false // No client available
}

// requeueAssignedReviews removes all reviews from a client and requeues them
func (h *Hub) requeueAssignedReviews(client *Client) {
	if assignedReviews, ok := h.AssignedReviews[client]; ok {
		for reviewID := range assignedReviews {
			reviewID, err := uuid.Parse(reviewID)
			if err != nil {
				fmt.Printf("Error parsing review ID: %v\n", err)
				continue
			}

			err = h.Store.CreateReviewStatus(context.Background(), reviewID, ReviewStatus{
				Status: Pending,
			})
			if err != nil {
				fmt.Printf("Error getting review from store: %v\n", err)
				continue
			}
		}

		// Remove the client from the AssignedReviews map
		delete(h.AssignedReviews, client)
	}
}

// Client represents a single WebSocket connection
type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan ReviewRequest
}

// WritePump handles the sending of reviews to the client
func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
		c.Hub.Unregister <- c
	}()

	for review := range c.Send {
		if err := c.Conn.WriteJSON(review); err != nil {
			log.Println("Error sending review to client:", err)
			break
		}
		log.Printf("Sent review.RequestId %s to client.", review.Id)
	}
}

// ReadPump handles the reading of messages from the client and handles the processing of responses
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

		var response ReviewResult
		if err := json.Unmarshal(message, &response); err != nil {
			log.Println("Error unmarshaling reviewer response:", err)
			continue
		}

		// Handle the response
		if chInterface, ok := reviewChannels.Load(response.Id); ok {
			responseChan, ok := chInterface.(chan ReviewResult)
			if ok {
				// Send the response non-blocking to prevent potential deadlocks
				select {
				case responseChan <- response:
					log.Printf("ReviewerResponse for ID %s sent to response channel.", response.Id)
				default:
					log.Printf("Response channel for ID %s is blocked. Skipping.", response.Id)
				}
			} else {
				log.Printf("Response channel for ID %s has an unexpected type.", response.Id)
			}
		} else {
			log.Printf("No response channel found for ID %s.", response.Id)
		}

		// Thread-safe removal of the review ID from assigned reviews
		c.Hub.AssignedReviewsMutex.Lock()
		if _, exists := c.Hub.AssignedReviews[c]; exists {
			delete(c.Hub.AssignedReviews[c], response.Id.String())
		}
		c.Hub.AssignedReviewsMutex.Unlock()
	}
}

func (h *Hub) getStats() (HubStats, error) {
	ctx := context.Background()

	pendingCount, err := h.Store.CountReviewRequests(ctx, Pending)
	if err != nil {
		return HubStats{}, fmt.Errorf("error counting pending reviews: %w", err)
	}

	completedCount, err := h.Store.CountReviewRequests(ctx, Completed)
	if err != nil {
		return HubStats{}, fmt.Errorf("error counting completed reviews: %w", err)
	}

	assignedCount, err := h.Store.CountReviewRequests(ctx, Assigned)
	if err != nil {
		return HubStats{}, fmt.Errorf("error counting assigned reviews: %w", err)
	}

	stats := HubStats{
		ConnectedClients:   len(h.Clients),
		ReviewDistribution: make(map[string]int),
		AssignedReviews:    make(map[string]int),
		FreeClients:        0,
		BusyClients:        0,

		PendingReviewsCount:   pendingCount,
		CompletedReviewsCount: completedCount,
		AssignedReviewsCount:  assignedCount,
	}

	totalAssignedReviews := 0

	h.AssignedReviewsMutex.RLock()
	for client, reviews := range h.AssignedReviews {
		clientKey := fmt.Sprintf("%p", client)
		assignedCount := len(reviews)

		// Annoyingly the ReviewDistribution map is a map[string]int so we need to
		// convert the assignedCount to a string
		assignedCountStr := strconv.Itoa(assignedCount)

		stats.AssignedReviews[clientKey] = assignedCount
		stats.ReviewDistribution[assignedCountStr]++
		totalAssignedReviews += assignedCount
	}
	h.AssignedReviewsMutex.RUnlock()

	stats.BusyClients = 0
	stats.FreeClients = 0

	for assignedCountStr, count := range stats.ReviewDistribution {
		assignedCount, err := strconv.Atoi(assignedCountStr)
		if err != nil {
			log.Printf("Error converting assignedCountStr to int: %v", err)
			continue
		}
		if assignedCount < MAX_REVIEWS_PER_CLIENT {
			stats.FreeClients += count
		} else {
			stats.BusyClients += count
		}
	}

	return stats, nil
}
