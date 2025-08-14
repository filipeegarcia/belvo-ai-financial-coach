package models

import "time"

// AssetType represents different types of financial assets
type AssetType string

const (
	AssetTypeEquity      AssetType = "equity"
	AssetTypeCrypto      AssetType = "crypto"
	AssetTypeFixedIncome AssetType = "fixed_income"
	AssetTypeETF         AssetType = "etf"
)

// Asset represents a financial asset that can be invested in
type Asset struct {
	Symbol      string    `json:"symbol"`
	Name        string    `json:"name"`
	Type        AssetType `json:"type"`
	Currency    string    `json:"currency"`
	Exchange    string    `json:"exchange"`
	Description string    `json:"description"`
}

// PriceData represents historical price information for an asset
type PriceData struct {
	Symbol   string    `json:"symbol"`
	Date     time.Time `json:"date"`
	Open     float64   `json:"open"`
	High     float64   `json:"high"`
	Low      float64   `json:"low"`
	Close    float64   `json:"close"`
	Volume   int64     `json:"volume"`
	AdjClose float64   `json:"adj_close"`
	Source   string    `json:"source"` // "yahoo_finance", "coingecko", "bcb"
}

// AssetPerformance represents calculated performance metrics for an asset
type AssetPerformance struct {
	Symbol                string    `json:"symbol"`
	Name                  string    `json:"name"`
	Type                  AssetType `json:"type"`
	Currency              string    `json:"currency"`
	CurrentPrice          float64   `json:"current_price"`
	PriceChange24h        float64   `json:"price_change_24h"`
	PriceChangePercent24h float64   `json:"price_change_percent_24h"`
	Returns1M             float64   `json:"returns_1m"`
	Returns3M             float64   `json:"returns_3m"`
	Returns6M             float64   `json:"returns_6m"`
	Returns1Y             float64   `json:"returns_1y"`
	Returns3Y             float64   `json:"returns_3y"`
	AnnualizedReturn      float64   `json:"annualized_return"`
	Volatility            float64   `json:"volatility"`
	LastUpdated           time.Time `json:"last_updated"`
}

// BrazilianRates represents Brazilian economic rates from Central Bank
type BrazilianRates struct {
	SelicRate   float64   `json:"selic_rate"` // Central Bank basic rate
	CDIRate     float64   `json:"cdi_rate"`   // Interbank deposit rate
	IPCARate    float64   `json:"ipca_rate"`  // Inflation rate
	Date        time.Time `json:"date"`
	Source      string    `json:"source"` // "bcb" (Banco Central do Brasil)
	LastUpdated time.Time `json:"last_updated"`
}

// CryptoData represents cryptocurrency data from CoinGecko
type CryptoData struct {
	ID                    string    `json:"id"`
	Symbol                string    `json:"symbol"`
	Name                  string    `json:"name"`
	CurrentPrice          float64   `json:"current_price"`
	MarketCap             int64     `json:"market_cap"`
	MarketCapRank         int       `json:"market_cap_rank"`
	Volume24h             int64     `json:"volume_24h"`
	PriceChange24h        float64   `json:"price_change_24h"`
	PriceChangePercent24h float64   `json:"price_change_percent_24h"`
	PriceChange7d         float64   `json:"price_change_7d"`
	PriceChange30d        float64   `json:"price_change_30d"`
	PriceChange1y         float64   `json:"price_change_1y"`
	LastUpdated           time.Time `json:"last_updated"`
}

// PortfolioAllocation represents asset allocation in a portfolio
type PortfolioAllocation struct {
	Symbol     string    `json:"symbol"`
	Name       string    `json:"name"`
	Type       AssetType `json:"type"`
	Percentage float64   `json:"percentage"` // 0.0 to 1.0
	Amount     float64   `json:"amount"`     // Amount in BRL
}

// PortfolioTemplate represents predefined portfolio allocations
type PortfolioTemplate struct {
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	RiskLevel      string                `json:"risk_level"` // "conservative", "balanced", "aggressive"
	Allocations    []PortfolioAllocation `json:"allocations"`
	ExpectedReturn float64               `json:"expected_return"` // Annual expected return
	MaxDrawdown    float64               `json:"max_drawdown"`    // Maximum expected loss
	Currency       string                `json:"currency"`
}

// PortfolioProjection represents future value projections
type PortfolioProjection struct {
	Years               int                   `json:"years"`
	MonthlyContribution float64               `json:"monthly_contribution"`
	InitialAmount       float64               `json:"initial_amount"`
	ExpectedReturn      float64               `json:"expected_return"`
	Allocations         []PortfolioAllocation `json:"allocations"`
	Projections         []ProjectionPoint     `json:"projections"`
	TotalFinalValue     float64               `json:"total_final_value"`
	TotalContributed    float64               `json:"total_contributed"`
	TotalGains          float64               `json:"total_gains"`
	Currency            string                `json:"currency"`
}

// ProjectionPoint represents a point in time in the portfolio projection
type ProjectionPoint struct {
	Month            int     `json:"month"`
	Year             int     `json:"year"`
	TotalValue       float64 `json:"total_value"`
	TotalContributed float64 `json:"total_contributed"`
	TotalGains       float64 `json:"total_gains"`
	MonthlyReturn    float64 `json:"monthly_return"`
}

// MarketDataSummary aggregates all market data for AI analysis
type MarketDataSummary struct {
	Assets         []AssetPerformance `json:"assets"`
	BrazilianRates BrazilianRates     `json:"brazilian_rates"`
	CryptoData     []CryptoData       `json:"crypto_data"`
	LastUpdated    time.Time          `json:"last_updated"`
	DataSources    []string           `json:"data_sources"`
}

// Default asset definitions for our system
var DefaultAssets = []Asset{
	{
		Symbol:      "BOVA11.SA",
		Name:        "iShares Ibovespa Fundo de Índice",
		Type:        AssetTypeETF,
		Currency:    "BRL",
		Exchange:    "B3",
		Description: "ETF que replica o desempenho do Índice Ibovespa",
	},
	{
		Symbol:      "IVVB11.SA",
		Name:        "iShares Core S&P 500 Fundo de Índice",
		Type:        AssetTypeETF,
		Currency:    "BRL",
		Exchange:    "B3",
		Description: "ETF que replica o desempenho do S&P 500 em reais",
	},
	{
		Symbol:      "BTC",
		Name:        "Bitcoin",
		Type:        AssetTypeCrypto,
		Currency:    "USD",
		Exchange:    "crypto",
		Description: "Criptomoeda descentralizada",
	},
	{
		Symbol:      "SELIC",
		Name:        "Taxa Selic",
		Type:        AssetTypeFixedIncome,
		Currency:    "BRL",
		Exchange:    "BCB",
		Description: "Taxa básica de juros da economia brasileira",
	},
	{
		Symbol:      "CDI",
		Name:        "Certificado de Depósito Interbancário",
		Type:        AssetTypeFixedIncome,
		Currency:    "BRL",
		Exchange:    "BCB",
		Description: "Taxa de referência do mercado interbancário",
	},
}

// Default portfolio templates
var DefaultPortfolioTemplates = []PortfolioTemplate{
	{
		Name:           "Conservador",
		Description:    "Foco em preservação de capital com baixo risco",
		RiskLevel:      "conservative",
		Currency:       "BRL",
		ExpectedReturn: 0.08, // 8% annual return
		MaxDrawdown:    0.05, // 5% maximum loss
		Allocations: []PortfolioAllocation{
			{Symbol: "SELIC", Type: AssetTypeFixedIncome, Percentage: 0.60},
			{Symbol: "CDI", Type: AssetTypeFixedIncome, Percentage: 0.25},
			{Symbol: "BOVA11.SA", Type: AssetTypeETF, Percentage: 0.15},
		},
	},
	{
		Name:           "Balanceado",
		Description:    "Equilibrio entre crescimento e segurança",
		RiskLevel:      "balanced",
		Currency:       "BRL",
		ExpectedReturn: 0.10, // 10% annual return
		MaxDrawdown:    0.15, // 15% maximum loss
		Allocations: []PortfolioAllocation{
			{Symbol: "SELIC", Type: AssetTypeFixedIncome, Percentage: 0.40},
			{Symbol: "BOVA11.SA", Type: AssetTypeETF, Percentage: 0.35},
			{Symbol: "IVVB11.SA", Type: AssetTypeETF, Percentage: 0.20},
			{Symbol: "BTC", Type: AssetTypeCrypto, Percentage: 0.05},
		},
	},
	{
		Name:           "Agressivo",
		Description:    "Foco em crescimento de longo prazo com maior risco",
		RiskLevel:      "aggressive",
		Currency:       "BRL",
		ExpectedReturn: 0.12, // 12% annual return
		MaxDrawdown:    0.25, // 25% maximum loss
		Allocations: []PortfolioAllocation{
			{Symbol: "BOVA11.SA", Type: AssetTypeETF, Percentage: 0.40},
			{Symbol: "IVVB11.SA", Type: AssetTypeETF, Percentage: 0.35},
			{Symbol: "BTC", Type: AssetTypeCrypto, Percentage: 0.15},
			{Symbol: "SELIC", Type: AssetTypeFixedIncome, Percentage: 0.10},
		},
	},
}
