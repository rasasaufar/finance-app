package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rasasaufar/finance-app/api/internal/httputil"
	"github.com/rasasaufar/finance-app/api/internal/middleware"
	"github.com/rasasaufar/finance-app/api/internal/types"
	"github.com/rasasaufar/finance-app/api/internal/validate"
)

func (h *Handler) HandleGetTransactions(w http.ResponseWriter, r *http.Request) {
	accountID := middleware.UserIDFromContext(r.Context())
	items, err := h.Store.ListTransactions(r.Context(), accountID)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}

	transactions := make([]types.Transaction, 0, len(items))
	for _, item := range items {
		transactions = append(transactions, types.Transaction{
			ID:       item.ID,
			Type:     item.Type,
			Category: item.Category,
			Amount:   item.Amount,
			Date:     item.Date.Format(types.DateLayout),
			Note:     item.Note,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, transactions)
}

func (h *Handler) HandleCreateTransaction(w http.ResponseWriter, r *http.Request) {
	var input types.TransactionInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	trx, err := validate.TransactionInput(input)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()
	accountID := middleware.UserIDFromContext(ctx)

	category, err := h.Store.FindCategoryByName(ctx, accountID, trx.Category)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusBadRequest, "kategori tidak ditemukan")
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	trxDate, _ := time.Parse(types.DateLayout, trx.Date)
	checks, err := h.Store.CheckBudgetForTransaction(ctx, accountID, category.ID, trx.Type, trxDate, trx.Amount, 0)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}

	var id int64
	err = h.Store.DB.QueryRow(
		ctx,
		`INSERT INTO transactions (account_id, type, category_id, amount, trx_date, note)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id`,
		accountID,
		trx.Type,
		category.ID,
		trx.Amount,
		trxDate,
		trx.Note,
	).Scan(&id)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}

	created := types.Transaction{
		ID:       id,
		Type:     trx.Type,
		Category: category.Name,
		Amount:   trx.Amount,
		Date:     trxDate.Format(types.DateLayout),
		Note:     trx.Note,
	}

	httputil.WriteJSON(w, http.StatusCreated, types.TransactionMutationResponse{
		Transaction:  created,
		BudgetChecks: checks,
		Warning:      anyOverBudget(checks),
	})
}

func (h *Handler) HandleUpdateTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseIDParam(r, "id")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "id transaksi tidak valid")
		return
	}

	var input types.TransactionInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	normalized, err := validate.TransactionInput(input)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()
	accountID := middleware.UserIDFromContext(ctx)

	category, err := h.Store.FindCategoryByName(ctx, accountID, normalized.Category)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusBadRequest, "kategori tidak ditemukan")
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	trxDate, _ := time.Parse(types.DateLayout, normalized.Date)
	checks, err := h.Store.CheckBudgetForTransaction(ctx, accountID, category.ID, normalized.Type, trxDate, normalized.Amount, id)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}

	var updatedID int64
	err = h.Store.DB.QueryRow(
		ctx,
		`UPDATE transactions
		 SET type = $1,
		     category_id = $2,
		     amount = $3,
		     trx_date = $4,
		     note = $5,
		     updated_at = NOW()
		 WHERE id = $6 AND account_id = $7
		 RETURNING id`,
		normalized.Type,
		category.ID,
		normalized.Amount,
		trxDate,
		normalized.Note,
		id,
		accountID,
	).Scan(&updatedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "transaksi tidak ditemukan")
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	updated := types.Transaction{
		ID:       updatedID,
		Type:     normalized.Type,
		Category: category.Name,
		Amount:   normalized.Amount,
		Date:     trxDate.Format(types.DateLayout),
		Note:     normalized.Note,
	}

	httputil.WriteJSON(w, http.StatusOK, types.TransactionMutationResponse{
		Transaction:  updated,
		BudgetChecks: checks,
		Warning:      anyOverBudget(checks),
	})
}

func (h *Handler) HandleDeleteTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseIDParam(r, "id")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "id transaksi tidak valid")
		return
	}

	accountID := middleware.UserIDFromContext(r.Context())
	result, err := h.Store.DB.Exec(r.Context(),
		`DELETE FROM transactions WHERE id = $1 AND account_id = $2`,
		id, accountID)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}
	if result.RowsAffected() == 0 {
		httputil.WriteError(w, http.StatusNotFound, "transaksi tidak ditemukan")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- helpers ---

func anyOverBudget(checks []types.BudgetCheck) bool {
	for _, check := range checks {
		if check.OverBudget {
			return true
		}
	}
	return false
}

func isUniqueViolation(err error) bool {
	type pgError interface {
		SQLState() string
	}
	var pgErr pgError
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23505"
	}
	return strings.Contains(err.Error(), "23505")
}
