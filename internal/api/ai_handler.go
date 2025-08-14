package api

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"gofr.dev/pkg/gofr"

	"ai-financial-coach/internal/models"
	"ai-financial-coach/internal/service"
)

// ContextCache stores pre-loaded financial contexts to avoid re-fetching
type ContextCache struct {
	mu       sync.RWMutex
	contexts map[string]*CachedContext
}

type CachedContext struct {
	Summary   *models.FinancialSummary
	LinkID    string
	OwnerName string
	CachedAt  time.Time
	ExpiresAt time.Time
}

// AIHandler handles HTTP requests related to AI financial coaching
type AIHandler struct {
	aiService     *service.AIService
	belvoService  *service.BelvoService
	marketService *service.MarketService
	contextCache  *ContextCache
}

// NewAIHandler creates a new AIHandler instance
func NewAIHandler(openAIAPIKey string, belvoService *service.BelvoService, marketService *service.MarketService) *AIHandler {
	return &AIHandler{
		aiService:     service.NewAIService(openAIAPIKey, marketService, belvoService),
		belvoService:  belvoService,
		marketService: marketService,
		contextCache: &ContextCache{
			contexts: make(map[string]*CachedContext),
		},
	}
}

// StoreContext caches financial context for a link
func (ah *AIHandler) StoreContext(linkID string, summary *models.FinancialSummary, ownerName string) {
	ah.contextCache.mu.Lock()
	defer ah.contextCache.mu.Unlock()

	sessionKey := linkID
	ah.contextCache.contexts[sessionKey] = &CachedContext{
		Summary:   summary,
		LinkID:    linkID,
		OwnerName: ownerName,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
}

// GetCachedContext retrieves cached financial context
func (ah *AIHandler) GetCachedContext(linkID string) (*models.FinancialSummary, bool) {
	ah.contextCache.mu.RLock()
	defer ah.contextCache.mu.RUnlock()

	sessionKey := linkID
	cached, exists := ah.contextCache.contexts[sessionKey]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(cached.ExpiresAt) {
		delete(ah.contextCache.contexts, sessionKey)
		return nil, false
	}
	return cached.Summary, true
}

// CacheContextFromSummary handles POST /api/ai/cache-context - stores financial context
func (ah *AIHandler) CacheContextFromSummary(ctx *gofr.Context) (interface{}, error) {
	var request struct {
		LinkID    string                   `json:"link_id"`
		OwnerName string                   `json:"owner_name"`
		Summary   *models.FinancialSummary `json:"financial_summary"`
	}

	if err := ctx.Bind(&request); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}

	if request.LinkID == "" || request.Summary == nil {
		return nil, fmt.Errorf("link_id and financial_summary are required")
	}

	// Store the context
	ah.StoreContext(request.LinkID, request.Summary, request.OwnerName)

	return map[string]interface{}{
		"message": fmt.Sprintf("Financial context cached for link: %s", request.LinkID[:8]),
		"status":  "success",
	}, nil
}

// AnalyzeFinancialProfile handles POST /api/ai/analyze
func (ah *AIHandler) AnalyzeFinancialProfile(ctx *gofr.Context) (interface{}, error) {
	var request models.AIAnalysisRequest
	if err := ctx.Bind(&request); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}

	// Validate required fields
	if request.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if request.RiskProfile == "" {
		request.RiskProfile = "balanced" // Default
	}
	if request.InvestmentHorizon == 0 {
		request.InvestmentHorizon = 5 // Default 5 years
	}
	if request.Language == "" {
		request.Language = "pt-BR" // Default Portuguese
	}

	// Perform AI analysis
	analysis, err := ah.aiService.AnalyzeFinancialProfile(&request)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze financial profile: %w", err)
	}

	return map[string]interface{}{
		"ai_analysis": analysis,
		"message":     "Análise financeira gerada com sucesso",
	}, nil
}

// GetPortfolioRecommendation handles GET /api/ai/portfolio-recommendation
func (ah *AIHandler) GetPortfolioRecommendation(ctx *gofr.Context) (interface{}, error) {
	// Get query parameters
	riskProfile := ctx.Param("risk_profile")
	if riskProfile == "" {
		riskProfile = "balanced"
	}

	monthlyBudgetStr := ctx.Param("monthly_budget")
	monthlyBudget := 0.0
	if monthlyBudgetStr != "" {
		var err error
		monthlyBudget, err = strconv.ParseFloat(monthlyBudgetStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid monthly_budget parameter")
		}
	}

	// Create a basic request for portfolio recommendation
	mockSummary := ah.createMockFinancialSummary(monthlyBudget)
	marketData, err := ah.marketService.GetMarketDataSummary()
	if err != nil {
		return nil, fmt.Errorf("failed to get market data: %w", err)
	}

	request := &models.AIAnalysisRequest{
		UserID:            "demo-user",
		FinancialSummary:  mockSummary,
		MarketData:        marketData,
		RiskProfile:       riskProfile,
		InvestmentHorizon: 5,
		MonthlyBudget:     monthlyBudget,
		Language:          "pt-BR",
	}

	analysis, err := ah.aiService.AnalyzeFinancialProfile(request)
	if err != nil {
		return nil, fmt.Errorf("failed to generate portfolio recommendation: %w", err)
	}

	return map[string]interface{}{
		"portfolio_recommendation": analysis.RecommendedPortfolio,
		"projections":              analysis.Projections,
		"risk_assessment":          analysis.RiskAssessment,
		"message":                  "Recomendação de portfolio gerada com sucesso",
	}, nil
}

// GenerateWhatIf handles POST /api/ai/what-if
func (ah *AIHandler) GenerateWhatIf(ctx *gofr.Context) (interface{}, error) {
	var request struct {
		BaseRequest    models.AIAnalysisRequest  `json:"base_request"`
		ScenarioParams models.ScenarioParameters `json:"scenario_params"`
	}

	if err := ctx.Bind(&request); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}

	scenario, err := ah.aiService.GenerateWhatIfScenario(&request.BaseRequest, &request.ScenarioParams)
	if err != nil {
		return nil, fmt.Errorf("failed to generate what-if scenario: %w", err)
	}

	return map[string]interface{}{
		"what_if_scenario": scenario,
		"message":          "Cenário what-if gerado com sucesso",
	}, nil
}

// GetQuickAnalysis handles GET /api/ai/quick-analysis/{link_id}
func (ah *AIHandler) GetQuickAnalysis(ctx *gofr.Context) (interface{}, error) {
	linkID := ctx.PathParam("link_id")
	if linkID == "" {
		return nil, fmt.Errorf("link_id parameter is required")
	}

	// Get financial data from Belvo
	financialSummary, err := ah.belvoService.GetFinancialSummary(linkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get financial summary: %w", err)
	}

	// Get market data
	marketData, err := ah.marketService.GetMarketDataSummary()
	if err != nil {
		return nil, fmt.Errorf("failed to get market data: %w", err)
	}

	// Create AI analysis request
	request := &models.AIAnalysisRequest{
		UserID:            "belvo-user-" + linkID,
		FinancialSummary:  financialSummary,
		MarketData:        marketData,
		RiskProfile:       ah.determineRiskProfile(financialSummary),
		InvestmentHorizon: 5,
		MonthlyBudget:     financialSummary.MonthlySurplus * 0.8, // 80% of surplus
		Language:          "pt-BR",
	}

	// Perform AI analysis
	analysis, err := ah.aiService.AnalyzeFinancialProfile(request)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze financial profile: %w", err)
	}

	return map[string]interface{}{
		"quick_analysis":    analysis,
		"link_id":           linkID,
		"financial_summary": financialSummary,
		"message":           "Análise rápida baseada nos dados do Belvo",
	}, nil
}

// GetMockAnalysis handles GET /api/ai/mock-analysis
func (ah *AIHandler) GetMockAnalysis(ctx *gofr.Context) (interface{}, error) {
	// Get query parameters for customization
	riskProfile := ctx.Param("risk_profile")
	if riskProfile == "" {
		riskProfile = "balanced"
	}

	language := ctx.Param("language")
	if language == "" {
		language = "pt-BR" // Default to Portuguese for existing behavior
	}
	// Convert to internal format
	if language == "en" {
		language = "en-US"
	} else if language == "pt" {
		language = "pt-BR"
	}

	credentialMode := ctx.Param("credential_mode")
	if credentialMode == "" {
		credentialMode = "demo" // Default to demo mode
	}

	linkID := ctx.Param("link_id")

	monthlyIncomeStr := ctx.Param("monthly_income")
	monthlyIncome := 8500.0 // Default
	if monthlyIncomeStr != "" {
		if parsed, err := strconv.ParseFloat(monthlyIncomeStr, 64); err == nil {
			monthlyIncome = parsed
		}
	}

	var mockSummary *models.FinancialSummary
	var dataSource string

	// Determine data source based on credential mode
	switch credentialMode {
	case "demo":
		// Use mock financial data
		mockSummary = ah.createMockFinancialSummaryWithIncome(monthlyIncome)
		dataSource = "Mock Data (Demo Mode)"
	case "test", "custom":
		if linkID != "" {
			// Try to get real Belvo data
			realSummary, err := ah.belvoService.GetFinancialSummary(linkID)
			if err != nil {
				// Fallback to mock data if Belvo fails
				mockSummary = ah.createMockFinancialSummaryWithIncome(monthlyIncome)
				dataSource = "Mock Data (Belvo Fallback)"
			} else {
				mockSummary = realSummary
				dataSource = fmt.Sprintf("Real Belvo Data (Link: %s)", linkID)
			}
		} else {
			// No link ID provided, use mock data
			mockSummary = ah.createMockFinancialSummaryWithIncome(monthlyIncome)
			dataSource = "Mock Data (No Link)"
		}
	default:
		// Default to mock data
		mockSummary = ah.createMockFinancialSummaryWithIncome(monthlyIncome)
		dataSource = "Mock Data (Default)"
	}

	// Get real market data
	marketData, err := ah.marketService.GetMarketDataSummary()
	if err != nil {
		return nil, fmt.Errorf("failed to get market data: %w", err)
	}

	// Create AI analysis request
	request := &models.AIAnalysisRequest{
		UserID:            "mock-user-demo",
		FinancialSummary:  mockSummary,
		MarketData:        marketData,
		RiskProfile:       riskProfile,
		InvestmentHorizon: 5,
		MonthlyBudget:     mockSummary.MonthlySurplus * 0.8,
		Language:          language,
		Goals: []models.InvestmentGoal{
			{
				Type:         "retirement",
				Description:  getLocalizedGoalDescription("retirement", language),
				TargetAmount: 1000000,
				TimeHorizon:  25,
				Priority:     "high",
			},
			{
				Type:         "emergency",
				Description:  getLocalizedGoalDescription("emergency", language),
				TargetAmount: 30000,
				TimeHorizon:  1,
				Priority:     "high",
			},
		},
	}

	// Perform AI analysis
	analysis, err := ah.aiService.AnalyzeFinancialProfile(request)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze mock profile: %w", err)
	}

	return map[string]interface{}{
		"mock_analysis": analysis,
		"parameters": map[string]interface{}{
			"risk_profile":    riskProfile,
			"monthly_income":  monthlyIncome,
			"language":        language,
			"credential_mode": credentialMode,
			"link_id":         linkID,
		},
		"data_source": dataSource,
		"message":     getLocalizedMessage("mock_analysis_success", language),
		"note":        getLocalizedMessage("mock_analysis_note", language),
	}, nil
}

// Helper function to get localized goal descriptions
func getLocalizedGoalDescription(goalType, language string) string {
	if language == "en-US" {
		switch goalType {
		case "retirement":
			return "Peaceful retirement"
		case "emergency":
			return "Emergency fund"
		default:
			return goalType
		}
	}
	// Default Portuguese
	switch goalType {
	case "retirement":
		return "Aposentadoria tranquila"
	case "emergency":
		return "Reserva de emergência"
	default:
		return goalType
	}
}

// Helper function to get localized messages
func getLocalizedMessage(messageType, language string) string {
	if language == "en-US" {
		switch messageType {
		case "mock_analysis_success":
			return "Mock analysis with simulated data"
		case "mock_analysis_note":
			return "This is a demonstration with fictional data for testing"
		default:
			return messageType
		}
	}
	// Default Portuguese
	switch messageType {
	case "mock_analysis_success":
		return "Análise de demonstração com dados simulados"
	case "mock_analysis_note":
		return "Esta é uma demonstração com dados fictícios para testing"
	default:
		return messageType
	}
}

// GetInvestmentAdvice handles GET /api/ai/advice
func (ah *AIHandler) GetInvestmentAdvice(ctx *gofr.Context) (interface{}, error) {
	// Get market data for current opportunities
	marketData, err := ah.marketService.GetMarketDataSummary()
	if err != nil {
		return nil, fmt.Errorf("failed to get market data: %w", err)
	}

	// Create simple mock for analysis
	mockSummary := ah.createMockFinancialSummary(2000) // R$ 2000 monthly investment

	request := &models.AIAnalysisRequest{
		UserID:            "advice-seeker",
		FinancialSummary:  mockSummary,
		MarketData:        marketData,
		RiskProfile:       "balanced",
		InvestmentHorizon: 3,
		MonthlyBudget:     2000,
		Language:          "pt-BR",
	}

	analysis, err := ah.aiService.AnalyzeFinancialProfile(request)
	if err != nil {
		return nil, fmt.Errorf("failed to generate investment advice: %w", err)
	}

	return map[string]interface{}{
		"investment_advice": map[string]interface{}{
			"market_opportunities": analysis.Analysis.MarketOpportunities,
			"recommendations":      analysis.Recommendations,
			"portfolio_suggestion": analysis.RecommendedPortfolio,
			"ai_summary":           analysis.Summary,
		},
		"market_context": map[string]interface{}{
			"brazilian_rates":   marketData.BrazilianRates,
			"asset_performance": marketData.Assets,
		},
		"message": "Conselhos de investimento baseados no mercado atual",
	}, nil
}

// Chat handles POST /api/ai/chat for conversational AI
func (ah *AIHandler) Chat(ctx *gofr.Context) (interface{}, error) {
	var request models.ChatRequest
	if err := ctx.Bind(&request); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}

	// Set default language if not provided
	if request.Language == "" {
		request.Language = "en"
	}

	// Try to use cached context first, then fetch if needed
	if request.UserContext == nil {
		if (request.CredentialMode == "test" || request.CredentialMode == "custom") && request.LinkID != "" {
			// First, try to get cached context
			if cachedSummary, found := ah.GetCachedContext(request.LinkID); found {
				request.UserContext = cachedSummary
			} else {
				// Use dynamic Belvo service with user credentials if provided
				var belvoService *service.BelvoService
				if request.SecretID != "" && request.SecretKey != "" {
					belvoService = service.NewBelvoService(request.SecretID, request.SecretKey, "sandbox")
				} else {
					belvoService = ah.belvoService
				}

				if summary, err := belvoService.GetFinancialSummary(request.LinkID); err == nil {
					request.UserContext = summary
					ah.StoreContext(request.LinkID, summary, "Unknown Customer")
				} else {
					request.UserContext = ah.createMockFinancialSummaryWithIncome(8500.0)
				}
			}
		} else {
			request.UserContext = ah.createMockFinancialSummaryWithIncome(8500.0)
		}
	}

	// Get market context
	if request.MarketContext == nil {
		marketData, err := ah.marketService.GetMarketDataSummary()
		if err == nil {
			request.MarketContext = marketData
		}
	}

	// Call AI service for conversational response
	response, err := ah.aiService.Chat(&request)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AI chat response: %w", err)
	}

	return map[string]interface{}{
		"chat_response": response,
		"message":       "AI chat response generated successfully",
	}, nil
}

// createMockFinancialSummary creates a realistic mock financial summary
func (ah *AIHandler) createMockFinancialSummary(monthlyBudget float64) *models.FinancialSummary {
	return ah.createMockFinancialSummaryWithIncome(8500.0) // Default income
}

// createMockFinancialSummaryWithIncome creates mock data with specified income
func (ah *AIHandler) createMockFinancialSummaryWithIncome(monthlyIncome float64) *models.FinancialSummary {
	fixedExpenses := monthlyIncome * 0.35    // 35% fixed expenses
	variableExpenses := monthlyIncome * 0.35 // 35% variable expenses
	surplus := monthlyIncome - fixedExpenses - variableExpenses

	return &models.FinancialSummary{
		UserID:                  "mock-user",
		MonthlyIncome:           monthlyIncome,
		MonthlyFixedExpenses:    fixedExpenses,
		MonthlyVariableExpenses: variableExpenses,
		MonthlySurplus:          surplus,
		TotalBalance:            surplus * 6, // 6 months of surplus as emergency fund
		Currency:                "BRL",
	}
}

// determineRiskProfile automatically determines risk profile based on financial data
func (ah *AIHandler) determineRiskProfile(summary *models.FinancialSummary) string {
	savingsRate := summary.MonthlySurplus / summary.MonthlyIncome
	emergencyMonths := summary.TotalBalance / (summary.MonthlyFixedExpenses + summary.MonthlyVariableExpenses)

	// Conservative: Low savings rate or insufficient emergency fund
	if savingsRate < 0.15 || emergencyMonths < 3 {
		return "conservative"
	}

	// Aggressive: High savings rate and good emergency fund
	if savingsRate > 0.25 && emergencyMonths > 6 {
		return "aggressive"
	}

	// Balanced: Middle ground
	return "balanced"
}
