package app

import (
	"fmt"
	"os"

	"gofr.dev/pkg/gofr"

	"ai-financial-coach/internal/api"
)

func CreateApp() *gofr.App {
	app := gofr.New()
	return app
}

func RegisterRoutes(app *gofr.App) {
	// Health check endpoint
	app.GET("/health", func(ctx *gofr.Context) (interface{}, error) {
		return map[string]interface{}{
			"status":  "healthy",
			"service": "ai-financial-coach",
			"version": "v1.0.0",
		}, nil
	})

	// API endpoints (to be implemented)
	app.GET("/", func(ctx *gofr.Context) (interface{}, error) {
		return map[string]interface{}{
			"message": "Welcome to AI Financial Coach API",
			"version": "v1.0.0",
			"endpoints": []string{
				"GET /health - Health check",
				"-- Belvo API (Phase 1) --",
				"GET /api/belvo/test-connection - Test Belvo API connection",
				"GET /api/belvo/institutions - Get available institutions",
				"POST /api/belvo/links - Create Belvo link",
				"GET /api/belvo/accounts/{link_id} - Get accounts for link",
				"GET /api/belvo/transactions/{link_id} - Get transactions for link",
				"GET /api/belvo/financial-summary/{link_id} - Get financial summary",
				"GET /api/belvo/mock-data - Get mock financial data for testing",
				"-- Market Data API (Phase 2) --",
				"GET /api/market/assets - Get available assets",
				"GET /api/market/assets/{symbol} - Get asset performance",
				"GET /api/market/data - Get market data summary",
				"GET /api/market/brazilian-rates - Get Brazilian economic rates",
				"GET /api/market/portfolio-templates - Get portfolio templates",
				"GET /api/market/portfolio-templates/{risk_level} - Get specific portfolio template",
				"-- AI Financial Coach (Phase 3) --",
				"POST /api/ai/analyze - Complete AI financial analysis",
				"GET /api/ai/portfolio-recommendation - Get AI portfolio recommendation",
				"POST /api/ai/what-if - Generate what-if scenarios",
				"GET /api/ai/quick-analysis/{link_id} - Quick AI analysis from Belvo data",
				"GET /api/ai/mock-analysis - AI analysis with mock data",
				"GET /api/ai/advice - General investment advice",
				"POST /api/ai/chat - Conversational AI chat",
			},
		}, nil
	})

	// Initialize Belvo handler with environment variables
	secretID := os.Getenv("BELVO_SECRET_ID")
	secretKey := os.Getenv("BELVO_SECRET_PASSWORD")
	environment := os.Getenv("BELVO_ENVIRONMENT")

	fmt.Printf("🔧 Environment Variables:\n")
	fmt.Printf("   BELVO_SECRET_ID: %s\n", func() string {
		if secretID != "" {
			return secretID[:8] + "..."
		} else {
			return "NOT SET"
		}
	}())
	fmt.Printf("   BELVO_SECRET_PASSWORD: %s\n", func() string {
		if secretKey != "" {
			return "***SET***"
		} else {
			return "NOT SET"
		}
	}())
	fmt.Printf("   BELVO_ENVIRONMENT: %s\n", func() string {
		if environment != "" {
			return environment
		} else {
			return "NOT SET (will default to sandbox)"
		}
	}())

	// Get test credentials from environment
	testSecretID := os.Getenv("BELVO_TEST_SECRET_ID")
	testSecretKey := os.Getenv("BELVO_TEST_SECRET_KEY")

	// Set defaults if not specified
	if environment == "" {
		environment = "sandbox"
	}
	if testSecretID == "" {
		testSecretID = "397581e3-22a5-4872-b11e-f12ff3c654b4"
	}
	if testSecretKey == "" {
		testSecretKey = "ifN@BQCu9s3xaad38j_*rNj@IbWEIK7LoAWXlH-pxhiPcOZfvKWKbivBeDlAv0k1"
	}

	// If primary credentials are not provided, fall back to test credentials
	if secretID == "" {
		secretID = testSecretID
	}
	if secretKey == "" {
		secretKey = testSecretKey
	}

	belvoHandler := api.NewBelvoHandler(secretID, secretKey, environment)

	// Initialize Market handler
	marketHandler := api.NewMarketHandler()

	// Initialize AI handler with OpenAI API key (optional)
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	if openAIAPIKey != "" {
		fmt.Printf("✅ OpenAI API key loaded: %s...\n", openAIAPIKey[:20])
	} else {
		fmt.Println("❌ OpenAI API key not found in environment")
	}
	aiHandler := api.NewAIHandler(openAIAPIKey, belvoHandler.GetBelvoService(), marketHandler.GetMarketService())

	// Set test credentials for belvo handler
	belvoHandler.SetTestCredentials(testSecretID, testSecretKey)

	// Core Belvo API routes
	app.POST("/api/belvo/test-connection", belvoHandler.TestConnection)
	app.POST("/api/belvo/create-erebor-link", belvoHandler.CreateEreborLink)
	app.POST("/api/belvo/links/for-selection", belvoHandler.GetLinksForSelection)
	app.POST("/api/belvo/links/detailed-info/{link_id}", belvoHandler.GetDetailedLinkInfo)

	// Development/debugging routes
	app.POST("/api/belvo/verify-data/{link_id}", belvoHandler.VerifyLinkData)

	// Market Data API routes (Phase 2)
	app.GET("/api/market/assets", marketHandler.GetAssets)
	app.GET("/api/market/assets/{symbol}", marketHandler.GetAssetPerformance)
	app.GET("/api/market/data", marketHandler.GetMarketData)
	app.GET("/api/market/brazilian-rates", marketHandler.GetBrazilianRates)
	app.GET("/api/market/portfolio-templates", marketHandler.GetPortfolioTemplates)
	app.GET("/api/market/portfolio-templates/{risk_level}", marketHandler.GetPortfolioTemplate)

	// AI Financial Coach API routes
	app.POST("/api/ai/chat", aiHandler.Chat)
	app.POST("/api/ai/cache-context", aiHandler.CacheContextFromSummary)
}
