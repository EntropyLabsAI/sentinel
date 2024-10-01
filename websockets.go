package main

import (
	"container/list"
	"encoding/json"
	"log"
	"math/rand"
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

// Hub maintains active connections and broadcasts messages
type Hub struct {
	Clients    map[*Client]bool
	Review     chan ReviewRequest
	Register   chan *Client
	Unregister chan *Client
	Queue      *list.List
	QueueMutex sync.Mutex
}

// ReviewerResponse represents the response from the reviewer
type ReviewerResponse struct {
	ID       string `json:"id"`
	Decision string `json:"decision"`
}

// NewHub initializes a new Hub
func NewHub() *Hub {
	return &Hub{
		Review:     make(chan ReviewRequest),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Queue:      list.New(),
	}
}

// Run the hub to manage client connections and messages
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
			go h.processQueue() // Process queue when a new client connects
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
		case review := <-h.Review:
			h.addReview(review)
		}
	}
}

func (h *Hub) sendToOneClient(review ReviewRequest) bool {
	if len(h.Clients) == 0 {
		return false
	}

	// Select a random client
	clients := make([]*Client, 0, len(h.Clients))
	for client := range h.Clients {
		clients = append(clients, client)
	}
	randomClient := clients[rand.Intn(len(clients))]

	// Try to send the review to the selected client
	select {
	case randomClient.Send <- review:
		return true
	default:
		// Client's channel is full, remove it
		close(randomClient.Send)
		delete(h.Clients, randomClient)
		return false
	}
}

func (h *Hub) processQueue() {
	h.QueueMutex.Lock()
	defer h.QueueMutex.Unlock()

	for h.Queue.Len() > 0 {
		element := h.Queue.Front()
		review := element.Value.(ReviewRequest)
		if h.sendToOneClient(review) {
			h.Queue.Remove(element)
		} else {
			// If we couldn't send to any client, stop processing
			break
		}
	}
}

func (h *Hub) addReview(review ReviewRequest) {
	if !h.sendToOneClient(review) {
		h.QueueMutex.Lock()
		h.Queue.PushBack(review)
		h.QueueMutex.Unlock()
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

		// Optional: send the response to an external API if needed
		// go sendResponseToAPI(response)
	}
}

// WritePump sends messages to the WebSocket client
func (c *Client) WritePump() {
	defer c.Conn.Close()

	for review := range c.Send {
		err := c.Conn.WriteJSON(review)
		if err != nil {
			log.Println("Error sending review to client:", err)
			return
		}
	}
}

// serveWs handles WebSocket requests from the frontend
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	client := &Client{Hub: hub, Conn: conn, Send: make(chan ReviewRequest)}
	client.Hub.Register <- client

	// Start reading and writing pumps
	go client.WritePump()
	go client.ReadPump()
}
