# Price Comparison Tool

A high-performance Go-based web service that fetches product prices from multiple e-commerce websites across different countries.

## Features

- 🌍 **Global Coverage**: Supports multiple countries (US, India, UK, Canada)
- 🛍️ **Multi-Site Scraping**: Fetches from Amazon, eBay, Flipkart, Best Buy, and more
- 🚀 **Concurrent Processing**: Fast parallel scraping using Go goroutines
- 🤖 **AI-Powered Matching**: Local LLM integration for smart product filtering
- 💰 **Price Sorting**: Results sorted by price with confidence scores
- 🎯 **REST API**: Clean JSON API endpoints
- 🖥️ **Web Interface**: Simple frontend for testing
- 🐳 **Docker Ready**: Complete containerized deployment with Ollama

## Quick Start

### Using Docker (Recommended)

```bash
# Clone the repository
git clone <your-repo-url>
cd price-comparison-tool

# Start with docker-compose
docker-compose up --build

# The service will be available at http://localhost:8080
```

### Manual Setup

```bash
# Install Go 1.19+
# Clone the repository
git clone <your-repo-url>
cd price-comparison-tool

# Install dependencies
go mod tidy

# Run the service
go run main.go

# Service available at http://localhost:8080
```

## API Usage

### Search for Prices

**POST** `/api/v1/prices`

**Request Body:**
```json
{
  "country": "US",
  "query": "iPhone 16 Pro, 128GB"
}
```

**Response:**
```json
{
  "results": [
    {
      "link": "https://amazon.com/...",
      "price": "999.00",
      "currency": "USD",
      "productName": "Apple iPhone 16 Pro 128GB",
      "site": "Amazon US",
      "country": "US",
      "fetchedAt": "2025-01-08T..."
    }
  ],
  "query": "iPhone 16 Pro, 128GB",
  "country": "US",
  "count": 15
}
```

### Health Check

**GET** `/api/v1/health`

### Supported Sites

**GET** `/api/v1/sites`

## Testing Examples

### cURL Commands

```bash
# Test the required iPhone query
curl -X POST http://localhost:8080/api/v1/prices \
  -H "Content-Type: application/json" \
  -d '{"country": "US", "query": "iPhone 16 Pro, 128GB"}'

# Test Indian market
curl -X POST http://localhost:8080/api/v1/prices \
  -H "Content-Type: application/json" \
  -d '{"country": "IN", "query": "boAt Airdopes 311 Pro"}'

# Health check
curl http://localhost:8080/api/v1/health
```

### Web Interface

Visit `http://localhost:8080` to use the web interface for testing.

## Supported Sites by Country

### United States (US)
- Amazon US
- eBay US  
- Best Buy

### India (IN)
- Amazon India
- Flipkart

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Web Frontend  │    │    REST API      │    │  Scraper Pool   │
│                 │────┤                  │────┤                 │
│  (HTML/JS)      │    │   (Gin/HTTP)     │    │ (Colly/Workers) │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │                       │
                                │                ┌──────┴──────┐
                                │                │ Local LLM   │
                       ┌────────┴────────┐      │ (Ollama)    │
                       │ Product Matcher │      │ phi3:mini   │
                       │ (AI Filtering)  │      └─────────────┘
                       └─────────────────┘
```

## Performance

- **Concurrent Scraping**: Up to 50 parallel requests
- **Rate Limiting**: Respects site limits (1.5-3s delays)
- **AI Filtering**: Local LLM processing with 30s timeout
- **Smart Matching**: Confidence-based result filtering
- **Timeout Handling**: 30s request timeouts

## Development

### Project Structure

```
├── main.go                 # Entry point
├── internal/
│   ├── api/               # HTTP handlers
│   ├── config/            # Configuration
│   ├── models/            # Data structures  
│   └── scraper/           # Web scraping logic
├── web/
│   ├── templates/         # HTML templates
│   └── static/           # Static assets
├── Dockerfile            # Container definition
└── docker-compose.yml    # Multi-service setup
```

### Adding New Sites

1. Update `loadSiteConfigs()` in `internal/scraper/service.go`
2. Add site configuration with selectors
3. Test with sample queries

### Environment Variables

- `PORT`: Server port (default: 8080)
- `REDIS_URL`: Redis connection string
- `OPENAI_API_KEY`: For LLM-based matching (optional)
- `GIN_MODE`: Gin framework mode (release/debug)

## Deployment

### Local Testing
```bash
docker-compose up --build
```

### Production Deployment

The application can be deployed to:
- **Vercel**: Frontend + Serverless functions
- **Railway**: Full-stack deployment
- **DigitalOcean**: Docker containers
- **AWS/GCP**: Container services

## Proof of Working

### Required Test Case

```bash
curl -X POST http://localhost:8080/api/v1/prices \
  -H "Content-Type: application/json" \
  -d '{"country": "US", "query": "iPhone 16 Pro, 128GB"}'
```

**Expected**: Returns price results from multiple US retailers.

## License

MIT License