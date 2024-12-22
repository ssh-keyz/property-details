// Package opencage provides types and utilities for interacting with the OpenCage API
package opencage

// Components represents the address components from OpenCage API response
type Components struct {
	Type           string `json:"type"`
	Category       string `json:"category"`
	BuildingUse    string `json:"building"`
	HouseNumber    string `json:"house_number"`
	Road           string `json:"road"`
	Suburb         string `json:"suburb"`
	City           string `json:"city"`
	State          string `json:"state"`
	Postcode       string `json:"postcode"`
	Country        string `json:"country"`
	BuildingLevels string `json:"building:levels"`
	Residential    string `json:"residential"`
	Apartments     string `json:"apartments"`
}

// Geometry represents geographical coordinates
type Geometry struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Timezone information from the response
type Timezone struct {
	Name string `json:"name"`
}

// RoadInfo contains details about the road at the location
type RoadInfo struct {
	SpeedLimit string `json:"speed_limit"`
	Surface    string `json:"surface"`
}

// OSM contains OpenStreetMap specific data
type OSM struct {
	BuildingLevels string `json:"building:levels"`
	Amenity        string `json:"amenity"`
	BuildingType   string `json:"building"`
}

// Annotations contains additional data about the location
type Annotations struct {
	Timezone Timezone `json:"timezone"`
	Roadinfo RoadInfo `json:"roadinfo"`
	OSM      OSM      `json:"OSM"`
}

// Result represents a single result from the OpenCage API
type Result struct {
	Confidence  int         `json:"confidence"`
	Components  Components  `json:"components"`
	Formatted   string      `json:"formatted"`
	Geometry    Geometry    `json:"geometry"`
	Annotations Annotations `json:"annotations"`
}

// Status represents the API response status
type Status struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Response represents the complete OpenCage API response
type Response struct {
	Results []Result `json:"results"`
	Status  Status   `json:"status"`
}
