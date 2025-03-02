package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Global shared text and mutex for safe concurrent access.
var (
	sharedText   string = ""
	textMutex    sync.RWMutex
	clients      = make(map[*websocket.Conn]bool)
	clientsMutex sync.Mutex
)

//go:embed templates/*
var templatesFS embed.FS

// Parse the index.html template.
var tmpl = template.Must(template.ParseFS(templatesFS, "templates/index.html"))

// Configure the WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin (for development).
		return true
	},
}

// serveIndex renders the index.html template.
func serveIndex(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Println("Template execution error:", err)
	}
}

// wsHandler upgrades HTTP connections to WebSocket and handles messaging.
func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	// Register the new client.
	clientsMutex.Lock()
	clients[conn] = true
	clientsMutex.Unlock()

	// Send the current shared text to the new client.
	textMutex.RLock()
	if err := conn.WriteJSON(map[string]string{"text": sharedText}); err != nil {
		log.Println("Error sending shared text:", err)
		textMutex.RUnlock()
		return
	}
	textMutex.RUnlock()

	// Listen for incoming messages.
	for {
		var msg map[string]string
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("Error reading JSON:", err)
			break
		}

		if newText, ok := msg["text"]; ok {
			// update the shared text.
			textMutex.Lock()
			sharedText = newText
			textMutex.Unlock()

			// Broadcast the update to all other clients
			broadcastUpdate(conn, newText)
		}
	}

	// Remove the client when done (e.g., on disconnect).
	clientsMutex.Lock()
	delete(clients, conn)
	clientsMutex.Unlock()
}

// broadcastUpdate sends the updated text to all clients except the sender.
func broadcastUpdate(sender *websocket.Conn, text string) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	message := map[string]string{"text": text}
	for client := range clients {
		if client == sender {
			continue
		}
		if err := client.WriteJSON(message); err != nil {
			log.Println("Error broadcasting to client:", err)
			client.Close()
			delete(clients, client)
		}
	}
}

func main() {
	// Route for the main page.
	http.HandleFunc("/", serveIndex)
	// WebSocket endpoint.
	http.HandleFunc("/ws", wsHandler)
	// serve static files.
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
