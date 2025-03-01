package event_bus

import (
	"encoding/json"
	"fmt"
	"net/http"
	"src/config"
	"sync"

	"golang.org/x/net/http2"
)

// Message represents an EventSource message
type Message struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Global variables to manage clients and events
var (
	clients   = make(map[chan string]bool) // Active clients
	clientsMu sync.Mutex                   // Mutex to protect the map
)

// Handle a client connection
func handleClient(w http.ResponseWriter, r *http.Request) {
	// Set headers for EventSource
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "->> Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Create a channel for this client
	clientChan := make(chan string, 10) // Buffer to prevent blocking

	// Add client to the list
	clientsMu.Lock()
	clients[clientChan] = true
	clientsMu.Unlock()

	// Remove client on disconnect
	defer func() {
		clientsMu.Lock()
		delete(clients, clientChan)
		clientsMu.Unlock()
		close(clientChan)
		fmt.Println("->> Client disconnected")
	}()

	// Listen for messages
	for msg := range clientChan {
		fmt.Fprintf(w, "data: %s\n\n", msg)
		flusher.Flush()
	}
}

// SendMessage sends a message to all connected clients
func SendMessage(msg Message) {
	jsonData, _ := json.Marshal(msg)

	clientsMu.Lock()
	for clientChan := range clients {
		select {
		case clientChan <- string(jsonData):
			// Successfully sent message
		default:
			// Client is blocked, remove it
			delete(clients, clientChan)
			close(clientChan)
			fmt.Println("->> Removed blocked client")
		}
	}
	clientsMu.Unlock()
}

// Publish starts an HTTP server that streams messages to clients
func Publish() {
	if config.App.Verbose {
		fmt.Println("->> Opening the event bus server on http://localhost:8001/messages")
	}

	// HTTP server (no TLS)
	server := &http.Server{
		Addr:    ":8001",
		Handler: nil,
	}
	http2.ConfigureServer(server, nil) // Enable HTTP2

	// Handle SSE messages
	http.HandleFunc("/messages", handleClient)

	// Start the server
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("->> Server error:", err)
	}
	SendMessage(Message{Type: "local_file_changed", Message: "The watcher is warming up..."})
}
