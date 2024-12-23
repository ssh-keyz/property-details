# Property Service

A Go-based service that provides detailed property information based on an address input. This service integrates with various data sources to provide comprehensive property and location details.

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

## Development

To run tests:

```bash
go test ./...
```

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