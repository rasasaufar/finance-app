package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/rasasaufar/finance-app/api/internal/httputil"
	"github.com/rasasaufar/finance-app/api/internal/middleware"
	"github.com/rasasaufar/finance-app/api/internal/types"
)

func (h *Handler) HandleGetCategories(w http.ResponseWriter, r *http.Request) {
	accountID := middleware.UserIDFromContext(r.Context())
	categories, err := h.Store.ListCategories(r.Context(), accountID)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, categories)
}

func (h *Handler) HandleCreateCategory(w http.ResponseWriter, r *http.Request) {
	var input types.CategoryInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "nama kategori wajib diisi")
		return
	}

	ctx := r.Context()
	accountID := middleware.UserIDFromContext(ctx)

	exists, err := h.Store.CategoryNameExists(ctx, accountID, name, 0)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}
	if exists {
		httputil.WriteError(w, http.StatusBadRequest, "nama kategori sudah ada")
		return
	}

	category := types.Category{}
	err = h.Store.DB.QueryRow(
		ctx,
		`INSERT INTO categories (account_id, name) VALUES ($1, $2) RETURNING id, name`,
		accountID, name,
	).Scan(&category.ID, &category.Name)
	if err != nil {
		if isUniqueViolation(err) {
			httputil.WriteError(w, http.StatusBadRequest, "nama kategori sudah ada")
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, category)
}

func (h *Handler) HandleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseIDParam(r, "id")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "id kategori tidak valid")
		return
	}

	var input types.CategoryInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "nama kategori wajib diisi")
		return
	}

	ctx := r.Context()
	accountID := middleware.UserIDFromContext(ctx)

	exists, err := h.Store.CategoryNameExists(ctx, accountID, name, id)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}
	if exists {
		httputil.WriteError(w, http.StatusBadRequest, "nama kategori sudah ada")
		return
	}

	category := types.Category{}
	err = h.Store.DB.QueryRow(
		ctx,
		`UPDATE categories SET name = $1 WHERE id = $2 AND account_id = $3 RETURNING id, name`,
		name, id, accountID,
	).Scan(&category.ID, &category.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "kategori tidak ditemukan")
			return
		}
		if isUniqueViolation(err) {
			httputil.WriteError(w, http.StatusBadRequest, "nama kategori sudah ada")
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, category)
}

func (h *Handler) HandleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseIDParam(r, "id")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "id kategori tidak valid")
		return
	}

	ctx := r.Context()
	accountID := middleware.UserIDFromContext(ctx)

	category, err := h.Store.FindCategoryByID(ctx, accountID, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "kategori tidak ditemukan")
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	var usageCount int64
	err = h.Store.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM transactions WHERE category_id = $1 AND account_id = $2`,
		id, accountID).Scan(&usageCount)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}
	if usageCount > 0 {
		httputil.WriteError(w, http.StatusBadRequest, "kategori masih digunakan transaksi")
		return
	}

	result, err := h.Store.DB.Exec(ctx,
		`DELETE FROM categories WHERE id = $1 AND account_id = $2`,
		category.ID, accountID)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}
	if result.RowsAffected() == 0 {
		httputil.WriteError(w, http.StatusNotFound, "kategori tidak ditemukan")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
