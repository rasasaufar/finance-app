package handler

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/rasasaufar/finance-app/api/internal/httputil"
	"github.com/rasasaufar/finance-app/api/internal/types"
	"github.com/rasasaufar/finance-app/api/internal/validate"
)

func (h *Handler) HandleGetSalaryMasters(w http.ResponseWriter, r *http.Request) {
	items, err := h.Store.ListSalaryMasters(r.Context())
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) HandleCreateSalaryMaster(w http.ResponseWriter, r *http.Request) {
	var input types.SalaryMasterInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := validate.SalaryMasterInput(input)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	created := types.SalaryMaster{}
	err = h.Store.DB.QueryRow(
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
			httputil.WriteError(w, http.StatusBadRequest, "master gaji bulan tersebut sudah ada")
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, created)
}

func (h *Handler) HandleUpdateSalaryMaster(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseIDParam(r, "id")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "id master gaji tidak valid")
		return
	}

	var input types.SalaryMasterInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	normalized, err := validate.SalaryMasterInput(input)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	updated := types.SalaryMaster{}
	err = h.Store.DB.QueryRow(
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
			httputil.WriteError(w, http.StatusNotFound, "master gaji tidak ditemukan")
			return
		}
		if isUniqueViolation(err) {
			httputil.WriteError(w, http.StatusBadRequest, "master gaji bulan tersebut sudah ada")
			return
		}
		httputil.WriteInternalServerError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, updated)
}

func (h *Handler) HandleDeleteSalaryMaster(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseIDParam(r, "id")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "id master gaji tidak valid")
		return
	}

	result, err := h.Store.DB.Exec(r.Context(), `DELETE FROM salary_masters WHERE id = $1`, id)
	if err != nil {
		httputil.WriteInternalServerError(w, err)
		return
	}
	if result.RowsAffected() == 0 {
		httputil.WriteError(w, http.StatusNotFound, "master gaji tidak ditemukan")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
