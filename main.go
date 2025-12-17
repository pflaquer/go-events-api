package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

// Event now includes an ImageURL field
type Event struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Date     string `json:"date"`
	Location string `json:"location"`
	ImageURL string `json:"image_url"` // New field for the front-end image
}

var (
	events = make(map[string]Event)
	mu     sync.RWMutex
)

func main() {
	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("GET /api/events", listEvents)
	mux.HandleFunc("POST /api/events", createEvent)
	mux.HandleFunc("GET /api/events/{id}", getEvent)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Events API with Images starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func listEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	mu.RLock()
	defer mu.RUnlock()

	var list []Event = make([]Event, 0)
	for _, e := range events {
		list = append(list, e)
	}
	json.NewEncoder(w).Encode(list)
}

func createEvent(w http.ResponseWriter, r *http.Request) {
	var e Event
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if e.ID == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	mu.Lock()
	events[e.ID] = e
	mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(e)
}

func getEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id") // Go 1.22+ feature

	mu.RLock()
	event, exists := events[id]
	mu.RUnlock()

	if !exists {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}
