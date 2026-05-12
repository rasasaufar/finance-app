package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/rasasaufar/finance-app/api/internal/httputil"
	"github.com/rasasaufar/finance-app/api/internal/middleware"
	"github.com/rasasaufar/finance-app/api/internal/types"
)

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var input types.LoginRequest
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	email := strings.TrimSpace(input.Email)
	password := strings.TrimSpace(input.Password)

	account, err := h.Store.FindAccountByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusUnauthorized, "email atau password salah")
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	if account.PasswordHash != password {
		httputil.WriteError(w, http.StatusUnauthorized, "email atau password salah")
		return
	}

	token, err := generateToken()
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}

	if err := h.Store.CreateSession(r.Context(), token, account.ID); err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, types.LoginResponse{
		Token: token,
		User: types.User{
			Name:  account.FullName,
			Email: account.Email,
			Role:  account.Role,
		},
	})
}

func (h *Handler) HandleMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	account, err := h.Store.FindAccountByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "akun tidak ditemukan")
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, types.User{
		Name:  account.FullName,
		Email: account.Email,
		Role:  account.Role,
	})
}

func (h *Handler) HandleUpdateMe(w http.ResponseWriter, r *http.Request) {
	var input types.UpdateProfileInput
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

	userID := middleware.UserIDFromContext(r.Context())
	updated, err := h.Store.UpdateSelfProfile(r.Context(), userID, name, email, strings.TrimSpace(input.Password))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "akun tidak ditemukan")
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, types.User{
		Name:  updated.FullName,
		Email: updated.Email,
		Role:  updated.Role,
	})
}
