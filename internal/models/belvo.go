package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// BelvoTime handles Belvo's timestamp format which may or may not include timezone
type BelvoTime time.Time

func (bt *BelvoTime) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)

	formats := []string{
		"2006-01-02T15:04:05.000000Z", // With Z timezone
		"2006-01-02T15:04:05.000000",  // Without timezone
		"2006-01-02T15:04:05Z",        // With Z, no microseconds
		"2006-01-02T15:04:05",         // Without timezone, no microseconds
		time.RFC3339,                  // Standard RFC3339
		time.RFC3339Nano,              // RFC3339 with nanoseconds
	}

	for _, format := range formats {
		if t, err := time.Parse(format, str); err == nil {
			*bt = BelvoTime(t)
			return nil
		}
	}

	return fmt.Errorf("unable to parse timestamp: %s", str)
}

func (bt BelvoTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(bt))
}

func (bt BelvoTime) Time() time.Time {
	return time.Time(bt)
}

// BelvoCredentials holds the API credentials for Belvo
type BelvoCredentials struct {
	SecretID    string `json:"secret_id"`
	SecretKey   string `json:"secret_key"`
	Environment string `json:"environment"` // "sandbox" or "production"
	BaseURL     string `json:"base_url"`
}

// BelvoLink represents a connection to a financial institution
type BelvoLink struct {
	ID                  string                 `json:"id"`
	Institution         string                 `json:"institution"`
	AccessMode          string                 `json:"access_mode"`
	LastAccessedAt      *time.Time             `json:"last_accessed_at"`
	Status              string                 `json:"status"`
	CreatedAt           time.Time              `json:"created_at"`
	ExternalID          string                 `json:"external_id"`
	InstitutionUserID   string                 `json:"institution_user_id"`
	RefreshRate         string                 `json:"refresh_rate"`
	Credentials         map[string]interface{} `json:"credentials"`
	FetchHistoricalData bool                   `json:"fetch_historical_data"`
}

// BelvoAccount represents a bank account
type BelvoAccount struct {
	ID                        string           `json:"id"`
	Link                      string           `json:"link"`
	Institution               BelvoInstitution `json:"institution"`
	CollectedAt               BelvoTime        `json:"collected_at"`
	CreatedAt                 BelvoTime        `json:"created_at"`
	Category                  string           `json:"category"`
	Type                      string           `json:"type"`
	Number                    string           `json:"number"`
	Name                      string           `json:"name"`
	Currency                  string           `json:"currency"`
	Balance                   BelvoBalance     `json:"balance"`
	LastAccessedAt            *BelvoTime       `json:"last_accessed_at"`
	Agency                    string           `json:"agency,omitempty"`
	BankProductID             string           `json:"bank_product_id,omitempty"`
	InternalIdentification    string           `json:"internal_identification,omitempty"`
	PublicIdentificationName  string           `json:"public_identification_name,omitempty"`
	PublicIdentificationValue string           `json:"public_identification_value,omitempty"`
	BalanceType               string           `json:"balance_type"`
}

// BelvoInstitution represents financial institution details
type BelvoInstitution struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Code         string `json:"code,omitempty"`
	DisplayName  string `json:"display_name,omitempty"`
	CountryCode  string `json:"country_code,omitempty"`
	Website      string `json:"website,omitempty"`
	PrimaryColor string `json:"primary_color,omitempty"`
	Logo         string `json:"logo,omitempty"`
}

// BelvoBalance represents account balance information
type BelvoBalance struct {
	Current   float64 `json:"current"`
	Available float64 `json:"available"`
}

// BelvoTransaction represents a financial transaction
type BelvoTransaction struct {
	ID                     string                 `json:"id"`
	Account                map[string]interface{} `json:"account"` // Account object, not just ID
	CollectedAt            BelvoTime              `json:"collected_at"`
	CreatedAt              BelvoTime              `json:"created_at"`
	ValueDate              string                 `json:"value_date"`
	AccountingDate         BelvoTime              `json:"accounting_date"`
	Amount                 float64                `json:"amount"`
	Balance                float64                `json:"balance"`
	Currency               string                 `json:"currency"`
	Description            string                 `json:"description"`
	Observations           *string                `json:"observations"`
	Merchant               *BelvoMerchant         `json:"merchant"`
	Category               string                 `json:"category"`
	Subcategory            *string                `json:"subcategory"`
	Reference              string                 `json:"reference"`
	Type                   string                 `json:"type"` // "INFLOW" or "OUTFLOW"
	Status                 string                 `json:"status"`
	InternalIdentification string                 `json:"internal_identification"`
}

// BelvoMerchant represents transaction merchant information
type BelvoMerchant struct {
	Name    string `json:"name"`
	Website string `json:"website,omitempty"`
	Logo    string `json:"logo,omitempty"`
}

// BelvoOwner represents account owner information
type BelvoOwner struct {
	ID                     string    `json:"id"`
	Link                   string    `json:"link"`
	CollectedAt            BelvoTime `json:"collected_at"`
	CreatedAt              BelvoTime `json:"created_at"`
	DisplayName            string    `json:"display_name"`
	FullName               string    `json:"full_name"`
	Email                  string    `json:"email"`
	PhoneNumber            string    `json:"phone_number"`
	Address                string    `json:"address"`
	InternalIdentification string    `json:"internal_identification"`
}

// BelvoIncome represents income information
type BelvoIncome struct {
	ID                    string                 `json:"id"`
	Account               string                 `json:"account"`
	CollectedAt           time.Time              `json:"collected_at"`
	CreatedAt             time.Time              `json:"created_at"`
	IncomeType            string                 `json:"income_type"`
	IncomeSourceType      string                 `json:"income_source_type"`
	Frequency             string                 `json:"frequency"`
	MonthlyAverage        float64                `json:"monthly_average"`
	Currency              string                 `json:"currency"`
	LastIncomeDescription string                 `json:"last_income_description"`
	LastIncomeDate        string                 `json:"last_income_date"`
	StabilityCoefficient  float64                `json:"stability_coefficient"`
	Regularity            string                 `json:"regularity"`
	TrendCoefficient      float64                `json:"trend_coefficient"`
	LookbackPeriods       int                    `json:"lookback_periods"`
	FullPeriods           int                    `json:"full_periods"`
	PeriodsWithIncome     int                    `json:"periods_with_income"`
	NumberOfIncomeStreams int                    `json:"number_of_income_streams"`
	ConfidenceInterval    map[string]interface{} `json:"confidence_interval"`
}

// BelvoRecurringExpense represents recurring expense information
type BelvoRecurringExpense struct {
	ID                       string                 `json:"id"`
	Account                  string                 `json:"account"`
	CollectedAt              time.Time              `json:"collected_at"`
	CreatedAt                time.Time              `json:"created_at"`
	Frequency                string                 `json:"frequency"`
	AverageTransactionAmount float64                `json:"average_transaction_amount"`
	MedianTransactionAmount  float64                `json:"median_transaction_amount"`
	DaysWithoutTransactions  []int                  `json:"days_without_transactions"`
	Category                 string                 `json:"category"`
	Currency                 string                 `json:"currency"`
	PaymentType              string                 `json:"payment_type"`
	TransactionsMeanAmount   float64                `json:"transactions_mean_amount"`
	Transactions             []BelvoTransaction     `json:"transactions"`
	ConfidenceInterval       map[string]interface{} `json:"confidence_interval"`
}

// BelvoAPIResponse represents the standard Belvo API response wrapper
type BelvoAPIResponse struct {
	Count    int                    `json:"count,omitempty"`
	Next     *string                `json:"next,omitempty"`
	Previous *string                `json:"previous,omitempty"`
	Results  []interface{}          `json:"results,omitempty"`
	Data     interface{}            `json:"data,omitempty"`
	Message  string                 `json:"message,omitempty"`
	Request  map[string]interface{} `json:"request_id,omitempty"`
}

// BelvoError represents API error responses
type BelvoError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Field     string `json:"field,omitempty"`
}

// FinancialSummary represents processed financial data for AI analysis
type FinancialSummary struct {
	UserID                  string                  `json:"user_id"`
	GeneratedAt             time.Time               `json:"generated_at"`
	MonthlyIncome           float64                 `json:"monthly_income"`
	MonthlyFixedExpenses    float64                 `json:"monthly_fixed_expenses"`
	MonthlyVariableExpenses float64                 `json:"monthly_variable_expenses"`
	MonthlySurplus          float64                 `json:"monthly_surplus"`
	TotalBalance            float64                 `json:"total_balance"`
	Accounts                []BelvoAccount          `json:"accounts"`
	RecentTransactions      []BelvoTransaction      `json:"recent_transactions"`
	IncomeStreams           []BelvoIncome           `json:"income_streams"`
	RecurringExpenses       []BelvoRecurringExpense `json:"recurring_expenses"`
	Currency                string                  `json:"currency"`
}

// CreateLinkRequest represents the request to create a Belvo link
type CreateLinkRequest struct {
	Institution         string `json:"institution"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	UsernameType        string `json:"username_type,omitempty"`
	AccessMode          string `json:"access_mode,omitempty"`
	ExternalID          string `json:"external_id,omitempty"`
	FetchHistoricalData bool   `json:"fetch_historical_data,omitempty"`
	SecretID            string `json:"secret_id,omitempty"`
	SecretKey           string `json:"secret_key,omitempty"`
}

// CreateLinkResponse represents the response from creating a Belvo link
type CreateLinkResponse struct {
	ID                  string     `json:"id"`
	Institution         string     `json:"institution"`
	AccessMode          string     `json:"access_mode"`
	LastAccessedAt      *time.Time `json:"last_accessed_at"`
	Status              string     `json:"status"`
	CreatedAt           time.Time  `json:"created_at"`
	ExternalID          string     `json:"external_id"`
	InstitutionUserID   string     `json:"institution_user_id"`
	RefreshRate         string     `json:"refresh_rate"`
	FetchHistoricalData bool       `json:"fetch_historical_data"`
}
