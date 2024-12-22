// main.go
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/anaheim/property-service/property"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: property-service \"<address>\"")
		os.Exit(1)
	}

	address := os.Args[1]
	service := property.NewService()

	info, err := service.GetInfo(address)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	output, _ := json.MarshalIndent(info, "", "  ")
	fmt.Println(string(output))
}
