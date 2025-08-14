package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gofr.dev/pkg/gofr"

	"ai-financial-coach/internal/models"
	"ai-financial-coach/internal/service"
)

// BelvoHandler handles HTTP requests related to Belvo API
type BelvoHandler struct {
	belvoService  *service.BelvoService
	testSecretID  string
	testSecretKey string
}

// NewBelvoHandler creates a new BelvoHandler instance
func NewBelvoHandler(secretID, secretKey, environment string) *BelvoHandler {
	if environment == "" {
		environment = "sandbox" // Default to sandbox
	}

	if secretID == "" || secretKey == "" {
		fmt.Printf("❌ BelvoHandler: Missing credentials - secretID: %s, secretKey: %s\n",
			secretID,
			func() string {
				if secretKey == "" {
					return "empty"
				} else {
					return "provided"
				}
			}())
		panic("Belvo credentials are required")
	}

	fmt.Printf("✅ BelvoHandler: Initializing with provided credentials\n")
	return &BelvoHandler{
		belvoService: service.NewBelvoService(secretID, secretKey, environment),
	}
}

// SetTestCredentials sets the test credentials for easy testing
func (bh *BelvoHandler) SetTestCredentials(testSecretID, testSecretKey string) {
	bh.testSecretID = testSecretID
	bh.testSecretKey = testSecretKey
}

// CreateLinkRequest represents the request body for creating a Belvo link
type CreateLinkRequest struct {
	Institution  string `json:"institution"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	UsernameType string `json:"username_type"`
	AccessMode   string `json:"access_mode"`
	SecretID     string `json:"secret_id,omitempty"`
	SecretKey    string `json:"secret_key,omitempty"`
}

// GetInstitutions handles GET /api/belvo/institutions
func (bh *BelvoHandler) GetInstitutions(ctx *gofr.Context) (interface{}, error) {
	institutions, err := bh.belvoService.GetInstitutions()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve institutions: %w", err)
	}

	return map[string]interface{}{
		"institutions": institutions,
		"count":        len(institutions),
	}, nil
}

// CreateLink handles POST /api/belvo/links
func (bh *BelvoHandler) CreateLink(ctx *gofr.Context) (interface{}, error) {
	var req CreateLinkRequest
	if err := ctx.Bind(&req); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}

	// Validate required fields
	if req.Institution == "" || req.Username == "" || req.Password == "" {
		return nil, fmt.Errorf("institution, username, and password are required")
	}

	// Determine which Belvo service to use
	var belvoService *service.BelvoService
	if req.SecretID != "" && req.SecretKey != "" {
		// Use user-provided credentials
		belvoService = service.NewBelvoService(req.SecretID, req.SecretKey, "sandbox")
	} else {
		// Use default service credentials
		belvoService = bh.belvoService
	}

	// Create link with additional parameters for Open Finance institutions
	linkParams := map[string]interface{}{
		"institution": req.Institution,
		"username":    req.Username,
		"password":    req.Password,
		"access_mode": "single",
	}

	// Add username_type if provided (required for some institutions like Open Finance)
	if req.UsernameType != "" {
		linkParams["username_type"] = req.UsernameType
	}

	link, err := belvoService.CreateLinkWithCustomParams(linkParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create link: %w", err)
	}

	return map[string]interface{}{
		"link":    link,
		"message": "Link created successfully",
	}, nil
}

// CreateLinkWithCredentials creates a Belvo link with specific credentials
func (bh *BelvoHandler) CreateLinkWithCredentials(ctx *gofr.Context) (interface{}, error) {
	var request struct {
		Institution    string `json:"institution"`
		Username       string `json:"username"`
		Password       string `json:"password"`
		CredentialType string `json:"credential_type"` // "demo", "test", "custom"
		SecretID       string `json:"secret_id,omitempty"`
		SecretKey      string `json:"secret_key,omitempty"`
	}

	if err := ctx.Bind(&request); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}

	// Determine which credentials to use
	var secretID, secretKey string
	switch request.CredentialType {
	case "demo":
		// Use mock data - no real Belvo call needed
		return map[string]interface{}{
			"mode":    "demo",
			"link_id": "demo-link-id",
			"message": "Demo mode activated - using mock financial data",
		}, nil
	case "test":
		// Use test credentials from environment
		secretID = bh.testSecretID
		secretKey = bh.testSecretKey
	case "custom":
		// Use user-provided credentials
		secretID = request.SecretID
		secretKey = request.SecretKey
	default:
		return nil, fmt.Errorf("invalid credential_type: %s", request.CredentialType)
	}

	// Create temporary Belvo service with appropriate credentials
	tempBelvoService := service.NewBelvoService(secretID, secretKey, "sandbox")

	// Standard link creation for all institutions
	linkRequest := map[string]interface{}{
		"institution": request.Institution,
		"username":    request.Username,
		"password":    request.Password,
	}

	// Create the link using the raw service call with custom parameters
	link, err := tempBelvoService.CreateLinkWithCustomParams(linkRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create Belvo link: %w", err)
	}

	return map[string]interface{}{
		"mode":    request.CredentialType,
		"link":    link,
		"message": "Successfully connected to Belvo",
	}, nil
}

// GetConnectToken handles POST /api/belvo/connect-token to generate widget token
func (bh *BelvoHandler) GetConnectToken(ctx *gofr.Context) (interface{}, error) {
	var req struct {
		Scopes    string `json:"scopes"`
		SecretID  string `json:"secret_id"`
		SecretKey string `json:"secret_key"`
	}
	_ = ctx.Bind(&req)

	// If client provided credentials, use them to generate a token; otherwise use default service
	var token map[string]interface{}
	var err error
	if req.SecretID != "" && req.SecretKey != "" {
		temp := service.NewBelvoService(req.SecretID, req.SecretKey, "sandbox")
		token, err = temp.GenerateAccessToken(req.Scopes)
	} else {
		token, err = bh.belvoService.GenerateAccessToken(req.Scopes)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to generate connect token: %w", err)
	}

	return map[string]interface{}{
		"token":   token,
		"message": "Belvo Connect token generated",
	}, nil
}

// GenerateOFDAWidgetTokenDirect handles POST /api/belvo/widget-token/direct - direct call to Belvo API
func (bh *BelvoHandler) GenerateOFDAWidgetTokenDirect(ctx *gofr.Context) (interface{}, error) {
	var req struct {
		SecretID  string `json:"secret_id"`
		SecretKey string `json:"secret_key"`
	}
	if err := ctx.Bind(&req); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}

	if req.SecretID == "" || req.SecretKey == "" {
		return nil, fmt.Errorf("secret_id and secret_key are required")
	}

	// Direct HTTP call to Belvo API (exactly like our working curl)

	payload := map[string]interface{}{
		"id":              req.SecretID,
		"password":        req.SecretKey,
		"scopes":          "read_institutions,write_links,read_consents,write_consents,write_consent_callback,delete_consents",
		"stale_in":        "7d",
		"fetch_resources": []string{"ACCOUNTS", "TRANSACTIONS", "OWNERS"},
		"widget": map[string]interface{}{
			"branding": map[string]interface{}{
				"company_name": "AI Financial Coach",
			},
			"customer_id": "test-user",
			"identification_info": []map[string]interface{}{
				{
					"type":  "CPF",
					"value": "12345678901",
				},
			},
			"permissions": []string{"REGISTER", "ACCOUNTS", "CREDIT_CARDS", "CREDIT_OPERATIONS"},
			"callback_urls": map[string]interface{}{
				"success": "http://localhost:3000/success",
				"exit":    "http://localhost:3000/exit",
				"event":   "http://localhost:3000/event",
			},
		},
	}

	jsonData, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequest("POST", "https://sandbox.belvo.com/api/token/", bytes.NewBuffer(jsonData))
	auth := base64.StdEncoding.EncodeToString([]byte(req.SecretID + ":" + req.SecretKey))
	httpReq.Header.Set("Authorization", "Basic "+auth)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Belvo API error %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	return map[string]interface{}{
		"widget_token": result,
		"widget_url":   fmt.Sprintf("https://widget.belvo.io/#/connect?token=%s", result["access"]),
		"message":      "Direct OFDA widget token generated successfully",
	}, nil
}

// CreateEreborLink handles POST /api/belvo/create-erebor-link to create erebor_br_retail link with real data
func (bh *BelvoHandler) CreateEreborLink(ctx *gofr.Context) (interface{}, error) {
	var req struct {
		SecretID  string `json:"secret_id"`
		SecretKey string `json:"secret_key"`
		Username  string `json:"username,omitempty"`
		Password  string `json:"password,omitempty"`
	}

	if err := ctx.Bind(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Create a dynamic Belvo service with user credentials
	var belvoService *service.BelvoService
	if req.SecretID != "" && req.SecretKey != "" {
		belvoService = service.NewBelvoService(req.SecretID, req.SecretKey, "sandbox")
	} else {
		belvoService = bh.belvoService
	}

	// Use default sandbox credentials for erebor_br_retail (works without consent issues)
	username := req.Username
	password := req.Password
	if username == "" {
		username = "testuser100" // Username ending in 100 for frequent user data
	}
	if password == "" {
		password = "password123" // Standard sandbox password
	}

	// Create erebor_br_retail link (regular institution - no consent issues)
	linkRequest := map[string]interface{}{
		"institution":     "erebor_br_retail",
		"username":        username,
		"password":        password,
		"access_mode":     "single",
		"external_id":     fmt.Sprintf("erebor-link-%d", time.Now().Unix()),
		"fetch_resources": []string{"ACCOUNTS", "TRANSACTIONS", "OWNERS"},
	}

	// Debug logging
	fmt.Printf("🔗 Creating erebor_br_retail link with request: %+v\n", linkRequest)

	link, err := belvoService.CreateLinkWithCustomParams(linkRequest)
	if err != nil {
		errMsg := fmt.Sprintf("%v", err)
		fmt.Printf("❌ Belvo API Error: %s\n", errMsg)
		return nil, fmt.Errorf("failed to create erebor_br_retail link: %w", err)
	}

	fmt.Printf("✅ erebor_br_retail link created successfully: %s\n", link.ID)

	return map[string]interface{}{
		"success":     true,
		"link":        link,
		"message":     "erebor_br_retail link created successfully with real Belvo data",
		"institution": "erebor_br_retail",
		"link_id":     link.ID,
		"status":      link.Status,
	}, nil
}

// GenerateOFDAWidgetToken - DEPRECATED: Widget approach not implemented due to incomplete Belvo documentation
// See README.md for details on Open Finance limitations
func (bh *BelvoHandler) GenerateOFDAWidgetToken(ctx *gofr.Context) (interface{}, error) {
	var req struct {
		SecretID    string `json:"secret_id"`
		SecretKey   string `json:"secret_key"`
		CustomerID  string `json:"customer_id,omitempty"`
		CPF         string `json:"cpf,omitempty"`
		CompanyName string `json:"company_name,omitempty"`
	}
	if err := ctx.Bind(&req); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}

	// Validate required fields
	if req.SecretID == "" || req.SecretKey == "" {
		return nil, fmt.Errorf("secret_id and secret_key are required")
	}

	// Set defaults
	if req.CustomerID == "" {
		req.CustomerID = "ai-financial-coach-user"
	}
	if req.CPF == "" {
		req.CPF = "12345678901" // Default test CPF
	}
	if req.CompanyName == "" {
		req.CompanyName = "AI Financial Coach"
	}

	// Create the OFDA widget token request (matching the working direct API call format)
	widgetRequest := map[string]interface{}{
		"id":       req.SecretID,
		"password": req.SecretKey,
		"scopes":   "read_institutions,write_links,read_consents,write_consents,write_consent_callback,delete_consents",
		"stale_in": "7d",
		"fetch_resources": []string{
			"ACCOUNTS",
			"TRANSACTIONS",
			"OWNERS",
		},
		"widget": map[string]interface{}{
			"branding": map[string]interface{}{
				"company_name": req.CompanyName,
			},
			"customer_id": req.CustomerID,
			"identification_info": []map[string]interface{}{
				{
					"type":  "CPF",
					"value": req.CPF,
				},
			},
			"permissions": []string{
				"REGISTER",
				"ACCOUNTS",
				"CREDIT_CARDS",
				"CREDIT_OPERATIONS",
			},
			"callback_urls": map[string]interface{}{
				"success": "http://localhost:3000/success",
				"exit":    "http://localhost:3000/exit",
				"event":   "http://localhost:3000/event",
			},
		},
	}

	// Debug: Log the request we're sending
	fmt.Printf("🔧 Widget Request: %+v\n", widgetRequest)

	// Create Belvo service with user credentials
	belvoService := service.NewBelvoService(req.SecretID, req.SecretKey, "sandbox")

	// Generate the widget token
	token, err := belvoService.GenerateOFDAWidgetToken(widgetRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to generate OFDA widget token: %w", err)
	}

	return map[string]interface{}{
		"widget_token": token,
		"widget_url":   fmt.Sprintf("https://widget.belvo.io/#/connect?token=%s", token["access"]),
		"message":      "OFDA widget token generated successfully for ofmockbank_br_retail",
		"instructions": "Open the widget_url in a browser to complete the ofmockbank_br_retail connection",
	}, nil
}

// GetLatestLink handles GET /api/belvo/links/latest - returns the most recent link id
func (bh *BelvoHandler) GetLatestLink(ctx *gofr.Context) (interface{}, error) {
	links, err := bh.belvoService.GetLinks()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch links: %w", err)
	}
	if len(links) == 0 {
		return map[string]interface{}{
			"link_id": "",
			"message": "No links found",
		}, nil
	}

	latest := links[0]
	// Return first; Belvo returns most recent first in many listings. No strong ordering guaranteed.
	return map[string]interface{}{
		"link_id":     latest.ID,
		"status":      latest.Status,
		"institution": latest.Institution,
	}, nil
}

// GetLinksWithCredentials handles POST /api/belvo/links/with-credentials to get links using user credentials
func (bh *BelvoHandler) GetLinksWithCredentials(ctx *gofr.Context) (interface{}, error) {
	var req struct {
		SecretID  string `json:"secret_id"`
		SecretKey string `json:"secret_key"`
	}
	if err := ctx.Bind(&req); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}

	// Validate credentials provided
	if req.SecretID == "" || req.SecretKey == "" {
		return nil, fmt.Errorf("secret_id and secret_key are required")
	}

	// Create temporary Belvo service with user credentials
	belvoService := service.NewBelvoService(req.SecretID, req.SecretKey, "sandbox")

	links, err := belvoService.GetLinks()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch links with user credentials: %w", err)
	}

	// Filter for erebor_br_retail links and check which ones have actual data
	var validLinksWithData []models.BelvoLink
	var validLinksNoData []models.BelvoLink

	for _, link := range links {
		if link.Institution == "erebor_br_retail" && link.Status == "valid" {
			// Check if this link has actual financial data
			hasData := belvoService.CheckLinkHasData(link.ID)
			if hasData {
				validLinksWithData = append(validLinksWithData, link)
			} else {
				validLinksNoData = append(validLinksNoData, link)
			}
		}
	}

	return map[string]interface{}{
		"links":                    validLinksWithData,
		"links_without_data":       validLinksNoData,
		"all_links":                links,
		"links_with_data_count":    len(validLinksWithData),
		"links_without_data_count": len(validLinksNoData),
		"total_count":              len(links),
		"has_data_available":       len(validLinksWithData) > 0,
		"message":                  "Links retrieved and data availability checked",
	}, nil
}

// GetLinksForSelection handles POST /api/belvo/links/for-selection
func (bh *BelvoHandler) GetLinksForSelection(ctx *gofr.Context) (interface{}, error) {
	var req struct {
		SecretID  string `json:"secret_id"`
		SecretKey string `json:"secret_key"`
	}

	if err := ctx.Bind(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if req.SecretID == "" || req.SecretKey == "" {
		return nil, fmt.Errorf("secret_id and secret_key are required")
	}

	// Create dynamic Belvo service with user credentials
	belvoService := service.NewBelvoService(req.SecretID, req.SecretKey, "sandbox")

	// Get all links
	links, err := belvoService.GetLinks()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve links: %w", err)
	}

	var basicLinks []map[string]interface{}

	for i, link := range links {
		if link.Status == "valid" {
			customerNumber := i + 1
			displayName := fmt.Sprintf("Customer Account #%d", customerNumber)

			basicLinks = append(basicLinks, map[string]interface{}{
				"id":              link.ID,
				"institution":     link.Institution,
				"display_name":    displayName,
				"status":          link.Status,
				"created_at":      link.CreatedAt,
				"customer_number": customerNumber,
				"short_id":        link.ID[:8] + "...",
			})
		}
	}

	return map[string]interface{}{
		"links":        basicLinks,
		"total_count":  len(basicLinks),
		"message":      "Customer links retrieved instantly",
		"loading_type": "instant",
	}, nil
}

// GetDetailedLinkInfo handles POST /api/belvo/links/detailed-info/{link_id} - DETAILED loading for selected customer
func (bh *BelvoHandler) GetDetailedLinkInfo(ctx *gofr.Context) (interface{}, error) {
	linkID := ctx.PathParam("link_id")
	if linkID == "" {
		return nil, fmt.Errorf("link_id parameter is required")
	}

	var req struct {
		SecretID  string `json:"secret_id"`
		SecretKey string `json:"secret_key"`
	}

	if err := ctx.Bind(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if req.SecretID == "" || req.SecretKey == "" {
		return nil, fmt.Errorf("secret_id and secret_key are required")
	}

	// Create dynamic Belvo service with user credentials
	belvoService := service.NewBelvoService(req.SecretID, req.SecretKey, "sandbox")

	// 🚀 PARALLEL DATA FETCHING for maximum speed
	type fetchResult struct {
		owners       []models.BelvoOwner
		accounts     []models.BelvoAccount
		transactions []models.BelvoTransaction
		summary      *models.FinancialSummary
		ownerErr     error
		accountErr   error
		transErr     error
		summaryErr   error
	}

	result := &fetchResult{}
	done := make(chan bool, 4)

	// Parallel fetch 1: Owner info
	go func() {
		defer func() { done <- true }()
		result.owners, result.ownerErr = belvoService.GetOwners(linkID)
	}()

	// Parallel fetch 2: Accounts
	go func() {
		defer func() { done <- true }()
		result.accounts, result.accountErr = belvoService.GetAccounts(linkID)
	}()

	// Parallel fetch 3: Recent transactions (last 3 months for speed)
	go func() {
		defer func() { done <- true }()
		now := time.Now()
		dateFrom := now.AddDate(0, -3, 0)
		dateTo := now
		result.transactions, result.transErr = belvoService.GetTransactions(linkID, &dateFrom, &dateTo)
	}()

	// Parallel fetch 4: Incomes (for financial summary calculation)
	go func() {
		defer func() { done <- true }()
		incomes, err := belvoService.GetIncomes(linkID)
		if err != nil {
			result.summaryErr = err
		} else {
			result.summary = &models.FinancialSummary{}
			result.summary.IncomeStreams = incomes
		}
	}()

	// Wait for all parallel fetches to complete
	for i := 0; i < 4; i++ {
		<-done
	}

	// Process owner info
	var ownerName string = "Unknown Customer"
	if len(result.owners) > 0 && result.owners[0].DisplayName != "" {
		ownerName = result.owners[0].DisplayName
	} else if len(result.owners) > 0 && result.owners[0].FullName != "" {
		ownerName = result.owners[0].FullName
	}

	// Handle accounts
	accounts := result.accounts
	if result.accountErr != nil {
		accounts = []models.BelvoAccount{}
	}

	// Calculate total balance
	totalBalance := 0.0
	for _, account := range accounts {
		totalBalance += account.Balance.Available
	}

	// Handle transactions
	transactions := result.transactions
	if result.transErr != nil {
		transactions = []models.BelvoTransaction{}
	}

	// Check data availability
	hasData := len(accounts) > 0 || len(transactions) > 0

	// Build financial summary from already-fetched data (no duplicate API calls)
	var financialSummary *models.FinancialSummary
	if result.summaryErr != nil {
		financialSummary = createBasicFinancialSummary(linkID, accounts, transactions)
	} else {
		financialSummary = createFinancialSummaryFromData(linkID, accounts, transactions, result.summary.IncomeStreams)
	}

	// Pre-generate comprehensive AI context summary for instant responses
	var monthlyIncome, monthlyExpenses float64
	if financialSummary != nil {
		monthlyIncome = financialSummary.MonthlyIncome
		monthlyExpenses = financialSummary.MonthlyVariableExpenses
	}

	// Categorize accounts by type
	accountCategories := make(map[string]int)
	accountsByCategory := make(map[string][]map[string]interface{})
	for _, account := range accounts {
		category := account.Category
		accountCategories[category]++
		if accountsByCategory[category] == nil {
			accountsByCategory[category] = []map[string]interface{}{}
		}
		accountsByCategory[category] = append(accountsByCategory[category], map[string]interface{}{
			"name":    account.Name,
			"type":    account.Type,
			"balance": account.Balance.Current,
		})
	}

	aiContextSummary := fmt.Sprintf(
		"Customer: %s | Link: %s | Total Balance: %.2f BRL | Accounts: %d (%s) | Transactions: %d | Monthly Income: %.2f BRL | Monthly Expenses: %.2f BRL | Net Flow: %.2f BRL",
		ownerName, linkID[:8], totalBalance, len(accounts),
		fmt.Sprintf("%v", accountCategories), len(transactions),
		monthlyIncome, monthlyExpenses, monthlyIncome-monthlyExpenses,
	)

	return map[string]interface{}{
		"link_id":              linkID,
		"owner_name":           ownerName,
		"account_count":        len(accounts),
		"accounts":             accounts,
		"account_categories":   accountCategories,
		"accounts_by_category": accountsByCategory,
		"transaction_count":    len(transactions),
		"recent_transactions":  transactions,
		"total_balance":        totalBalance,
		"currency":             "BRL",
		"has_data":             hasData,
		"financial_summary":    financialSummary,
		"ai_context_summary":   aiContextSummary,
		"message":              fmt.Sprintf("Detailed analysis completed for %s", ownerName),
		"loading_type":         "detailed",
		"load_time_optimized":  true,
		"data_scope":           "recent",
	}, nil
}

// Helper functions for data verification
func extractAccountIDs(accounts []models.BelvoAccount) []string {
	ids := make([]string, len(accounts))
	for i, account := range accounts {
		ids[i] = account.ID[:8] + "..." // Short IDs for readability
	}
	return ids
}

func verifyBalanceCalculation(accounts []models.BelvoAccount, calculatedTotal float64) map[string]interface{} {
	var recalculated float64
	balances := make([]map[string]interface{}, len(accounts))

	for i, account := range accounts {
		recalculated += account.Balance.Available
		balances[i] = map[string]interface{}{
			"account_id": account.ID[:8] + "...",
			"name":       account.Name,
			"balance":    account.Balance.Available,
		}
	}

	return map[string]interface{}{
		"calculated_total":    calculatedTotal,
		"recalculated":        recalculated,
		"matches":             calculatedTotal == recalculated,
		"individual_balances": balances,
	}
}

func getSampleTransaction(transactions []models.BelvoTransaction) map[string]interface{} {
	if len(transactions) == 0 {
		return map[string]interface{}{"message": "No transactions found"}
	}

	tx := transactions[0]
	return map[string]interface{}{
		"id":          tx.ID[:8] + "...",
		"amount":      tx.Amount,
		"description": tx.Description,
		"date":        tx.AccountingDate,
		"type":        tx.Type,
	}
}

// createFinancialSummaryFromData builds financial summary from already-fetched data
func createFinancialSummaryFromData(linkID string, accounts []models.BelvoAccount, transactions []models.BelvoTransaction, incomes []models.BelvoIncome) *models.FinancialSummary {
	// Calculate total balance
	totalBalance := 0.0
	for _, account := range accounts {
		totalBalance += account.Balance.Available
	}

	// Calculate monthly income from income streams
	monthlyIncome := 0.0
	for _, income := range incomes {
		monthlyIncome += income.MonthlyAverage
	}

	// Calculate income and expenses from transactions
	totalInflow := 0.0
	totalOutflow := 0.0
	for _, transaction := range transactions {
		if transaction.Type == "INFLOW" {
			totalInflow += transaction.Amount
		} else if transaction.Type == "OUTFLOW" {
			totalOutflow += transaction.Amount
		}
	}

	// Convert to monthly averages (transactions are from last 3 months)
	monthsOfData := 3.0
	monthlyIncomeFromTransactions := totalInflow / monthsOfData
	monthlyExpensesFromTransactions := totalOutflow / monthsOfData

	// Use transaction-based income if no formal income streams found
	if monthlyIncome == 0 {
		monthlyIncome = monthlyIncomeFromTransactions
	}

	monthlyVariableExpenses := monthlyExpensesFromTransactions
	monthlySurplus := monthlyIncome - monthlyVariableExpenses

	currency := "BRL"
	if len(accounts) > 0 {
		currency = accounts[0].Currency
	}

	return &models.FinancialSummary{
		UserID:                  linkID,
		GeneratedAt:             time.Now(),
		MonthlyIncome:           monthlyIncome,
		MonthlyFixedExpenses:    0.0,
		MonthlyVariableExpenses: monthlyVariableExpenses,
		MonthlySurplus:          monthlySurplus,
		TotalBalance:            totalBalance,
		Accounts:                accounts,
		RecentTransactions:      transactions,
		IncomeStreams:           incomes,
		RecurringExpenses:       []models.BelvoRecurringExpense{},
		Currency:                currency,
	}
}

// createBasicFinancialSummary creates a basic summary when income data fails
func createBasicFinancialSummary(linkID string, accounts []models.BelvoAccount, transactions []models.BelvoTransaction) *models.FinancialSummary {
	return createFinancialSummaryFromData(linkID, accounts, transactions, []models.BelvoIncome{})
}

// VerifyLinkData handles POST /api/belvo/verify-data/{link_id} - Comprehensive data verification
func (bh *BelvoHandler) VerifyLinkData(ctx *gofr.Context) (interface{}, error) {
	linkID := ctx.PathParam("link_id")
	if linkID == "" {
		return nil, fmt.Errorf("link_id parameter is required")
	}

	var req struct {
		SecretID  string `json:"secret_id"`
		SecretKey string `json:"secret_key"`
	}

	if err := ctx.Bind(&req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if req.SecretID == "" || req.SecretKey == "" {
		return nil, fmt.Errorf("secret_id and secret_key are required")
	}

	// Create dynamic Belvo service with user credentials
	belvoService := service.NewBelvoService(req.SecretID, req.SecretKey, "sandbox")

	fmt.Printf("🔍 VERIFYING data integrity for link: %s\n", linkID)

	// Get fresh data directly from Belvo for verification
	owners, ownerErr := belvoService.GetOwners(linkID)
	accounts, accountErr := belvoService.GetAccounts(linkID)

	// Get a small sample of transactions for verification
	now := time.Now()
	sampleFrom := now.AddDate(0, 0, -7) // Last week only for quick verification
	sampleTo := now
	sampleTransactions, transErr := belvoService.GetTransactions(linkID, &sampleFrom, &sampleTo)

	verification := map[string]interface{}{
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"link_id":   linkID,
		"owners": map[string]interface{}{
			"count":   len(owners),
			"error":   getErrorString(ownerErr),
			"details": owners,
		},
		"accounts": map[string]interface{}{
			"count":         len(accounts),
			"error":         getErrorString(accountErr),
			"ids":           extractAccountIDs(accounts),
			"names":         extractAccountNames(accounts),
			"total_balance": calculateTotalBalance(accounts),
		},
		"sample_transactions": map[string]interface{}{
			"count":      len(sampleTransactions),
			"error":      getErrorString(transErr),
			"date_range": fmt.Sprintf("%s to %s", sampleFrom.Format("2006-01-02"), sampleTo.Format("2006-01-02")),
			"sample":     getSampleTransaction(sampleTransactions),
		},
		"consistency_check": map[string]interface{}{
			"owners_consistent": len(owners) > 0,
			"accounts_exist":    len(accounts) > 0,
			"data_available":    len(accounts) > 0 || len(sampleTransactions) > 0,
		},
	}

	return map[string]interface{}{
		"verification": verification,
		"status":       "completed",
	}, nil
}

// Helper functions for verification
func getErrorString(err error) string {
	if err != nil {
		return err.Error()
	}
	return "none"
}

func extractAccountNames(accounts []models.BelvoAccount) []string {
	names := make([]string, len(accounts))
	for i, account := range accounts {
		names[i] = account.Name
	}
	return names
}

func calculateTotalBalance(accounts []models.BelvoAccount) float64 {
	var total float64
	for _, account := range accounts {
		total += account.Balance.Available
	}
	return total
}

// GetAccounts handles GET /api/belvo/accounts
func (bh *BelvoHandler) GetAccounts(ctx *gofr.Context) (interface{}, error) {
	linkID := ctx.PathParam("link_id")
	if linkID == "" {
		return nil, fmt.Errorf("link_id parameter is required")
	}

	accounts, err := bh.belvoService.GetAccounts(linkID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve accounts: %w", err)
	}

	return map[string]interface{}{
		"accounts": accounts,
		"count":    len(accounts),
		"link_id":  linkID,
	}, nil
}

// GetTransactions handles GET /api/belvo/transactions
func (bh *BelvoHandler) GetTransactions(ctx *gofr.Context) (interface{}, error) {
	linkID := ctx.PathParam("link_id")
	if linkID == "" {
		return nil, fmt.Errorf("link_id parameter is required")
	}

	// Optional date filters can be added here
	transactions, err := bh.belvoService.GetTransactions(linkID, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve transactions: %w", err)
	}

	return map[string]interface{}{
		"transactions": transactions,
		"count":        len(transactions),
		"link_id":      linkID,
	}, nil
}

// GetFinancialSummary handles GET /api/belvo/financial-summary
func (bh *BelvoHandler) GetFinancialSummary(ctx *gofr.Context) (interface{}, error) {
	linkID := ctx.PathParam("link_id")
	if linkID == "" {
		return nil, fmt.Errorf("link_id parameter is required")
	}

	// For now, use default service credentials
	// TODO: Add proper credential handling via request context
	belvoService := bh.belvoService

	summary, err := belvoService.GetFinancialSummary(linkID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate financial summary: %w", err)
	}

	return map[string]interface{}{
		"financial_summary": summary,
		"link_id":           linkID,
		"message":           "Financial summary generated successfully",
	}, nil
}

// GetFinancialSummaryWithCredentials handles POST /api/belvo/financial-summary/with-credentials
func (bh *BelvoHandler) GetFinancialSummaryWithCredentials(ctx *gofr.Context) (interface{}, error) {
	var req struct {
		LinkID    string `json:"link_id"`
		SecretID  string `json:"secret_id"`
		SecretKey string `json:"secret_key"`
	}
	if err := ctx.Bind(&req); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}

	// Validate required fields
	if req.LinkID == "" {
		return nil, fmt.Errorf("link_id is required")
	}
	if req.SecretID == "" || req.SecretKey == "" {
		return nil, fmt.Errorf("secret_id and secret_key are required")
	}

	// Create Belvo service with user credentials
	belvoService := service.NewBelvoService(req.SecretID, req.SecretKey, "sandbox")

	summary, err := belvoService.GetFinancialSummary(req.LinkID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate financial summary: %w", err)
	}

	return map[string]interface{}{
		"financial_summary": summary,
		"link_id":           req.LinkID,
		"credential_source": "user-provided",
		"message":           "Financial summary generated successfully with user credentials",
	}, nil
}

// GetMockData provides mock financial data for development/testing
func (bh *BelvoHandler) GetMockData(ctx *gofr.Context) (interface{}, error) {
	// Return mock data based on our previous Belvo exploration
	mockSummary := &models.FinancialSummary{
		UserID:                  "mock-user-001",
		GeneratedAt:             time.Now(),
		MonthlyIncome:           8500.00,
		MonthlyFixedExpenses:    3200.00,
		MonthlyVariableExpenses: 2800.00,
		MonthlySurplus:          2500.00,
		TotalBalance:            67355.41,
		Currency:                "BRL",
		Accounts: []models.BelvoAccount{
			{
				ID:       "mock-account-001",
				Name:     "Conta corrente",
				Category: "CHECKING_ACCOUNT",
				Type:     "Contas",
				Number:   "5534",
				Currency: "BRL",
				Balance: models.BelvoBalance{
					Current:   67355.41,
					Available: 67355.41,
				},
				Institution: models.BelvoInstitution{
					Name: "erebor_br_retail",
					Type: "bank",
				},
			},
		},
	}

	return map[string]interface{}{
		"financial_summary": mockSummary,
		"message":           "Mock financial data for development",
		"note":              "This is sample data for testing. Use real Belvo credentials for actual financial data.",
	}, nil
}

// TestConnection handles POST /api/belvo/test-connection
func (bh *BelvoHandler) TestConnection(ctx *gofr.Context) (interface{}, error) {
	var req struct {
		SecretID  string `json:"secret_id,omitempty"`
		SecretKey string `json:"secret_key,omitempty"`
	}

	// Try to bind request body - if it fails, use default credentials
	_ = ctx.Bind(&req)

	// Determine which Belvo service to use
	var belvoService *service.BelvoService
	var credentialSource string

	if req.SecretID != "" && req.SecretKey != "" {
		// Use user-provided credentials
		belvoService = service.NewBelvoService(req.SecretID, req.SecretKey, "sandbox")
		credentialSource = "user-provided"
	} else {
		// Use default service credentials
		belvoService = bh.belvoService
		credentialSource = "default"
	}

	institutions, err := belvoService.GetInstitutions()
	if err != nil {
		return map[string]interface{}{
			"status":            "failed",
			"message":           "Failed to connect to Belvo API",
			"error":             err.Error(),
			"connected":         false,
			"credential_source": credentialSource,
		}, nil
	}

	return map[string]interface{}{
		"status":             "success",
		"message":            "Successfully connected to Belvo API",
		"connected":          true,
		"institutions_count": len(institutions),
		"environment":        belvoService.GetEnvironment(),
		"credential_source":  credentialSource,
	}, nil
}

// GetBelvoService returns the belvo service instance
func (bh *BelvoHandler) GetBelvoService() *service.BelvoService {
	return bh.belvoService
}
