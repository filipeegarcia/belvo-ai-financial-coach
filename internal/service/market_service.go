package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ai-financial-coach/internal/models"
)

// MarketService handles fetching market data from multiple sources
type MarketService struct {
	httpClient *http.Client
}

// NewMarketService creates a new MarketService instance
func NewMarketService() *MarketService {
	return &MarketService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// YahooFinanceResponse represents the structure from Yahoo Finance API
type YahooFinanceResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Currency           string  `json:"currency"`
				Symbol             string  `json:"symbol"`
				ExchangeName       string  `json:"exchangeName"`
				InstrumentType     string  `json:"instrumentType"`
				RegularMarketPrice float64 `json:"regularMarketPrice"`
				PreviousClose      float64 `json:"previousClose"`
				ChartPreviousClose float64 `json:"chartPreviousClose"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []int64   `json:"volume"`
				} `json:"quote"`
				AdjClose []struct {
					AdjClose []float64 `json:"adjclose"`
				} `json:"adjclose"`
			} `json:"indicators"`
		} `json:"result"`
	} `json:"chart"`
}

// CoinGeckoResponse represents Bitcoin data from CoinGecko
type CoinGeckoResponse struct {
	Bitcoin struct {
		USD float64 `json:"usd"`
		BRL float64 `json:"brl"`
	} `json:"bitcoin"`
}

// CoinGeckoHistoricalResponse represents historical data from CoinGecko
type CoinGeckoHistoricalResponse struct {
	Prices [][]float64 `json:"prices"` // [timestamp, price]
}

// BrazilianCentralBankResponse represents data from Brazilian Central Bank API
type BrazilianCentralBankResponse struct {
	Value []struct {
		Date  string  `json:"data"`
		Value float64 `json:"valor"`
	} `json:"valor"`
}

// GetAssetPerformance fetches and calculates performance metrics for an asset
func (ms *MarketService) GetAssetPerformance(symbol string) (*models.AssetPerformance, error) {
	asset := ms.getAssetInfo(symbol)

	switch asset.Type {
	case models.AssetTypeETF:
		return ms.getETFPerformance(symbol)
	case models.AssetTypeCrypto:
		return ms.getCryptoPerformance(symbol)
	case models.AssetTypeFixedIncome:
		return ms.getFixedIncomePerformance(symbol)
	default:
		return nil, fmt.Errorf("unsupported asset type: %s", asset.Type)
	}
}

// getAssetInfo returns asset information from our default assets
func (ms *MarketService) getAssetInfo(symbol string) models.Asset {
	for _, asset := range models.DefaultAssets {
		if asset.Symbol == symbol {
			return asset
		}
	}
	// Return default if not found
	return models.Asset{
		Symbol:   symbol,
		Name:     symbol,
		Type:     models.AssetTypeEquity,
		Currency: "BRL",
	}
}

// getETFPerformance fetches ETF data from Yahoo Finance
func (ms *MarketService) getETFPerformance(symbol string) (*models.AssetPerformance, error) {
	// Fetch historical data (1 year)
	endTime := time.Now().Unix()
	startTime := time.Now().AddDate(-1, 0, 0).Unix()

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?period1=%d&period2=%d&interval=1d",
		symbol, startTime, endTime)

	resp, err := ms.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Yahoo Finance data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Yahoo Finance API error: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var yahooResp YahooFinanceResponse
	if err := json.Unmarshal(body, &yahooResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Yahoo response: %w", err)
	}

	if len(yahooResp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data returned from Yahoo Finance for %s", symbol)
	}

	result := yahooResp.Chart.Result[0]
	asset := ms.getAssetInfo(symbol)

	// Calculate performance metrics
	prices := result.Indicators.Quote[0].Close
	if len(prices) == 0 {
		return nil, fmt.Errorf("no price data available for %s", symbol)
	}

	currentPrice := result.Meta.RegularMarketPrice
	previousClose := result.Meta.PreviousClose

	// Calculate returns
	returns1M := ms.calculateReturn(prices, 30)
	returns3M := ms.calculateReturn(prices, 90)
	returns6M := ms.calculateReturn(prices, 180)
	returns1Y := ms.calculateReturn(prices, 365)

	// Calculate annualized return and volatility
	annualizedReturn := ms.calculateAnnualizedReturn(prices)
	volatility := ms.calculateVolatility(prices)

	return &models.AssetPerformance{
		Symbol:                symbol,
		Name:                  asset.Name,
		Type:                  asset.Type,
		Currency:              result.Meta.Currency,
		CurrentPrice:          currentPrice,
		PriceChange24h:        currentPrice - previousClose,
		PriceChangePercent24h: ((currentPrice - previousClose) / previousClose) * 100,
		Returns1M:             returns1M,
		Returns3M:             returns3M,
		Returns6M:             returns6M,
		Returns1Y:             returns1Y,
		AnnualizedReturn:      annualizedReturn,
		Volatility:            volatility,
		LastUpdated:           time.Now(),
	}, nil
}

// getCryptoPerformance fetches crypto data from CoinGecko
func (ms *MarketService) getCryptoPerformance(symbol string) (*models.AssetPerformance, error) {
	// Get current price
	url := "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd,brl&include_24hr_change=true"

	resp, err := ms.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CoinGecko data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CoinGecko API error: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse current price response
	var priceResp map[string]map[string]interface{}
	if err := json.Unmarshal(body, &priceResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CoinGecko response: %w", err)
	}

	bitcoinData, exists := priceResp["bitcoin"]
	if !exists {
		return nil, fmt.Errorf("bitcoin data not found in response")
	}

	currentPriceUSD, _ := bitcoinData["usd"].(float64)
	change24h, _ := bitcoinData["usd_24h_change"].(float64)

	// Get historical data for longer-term returns
	historicalURL := "https://api.coingecko.com/api/v3/coins/bitcoin/market_chart?vs_currency=usd&days=365"
	histResp, err := ms.httpClient.Get(historicalURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch historical data: %w", err)
	}
	defer histResp.Body.Close()

	histBody, err := io.ReadAll(histResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read historical response: %w", err)
	}

	var histData CoinGeckoHistoricalResponse
	if err := json.Unmarshal(histBody, &histData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal historical data: %w", err)
	}

	// Extract prices for calculations
	var prices []float64
	for _, pricePoint := range histData.Prices {
		if len(pricePoint) >= 2 {
			prices = append(prices, pricePoint[1])
		}
	}

	// Calculate returns
	returns1M := ms.calculateReturn(prices, 30)
	returns3M := ms.calculateReturn(prices, 90)
	returns6M := ms.calculateReturn(prices, 180)
	returns1Y := ms.calculateReturn(prices, 365)

	annualizedReturn := ms.calculateAnnualizedReturn(prices)
	volatility := ms.calculateVolatility(prices)

	asset := ms.getAssetInfo(symbol)

	return &models.AssetPerformance{
		Symbol:                symbol,
		Name:                  asset.Name,
		Type:                  asset.Type,
		Currency:              "USD",
		CurrentPrice:          currentPriceUSD,
		PriceChange24h:        (currentPriceUSD * change24h) / 100,
		PriceChangePercent24h: change24h,
		Returns1M:             returns1M,
		Returns3M:             returns3M,
		Returns6M:             returns6M,
		Returns1Y:             returns1Y,
		AnnualizedReturn:      annualizedReturn,
		Volatility:            volatility,
		LastUpdated:           time.Now(),
	}, nil
}

// getFixedIncomePerformance gets Brazilian fixed income rates
func (ms *MarketService) getFixedIncomePerformance(symbol string) (*models.AssetPerformance, error) {
	rates, err := ms.GetBrazilianRates()
	if err != nil {
		return nil, fmt.Errorf("failed to get Brazilian rates: %w", err)
	}

	asset := ms.getAssetInfo(symbol)
	var currentRate float64

	switch symbol {
	case "SELIC":
		currentRate = rates.SelicRate
	case "CDI":
		currentRate = rates.CDIRate
	default:
		currentRate = rates.SelicRate // Default to Selic
	}

	// Convert annual rate to daily return for consistency

	return &models.AssetPerformance{
		Symbol:                symbol,
		Name:                  asset.Name,
		Type:                  asset.Type,
		Currency:              "BRL",
		CurrentPrice:          currentRate, // Use rate as "price"
		PriceChange24h:        0,           // Rates don't change daily
		PriceChangePercent24h: 0,
		Returns1M:             currentRate / 12,     // Monthly approximation
		Returns3M:             currentRate * 3 / 12, // Quarterly approximation
		Returns6M:             currentRate / 2,      // Semi-annual
		Returns1Y:             currentRate,          // Annual rate
		AnnualizedReturn:      currentRate,
		Volatility:            0.01, // Low volatility for fixed income
		LastUpdated:           time.Now(),
	}, nil
}

// GetBrazilianRates fetches current Brazilian economic rates
func (ms *MarketService) GetBrazilianRates() (*models.BrazilianRates, error) {
	// For now, we'll use approximated rates since the Central Bank API requires specific setup
	// In production, you would integrate with: https://olinda.bcb.gov.br/olinda/servico/PTAX/versao/v1/odata/

	return &models.BrazilianRates{
		SelicRate:   10.75, // Current approximate Selic rate
		CDIRate:     10.40, // CDI typically slightly below Selic
		IPCARate:    4.50,  // Current approximate inflation
		Date:        time.Now(),
		Source:      "bcb_approximated",
		LastUpdated: time.Now(),
	}, nil
}

// calculateReturn calculates return over a specific number of days
func (ms *MarketService) calculateReturn(prices []float64, days int) float64 {
	if len(prices) < days || days <= 0 {
		return 0
	}

	currentPrice := prices[len(prices)-1]
	pastPrice := prices[len(prices)-days]

	if pastPrice == 0 {
		return 0
	}

	return ((currentPrice - pastPrice) / pastPrice) * 100
}

// calculateAnnualizedReturn calculates the annualized return from price series
func (ms *MarketService) calculateAnnualizedReturn(prices []float64) float64 {
	if len(prices) < 2 {
		return 0
	}

	firstPrice := prices[0]
	lastPrice := prices[len(prices)-1]
	days := float64(len(prices))

	if firstPrice == 0 || days == 0 {
		return 0
	}

	// Calculate compound annual growth rate (CAGR)
	years := days / 365.0
	totalReturn := lastPrice / firstPrice
	annualizedReturn := (totalReturn - 1) * 100

	if years > 0 {
		annualizedReturn = (totalReturn - 1) * (365.0 / days) * 100
	}

	return annualizedReturn
}

// calculateVolatility calculates price volatility (standard deviation of returns)
func (ms *MarketService) calculateVolatility(prices []float64) float64 {
	if len(prices) < 2 {
		return 0
	}

	// Calculate daily returns
	var returns []float64
	for i := 1; i < len(prices); i++ {
		if prices[i-1] != 0 {
			dailyReturn := (prices[i] - prices[i-1]) / prices[i-1]
			returns = append(returns, dailyReturn)
		}
	}

	if len(returns) == 0 {
		return 0
	}

	// Calculate mean return
	var sum float64
	for _, ret := range returns {
		sum += ret
	}
	mean := sum / float64(len(returns))

	// Calculate variance
	var variance float64
	for _, ret := range returns {
		variance += (ret - mean) * (ret - mean)
	}
	variance = variance / float64(len(returns))

	// Return annualized volatility (standard deviation * sqrt(365))
	dailyVolatility := variance
	if dailyVolatility < 0 {
		dailyVolatility = 0
	}

	return dailyVolatility * 19.1 * 100 // sqrt(365) ≈ 19.1, convert to percentage
}

// GetMarketDataSummary aggregates all market data
func (ms *MarketService) GetMarketDataSummary() (*models.MarketDataSummary, error) {
	var assets []models.AssetPerformance
	var dataSources []string

	// Fetch data for all default assets
	for _, asset := range models.DefaultAssets {
		performance, err := ms.GetAssetPerformance(asset.Symbol)
		if err != nil {
			// Log error but continue with other assets
			continue
		}
		assets = append(assets, *performance)
	}

	// Get Brazilian rates
	rates, err := ms.GetBrazilianRates()
	if err != nil {
		return nil, fmt.Errorf("failed to get Brazilian rates: %w", err)
	}

	dataSources = []string{"yahoo_finance", "coingecko", "bcb_approximated"}

	return &models.MarketDataSummary{
		Assets:         assets,
		BrazilianRates: *rates,
		LastUpdated:    time.Now(),
		DataSources:    dataSources,
	}, nil
}
