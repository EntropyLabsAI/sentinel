package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

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
	Broadcast  chan ReviewRequest
	Register   chan *Client
	Unregister chan *Client
}

// ReviewerResponse represents the response from the reviewer
type ReviewerResponse struct {
	ID       string `json:"id"`
	Decision string `json:"decision"`
}

// NewHub initializes a new Hub
func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan ReviewRequest),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

// Run the hub to manage client connections and messages
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
		case review := <-h.Broadcast:
			// Broadcast review to all connected clients
			for client := range h.Clients {
				select {
				case client.Send <- review:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
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
	for {
		select {
		case review, ok := <-c.Send:
			if !ok {
				// Channel closed
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			// Send review request to client
			c.Conn.WriteJSON(review)
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

func sendResponseToAPI(response map[string]string) {
	// Prepare the request
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Println("Error marshaling response:", err)
		return
	}

	// Send POST request to external API
	resp, err := http.Post("https://external-api.example.com/response",
		"application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error sending response to external API:", err)
		return
	}
	defer resp.Body.Close()
	log.Println("Response sent to external API with status:", resp.Status)
}
