// main.go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type PropertyService struct {
	httpClient *http.Client
}

type PropertyInfo struct {
	Address     string          `json:"address"`
	Coordinates Coordinates     `json:"coordinates"`
	Details     PropertyDetails `json:"details"`
	Schools     []School        `json:"schools"`
}

type Coordinates struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type PropertyDetails struct {
	Size        string  `json:"size"`
	Rooms       int     `json:"rooms"`
	Value       float64 `json:"value"`
	LastUpdated string  `json:"last_updated"`
}

type School struct {
	Name     string  `json:"name"`
	Distance float64 `json:"distance_km"`
	Rating   float64 `json:"rating"`
	Type     string  `json:"type"`
}

func NewPropertyService() *PropertyService {
	return &PropertyService{
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

func (s *PropertyService) ValidateAddress(address string) error {
	if strings.TrimSpace(address) == "" {
		return fmt.Errorf("address cannot be empty")
	}

	parts := strings.Split(address, ",")
	if len(parts) < 3 {
		return fmt.Errorf("address must include street, city, and state")
	}

	addressRegex := regexp.MustCompile(`^\d+\s+[A-Za-z0-9\s.-]+,\s*[A-Za-z\s]+,\s*[A-Z]{2}\s*\d{5}?$`)
	if !addressRegex.MatchString(strings.TrimSpace(address)) {
		return fmt.Errorf("invalid address format")
	}

	return nil
}

func (s *PropertyService) GetPropertyInfo(address string) (*PropertyInfo, error) {
	if err := s.ValidateAddress(address); err != nil {
		return nil, fmt.Errorf("address validation failed: %w", err)
	}

	coords, err := s.geocodeAddress(address)
	if err != nil {
		return nil, fmt.Errorf("geocoding failed: %w", err)
	}

	details, err := s.getPropertyDetails(coords)
	if err != nil {
		return nil, fmt.Errorf("failed to get property details: %w", err)
	}

	schools, err := s.getNearbySchools(coords)
	if err != nil {
		return nil, fmt.Errorf("failed to get nearby schools: %w", err)
	}

	return &PropertyInfo{
		Address:     address,
		Coordinates: *coords,
		Details:     *details,
		Schools:     schools,
	}, nil
}

func (s *PropertyService) geocodeAddress(address string) (*Coordinates, error) {
	endpoint := fmt.Sprintf(
		"https://nominatim.openstreetmap.org/search?q=%s&format=json&limit=1",
		url.QueryEscape(address),
	)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "PropertyInfoService/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("address not found")
	}

	lat, err := strconv.ParseFloat(results[0].Lat, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid latitude value: %w", err)
	}

	lon, err := strconv.ParseFloat(results[0].Lon, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid longitude value: %w", err)
	}

	return &Coordinates{
		Lat: lat,
		Lon: lon,
	}, nil
}

func (s *PropertyService) getPropertyDetails(coords *Coordinates) (*PropertyDetails, error) {
	endpoint := fmt.Sprintf(
		"https://api.opencagedata.com/geocode/v1/json?q=%f+%f&key=%s",
		coords.Lat, coords.Lon, os.Getenv("OPENCAGE_API_KEY"),
	)

	// endpoint := fmt.Sprintf(
	// 	"https://api.opencagedata.com/geocode/v1/json?q=%f+%f&key=%s",
	// 	coords.Lat, coords.Lon, "97d4254c990844bfa9edbd65c5b4fd1b",
	// )

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch property details: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Results []struct {
			Components struct {
				Type        string `json:"type"`
				BuildingUse string `json:"building"`
				HouseNumber string `json:"house_number"`
			} `json:"components"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Default values when data is unavailable
	details := &PropertyDetails{
		Size:        "Unknown",
		Rooms:       0,
		Value:       0,
		LastUpdated: time.Now().Format(time.RFC3339),
	}

	if len(result.Results) > 0 {
		// Populate available data
		if result.Results[0].Components.Type == "residential" {
			details.Size = "Residential Property"
		}
	}

	return details, nil
}

func (s *PropertyService) getNearbySchools(coords *Coordinates) ([]School, error) {
	query := fmt.Sprintf(
		`[out:json][timeout:25];
        (
            way["amenity"="school"]["name"](around:2000,%f,%f);
            relation["amenity"="school"]["name"](around:2000,%f,%f);
        );
        out body;
        >;
        out skel qt;`,
		coords.Lat, coords.Lon,
		coords.Lat, coords.Lon,
	)

	endpoint := "https://overpass-api.de/api/interpreter"
	resp, err := s.httpClient.Post(endpoint, "text/plain", strings.NewReader(query))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schools: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Elements []struct {
			Tags struct {
				Name       string `json:"name"`
				SchoolType string `json:"school:type"`
				IsrcRating string `json:"isced:rating"`
			} `json:"tags"`
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"elements"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	schools := make([]School, 0)
	for _, element := range result.Elements {
		if element.Tags.Name == "" {
			continue
		}

		distance := calculateDistance(coords.Lat, coords.Lon, element.Lat, element.Lon)

		school := School{
			Name:     element.Tags.Name,
			Distance: distance,
			Rating:   parseRating(element.Tags.IsrcRating),
			Type:     parseSchoolType(element.Tags.SchoolType),
		}
		schools = append(schools, school)
	}

	return schools, nil
}

func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	endpoint := fmt.Sprintf(
		"https://router.project-osrm.org/route/v1/driving/%f,%f;%f,%f?overview=false",
		lon1, lat1, // OSRM expects coordinates in lon,lat order
		lon2, lat2,
	)

	resp, err := http.Get(endpoint)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	var result struct {
		Routes []struct {
			Distance float64 `json:"distance"`
		} `json:"routes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0
	}

	if len(result.Routes) == 0 {
		return 0
	}

	// Convert meters to miles
	return result.Routes[0].Distance / 1609.34
}

func parseRating(rating string) float64 {
	if rating == "" {
		return 0
	}
	r, err := strconv.ParseFloat(rating, 64)
	if err != nil {
		return 0
	}
	return r
}

func parseSchoolType(schoolType string) string {
	if schoolType == "" {
		return "Unknown"
	}
	return strings.Title(schoolType)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: property-service \"<address>\"")
		os.Exit(1)
	}

	address := os.Args[1]
	service := NewPropertyService()

	info, err := service.GetPropertyInfo(address)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	output, _ := json.MarshalIndent(info, "", "  ")
	fmt.Println(string(output))
}
