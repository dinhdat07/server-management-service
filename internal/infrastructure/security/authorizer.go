package security

import (
	"context"
)

type Authorizer struct {
}

func NewAuthorizer() *Authorizer {
	return &Authorizer{}
}

func (a *Authorizer) HasRole(ctx context.Context, principal *Principal, requiredRole string) bool {
	// If no specific role is required, allow access
	if requiredRole == "" {
		return true
	}

	// Admin has access to everything
	if principal.RoleCode == "ADMIN" {
		return true
	}

	return principal.RoleCode == requiredRole
}
