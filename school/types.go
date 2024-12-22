// Package school provides types and utilities for handling school-related data
package school

// Tags represents the OpenStreetMap tags related to schools
type Tags struct {
	Name           string `json:"name"`
	SchoolType     string `json:"amenity:school:type"`
	IsrcRating     string `json:"isced:rating"`
	School         string `json:"amenity"`
	SchoolCategory string `json:"school_category"`
	SchoolLevel    string `json:"school_level"`
	Education      string `json:"education"`
	EducationType  string `json:"education:type"`
}
