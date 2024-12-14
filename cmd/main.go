package main

import (
	"1/1/functions/config"
	"1/1/functions/setup"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: static [setup|add <page name>]")
		return
	}

	var staticjson config.StaticJson

	command := os.Args[1]
	switch command {
	case "setup":
		if len(os.Args) < 3 {
			fmt.Println("Please provide the project name: static setup <project name>")
			return
		}
		projectName := os.Args[2]
		setup.SetupProject(projectName, staticjson)
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a page name: static add <page name>")
			return
		}
		pageName := os.Args[2]
		setup.AddPage(pageName)
	case "compile":
		setup.CompileProject()
	case "watch":
		watchProject()
	default:
		fmt.Println("Unknown command. Use 'setup' or 'add <page name>'.")
	}
}

func watchProject() {
	// WebSocket upgrader from the Gorilla WebSocket package.
	var upgrader = websocket.Upgrader{
		// Allow connections from any origin; for production, adjust this.
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// Map to keep track of connected clients.
	var clients = make(map[*websocket.Conn]bool)
	var clientsMutex sync.Mutex

	// Handle WebSocket connections.
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// Upgrade the HTTP connection to a WebSocket connection.
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Failed to set WebSocket upgrade:", err)
			return
		}
		defer conn.Close()

		// Register the new client.
		clientsMutex.Lock()
		clients[conn] = true
		clientsMutex.Unlock()

		fmt.Println("New client connected:", conn.RemoteAddr())

		// Listen for messages from the client (this keeps the connection open).
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}

		// Unregister the client when done.
		clientsMutex.Lock()
		delete(clients, conn)
		clientsMutex.Unlock()
		fmt.Println("Client disconnected:", conn.RemoteAddr())
	})

	// Serve static files from the "output" directory.
	fileServer := http.FileServer(http.Dir("./output"))
	http.Handle("/", fileServer)

	// Function to notify all connected clients to reload.
	notifyClients := func() {
		clientsMutex.Lock()
		defer clientsMutex.Unlock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte("reload"))
			if err != nil {
				fmt.Println("Error sending message to", client.RemoteAddr(), ":", err)
				client.Close()
				delete(clients, client)
			}
		}
	}

	// Set up a file watcher using fsnotify.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Error creating watcher:", err)
		return
	}
	defer watcher.Close()

	// Function to add all subdirectories to the watcher.
	addDirsToWatcher := func(root string) {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if err := watcher.Add(path); err != nil {
					fmt.Println("Error adding", path, "to watcher:", err)
				}
			}
			return nil
		})
		if err != nil {
			fmt.Println("Error walking through", root, ":", err)
		}
	}

	// Monitor all subdirectories inside "./views" and "./public".
	addDirsToWatcher("./views")
	addDirsToWatcher("./public")

	// Start a goroutine to handle file system events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if filepath.Ext(event.Name) == ".go" {
						// fmt.Println("Detected change in a Go file, ignoring:", event.Name)
						continue
					}
					// fmt.Println("Detected change in:", event.Name)
					setup.CompileProject()
					notifyClients()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("Watcher error:", err)
			}
		}
	}()

	// Start the single HTTP server.
	fmt.Println("Starting server at http://localhost:8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}

	// The server will block here, so no need for a select{} or similar.
}
