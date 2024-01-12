package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Relaxed CheckOrigin - for development purposes only
	CheckOrigin: func(r *http.Request) bool {
		log.Printf("WebSocket connection request from origin: %s\n", r.Header.Get("Origin"))
		return true // Allow all origins
	},
}

func handleTerminalConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v\n", err)
		return
	}
	log.Println("WebSocket connection established")
	defer conn.Close()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading WebSocket message: %v\n", err)
			return
		}
		log.Printf("Received message: %s (type: %d)\n", string(p), messageType)

		// Execute the command
		cmdOutput, err := exec.Command("sh", "-c", string(p)).CombinedOutput()
		if err != nil {
			errMsg := fmt.Sprintf("Error executing command: %s\n", err)
			log.Printf("Command execution error: %s\n", errMsg)
			if err := conn.WriteMessage(websocket.TextMessage, []byte(errMsg)); err != nil {
				log.Printf("Error sending WebSocket message: %v\n", err)
				return
			}
			continue
		}

		log.Printf("Command executed successfully. Output: %s\n", string(cmdOutput))

		// Send the command output back to the client
		if err := conn.WriteMessage(websocket.TextMessage, cmdOutput); err != nil {
			log.Printf("Error sending WebSocket message: %v\n", err)
			return
		}
	}
}

func main() {
	http.HandleFunc("/terminal", handleTerminalConnection)

	port := "8081"
	log.Printf("Server is starting on port %s...\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Printf("Error starting server: %v\n", err)
	}
}
