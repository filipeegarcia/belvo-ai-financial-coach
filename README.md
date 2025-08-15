# 🤖 AI Financial Coach

A comprehensive AI-powered financial coaching application that provides personalized financial analysis and recommendations using real banking data through Belvo API Sandbox integration.

## 🌟 Features

- **Real Banking Data Integration** via Belvo API (`erebor_br_retail`)
- **AI-Powered Financial Analysis** using OpenAI GPT-4
- **Interactive Chat Interface** for personalized financial coaching
- **Comprehensive Transaction Analysis** with detailed spending insights
- **Account Management** with multi-account support
- **Production-Ready Deployment** with auto-deployment pipeline

## 🚀 Live Demo

- **Production App**: `https://filipegarcia.co/belvo`

## 🏗️ Architecture

### Backend
- **Framework**: GoFr v1.35.0 (Go web framework)
- **AI Integration**: OpenAI GPT-4 with transaction context
- **Financial Data**: Belvo API (sandbox environment)
- **Authentication**: Belvo credentials-based auth
- **Caching**: In-memory context cache for AI responses

### Frontend
- **Framework**: Next.js 15 with React 19
- **Styling**: Tailwind CSS with Typography plugin
- **Markdown**: React Markdown for AI responses
- **State Management**: React useState with sessionStorage

### Deployment
- **Backend**: Railway (auto-deploy from GitHub)
- **Frontend**: Vercel (auto-deploy from GitHub)
- **CI/CD**: GitHub Actions
- **Domains**: Custom domain support

## 🛠️ Technology Stack

| Component | Technology | Version |
|-----------|------------|---------|
| Backend Language | Go | 1.21 |
| Web Framework | GoFr | v1.35.0 |
| Frontend Framework | Next.js | 15.4.6 |
| React | React | 19.1.0 |
| AI Provider | OpenAI | GPT-4 |
| Financial API | Belvo | Sandbox |
| Deployment | Railway + Vercel | - |
| Styling | Tailwind CSS | 3.4.17 |

## 📋 Prerequisites

- **Go 1.21+** installed
- **Node.js 18+** installed
- **Git** for version control
- **Belvo API credentials** (sandbox environment)
- **OpenAI API key**

## 🚀 Quick Start

### 1. Clone & Setup
```bash
git clone <repository-url>
cd poc

# Backend setup
go mod tidy

# Frontend setup
cd frontend
npm install
cd ..
```

### 2. Environment Configuration
```bash
# Backend Environment Variables
export BELVO_SECRET_ID=your-belvo-secret-id
export BELVO_SECRET_PASSWORD=your-belvo-secret-password
export OPENAI_API_KEY=your-openai-api-key
export BELVO_ENVIRONMENT=sandbox
```

### 3. Run Development Environment
```bash
# Terminal 1: Start Backend (port 8000)
go run cmd/api/main.go

# Terminal 2: Start Frontend (port 3000)
cd frontend
npm run dev
```

### 4. Access the Application
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8000
- **Health Check**: http://localhost:8000/health

## 🔄 User Flow

1. **Authentication**: Enter Belvo credentials
2. **Account Selection**: Choose from available customer accounts
3. **Data Loading**: Progressive loading of financial data (accounts, transactions, income)
4. **AI Chat**: Interactive financial coaching with full transaction context

## 📁 Project Structure

```
poc/
├── cmd/api/                 # Application entry point
│   └── main.go             # Main server file
├── internal/               # Private backend code
│   ├── app/               # Application setup
│   ├── api/               # HTTP handlers
│   │   ├── belvo_handler.go
│   │   └── ai_handler.go
│   ├── service/           # Business logic
│   │   ├── belvo_service.go
│   │   └── ai_service.go
│   └── models/            # Data models
├── frontend/              # Next.js frontend
│   ├── src/app/          # App router pages
│   │   ├── page.tsx      # Auth page
│   │   └── chat/page.tsx # Chat interface
│   └── package.json
├── .github/workflows/     # CI/CD pipelines
├── railway.json          # Railway deployment config
├── vercel.json           # Vercel deployment config
├── go.mod               # Go dependencies
└── README.md           # This file
```

## 🌐 API Endpoints

### Authentication & Connection
- `POST /api/belvo/test-connection` - Test Belvo credentials
- `POST /api/belvo/create-erebor-link` - Create banking connection

### Account Management  
- `POST /api/belvo/links/for-selection` - Get available accounts (fast)
- `POST /api/belvo/links/detailed-info/{link_id}` - Get detailed account data

### AI Financial Coach
- `POST /api/ai/chat` - Chat with AI financial coach
- `POST /api/ai/cache-context` - Cache financial context

### Utilities
- `GET /health` - Health check
- `POST /api/belvo/verify-data/{link_id}` - Verify data integrity

## 🎯 Belvo Integration Details

### Institution Used
**`erebor_br_retail`** - Chosen for reliability and compatibility

### Why Not `ofmockbank_br_retail`?
- Requires Open Finance Brazil specific authentication flow
- Needs Belvo Connect Widget for consent management  
- Not accessible via direct API calls
- Lacks proper documentation for direct integration

### Data Retrieved
- **Accounts**: Account details, balances, types
- **Transactions**: Last 3 months of transaction history
- **Owners**: Account holder information
- **Income**: Income stream analysis
- **Financial Summary**: Aggregated financial metrics

## 🤖 AI Capabilities

The AI Financial Coach has access to:
- **Complete transaction history** with descriptions, amounts, dates
- **Account balances** and types
- **Income and expense patterns**
- **Monthly financial summaries**
- **Spending categorization**

### Sample AI Interactions
- "Show me my last 10 transactions"
- "How much did I spend on credit cards this month?"
- "What's my monthly income vs expenses?"
- "Analyze my spending patterns"

## 🚀 Deployment

### Automatic Deployment
Every push to `main` branch triggers:
1. **Frontend deployment** to Vercel
2. **Backend deployment** to Railway
3. **End-to-end testing** via GitHub Actions

### Production URLs
- **Frontend**: `https://filipegarcia.co/belvo`
- **Backend**: `https://api.filipegarcia.co`

### Manual Deployment

#### Backend (Railway)
1. Connect GitHub repository to Railway
2. Set environment variables
3. Deploy with auto-detect Go buildpack

#### Frontend (Vercel)  
1. Connect GitHub repository to Vercel
2. Set build directory to `frontend`
3. Configure custom domain with `/belvo` path

## 🔧 Development

### Backend Development
```bash
# Run with auto-restart
go run cmd/api/main.go

# Build binary
go build -o bin/api cmd/api/main.go

# Test endpoints
curl http://localhost:8000/health
```

### Frontend Development
```bash
cd frontend

# Development server
npm run dev

# Build for production
npm run build

# Start production server
npm start
```

### Environment Variables

#### Backend (Required)
```bash
BELVO_SECRET_ID=your-belvo-secret-id
BELVO_SECRET_PASSWORD=your-belvo-secret-password  
OPENAI_API_KEY=your-openai-api-key
BELVO_ENVIRONMENT=sandbox
```

#### Frontend (Production)
```bash
NEXT_PUBLIC_API_URL=https://api.filipegarcia.co
NEXT_PUBLIC_ENVIRONMENT=production
```

## 🔍 Key Features Explained

### Progressive Data Loading
1. **Phase 1**: Instant display of basic account information
2. **Phase 2**: Detailed financial data loading after account selection
3. **Optimization**: Parallel API calls for maximum speed

### AI Context Management
- **Caching**: Financial data cached for 24 hours
- **Context Building**: Comprehensive transaction summaries for AI
- **Real-time**: AI responses with full financial context

### Error Handling
- **Graceful degradation** when Belvo API is unavailable
- **User-friendly error messages**
- **Retry mechanisms** for transient failures

## 🧪 Testing

### Backend Testing
```bash
# Test health endpoint
curl http://localhost:8000/health

# Test with real credentials
curl -X POST http://localhost:8000/api/belvo/test-connection \
  -H "Content-Type: application/json" \
  -d '{"secret_id":"your-id","secret_key":"your-key"}'
```

### Frontend Testing  
```bash
cd frontend
npm test
```

## 📈 Performance

- **Initial load**: ~2-3 seconds for account list
- **Detailed data**: ~15-30 seconds for comprehensive financial data
- **AI responses**: ~3-5 seconds with cached context
- **Caching**: 24-hour financial data cache for repeat interactions

## 🔒 Security

- **No data persistence**: Financial data not stored permanently
- **Credential handling**: Secure credential transmission
- **Session management**: Client-side session storage only
- **HTTPS**: All production traffic encrypted

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📄 License

This project is for demonstration purposes and educational use.

---

**Built with ❤️ using Go, Next.js, and AI**
