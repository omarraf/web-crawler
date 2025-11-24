# Web Scraper

A high-performance web scraping and crawling service built with Go.

## Tech Stack

- **Go** (1.25.3) - Core programming language
- **Chi Router** (v5.2.3) - Lightweight, idiomatic HTTP router
- **CORS Middleware** (v1.2.2) - Cross-origin resource sharing support
- **godotenv** (v1.5.1) - Environment configuration management

## System Architecture

```
┌─────────────────┐
│   HTTP Client   │
│  (Web Browser)  │
└────────┬────────┘
         │
         │ HTTP/HTTPS
         │
         ▼
┌─────────────────────────────────┐
│      CORS Middleware            │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│      Chi Router (v1)            │
│  ┌───────────────────────────┐  │
│  │  /v1/healthz              │  │
│  │  /v1/scrape (future)      │  │
│  │  /v1/crawl  (future)      │  │
│  └───────────────────────────┘  │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│     Handler Layer               │
│  - Health Check                 │
│  - JSON Response Utilities      │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│     Business Logic              │
│  - Web Scraping Engine          │
│  - Data Extraction              │
│  - Content Processing           │
└─────────────────────────────────┘
```

## Features

- RESTful API architecture with versioned endpoints
- CORS-enabled for cross-origin requests
- Health monitoring endpoint
- Environment-based configuration
- JSON response formatting
- Extensible middleware pipeline
- Structured error handling

## Installation

### Prerequisites

- Go 1.25.3 or higher
- Git

### Setup

1. Clone the repository:
```bash
git clone https://github.com/omarraf/web-scraper.git
cd web-scraper
```

2. Install dependencies:
```bash
go mod download
go mod vendor
```

3. Create a `.env` file in the root directory:
```bash
PORT=8080
```

4. Build the application:
```bash
go build -o web-scraper
```

## Configuration

The application uses environment variables for configuration. Create a `.env` file with the following variables:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PORT` | HTTP server port | - | Yes |

## Usage

### Running the Server

```bash
./web-scraper
```

Or run directly with Go:

```bash
go run .
```

The server will start on the configured port (default: 8080).

### API Endpoints

#### Health Check

**GET** `/v1/healthz`

Check the health status of the service.

**Response:**
```json
{}
```

**Status Codes:**
- `200 OK` - Service is healthy

## Development

### Project Structure

```
.
├── main.go                  # Application entry point
├── handler_readiness.go     # Health check handler
├── json.go                  # JSON response utilities
├── go.mod                   # Go module definition
├── go.sum                   # Dependency checksums
├── .env                     # Environment configuration
└── vendor/                  # Vendored dependencies
```

### Building from Source

```bash
go build -o web-scraper
```

### Running Tests

```bash
go test ./...
```






