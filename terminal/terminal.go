package main

import (
	"fmt"
	"io/ioutil"
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

func executeCommandHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v\n", err)
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	command := string(body)
	log.Printf("Received command to execute: %s\n", command)

	cmdOutput, err := exec.Command("sh", "-c", command).CombinedOutput()
	if err != nil {
		errMsg := fmt.Sprintf("Error executing command: %s\n", err)
		log.Printf("%s\n", errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	log.Printf("Command executed successfully. Output: %s\n", string(cmdOutput))
	w.Header().Set("Content-Type", "text/plain")
	w.Write(cmdOutput)
}

func main() {
	http.HandleFunc("/terminal", handleTerminalConnection) // WebSocket handler
	http.HandleFunc("/execute", executeCommandHandler)     // New HTTP POST handler for commands

	port := "8081"
	log.Printf("Server is starting on port %s...\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Printf("Error starting server: %v\n", err)
	}
}
