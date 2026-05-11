package handler

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/rasasaufar/finance-app/api/internal/httputil"
	"github.com/rasasaufar/finance-app/api/internal/store"
	"github.com/rasasaufar/finance-app/api/internal/types"
)

func (h *Handler) HandleDashboardSummary(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	today := now.Format(types.DateLayout)
	month := now.Format(types.MonthLayout)
	ctx := r.Context()

	transactions, err := h.Store.ListTransactions(ctx)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}

	rules, err := h.Store.ListBudgetRules(ctx)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}

	salaryCurrentMonth, err := h.Store.SalaryForMonth(ctx, month)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}

	salaryTotalToDate, err := h.Store.SalaryToMonth(ctx, month)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
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
			if !strings.EqualFold(trx.Category, "Gaji") {
				additionalIncome += trx.Amount
			}
			continue
		}

		totalExpense += trx.Amount
		trxDate := trx.Date.Format(types.DateLayout)
		if trxDate == today {
			todayExpense += trx.Amount
		}
		if trx.Date.Format(types.MonthLayout) == month {
			monthExpense += trx.Amount
			categoryKey := strings.ToLower(strings.TrimSpace(trx.Category))
			monthExpenseByCategory[categoryKey] += trx.Amount
			if _, exists := monthExpenseLabelByCategory[categoryKey]; !exists {
				monthExpenseLabelByCategory[categoryKey] = strings.TrimSpace(trx.Category)
			}
		}
	}

	budgetUsage := make([]types.BudgetCheck, 0, len(rules))
	budgetRuleCategorySet := map[string]struct{}{}
	var bensinMonthlyRuleLimit int64
	var dailyMakanRule *types.DBBudgetRule

	for _, rule := range rules {
		ruleCategoryKey := strings.ToLower(strings.TrimSpace(rule.Category))
		if ruleCategoryKey != "" {
			budgetRuleCategorySet[ruleCategoryKey] = struct{}{}
		}
		if strings.EqualFold(rule.Category, "Bensin") && rule.Period == "monthly" {
			bensinMonthlyRuleLimit += rule.Limit
		}

		used, err := h.Store.SumExpenseForRule(ctx, rule.CategoryID, rule.Period, now, 0)
		if err != nil {
			httputil.WriteInternalServerError(w, err)
			return
		}

		check := store.BuildBudgetCheck(rule.Category, rule.Period, rule.Limit, used)
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

		budgetUsage = append(budgetUsage, types.BudgetCheck{
			Category:   categoryLabel,
			Period:     "monthly",
			Limit:      otherCategoryLimit,
			Used:       used,
			Remaining:  otherCategoryLimit - used,
			Percentage: percentage,
			OverBudget: used > otherCategoryLimit,
		})
	}

	var makanToday *types.BudgetCheck
	var bensinMonth *types.BudgetCheck
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

	var makanMonth *types.BudgetCheck
	if dailyMakanRule != nil {
		daysCount := int64(daysInMonth(now))
		monthlyLimit := dailyMakanRule.Limit * daysCount
		monthlyUsed := monthExpenseByCategory[strings.ToLower(strings.TrimSpace(dailyMakanRule.Category))]
		percentage := 0.0
		if monthlyLimit > 0 {
			percentage = (float64(monthlyUsed) / float64(monthlyLimit)) * 100
		}

		check := types.BudgetCheck{
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

	httputil.WriteJSON(w, http.StatusOK, types.DashboardSummaryResponse{
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

func (h *Handler) HandleMonthlyReport(w http.ResponseWriter, r *http.Request) {
	month := strings.TrimSpace(r.URL.Query().Get("month"))
	if month == "" {
		month = time.Now().Format(types.MonthLayout)
	}

	reference, err := time.Parse(types.MonthLayout, month)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "format bulan harus YYYY-MM")
		return
	}

	transactions, err := h.Store.ListTransactions(r.Context())
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}

	rules, err := h.Store.ListBudgetRules(r.Context())
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}

	var totalIncome int64
	var totalExpense int64
	// key: lowercase category name → total expense for the month
	byCategoryMap := map[string]int64{}
	// key: lowercase category name → original display name
	categoryDisplayName := map[string]string{}

	for _, trx := range transactions {
		if trx.Date.Format(types.MonthLayout) != month {
			continue
		}

		if trx.Type == "income" {
			totalIncome += trx.Amount
			continue
		}

		totalExpense += trx.Amount
		key := strings.ToLower(strings.TrimSpace(trx.Category))
		byCategoryMap[key] += trx.Amount
		if _, exists := categoryDisplayName[key]; !exists {
			categoryDisplayName[key] = strings.TrimSpace(trx.Category)
		}
	}

	spendingByCategory := make([]types.CategorySpending, 0, len(byCategoryMap))
	for key, total := range byCategoryMap {
		name := categoryDisplayName[key]
		if name == "" {
			name = key
		}
		spendingByCategory = append(spendingByCategory, types.CategorySpending{Category: name, Total: total})
	}
	sort.Slice(spendingByCategory, func(i, j int) bool {
		return spendingByCategory[i].Total > spendingByCategory[j].Total
	})

	daysInRef := int64(daysInMonth(reference))

	budgetUsage := make([]types.BudgetCheck, 0, len(rules))
	for _, rule := range rules {
		var used int64
		effectiveLimit := rule.Limit

		if rule.Period == "monthly" {
			// For monthly rules, query the DB directly (already sums the whole month)
			used, err = h.Store.SumExpenseForRule(r.Context(), rule.CategoryID, rule.Period, reference, 0)
			if err != nil {
				httputil.WriteInternalServerError(w, err)
				return
			}
		} else {
			// For daily/weekly rules in a monthly report, sum all expenses in the month
			used = byCategoryMap[strings.ToLower(strings.TrimSpace(rule.Category))]
			if rule.Period == "daily" {
				// Scale daily limit to the full month
				effectiveLimit = rule.Limit * daysInRef
			}
		}
		budgetUsage = append(budgetUsage, store.BuildBudgetCheck(rule.Category, rule.Period, effectiveLimit, used))
	}

	httputil.WriteJSON(w, http.StatusOK, types.MonthlyReportResponse{
		Month:              month,
		TotalIncome:        totalIncome,
		TotalExpense:       totalExpense,
		Net:                totalIncome - totalExpense,
		SpendingByCategory: spendingByCategory,
		BudgetUsage:        budgetUsage,
	})
}

func daysInMonth(source time.Time) int {
	return time.Date(source.Year(), source.Month()+1, 0, 0, 0, 0, 0, source.Location()).Day()
}
