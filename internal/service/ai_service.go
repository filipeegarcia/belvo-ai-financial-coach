package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"ai-financial-coach/internal/models"
)

// AIService handles AI-powered financial analysis
type AIService struct {
	httpClient    *http.Client
	openAIAPIKey  string
	openAIBaseURL string
	model         string
	marketService *MarketService
	belvoService  *BelvoService
}

// NewAIService creates a new AIService instance
func NewAIService(openAIAPIKey string, marketService *MarketService, belvoService *BelvoService) *AIService {
	return &AIService{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		openAIAPIKey:  openAIAPIKey,
		openAIBaseURL: "https://api.openai.com/v1",
		model:         "gpt-4o-mini",
		marketService: marketService,
		belvoService:  belvoService,
	}
}

// AnalyzeFinancialProfile performs comprehensive AI analysis
func (ai *AIService) AnalyzeFinancialProfile(request *models.AIAnalysisRequest) (*models.AIAnalysisResponse, error) {
	// 1. Calculate portfolio recommendation
	portfolio, err := ai.calculatePortfolioRecommendation(request)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate portfolio recommendation: %w", err)
	}

	// 2. Generate projections
	projections, err := ai.calculateProjections(portfolio, request.InvestmentHorizon, request.MonthlyBudget)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate projections: %w", err)
	}

	// 3. Analyze financial data
	analysis, err := ai.analyzeFinancialData(request.FinancialSummary, request.MarketData)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze financial data: %w", err)
	}

	// 4. Generate AI recommendations
	recommendations, err := ai.generateRecommendations(request, portfolio, analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to generate recommendations: %w", err)
	}

	// 5. Assess risks
	riskAssessment, err := ai.assessRisks(portfolio, request.FinancialSummary, projections)
	if err != nil {
		return nil, fmt.Errorf("failed to assess risks: %w", err)
	}

	// 6. Generate AI summary
	summary, err := ai.generateAISummary(request, portfolio, projections, analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AI summary: %w", err)
	}

	return &models.AIAnalysisResponse{
		UserID:               request.UserID,
		GeneratedAt:          time.Now(),
		RiskProfile:          request.RiskProfile,
		RecommendedPortfolio: *portfolio,
		Projections:          *projections,
		Analysis:             *analysis,
		Recommendations:      recommendations,
		RiskAssessment:       *riskAssessment,
		Summary:              summary,
		Language:             request.Language,
	}, nil
}

// calculatePortfolioRecommendation determines the best portfolio for the user
func (ai *AIService) calculatePortfolioRecommendation(request *models.AIAnalysisRequest) (*models.PortfolioRecommendation, error) {
	// Select base template based on risk profile
	var template models.PortfolioTemplate
	for _, t := range models.DefaultPortfolioTemplates {
		if t.RiskLevel == request.RiskProfile {
			template = t
			break
		}
	}

	// Localize template name
	template.Name = ai.localizeTemplateName(template.Name, request.Language)

	// Calculate safe monthly investment amount
	surplus := request.FinancialSummary.MonthlySurplus
	safeInvestmentPercentage := 0.8 // Invest 80% of surplus by default

	switch request.RiskProfile {
	case "conservative":
		safeInvestmentPercentage = 0.6
	case "balanced":
		safeInvestmentPercentage = 0.8
	case "aggressive":
		safeInvestmentPercentage = 0.9
	}

	monthlyInvestment := surplus * safeInvestmentPercentage
	if request.MonthlyBudget > 0 && request.MonthlyBudget < monthlyInvestment {
		monthlyInvestment = request.MonthlyBudget
	}

	// Calculate expected return based on current market data
	expectedReturn := ai.calculateBlendedReturn(template, request.MarketData)

	rationale := ai.localizeRationale(request.RiskProfile, request.Language)

	return &models.PortfolioRecommendation{
		Template:          template,
		CustomAllocations: template.Allocations,
		ExpectedReturn:    expectedReturn,
		ExpectedRisk:      template.MaxDrawdown,
		MonthlyInvestment: monthlyInvestment,
		Rationale:         rationale,
	}, nil
}

// localizeTemplateName translates portfolio template names
func (ai *AIService) localizeTemplateName(name, language string) string {
	if language == "en-US" {
		switch name {
		case "Conservador":
			return "Conservative"
		case "Balanceado":
			return "Balanced"
		case "Agressivo":
			return "Aggressive"
		default:
			return name
		}
	}
	// Default Portuguese
	switch name {
	case "Conservative":
		return "Conservador"
	case "Balanced":
		return "Balanceado"
	case "Aggressive":
		return "Agressivo"
	default:
		return name
	}
}

// localizeRationale provides localized rationale text
func (ai *AIService) localizeRationale(riskProfile, language string) string {
	if language == "en-US" {
		return fmt.Sprintf("Based on your %s profile and current financial situation", riskProfile)
	}
	return fmt.Sprintf("Baseado no seu perfil %s e situação financeira atual", riskProfile)
}

// calculateBlendedReturn calculates expected portfolio return based on market data
func (ai *AIService) calculateBlendedReturn(template models.PortfolioTemplate, marketData *models.MarketDataSummary) float64 {
	var weightedReturn float64

	for _, allocation := range template.Allocations {
		var assetReturn float64

		// Find asset performance in market data
		for _, asset := range marketData.Assets {
			if asset.Symbol == allocation.Symbol {
				assetReturn = asset.AnnualizedReturn / 100 // Convert percentage to decimal
				break
			}
		}

		// If no market data, use template expected return
		if assetReturn == 0 {
			assetReturn = template.ExpectedReturn
		}

		weightedReturn += allocation.Percentage * assetReturn
	}

	return weightedReturn
}

// calculateProjections generates future value projections
func (ai *AIService) calculateProjections(portfolio *models.PortfolioRecommendation, years int, monthlyContribution float64) (*models.PortfolioProjection, error) {
	if monthlyContribution == 0 {
		monthlyContribution = portfolio.MonthlyInvestment
	}

	monthlyRate := portfolio.ExpectedReturn / 12
	totalMonths := years * 12

	var projections []models.ProjectionPoint
	var totalContributed, totalValue float64

	for month := 1; month <= totalMonths; month++ {
		totalContributed += monthlyContribution

		// Future Value of Annuity: FV = P * (((1+r)^n - 1) / r)
		if monthlyRate > 0 {
			totalValue = monthlyContribution * ((math.Pow(1+monthlyRate, float64(month)) - 1) / monthlyRate)
		} else {
			totalValue = totalContributed
		}

		gains := totalValue - totalContributed
		monthlyReturn := 0.0
		if month > 1 {
			monthlyReturn = monthlyRate * 100
		}

		projections = append(projections, models.ProjectionPoint{
			Month:            month,
			Year:             (month-1)/12 + 1,
			TotalValue:       totalValue,
			TotalContributed: totalContributed,
			TotalGains:       gains,
			MonthlyReturn:    monthlyReturn,
		})
	}

	return &models.PortfolioProjection{
		Years:               years,
		MonthlyContribution: monthlyContribution,
		InitialAmount:       0,
		ExpectedReturn:      portfolio.ExpectedReturn,
		Allocations:         portfolio.CustomAllocations,
		Projections:         projections,
		TotalFinalValue:     totalValue,
		TotalContributed:    totalContributed,
		TotalGains:          totalValue - totalContributed,
		Currency:            "BRL",
	}, nil
}

// analyzeFinancialData performs detailed financial analysis
func (ai *AIService) analyzeFinancialData(summary *models.FinancialSummary, marketData *models.MarketDataSummary) (*models.FinancialAnalysis, error) {
	// Calculate financial health score
	healthScore := ai.calculateFinancialHealthScore(summary)

	// Analyze surplus
	surplusAnalysis := ai.analyzeSurplus(summary)

	// Analyze spending patterns
	spendingPatterns := ai.analyzeSpendingPatterns(summary)

	// Assess investment readiness
	readiness := ai.assessInvestmentReadiness(summary, healthScore)

	// Identify market opportunities
	opportunities := ai.identifyMarketOpportunities(marketData)

	return &models.FinancialAnalysis{
		SurplusAnalysis:      *surplusAnalysis,
		SpendingPatterns:     *spendingPatterns,
		InvestmentReadiness:  *readiness,
		MarketOpportunities:  opportunities,
		FinancialHealthScore: healthScore,
	}, nil
}

// calculateFinancialHealthScore calculates a 0-100 financial health score
func (ai *AIService) calculateFinancialHealthScore(summary *models.FinancialSummary) float64 {
	score := 0.0

	// Income stability (25 points)
	if summary.MonthlyIncome > 0 {
		score += 25
	}

	// Expense management (25 points)
	expenseRatio := (summary.MonthlyFixedExpenses + summary.MonthlyVariableExpenses) / summary.MonthlyIncome
	if expenseRatio < 0.7 {
		score += 25
	} else if expenseRatio < 0.9 {
		score += 15
	} else {
		score += 5
	}

	// Savings rate (25 points)
	savingsRate := summary.MonthlySurplus / summary.MonthlyIncome
	if savingsRate > 0.2 {
		score += 25
	} else if savingsRate > 0.1 {
		score += 15
	} else if savingsRate > 0 {
		score += 10
	}

	// Emergency fund (25 points)
	monthlyExpenses := summary.MonthlyFixedExpenses + summary.MonthlyVariableExpenses
	emergencyMonths := summary.TotalBalance / monthlyExpenses
	if emergencyMonths >= 6 {
		score += 25
	} else if emergencyMonths >= 3 {
		score += 15
	} else if emergencyMonths >= 1 {
		score += 10
	}

	return math.Min(score, 100)
}

// analyzeSurplus analyzes available investment capacity
func (ai *AIService) analyzeSurplus(summary *models.FinancialSummary) *models.SurplusAnalysis {
	monthlyExpenses := summary.MonthlyFixedExpenses + summary.MonthlyVariableExpenses
	emergencyTarget := monthlyExpenses * 3 // 3 months minimum

	// Conservative approach: invest only after emergency fund
	availableForInvestment := summary.MonthlySurplus
	if summary.TotalBalance < emergencyTarget {
		availableForInvestment = math.Max(0, summary.MonthlySurplus*0.5) // 50% until emergency fund is built
	}

	return &models.SurplusAnalysis{
		MonthlySurplus:         summary.MonthlySurplus,
		RecommendedInvestment:  0.8, // 80% of surplus
		EmergencyFundTarget:    emergencyTarget,
		CurrentEmergencyFund:   summary.TotalBalance,
		SafeInvestmentAmount:   availableForInvestment,
		ConservativePercentage: 0.6,
	}
}

// analyzeSpendingPatterns analyzes user spending behavior
func (ai *AIService) analyzeSpendingPatterns(summary *models.FinancialSummary) *models.SpendingPatterns {
	totalExpenses := summary.MonthlyFixedExpenses + summary.MonthlyVariableExpenses

	return &models.SpendingPatterns{
		FixedExpenseRatio:    summary.MonthlyFixedExpenses / summary.MonthlyIncome,
		VariableExpenseRatio: summary.MonthlyVariableExpenses / summary.MonthlyIncome,
		SavingsRate:          summary.MonthlySurplus / summary.MonthlyIncome,
		TopCategories: []models.ExpenseCategory{
			{Category: "Gastos Fixos", Amount: summary.MonthlyFixedExpenses, Percentage: summary.MonthlyFixedExpenses / totalExpenses, Trend: "stable"},
			{Category: "Gastos Variáveis", Amount: summary.MonthlyVariableExpenses, Percentage: summary.MonthlyVariableExpenses / totalExpenses, Trend: "stable"},
		},
		OptimizationSuggestions: []string{
			"Revise gastos variáveis mensalmente",
			"Considere renegociar contratos fixos",
			"Automatize investimentos para facilitar poupança",
		},
	}
}

// assessInvestmentReadiness evaluates how ready the user is to invest
func (ai *AIService) assessInvestmentReadiness(summary *models.FinancialSummary, healthScore float64) *models.InvestmentReadiness {
	score := healthScore
	level := "not_ready"

	if score >= 80 {
		level = "very_ready"
	} else if score >= 60 {
		level = "ready"
	} else if score >= 40 {
		level = "somewhat_ready"
	}

	return &models.InvestmentReadiness{
		Score:          score,
		ReadinessLevel: level,
		KeyFactors: []string{
			fmt.Sprintf("Renda mensal: R$ %.2f", summary.MonthlyIncome),
			fmt.Sprintf("Sobra mensal: R$ %.2f", summary.MonthlySurplus),
			fmt.Sprintf("Reserva atual: R$ %.2f", summary.TotalBalance),
		},
		ImprovementAreas: []string{
			"Construir reserva de emergência",
			"Reduzir gastos desnecessários",
			"Aumentar renda",
		},
		RecommendedStrategy: "Início gradual com valores baixos",
	}
}

// identifyMarketOpportunities analyzes current market conditions
func (ai *AIService) identifyMarketOpportunities(marketData *models.MarketDataSummary) []models.MarketOpportunity {
	var opportunities []models.MarketOpportunity

	for _, asset := range marketData.Assets {
		opportunity := "hold"
		confidence := 0.5

		// Simple analysis based on returns
		if asset.AnnualizedReturn > 15 {
			opportunity = "buy"
			confidence = 0.7
		} else if asset.AnnualizedReturn < -10 {
			opportunity = "avoid"
			confidence = 0.8
		}

		opportunities = append(opportunities, models.MarketOpportunity{
			AssetClass:  string(asset.Type),
			Symbol:      asset.Symbol,
			Opportunity: opportunity,
			Rationale:   fmt.Sprintf("Retorno anualizado de %.2f%%", asset.AnnualizedReturn),
			Confidence:  confidence,
			TimeHorizon: "medium",
		})
	}

	return opportunities
}

// generateRecommendations creates actionable recommendations
func (ai *AIService) generateRecommendations(request *models.AIAnalysisRequest, portfolio *models.PortfolioRecommendation, analysis *models.FinancialAnalysis) ([]models.ActionRecommendation, error) {
	var recommendations []models.ActionRecommendation
	language := request.Language

	// Emergency fund recommendation
	if analysis.SurplusAnalysis.CurrentEmergencyFund < analysis.SurplusAnalysis.EmergencyFundTarget {
		recommendations = append(recommendations, models.ActionRecommendation{
			Priority:    "immediate",
			Action:      ai.localizeText("build_emergency_fund", language),
			Description: ai.localizeDescription("emergency_fund_desc", language, analysis.SurplusAnalysis.EmergencyFundTarget),
			Impact:      "high",
			Effort:      "moderate",
			Timeline:    ai.localizeText("3_6_months", language),
		})
	}

	// Investment start recommendation
	if analysis.InvestmentReadiness.ReadinessLevel != "not_ready" {
		recommendations = append(recommendations, models.ActionRecommendation{
			Priority:    "short_term",
			Action:      ai.localizeText("start_investments", language),
			Description: ai.localizeDescription("start_investments_desc", language, portfolio.MonthlyInvestment, request.RiskProfile),
			Impact:      "high",
			Effort:      "easy",
			Timeline:    ai.localizeText("1_month", language),
		})
	}

	// Portfolio diversification
	recommendations = append(recommendations, models.ActionRecommendation{
		Priority:    "medium_term",
		Action:      ai.localizeText("diversify_portfolio", language),
		Description: ai.localizeText("diversify_portfolio_desc", language),
		Impact:      "medium",
		Effort:      "moderate",
		Timeline:    ai.localizeText("6_months", language),
	})

	return recommendations, nil
}

// localizeText provides translation for common texts
func (ai *AIService) localizeText(key, language string) string {
	if language == "en-US" {
		switch key {
		case "build_emergency_fund":
			return "Build Emergency Fund"
		case "start_investments":
			return "Start Investments"
		case "diversify_portfolio":
			return "Diversify Portfolio"
		case "diversify_portfolio_desc":
			return "Implement the recommended allocation gradually"
		case "3_6_months":
			return "3-6 months"
		case "1_month":
			return "1 month"
		case "6_months":
			return "6 months"
		default:
			return key
		}
	}
	// Default Portuguese
	switch key {
	case "build_emergency_fund":
		return "Construir Reserva de Emergência"
	case "start_investments":
		return "Iniciar Investimentos"
	case "diversify_portfolio":
		return "Diversificar Portfolio"
	case "diversify_portfolio_desc":
		return "Implemente a alocação recomendada gradualmente"
	case "3_6_months":
		return "3-6 meses"
	case "1_month":
		return "1 mês"
	case "6_months":
		return "6 meses"
	default:
		return key
	}
}

// localizeDescription provides localized descriptions with parameters
func (ai *AIService) localizeDescription(key, language string, params ...interface{}) string {
	if language == "en-US" {
		switch key {
		case "emergency_fund_desc":
			if len(params) > 0 {
				return fmt.Sprintf("Accumulate $%.2f to have 3-6 months of expenses", params[0])
			}
			return "Accumulate emergency fund for 3-6 months of expenses"
		case "start_investments_desc":
			if len(params) >= 2 {
				return fmt.Sprintf("Start investing $%.2f monthly with %s profile", params[0], params[1])
			}
			return "Start investing monthly with recommended profile"
		default:
			return key
		}
	}
	// Default Portuguese
	switch key {
	case "emergency_fund_desc":
		if len(params) > 0 {
			return fmt.Sprintf("Acumule R$ %.2f para ter 3-6 meses de gastos", params[0])
		}
		return "Acumule reserva de emergência para 3-6 meses de gastos"
	case "start_investments_desc":
		if len(params) >= 2 {
			return fmt.Sprintf("Comece investindo R$ %.2f mensalmente no perfil %s", params[0], params[1])
		}
		return "Comece investindo mensalmente no perfil recomendado"
	default:
		return key
	}
}

// assessRisks evaluates investment risks
func (ai *AIService) assessRisks(portfolio *models.PortfolioRecommendation, summary *models.FinancialSummary, projections *models.PortfolioProjection) (*models.RiskAssessment, error) {
	overallRisk := "medium"
	if portfolio.ExpectedRisk < 0.1 {
		overallRisk = "low"
	} else if portfolio.ExpectedRisk > 0.2 {
		overallRisk = "high"
	}

	potentialLoss := projections.TotalFinalValue * portfolio.ExpectedRisk

	return &models.RiskAssessment{
		OverallRisk: overallRisk,
		RiskFactors: []models.RiskFactor{
			{Type: "market", Description: "Volatilidade do mercado", Severity: "medium", Probability: 0.7},
			{Type: "inflation", Description: "Risco de inflação", Severity: "medium", Probability: 0.5},
		},
		MitigationSteps: []string{
			"Diversificação de ativos",
			"Investimento regular (dollar-cost averaging)",
			"Revisão periódica da carteira",
		},
		RiskTolerance: portfolio.Template.RiskLevel,
		WorstCaseScenario: models.WorstCaseScenario{
			PotentialLoss:  potentialLoss,
			LossPercentage: portfolio.ExpectedRisk * 100,
			RecoveryTime:   "12-24 meses",
			Description:    "Em cenário adverso, possível perda temporária",
		},
	}, nil
}

// generateAISummary creates a comprehensive AI-generated summary
func (ai *AIService) generateAISummary(request *models.AIAnalysisRequest, portfolio *models.PortfolioRecommendation, projections *models.PortfolioProjection, analysis *models.FinancialAnalysis) (string, error) {
	if ai.openAIAPIKey == "" {
		// Return a template summary if no API key
		return ai.generateTemplateSummary(request, portfolio, projections, analysis), nil
	}

	prompt := ai.buildAIPrompt(request, portfolio, projections, analysis)

	llmRequest := models.LLMRequest{
		Model:       ai.model,
		Temperature: 0.7,
		MaxTokens:   500,
		Messages: []models.LLMMessage{
			{Role: "system", Content: ai.getSystemPrompt(request.Language)},
			{Role: "user", Content: prompt},
		},
	}

	response, err := ai.callOpenAI(llmRequest)
	if err != nil {
		// Fallback to template summary
		return ai.generateTemplateSummary(request, portfolio, projections, analysis), nil
	}

	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content, nil
	}

	return ai.generateTemplateSummary(request, portfolio, projections, analysis), nil
}

// callOpenAI makes a request to OpenAI API
func (ai *AIService) callOpenAI(request models.LLMRequest) (*models.LLMResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", ai.openAIBaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ai.openAIAPIKey)

	resp, err := ai.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response models.LLMResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// getSystemPrompt returns the system prompt for the AI coach
func (ai *AIService) getSystemPrompt(language string) string {
	if language == "pt-BR" {
		return `Você é um consultor financeiro amigável e experiente especializado em investimentos brasileiros. 

INSTRUÇÕES:
- Seja conciso, claro e encorajador
- Use linguagem simples e evite jargões
- Inclua sempre um aviso legal sobre consultoria financeira
- Foque em ações práticas e exequíveis
- Limite sua resposta a 300 palavras
- Use reais (R$) para valores monetários

INCLUA:
1. Análise da situação financeira
2. Justificativa para o portfolio recomendado
3. Próximos passos práticos
4. Aviso legal sobre investimentos

Seja sempre otimista mas realista sobre expectativas de retorno.`
	}

	return `You are a friendly, experienced financial advisor specializing in Brazilian investments.

INSTRUCTIONS:
- Be concise, clear, and encouraging
- Use simple language and avoid jargon
- Always include a legal disclaimer
- Focus on practical, actionable advice
- Limit response to 300 words
- Use BRL (R$) for monetary values

INCLUDE:
1. Financial situation analysis
2. Portfolio recommendation rationale
3. Practical next steps
4. Investment disclaimer

Be optimistic but realistic about return expectations.`
}

// buildAIPrompt creates the user prompt with context
func (ai *AIService) buildAIPrompt(request *models.AIAnalysisRequest, portfolio *models.PortfolioRecommendation, projections *models.PortfolioProjection, analysis *models.FinancialAnalysis) string {
	return fmt.Sprintf(`Analise esta situação financeira e forneça recomendações:

SITUAÇÃO FINANCEIRA:
- Renda mensal: R$ %.2f
- Gastos totais: R$ %.2f  
- Sobra mensal: R$ %.2f
- Reserva atual: R$ %.2f
- Saúde financeira: %.0f/100

RECOMENDAÇÃO DE PORTFOLIO:
- Perfil: %s
- Investimento mensal: R$ %.2f
- Retorno esperado: %.1f%% ao ano
- Risco máximo: %.1f%%

PROJEÇÃO (%d ANOS):
- Valor final estimado: R$ %.2f
- Total investido: R$ %.2f
- Ganhos projetados: R$ %.2f

Forneça uma análise personalizada e recomendações práticas.`,
		request.FinancialSummary.MonthlyIncome,
		request.FinancialSummary.MonthlyFixedExpenses+request.FinancialSummary.MonthlyVariableExpenses,
		request.FinancialSummary.MonthlySurplus,
		request.FinancialSummary.TotalBalance,
		analysis.FinancialHealthScore,
		portfolio.Template.RiskLevel,
		portfolio.MonthlyInvestment,
		portfolio.ExpectedReturn*100,
		portfolio.ExpectedRisk*100,
		request.InvestmentHorizon,
		projections.TotalFinalValue,
		projections.TotalContributed,
		projections.TotalGains)
}

// generateTemplateSummary creates a fallback summary without AI
func (ai *AIService) generateTemplateSummary(request *models.AIAnalysisRequest, portfolio *models.PortfolioRecommendation, projections *models.PortfolioProjection, analysis *models.FinancialAnalysis) string {
	return fmt.Sprintf(`📊 **Análise Financeira Personalizada**

**Sua Situação:** Com uma renda de R$ %.2f e sobra mensal de R$ %.2f, você tem uma base sólida para investir. Seu score de saúde financeira é %.0f/100.

**Recomendação:** Portfolio %s com investimento mensal de R$ %.2f. Esta estratégia oferece retorno esperado de %.1f%% ao ano com risco controlado.

**Projeção em %d anos:** Investindo regularmente, você pode acumular aproximadamente R$ %.2f, com ganhos de R$ %.2f sobre o valor investido.

**Próximos Passos:**
1. %s
2. Comece com valores pequenos e aumente gradualmente
3. Revise sua carteira a cada 6 meses

⚠️ **Importante:** Esta análise é educativa. Consulte um consultor financeiro certificado antes de tomar decisões de investimento.`,
		request.FinancialSummary.MonthlyIncome,
		request.FinancialSummary.MonthlySurplus,
		analysis.FinancialHealthScore,
		portfolio.Template.RiskLevel,
		portfolio.MonthlyInvestment,
		portfolio.ExpectedReturn*100,
		request.InvestmentHorizon,
		projections.TotalFinalValue,
		projections.TotalGains,
		func() string {
			if analysis.SurplusAnalysis.CurrentEmergencyFund < analysis.SurplusAnalysis.EmergencyFundTarget {
				return "Construa sua reserva de emergência primeiro"
			}
			return "Inicie seus investimentos imediatamente"
		}())
}

// GenerateWhatIfScenario creates scenario analysis
func (ai *AIService) GenerateWhatIfScenario(baseRequest *models.AIAnalysisRequest, scenarioParams *models.ScenarioParameters) (*models.WhatIfScenario, error) {
	// Create modified request for scenario
	modifiedRequest := *baseRequest
	modifiedRequest.RiskProfile = scenarioParams.RiskLevel
	modifiedRequest.InvestmentHorizon = scenarioParams.InvestmentHorizon
	modifiedRequest.MonthlyBudget = scenarioParams.MonthlyContribution

	// Calculate scenario projections
	portfolio, err := ai.calculatePortfolioRecommendation(&modifiedRequest)
	if err != nil {
		return nil, err
	}

	projections, err := ai.calculateProjections(portfolio, scenarioParams.InvestmentHorizon, scenarioParams.MonthlyContribution)
	if err != nil {
		return nil, err
	}

	// Calculate baseline for comparison
	baselinePortfolio, _ := ai.calculatePortfolioRecommendation(baseRequest)
	baselineProjections, _ := ai.calculateProjections(baselinePortfolio, baseRequest.InvestmentHorizon, baseRequest.MonthlyBudget)

	// Calculate impact
	impact := models.ScenarioImpact{
		FinalValueDifference: projections.TotalFinalValue - baselineProjections.TotalFinalValue,
		PercentageDifference: ((projections.TotalFinalValue - baselineProjections.TotalFinalValue) / baselineProjections.TotalFinalValue) * 100,
		RiskAdjustedReturn:   portfolio.ExpectedReturn,
	}

	return &models.WhatIfScenario{
		Name:               fmt.Sprintf("Cenário: %s - %d anos", scenarioParams.RiskLevel, scenarioParams.InvestmentHorizon),
		Description:        fmt.Sprintf("Investindo R$ %.2f mensalmente com perfil %s", scenarioParams.MonthlyContribution, scenarioParams.RiskLevel),
		Parameters:         *scenarioParams,
		Projections:        *projections,
		ComparisonBaseline: *baselineProjections,
		Impact:             impact,
	}, nil
}

// Chat handles conversational AI interactions with financial context
func (ai *AIService) Chat(request *models.ChatRequest) (*models.ChatResponse, error) {
	if ai.openAIAPIKey == "" {
		// Fallback mode when OpenAI API key is not configured
		return ai.generateMockChatResponse(request), nil
	}

	// Build the system prompt with financial coaching context
	systemPrompt := ai.buildFinancialCoachSystemPrompt(request.Language)

	// Build user context if available
	contextMessage := ai.buildUserContextMessage(request)

	// Prepare conversation messages
	messages := []models.LLMMessage{
		{Role: "system", Content: systemPrompt},
	}

	// Add user context as background information
	if contextMessage != "" {
		messages = append(messages, models.LLMMessage{
			Role:    "system",
			Content: fmt.Sprintf("User Financial Context: %s", contextMessage),
		})
	}

	// Add chat history (keep last 10 messages to stay within token limits)
	historyLimit := 10
	if len(request.ChatHistory) > historyLimit {
		request.ChatHistory = request.ChatHistory[len(request.ChatHistory)-historyLimit:]
	}
	messages = append(messages, request.ChatHistory...)

	// Add current user message
	messages = append(messages, models.LLMMessage{
		Role:    "user",
		Content: request.Message,
	})

	// Check if user is asking for dashboard
	showDashboard := ai.shouldShowDashboard(request.Message, request.Language)

	// Call OpenAI
	llmRequest := models.LLMRequest{
		Model:       ai.model,
		Temperature: 0.7,
		MaxTokens:   500,
		Messages:    messages,
	}

	response, err := ai.callOpenAI(llmRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	// Generate conversation ID if not provided
	conversationID := request.ConversationID
	if conversationID == "" {
		conversationID = fmt.Sprintf("chat_%d", time.Now().Unix())
	}

	return &models.ChatResponse{
		Message:       response.Choices[0].Message.Content,
		ShowDashboard: showDashboard,
		Language:      request.Language,
		GeneratedAt:   time.Now(),
	}, nil
}

// buildFinancialCoachSystemPrompt creates the system prompt for financial coaching
func (ai *AIService) buildFinancialCoachSystemPrompt(language string) string {
	if language == "pt" {
		return `Você é um consultor financeiro IA amigável, conciso e conhecedor. Você NÃO é um consultor licenciado — inclua sempre um breve disclaimer em cada recomendação.

Contexto: Você está ajudando usuários com dados financeiros reais do Belvo (bancos brasileiros) e dados de mercado em tempo real. Você pode:
- Analisar saúde financeira
- Recomendar investimentos personalizados (SELIC, CDI, BOVA11, IVVB11, BTC)
- Simular cenários futuros
- Explicar conceitos financeiros

Diretrizes:
- Mantenha respostas ≤ 300 palavras
- Seja amigável mas profissional
- Use dados fornecidos quando disponíveis
- Inclua 3 itens de ação curtos quando dar conselhos
- Sempre inclua disclaimer sobre não ser consultor licenciado
- Se perguntarem sobre dashboard, sugira que digitem "dashboard"

Disclaimer padrão: "Lembre-se: sou uma IA assistente, não um consultor financeiro licenciado. Sempre consulte um profissional antes de decisões importantes."`

	} else {
		return `You are a friendly, concise, knowledgeable financial coach AI. You are NOT a licensed advisor — include a brief disclaimer in every recommendation.

Context: You're helping users with real financial data from Belvo (Brazilian banks) and live market data. You have access to:
- Complete transaction history with detailed information (descriptions, amounts, dates, merchants)
- Account information and balances
- Financial health metrics
- Live market data

You can:
- Show and analyze individual transactions by date, amount, merchant, category
- List recent transactions with full details when requested
- Analyze spending patterns by merchant and category
- Recommend personalized investments (SELIC, CDI, BOVA11, IVVB11, BTC)
- Simulate future scenarios
- Explain financial concepts

Guidelines:
- Keep responses ≤ 500 words when possible, unless user asks for details for transactions or things like that
- Be friendly but professional
- Use provided data when available
- Include 3 short action items when giving advice, if you think it's relevant
- Always include disclaimer about not being licensed advisor when giving advice

Standard disclaimer: "Remember: I'm an AI assistant, not a licensed financial advisor. Always consult a professional before making important financial decisions."`
	}
}

// buildUserContextMessage creates context from user's financial data INCLUDING TRANSACTIONS
func (ai *AIService) buildUserContextMessage(request *models.ChatRequest) string {
	if request.UserContext == nil {
		return ""
	}

	// Basic financial summary
	context := fmt.Sprintf("Monthly Income: $%.2f, Monthly Expenses: $%.2f, Monthly Surplus: $%.2f, Total Balance: $%.2f",
		request.UserContext.MonthlyIncome,
		request.UserContext.MonthlyFixedExpenses+request.UserContext.MonthlyVariableExpenses,
		request.UserContext.MonthlySurplus,
		request.UserContext.TotalBalance,
	)

	// Add account information
	if len(request.UserContext.Accounts) > 0 {
		context += fmt.Sprintf(", Accounts: %d accounts", len(request.UserContext.Accounts))
		for i, account := range request.UserContext.Accounts {
			if i < 3 { // Limit to first 3 accounts to avoid token overflow
				context += fmt.Sprintf(" | %s: $%.2f", account.Name, account.Balance.Available)
			}
		}
	}

	// Add transaction information - THIS IS THE CRITICAL PART!
	if len(request.UserContext.RecentTransactions) > 0 {
		context += fmt.Sprintf(", Transaction History: %d recent transactions available", len(request.UserContext.RecentTransactions))

		// Add sample recent transactions for context
		transactionSample := ""
		for i, transaction := range request.UserContext.RecentTransactions {
			if i < 5 { // Show first 5 transactions as examples
				transactionSample += fmt.Sprintf(" | %s: $%.2f (%s)",
					transaction.Description, transaction.Amount, transaction.Type)
			}
		}
		if transactionSample != "" {
			context += fmt.Sprintf(", Sample Transactions: %s", transactionSample)
		}
	}

	// Add market context
	if request.MarketContext != nil && len(request.MarketContext.Assets) > 0 {
		context += fmt.Sprintf(", SELIC Rate: %.2f%%, Market Context Available",
			request.MarketContext.BrazilianRates.SelicRate)
	}

	return context
}

// shouldShowDashboard determines if the AI should suggest showing the dashboard
func (ai *AIService) shouldShowDashboard(message, language string) bool {
	lowerMsg := strings.ToLower(message)

	if language == "pt" {
		return strings.Contains(lowerMsg, "dashboard") ||
			strings.Contains(lowerMsg, "painel") ||
			strings.Contains(lowerMsg, "visão geral") ||
			strings.Contains(lowerMsg, "gráfico") ||
			strings.Contains(lowerMsg, "mostrar")
	}

	return strings.Contains(lowerMsg, "dashboard") ||
		strings.Contains(lowerMsg, "show") ||
		strings.Contains(lowerMsg, "display") ||
		strings.Contains(lowerMsg, "visual") ||
		strings.Contains(lowerMsg, "chart")
}

// generateMockChatResponse provides fallback responses when OpenAI API key is not configured
func (ai *AIService) generateMockChatResponse(request *models.ChatRequest) *models.ChatResponse {
	message := strings.ToLower(request.Message)

	// Determine if dashboard should be shown based on keywords
	showDashboard := strings.Contains(message, "dashboard") ||
		strings.Contains(message, "overview") ||
		strings.Contains(message, "analysis") ||
		strings.Contains(message, "portfolio") ||
		strings.Contains(message, "chart")

	var responseText string
	if request.Language == "pt" {
		if showDashboard {
			responseText = "Olá! Aqui está sua análise financeira completa. Você pode ver seu dashboard com dados detalhados sobre investimentos, gastos e recomendações personalizadas. Seu perfil financeiro mostra um bom potencial para investimentos!"
		} else if strings.Contains(message, "investir") || strings.Contains(message, "investimento") {
			responseText = "Com base em sua situação financeira, recomendo começar com investimentos conservadores. Você tem um bom excedente mensal para investir. Que tal explorar o Tesouro Selic e ETFs diversificados?"
		} else if strings.Contains(message, "gastos") || strings.Contains(message, "economia") {
			responseText = "Seus gastos estão bem controlados! Você tem uma boa disciplina financeira. Para economizar ainda mais, considere revisar seus gastos variáveis e automatizar seus investimentos."
		} else {
			responseText = "Olá! Sou seu coach financeiro. Posso ajudar com análises de investimentos, planejamento financeiro e recomendações personalizadas. Como posso ajudá-lo hoje?"
		}
	} else {
		if showDashboard {
			responseText = "Hello! Here's your complete financial analysis. You can view your dashboard with detailed data about investments, expenses, and personalized recommendations. Your financial profile shows good investment potential!"
		} else if strings.Contains(message, "invest") || strings.Contains(message, "investment") {
			responseText = "Based on your financial situation, I recommend starting with conservative investments. You have a good monthly surplus for investing. How about exploring Tesouro Selic and diversified ETFs?"
		} else if strings.Contains(message, "expense") || strings.Contains(message, "saving") {
			responseText = "Your expenses are well controlled! You have good financial discipline. To save even more, consider reviewing your variable expenses and automating your investments."
		} else {
			responseText = "Hello! I'm your AI Financial Coach. I can help with investment analysis, financial planning, and personalized recommendations. How can I help you today?"
		}
	}

	return &models.ChatResponse{
		Message:       responseText,
		ShowDashboard: showDashboard,
		Language:      request.Language,
		GeneratedAt:   time.Now(),
	}
}
