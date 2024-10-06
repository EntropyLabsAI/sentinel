package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

const MAX_REVIEWS_PER_CLIENT = 3

// Upgrade HTTP connection to WebSocket with proper settings
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Adjust as needed for security
}

// Hub maintains active connections and broadcasts messages
type Hub struct {
	Clients              map[*Client]bool
	ClientsMutex         sync.RWMutex
	ReviewChan           chan ReviewRequest
	Register             chan *Client
	Unregister           chan *Client
	AssignedReviews      map[*Client]map[string]bool
	AssignedReviewsMutex sync.RWMutex
	ReviewStore          *ReviewStore
	Queue                *list.List
}

func NewHub() *Hub {
	return &Hub{
		Clients:         make(map[*Client]bool),
		ReviewChan:      make(chan ReviewRequest, 100),
		Register:        make(chan *Client),
		Unregister:      make(chan *Client),
		AssignedReviews: make(map[*Client]map[string]bool),
		ReviewStore:     NewReviewStore(),
		Queue:           list.New(),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)
		case client := <-h.Unregister:
			h.unregisterClient(client)
		case review := <-h.ReviewChan:
			fmt.Printf("Received review from ReviewChan: %v\n", review.RequestID)
			h.assignReview(review)
		}
	}
}

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

func (h *Hub) unregisterClient(client *Client) {
	h.ClientsMutex.Lock()
	if _, exists := h.Clients[client]; exists {
		delete(h.Clients, client)
		h.ClientsMutex.Unlock()

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

func (h *Hub) assignReview(review ReviewRequest) {
	h.ClientsMutex.RLock()
	defer h.ClientsMutex.RUnlock()
	h.AssignedReviewsMutex.Lock()
	defer h.AssignedReviewsMutex.Unlock()

	for client := range h.Clients {
		assignedReviewsCount := len(h.AssignedReviews[client])
		if assignedReviewsCount < MAX_REVIEWS_PER_CLIENT {
			// Assign the review to the client
			client.Send <- review
			h.ReviewStore.Add(review)
			h.AssignedReviews[client][review.RequestID] = true
			log.Printf("Assigned review.RequestID %s to a client.", review.RequestID)
			return // Review assigned
		}
	}

	// If no client is available, queue the review
	h.Queue.PushBack(review)
	log.Printf("No available clients with capacity. Queued review.RequestID %s.", review.RequestID)
}

func (h *Hub) processQueue() {
	h.ClientsMutex.RLock()
	defer h.ClientsMutex.RUnlock()
	h.AssignedReviewsMutex.Lock()
	defer h.AssignedReviewsMutex.Unlock()

	var next *list.Element
	for e := h.Queue.Front(); e != nil; e = next {
		next = e.Next()
		review := e.Value.(ReviewRequest)
		assigned := false

		for client := range h.Clients {
			assignedReviewsCount := len(h.AssignedReviews[client])
			if assignedReviewsCount < MAX_REVIEWS_PER_CLIENT {
				// Assign the review to the client
				client.Send <- review
				h.ReviewStore.Add(review)
				h.AssignedReviews[client][review.RequestID] = true
				log.Printf("Assigned queued review.RequestID %s to client.", review.RequestID)
				h.Queue.Remove(e)
				assigned = true
				break
			}
		}

		if !assigned {
			// No clients with capacity available at this time
			break
		}
	}
}

func (h *Hub) requeueAssignedReviews(client *Client) {
	if assignedReviews, ok := h.AssignedReviews[client]; ok {
		for reviewID := range assignedReviews {
			review, exists := h.ReviewStore.Get(reviewID)
			fmt.Printf("Review details for ID %s: %v\n", reviewID, review)
			if !exists {
				fmt.Printf("Review details for ID %s not found in ReviewStore. Skipping requeue.", reviewID)
				continue
			}

			fmt.Printf("Review details for ID %s have been retrieved from the store\n", reviewID)
			h.ReviewStore.Delete(reviewID)
			fmt.Printf("Review details for ID %s have been deleted from the store\n", reviewID)
			h.ReviewChan <- review
			fmt.Printf("Review details for ID %s have been sent to the ReviewChan\n", reviewID)

			log.Printf("Re-queuing review.RequestID %s as client disconnected.", reviewID)
		}
	}
}

// Client represents a single WebSocket connection
type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan ReviewRequest
}

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

		var response ReviewerResponse
		if err := json.Unmarshal(message, &response); err != nil {
			log.Println("Error unmarshaling reviewer response:", err)
			continue
		}

		// Handle the response
		if chInterface, ok := reviewChannels.Load(response.ID); ok {
			responseChan, ok := chInterface.(chan ReviewerResponse)
			if ok {
				// Send the response non-blocking to prevent potential deadlocks
				select {
				case responseChan <- response:
					log.Printf("ReviewerResponse for ID %s sent to response channel.", response.ID)
				default:
					log.Printf("Response channel for ID %s is blocked. Skipping.", response.ID)
				}
			} else {
				log.Printf("Response channel for ID %s has an unexpected type.", response.ID)
			}
		} else {
			log.Printf("No response channel found for ID %s.", response.ID)
		}

		// Thread-safe removal of the review ID from assigned reviews
		c.Hub.AssignedReviewsMutex.Lock()
		if _, exists := c.Hub.AssignedReviews[c]; exists {
			delete(c.Hub.AssignedReviews[c], response.ID)
		}
		c.Hub.AssignedReviewsMutex.Unlock()

		// Remove the review from the ReviewStore
		c.Hub.ReviewStore.Delete(response.ID)
	}
}

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
		log.Printf("Sent review.RequestID %s to client.", review.RequestID)
	}
}

type HubStats struct {
	ConnectedClients   int            `json:"connected_clients"`
	QueuedReviews      int            `json:"queued_reviews"`
	StoredReviews      int            `json:"stored_reviews"`
	FreeClients        int            `json:"free_clients"`
	BusyClients        int            `json:"busy_clients"`
	AssignedReviews    map[string]int `json:"assigned_reviews"`
	ReviewDistribution map[int]int    `json:"review_distribution"`
}

func (h *Hub) getStats() HubStats {
	stats := HubStats{
		ConnectedClients:   len(h.Clients),
		QueuedReviews:      h.Queue.Len(),
		StoredReviews:      h.ReviewStore.Count(),
		AssignedReviews:    make(map[string]int),
		ReviewDistribution: make(map[int]int),
	}

	totalAssignedReviews := 0

	h.AssignedReviewsMutex.RLock()
	for client, reviews := range h.AssignedReviews {
		clientKey := fmt.Sprintf("%p", client)
		assignedCount := len(reviews)
		stats.AssignedReviews[clientKey] = assignedCount
		stats.ReviewDistribution[assignedCount]++
		totalAssignedReviews += assignedCount
	}
	h.AssignedReviewsMutex.RUnlock()

	stats.BusyClients = 0
	stats.FreeClients = 0

	h.ClientsMutex.RLock()
	h.ClientsMutex.RUnlock()

	for _, assignedCount := range stats.ReviewDistribution {
		if assignedCount < MAX_REVIEWS_PER_CLIENT {
			stats.FreeClients++
		} else {
			stats.BusyClients++
		}
	}

	return stats
}
