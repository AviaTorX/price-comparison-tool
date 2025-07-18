# 🛒 AI-Powered Price Comparison Tool

A comprehensive price comparison engine that uses **AI-powered extraction** to intelligently gather product pricing from major e-commerce websites across 8 countries. Get accurate pricing with confidence scoring in seconds.

## 🚀 Key Features

### 🧠 **LLM-First Intelligence**
- **Smart Content Extraction**: AI analyzes entire web pages instead of fragile CSS selectors
- **Contextual Understanding**: LLM understands product specifications, variants, and pricing
- **Robust Fallback**: Automatic fallback to CSS selectors when needed
- **Confidence Scoring**: AI-powered relevance scoring (0-100%) for each result

### ⚡ **High Performance**
- **Fast Results**: Get comprehensive pricing in 30-45 seconds
- **Progress Tracking**: Visual progress bars with estimated completion times
- **Parallel Processing**: Concurrent scraping across all sites for speed
- **Smart Timeouts**: Optimized request handling for reliability

### 🌍 **Global Coverage**
- **8 Countries**: US, Canada, UK, India, Germany, France, Japan, Australia
- **19+ E-commerce Sites**: Amazon (6 regions), eBay, Flipkart, Walmart, Target, Best Buy, etc.
- **Universal Categories**: Electronics, fashion, home goods, and more

### 🔧 **Advanced Architecture**
- **Parallel Processing**: Concurrent scraping across all sites
- **Worker Pools**: 5-worker LLM processing for optimal performance
- **Error Resilience**: Multiple fallback layers ensure reliability
- **Smart Filtering**: Automatic removal of ads and irrelevant content

## 🏃 Quick Start

### Using Docker (Recommended)

1. **Start the enhanced system:**
   ```bash
   docker-compose up -d --build
   ```

2. **Access the web interface:**
   - **🌐 Web UI**: http://localhost:8080
   - **Try searching** for "iPhone 16 Pro, 128GB" in India
   - **Get comprehensive results** from all major e-commerce sites!

3. **Test the API endpoint:**
   ```bash
   # Price Comparison API
   curl -X POST http://localhost:8080/api/v1/prices \
     -H "Content-Type: application/json" \
     -d '{"country": "IN", "query": "iPhone 16 Pro 128GB"}'
   ```

## 📊 Supported Markets

| Country | Sites | Example Results |
|---------|-------|----------------|
| 🇮🇳 **India** | Amazon, Flipkart, Snapdeal, Myntra | iPhone 16 Pro: ₹107,900 (Flipkart) |
| 🇺🇸 **United States** | Amazon, eBay, Walmart, Target, Best Buy | MacBook Air: $999 (Best Buy) |
| 🇨🇦 **Canada** | Amazon, eBay, Walmart | Galaxy S24: CAD $1,200 (Amazon) |
| 🇬🇧 **United Kingdom** | Amazon, eBay | AirPods Pro: £249 (Amazon UK) |
| 🇩🇪 **Germany** | Amazon DE | iPhone 15: €899 (Amazon) |
| 🇫🇷 **France** | Amazon FR | iPad Pro: €1,199 (Amazon) |
| 🇯🇵 **Japan** | Amazon JP | Nintendo Switch: ¥32,978 (Amazon) |
| 🇦🇺 **Australia** | Amazon AU | Surface Pro: AUD $1,699 (Amazon) |

## 🔌 API Reference

### Core Endpoints
- **`GET /api/v1/health`** - System health check
- **`POST /api/v1/prices`** - Price comparison across all sites
- **`GET /api/v1/sites`** - List all supported e-commerce sites

### API Response Format
```json
{
  "results": [
    {
      "productName": "iPhone 16 Pro 128GB",
      "price": "107900",
      "currency": "INR",
      "site": "Flipkart",
      "country": "IN",
      "confidence": 0.95,
      "link": "https://..."
    }
  ],
  "query": "iPhone 16 Pro 128GB",
  "country": "IN",
  "count": 25
}
```

## 🧪 Example Searches

### Web Interface Examples
Visit **http://localhost:8080** and try:

**📱 Electronics:**
- "iPhone 16 Pro, 128GB" (India) → Flipkart ₹107,900, Amazon ₹111,900
- "MacBook Air M2" (US) → Multiple retailers with price comparison
- "Samsung Galaxy S24 Ultra" (UK) → Cross-site pricing

**👕 Fashion:**
- "Nike Air Max 270" (US) → Target, Best Buy, Amazon results
- "Adidas Ultraboost 22" (Canada) → Multi-retailer comparison

**🏠 Home & Garden:**
- "Dyson V15 Vacuum" (Australia) → Local pricing insights
- "KitchenAid Stand Mixer" (Germany) → European market pricing

### API Examples
```bash
# High-confidence electronics search
curl -X POST http://localhost:8080/api/v1/prices \
  -H "Content-Type: application/json" \
  -d '{
    "country": "US",
    "query": "iPad Pro 12.9 inch 256GB"
  }'

# Fashion search with brand specificity  
curl -X POST http://localhost:8080/api/v1/prices \
  -H "Content-Type: application/json" \
  -d '{
    "country": "UK", 
    "query": "Nike Air Force 1 size 10"
  }'

# Home appliance search
curl -X POST http://localhost:8080/api/v1/prices \
  -H "Content-Type: application/json" \
  -d '{
    "country": "CA",
    "query": "Instant Pot Duo 8 quart"
  }'
```

## 🏗️ Architecture Overview

### LLM-First Processing Pipeline
```
Web Page → Content Extraction → LLM Analysis → JSON Parsing → Confidence Scoring → Results
     ↓              ↓                ↓              ↓               ↓            ↓
Navigation     Smart Filtering    Product         Data         AI Relevance   Price
Removal        (Skip ads/nav)     Understanding   Validation   Assessment     Sorting
```

### Technology Stack
- **🔧 Backend**: Go 1.21 with Gin framework
- **🧠 AI/LLM**: Ollama with phi3:mini model for intelligent extraction
- **🕷️ Scraping**: Colly with parallel processing and smart content targeting
- **🎨 Frontend**: Modern HTML5/CSS3/JavaScript with enhanced progress tracking
- **🐳 Deployment**: Docker Compose with multi-service orchestration

### Performance Optimizations
- **Parallel Architecture**: All 19 sites scraped concurrently
- **Worker Pools**: 5 concurrent LLM evaluations
- **Content Chunking**: Optimized 8KB content blocks for LLM processing
- **Smart Caching**: Reduced redundant processing
- **Fallback Systems**: Multiple reliability layers

## 🧑‍💻 Development

### Local Development Setup
```bash
# Clone and setup
git clone <repository-url>
cd price-comparison-tool

# Install dependencies
go mod download

# Run locally (requires Ollama)
go run main.go

# Build production binary
go build -o price-comparison-tool
```

### Environment Configuration
```bash
# Core settings
PORT=8080                    # Server port
OLLAMA_HOST=http://localhost:11434  # LLM service URL

# Development settings  
GIN_MODE=debug              # Enable debug logging
LOG_LEVEL=info              # Logging verbosity
```

### Testing the System
```bash
# Health check
curl http://localhost:8080/api/v1/health

# Test site coverage
curl http://localhost:8080/api/v1/sites | jq '.count'  # Should show 19

# End-to-end test
curl -X POST http://localhost:8080/api/v1/prices \
  -H "Content-Type: application/json" \
  -d '{"country": "IN", "query": "iPhone 16 Pro 128GB"}' | \
  jq '.results[0] | {site, price, currency, confidence}'
```


## 🤝 Contributing

1. **Fork** the repository
2. **Create** feature branch (`git checkout -b feature/amazing-feature`)
3. **Implement** with comprehensive tests
4. **Test** across multiple countries and product types
5. **Submit** pull request with detailed description

## 📄 License

MIT License - See [LICENSE](LICENSE) file for details

---

**🚀 Ready to explore global pricing?** Visit **http://localhost:8080** and start comparing! 

*Built with ❤️ using Go, LLM, and modern web technologies*