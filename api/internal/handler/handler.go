// Package handler contains all HTTP handlers grouped by domain.
package handler

import (
	"github.com/rasasaufar/finance-app/api/internal/store"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	Store *store.Store
}

func New(s *store.Store) *Handler {
	return &Handler{Store: s}
}
