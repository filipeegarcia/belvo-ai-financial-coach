package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ai-financial-coach/internal/models"
)

// BelvoService handles interactions with the Belvo API
type BelvoService struct {
	credentials *models.BelvoCredentials
	httpClient  *http.Client
}

// NewBelvoService creates a new instance of BelvoService
func NewBelvoService(secretID, secretKey, environment string) *BelvoService {
	baseURL := "https://sandbox.belvo.com" // Default to sandbox
	if environment == "production" {
		baseURL = "https://api.belvo.com"
	}

	// Service created successfully

	return &BelvoService{
		credentials: &models.BelvoCredentials{
			SecretID:    secretID,
			SecretKey:   secretKey,
			Environment: environment,
			BaseURL:     baseURL,
		},
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetEnvironment returns the current environment (sandbox/production)
func (bs *BelvoService) GetEnvironment() string {
	return bs.credentials.Environment
}

// makeRequest performs authenticated HTTP requests to Belvo API
func (bs *BelvoService) makeRequest(method, endpoint string, body []byte) (*http.Response, error) {
	url := bs.credentials.BaseURL + endpoint

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Making authenticated request to Belvo API

	// Set basic auth
	req.SetBasicAuth(bs.credentials.SecretID, bs.credentials.SecretKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return bs.httpClient.Do(req)
}

// GenerateAccessToken creates a short-lived access token for Belvo Connect Widget
func (bs *BelvoService) GenerateAccessToken(scopes string) (map[string]interface{}, error) {
	if scopes == "" {
		// Minimal valid scopes for launching Connect and creating links
		scopes = "read_institutions,write_links,read_links"
	}

	payload := map[string]string{
		"id":       bs.credentials.SecretID,
		"password": bs.credentials.SecretKey,
		"scopes":   scopes,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token payload: %w", err)
	}

	// Token endpoint expects id/password in body instead of Basic Auth
	url := bs.credentials.BaseURL + "/api/token/"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := bs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request access token: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Response structure can evolve; return as generic map and let caller pick fields
	var tokenResp map[string]interface{}
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return tokenResp, nil
}

// GetLinks returns existing Belvo links
func (bs *BelvoService) GetLinks() ([]models.BelvoLink, error) {
	resp, err := bs.makeRequest("GET", "/api/links/", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get links: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse models.BelvoAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var links []models.BelvoLink
	for _, result := range apiResponse.Results {
		linkBytes, _ := json.Marshal(result)
		var link models.BelvoLink
		if err := json.Unmarshal(linkBytes, &link); err == nil {
			links = append(links, link)
		}
	}

	return links, nil
}

// CheckLinkHasData checks if a link has actual financial data by testing accounts endpoint
func (bs *BelvoService) CheckLinkHasData(linkID string) bool {
	endpoint := fmt.Sprintf("%s/api/accounts/?link=%s", bs.credentials.BaseURL, linkID)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return false
	}

	req.SetBasicAuth(bs.credentials.SecretID, bs.credentials.SecretKey)
	req.Header.Set("accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false
	}

	var result struct {
		Results []interface{} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}

	// Return true if we have at least one account
	return len(result.Results) > 0
}

// GetInstitutions retrieves available financial institutions
func (bs *BelvoService) GetInstitutions() ([]models.BelvoInstitution, error) {
	resp, err := bs.makeRequest("GET", "/api/institutions/", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get institutions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse models.BelvoAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert interface{} to BelvoInstitution
	var institutions []models.BelvoInstitution
	for _, result := range apiResponse.Results {
		institutionBytes, _ := json.Marshal(result)
		var institution models.BelvoInstitution
		if err := json.Unmarshal(institutionBytes, &institution); err == nil {
			institutions = append(institutions, institution)
		}
	}

	return institutions, nil
}

// CreateLink creates a new connection to a financial institution
func (bs *BelvoService) CreateLink(institution, username, password string) (*models.CreateLinkResponse, error) {
	linkRequest := models.CreateLinkRequest{
		Institution:         institution,
		Username:            username,
		Password:            password,
		AccessMode:          "single",
		FetchHistoricalData: true,
	}

	bodyBytes, err := json.Marshal(linkRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal link request: %w", err)
	}

	resp, err := bs.makeRequest("POST", "/api/links/", bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create link: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		var belvoError models.BelvoError
		if json.Unmarshal(body, &belvoError) == nil {
			return nil, fmt.Errorf("Belvo API error: %s - %s", belvoError.Code, belvoError.Message)
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var linkResponse models.CreateLinkResponse
	if err := json.Unmarshal(body, &linkResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal link response: %w", err)
	}

	return &linkResponse, nil
}

// CreateLinkWithCustomParams creates a Belvo link with custom parameters (for Open Finance Brazil institutions)
func (bs *BelvoService) CreateLinkWithCustomParams(linkRequest map[string]interface{}) (*models.BelvoLink, error) {
	endpoint := fmt.Sprintf("%s/api/links/", bs.credentials.BaseURL)

	// Convert the map to JSON
	requestBody, err := json.Marshal(linkRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(bs.credentials.SecretID, bs.credentials.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := bs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var link models.BelvoLink
	if err := json.Unmarshal(body, &link); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &link, nil
}

// GetAccounts retrieves accounts for a specific link using POST request (as required by Belvo)
func (bs *BelvoService) GetAccounts(linkID string) ([]models.BelvoAccount, error) {
	endpoint := fmt.Sprintf("%s/api/accounts/", bs.credentials.BaseURL)

	// Belvo requires POST request with link in body
	requestBody := map[string]interface{}{
		"link": linkID,
	}

	reqBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(bs.credentials.SecretID, bs.credentials.SecretKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Debug: Log response details
	fmt.Printf("🔍 Accounts API Response: Status=%d, BodyLength=%d\n", resp.StatusCode, len(body))
	if len(body) > 200 {
		fmt.Printf("🔍 First 200 chars of response: %s...\n", string(body[:200]))
	} else {
		fmt.Printf("🔍 Full response: %s\n", string(body))
	}

	// Belvo returns accounts directly as an array, not wrapped in "results"
	var accounts []models.BelvoAccount
	if err := json.Unmarshal(body, &accounts); err != nil {
		fmt.Printf("❌ Failed to unmarshal accounts: %v\n", err)
		return nil, fmt.Errorf("failed to unmarshal accounts: %w", err)
	}

	fmt.Printf("✅ Successfully parsed %d accounts\n", len(accounts))
	return accounts, nil
}

// GetTransactions retrieves transactions for a specific link using POST request (as required by Belvo)
func (bs *BelvoService) GetTransactions(linkID string, dateFrom, dateTo *time.Time) ([]models.BelvoTransaction, error) {
	endpoint := fmt.Sprintf("%s/api/transactions/", bs.credentials.BaseURL)

	// Belvo requires POST request with parameters in body
	requestBody := map[string]interface{}{
		"link": linkID,
	}

	// Add date filters if provided
	if dateFrom != nil {
		requestBody["date_from"] = dateFrom.Format("2006-01-02")
	}
	if dateTo != nil {
		requestBody["date_to"] = dateTo.Format("2006-01-02")
	}

	reqBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(bs.credentials.SecretID, bs.credentials.SecretKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Belvo returns transactions directly as an array, not wrapped in "results"
	var transactions []models.BelvoTransaction
	if err := json.Unmarshal(body, &transactions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transactions: %w", err)
	}

	return transactions, nil
}

// GetIncomes retrieves income information for a specific link
func (bs *BelvoService) GetIncomes(linkID string) ([]models.BelvoIncome, error) {
	endpoint := "/api/incomes/"
	if linkID != "" {
		endpoint += "?link=" + linkID
	}

	resp, err := bs.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get incomes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse models.BelvoAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var incomes []models.BelvoIncome
	for _, result := range apiResponse.Results {
		incomeBytes, _ := json.Marshal(result)
		var income models.BelvoIncome
		if err := json.Unmarshal(incomeBytes, &income); err == nil {
			incomes = append(incomes, income)
		}
	}

	return incomes, nil
}

// GetRecurringExpenses retrieves recurring expense information for a specific link
func (bs *BelvoService) GetRecurringExpenses(linkID string) ([]models.BelvoRecurringExpense, error) {
	endpoint := "/api/recurring-expenses/"
	if linkID != "" {
		endpoint += "?link=" + linkID
	}

	resp, err := bs.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get recurring expenses: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse models.BelvoAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var expenses []models.BelvoRecurringExpense
	for _, result := range apiResponse.Results {
		expenseBytes, _ := json.Marshal(result)
		var expense models.BelvoRecurringExpense
		if err := json.Unmarshal(expenseBytes, &expense); err == nil {
			expenses = append(expenses, expense)
		}
	}

	return expenses, nil
}

// GetFinancialSummary aggregates all financial data into a summary for AI analysis
func (bs *BelvoService) GetFinancialSummary(linkID string) (*models.FinancialSummary, error) {
	fmt.Printf("🔍 Starting GetFinancialSummary for link: %s\n", linkID)

	// Get all financial data in parallel
	accounts, err := bs.GetAccounts(linkID)
	if err != nil {
		fmt.Printf("❌ Failed to get accounts: %v\n", err)
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	fmt.Printf("✅ Retrieved %d accounts\n", len(accounts))

	// Get transactions for the last 90 days
	dateTo := time.Now()
	dateFrom := dateTo.AddDate(0, -3, 0) // 3 months ago
	fmt.Printf("📅 Fetching transactions from %s to %s\n", dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))
	transactions, err := bs.GetTransactions(linkID, &dateFrom, &dateTo)
	if err != nil {
		fmt.Printf("❌ Failed to get transactions: %v\n", err)
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	fmt.Printf("✅ Retrieved %d transactions\n", len(transactions))

	incomes, err := bs.GetIncomes(linkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get incomes: %w", err)
	}

	recurringExpenses, err := bs.GetRecurringExpenses(linkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recurring expenses: %w", err)
	}

	// Calculate financial metrics
	totalBalance := 0.0
	for _, account := range accounts {
		totalBalance += account.Balance.Available
	}

	monthlyIncome := 0.0
	for _, income := range incomes {
		monthlyIncome += income.MonthlyAverage
	}

	monthlyFixedExpenses := 0.0
	monthlyVariableExpenses := 0.0
	for _, expense := range recurringExpenses {
		monthlyFixedExpenses += expense.AverageTransactionAmount
	}

	// Calculate income and expenses from transactions
	totalInflow := 0.0
	totalOutflow := 0.0

	fmt.Printf("🔍 Processing %d transactions for link %s\n", len(transactions), linkID)

	for _, transaction := range transactions {
		if transaction.Type == "INFLOW" {
			totalInflow += transaction.Amount
		} else if transaction.Type == "OUTFLOW" {
			totalOutflow += transaction.Amount
		}
	}

	fmt.Printf("💰 Total inflow: %.2f, Total outflow: %.2f\n", totalInflow, totalOutflow)

	// Convert to monthly averages (transactions are from last 3 months)
	monthsOfData := 3.0
	if len(transactions) > 0 {
		// Calculate estimated monthly averages
		monthlyIncomeFromTransactions := totalInflow / monthsOfData
		monthlyExpensesFromTransactions := totalOutflow / monthsOfData

		fmt.Printf("📊 Monthly income from transactions: %.2f\n", monthlyIncomeFromTransactions)
		fmt.Printf("📊 Monthly expenses from transactions: %.2f\n", monthlyExpensesFromTransactions)

		// Use transaction-based income if no formal income streams found
		if monthlyIncome == 0 {
			monthlyIncome = monthlyIncomeFromTransactions
			fmt.Printf("✅ Using transaction-based income: %.2f\n", monthlyIncome)
		}

		// Add transaction-based expenses to variable expenses
		monthlyVariableExpenses += monthlyExpensesFromTransactions
		fmt.Printf("✅ Variable expenses set to: %.2f\n", monthlyVariableExpenses)
	}

	monthlySurplus := monthlyIncome - monthlyFixedExpenses - monthlyVariableExpenses

	currency := "BRL" // Default for Brazil
	if len(accounts) > 0 {
		currency = accounts[0].Currency
	}

	return &models.FinancialSummary{
		UserID:                  linkID, // Using linkID as user identifier for now
		GeneratedAt:             time.Now(),
		MonthlyIncome:           monthlyIncome,
		MonthlyFixedExpenses:    monthlyFixedExpenses,
		MonthlyVariableExpenses: monthlyVariableExpenses,
		MonthlySurplus:          monthlySurplus,
		TotalBalance:            totalBalance,
		Accounts:                accounts,
		RecentTransactions:      transactions,
		IncomeStreams:           incomes,
		RecurringExpenses:       recurringExpenses,
		Currency:                currency,
	}, nil
}

// GenerateOFDAWidgetToken generates a widget token for Open Finance Data Aggregation in Brazil
func (bs *BelvoService) GenerateOFDAWidgetToken(widgetRequest map[string]interface{}) (map[string]interface{}, error) {
	// Prepare the request body
	reqBody, err := json.Marshal(widgetRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal widget request: %w", err)
	}

	// Make request to Belvo's widget token endpoint
	resp, err := bs.makeRequest("POST", "/api/token/", reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create widget token request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read widget token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("widget token request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse the response
	var tokenResp map[string]interface{}
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse widget token response: %w", err)
	}

	return tokenResp, nil
}

// GetOwners retrieves owner information for a specific link
func (bs *BelvoService) GetOwners(linkID string) ([]models.BelvoOwner, error) {
	endpoint := fmt.Sprintf("%s/api/owners/", bs.credentials.BaseURL)

	requestBody := map[string]interface{}{
		"link": linkID,
	}

	reqBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(bs.credentials.SecretID, bs.credentials.SecretKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to retrieve owners: status %d", resp.StatusCode)
	}

	var owners []models.BelvoOwner
	if err := json.NewDecoder(resp.Body).Decode(&owners); err != nil {
		return nil, fmt.Errorf("failed to decode owners response: %w", err)
	}

	return owners, nil
}
