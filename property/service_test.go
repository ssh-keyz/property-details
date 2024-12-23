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
	tests := []struct {
		name          string
		geocodeResp   interface{}
		detailsResp   interface{}
		schoolsResp   interface{}
		address       string
		wantErr       bool
		errorResponse bool
		invalidJSON   bool
		httpError     bool
		missingAPIKey bool
		emptyResults  bool
		invalidLatLon bool
		skipTest      bool
	}{
		{
			name: "successful case",
			geocodeResp: []struct {
				Lat string `json:"lat"`
				Lon string `json:"lon"`
			}{
				{Lat: "37.7749", Lon: "-122.4194"},
			},
			detailsResp: map[string]interface{}{
				"results": []map[string]interface{}{
					{
						"components": map[string]interface{}{
							"type":            "residential",
							"building":        "house",
							"building_levels": "2",
							"apartments":      "yes",
						},
						"annotations": map[string]interface{}{
							"OSM": map[string]interface{}{
								"building":      "residential",
								"building_type": "apartments",
							},
						},
					},
				},
			},
			schoolsResp: map[string]interface{}{
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
			},
			address: "123 Main St, San Francisco, CA 94105",
		},
		{
			name:          "geocoding error",
			errorResponse: true,
			address:       "123 Main St, San Francisco, CA 94105",
			wantErr:       true,
		},
		{
			name:        "invalid json response",
			invalidJSON: true,
			address:     "123 Main St, San Francisco, CA 94105",
			wantErr:     true,
		},
		{
			name:      "http error",
			httpError: true,
			address:   "123 Main St, San Francisco, CA 94105",
			wantErr:   true,
			skipTest:  true, // Skip this test as it causes a panic in the test server
		},
		{
			name:          "missing API key",
			missingAPIKey: true,
			address:       "123 Main St, San Francisco, CA 94105",
			wantErr:       true,
		},
		{
			name:         "empty results",
			emptyResults: true,
			address:      "123 Main St, San Francisco, CA 94105",
			wantErr:      true,
		},
		{
			name:          "invalid lat/lon",
			invalidLatLon: true,
			address:       "123 Main St, San Francisco, CA 94105",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		if tt.skipTest {
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.errorResponse {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}

				if tt.httpError {
					panic("connection error")
				}

				if tt.invalidJSON {
					w.Write([]byte("invalid json"))
					return
				}

				w.Header().Set("Content-Type", "application/json")

				switch {
				case strings.Contains(r.URL.Path, "search"):
					if tt.emptyResults {
						json.NewEncoder(w).Encode([]struct{}{})
					} else if tt.invalidLatLon {
						json.NewEncoder(w).Encode([]struct {
							Lat string `json:"lat"`
							Lon string `json:"lon"`
						}{
							{Lat: "invalid", Lon: "invalid"},
						})
					} else {
						json.NewEncoder(w).Encode(tt.geocodeResp)
					}
				case strings.Contains(r.URL.Path, "geocode"):
					if tt.missingAPIKey && r.URL.Query().Get("key") == "" {
						http.Error(w, "Missing API key", http.StatusUnauthorized)
						return
					}
					json.NewEncoder(w).Encode(tt.detailsResp)
				case strings.Contains(r.URL.Path, "interpreter"):
					json.NewEncoder(w).Encode(tt.schoolsResp)
				default:
					http.Error(w, "Not found", http.StatusNotFound)
				}
			}))
			defer server.Close()

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

			client.Transport = RoundTripFunc(func(req *http.Request) (*http.Response, error) {
				req.URL.Scheme = "http"
				req.URL.Host = strings.TrimPrefix(server.URL, "http://")
				return http.DefaultTransport.RoundTrip(req)
			})

			service := &Service{httpClient: client}

			if tt.missingAPIKey {
				os.Unsetenv("OPENCAGE_API_KEY")
			} else {
				os.Setenv("OPENCAGE_API_KEY", "test-key")
			}

			info, err := service.GetInfo(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if info.Address != tt.address {
					t.Errorf("GetInfo().Address = %v, want %v", info.Address, tt.address)
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
			}
		})
	}
}

func TestGeocodeAddress(t *testing.T) {
	tests := []struct {
		name       string
		address    string
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful geocoding",
			address:    "123 Main St, San Francisco, CA 94105",
			response:   `[{"lat": "37.7749", "lon": "-122.4194"}]`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty response",
			address:    "Invalid Address",
			response:   `[]`,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "invalid json",
			address:    "123 Main St",
			response:   `invalid json`,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "server error",
			address:    "123 Main St",
			response:   `Internal Server Error`,
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := &http.Client{
				Transport: RoundTripFunc(func(req *http.Request) (*http.Response, error) {
					req.URL.Scheme = "http"
					req.URL.Host = strings.TrimPrefix(server.URL, "http://")
					return http.DefaultTransport.RoundTrip(req)
				}),
			}

			service := &Service{httpClient: client}
			coords, err := service.geocodeAddress(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("geocodeAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if coords.Lat == 0 || coords.Lon == 0 {
					t.Error("geocodeAddress() returned zero coordinates")
				}
			}
		})
	}
}

func TestGetPropertyDetails(t *testing.T) {
	tests := []struct {
		name       string
		address    string
		response   string
		statusCode int
		wantErr    bool
		wantSize   string
		wantRooms  int
	}{
		{
			name:    "successful case with all details",
			address: "123 Main St, San Francisco, CA 94105",
			response: `{
				"results": [{
					"components": {
						"type": "residential",
						"building": "house",
						"building_levels": "3",
						"apartments": "yes"
					},
					"annotations": {
						"OSM": {
							"building": "residential",
							"building_type": "apartments"
						}
					}
				}]
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantSize:   "house residential apartment building",
			wantRooms:  3,
		},
		{
			name:    "minimal property details",
			address: "123 Main St",
			response: `{
				"results": [{
					"components": {
						"type": "residential"
					}
				}]
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantSize:   "residential",
			wantRooms:  3,
		},
		{
			name:       "server error",
			address:    "123 Main St",
			response:   "Internal Server Error",
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
		{
			name:       "invalid json",
			address:    "123 Main St",
			response:   "invalid json",
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "empty response",
			address:    "123 Main St",
			response:   `{"results": []}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantSize:   "Mock-Data",
			wantRooms:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := &http.Client{
				Transport: RoundTripFunc(func(req *http.Request) (*http.Response, error) {
					req.URL.Scheme = "http"
					req.URL.Host = strings.TrimPrefix(server.URL, "http://")
					return http.DefaultTransport.RoundTrip(req)
				}),
			}

			service := &Service{httpClient: client}
			os.Setenv("OPENCAGE_API_KEY", "test-key")
			defer os.Unsetenv("OPENCAGE_API_KEY")

			details, err := service.getPropertyDetails(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPropertyDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if details.Size != tt.wantSize {
					t.Errorf("getPropertyDetails() size = %v, want %v", details.Size, tt.wantSize)
				}
				if details.Rooms != tt.wantRooms {
					t.Errorf("getPropertyDetails() rooms = %v, want %v", details.Rooms, tt.wantRooms)
				}
				if details.Value != 500000 {
					t.Errorf("getPropertyDetails() value = %v, want 500000", details.Value)
				}
				if details.LastUpdated == "" {
					t.Error("getPropertyDetails() lastUpdated is empty")
				}
			}
		})
	}
}

func TestGetNearbySchools(t *testing.T) {
	tests := []struct {
		name       string
		coords     *Coordinates
		response   string
		statusCode int
		wantErr    bool
		wantCount  int
	}{
		{
			name: "successful case with multiple schools",
			coords: &Coordinates{
				Lat: 37.7749,
				Lon: -122.4194,
			},
			response: `{
				"elements": [
					{
						"type": "node",
						"lat": 37.7749,
						"lon": -122.4194,
						"tags": {
							"name": "Elementary School",
							"amenity": "school",
							"school_type": "elementary"
						}
					},
					{
						"type": "way",
						"center": {
							"lat": 37.7750,
							"lon": -122.4195
						},
						"tags": {
							"name": "High School",
							"amenity": "school",
							"school_level": "secondary"
						}
					},
					{
						"type": "relation",
						"tags": {
							"name": "Invalid School",
							"amenity": "school"
						}
					}
				]
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name: "no schools found",
			coords: &Coordinates{
				Lat: 37.7749,
				Lon: -122.4194,
			},
			response:   `{"elements": []}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  0,
		},
		{
			name: "server error",
			coords: &Coordinates{
				Lat: 37.7749,
				Lon: -122.4194,
			},
			response:   "Internal Server Error",
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
		{
			name: "invalid json",
			coords: &Coordinates{
				Lat: 37.7749,
				Lon: -122.4194,
			},
			response:   "invalid json",
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name: "invalid coordinates in response",
			coords: &Coordinates{
				Lat: 37.7749,
				Lon: -122.4194,
			},
			response: `{
				"elements": [
					{
						"type": "node",
						"lat": 91.0,
						"lon": 181.0,
						"tags": {
							"name": "Invalid School",
							"amenity": "school"
						}
					}
				]
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := &http.Client{
				Transport: RoundTripFunc(func(req *http.Request) (*http.Response, error) {
					req.URL.Scheme = "http"
					req.URL.Host = strings.TrimPrefix(server.URL, "http://")
					return http.DefaultTransport.RoundTrip(req)
				}),
			}

			service := &Service{httpClient: client}
			schools, err := service.getNearbySchools(tt.coords)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNearbySchools() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if len(schools) != tt.wantCount {
					t.Errorf("getNearbySchools() returned %v schools, want %v", len(schools), tt.wantCount)
				}

				for _, school := range schools {
					if school.Name == "" {
						t.Error("getNearbySchools() returned school with empty name")
					}
					if school.Type == "" {
						t.Error("getNearbySchools() returned school with empty type")
					}
					if school.Distance < 0 {
						t.Error("getNearbySchools() returned negative distance")
					}
					if school.Rating < 3.0 || school.Rating > 5.0 {
						t.Error("getNearbySchools() returned invalid rating")
					}
				}
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
