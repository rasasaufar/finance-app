package handler

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/rasasaufar/finance-app/api/internal/httputil"
	"github.com/rasasaufar/finance-app/api/internal/middleware"
	"github.com/rasasaufar/finance-app/api/internal/types"
	"github.com/rasasaufar/finance-app/api/internal/validate"
)

func (h *Handler) HandleGetBudgetRules(w http.ResponseWriter, r *http.Request) {
	accountID := middleware.UserIDFromContext(r.Context())
	rules, err := h.Store.ListBudgetRules(r.Context(), accountID)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}

	response := make([]types.BudgetRule, 0, len(rules))
	for _, rule := range rules {
		response = append(response, types.BudgetRule{
			ID:       rule.ID,
			Category: rule.Category,
			Period:   rule.Period,
			Limit:    rule.Limit,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) HandleCreateBudgetRule(w http.ResponseWriter, r *http.Request) {
	var input types.BudgetRuleInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	rule, err := validate.BudgetRuleInput(input)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()
	accountID := middleware.UserIDFromContext(ctx)

	category, err := h.Store.FindCategoryByName(ctx, accountID, rule.Category)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusBadRequest, "kategori tidak ditemukan")
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	created := types.BudgetRule{Category: category.Name, Period: rule.Period, Limit: rule.Limit}
	err = h.Store.DB.QueryRow(
		ctx,
		`INSERT INTO budget_rules (account_id, category_id, period, limit_amount)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		accountID,
		category.ID,
		rule.Period,
		rule.Limit,
	).Scan(&created.ID)
	if err != nil {
		if isUniqueViolation(err) {
			httputil.WriteError(w, http.StatusBadRequest, "budget rule untuk kategori dan periode tersebut sudah ada")
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, created)
}

func (h *Handler) HandleUpdateBudgetRule(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseIDParam(r, "id")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "id budget rule tidak valid")
		return
	}

	var input types.BudgetRuleInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	normalized, err := validate.BudgetRuleInput(input)
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

	updated := types.BudgetRule{ID: id, Category: category.Name, Period: normalized.Period, Limit: normalized.Limit}
	err = h.Store.DB.QueryRow(
		ctx,
		`UPDATE budget_rules
		 SET category_id = $1,
		     period = $2,
		     limit_amount = $3
		 WHERE id = $4 AND account_id = $5
		 RETURNING id`,
		category.ID,
		normalized.Period,
		normalized.Limit,
		id,
		accountID,
	).Scan(&updated.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "budget rule tidak ditemukan")
			return
		}
		if isUniqueViolation(err) {
			httputil.WriteError(w, http.StatusBadRequest, "budget rule untuk kategori dan periode tersebut sudah ada")
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, updated)
}

func (h *Handler) HandleDeleteBudgetRule(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseIDParam(r, "id")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "id budget rule tidak valid")
		return
	}

	accountID := middleware.UserIDFromContext(r.Context())
	result, err := h.Store.DB.Exec(r.Context(),
		`DELETE FROM budget_rules WHERE id = $1 AND account_id = $2`,
		id, accountID)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}
	if result.RowsAffected() == 0 {
		httputil.WriteError(w, http.StatusNotFound, "budget rule tidak ditemukan")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
