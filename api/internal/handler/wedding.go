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

// HandleGetWeddingSummary returns config + all deposits for the current user.
func (h *Handler) HandleGetWeddingSummary(w http.ResponseWriter, r *http.Request) {
	accountID := middleware.UserIDFromContext(r.Context())

	cfg, err := h.Store.GetWeddingConfig(r.Context(), accountID)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}

	deposits, err := h.Store.ListWeddingDeposits(r.Context(), accountID)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, types.WeddingSummary{
		Config:   cfg,
		Deposits: deposits,
	})
}

// HandleUpdateWeddingConfig upserts the wedding config for the current user.
func (h *Handler) HandleUpdateWeddingConfig(w http.ResponseWriter, r *http.Request) {
	var input types.WeddingConfig
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	normalized, err := validate.WeddingConfigInput(input)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	accountID := middleware.UserIDFromContext(r.Context())
	cfg, err := h.Store.UpsertWeddingConfig(r.Context(), accountID, normalized)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, cfg)
}

// HandleCreateWeddingDeposit adds a new deposit.
func (h *Handler) HandleCreateWeddingDeposit(w http.ResponseWriter, r *http.Request) {
	var input types.WeddingDepositInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	normalized, err := validate.WeddingDepositInput(input)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	accountID := middleware.UserIDFromContext(r.Context())
	created, err := h.Store.CreateWeddingDeposit(r.Context(), accountID, normalized)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, created)
}

// HandleUpdateWeddingDeposit updates an existing deposit.
func (h *Handler) HandleUpdateWeddingDeposit(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseIDParam(r, "id")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "id setoran tidak valid")
		return
	}

	var input types.WeddingDepositInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	normalized, err := validate.WeddingDepositInput(input)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	accountID := middleware.UserIDFromContext(r.Context())
	updated, err := h.Store.UpdateWeddingDeposit(r.Context(), accountID, id, normalized)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "setoran tidak ditemukan")
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, updated)
}

// HandleDeleteWeddingDeposit removes a deposit.
func (h *Handler) HandleDeleteWeddingDeposit(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseIDParam(r, "id")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "id setoran tidak valid")
		return
	}

	accountID := middleware.UserIDFromContext(r.Context())
	if err := h.Store.DeleteWeddingDeposit(r.Context(), accountID, id); err != nil {
		if err.Error() == "setoran tidak ditemukan" {
			httputil.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
