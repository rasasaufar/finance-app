package httputil

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

func WriteInternalServerError(w http.ResponseWriter, err error) {
	log.Printf("internal server error: %v", err)
	WriteError(w, http.StatusInternalServerError, "terjadi kesalahan internal")
}

func DecodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("payload tidak valid: %w", err)
	}
	return nil
}

func ParseIDParam(r *http.Request, key string) (int64, error) {
	idStr := strings.TrimSpace(chi.URLParam(r, key))
	if idStr == "" {
		return 0, fmt.Errorf("id kosong")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("id tidak valid")
	}
	return id, nil
}
