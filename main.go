package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/lkhume/simple-text-editor/crdt"
)

//go:embed templates/*
var templatesFS embed.FS

type Operation struct {
	Type    string `json:"type"` // "insert"/"delete"
	Pos     int    `json:"pos"`
	Char    string `json:"char,omitempty"`
	Site    string `json:"site,omitempty"`
	Counter int    `json:"counter,omitempty"`
}

// Parse the index.html template.
var tmpl = template.Must(template.ParseFS(templatesFS, "templates/index.html"))

// Configure the WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin (for development).
		return true
	},
}

// global map tracking all active websocket connections
var clients = make(map[*websocket.Conn]bool)

// client map mutex
var clientsMutex sync.Mutex

// initialize CRDT document
var doc = crdt.NewDocument()
var localSite = "local"
var localCounter = 0

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
	if err := conn.WriteJSON(map[string]string{"text": doc.ToString()}); err != nil {
		log.Println("Error sending shared document:", err)
		return
	}

	// Listen for incoming messages.
	for {
		var op Operation
		if err := conn.ReadJSON(&op); err != nil {
			log.Println("Error reading operation:", err)
			break
		}

		switch op.Type {
		case "insert":
			// If the client didn't supply a site and counter, assign them.
			if op.Site == "" {
				op.Site = localSite
				localCounter++
				op.Counter = localCounter
			}
			if len(op.Char) != 1 {
				log.Println("Insert operation: 'char' should be a single character")
				continue
			}
			newElem := crdt.Element{
				ID:      crdt.Identifier{Site: op.Site, Counter: op.Counter},
				Char:    []rune(op.Char)[0],
				Deleted: false,
			}
			if err := doc.Insert(newElem, op.Pos); err != nil {
				log.Println("Insert error:", err)
				continue
			}
		case "delete":
			// For a delete operation, mark the element at the given position as deleted.
			if err := doc.Delete(op.Pos); err != nil {
				log.Println("Delete error:", err)
				continue
			}
		default:
			log.Println("Unknown operation type:", op.Type)
			continue
		}

		// Sort the document to ensure consistent ordering
		doc.Merge()

		broadcastUpdate(map[string]string{"text": doc.ToString()})
	}

	// Once the connection loop ends, remove the client from the clients map
	clientsMutex.Lock()
	delete(clients, conn)
	clientsMutex.Unlock()
}

// broadcastUpdate sends the updated text to all clients except the sender.
func broadcastUpdate(message interface{}) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for client := range clients {
		if err := client.WriteJSON(message); err != nil {
			log.Println("Error broadcasting to client:", err)
			client.Close()
			delete(clients, client)
		}
	}
}

func main() {
	dir, err := os.Getwd()

	if err != nil {
		log.Fatal("Error fetching working directory: ", err)
	}

	log.Println("Roxane working directory: ", dir)

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
