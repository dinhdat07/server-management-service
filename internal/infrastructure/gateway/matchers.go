package gateway

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func CustomIncomingMatcher(key string) (string, bool) {
	switch key {
	case "Cookie":
		return "cookie", true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

func CustomOutgoingMatcher(key string) (string, bool) {
	switch key {
	case "set-cookie-access-token":
		return "Set-Cookie", true
	case "set-cookie-refresh-token":
		return "Set-Cookie", true
	case "set-cookie-csrf-token":
		return "Set-Cookie", true
	case "clear-cookie":
		return "Set-Cookie", true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}
