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
	Clients         map[*Client]bool
	ReviewChan      chan ReviewRequest
	Register        chan *Client
	Unregister      chan *Client
	FreeClients     chan *Client
	AssignedReviews map[string]*Client
	ReviewStore     *ReviewStore
	Queue           *list.List
}

func NewHub() *Hub {
	return &Hub{
		Clients:         make(map[*Client]bool),
		ReviewChan:      make(chan ReviewRequest),
		Register:        make(chan *Client),
		Unregister:      make(chan *Client),
		FreeClients:     make(chan *Client, 100),
		AssignedReviews: make(map[string]*Client),
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
			fmt.Printf("Received review from ReviewChan: %v\n", review.ID)
			h.assignReview(review)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.Clients[client] = true
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
	}
}

func (h *Hub) assignReview(review ReviewRequest) {
	select {
	case client := <-h.FreeClients:
		if _, exists := h.Clients[client]; !exists {
			log.Printf("Client was unregistered. Skipping review ID %s.", review.ID)
			h.assignReview(review) // Retry assignment
			return
		}

		select {
		case client.Send <- review:
			h.AssignedReviews[review.ID] = client
			h.ReviewStore.Add(review)
			log.Printf("Assigned review ID %s to a client.", review.ID)
		default:
			h.Queue.PushBack(review)
			log.Printf("Client's send channel full. Queued review ID %s.", review.ID)
			h.FreeClients <- client // Put the client back as it's still available
		}
	default:
		h.Queue.PushBack(review)
		log.Printf("No available clients. Queued review ID %s.", review.ID)
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

			select {
			case client.Send <- review:
				h.AssignedReviews[review.ID] = client
				h.ReviewStore.Add(review)
				h.Queue.Remove(element)
				log.Printf("Assigned queued review ID %s to a client.", review.ID)
			default:
				log.Printf("Client's send channel full. Keeping review ID %s in queue.", review.ID)
				h.FreeClients <- client // Put the client back as it's still available
				return
			}
		default:
			return
		}
	}
}

func (h *Hub) requeueAssignedReviews(client *Client) {
	for reviewID, assignedClient := range h.AssignedReviews {
		if assignedClient == client {
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
			delete(h.AssignedReviews, reviewID)
			fmt.Printf("Review details for ID %s have been deleted from the AssignedReviews map\n", reviewID)

			log.Printf("Re-queuing review ID %s as client disconnected.", reviewID)
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

		// Handle the response (e.g., send to HTTP handler or process internally)
		// Implement reviewChannels or alternative handling as needed

		// Mark the client as available after processing
		if _, exists := c.Hub.AssignedReviews[response.ID]; exists {
			delete(c.Hub.AssignedReviews, response.ID)
			c.Hub.ReviewStore.Delete(response.ID)
			c.Hub.FreeClients <- c
			log.Printf("Client marked as available after handling review ID %s.", response.ID)
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
		log.Printf("Sent review ID %s to client.", review.ID)
	}
}

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
	hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
