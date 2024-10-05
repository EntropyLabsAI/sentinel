package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrade HTTP connection to WebSocket with proper settings
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Adjust as needed for security
}

// Hub maintains active connections and broadcasts messages
type Hub struct {
	Clients         map[*Client]bool
	ReviewChan      chan ReviewRequest
	Register        chan *Client
	Unregister      chan *Client
	FreeClients     chan *Client
	AssignedReviews map[*Client]map[string]bool // Map of clients to their assigned review IDs
	ReviewStore     *ReviewStore
	Queue           *list.List
}

func NewHub() *Hub {
	return &Hub{
		Clients:         make(map[*Client]bool),
		ReviewChan:      make(chan ReviewRequest, 100),
		Register:        make(chan *Client),
		Unregister:      make(chan *Client),
		FreeClients:     make(chan *Client, 100),
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
	h.Clients[client] = true
	h.AssignedReviews[client] = make(map[string]bool)
	h.FreeClients <- client
	log.Println("Client registered and marked as available.")
	h.processQueue()
}

func (h *Hub) unregisterClient(client *Client) {
	if _, exists := h.Clients[client]; exists {
		delete(h.Clients, client)
		close(client.Send)
		log.Println("Client unregistered.")
		h.requeueAssignedReviews(client)
		delete(h.AssignedReviews, client)
	}
}

func (h *Hub) assignReview(review ReviewRequest) {
	for {
		select {
		case client := <-h.FreeClients:
			if _, exists := h.Clients[client]; !exists {
				log.Printf("Client was unregistered. Skipping review.RequestID %s.", review.RequestID)
				continue // Try next client
			}
			if len(client.Send) < cap(client.Send) {
				// Assign the review to the client
				client.Send <- review
				h.ReviewStore.Add(review)
				h.AssignedReviews[client][review.RequestID] = true
				log.Printf("Assigned review.RequestID %s to a client.", review.RequestID)

				// Return client to FreeClients if they still have capacity
				if len(client.Send) < cap(client.Send) {
					h.FreeClients <- client
				}
			} else {
				// Client's send channel is full, try next client
				// Return client back to FreeClients as it may be able to accept more reviews later
				h.FreeClients <- client
				continue
			}
			return // Review assigned
		default:
			h.Queue.PushBack(review)
			log.Printf("No available clients. Queued review.RequestID %s.", review.RequestID)
			return
		}
	}
}

func (h *Hub) processQueue() {
	for h.Queue.Len() > 0 {
		select {
		case client := <-h.FreeClients:
			if _, exists := h.Clients[client]; !exists {
				log.Println("Client was unregistered while processing queue.")
				continue
			}
			element := h.Queue.Front()
			review := element.Value.(ReviewRequest)

			if len(client.Send) < cap(client.Send) {
				client.Send <- review
				h.ReviewStore.Add(review)
				h.AssignedReviews[client][review.RequestID] = true
				h.Queue.Remove(element)
				log.Printf("Assigned queued review.RequestID %s to client.", review.RequestID)

				// Return client to FreeClients if they still have capacity
				if len(client.Send) < cap(client.Send) {
					h.FreeClients <- client
				}
			} else {
				// Client's send channel is full, try next client
				// Return client back to FreeClients as it may be able to accept more reviews later
				h.FreeClients <- client
				continue
			}
		default:
			// No free clients available
			return
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

		// Handle the response by sending it to the corresponding response channel
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

		// Remove the review ID from the client's assigned reviews
		if _, exists := c.Hub.AssignedReviews[c]; exists {
			delete(c.Hub.AssignedReviews[c], response.ID)
		}

		// Remove the review from the ReviewStore
		c.Hub.ReviewStore.Delete(response.ID)

		// If client has capacity, add back to FreeClients
		if len(c.Send) < cap(c.Send) {
			c.Hub.FreeClients <- c
			log.Printf("Client marked as available after handling review.RequestID %s.", response.ID)
		}
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
