package property

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

	"github.com/anaheim/property-service/opencage"
	"github.com/anaheim/property-service/school"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ValidateAddress checks if the provided address is valid
func (s *Service) ValidateAddress(address string) error {
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

// GetInfo retrieves comprehensive information about a property
func (s *Service) GetInfo(address string) (*Info, error) {
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

	return &Info{
		Address:     address,
		Coordinates: *coords,
		Details:     *details,
		Schools:     schools,
	}, nil
}

func (s *Service) geocodeAddress(address string) (*Coordinates, error) {
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

func (s *Service) getPropertyDetails(address string) (*Details, error) {
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

	var rawResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawResponse); err != nil {
		return nil, fmt.Errorf("failed to decode raw response: %w", err)
	}

	var result opencage.Response
	rawJSON, _ := json.Marshal(rawResponse)
	if err := json.Unmarshal(rawJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to parse structured response: %w", err)
	}

	details := &Details{
		Size:        "Mock-Data",
		Rooms:       3,
		Value:       500000,
		LastUpdated: time.Now().Format(time.RFC3339),
	}

	if len(result.Results) > 0 {
		components := result.Results[0].Components
		annotations := result.Results[0].Annotations

		sizeDetails := []string{}

		if components.Type == "residential" || components.Category == "building" {
			if components.BuildingUse != "" {
				sizeDetails = append(sizeDetails, components.BuildingUse)
			}
			if components.Type != "" {
				sizeDetails = append(sizeDetails, components.Type)
			}
		}

		if components.BuildingLevels != "" {
			sizeDetails = append(sizeDetails, fmt.Sprintf("%s stories", components.BuildingLevels))
		} else if annotations.OSM.BuildingLevels != "" {
			sizeDetails = append(sizeDetails, fmt.Sprintf("%s stories", annotations.OSM.BuildingLevels))
		}

		if components.Apartments != "" {
			sizeDetails = append(sizeDetails, "apartment building")
		}

		if len(sizeDetails) > 0 {
			details.Size = strings.Join(sizeDetails, " ")
		} else {
			details.Size = "Residential Property"
		}

		if levels, err := strconv.Atoi(components.BuildingLevels); err == nil && levels > 0 {
			details.Rooms = levels * 2
		}

		if annotations.OSM.BuildingType != "" {
			fmt.Printf("  OSM Building Type: %s\n", annotations.OSM.BuildingType)
		}
		if annotations.OSM.BuildingLevels != "" {
			fmt.Printf("  OSM Building Levels: %s\n", annotations.OSM.BuildingLevels)
		}
	}

	return details, nil
}

func (s *Service) getNearbySchools(coords *Coordinates) ([]School, error) {
	query := fmt.Sprintf(
		`[out:json][timeout:25];
        (
            way["amenity"="school"]["name"](around:2000,%f,%f);
            relation["amenity"="school"]["name"](around:2000,%f,%f);
            node["amenity"="school"]["name"](around:2000,%f,%f);
        );
        out center;`,
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

	var rawResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawResponse); err != nil {
		return nil, fmt.Errorf("failed to decode raw response: %w", err)
	}

	var result struct {
		Elements []struct {
			Type   string      `json:"type"`
			Tags   school.Tags `json:"tags"`
			Lat    float64     `json:"lat"`
			Lon    float64     `json:"lon"`
			Center *struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			} `json:"center"`
		} `json:"elements"`
	}

	rawJSON, _ := json.Marshal(rawResponse)
	if err := json.Unmarshal(rawJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to parse structured response: %w", err)
	}

	schools := make([]School, 0)
	for _, element := range result.Elements {
		if element.Tags.Name == "" {
			continue
		}

		var schoolLat, schoolLon float64
		if element.Type == "node" {
			schoolLat = element.Lat
			schoolLon = element.Lon
		} else if element.Center != nil {
			schoolLat = element.Center.Lat
			schoolLon = element.Center.Lon
		} else {
			fmt.Printf("Skipping school %s: no valid coordinates\n", element.Tags.Name)
			continue
		}

		if !areValidCoordinates(schoolLat, schoolLon) {
			fmt.Printf("Skipping school %s: invalid coordinates (%f,%f)\n",
				element.Tags.Name, schoolLat, schoolLon)
			continue
		}

		distance := calculateDistance(coords.Lat, coords.Lon, schoolLat, schoolLon)
		schoolType := determineSchoolType(element.Tags)
		rating := determineSchoolRating(element.Tags)

		school := School{
			Name:     element.Tags.Name,
			Distance: distance,
			Rating:   rating,
			Type:     schoolType,
		}

		schools = append(schools, school)
	}

	return schools, nil
}

// Helper functions

func areValidCoordinates(lat, lon float64) bool {
	return lat != 0 && lon != 0 &&
		lat >= -90 && lat <= 90 &&
		lon >= -180 && lon <= 180
}

func determineSchoolType(tags school.Tags) string {
	caser := cases.Title(language.English)

	if tags.School != "school" {
		return "Unknown"
	}

	if tags.SchoolType != "" {
		return caser.String(strings.ReplaceAll(tags.SchoolType, "_", " "))
	}
	if tags.SchoolLevel != "" {
		return caser.String(strings.ReplaceAll(tags.SchoolLevel, "_", " "))
	}
	if tags.SchoolCategory != "" {
		return caser.String(strings.ReplaceAll(tags.SchoolCategory, "_", " "))
	}
	if tags.Education != "" {
		return caser.String(strings.ReplaceAll(tags.Education, "_", " "))
	}
	if tags.EducationType != "" {
		return caser.String(strings.ReplaceAll(tags.EducationType, "_", " "))
	}

	return "General School"
}

func mockSchoolRating(name string) float64 {
	var hash uint32
	for i := 0; i < len(name); i++ {
		hash = hash*31 + uint32(name[i])
	}
	rating := 3.0 + (float64(hash%20) / 10.0)
	return math.Round(rating*10) / 10
}

func determineSchoolRating(tags school.Tags) float64 {
	return mockSchoolRating(tags.Name)
}

func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0

	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := R * c

	return math.Round(distance*100) / 100
}
