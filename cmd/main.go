package main

import (
	"bufio"    // For reading user input
	"fmt"      // For formatted I/O operations
	"net/http" // For HTTP server functionality
	"os"       // For file and environment operations
	"os/exec"
	"path/filepath" // For handling file paths
	"strings"       // For string manipulations
	"sync"          // For synchronizing concurrent operations
	"time"

	"github.com/fsnotify/fsnotify" // For file system notifications
	"github.com/gorilla/websocket" // For WebSocket handling

	"1/1/functions/config" // Custom configuration handling
	"1/1/functions/setup"  // Custom project setup utilities
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
)

// handling the add command
func handleAddCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Please provide a page name: static add <page name>")
		return
	}
	pageName := os.Args[2]

	// Validate page name
	if strings.ContainsAny(pageName, " -./\\<>:\"|?*") || strings.HasSuffix(pageName, ".exe") {
		fmt.Println(Red + "Error: Invalid page name. Use only alphanumeric characters and avoid special symbols or '.exe' extensions." + Reset)
		return
	}

	fmt.Printf(Yellow+"You are about to add a page named '%s'. Confirm (y/n): "+Reset, pageName)
	reader := bufio.NewReader(os.Stdin)
	confirmation, _ := reader.ReadString('\n')
	confirmation = strings.TrimSpace(strings.ToLower(confirmation))

	// Validate confirmation input
	if confirmation != "y" && confirmation != "n" {
		fmt.Println(Red + "Invalid input. Please type 'y' for yes or 'n' for no." + Reset)
		return
	}

	if confirmation == "n" {
		fmt.Println(Yellow + "Page addition canceled." + Reset)
		return
	}

	// Add the page
	setup.AddPage(pageName)
	fmt.Println(Green + "Page created successfully. Running compile command..." + Reset)

	// Automatically run the compile command
	compileCmd := exec.Command(".\\static.exe", "compile")
	compileCmd.Stdout = os.Stdout
	compileCmd.Stderr = os.Stderr

	err := compileCmd.Run()
	if err != nil {
		fmt.Println(Red+"Error running compile command:"+Reset, err)
	} else {
		fmt.Println(Green + "Compile command executed successfully!" + Reset)
	}
}

// loading bar component
func greenRectangularLoadingBar(duration time.Duration) {
	green := "\033[32m" // ANSI code for green
	reset := "\033[0m"  // ANSI code to reset color

	totalLength := 30 // Length of the progress bar (number of blocks)

	// Print the start of the progress bar with the message
	fmt.Print("Setting up the project: [")

	// Display empty part of the progress bar (filled with spaces for now)
	for i := 0; i < totalLength; i++ {
		fmt.Print(" ")
	}
	fmt.Print("]")

	// Simulate the progress and update the progress bar
	for i := 0; i <= totalLength; i++ {
		fmt.Print("\rSetting up the project: [")
		// Print the filled portion (green blocks)
		for j := 0; j < i; j++ {
			fmt.Print(green + "█" + reset)
		}
		// Print the remaining empty portion
		for j := i; j < totalLength; j++ {
			fmt.Print(" ")
		}
		fmt.Print("]")

		// Adjust the speed of the progress bar filling
		time.Sleep(duration / time.Duration(totalLength))
	}

	// After the bar is filled, print a completion message
	fmt.Println("\nProject setup complete!")
}

// main function
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: static [setup|add <page name>]")
		return
	}

	var staticjson config.StaticJson

	command := os.Args[1]
	switch command {
	case "setup":
		reader := bufio.NewReader(os.Stdin)

		// Prompt for project name
		fmt.Print("Project name: ")
		projectName, _ := reader.ReadString('\n')
		projectName = strings.TrimSpace(projectName)

		// Validate project name
		if strings.Contains(projectName, " ") {
			fmt.Println("Error: Project name cannot contain spaces.")
			return
		}
		if len(projectName) > 100 {
			fmt.Println("Error: Project name cannot exceed 100 characters.")
			return
		}

		// Prompt for description
		fmt.Print(Yellow + "Description: " + Reset)
		description, _ := reader.ReadString('\n')
		description = strings.TrimSpace(description)

		// Prompt for author name
		fmt.Print(Yellow + "Author: " + Reset)
		author, _ := reader.ReadString('\n')
		author = strings.TrimSpace(author)

		// Simulate loading bar
		greenRectangularLoadingBar(2 * time.Second)

		// Output details
		fmt.Printf(Green + "\nProject setup complete:\n" + Reset)
		fmt.Printf(Green+"Project name: %s\nDescription: %s\nAuthor: %s\n"+Reset, projectName, description, author)

		// Call the setup function
		setup.SetupProject(projectName, staticjson)

	case "add":
		handleAddCommand()

	case "compile":
		setup.CompileProject()

	case "watch":
		watchProject()

	case "help":
		showHelp()

	default:
		fmt.Println("Unknown command. Use 'setup' or 'add <page name>'.")
	}
}

// watch command to host the project
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

// Show .help text
func showHelp() {
	fmt.Println("Usage: static [command] [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  setup          Configure your project interactively. It will guide you through the setup process.")
	fmt.Println("  add <page name> Add one or more pages to the project in a loop.")
	fmt.Println("  compile        Compile the entire project for production.")
	fmt.Println("  watch          Start a development server and watch for file changes.")
	fmt.Println("  help           Show this help message and list all available commands.")
}
