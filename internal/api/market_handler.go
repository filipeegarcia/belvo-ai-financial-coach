package api

import (
	"fmt"

	"gofr.dev/pkg/gofr"

	"ai-financial-coach/internal/models"
	"ai-financial-coach/internal/service"
)

// MarketHandler handles HTTP requests related to market data
type MarketHandler struct {
	marketService *service.MarketService
}

// NewMarketHandler creates a new MarketHandler instance
func NewMarketHandler() *MarketHandler {
	return &MarketHandler{
		marketService: service.NewMarketService(),
	}
}

// GetAssets handles GET /api/market/assets
func (mh *MarketHandler) GetAssets(ctx *gofr.Context) (interface{}, error) {
	return map[string]interface{}{
		"assets": models.DefaultAssets,
		"count":  len(models.DefaultAssets),
	}, nil
}

// GetAssetPerformance handles GET /api/market/assets/{symbol}
func (mh *MarketHandler) GetAssetPerformance(ctx *gofr.Context) (interface{}, error) {
	symbol := ctx.PathParam("symbol")
	if symbol == "" {
		return nil, fmt.Errorf("symbol parameter is required")
	}

	performance, err := mh.marketService.GetAssetPerformance(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset performance for %s: %w", symbol, err)
	}

	return map[string]interface{}{
		"asset_performance": performance,
		"symbol":            symbol,
	}, nil
}

// GetMarketData handles GET /api/market/data
func (mh *MarketHandler) GetMarketData(ctx *gofr.Context) (interface{}, error) {
	summary, err := mh.marketService.GetMarketDataSummary()
	if err != nil {
		return nil, fmt.Errorf("failed to get market data summary: %w", err)
	}

	return map[string]interface{}{
		"market_data": summary,
		"message":     "Market data summary retrieved successfully",
	}, nil
}

// GetBrazilianRates handles GET /api/market/brazilian-rates
func (mh *MarketHandler) GetBrazilianRates(ctx *gofr.Context) (interface{}, error) {
	rates, err := mh.marketService.GetBrazilianRates()
	if err != nil {
		return nil, fmt.Errorf("failed to get Brazilian rates: %w", err)
	}

	return map[string]interface{}{
		"brazilian_rates": rates,
		"message":         "Brazilian economic rates retrieved successfully",
	}, nil
}

// GetPortfolioTemplates handles GET /api/market/portfolio-templates
func (mh *MarketHandler) GetPortfolioTemplates(ctx *gofr.Context) (interface{}, error) {
	return map[string]interface{}{
		"portfolio_templates": models.DefaultPortfolioTemplates,
		"count":               len(models.DefaultPortfolioTemplates),
		"message":             "Portfolio templates retrieved successfully",
	}, nil
}

// GetPortfolioTemplate handles GET /api/market/portfolio-templates/{risk_level}
func (mh *MarketHandler) GetPortfolioTemplate(ctx *gofr.Context) (interface{}, error) {
	riskLevel := ctx.PathParam("risk_level")
	if riskLevel == "" {
		return nil, fmt.Errorf("risk_level parameter is required")
	}

	for _, template := range models.DefaultPortfolioTemplates {
		if template.RiskLevel == riskLevel {
			return map[string]interface{}{
				"portfolio_template": template,
				"risk_level":         riskLevel,
			}, nil
		}
	}

	return nil, fmt.Errorf("portfolio template not found for risk level: %s", riskLevel)
}

// GetMarketService returns the market service instance
func (mh *MarketHandler) GetMarketService() *service.MarketService {
	return mh.marketService
}
