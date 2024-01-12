package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

// jsonResponse struct for a simple JSON response
type jsonResponse struct {
	Message string `json:"message"`
}

// configResponse struct to send the Load Balancer URL
type configResponse struct {
	LoadBalancerURL string `json:"loadBalancerUrl"`
}

func getS3URLHandler(w http.ResponseWriter, r *http.Request) {
	// Logic to get the S3 URL
	s3URL := os.Getenv("s3_object_url")
	if s3URL == "" {
		log.Println("s3_object_url environment variable is not set")
		http.Error(w, "s3_object_url environment variable is not set", http.StatusInternalServerError)
		return
	}

	// Create a JSON response with the S3 URL
	response := struct {
		S3URL string `json:"s3_url"`
	}{
		S3URL: s3URL,
	}

	// Set content type and send the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// serverlessAttackHandler handles the serverless attack path
func serverlessAttackHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling /api/serverlesspath request from %s", r.RemoteAddr)
	// Your logic for serverless attack path
	doitformeURL := "http://lazymode:4200/serverless/attack"

	// Make a GET request to the doitforme service
	resp, err := http.Get(doitformeURL)
	if err != nil {
		log.Printf("Error making request to doitforme service: %s", err)
		http.Error(w, "Failed to trigger serverless attack path", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the response from the doitforme service
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response from doitforme service: %s", err)
		http.Error(w, "Failed to read response from serverless attack path", http.StatusInternalServerError)
		return
	}

	// Write the response back to the client
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseBody)
}

// configHandler to handle the configuration route
func configHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling /api/config request from %s", r.RemoteAddr)

	loadBalancerURL := os.Getenv("LOAD_BALANCER_URL")
	if loadBalancerURL == "" {
		log.Println("LOAD_BALANCER_URL environment variable is not set")
		http.Error(w, "LOAD_BALANCER_URL environment variable is not set", http.StatusInternalServerError)
		return
	}

	resp := configResponse{LoadBalancerURL: loadBalancerURL}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding JSON response: %s", err)
	}
}

// productsHandler to handle the products route
func productsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling /api/products request from %s", r.RemoteAddr)

	envVarValue := os.Getenv("AGW_URL")
	resp := jsonResponse{Message: envVarValue}

	log.Printf("AGW_URL environment variable value: %s", envVarValue)

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Error encoding JSON response: %s", err)
	}
}

// uploadHandler to handle the orders route
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling /api/orders request from %s", r.RemoteAddr)

	pythonServiceURL := "http://orders:5000"
	if pythonServiceURL == "" {
		log.Println("Python service URL is not set")
		http.Error(w, "Python service URL is not set", http.StatusInternalServerError)
		return
	}

	client := &http.Client{
		Timeout: time.Second * 30, // Timeout for the HTTP request
	}
	req, err := http.NewRequest("POST", pythonServiceURL+"/upload", r.Body)
	if err != nil {
		log.Printf("Error creating request to Python service: %s", err)
		http.Error(w, "Error creating request to Python service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy headers from the original request
	req.Header = r.Header

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error forwarding request to Python service: %s", err)
		http.Error(w, "Error forwarding request to Python service: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the response body and forward it to the client
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response from Python service: %s", err)
		http.Error(w, "Error reading response from Python service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, err = w.Write(body)
	if err != nil {
		log.Printf("Error writing response: %s", err)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

// handleTerminalConnection handles WebSocket connections for terminal interactions
func handleTerminalConnection(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling /api/terminal WebSocket connection from %s", r.RemoteAddr)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %s", err)
		return
	}
	defer conn.Close()

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading WebSocket message: %s", err)
			return
		}
		log.Printf("Received WebSocket message: %s", p)

		// Execute the command
		cmdOutput, err := exec.Command("sh", "-c", string(p)).CombinedOutput()
		if err != nil {
			errMsg := fmt.Sprintf("Error executing command: %s", err)
			log.Println(errMsg)
			if err := conn.WriteMessage(websocket.TextMessage, []byte(errMsg)); err != nil {
				log.Printf("Error sending WebSocket message: %s", err)
				return
			}
			continue
		}

		log.Printf("Command executed successfully. Output: %s", string(cmdOutput))

		// Send the command output back to the client
		if err := conn.WriteMessage(websocket.TextMessage, cmdOutput); err != nil {
			log.Printf("Error sending WebSocket message: %s", err)
			return
		}
	}
}

func main() {
	log.Println("Starting Backend API server...")

	mux := http.NewServeMux()
	mux.HandleFunc("/api/products", productsHandler)
	mux.HandleFunc("/api/orders", uploadHandler)
	mux.HandleFunc("/api/terminal", handleTerminalConnection)
	mux.HandleFunc("/api/config", configHandler)
	mux.HandleFunc("/api/serverlesspath", serverlessAttackHandler)
	mux.HandleFunc("/api/get-s3", getS3URLHandler)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	handler := c.Handler(mux)
	port := "4567"
	fmt.Printf("Backend API server is running on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
