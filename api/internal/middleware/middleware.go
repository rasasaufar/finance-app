package middleware

import (
	"net/http"
	"strings"

	"github.com/rasasaufar/finance-app/api/internal/httputil"
	"github.com/rasasaufar/finance-app/api/internal/types"
)

// Auth validates the Bearer token on protected routes.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(authorization, "Bearer ") {
			httputil.WriteError(w, http.StatusUnauthorized, "token tidak valid")
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
		if token != types.DummyToken {
			httputil.WriteError(w, http.StatusUnauthorized, "token tidak valid")
			return
		}

		next.ServeHTTP(w, r)
	})
}
