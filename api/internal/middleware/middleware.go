package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/rasasaufar/finance-app/api/internal/httputil"
	"github.com/rasasaufar/finance-app/api/internal/store"
	"github.com/rasasaufar/finance-app/api/internal/types"
)

// Auth validates the Bearer token and injects userID + role into the request context.
func Auth(st *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorization := strings.TrimSpace(r.Header.Get("Authorization"))
			if !strings.HasPrefix(authorization, "Bearer ") {
				httputil.WriteError(w, http.StatusUnauthorized, "token tidak valid")
				return
			}

			token := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
			if token == "" {
				httputil.WriteError(w, http.StatusUnauthorized, "token tidak valid")
				return
			}

			account, err := st.FindAccountByToken(r.Context(), token)
			if err != nil {
				httputil.WriteError(w, http.StatusUnauthorized, "token tidak valid atau sudah kadaluarsa")
				return
			}

			ctx := context.WithValue(r.Context(), types.UserIDKey, account.ID)
			ctx = context.WithValue(ctx, types.UserRoleKey, account.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminOnly rejects requests from non-admin users.
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, _ := r.Context().Value(types.UserRoleKey).(string)
		if role != types.RoleAdmin {
			httputil.WriteError(w, http.StatusForbidden, "akses ditolak: hanya admin")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// UserIDFromContext extracts the userID from context (returns 0 if not set).
func UserIDFromContext(ctx context.Context) int64 {
	id, _ := ctx.Value(types.UserIDKey).(int64)
	return id
}
