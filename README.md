# 🤖 AI Financial Coach

An intelligent financial coaching platform that provides personalized financial analysis and investment recommendations using real banking data through Belvo API integration and OpenAI GPT-4.

**Live Demo**: https://belvo-ai-financial-coach.vercel.app

## Quick Test

For immediate testing, load the "Test Credentials" on auth screen.

## Features

- **Real Banking Data Integration**: Connects to real financial institutions via Belvo API
- **AI-Powered Financial Analysis**: Advanced GPT-4 analysis of spending patterns and financial health
- **Interactive Chat Interface**: Natural language conversations about your finances
- **Investment Recommendations**: Personalized portfolio suggestions based on your financial profile
- **Real-time Market Data**: Live market data from Yahoo Finance, CoinGecko, and Brazilian Central Bank
- **Multi-Account Support**: Analyze multiple bank accounts and customer links

## AI Financial Coach Capabilities

The AI coach provides comprehensive financial guidance by analyzing:
- **Transaction History**: Complete spending analysis with categorization
- **Account Balances**: Real-time account information across multiple institutions
- **Income vs Expenses**: Monthly financial flow analysis
- **Investment Opportunities**: Market-aware investment suggestions
- **Risk Assessment**: Personalized risk profiling for investment recommendations
- **Budget Optimization**: Actionable insights for expense management

## User Workflow

1. **Authentication**: User provides Belvo credentials for secure access
2. **Link Selection**: System retrieves all available customer links from connected institutions
3. **Data Analysis**: User selects a specific customer link for detailed financial analysis
4. **AI Coaching**: Interactive chat interface provides personalized financial guidance
5. **Investment Recommendations**: AI suggests portfolio allocations based on financial profile

## Banking Integration

**Primary Institution**: `erebor_br_retail` (Belvo Sandbox)
- Chosen for reliability and direct API compatibility
- Provides comprehensive financial data including accounts, transactions, and owner information

**Why not `ofmockbank_br_retail`?**
- Requires Open Finance Brazil authentication flow
- Needs Belvo Connect Widget for consent management
- Limited direct API access

## Technology Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: GoFr v1.43.0
- **APIs**: Belvo (Banking), OpenAI (AI), Yahoo Finance (Market Data), CoinGecko (Crypto)

### Frontend
- **Framework**: Next.js 15.4.6
- **React**: 19.1.0
- **Styling**: Tailwind CSS 3.4.17

### Infrastructure
- **Backend Deployment**: Railway
- **Frontend Deployment**: Vercel
- **Domain**: Custom subdomain (belvo.filipegarcia.co)

## Architecture

```
Frontend (Next.js) → Backend API (Go) → External APIs
                                    ├── Belvo API (Banking Data)
                                    ├── OpenAI API (AI Analysis)
                                    ├── Yahoo Finance (Market Data)
                                    └── CoinGecko (Crypto Data)
```

## Automated Deployment

- **CI/CD**: Native Vercel/Railway integrations with GitHub
- **Backend**: Auto-deploys to Railway on main branch commits
- **Frontend**: Auto-deploys to Vercel with custom domain configuration
- **Environment**: Production environment variables managed through secrets

## Security

- **No Data Persistence**: Financial data is not stored permanently
- **Secure Transmission**: All API communications over HTTPS
- **Session Management**: Client-side session storage only
- **Credential Handling**: Secure credential transmission without server-side storage
