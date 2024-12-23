package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ssh-keyz/property-details/property"
)

func TestHandleGetProperty(t *testing.T) {
	// Create a new server instance
	server := &Server{
		service: property.NewService(),
	}

	tests := []struct {
		name           string
		address        string
		expectedStatus int
		wantErr        bool
	}{
		{
			name:           "valid address",
			address:        "1600 Amphitheatre Parkway, Mountain View, CA 94043",
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "empty address",
			address:        "",
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:           "invalid address format",
			address:        "invalid!!!address",
			expectedStatus: http.StatusInternalServerError,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new HTTP request with URL encoded address
			reqURL := "/property"
			if tt.address != "" {
				reqURL += "?address=" + url.QueryEscape(tt.address)
			}
			req := httptest.NewRequest(http.MethodGet, reqURL, nil)
			w := httptest.NewRecorder()

			// Call the handler
			server.handleGetProperty(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("handleGetProperty() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			// For successful requests, verify the response structure
			if !tt.wantErr && w.Code == http.StatusOK {
				var response property.Info
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				// Verify the response contains the expected address
				if response.Address != tt.address {
					t.Errorf("Response address = %v, want %v", response.Address, tt.address)
				}

				// Verify coordinates are present
				if response.Coordinates.Lat == 0 || response.Coordinates.Lon == 0 {
					t.Error("Response coordinates are missing")
				}

				// Verify property details are present
				if response.Details.Size == "" {
					t.Error("Response property details are missing")
				}

				// Verify schools array is present and not empty
				if len(response.Schools) == 0 {
					t.Error("Response schools array is empty")
				}
			}
		})
	}
}

func TestMethodNotAllowed(t *testing.T) {
	server := &Server{
		service: property.NewService(),
	}

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/property?address=test", nil)
			w := httptest.NewRecorder()

			server.handleGetProperty(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s request: got status %v, want %v", method, w.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}
