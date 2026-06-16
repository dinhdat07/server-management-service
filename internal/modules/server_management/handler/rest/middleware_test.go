package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"server-management-service/internal/infrastructure/security"
	"server-management-service/internal/modules/identity/domain"
)

func generateTestToken(secret string, userID string, role domain.RoleCode) string {
	claims := jwt.MapClaims{
		"user_id":   userID,
		"email":     "test@test.com",
		"role_code": string(role),
		"exp":       time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestAuthMiddleware(t *testing.T) {
	secret := "super-secret-key"
	auth := security.NewAuthenticator(secret, nil)

	middleware := AuthMiddleware(auth)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := PrincipalFromContext(r.Context())
		assert.NotNil(t, p)
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("missing token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer invalid")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("valid token", func(t *testing.T) {
		token := generateTestToken(secret, "user1", domain.RoleCodeUser)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestPermissionMiddleware(t *testing.T) {
	authorizer := security.NewAuthorizer()

	middleware := PermissionMiddleware(authorizer, security.PermServerCreate)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("no principal in context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("insufficient permissions", func(t *testing.T) {
		p := &security.Principal{UserID: "1", RoleCode: domain.RoleCodeUser}
		ctx := principalWithContext(context.Background(), p)
		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("sufficient permissions", func(t *testing.T) {
		p := &security.Principal{UserID: "2", RoleCode: domain.RoleCodeAdmin}
		ctx := principalWithContext(context.Background(), p)
		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
