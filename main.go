// main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/ssh-keyz/property-details/property"
)

type Server struct {
	service *property.Service
}

func (s *Server) handleGetProperty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Address parameter is required", http.StatusBadRequest)
		return
	}

	// URL decode the address
	decodedAddress, err := url.QueryUnescape(address)
	if err != nil {
		http.Error(w, "Invalid address format", http.StatusBadRequest)
		return
	}

	info, err := s.service.GetInfo(decodedAddress)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting property info: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func main() {
	server := &Server{
		service: property.NewService(),
	}

	http.HandleFunc("/property", server.handleGetProperty)

	port := ":8080"
	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
