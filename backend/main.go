package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

type jsonResponse struct {
	Message string `json:"message"`
}

type configResponse struct {
	LoadBalancerURL string `json:"loadBalancerUrl"`
}

func getS3URLHandler(w http.ResponseWriter, r *http.Request) {
	s3URL := os.Getenv("s3_object_url")
	if s3URL == "" {
		log.Println("s3_object_url environment variable is not set")
		http.Error(w, "s3_object_url environment variable is not set", http.StatusInternalServerError)
		return
	}

	response := struct {
		S3URL string `json:"s3_url"`
	}{S3URL: s3URL}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func serverlessAttackHandler(w http.ResponseWriter, r *http.Request) {
	doitformeURL := "http://lazymode:4200/serverless/attack"
	resp, err := http.Get(doitformeURL)
	if err != nil {
		log.Printf("Error making request to doitforme service: %s", err)
		http.Error(w, "Failed to trigger serverless attack path", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response from doitforme service: %s", err)
		http.Error(w, "Failed to read response from serverless attack path", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseBody)
}

func configHandler(w http.ResponseWriter, r *http.Request) {
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

func productsHandler(w http.ResponseWriter, r *http.Request) {
	envVarValue := os.Getenv("AGW_URL")
	resp := jsonResponse{Message: envVarValue}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding JSON response: %s", err)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	pythonServiceURL := "http://orders:5000"
	client := &http.Client{
		Timeout: time.Second * 30,
	}
	req, err := http.NewRequest("POST", pythonServiceURL+"/upload", r.Body)
	if err != nil {
		log.Printf("Error creating request to Python service: %s", err)
		http.Error(w, "Error creating request to Python service: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header = r.Header

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error forwarding request to Python service: %s", err)
		http.Error(w, "Error forwarding request to Python service: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response from Python service: %s", err)
		http.Error(w, "Error reading response from Python service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func forwardCommandToTerminal(command string) (string, error) {
	terminalURL := "http://terminal:8081/execute"
	client := &http.Client{
		Timeout: time.Second * 30,
	}
	req, err := http.NewRequest("POST", terminalURL, strings.NewReader(command))
	if err != nil {
		return "", fmt.Errorf("creating request failed: %v", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending command to terminal failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading terminal response failed: %v", err)
	}

	return string(body), nil
}

func handleTerminalConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading WebSocket message: %v", err)
			return
		}

		output, err := forwardCommandToTerminal(string(p))
		if err != nil {
			errorMsg := fmt.Sprintf("Error forwarding command: %v", err)
			conn.WriteMessage(websocket.TextMessage, []byte(errorMsg))
			continue
		}

		if err := conn.WriteMessage(websocket.TextMessage, []byte(output)); err != nil {
			log.Printf("Error sending WebSocket message: %v", err)
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
