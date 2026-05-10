package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

const (
	hardcodedEmail    = "rasas@example.com"
	hardcodedPassword = "password123"
	dummyToken        = "dummy-token-rasas"
	dateLayout        = "2006-01-02"
	monthLayout       = "2006-01"
)

type appServer struct {
	store *store
}

type store struct {
	mu              sync.RWMutex
	nextTransaction int64
	nextCategory    int64
	nextBudgetRule  int64
	transactions    []Transaction
	categories      []Category
	budgetRules     []BudgetRule
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Transaction struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"`
	Category string `json:"category"`
	Amount   int64  `json:"amount"`
	Date     string `json:"date"`
	Note     string `json:"note"`
}

type transactionInput struct {
	Type     string `json:"type"`
	Category string `json:"category"`
	Amount   int64  `json:"amount"`
	Date     string `json:"date"`
	Note     string `json:"note"`
}

type transactionMutationResponse struct {
	Transaction  Transaction   `json:"transaction"`
	BudgetChecks []BudgetCheck `json:"budget_checks"`
	Warning      bool          `json:"warning"`
}

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type categoryInput struct {
	Name string `json:"name"`
}

type BudgetRule struct {
	ID       int64  `json:"id"`
	Category string `json:"category"`
	Period   string `json:"period"`
	Limit    int64  `json:"limit"`
}

type budgetRuleInput struct {
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

type dashboardSummaryResponse struct {
	CurrentBalance  int64         `json:"current_balance"`
	TodayExpense    int64         `json:"today_expense"`
	MonthExpense    int64         `json:"month_expense"`
	RemainingBudget int64         `json:"remaining_budget"`
	MakanToday      *BudgetCheck  `json:"makan_today,omitempty"`
	MakanMonth      *BudgetCheck  `json:"makan_month,omitempty"`
	BensinMonth     *BudgetCheck  `json:"bensin_month,omitempty"`
	BudgetUsage     []BudgetCheck `json:"budget_usage"`
}

type categorySpending struct {
	Category string `json:"category"`
	Total    int64  `json:"total"`
}

type monthlyReportResponse struct {
	Month              string             `json:"month"`
	TotalIncome        int64              `json:"total_income"`
	TotalExpense       int64              `json:"total_expense"`
	Net                int64              `json:"net"`
	SpendingByCategory []categorySpending `json:"spending_by_category"`
	BudgetUsage        []BudgetCheck      `json:"budget_usage"`
}

func main() {
	app := &appServer{store: newStore()}
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://localhost:5174",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Post("/auth/login", app.handleLogin)

	r.Group(func(pr chi.Router) {
		pr.Use(authMiddleware)
		pr.Get("/me", app.handleMe)
		pr.Get("/dashboard/summary", app.handleDashboardSummary)

		pr.Get("/transactions", app.handleGetTransactions)
		pr.Post("/transactions", app.handleCreateTransaction)
		pr.Put("/transactions/{id}", app.handleUpdateTransaction)
		pr.Delete("/transactions/{id}", app.handleDeleteTransaction)

		pr.Get("/categories", app.handleGetCategories)
		pr.Post("/categories", app.handleCreateCategory)
		pr.Put("/categories/{id}", app.handleUpdateCategory)
		pr.Delete("/categories/{id}", app.handleDeleteCategory)

		pr.Get("/budget-rules", app.handleGetBudgetRules)
		pr.Post("/budget-rules", app.handleCreateBudgetRule)
		pr.Put("/budget-rules/{id}", app.handleUpdateBudgetRule)
		pr.Delete("/budget-rules/{id}", app.handleDeleteBudgetRule)

		pr.Get("/reports/monthly", app.handleMonthlyReport)
	})

	log.Println("API running on http://localhost:8080")

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}

func newStore() *store {
	categories := []Category{
		{ID: 1, Name: "Makan"},
		{ID: 2, Name: "Bensin"},
		{ID: 3, Name: "Transport"},
		{ID: 4, Name: "Belanja"},
		{ID: 5, Name: "Hiburan"},
		{ID: 6, Name: "Tagihan"},
		{ID: 7, Name: "Gaji"},
	}

	budgetRules := []BudgetRule{
		{ID: 1, Category: "Makan", Period: "daily", Limit: 60000},
		{ID: 2, Category: "Bensin", Period: "monthly", Limit: 240000},
	}

	return &store{
		nextTransaction: 1,
		nextCategory:    int64(len(categories) + 1),
		nextBudgetRule:  int64(len(budgetRules) + 1),
		transactions:    []Transaction{},
		categories:      categories,
		budgetRules:     budgetRules,
	}
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(authorization, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "token tidak valid")
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
		if token != dummyToken {
			writeError(w, http.StatusUnauthorized, "token tidak valid")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *appServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	var input loginRequest
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if strings.TrimSpace(input.Email) != hardcodedEmail || strings.TrimSpace(input.Password) != hardcodedPassword {
		writeError(w, http.StatusUnauthorized, "email atau password salah")
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		Token: dummyToken,
		User: User{
			Name:  "Rasas",
			Email: hardcodedEmail,
		},
	})
}

func (a *appServer) handleMe(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, User{Name: "Rasas", Email: hardcodedEmail})
}

func (a *appServer) handleDashboardSummary(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	today := now.Format(dateLayout)
	month := now.Format(monthLayout)

	a.store.mu.RLock()
	defer a.store.mu.RUnlock()

	var totalIncome int64
	var totalExpense int64
	var todayExpense int64
	var monthExpense int64
	monthExpenseByCategory := map[string]int64{}

	for _, trx := range a.store.transactions {
		if trx.Type == "income" {
			totalIncome += trx.Amount
			continue
		}
		totalExpense += trx.Amount
		if trx.Date == today {
			todayExpense += trx.Amount
		}
		if strings.HasPrefix(trx.Date, month) {
			monthExpense += trx.Amount
			monthExpenseByCategory[strings.ToLower(trx.Category)] += trx.Amount
		}
	}

	budgetUsage := make([]BudgetCheck, 0, len(a.store.budgetRules))

	var dailyMakanRule *BudgetRule
	for _, rule := range a.store.budgetRules {
		referenceDate, err := referenceDateForRule(rule.Period, now)
		if err != nil {
			continue
		}
		if strings.EqualFold(rule.Category, "Makan") && rule.Period == "daily" {
			ruleCopy := rule
			dailyMakanRule = &ruleCopy
		}
		check := a.store.buildBudgetCheck(rule, referenceDate, 0, 0)
		budgetUsage = append(budgetUsage, check)
	}

	var makanToday *BudgetCheck
	var makanMonth *BudgetCheck
	var bensinMonth *BudgetCheck
	for i := range budgetUsage {
		item := budgetUsage[i]
		if strings.EqualFold(item.Category, "Makan") && item.Period == "daily" {
			copyValue := item
			makanToday = &copyValue
		}
		if strings.EqualFold(item.Category, "Bensin") && item.Period == "monthly" {
			copyValue := item
			bensinMonth = &copyValue
		}
	}

	if dailyMakanRule != nil {
		daysCount := int64(daysInMonth(now))
		monthlyLimit := dailyMakanRule.Limit * daysCount
		monthlyUsed := monthExpenseByCategory[strings.ToLower(dailyMakanRule.Category)]
		percentage := 0.0
		if monthlyLimit > 0 {
			percentage = (float64(monthlyUsed) / float64(monthlyLimit)) * 100
		}

		check := BudgetCheck{
			Category:   dailyMakanRule.Category,
			Period:     "monthly",
			Limit:      monthlyLimit,
			Used:       monthlyUsed,
			Remaining:  monthlyLimit - monthlyUsed,
			Percentage: percentage,
			OverBudget: monthlyUsed > monthlyLimit,
		}
		makanMonth = &check
	}

	var remainingBudget int64
	if makanMonth != nil {
		remainingBudget += makanMonth.Remaining
	}
	if bensinMonth != nil {
		remainingBudget += bensinMonth.Remaining
	}

	writeJSON(w, http.StatusOK, dashboardSummaryResponse{
		CurrentBalance:  totalIncome - totalExpense,
		TodayExpense:    todayExpense,
		MonthExpense:    monthExpense,
		RemainingBudget: remainingBudget,
		MakanToday:      makanToday,
		MakanMonth:      makanMonth,
		BensinMonth:     bensinMonth,
		BudgetUsage:     budgetUsage,
	})
}

func (a *appServer) handleGetTransactions(w http.ResponseWriter, r *http.Request) {
	a.store.mu.RLock()
	transactions := append([]Transaction(nil), a.store.transactions...)
	a.store.mu.RUnlock()

	sort.Slice(transactions, func(i, j int) bool {
		if transactions[i].Date == transactions[j].Date {
			return transactions[i].ID > transactions[j].ID
		}
		return transactions[i].Date > transactions[j].Date
	})

	writeJSON(w, http.StatusOK, transactions)
}

func (a *appServer) handleCreateTransaction(w http.ResponseWriter, r *http.Request) {
	var input transactionInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	trx, err := normalizeTransactionInput(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	a.store.mu.Lock()
	defer a.store.mu.Unlock()

	if !a.store.categoryExistsLocked(trx.Category) {
		writeError(w, http.StatusBadRequest, "kategori tidak ditemukan")
		return
	}

	trx.ID = a.store.nextTransaction
	a.store.nextTransaction++
	a.store.transactions = append(a.store.transactions, trx)

	checks := a.store.checkBudgetForTransactionLocked(trx, trx.ID)
	warning := anyOverBudget(checks)

	writeJSON(w, http.StatusCreated, transactionMutationResponse{
		Transaction:  trx,
		BudgetChecks: checks,
		Warning:      warning,
	})
}

func (a *appServer) handleUpdateTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "id transaksi tidak valid")
		return
	}

	var input transactionInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	normalized, err := normalizeTransactionInput(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	a.store.mu.Lock()
	defer a.store.mu.Unlock()

	if !a.store.categoryExistsLocked(normalized.Category) {
		writeError(w, http.StatusBadRequest, "kategori tidak ditemukan")
		return
	}

	index := -1
	for i := range a.store.transactions {
		if a.store.transactions[i].ID == id {
			index = i
			break
		}
	}
	if index == -1 {
		writeError(w, http.StatusNotFound, "transaksi tidak ditemukan")
		return
	}

	normalized.ID = id
	a.store.transactions[index] = normalized

	checks := a.store.checkBudgetForTransactionLocked(normalized, id)
	warning := anyOverBudget(checks)

	writeJSON(w, http.StatusOK, transactionMutationResponse{
		Transaction:  normalized,
		BudgetChecks: checks,
		Warning:      warning,
	})
}

func (a *appServer) handleDeleteTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "id transaksi tidak valid")
		return
	}

	a.store.mu.Lock()
	defer a.store.mu.Unlock()

	index := -1
	for i := range a.store.transactions {
		if a.store.transactions[i].ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		writeError(w, http.StatusNotFound, "transaksi tidak ditemukan")
		return
	}

	a.store.transactions = append(a.store.transactions[:index], a.store.transactions[index+1:]...)
	w.WriteHeader(http.StatusNoContent)
}

func (a *appServer) handleGetCategories(w http.ResponseWriter, r *http.Request) {
	a.store.mu.RLock()
	categories := append([]Category(nil), a.store.categories...)
	a.store.mu.RUnlock()

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name < categories[j].Name
	})

	writeJSON(w, http.StatusOK, categories)
}

func (a *appServer) handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	var input categoryInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "nama kategori wajib diisi")
		return
	}

	a.store.mu.Lock()
	defer a.store.mu.Unlock()

	if a.store.categoryNameExistsLocked(name, 0) {
		writeError(w, http.StatusBadRequest, "nama kategori sudah ada")
		return
	}

	category := Category{ID: a.store.nextCategory, Name: name}
	a.store.nextCategory++
	a.store.categories = append(a.store.categories, category)

	writeJSON(w, http.StatusCreated, category)
}

func (a *appServer) handleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "id kategori tidak valid")
		return
	}

	var input categoryInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "nama kategori wajib diisi")
		return
	}

	a.store.mu.Lock()
	defer a.store.mu.Unlock()

	if a.store.categoryNameExistsLocked(name, id) {
		writeError(w, http.StatusBadRequest, "nama kategori sudah ada")
		return
	}

	index := -1
	oldName := ""
	for i := range a.store.categories {
		if a.store.categories[i].ID == id {
			index = i
			oldName = a.store.categories[i].Name
			break
		}
	}
	if index == -1 {
		writeError(w, http.StatusNotFound, "kategori tidak ditemukan")
		return
	}

	a.store.categories[index].Name = name

	for i := range a.store.transactions {
		if strings.EqualFold(a.store.transactions[i].Category, oldName) {
			a.store.transactions[i].Category = name
		}
	}
	for i := range a.store.budgetRules {
		if strings.EqualFold(a.store.budgetRules[i].Category, oldName) {
			a.store.budgetRules[i].Category = name
		}
	}

	writeJSON(w, http.StatusOK, a.store.categories[index])
}

func (a *appServer) handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "id kategori tidak valid")
		return
	}

	a.store.mu.Lock()
	defer a.store.mu.Unlock()

	index := -1
	name := ""
	for i := range a.store.categories {
		if a.store.categories[i].ID == id {
			index = i
			name = a.store.categories[i].Name
			break
		}
	}
	if index == -1 {
		writeError(w, http.StatusNotFound, "kategori tidak ditemukan")
		return
	}

	for _, trx := range a.store.transactions {
		if strings.EqualFold(trx.Category, name) {
			writeError(w, http.StatusBadRequest, "kategori masih digunakan transaksi")
			return
		}
	}

	a.store.categories = append(a.store.categories[:index], a.store.categories[index+1:]...)

	filteredRules := make([]BudgetRule, 0, len(a.store.budgetRules))
	for _, rule := range a.store.budgetRules {
		if !strings.EqualFold(rule.Category, name) {
			filteredRules = append(filteredRules, rule)
		}
	}
	a.store.budgetRules = filteredRules

	w.WriteHeader(http.StatusNoContent)
}

func (a *appServer) handleGetBudgetRules(w http.ResponseWriter, r *http.Request) {
	a.store.mu.RLock()
	rules := append([]BudgetRule(nil), a.store.budgetRules...)
	a.store.mu.RUnlock()

	sort.Slice(rules, func(i, j int) bool {
		if rules[i].Category == rules[j].Category {
			return rules[i].Period < rules[j].Period
		}
		return rules[i].Category < rules[j].Category
	})

	writeJSON(w, http.StatusOK, rules)
}

func (a *appServer) handleCreateBudgetRule(w http.ResponseWriter, r *http.Request) {
	var input budgetRuleInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	rule, err := normalizeBudgetRuleInput(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	a.store.mu.Lock()
	defer a.store.mu.Unlock()

	if !a.store.categoryExistsLocked(rule.Category) {
		writeError(w, http.StatusBadRequest, "kategori tidak ditemukan")
		return
	}

	rule.ID = a.store.nextBudgetRule
	a.store.nextBudgetRule++
	a.store.budgetRules = append(a.store.budgetRules, rule)

	writeJSON(w, http.StatusCreated, rule)
}

func (a *appServer) handleUpdateBudgetRule(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "id budget rule tidak valid")
		return
	}

	var input budgetRuleInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	normalized, err := normalizeBudgetRuleInput(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	a.store.mu.Lock()
	defer a.store.mu.Unlock()

	if !a.store.categoryExistsLocked(normalized.Category) {
		writeError(w, http.StatusBadRequest, "kategori tidak ditemukan")
		return
	}

	index := -1
	for i := range a.store.budgetRules {
		if a.store.budgetRules[i].ID == id {
			index = i
			break
		}
	}
	if index == -1 {
		writeError(w, http.StatusNotFound, "budget rule tidak ditemukan")
		return
	}

	normalized.ID = id
	a.store.budgetRules[index] = normalized

	writeJSON(w, http.StatusOK, normalized)
}

func (a *appServer) handleDeleteBudgetRule(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "id budget rule tidak valid")
		return
	}

	a.store.mu.Lock()
	defer a.store.mu.Unlock()

	index := -1
	for i := range a.store.budgetRules {
		if a.store.budgetRules[i].ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		writeError(w, http.StatusNotFound, "budget rule tidak ditemukan")
		return
	}

	a.store.budgetRules = append(a.store.budgetRules[:index], a.store.budgetRules[index+1:]...)
	w.WriteHeader(http.StatusNoContent)
}

func (a *appServer) handleMonthlyReport(w http.ResponseWriter, r *http.Request) {
	month := strings.TrimSpace(r.URL.Query().Get("month"))
	if month == "" {
		month = time.Now().Format(monthLayout)
	}

	reference, err := time.Parse(monthLayout, month)
	if err != nil {
		writeError(w, http.StatusBadRequest, "format bulan harus YYYY-MM")
		return
	}

	a.store.mu.RLock()
	defer a.store.mu.RUnlock()

	var totalIncome int64
	var totalExpense int64
	byCategoryMap := map[string]int64{}

	for _, trx := range a.store.transactions {
		if !strings.HasPrefix(trx.Date, month) {
			continue
		}

		if trx.Type == "income" {
			totalIncome += trx.Amount
			continue
		}

		totalExpense += trx.Amount
		byCategoryMap[trx.Category] += trx.Amount
	}

	spendingByCategory := make([]categorySpending, 0, len(byCategoryMap))
	for category, total := range byCategoryMap {
		spendingByCategory = append(spendingByCategory, categorySpending{Category: category, Total: total})
	}
	sort.Slice(spendingByCategory, func(i, j int) bool {
		return spendingByCategory[i].Total > spendingByCategory[j].Total
	})

	budgetUsage := make([]BudgetCheck, 0, len(a.store.budgetRules))
	for _, rule := range a.store.budgetRules {
		referenceDate, err := referenceDateForRule(rule.Period, reference)
		if err != nil {
			continue
		}
		check := a.store.buildBudgetCheck(rule, referenceDate, 0, 0)
		budgetUsage = append(budgetUsage, check)
	}

	writeJSON(w, http.StatusOK, monthlyReportResponse{
		Month:              month,
		TotalIncome:        totalIncome,
		TotalExpense:       totalExpense,
		Net:                totalIncome - totalExpense,
		SpendingByCategory: spendingByCategory,
		BudgetUsage:        budgetUsage,
	})
}

func (s *store) categoryExistsLocked(name string) bool {
	for _, category := range s.categories {
		if strings.EqualFold(category.Name, name) {
			return true
		}
	}
	return false
}

func (s *store) categoryNameExistsLocked(name string, exceptID int64) bool {
	for _, category := range s.categories {
		if category.ID == exceptID {
			continue
		}
		if strings.EqualFold(category.Name, name) {
			return true
		}
	}
	return false
}

func (s *store) checkBudgetForTransactionLocked(trx Transaction, excludeTransactionID int64) []BudgetCheck {
	if trx.Type != "expense" {
		return []BudgetCheck{}
	}

	referenceDate, err := time.Parse(dateLayout, trx.Date)
	if err != nil {
		return []BudgetCheck{}
	}

	checks := make([]BudgetCheck, 0)
	for _, rule := range s.budgetRules {
		if !strings.EqualFold(rule.Category, trx.Category) {
			continue
		}

		check := s.buildBudgetCheck(rule, referenceDate, trx.Amount, excludeTransactionID)
		checks = append(checks, check)
	}

	sort.Slice(checks, func(i, j int) bool {
		return periodRank(checks[i].Period) < periodRank(checks[j].Period)
	})

	return checks
}

func (s *store) buildBudgetCheck(rule BudgetRule, referenceDate time.Time, additionalAmount int64, excludeTransactionID int64) BudgetCheck {
	used := s.sumExpenseForRuleLocked(rule, referenceDate, excludeTransactionID) + additionalAmount
	remaining := rule.Limit - used

	percentage := 0.0
	if rule.Limit > 0 {
		percentage = (float64(used) / float64(rule.Limit)) * 100
	}

	return BudgetCheck{
		Category:   rule.Category,
		Period:     rule.Period,
		Limit:      rule.Limit,
		Used:       used,
		Remaining:  remaining,
		Percentage: percentage,
		OverBudget: used > rule.Limit,
	}
}

func (s *store) sumExpenseForRuleLocked(rule BudgetRule, referenceDate time.Time, excludeTransactionID int64) int64 {
	total := int64(0)

	for _, trx := range s.transactions {
		if excludeTransactionID > 0 && trx.ID == excludeTransactionID {
			continue
		}
		if trx.Type != "expense" {
			continue
		}
		if !strings.EqualFold(trx.Category, rule.Category) {
			continue
		}

		trxDate, err := time.Parse(dateLayout, trx.Date)
		if err != nil {
			continue
		}

		if isSamePeriod(rule.Period, trxDate, referenceDate) {
			total += trx.Amount
		}
	}

	return total
}

func isSamePeriod(period string, trxDate, referenceDate time.Time) bool {
	switch period {
	case "daily":
		ty, tm, td := trxDate.Date()
		ry, rm, rd := referenceDate.Date()
		return ty == ry && tm == rm && td == rd
	case "weekly":
		tYear, tWeek := trxDate.ISOWeek()
		rYear, rWeek := referenceDate.ISOWeek()
		return tYear == rYear && tWeek == rWeek
	case "monthly":
		ty, tm, _ := trxDate.Date()
		ry, rm, _ := referenceDate.Date()
		return ty == ry && tm == rm
	default:
		return false
	}
}

func referenceDateForRule(period string, source time.Time) (time.Time, error) {
	switch period {
	case "daily", "weekly", "monthly":
		return source, nil
	default:
		return time.Time{}, errors.New("periode tidak valid")
	}
}

func daysInMonth(source time.Time) int {
	return time.Date(source.Year(), source.Month()+1, 0, 0, 0, 0, 0, source.Location()).Day()
}

func normalizeTransactionInput(input transactionInput) (Transaction, error) {
	transactionType := strings.ToLower(strings.TrimSpace(input.Type))
	if transactionType != "income" && transactionType != "expense" {
		return Transaction{}, errors.New("tipe transaksi harus income atau expense")
	}

	category := strings.TrimSpace(input.Category)
	if category == "" {
		return Transaction{}, errors.New("kategori wajib diisi")
	}

	if input.Amount <= 0 {
		return Transaction{}, errors.New("nominal harus lebih dari 0")
	}

	dateValue := strings.TrimSpace(input.Date)
	if _, err := time.Parse(dateLayout, dateValue); err != nil {
		return Transaction{}, errors.New("tanggal harus berformat YYYY-MM-DD")
	}

	return Transaction{
		Type:     transactionType,
		Category: category,
		Amount:   input.Amount,
		Date:     dateValue,
		Note:     strings.TrimSpace(input.Note),
	}, nil
}

func normalizeBudgetRuleInput(input budgetRuleInput) (BudgetRule, error) {
	category := strings.TrimSpace(input.Category)
	if category == "" {
		return BudgetRule{}, errors.New("kategori wajib diisi")
	}

	period := strings.ToLower(strings.TrimSpace(input.Period))
	if period != "daily" && period != "weekly" && period != "monthly" {
		return BudgetRule{}, errors.New("periode harus daily, weekly, atau monthly")
	}

	if input.Limit <= 0 {
		return BudgetRule{}, errors.New("limit budget harus lebih dari 0")
	}

	return BudgetRule{
		Category: category,
		Period:   period,
		Limit:    input.Limit,
	}, nil
}

func periodRank(period string) int {
	switch period {
	case "daily":
		return 1
	case "weekly":
		return 2
	case "monthly":
		return 3
	default:
		return 99
	}
}

func parseIDParam(r *http.Request, key string) (int64, error) {
	idStr := strings.TrimSpace(chi.URLParam(r, key))
	if idStr == "" {
		return 0, errors.New("id kosong")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("id tidak valid")
	}
	return id, nil
}

func anyOverBudget(checks []BudgetCheck) bool {
	for _, check := range checks {
		if check.OverBudget {
			return true
		}
	}
	return false
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("payload tidak valid: %w", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
