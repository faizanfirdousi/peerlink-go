package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// structure of the data a peer sends when registering
type RegisterRequest struct {
	Peer string `json:"peer"`
	Files []string `json:"files"`
}


var (
	fileIndex = make(map[string][]string) //global map where filename : list of peers hosting that file
	mu sync.Mutex
)

// POST /register

func registerHandler(w http.ResponseWriter, r *http.Request){
	// ensures the handler only accepts POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	//JSON decoding, parses the incoming JSON request body and stores it in req
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		//If the JSON is malformed, it returns an error with 400 Bad Request.
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}


	mu.Lock()
	for _, file := range req.Files {
		fileIndex[file] = append(fileIndex[file], req.Peer)
	}
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Peer %s registered with %d files\n", req.Peer, len(req.Files))
}

func findHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	file := r.URL.Query().Get("file")
	if file == "" {
		http.Error(w, "file query param required", http.StatusBadRequest)
		return
	}

	mu.Lock()
	peers, exists := fileIndex[file]
	mu.Unlock()

	if !exists{
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string][]string{"peers": peers})
}

func main() {
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/find", findHandler)

	fmt.Println("Bootstrap server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}



