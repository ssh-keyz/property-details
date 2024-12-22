// main_test.go
package main

import (
    "testing"
    "time"
)

func TestValidateAddress(t *testing.T) {
    tests := []struct {
        name    string
        address string
        wantErr bool
    }{
        {
            name:    "valid address",
            address: "123 Main St, Springfield, IL 62701",
            wantErr: false,
        },
        {
            name:    "empty address",
            address: "",
            wantErr: true,
        },
        {
            name:    "missing components",
            address: "123 Main St",
            wantErr: true,
        },
        {
            name:    "invalid format",
            address: "123,,IL",
            wantErr: true,
        },
        {
            name:    "special characters",
            address: "123 Main St. #4B, Springfield, IL 62701",
            wantErr: true,
        },
    }

    service := NewPropertyService()

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := service.ValidateAddress(tt.address)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateAddress() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

func TestGetPropertyInfo(t *testing.T) {
    service := NewPropertyService()

    t.Run("successful request", func(t *testing.T) {
        address := "123 Main St, Springfield, IL 62701"
        info, err := service.GetPropertyInfo(address)
        if err != nil {
            t.Errorf("GetPropertyInfo() error = %v", err)
            return
        }

        if info.Address != address {
            t.Errorf("Expected address %v, got %v", address, info.Address)
        }

        if len(info.Schools) == 0 {
            t.Error("Expected schools to be returned")
        }

        if info.Details.Size == "" {
            t.Error("Expected property size to be non-empty")
        }
    })

    t.Run("invalid address", func(t *testing.T) {
        _, err := service.GetPropertyInfo("")
        if err == nil {
            t.Error("Expected error for empty address")
        }
    })
}

func TestGeocodeAddress(t *testing.T) {
    service := NewPropertyService()

    t.Run("valid address", func(t *testing.T) {
        coords, err := service.geocodeAddress("123 Main St, Springfield, IL 62701")
        if err != nil {
            t.Errorf("geocodeAddress() error = %v", err)
            return
        }

        if coords.Lat == 0 || coords.Lon == 0 {
            t.Error("Expected non-zero coordinates")
        }
    })

    t.Run("invalid address", func(t *testing.T) {
        _, err := service.geocodeAddress("Invalid Address")
        if err == nil {
            t.Error("Expected error for invalid address")
        }
    })
}
