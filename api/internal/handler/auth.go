package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/rasasaufar/finance-app/api/internal/httputil"
	"github.com/rasasaufar/finance-app/api/internal/types"
)

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var input types.LoginRequest
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if strings.TrimSpace(input.Email) != types.HardcodedEmail || strings.TrimSpace(input.Password) != types.HardcodedPassword {
		httputil.WriteError(w, http.StatusUnauthorized, "email atau password salah")
		return
	}

	profileName := "Rasas"
	if profile, err := h.Store.FindUserProfileByEmail(r.Context(), types.HardcodedEmail); err == nil {
		profileName = profile.FullName
	}

	httputil.WriteJSON(w, http.StatusOK, types.LoginResponse{
		Token: types.DummyToken,
		User: types.User{
			Name:  profileName,
			Email: types.HardcodedEmail,
		},
	})
}

func (h *Handler) HandleMe(w http.ResponseWriter, r *http.Request) {
	profile, err := h.Store.FindUserProfileByEmail(r.Context(), types.HardcodedEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteJSON(w, http.StatusOK, types.User{Name: "Rasas", Email: types.HardcodedEmail})
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, types.User{Name: profile.FullName, Email: profile.Email})
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

	ctx := r.Context()
	updated, err := h.Store.UpdateUserProfile(ctx, types.HardcodedEmail, name, email, strings.TrimSpace(input.Password))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "profil tidak ditemukan")
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, types.User{Name: updated.FullName, Email: updated.Email})
}
