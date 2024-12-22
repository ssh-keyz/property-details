// Package property provides types and utilities for handling property-related data
package property

import (
	"net/http"
	"time"
)

// Service handles property-related operations
type Service struct {
	httpClient *http.Client
}

// Info represents comprehensive information about a property
type Info struct {
	Address     string      `json:"address"`
	Coordinates Coordinates `json:"coordinates"`
	Details     Details     `json:"details"`
	Schools     []School    `json:"schools"`
}

// Coordinates represents a geographical location
type Coordinates struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// Details contains specific information about a property
type Details struct {
	Size        string  `json:"size"`
	Rooms       int     `json:"rooms"`
	Value       float64 `json:"value"`
	LastUpdated string  `json:"last_updated"`
}

// School represents information about a school near a property
type School struct {
	Name     string  `json:"name"`
	Distance float64 `json:"distance_km"`
	Rating   float64 `json:"rating"`
	Type     string  `json:"type"`
}

// NewService creates a new instance of the property service
func NewService() *Service {
	return &Service{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        60,
				IdleConnTimeout:     60 * time.Second,
				DisableCompression:  false,
				DisableKeepAlives:   false,
				MaxIdleConnsPerHost: 30,
			},
		},
	}
}
