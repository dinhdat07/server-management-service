package rest

import (
	"context"
	"net/http"

	"server-management-service/internal/infrastructure/security"
)

type contextKey string

const principalKey contextKey = "principal"

func PrincipalFromContext(ctx context.Context) *security.Principal {
	p, _ := ctx.Value(principalKey).(*security.Principal)
	return p
}

func principalWithContext(ctx context.Context, p *security.Principal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}

func AuthMiddleware(auth *security.Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearer(r)
			if token == "" {
				writeError(w, http.StatusUnauthorized, "missing authorization token")
				return
			}
			principal, err := auth.Authenticate(r.Context(), token)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid token")
				return
			}
			next.ServeHTTP(w, r.WithContext(principalWithContext(r.Context(), principal)))
		})
	}
}

func PermissionMiddleware(authorizer *security.Authorizer, required security.PermissionCode) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal := PrincipalFromContext(r.Context())
			if principal == nil {
				writeError(w, http.StatusUnauthorized, "authentication required")
				return
			}
			if !authorizer.HasPermission(r.Context(), principal, required) {
				writeError(w, http.StatusForbidden, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
