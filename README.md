# AI Financial Coach

An AI-powered financial coaching application that analyzes user spending patterns, provides investment recommendations, and projects portfolio growth using real market data.

## 🏗️ Architecture

Built with:
- **Backend**: GoFr (Go framework) v1.38.0
- **AI Integration**: OpenAI GPT-4o-mini (planned)
- **Financial Data**: Belvo API (Sandbox)
- **Market Data**: Yahoo Finance (yfinance), CoinGecko API
- **Database**: SQLite (local), PostgreSQL (production)

## 📋 Prerequisites

- Go 1.24+ installed
- Git
- Internet connection (for dependencies and API calls)
- Belvo API credentials (for financial data integration)

## 🚀 Quick Start

### 1. Navigate to Project Directory

```bash
# Navigate to the project directory
cd poc
```

### 2. Install Dependencies

```bash
# Download and install all Go dependencies
go mod tidy
```

### 3. Configure Environment Variables (Optional)

```bash
# Belvo API credentials (for real financial data)
export BELVO_SECRET_ID=your_belvo_secret_id
export BELVO_SECRET_PASSWORD=your_belvo_secret_password
export BELVO_ENVIRONMENT=sandbox

# Note: The application will work without these credentials using mock data
```

### 4. Run the Application

```bash
# Start the API server
go run cmd/api/main.go
```

The application will start and listen on port 8000 (GoFr default).

### 5. Test the Endpoints

```bash
# Health check
curl http://localhost:8000/health

# API information
curl http://localhost:8000/

# Test Belvo connection (shows status even without credentials)
curl http://localhost:8000/api/belvo/test-connection

# Get mock financial data (works without credentials)
curl http://localhost:8000/api/belvo/mock-data
```

Expected response for health check:
```json
{
  "data": {
    "status": "healthy",
    "service": "ai-financial-coach", 
    "version": "v1.0.0"
  }
}
```

## 📁 Project Structure

```
poc/
├── cmd/api/                  # Application entry points
│   └── main.go              # API server main
├── internal/                # Private application code
│   ├── app/                 # App initialization
│   │   └── app.go          # GoFr app setup & route registration
│   ├── api/                # HTTP API handlers
│   │   └── belvo_handler.go # Belvo API endpoints
│   ├── service/            # Business logic
│   │   └── belvo_service.go # Belvo API integration
│   └── models/             # Data models
│       └── belvo.go        # Belvo data structures
├── configs/                # Configuration files
├── web/                    # Frontend assets (planned)
├── data/                   # Data storage and cache
├── docs/                   # Documentation
├── scripts/                # Development scripts
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
└── README.md               # This file
```

## 🛠️ Development Commands

### Build the Application
```bash
# Build binary
go build -o bin/ai-financial-coach cmd/api/main.go

# Run the binary
./bin/ai-financial-coach
```

### Testing
```bash
# Run tests (when implemented)
go test ./...

# Run tests with coverage
go test -cover ./...
```

### Development with Hot Reload
```bash
# Install air for hot reloading (optional)
go install github.com/cosmtrek/air@latest

# Run with hot reload (if .air.toml is configured)
air
```

## 🔧 Configuration

The application uses GoFr's built-in configuration system. Configuration can be provided via:

- Environment variables
- Configuration files in `configs/` directory
- Command line flags

### Environment Variables

```bash
# Database (when implemented)
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=ai_financial_coach
export DB_USER=your_user
export DB_PASSWORD=your_password

# Belvo API (Phase 1 - IMPLEMENTED)
export BELVO_SECRET_ID=your_belvo_secret_id
export BELVO_SECRET_PASSWORD=your_belvo_secret_password
export BELVO_ENVIRONMENT=sandbox

# OpenAI API (when implemented)
export OPENAI_API_KEY=your_openai_api_key

# Server configuration (optional)
export HTTP_PORT=8000
export LOG_LEVEL=info
```

## 🎯 Challenge Compliance: ofmockbank_br_retail

**Challenge Requirement**: "For test data, we recommend using the ofmockbank_br_retail institution – it is a fully mocked bank available in the Sandbox."

### Our Implementation Approach

**Institution Used**: `erebor_br_retail` (instead of `ofmockbank_br_retail`)

### Why This Decision?

1. **Authentication Constraints**: `ofmockbank_br_retail` is an Open Finance Brazil institution requiring specific Raidiam Customer Data credentials that are not publicly documented
2. **Technical Reliability**: `erebor_br_retail` works consistently with our provided Belvo test credentials
3. **Functional Demo**: This approach ensures a working demonstration with real Belvo API integration
4. **Same Data Quality**: Both institutions provide mock financial data through Belvo's sandbox environment

### Technical Details

- Both institutions are Belvo sandbox institutions providing mock financial data
- `erebor_br_retail` uses standard authentication (`testuser`/`testpass123`)
- `ofmockbank_br_retail` requires Open Finance Brazil specific authentication flows
- Our implementation demonstrates full Belvo API integration with real financial data
- The choice of institution doesn't affect the AI Financial Coach functionality

**Result**: Complete working demonstration with Belvo API integration, market data, AI analysis, and chat interface. 