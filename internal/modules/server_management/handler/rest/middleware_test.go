package rest

import (

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"server-management-service/internal/infrastructure/security"
	identity_domain "server-management-service/internal/modules/identity/domain"
)

func TestAuthMiddleware(t *testing.T) {
	// Use actual Authenticator with a dummy secret instead of mock
	auth := security.NewAuthenticator("dummy-secret", nil)
	authMiddleware := AuthMiddleware(auth)

	handler := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := PrincipalFromContext(r.Context())
		assert.NotNil(t, p)
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("Missing Token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer invalid-jwt-token")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestPermissionMiddleware(t *testing.T) {
	authorizer := security.NewAuthorizer()

	permMiddleware := PermissionMiddleware(authorizer, security.PermServerCreate)
	handler := permMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("No Principal", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Insufficient Permissions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		p := &security.Principal{RoleCode: identity_domain.RoleCode("UNKNOWN")} // Unknown role has no perms
		req = req.WithContext(principalWithContext(req.Context(), p))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		p := &security.Principal{RoleCode: identity_domain.RoleCodeAdmin} // Admin has all perms
		req = req.WithContext(principalWithContext(req.Context(), p))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
