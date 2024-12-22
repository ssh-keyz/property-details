// main_test.go
package property

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/ssh-keyz/property-details/school"
)

func TestValidateAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string

		wantErr bool
	}{
		{
			name:    "valid address",
			address: "123 Main St, San Francisco, CA 94105",
			wantErr: false,
		},
		{
			name:    "empty address",
			address: "",
			wantErr: true,
		},
		{
			name:    "incomplete address",
			address: "123 Main St",
			wantErr: true,
		},
		{
			name:    "invalid format",
			address: "Invalid Address Format",
			wantErr: true,
		},
	}

	service := NewService()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateAddress(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAreValidCoordinates(t *testing.T) {
	tests := []struct {
		name string
		lat  float64
		lon  float64
		want bool
	}{
		{
			name: "valid coordinates",
			lat:  37.7749,
			lon:  -122.4194,
			want: true,
		},
		{
			name: "invalid latitude",
			lat:  91.0,
			lon:  -122.4194,
			want: false,
		},
		{
			name: "invalid longitude",
			lat:  37.7749,
			lon:  181.0,
			want: false,
		},
		{
			name: "zero coordinates",
			lat:  0.0,
			lon:  0.0,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AreValidCoordinates(tt.lat, tt.lon); got != tt.want {
				t.Errorf("AreValidCoordinates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetermineSchoolType(t *testing.T) {
	tests := []struct {
		name string
		tags school.Tags
		want string
	}{
		{
			name: "school type specified",
			tags: school.Tags{
				School:     "school",
				SchoolType: "elementary",
			},
			want: "Elementary",
		},
		{
			name: "school level specified",
			tags: school.Tags{
				School:      "school",
				SchoolLevel: "secondary",
			},
			want: "Secondary",
		},
		{
			name: "education type specified",
			tags: school.Tags{
				School:        "school",
				EducationType: "high_school",
			},
			want: "High School",
		},
		{
			name: "no specific type",
			tags: school.Tags{
				School: "school",
			},
			want: "General School",
		},
		{
			name: "not a school",
			tags: school.Tags{
				School: "not_school",
			},
			want: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetermineSchoolType(tt.tags); got != tt.want {
				t.Errorf("DetermineSchoolType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateDistance(t *testing.T) {
	tests := []struct {
		name     string
		lat1     float64
		lon1     float64
		lat2     float64
		lon2     float64
		wantDist float64
		epsilon  float64
	}{
		{
			name:     "same point",
			lat1:     37.7749,
			lon1:     -122.4194,
			lat2:     37.7749,
			lon2:     -122.4194,
			wantDist: 0.0,
			epsilon:  0.01,
		},
		{
			name:     "SF to LA",
			lat1:     37.7749,
			lon1:     -122.4194,
			lat2:     34.0522,
			lon2:     -118.2437,
			wantDist: 559.12, // Approximate distance in km
			epsilon:  1.0,    // Allow 1km error margin
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateDistance(tt.lat1, tt.lon1, tt.lat2, tt.lon2)
			diff := got - tt.wantDist
			if diff < -tt.epsilon || diff > tt.epsilon {
				t.Errorf("CalculateDistance() = %v, want %v (±%v)", got, tt.wantDist, tt.epsilon)
			}
		})
	}
}

func TestMockSchoolRating(t *testing.T) {
	tests := []struct {
		name       string
		schoolName string
		wantRange  [2]float64 // min and max expected rating
	}{
		{
			name:       "basic school",
			schoolName: "Test School",
			wantRange:  [2]float64{3.0, 5.0},
		},
		{
			name:       "empty name",
			schoolName: "",
			wantRange:  [2]float64{3.0, 5.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mockSchoolRating(tt.schoolName)
			if got < tt.wantRange[0] || got > tt.wantRange[1] {
				t.Errorf("mockSchoolRating() = %v, want between %v and %v",
					got, tt.wantRange[0], tt.wantRange[1])
			}
		})
	}
}

func TestGetInfo(t *testing.T) {
	// Create a mock server for all HTTP requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "search"):
			json.NewEncoder(w).Encode([]struct {
				Lat string `json:"lat"`
				Lon string `json:"lon"`
			}{
				{Lat: "37.7749", Lon: "-122.4194"},
			})
		case strings.Contains(r.URL.Path, "geocode"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"results": []map[string]interface{}{
					{
						"components": map[string]interface{}{
							"type":            "residential",
							"building":        "house",
							"building_levels": "2",
						},
						"annotations": map[string]interface{}{
							"OSM": map[string]interface{}{
								"building": "residential",
							},
						},
					},
				},
			})
		case strings.Contains(r.URL.Path, "interpreter"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"elements": []map[string]interface{}{
					{
						"type": "node",
						"lat":  37.7749,
						"lon":  -122.4194,
						"tags": map[string]interface{}{
							"name":        "Test School",
							"amenity":     "school",
							"school_type": "elementary",
						},
					},
				},
			})
		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create a custom client that redirects all requests to our test server
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(server.URL)
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// Create a custom RoundTripper that modifies the request URL
	client.Transport = RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		// Replace the https URL with our test server URL
		req.URL.Scheme = "http"
		req.URL.Host = strings.TrimPrefix(server.URL, "http://")
		return http.DefaultTransport.RoundTrip(req)
	})

	service := &Service{httpClient: client}
	validAddress := "123 Main St, San Francisco, CA 94105"

	// Set environment variable for testing
	os.Setenv("OPENCAGE_API_KEY", "test-key")
	defer os.Unsetenv("OPENCAGE_API_KEY")

	info, err := service.GetInfo(validAddress)
	if err != nil {
		t.Fatalf("GetInfo() error = %v", err)
	}

	// Verify the response structure
	if info.Address != validAddress {
		t.Errorf("GetInfo().Address = %v, want %v", info.Address, validAddress)
	}

	if len(info.Schools) == 0 {
		t.Error("GetInfo().Schools is empty")
	}

	if info.Details.Size == "" {
		t.Error("GetInfo().Details.Size is empty")
	}

	if info.Details.Rooms <= 0 {
		t.Error("GetInfo().Details.Rooms should be positive")
	}

	if info.Details.Value <= 0 {
		t.Error("GetInfo().Details.Value should be positive")
	}

	if info.Details.LastUpdated == "" {
		t.Error("GetInfo().Details.LastUpdated is empty")
	}

	// Test error cases
	errorCases := []struct {
		name    string
		address string
	}{
		{
			name:    "empty address",
			address: "",
		},
		{
			name:    "invalid address",
			address: "invalid",
		},
	}

	for _, tt := range errorCases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetInfo(tt.address)
			if err == nil {
				t.Error("GetInfo() error = nil, want error")
			}
		})
	}
}

// RoundTripFunc allows us to use a function as an http.RoundTripper
type RoundTripFunc func(*http.Request) (*http.Response, error)

// RoundTrip implements the http.RoundTripper interface
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
