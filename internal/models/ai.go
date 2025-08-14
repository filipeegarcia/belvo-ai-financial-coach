package models

import "time"

// AIAnalysisRequest represents a request for AI financial analysis
type AIAnalysisRequest struct {
	UserID            string             `json:"user_id"`
	FinancialSummary  *FinancialSummary  `json:"financial_summary"`
	MarketData        *MarketDataSummary `json:"market_data"`
	RiskProfile       string             `json:"risk_profile"`       // "conservative", "balanced", "aggressive"
	InvestmentHorizon int                `json:"investment_horizon"` // Years
	MonthlyBudget     float64            `json:"monthly_budget"`     // Monthly investment amount
	Goals             []InvestmentGoal   `json:"goals"`
	Language          string             `json:"language"` // "pt-BR", "en-US"
}

// InvestmentGoal represents user's investment objectives
type InvestmentGoal struct {
	Type         string  `json:"type"` // "retirement", "house", "emergency", "general_wealth"
	Description  string  `json:"description"`
	TargetAmount float64 `json:"target_amount"`
	TimeHorizon  int     `json:"time_horizon"` // Years
	Priority     string  `json:"priority"`     // "high", "medium", "low"
}

// AIAnalysisResponse represents the AI's comprehensive financial analysis
type AIAnalysisResponse struct {
	UserID               string                  `json:"user_id"`
	GeneratedAt          time.Time               `json:"generated_at"`
	RiskProfile          string                  `json:"risk_profile"`
	RecommendedPortfolio PortfolioRecommendation `json:"recommended_portfolio"`
	Projections          PortfolioProjection     `json:"projections"`
	Analysis             FinancialAnalysis       `json:"analysis"`
	Recommendations      []ActionRecommendation  `json:"recommendations"`
	RiskAssessment       RiskAssessment          `json:"risk_assessment"`
	Summary              string                  `json:"summary"`
	Language             string                  `json:"language"`
}

// PortfolioRecommendation represents the AI's suggested portfolio
type PortfolioRecommendation struct {
	Template          PortfolioTemplate     `json:"template"`
	CustomAllocations []PortfolioAllocation `json:"custom_allocations"`
	ExpectedReturn    float64               `json:"expected_return"`
	ExpectedRisk      float64               `json:"expected_risk"`
	Rationale         string                `json:"rationale"`
	MonthlyInvestment float64               `json:"monthly_investment"`
}

// FinancialAnalysis represents detailed financial insights
type FinancialAnalysis struct {
	SurplusAnalysis      SurplusAnalysis     `json:"surplus_analysis"`
	SpendingPatterns     SpendingPatterns    `json:"spending_patterns"`
	InvestmentReadiness  InvestmentReadiness `json:"investment_readiness"`
	MarketOpportunities  []MarketOpportunity `json:"market_opportunities"`
	FinancialHealthScore float64             `json:"financial_health_score"` // 0-100
}

// SurplusAnalysis breaks down available investment capacity
type SurplusAnalysis struct {
	MonthlySurplus         float64 `json:"monthly_surplus"`
	RecommendedInvestment  float64 `json:"recommended_investment"` // Percentage of surplus
	EmergencyFundTarget    float64 `json:"emergency_fund_target"`  // 3-6 months expenses
	CurrentEmergencyFund   float64 `json:"current_emergency_fund"`
	SafeInvestmentAmount   float64 `json:"safe_investment_amount"`
	ConservativePercentage float64 `json:"conservative_percentage"` // % of surplus to invest
}

// SpendingPatterns analyzes user's spending behavior
type SpendingPatterns struct {
	FixedExpenseRatio       float64           `json:"fixed_expense_ratio"`
	VariableExpenseRatio    float64           `json:"variable_expense_ratio"`
	SavingsRate             float64           `json:"savings_rate"`
	TopCategories           []ExpenseCategory `json:"top_categories"`
	OptimizationSuggestions []string          `json:"optimization_suggestions"`
}

// ExpenseCategory represents spending by category
type ExpenseCategory struct {
	Category   string  `json:"category"`
	Amount     float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
	Trend      string  `json:"trend"` // "increasing", "stable", "decreasing"
}

// InvestmentReadiness assesses how ready the user is to invest
type InvestmentReadiness struct {
	Score               float64  `json:"score"`           // 0-100
	ReadinessLevel      string   `json:"readiness_level"` // "not_ready", "somewhat_ready", "ready", "very_ready"
	KeyFactors          []string `json:"key_factors"`
	ImprovementAreas    []string `json:"improvement_areas"`
	RecommendedStrategy string   `json:"recommended_strategy"`
}

// MarketOpportunity represents current market conditions
type MarketOpportunity struct {
	AssetClass  string  `json:"asset_class"`
	Symbol      string  `json:"symbol"`
	Opportunity string  `json:"opportunity"` // "buy", "hold", "avoid"
	Rationale   string  `json:"rationale"`
	Confidence  float64 `json:"confidence"`   // 0-1
	TimeHorizon string  `json:"time_horizon"` // "short", "medium", "long"
}

// ActionRecommendation represents specific actions the user should take
type ActionRecommendation struct {
	Priority    string `json:"priority"` // "immediate", "short_term", "medium_term", "long_term"
	Action      string `json:"action"`
	Description string `json:"description"`
	Impact      string `json:"impact"`   // "high", "medium", "low"
	Effort      string `json:"effort"`   // "easy", "moderate", "complex"
	Timeline    string `json:"timeline"` // "1 week", "1 month", "3 months", etc.
}

// RiskAssessment evaluates investment risks
type RiskAssessment struct {
	OverallRisk       string            `json:"overall_risk"` // "low", "medium", "high"
	RiskFactors       []RiskFactor      `json:"risk_factors"`
	MitigationSteps   []string          `json:"mitigation_steps"`
	RiskTolerance     string            `json:"risk_tolerance"`
	WorstCaseScenario WorstCaseScenario `json:"worst_case_scenario"`
}

// RiskFactor represents a specific risk
type RiskFactor struct {
	Type        string  `json:"type"` // "market", "inflation", "liquidity", "concentration"
	Description string  `json:"description"`
	Severity    string  `json:"severity"`    // "low", "medium", "high"
	Probability float64 `json:"probability"` // 0-1
}

// WorstCaseScenario projects potential losses
type WorstCaseScenario struct {
	PotentialLoss  float64 `json:"potential_loss"`  // Amount
	LossPercentage float64 `json:"loss_percentage"` // % of portfolio
	RecoveryTime   string  `json:"recovery_time"`   // "6 months", "2 years"
	Description    string  `json:"description"`
}

// AIPromptContext contains all context for the AI prompt
type AIPromptContext struct {
	UserProfile      UserProfile         `json:"user_profile"`
	FinancialData    *FinancialSummary   `json:"financial_data"`
	MarketData       *MarketDataSummary  `json:"market_data"`
	SelectedTemplate PortfolioTemplate   `json:"selected_template"`
	Projections      PortfolioProjection `json:"projections"`
	Language         string              `json:"language"`
}

// UserProfile represents user characteristics for AI analysis
type UserProfile struct {
	Age           int              `json:"age"`
	IncomeLevel   string           `json:"income_level"`   // "low", "medium", "high"
	FamilyStatus  string           `json:"family_status"`  // "single", "married", "family_with_children"
	RiskTolerance string           `json:"risk_tolerance"` // "conservative", "moderate", "aggressive"
	InvestmentExp string           `json:"investment_exp"` // "beginner", "intermediate", "advanced"
	Goals         []InvestmentGoal `json:"goals"`
	TimeHorizon   int              `json:"time_horizon"` // Years
}

// WhatIfScenario represents scenario analysis
type WhatIfScenario struct {
	Name               string              `json:"name"`
	Description        string              `json:"description"`
	Parameters         ScenarioParameters  `json:"parameters"`
	Projections        PortfolioProjection `json:"projections"`
	ComparisonBaseline PortfolioProjection `json:"comparison_baseline"`
	Impact             ScenarioImpact      `json:"impact"`
}

// ScenarioParameters defines the scenario variables
type ScenarioParameters struct {
	MonthlyContribution float64 `json:"monthly_contribution"`
	InvestmentHorizon   int     `json:"investment_horizon"`
	RiskLevel           string  `json:"risk_level"`
	MarketCondition     string  `json:"market_condition"` // "bull", "bear", "normal"
}

// ScenarioImpact shows the difference from baseline
type ScenarioImpact struct {
	FinalValueDifference float64 `json:"final_value_difference"`
	PercentageDifference float64 `json:"percentage_difference"`
	TimeToGoalDifference int     `json:"time_to_goal_difference"` // Days
	RiskAdjustedReturn   float64 `json:"risk_adjusted_return"`
}

// LLMRequest represents a request to the language model
type LLMRequest struct {
	Model       string       `json:"model"`
	Messages    []LLMMessage `json:"messages"`
	Temperature float64      `json:"temperature"`
	MaxTokens   int          `json:"max_tokens"`
}

// LLMMessage represents a message in the conversation
type LLMMessage struct {
	Role    string `json:"role"` // "system", "user", "assistant"
	Content string `json:"content"`
}

// LLMResponse represents the language model's response
type LLMResponse struct {
	ID      string      `json:"id"`
	Object  string      `json:"object"`
	Created int64       `json:"created"`
	Model   string      `json:"model"`
	Choices []LLMChoice `json:"choices"`
	Usage   LLMUsage    `json:"usage"`
}

// LLMChoice represents a choice in the LLM response
type LLMChoice struct {
	Index        int        `json:"index"`
	Message      LLMMessage `json:"message"`
	FinishReason string     `json:"finish_reason"`
}

// LLMUsage represents token usage statistics
type LLMUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatRequest represents a request for conversational AI
type ChatRequest struct {
	Message        string             `json:"message"`
	ConversationID string             `json:"conversation_id,omitempty"`
	Language       string             `json:"language"` // "en" or "pt"
	UserContext    *FinancialSummary  `json:"user_context,omitempty"`
	MarketContext  *MarketDataSummary `json:"market_context,omitempty"`
	ChatHistory    []LLMMessage       `json:"chat_history,omitempty"`
	// Optional: if provided, backend will fetch real data from Belvo
	CredentialMode string `json:"credential_mode,omitempty"` // "demo", "test", "custom"
	LinkID         string `json:"link_id,omitempty"`
	SecretID       string `json:"secret_id,omitempty"`
	SecretKey      string `json:"secret_key,omitempty"`
}

// ChatResponse represents the AI's conversational response
type ChatResponse struct {
	ConversationID string    `json:"conversation_id"`
	Message        string    `json:"message"`
	Language       string    `json:"language"`
	GeneratedAt    time.Time `json:"generated_at"`
	TokensUsed     int       `json:"tokens_used,omitempty"`
	ShowDashboard  bool      `json:"show_dashboard,omitempty"`
}
