package gateway

import (
	"net/http"
)

// CookieMiddleware parses cookies and injects them as Authorization metadata for gRPC Gateway
func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("access_token"); err == nil {
			r.Header.Add("Authorization", "Bearer "+cookie.Value)
		}

		next.ServeHTTP(w, r)
	})
}
