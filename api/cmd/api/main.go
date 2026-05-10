package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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
	db *pgxpool.Pool
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

type SalaryMaster struct {
	ID     int64  `json:"id"`
	Month  string `json:"month"`
	Amount int64  `json:"amount"`
	Note   string `json:"note"`
}

type salaryMasterInput struct {
	Month  string `json:"month"`
	Amount int64  `json:"amount"`
	Note   string `json:"note"`
}

type dashboardSummaryResponse struct {
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

type dbTransaction struct {
	ID         int64
	Type       string
	CategoryID int64
	Category   string
	Amount     int64
	Date       time.Time
	Note       string
}

type dbBudgetRule struct {
	ID         int64
	CategoryID int64
	Category   string
	Period     string
	Limit      int64
}

type dbUserProfile struct {
	ID       int64
	FullName string
	Email    string
}

func main() {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		databaseURL = "postgres://finance_user:finance_pass@localhost:5432/finance_app?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("gagal inisialisasi postgres: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("gagal konek ke postgres: %v", err)
	}

	st := &store{db: pool}
	if err := st.ensureSchema(ctx); err != nil {
		log.Fatalf("gagal setup schema: %v", err)
	}
	if err := st.seedDefaults(ctx); err != nil {
		log.Fatalf("gagal seed data default: %v", err)
	}

	app := &appServer{store: st}
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
		pr.Put("/me", app.handleUpdateMe)
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

		pr.Get("/salary-masters", app.handleGetSalaryMasters)
		pr.Post("/salary-masters", app.handleCreateSalaryMaster)
		pr.Put("/salary-masters/{id}", app.handleUpdateSalaryMaster)
		pr.Delete("/salary-masters/{id}", app.handleDeleteSalaryMaster)

		pr.Get("/reports/monthly", app.handleMonthlyReport)
	})

	log.Printf("API running on http://localhost:8080 (postgres connected)")

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
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

	profileName := "Rasas"
	if profile, err := a.store.findUserProfileByEmail(r.Context(), hardcodedEmail); err == nil {
		profileName = profile.FullName
	}

	writeJSON(w, http.StatusOK, loginResponse{
		Token: dummyToken,
		User: User{
			Name:  profileName,
			Email: hardcodedEmail,
		},
	})
}

func (a *appServer) handleMe(w http.ResponseWriter, r *http.Request) {
	profile, err := a.store.findUserProfileByEmail(r.Context(), hardcodedEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusOK, User{Name: "Rasas", Email: hardcodedEmail})
			return
		}
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, User{Name: profile.FullName, Email: profile.Email})
}

type updateProfileInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
}

func (a *appServer) handleUpdateMe(w http.ResponseWriter, r *http.Request) {
	var input updateProfileInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "nama tidak boleh kosong")
		return
	}

	email := strings.TrimSpace(input.Email)
	if email == "" {
		writeError(w, http.StatusBadRequest, "email tidak boleh kosong")
		return
	}

	ctx := r.Context()
	updated, err := a.store.updateUserProfile(ctx, hardcodedEmail, name, email, strings.TrimSpace(input.Password))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "profil tidak ditemukan")
			return
		}
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, User{Name: updated.FullName, Email: updated.Email})
}

func (a *appServer) handleDashboardSummary(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	today := now.Format(dateLayout)
	month := now.Format(monthLayout)
	ctx := r.Context()

	transactions, err := a.store.listTransactionsDB(ctx)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	rules, err := a.store.listBudgetRulesDB(ctx)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	salaryCurrentMonth, err := a.store.salaryForMonth(ctx, month)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	salaryTotalToDate, err := a.store.salaryToMonth(ctx, month)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	var additionalIncome int64
	var totalExpense int64
	var todayExpense int64
	var monthExpense int64
	monthExpenseByCategory := map[string]int64{}
	monthExpenseLabelByCategory := map[string]string{}

	for _, trx := range transactions {
		if trx.Type == "income" {
			// Gaji dikelola lewat master saldo bulanan agar tidak dobel hitung.
			if !strings.EqualFold(trx.Category, "Gaji") {
				additionalIncome += trx.Amount
			}
			continue
		}

		totalExpense += trx.Amount
		trxDate := trx.Date.Format(dateLayout)
		if trxDate == today {
			todayExpense += trx.Amount
		}
		if trx.Date.Format(monthLayout) == month {
			monthExpense += trx.Amount
			categoryKey := strings.ToLower(strings.TrimSpace(trx.Category))
			monthExpenseByCategory[categoryKey] += trx.Amount
			if _, exists := monthExpenseLabelByCategory[categoryKey]; !exists {
				monthExpenseLabelByCategory[categoryKey] = strings.TrimSpace(trx.Category)
			}
		}
	}

	budgetUsage := make([]BudgetCheck, 0, len(rules))
	budgetRuleCategorySet := map[string]struct{}{}
	var bensinMonthlyRuleLimit int64
	var dailyMakanRule *dbBudgetRule

	for _, rule := range rules {
		ruleCategoryKey := strings.ToLower(strings.TrimSpace(rule.Category))
		if ruleCategoryKey != "" {
			budgetRuleCategorySet[ruleCategoryKey] = struct{}{}
		}
		if strings.EqualFold(rule.Category, "Bensin") && rule.Period == "monthly" {
			bensinMonthlyRuleLimit += rule.Limit
		}

		used, err := a.store.sumExpenseForRule(ctx, rule.CategoryID, rule.Period, now, 0)
		if err != nil {
			writeInternalServerError(w, err)
			return
		}

		check := buildBudgetCheck(rule.Category, rule.Period, rule.Limit, used)
		budgetUsage = append(budgetUsage, check)

		if strings.EqualFold(rule.Category, "Makan") && rule.Period == "daily" {
			ruleCopy := rule
			dailyMakanRule = &ruleCopy
		}
	}

	makanMonthlyRuleLimit := int64(0)
	if dailyMakanRule != nil {
		makanMonthlyRuleLimit = dailyMakanRule.Limit * int64(daysInMonth(now))
	}

	otherCategoryLimit := salaryCurrentMonth - makanMonthlyRuleLimit - bensinMonthlyRuleLimit
	if otherCategoryLimit < 0 {
		otherCategoryLimit = 0
	}

	// Tetap tampilkan pemakaian kategori yang punya transaksi bulanan walau belum punya rule budget.
	for categoryKey, used := range monthExpenseByCategory {
		if used <= 0 {
			continue
		}
		if _, exists := budgetRuleCategorySet[categoryKey]; exists {
			continue
		}

		categoryLabel := monthExpenseLabelByCategory[categoryKey]
		if categoryLabel == "" {
			categoryLabel = categoryKey
		}

		percentage := 0.0
		if otherCategoryLimit > 0 {
			percentage = (float64(used) / float64(otherCategoryLimit)) * 100
		} else if used > 0 {
			percentage = 100
		}

		budgetUsage = append(budgetUsage, BudgetCheck{
			Category:   categoryLabel,
			Period:     "monthly",
			Limit:      otherCategoryLimit,
			Used:       used,
			Remaining:  otherCategoryLimit - used,
			Percentage: percentage,
			OverBudget: used > otherCategoryLimit,
		})
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
		monthlyUsed := monthExpenseByCategory[strings.ToLower(strings.TrimSpace(dailyMakanRule.Category))]
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
		CurrentBalance:     salaryTotalToDate + additionalIncome - totalExpense,
		SalaryCurrentMonth: salaryCurrentMonth,
		SalaryTotalToDate:  salaryTotalToDate,
		TodayExpense:       todayExpense,
		MonthExpense:       monthExpense,
		RemainingBudget:    remainingBudget,
		MakanToday:         makanToday,
		MakanMonth:         makanMonth,
		BensinMonth:        bensinMonth,
		BudgetUsage:        budgetUsage,
	})
}

func (a *appServer) handleGetTransactions(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.listTransactionsDB(r.Context())
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	transactions := make([]Transaction, 0, len(items))
	for _, item := range items {
		transactions = append(transactions, Transaction{
			ID:       item.ID,
			Type:     item.Type,
			Category: item.Category,
			Amount:   item.Amount,
			Date:     item.Date.Format(dateLayout),
			Note:     item.Note,
		})
	}

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

	ctx := r.Context()
	category, err := a.store.findCategoryByName(ctx, trx.Category)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusBadRequest, "kategori tidak ditemukan")
			return
		}
		writeInternalServerError(w, err)
		return
	}

	trxDate, _ := time.Parse(dateLayout, trx.Date)
	checks, err := a.store.checkBudgetForTransaction(ctx, category.ID, trx.Type, trxDate, trx.Amount, 0)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	var id int64
	err = a.store.db.QueryRow(
		ctx,
		`INSERT INTO transactions (type, category_id, amount, trx_date, note)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		trx.Type,
		category.ID,
		trx.Amount,
		trxDate,
		trx.Note,
	).Scan(&id)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	created := Transaction{
		ID:       id,
		Type:     trx.Type,
		Category: category.Name,
		Amount:   trx.Amount,
		Date:     trxDate.Format(dateLayout),
		Note:     trx.Note,
	}

	writeJSON(w, http.StatusCreated, transactionMutationResponse{
		Transaction:  created,
		BudgetChecks: checks,
		Warning:      anyOverBudget(checks),
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

	ctx := r.Context()
	category, err := a.store.findCategoryByName(ctx, normalized.Category)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusBadRequest, "kategori tidak ditemukan")
			return
		}
		writeInternalServerError(w, err)
		return
	}

	trxDate, _ := time.Parse(dateLayout, normalized.Date)
	checks, err := a.store.checkBudgetForTransaction(ctx, category.ID, normalized.Type, trxDate, normalized.Amount, id)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	var updatedID int64
	err = a.store.db.QueryRow(
		ctx,
		`UPDATE transactions
		 SET type = $1,
		     category_id = $2,
		     amount = $3,
		     trx_date = $4,
		     note = $5,
		     updated_at = NOW()
		 WHERE id = $6
		 RETURNING id`,
		normalized.Type,
		category.ID,
		normalized.Amount,
		trxDate,
		normalized.Note,
		id,
	).Scan(&updatedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "transaksi tidak ditemukan")
			return
		}
		writeInternalServerError(w, err)
		return
	}

	updated := Transaction{
		ID:       updatedID,
		Type:     normalized.Type,
		Category: category.Name,
		Amount:   normalized.Amount,
		Date:     trxDate.Format(dateLayout),
		Note:     normalized.Note,
	}

	writeJSON(w, http.StatusOK, transactionMutationResponse{
		Transaction:  updated,
		BudgetChecks: checks,
		Warning:      anyOverBudget(checks),
	})
}

func (a *appServer) handleDeleteTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "id transaksi tidak valid")
		return
	}

	result, err := a.store.db.Exec(r.Context(), `DELETE FROM transactions WHERE id = $1`, id)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	if result.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "transaksi tidak ditemukan")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *appServer) handleGetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := a.store.listCategoriesDB(r.Context())
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
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

	exists, err := a.store.categoryNameExists(r.Context(), name, 0)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	if exists {
		writeError(w, http.StatusBadRequest, "nama kategori sudah ada")
		return
	}

	category := Category{}
	err = a.store.db.QueryRow(
		r.Context(),
		`INSERT INTO categories (name) VALUES ($1) RETURNING id, name`,
		name,
	).Scan(&category.ID, &category.Name)
	if err != nil {
		if isUniqueViolation(err) {
			writeError(w, http.StatusBadRequest, "nama kategori sudah ada")
			return
		}
		writeInternalServerError(w, err)
		return
	}

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

	exists, err := a.store.categoryNameExists(r.Context(), name, id)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	if exists {
		writeError(w, http.StatusBadRequest, "nama kategori sudah ada")
		return
	}

	category := Category{}
	err = a.store.db.QueryRow(
		r.Context(),
		`UPDATE categories SET name = $1 WHERE id = $2 RETURNING id, name`,
		name,
		id,
	).Scan(&category.ID, &category.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "kategori tidak ditemukan")
			return
		}
		if isUniqueViolation(err) {
			writeError(w, http.StatusBadRequest, "nama kategori sudah ada")
			return
		}
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, category)
}

func (a *appServer) handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "id kategori tidak valid")
		return
	}

	ctx := r.Context()
	category, err := a.store.findCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "kategori tidak ditemukan")
			return
		}
		writeInternalServerError(w, err)
		return
	}

	var usageCount int64
	err = a.store.db.QueryRow(ctx, `SELECT COUNT(*) FROM transactions WHERE category_id = $1`, id).Scan(&usageCount)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	if usageCount > 0 {
		writeError(w, http.StatusBadRequest, "kategori masih digunakan transaksi")
		return
	}

	result, err := a.store.db.Exec(ctx, `DELETE FROM categories WHERE id = $1`, category.ID)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	if result.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "kategori tidak ditemukan")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *appServer) handleGetBudgetRules(w http.ResponseWriter, r *http.Request) {
	rules, err := a.store.listBudgetRulesDB(r.Context())
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	response := make([]BudgetRule, 0, len(rules))
	for _, rule := range rules {
		response = append(response, BudgetRule{
			ID:       rule.ID,
			Category: rule.Category,
			Period:   rule.Period,
			Limit:    rule.Limit,
		})
	}

	writeJSON(w, http.StatusOK, response)
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

	ctx := r.Context()
	category, err := a.store.findCategoryByName(ctx, rule.Category)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusBadRequest, "kategori tidak ditemukan")
			return
		}
		writeInternalServerError(w, err)
		return
	}

	created := BudgetRule{Category: category.Name, Period: rule.Period, Limit: rule.Limit}
	err = a.store.db.QueryRow(
		ctx,
		`INSERT INTO budget_rules (category_id, period, limit_amount)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		category.ID,
		rule.Period,
		rule.Limit,
	).Scan(&created.ID)
	if err != nil {
		if isUniqueViolation(err) {
			writeError(w, http.StatusBadRequest, "budget rule untuk kategori dan periode tersebut sudah ada")
			return
		}
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, created)
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

	ctx := r.Context()
	category, err := a.store.findCategoryByName(ctx, normalized.Category)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusBadRequest, "kategori tidak ditemukan")
			return
		}
		writeInternalServerError(w, err)
		return
	}

	updated := BudgetRule{ID: id, Category: category.Name, Period: normalized.Period, Limit: normalized.Limit}
	err = a.store.db.QueryRow(
		ctx,
		`UPDATE budget_rules
		 SET category_id = $1,
		     period = $2,
		     limit_amount = $3
		 WHERE id = $4
		 RETURNING id`,
		category.ID,
		normalized.Period,
		normalized.Limit,
		id,
	).Scan(&updated.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "budget rule tidak ditemukan")
			return
		}
		if isUniqueViolation(err) {
			writeError(w, http.StatusBadRequest, "budget rule untuk kategori dan periode tersebut sudah ada")
			return
		}
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (a *appServer) handleDeleteBudgetRule(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "id budget rule tidak valid")
		return
	}

	result, err := a.store.db.Exec(r.Context(), `DELETE FROM budget_rules WHERE id = $1`, id)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	if result.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "budget rule tidak ditemukan")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *appServer) handleGetSalaryMasters(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.listSalaryMastersDB(r.Context())
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (a *appServer) handleCreateSalaryMaster(w http.ResponseWriter, r *http.Request) {
	var input salaryMasterInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := normalizeSalaryMasterInput(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	created := SalaryMaster{}
	err = a.store.db.QueryRow(
		r.Context(),
		`INSERT INTO salary_masters (month, amount, note)
		 VALUES ($1, $2, $3)
		 RETURNING id, month, amount, note`,
		item.Month,
		item.Amount,
		item.Note,
	).Scan(&created.ID, &created.Month, &created.Amount, &created.Note)
	if err != nil {
		if isUniqueViolation(err) {
			writeError(w, http.StatusBadRequest, "master gaji bulan tersebut sudah ada")
			return
		}
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (a *appServer) handleUpdateSalaryMaster(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "id master gaji tidak valid")
		return
	}

	var input salaryMasterInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	normalized, err := normalizeSalaryMasterInput(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	updated := SalaryMaster{}
	err = a.store.db.QueryRow(
		r.Context(),
		`UPDATE salary_masters
		 SET month = $1,
		     amount = $2,
		     note = $3,
		     updated_at = NOW()
		 WHERE id = $4
		 RETURNING id, month, amount, note`,
		normalized.Month,
		normalized.Amount,
		normalized.Note,
		id,
	).Scan(&updated.ID, &updated.Month, &updated.Amount, &updated.Note)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "master gaji tidak ditemukan")
			return
		}
		if isUniqueViolation(err) {
			writeError(w, http.StatusBadRequest, "master gaji bulan tersebut sudah ada")
			return
		}
		writeInternalServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (a *appServer) handleDeleteSalaryMaster(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "id master gaji tidak valid")
		return
	}

	result, err := a.store.db.Exec(r.Context(), `DELETE FROM salary_masters WHERE id = $1`, id)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	if result.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "master gaji tidak ditemukan")
		return
	}

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

	transactions, err := a.store.listTransactionsDB(r.Context())
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	rules, err := a.store.listBudgetRulesDB(r.Context())
	if err != nil {
		writeInternalServerError(w, err)
		return
	}

	var totalIncome int64
	var totalExpense int64
	byCategoryMap := map[string]int64{}

	for _, trx := range transactions {
		if trx.Date.Format(monthLayout) != month {
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

	budgetUsage := make([]BudgetCheck, 0, len(rules))
	for _, rule := range rules {
		used, err := a.store.sumExpenseForRule(r.Context(), rule.CategoryID, rule.Period, reference, 0)
		if err != nil {
			writeInternalServerError(w, err)
			return
		}
		budgetUsage = append(budgetUsage, buildBudgetCheck(rule.Category, rule.Period, rule.Limit, used))
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

func (s *store) ensureSchema(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS categories (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_categories_lower_name ON categories ((LOWER(name)))`,
		`CREATE TABLE IF NOT EXISTS budget_rules (
			id BIGSERIAL PRIMARY KEY,
			category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
			period TEXT NOT NULL CHECK (period IN ('daily', 'weekly', 'monthly')),
			limit_amount BIGINT NOT NULL CHECK (limit_amount > 0),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE (category_id, period)
		)`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id BIGSERIAL PRIMARY KEY,
			type TEXT NOT NULL CHECK (type IN ('income', 'expense')),
			category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
			amount BIGINT NOT NULL CHECK (amount > 0),
			trx_date DATE NOT NULL,
			note TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_trx_date ON transactions (trx_date)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_category_id ON transactions (category_id)`,
		`CREATE TABLE IF NOT EXISTS salary_masters (
			id BIGSERIAL PRIMARY KEY,
			month TEXT NOT NULL UNIQUE CHECK (month ~ '^[0-9]{4}-[0-9]{2}$'),
			amount BIGINT NOT NULL CHECK (amount > 0),
			note TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS user_profiles (
			id BIGSERIAL PRIMARY KEY,
			full_name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL DEFAULT '',
			avatar_url TEXT NOT NULL DEFAULT '',
			avatar_initials TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_user_profiles_lower_email ON user_profiles ((LOWER(email)))`,
	}

	for _, statement := range statements {
		if _, err := s.db.Exec(ctx, statement); err != nil {
			return err
		}
	}

	return nil
}

func (s *store) seedDefaults(ctx context.Context) error {
	if _, err := s.db.Exec(ctx, `
		INSERT INTO categories (name) VALUES
		('Makan'),
		('Bensin'),
		('Transport'),
		('Belanja'),
		('Hiburan'),
		('Tagihan'),
		('Gaji')
		ON CONFLICT (name) DO NOTHING`); err != nil {
		return err
	}

	if _, err := s.db.Exec(ctx, `
		INSERT INTO budget_rules (category_id, period, limit_amount)
		SELECT id, 'daily', 60000
		FROM categories
		WHERE name = 'Makan'
		ON CONFLICT (category_id, period) DO NOTHING`); err != nil {
		return err
	}

	if _, err := s.db.Exec(ctx, `
		INSERT INTO budget_rules (category_id, period, limit_amount)
		SELECT id, 'monthly', 240000
		FROM categories
		WHERE name = 'Bensin'
		ON CONFLICT (category_id, period) DO NOTHING`); err != nil {
		return err
	}

	if _, err := s.db.Exec(ctx, `
		INSERT INTO user_profiles (full_name, email, password_hash, avatar_initials)
		VALUES ('Rasa Saufar', $1, $2, 'RS')
		ON CONFLICT (email) DO NOTHING`,
		hardcodedEmail,
		hardcodedPassword,
	); err != nil {
		return err
	}

	return nil
}

func (s *store) listTransactionsDB(ctx context.Context) ([]dbTransaction, error) {
	rows, err := s.db.Query(ctx, `
		SELECT
			t.id,
			t.type,
			c.id,
			c.name,
			t.amount,
			t.trx_date,
			t.note
		FROM transactions t
		JOIN categories c ON c.id = t.category_id
		ORDER BY t.trx_date DESC, t.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]dbTransaction, 0)
	for rows.Next() {
		item := dbTransaction{}
		if err := rows.Scan(
			&item.ID,
			&item.Type,
			&item.CategoryID,
			&item.Category,
			&item.Amount,
			&item.Date,
			&item.Note,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (s *store) listCategoriesDB(ctx context.Context) ([]Category, error) {
	rows, err := s.db.Query(ctx, `SELECT id, name FROM categories ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Category, 0)
	for rows.Next() {
		item := Category{}
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (s *store) listBudgetRulesDB(ctx context.Context) ([]dbBudgetRule, error) {
	rows, err := s.db.Query(ctx, `
		SELECT
			br.id,
			c.id,
			c.name,
			br.period,
			br.limit_amount
		FROM budget_rules br
		JOIN categories c ON c.id = br.category_id
		ORDER BY c.name ASC, br.period ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]dbBudgetRule, 0)
	for rows.Next() {
		item := dbBudgetRule{}
		if err := rows.Scan(
			&item.ID,
			&item.CategoryID,
			&item.Category,
			&item.Period,
			&item.Limit,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (s *store) listSalaryMastersDB(ctx context.Context) ([]SalaryMaster, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, month, amount, note
		FROM salary_masters
		ORDER BY month DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]SalaryMaster, 0)
	for rows.Next() {
		item := SalaryMaster{}
		if err := rows.Scan(&item.ID, &item.Month, &item.Amount, &item.Note); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (s *store) findCategoryByName(ctx context.Context, name string) (Category, error) {
	item := Category{}
	err := s.db.QueryRow(
		ctx,
		`SELECT id, name FROM categories WHERE LOWER(name) = LOWER($1) ORDER BY id LIMIT 1`,
		strings.TrimSpace(name),
	).Scan(&item.ID, &item.Name)
	if err != nil {
		return Category{}, err
	}
	return item, nil
}

func (s *store) findCategoryByID(ctx context.Context, id int64) (Category, error) {
	item := Category{}
	err := s.db.QueryRow(ctx, `SELECT id, name FROM categories WHERE id = $1`, id).Scan(&item.ID, &item.Name)
	if err != nil {
		return Category{}, err
	}
	return item, nil
}

func (s *store) categoryNameExists(ctx context.Context, name string, exceptID int64) (bool, error) {
	var exists bool
	err := s.db.QueryRow(
		ctx,
		`SELECT EXISTS (
			SELECT 1 FROM categories
			WHERE LOWER(name) = LOWER($1)
			  AND ($2 = 0 OR id <> $2)
		)`,
		strings.TrimSpace(name),
		exceptID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *store) findUserProfileByEmail(ctx context.Context, email string) (dbUserProfile, error) {
	item := dbUserProfile{}
	err := s.db.QueryRow(
		ctx,
		`SELECT id, full_name, email
		 FROM user_profiles
		 WHERE LOWER(email) = LOWER($1)
		 ORDER BY id
		 LIMIT 1`,
		strings.TrimSpace(email),
	).Scan(&item.ID, &item.FullName, &item.Email)
	if err != nil {
		return dbUserProfile{}, err
	}
	return item, nil
}

func (s *store) updateUserProfile(ctx context.Context, currentEmail, newName, newEmail, newPassword string) (dbUserProfile, error) {
	var item dbUserProfile

	if newPassword != "" {
		err := s.db.QueryRow(
			ctx,
			`UPDATE user_profiles
			 SET full_name = $1,
			     email = $2,
			     password_hash = $3,
			     updated_at = NOW()
			 WHERE LOWER(email) = LOWER($4)
			 RETURNING id, full_name, email`,
			newName, newEmail, newPassword, currentEmail,
		).Scan(&item.ID, &item.FullName, &item.Email)
		if err != nil {
			return dbUserProfile{}, err
		}
	} else {
		err := s.db.QueryRow(
			ctx,
			`UPDATE user_profiles
			 SET full_name = $1,
			     email = $2,
			     updated_at = NOW()
			 WHERE LOWER(email) = LOWER($3)
			 RETURNING id, full_name, email`,
			newName, newEmail, currentEmail,
		).Scan(&item.ID, &item.FullName, &item.Email)
		if err != nil {
			return dbUserProfile{}, err
		}
	}

	return item, nil
}

func (s *store) salaryForMonth(ctx context.Context, month string) (int64, error) {
	var total int64
	err := s.db.QueryRow(
		ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM salary_masters WHERE month = $1`,
		month,
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (s *store) salaryToMonth(ctx context.Context, month string) (int64, error) {
	var total int64
	err := s.db.QueryRow(
		ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM salary_masters WHERE month <= $1`,
		month,
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (s *store) checkBudgetForTransaction(ctx context.Context, categoryID int64, trxType string, referenceDate time.Time, additionalAmount int64, excludeTransactionID int64) ([]BudgetCheck, error) {
	if trxType != "expense" {
		return []BudgetCheck{}, nil
	}

	rules, err := s.listBudgetRulesByCategoryID(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	checks := make([]BudgetCheck, 0, len(rules))
	for _, rule := range rules {
		used, err := s.sumExpenseForRule(ctx, categoryID, rule.Period, referenceDate, excludeTransactionID)
		if err != nil {
			return nil, err
		}
		used += additionalAmount

		checks = append(checks, buildBudgetCheck(rule.Category, rule.Period, rule.Limit, used))
	}

	sort.Slice(checks, func(i, j int) bool {
		return periodRank(checks[i].Period) < periodRank(checks[j].Period)
	})

	return checks, nil
}

func (s *store) listBudgetRulesByCategoryID(ctx context.Context, categoryID int64) ([]dbBudgetRule, error) {
	rows, err := s.db.Query(ctx, `
		SELECT
			br.id,
			c.id,
			c.name,
			br.period,
			br.limit_amount
		FROM budget_rules br
		JOIN categories c ON c.id = br.category_id
		WHERE br.category_id = $1`, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]dbBudgetRule, 0)
	for rows.Next() {
		item := dbBudgetRule{}
		if err := rows.Scan(&item.ID, &item.CategoryID, &item.Category, &item.Period, &item.Limit); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (s *store) sumExpenseForRule(ctx context.Context, categoryID int64, period string, referenceDate time.Time, excludeTransactionID int64) (int64, error) {
	var query string
	switch period {
	case "daily":
		query = `
			SELECT COALESCE(SUM(amount), 0)
			FROM transactions
			WHERE type = 'expense'
			  AND category_id = $1
			  AND ($2 = 0 OR id <> $2)
			  AND trx_date = $3::date`
	case "weekly":
		query = `
			SELECT COALESCE(SUM(amount), 0)
			FROM transactions
			WHERE type = 'expense'
			  AND category_id = $1
			  AND ($2 = 0 OR id <> $2)
			  AND DATE_TRUNC('week', trx_date) = DATE_TRUNC('week', $3::date)`
	case "monthly":
		query = `
			SELECT COALESCE(SUM(amount), 0)
			FROM transactions
			WHERE type = 'expense'
			  AND category_id = $1
			  AND ($2 = 0 OR id <> $2)
			  AND DATE_TRUNC('month', trx_date) = DATE_TRUNC('month', $3::date)`
	default:
		return 0, errors.New("periode tidak valid")
	}

	var total int64
	err := s.db.QueryRow(ctx, query, categoryID, excludeTransactionID, referenceDate).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func buildBudgetCheck(category, period string, limit, used int64) BudgetCheck {
	remaining := limit - used
	percentage := 0.0
	if limit > 0 {
		percentage = (float64(used) / float64(limit)) * 100
	}

	return BudgetCheck{
		Category:   category,
		Period:     period,
		Limit:      limit,
		Used:       used,
		Remaining:  remaining,
		Percentage: percentage,
		OverBudget: used > limit,
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
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

func normalizeSalaryMasterInput(input salaryMasterInput) (SalaryMaster, error) {
	month := strings.TrimSpace(input.Month)
	if _, err := time.Parse(monthLayout, month); err != nil {
		return SalaryMaster{}, errors.New("bulan harus berformat YYYY-MM")
	}

	if input.Amount <= 0 {
		return SalaryMaster{}, errors.New("nominal gaji harus lebih dari 0")
	}

	return SalaryMaster{
		Month:  month,
		Amount: input.Amount,
		Note:   strings.TrimSpace(input.Note),
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

func writeInternalServerError(w http.ResponseWriter, err error) {
	log.Printf("internal server error: %v", err)
	writeError(w, http.StatusInternalServerError, "terjadi kesalahan internal")
}
