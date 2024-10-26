package sentinel

import (
	"container/list"
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

const MAX_REVIEWS_PER_CLIENT = 10

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
	// ReviewChan is a channel that receives new reviews from agents
	ReviewChan chan Review
	// Register is used when a new client connects
	Register chan *Client
	// Unregister is used when a client disconnects
	Unregister chan *Client
	// AssignedReviews is a map of clients to the reviews they are currently processing
	AssignedReviews      map[*Client]map[string]bool
	AssignedReviewsMutex sync.RWMutex
	// ReviewStore is used to store reviews that have been assigned to a client and are waiting to be completed
	// This is used to ensure that reviews are not lost if a client disconnects unexpectedly
	// ReviewStore *ReviewStoreType
	// Queue is a list of reviews that are waiting to be assigned to a client
	Queue *list.List
	// CompletedReviewCount is used to count the number of reviews that have been completed
	CompletedReviewCount int
	Store                Store
}

func NewHub(store Store) *Hub {
	return &Hub{
		Clients:         make(map[*Client]bool),
		ReviewChan:      make(chan Review, 100),
		Register:        make(chan *Client),
		Unregister:      make(chan *Client),
		AssignedReviews: make(map[*Client]map[string]bool),
		// ReviewStore:     NewReviewStore(),
		Queue: list.New(),
		Store: store,
	}
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
	h.processQueue()
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
func (h *Hub) assignReview(review Review) {
	h.ClientsMutex.RLock()
	defer h.ClientsMutex.RUnlock()

	// Attempt to assign the review to a client
	if !h.assignReviewToClient(review) {
		// If no client is available, queue the review
		h.Queue.PushBack(review)
		log.Printf("No available clients with capacity. Queued review.RequestId %s.", review.Id)
	}
}

// processQueue is a loop that continuously checks the queue for available reviews to assign to clients
func (h *Hub) processQueue() {
	h.ClientsMutex.RLock()
	defer h.ClientsMutex.RUnlock()

	var next *list.Element
	for e := h.Queue.Front(); e != nil; e = next {
		next = e.Next()
		review := e.Value.(Review)

		if h.assignReviewToClient(review) {
			h.Queue.Remove(e)
		} else {
			// No clients with capacity available at this time
			break
		}
	}
}

// assignReviewToClient attempts to assign a review to a client if they have capacity
func (h *Hub) assignReviewToClient(review Review) bool {
	ctx := context.Background()

	h.AssignedReviewsMutex.Lock()
	defer h.AssignedReviewsMutex.Unlock()

	// Iterate over all clients and assign the review if they have capacity
	for client := range h.Clients {
		assignedReviewsCount := len(h.AssignedReviews[client])

		if assignedReviewsCount < MAX_REVIEWS_PER_CLIENT {
			client.Send <- review

			err := h.Store.CreateReview(ctx, review)
			if err != nil {
				fmt.Printf("Error creating review in store: %v\n", err)
				continue
			}

			h.AssignedReviews[client][review.Id.String()] = true
			log.Printf("Assigned review.RequestId %s to client.", review.Id)
			return true // Review assignekd
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

			review, err := h.Store.GetReview(context.Background(), reviewID)
			if err != nil {
				fmt.Printf("Error getting review from store: %v\n", err)
				continue
			}

			if review == nil {
				fmt.Printf("Review details for ID %s not found in store. Skipping requeue.", reviewID)
				continue
			}

			fmt.Printf("Review details for ID %s have been retrieved from the store\n", reviewID)

			err = h.Store.DeleteReview(context.Background(), reviewID)
			if err != nil {
				fmt.Printf("Error deleting review from store: %v\n", err)
				continue
			}

			fmt.Printf("Review details for ID %s have been deleted from the store\n", reviewID)
			h.ReviewChan <- *review
			fmt.Printf("Review details for ID %s have been sent to the ReviewChan\n", reviewID)

			log.Printf("Re-queuing review.RequestId %s as client disconnected.", reviewID)
		}
	}
}

// Client represents a single WebSocket connection
type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan Review
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

		// Remove the review from the ReviewStore
		err = c.Hub.Store.DeleteReview(context.Background(), response.Id)
		if err != nil {
			fmt.Printf("Error deleting review from store: %v\n", err)
		}

		c.Hub.CompletedReviewCount++

		// Process the next review in the queue
		c.Hub.processQueue()
	}
}

type HubStats struct {
	ConnectedClients   int
	QueuedReviews      int
	StoredReviews      int
	AssignedReviews    map[string]int
	BusyClients        int
	FreeClients        int
	CompletedReviews   int
	ReviewDistribution map[string]int
}

func (h *Hub) getStats() HubStats {
	ctx := context.Background()

	count, err := h.Store.CountReviews(ctx)
	if err != nil {
		fmt.Printf("Error counting reviews in store: %v\n", err)
		count = 0
	}

	stats := HubStats{
		ConnectedClients:   len(h.Clients),
		QueuedReviews:      h.Queue.Len(),
		StoredReviews:      count,
		AssignedReviews:    make(map[string]int),
		ReviewDistribution: make(map[string]int),
		CompletedReviews:   h.CompletedReviewCount,
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

	return stats
}
