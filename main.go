// main.go
package main

import (
	"encoding/json"
	"fmt"
	"math"
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

type SchoolTags struct {
	Name           string `json:"name"`
	SchoolType     string `json:"school:type"`
	IsrcRating     string `json:"isced:rating"`
	School         string `json:"school"`
	SchoolCategory string `json:"school:category"`
	SchoolLevel    string `json:"school:level"`
	Education      string `json:"education"`
	EducationType  string `json:"education_type"`
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

// Zillow API (limited free tier): Property details, valuations, Requires registration, Rate limits apply
// Attom Data Solutions API: Property characteristics, assessments, Free trial available, Good documentation
// APIfy for Realtor.com: Scrape property details, sold properties, rental properties, scrape by keyword, by filter
func (s *PropertyService) GetPropertyInfo(address string) (*PropertyInfo, error) {
	if err := s.ValidateAddress(address); err != nil {
		return nil, fmt.Errorf("address validation failed: %w", err)
	}

	coords, err := s.geocodeAddress(address)
	if err != nil {
		return nil, fmt.Errorf("geocoding failed: %w", err)
	}

	details, err := s.getPropertyDetails(address)
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

func (s *PropertyService) getPropertyDetails(address string) (*PropertyDetails, error) {
	endpoint := fmt.Sprintf(
		"https://api.opencagedata.com/geocode/v1/json?q=%s&key=%s",
		url.QueryEscape(address), os.Getenv("OPENCAGE_API_KEY"),
	)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch property details: %w", err)
	}
	defer resp.Body.Close()

	// First read the raw response for debugging
	var rawResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawResponse); err != nil {
		return nil, fmt.Errorf("failed to decode raw response: %w", err)
	}

	// Print the raw response for debugging
	rawJSON, _ := json.MarshalIndent(rawResponse, "", "  ")
	fmt.Printf("OpenCage API Response:\n%s\n", rawJSON)

	// Now parse into our structured format
	var result struct {
		Results []struct {
			Confidence int `json:"confidence"`
			Components struct {
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
			} `json:"components"`
			Formatted string `json:"formatted"`
			Geometry  struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"geometry"`
			Annotations struct {
				Timezone struct {
					Name string `json:"name"`
				} `json:"timezone"`
				Roadinfo struct {
					SpeedLimit string `json:"speed_limit"`
					Surface    string `json:"surface"`
				} `json:"roadinfo"`
				OSM struct {
					BuildingLevels string `json:"building:levels"`
					Amenity        string `json:"amenity"`
					BuildingType   string `json:"building"`
				} `json:"OSM"`
			} `json:"annotations"`
		} `json:"results"`
		Status struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"status"`
	}

	// Reset the response body with the raw data for parsing
	rawJSON, _ = json.Marshal(rawResponse)
	if err := json.Unmarshal(rawJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to parse structured response: %w", err)
	}

	// Default values when data is unavailable
	details := &PropertyDetails{
		Size:        "Unknown",
		Rooms:       0,
		Value:       0,
		LastUpdated: time.Now().Format(time.RFC3339),
	}

	if len(result.Results) > 0 {
		components := result.Results[0].Components
		annotations := result.Results[0].Annotations

		// Build a comprehensive size description
		sizeDetails := []string{}

		// Check building type
		if components.Type == "residential" || components.Category == "building" {
			if components.BuildingUse != "" {
				sizeDetails = append(sizeDetails, components.BuildingUse)
			}
			if components.Type != "" {
				sizeDetails = append(sizeDetails, components.Type)
			}
		}

		// Add number of levels if available
		if components.BuildingLevels != "" {
			sizeDetails = append(sizeDetails, fmt.Sprintf("%s stories", components.BuildingLevels))
		} else if annotations.OSM.BuildingLevels != "" {
			sizeDetails = append(sizeDetails, fmt.Sprintf("%s stories", annotations.OSM.BuildingLevels))
		}

		// Check if it's an apartment
		if components.Apartments != "" {
			sizeDetails = append(sizeDetails, "apartment building")
		}

		// Set the size description
		if len(sizeDetails) > 0 {
			details.Size = strings.Join(sizeDetails, " ")
		} else {
			details.Size = "Residential Property"
		}

		// Estimate rooms based on building levels
		if levels, err := strconv.Atoi(components.BuildingLevels); err == nil && levels > 0 {
			// Rough estimate: 2 rooms per level for residential buildings
			details.Rooms = levels * 2
		}

		// For debugging
		// fmt.Printf("Property Components:\n")
		// fmt.Printf("  Type: %s\n", components.Type)
		// fmt.Printf("  Category: %s\n", components.Category)
		// fmt.Printf("  BuildingUse: %s\n", components.BuildingUse)
		// fmt.Printf("  Building Levels: %s\n", components.BuildingLevels)
		// fmt.Printf("  Apartments: %s\n", components.Apartments)
		// fmt.Printf("  Address: %s %s, %s, %s %s\n",
		// 	components.HouseNumber,
		// 	components.Road,
		// 	components.City,
		// 	components.State,
		// 	components.Postcode)

		if annotations.OSM.BuildingType != "" {
			fmt.Printf("  OSM Building Type: %s\n", annotations.OSM.BuildingType)
		}
		if annotations.OSM.BuildingLevels != "" {
			fmt.Printf("  OSM Building Levels: %s\n", annotations.OSM.BuildingLevels)
		}
	}

	return details, nil
}

func (s *PropertyService) getNearbySchools(coords *Coordinates) ([]School, error) {
	// First query to get schools
	query := fmt.Sprintf(
		`[out:json][timeout:25];
        (
            way["amenity"="school"]["name"](around:2000,%f,%f);
            relation["amenity"="school"]["name"](around:2000,%f,%f);
            node["amenity"="school"]["name"](around:2000,%f,%f);
        );
        out center;`, // Use 'out center' to get center points for ways and relations
		coords.Lat, coords.Lon,
		coords.Lat, coords.Lon,
		coords.Lat, coords.Lon,
	)

	endpoint := "https://overpass-api.de/api/interpreter"
	resp, err := s.httpClient.Post(endpoint, "text/plain", strings.NewReader(query))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schools: %w", err)
	}
	defer resp.Body.Close()

	// First read the raw response for debugging
	var rawResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawResponse); err != nil {
		return nil, fmt.Errorf("failed to decode raw response: %w", err)
	}

	// Print the raw response for debugging
	rawJSON, _ := json.MarshalIndent(rawResponse, "", "  ")
	fmt.Printf("Overpass API Response:\n%s\n", rawJSON)

	var result struct {
		Elements []struct {
			Type   string     `json:"type"`
			Tags   SchoolTags `json:"tags"`
			Lat    float64    `json:"lat"`
			Lon    float64    `json:"lon"`
			Center *struct {  // Center coordinates for ways and relations
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			} `json:"center"`
		} `json:"elements"`
	}

	// Reset the response body with the raw data for parsing
	rawJSON, _ = json.Marshal(rawResponse)
	if err := json.Unmarshal(rawJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to parse structured response: %w", err)
	}

	schools := make([]School, 0)
	for _, element := range result.Elements {
		if element.Tags.Name == "" {
			continue
		}

		// Get the correct coordinates based on element type
		var schoolLat, schoolLon float64
		if element.Type == "node" {
			schoolLat = element.Lat
			schoolLon = element.Lon
		} else if element.Center != nil {
			// Use center coordinates for ways and relations
			schoolLat = element.Center.Lat
			schoolLon = element.Center.Lon
		} else {
			// Skip if we can't determine coordinates
			fmt.Printf("Skipping school %s: no valid coordinates\n", element.Tags.Name)
			continue
		}

		// Skip if coordinates are invalid (0,0 or out of range)
		if !areValidCoordinates(schoolLat, schoolLon) {
			fmt.Printf("Skipping school %s: invalid coordinates (%f,%f)\n",
				element.Tags.Name, schoolLat, schoolLon)
			continue
		}

		// Calculate direct distance using haversine formula
		distance := calculateDistance(coords.Lat, coords.Lon, schoolLat, schoolLon)

		// Determine school type from various possible tags
		schoolType := determineSchoolType(element.Tags)

		// Parse rating - OSM typically uses values 1-5
		rating := determineSchoolRating(element.Tags)

		school := School{
			Name:     element.Tags.Name,
			Distance: distance,
			Rating:   rating,
			Type:     schoolType,
		}

		// Debug information
		fmt.Printf("School Found:\n")
		fmt.Printf("  Name: %s\n", school.Name)
		fmt.Printf("  Type: %s (%s)\n", element.Type, school.Type)
		fmt.Printf("  Coordinates: %.6f,%.6f\n", schoolLat, schoolLon)
		fmt.Printf("  Distance: %.2f km\n", school.Distance)
		fmt.Printf("  Rating: %.1f\n", school.Rating)
		fmt.Printf("  Raw Tags: %+v\n", element.Tags)

		schools = append(schools, school)
	}

	return schools, nil
}

// Helper function to validate coordinates
func areValidCoordinates(lat, lon float64) bool {
	return lat != 0 && lon != 0 && // Not null island
		lat >= -90 && lat <= 90 && // Valid latitude range
		lon >= -180 && lon <= 180 // Valid longitude range
}

func determineSchoolType(tags SchoolTags) string {
	// Check various OSM tags that might indicate school type
	if tags.SchoolType != "" {
		return strings.Title(tags.SchoolType)
	}
	if tags.SchoolLevel != "" {
		return strings.Title(tags.SchoolLevel)
	}
	if tags.SchoolCategory != "" {
		return strings.Title(tags.SchoolCategory)
	}
	if tags.Education != "" {
		return strings.Title(tags.Education)
	}
	if tags.EducationType != "" {
		return strings.Title(tags.EducationType)
	}
	if tags.School != "" {
		return strings.Title(tags.School)
	}
	return "Unknown"
}

func determineSchoolRating(tags SchoolTags) float64 {
	if tags.IsrcRating != "" {
		if rating, err := strconv.ParseFloat(tags.IsrcRating, 64); err == nil {
			// Normalize rating to 0-5 scale if needed
			if rating > 5 {
				return 5.0
			}
			return rating
		}
	}
	return 0 // Default rating when not available
}

// calculateDistance calculates the shortest distance between two points on Earth's surface using the haversine formula.
//
// The haversine formula determines the great-circle distance between two points on a sphere
// given their latitudes and longitudes. This is the shortest distance over the Earth's surface,
// ignoring terrain, elevation differences, and obstacles.
//
// Parameters:
//   - lat1: Latitude of the first point in decimal degrees
//   - lon1: Longitude of the first point in decimal degrees
//   - lat2: Latitude of the second point in decimal degrees
//   - lon2: Longitude of the second point in decimal degrees
//
// Returns:
//   - The distance between the points in kilometers, rounded to 2 decimal places
//
// Note: This implementation uses the Earth's mean radius of 6371.0 kilometers.
// The accuracy of this calculation decreases for very small distances and near the poles.
// It may be better to use OSRM for more accurate routing.
//
//	More information here: https://www.nextmv.io/blog/haversine-vs-osrm-distance-and-cost-experiments-on-a-vehicle-routing-problem-vrp
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth's radius in kilometers

	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// Differences in coordinates
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	// Haversine formula
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := R * c

	// Round to 2 decimal places
	return math.Round(distance*100) / 100
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
