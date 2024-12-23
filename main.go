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

// CORS middleware to handle cross-origin requests
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the origin from the request header
		origin := r.Header.Get("Origin")

		// Allow requests from these origins
		allowedOrigins := map[string]bool{
			"https://property-details-client.vercel.app": true,
			"http://localhost:4321":                      true,
		}

		// If the origin is allowed, set it in the response header
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
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

	// Apply CORS middleware to the property endpoint
	http.HandleFunc("/property", corsMiddleware(server.handleGetProperty))

	port := ":8080"
	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
