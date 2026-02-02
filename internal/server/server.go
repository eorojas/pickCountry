package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/eorojas/pickCountry/internal/app"
)

type Server struct {
	Manager *app.Manager
	Port    int
}

type InputRequest struct {
	Key  string `json:"key"`
	Code string `json:"code"` // JS event code
}

func (s *Server) stateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // For local dev if needed
	
	state := s.Manager.GetState()
	if err := json.NewEncoder(w).Encode(state); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) inputHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req InputRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.Manager.ProcessInput(req.Key, req.Code)

	// Return updated state immediately
	s.stateHandler(w, r)
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	// Simple static file server for now, or embedded HTML
	// For this step, let's serve a basic HTML string if index.html is missing,
	// or serve from a static dir.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "static/index.html")
}

func (s *Server) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/state", s.stateHandler)
	mux.HandleFunc("/api/input", s.inputHandler)
	mux.HandleFunc("/", s.indexHandler)

	addr := fmt.Sprintf(":%d", s.Port)
	log.Printf("Server listening on http://localhost%s", addr)
	return http.ListenAndServe(addr, mux)
}
