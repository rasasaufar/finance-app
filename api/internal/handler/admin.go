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

// HandleListAccounts GET /admin/accounts
func (h *Handler) HandleListAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.Store.ListAccounts(r.Context())
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}

	response := make([]types.AccountResponse, 0, len(accounts))
	for _, a := range accounts {
		response = append(response, types.AccountResponse{
			ID:    a.ID,
			Name:  a.FullName,
			Email: a.Email,
			Role:  a.Role,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, response)
}

// HandleCreateAccount POST /admin/accounts
func (h *Handler) HandleCreateAccount(w http.ResponseWriter, r *http.Request) {
	var input types.AccountInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "nama tidak boleh kosong")
		return
	}

	email := strings.TrimSpace(input.Email)
	if email == "" {
		httputil.WriteError(w, http.StatusBadRequest, "email tidak boleh kosong")
		return
	}

	password := strings.TrimSpace(input.Password)
	if password == "" {
		httputil.WriteError(w, http.StatusBadRequest, "password tidak boleh kosong")
		return
	}

	role := strings.TrimSpace(input.Role)
	if role != types.RoleAdmin && role != types.RoleUser {
		role = types.RoleUser
	}

	ctx := r.Context()
	account, err := h.Store.CreateAccount(ctx, name, email, password, role)
	if err != nil {
		if isUniqueViolation(err) {
			httputil.WriteError(w, http.StatusBadRequest, "email sudah digunakan")
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	// Seed default categories for the new account
	if err := h.Store.SeedDefaultCategoriesForAccount(ctx, account.ID); err != nil {
		// Non-fatal: log but don't fail the request
		_ = err
	}

	httputil.WriteJSON(w, http.StatusCreated, types.AccountResponse{
		ID:    account.ID,
		Name:  account.FullName,
		Email: account.Email,
		Role:  account.Role,
	})
}

// HandleUpdateAccount PUT /admin/accounts/{id}
func (h *Handler) HandleUpdateAccount(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseIDParam(r, "id")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "id akun tidak valid")
		return
	}

	// Prevent admin from editing their own account via this endpoint
	callerID := middleware.UserIDFromContext(r.Context())
	if callerID == id {
		httputil.WriteError(w, http.StatusBadRequest, "gunakan halaman Pengaturan Akun untuk mengubah profil sendiri")
		return
	}

	var input types.AccountUpdateInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "nama tidak boleh kosong")
		return
	}

	email := strings.TrimSpace(input.Email)
	if email == "" {
		httputil.WriteError(w, http.StatusBadRequest, "email tidak boleh kosong")
		return
	}

	role := strings.TrimSpace(input.Role)
	if role != types.RoleAdmin && role != types.RoleUser {
		role = types.RoleUser
	}

	updated, err := h.Store.UpdateAccount(r.Context(), id, name, email, strings.TrimSpace(input.Password), role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "akun tidak ditemukan")
			return
		}
		if isUniqueViolation(err) {
			httputil.WriteError(w, http.StatusBadRequest, "email sudah digunakan")
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, types.AccountResponse{
		ID:    updated.ID,
		Name:  updated.FullName,
		Email: updated.Email,
		Role:  updated.Role,
	})
}

// HandleDeleteAccount DELETE /admin/accounts/{id}
func (h *Handler) HandleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseIDParam(r, "id")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "id akun tidak valid")
		return
	}

	// Prevent admin from deleting their own account
	callerID := middleware.UserIDFromContext(r.Context())
	if callerID == id {
		httputil.WriteError(w, http.StatusBadRequest, "tidak bisa menghapus akun sendiri")
		return
	}

	if err := h.Store.DeleteAccount(r.Context(), id); err != nil {
		if err.Error() == "akun tidak ditemukan" {
			httputil.WriteError(w, http.StatusNotFound, "akun tidak ditemukan")
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
