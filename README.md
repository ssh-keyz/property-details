# Property Service

A service that provides detailed information about properties, including nearby schools and geographical data.

## Running the API Server

1. Build and run the server:
```bash
go run main.go
```

The server will start on port 8080.

## API Endpoints

### Get Property Information

Retrieves detailed information about a property including its location, details, and nearby schools.

```
GET /property?address={urlEncodedAddress}
```

#### Parameters
- `address` (required): The property address, URL encoded

#### Example Request
```bash
curl "http://localhost:8080/property?address=1600%20Amphitheatre%20Parkway%2C%20Mountain%20View%2C%20CA%2094043"
```

#### Example Response
```json
{
  "address": "1600 Amphitheatre Parkway, Mountain View, CA 94043",
  "coordinates": {
    "lat": 37.42248575,
    "lon": -122.08558456613565
  },
  "details": {
    "size": "Mock-Data",
    "rooms": 3,
    "value": 500000,
    "last_updated": "2024-12-22T16:10:22-08:00"
  },
  "schools": [
    {
      "name": "Example School",
      "distance_km": 1.15,
      "rating": 4.5,
      "type": "General School"
    }
  ]
}
```

#### Response Codes
- `200 OK`: Successfully retrieved property information
- `400 Bad Request`: Missing or invalid address parameter
- `500 Internal Server Error`: Server error or invalid address format

## Development

### Running Tests
```bash
go test -v
```

## Features

- Address-based property information lookup
- School district information
- Geocoding support via OpenCage
- Structured JSON output

## Prerequisites

- Go 1.21 or higher
- OpenCage API key (for geocoding functionality)

## Installation

1. Clone the repository:
   ```bash
   git clone github.com/anaheim/property-service
   cd property-service
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

## Building

Build the service using:

```bash
go build -o property-service
```

## Usage

Run the service by providing an address as a command-line argument:

```bash
./property-service "2510 Bancroft Way, Berkeley, CA 94704"
```

The service will return property information in JSON format.

## Project Structure

- `main.go` - Entry point and CLI interface
- `property/` - Core property information service
- `school/` - School district information
- `opencage/` - Geocoding integration

## Dependencies

- `golang.org/x/text` - Text processing utilities

### Code Coverage

?   	github.com/ssh-keyz/property-details/opencage	[no test files]
?   	github.com/ssh-keyz/property-details/school	[no test files]
ok  	github.com/ssh-keyz/property-details	17.095s	coverage: 65.2% of statements
ok  	github.com/ssh-keyz/property-details/property	0.245s	coverage: 86.8% of statements

To run tests with coverage analysis:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # Opens coverage report in browser
```

## License

Copyright © 2023 Anaheim Electronics