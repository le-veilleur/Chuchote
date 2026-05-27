package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/maxime/chuchote/application/dto"
	"github.com/maxime/chuchote/port/inbound"
)

type contextKey string

const ClaimsKey contextKey = "claims"

func Auth(authSvc inbound.AuthUseCase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(header, "Bearer ")
			claims, err := authSvc.ValidateToken(r.Context(), token)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ClaimsFromContext(ctx context.Context) (dto.UserClaims, bool) {
	c, ok := ctx.Value(ClaimsKey).(dto.UserClaims)
	return c, ok
}
