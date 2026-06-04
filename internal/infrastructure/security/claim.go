package security

import "github.com/golang-jwt/jwt/v5"

type claims struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	RoleID    string `json:"role_id"`
	RoleCode  string `json:"role_code"`
	jwt.RegisteredClaims
}
