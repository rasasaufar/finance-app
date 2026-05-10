package types

import "time"

const (
	HardcodedEmail    = "rasas@example.com"
	HardcodedPassword = "password123"
	DummyToken        = "dummy-token-rasas"
	DateLayout        = "2006-01-02"
	MonthLayout       = "2006-01"
)

// --- Auth ---

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UpdateProfileInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
}

// --- Transaction ---

type Transaction struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"`
	Category string `json:"category"`
	Amount   int64  `json:"amount"`
	Date     string `json:"date"`
	Note     string `json:"note"`
}

type TransactionInput struct {
	Type     string `json:"type"`
	Category string `json:"category"`
	Amount   int64  `json:"amount"`
	Date     string `json:"date"`
	Note     string `json:"note"`
}

type TransactionMutationResponse struct {
	Transaction  Transaction   `json:"transaction"`
	BudgetChecks []BudgetCheck `json:"budget_checks"`
	Warning      bool          `json:"warning"`
}

// --- Category ---

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CategoryInput struct {
	Name string `json:"name"`
}

// --- Budget ---

type BudgetRule struct {
	ID       int64  `json:"id"`
	Category string `json:"category"`
	Period   string `json:"period"`
	Limit    int64  `json:"limit"`
}

type BudgetRuleInput struct {
	Category string `json:"category"`
	Period   string `json:"period"`
	Limit    int64  `json:"limit"`
}

type BudgetCheck struct {
	Category   string  `json:"category"`
	Period     string  `json:"period"`
	Limit      int64   `json:"limit"`
	Used       int64   `json:"used"`
	Remaining  int64   `json:"remaining"`
	Percentage float64 `json:"percentage"`
	OverBudget bool    `json:"over_budget"`
}

// --- Salary ---

type SalaryMaster struct {
	ID     int64  `json:"id"`
	Month  string `json:"month"`
	Amount int64  `json:"amount"`
	Note   string `json:"note"`
}

type SalaryMasterInput struct {
	Month  string `json:"month"`
	Amount int64  `json:"amount"`
	Note   string `json:"note"`
}

// --- Dashboard & Reports ---

type DashboardSummaryResponse struct {
	CurrentBalance     int64         `json:"current_balance"`
	SalaryCurrentMonth int64         `json:"salary_current_month"`
	SalaryTotalToDate  int64         `json:"salary_total_to_date"`
	TodayExpense       int64         `json:"today_expense"`
	MonthExpense       int64         `json:"month_expense"`
	RemainingBudget    int64         `json:"remaining_budget"`
	MakanToday         *BudgetCheck  `json:"makan_today,omitempty"`
	MakanMonth         *BudgetCheck  `json:"makan_month,omitempty"`
	BensinMonth        *BudgetCheck  `json:"bensin_month,omitempty"`
	BudgetUsage        []BudgetCheck `json:"budget_usage"`
}

type CategorySpending struct {
	Category string `json:"category"`
	Total    int64  `json:"total"`
}

type MonthlyReportResponse struct {
	Month              string             `json:"month"`
	TotalIncome        int64              `json:"total_income"`
	TotalExpense       int64              `json:"total_expense"`
	Net                int64              `json:"net"`
	SpendingByCategory []CategorySpending `json:"spending_by_category"`
	BudgetUsage        []BudgetCheck      `json:"budget_usage"`
}

// --- DB internal types ---

type DBTransaction struct {
	ID         int64
	Type       string
	CategoryID int64
	Category   string
	Amount     int64
	Date       time.Time
	Note       string
}

type DBBudgetRule struct {
	ID         int64
	CategoryID int64
	Category   string
	Period     string
	Limit      int64
}

type DBUserProfile struct {
	ID       int64
	FullName string
	Email    string
}
