// Package validate contains input normalization and validation helpers.
package validate

import (
	"errors"
	"strings"
	"time"

	"github.com/rasasaufar/finance-app/api/internal/types"
)

func TransactionInput(input types.TransactionInput) (types.Transaction, error) {
	transactionType := strings.ToLower(strings.TrimSpace(input.Type))
	if transactionType != "income" && transactionType != "expense" {
		return types.Transaction{}, errors.New("tipe transaksi harus income atau expense")
	}

	category := strings.TrimSpace(input.Category)
	if category == "" {
		return types.Transaction{}, errors.New("kategori wajib diisi")
	}

	if input.Amount <= 0 {
		return types.Transaction{}, errors.New("nominal harus lebih dari 0")
	}

	dateValue := strings.TrimSpace(input.Date)
	if _, err := time.Parse(types.DateLayout, dateValue); err != nil {
		return types.Transaction{}, errors.New("tanggal harus berformat YYYY-MM-DD")
	}

	return types.Transaction{
		Type:     transactionType,
		Category: category,
		Amount:   input.Amount,
		Date:     dateValue,
		Note:     strings.TrimSpace(input.Note),
	}, nil
}

func BudgetRuleInput(input types.BudgetRuleInput) (types.BudgetRule, error) {
	category := strings.TrimSpace(input.Category)
	if category == "" {
		return types.BudgetRule{}, errors.New("kategori wajib diisi")
	}

	period := strings.ToLower(strings.TrimSpace(input.Period))
	if period != "daily" && period != "weekly" && period != "monthly" {
		return types.BudgetRule{}, errors.New("periode harus daily, weekly, atau monthly")
	}

	if input.Limit <= 0 {
		return types.BudgetRule{}, errors.New("limit budget harus lebih dari 0")
	}

	return types.BudgetRule{
		Category: category,
		Period:   period,
		Limit:    input.Limit,
	}, nil
}

func SalaryMasterInput(input types.SalaryMasterInput) (types.SalaryMaster, error) {
	month := strings.TrimSpace(input.Month)
	if _, err := time.Parse(types.MonthLayout, month); err != nil {
		return types.SalaryMaster{}, errors.New("bulan harus berformat YYYY-MM")
	}

	if input.Amount <= 0 {
		return types.SalaryMaster{}, errors.New("nominal gaji harus lebih dari 0")
	}

	return types.SalaryMaster{
		Month:  month,
		Amount: input.Amount,
		Note:   strings.TrimSpace(input.Note),
	}, nil
}
