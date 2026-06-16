package middlewares

import (
	"context"
	"net"
	"server-management-service/internal/shared/config"
	"server-management-service/internal/shared/grpcctx"
	"server-management-service/internal/infrastructure/ratelimit"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func PreAuthRateLimitInterceptor(
	limiter ratelimit.Limiter,
	keyBuilder ratelimit.KeyBuilder,
	rlCfg *config.RateLimitConfig,
) grpc.UnaryServerInterceptor {
	return rateLimitInterceptor(limiter, keyBuilder, rlCfg, ratelimit.PhasePreAuth)
}

func PostAuthRateLimitInterceptor(
	limiter ratelimit.Limiter,
	keyBuilder ratelimit.KeyBuilder,
	rlCfg *config.RateLimitConfig,
) grpc.UnaryServerInterceptor {
	return rateLimitInterceptor(limiter, keyBuilder, rlCfg, ratelimit.PhasePostAuth)
}

func rateLimitInterceptor(limiter ratelimit.Limiter,
	keyBuilder ratelimit.KeyBuilder,
	rlCfg *config.RateLimitConfig,
	phase ratelimit.Phase,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if rlCfg == nil || !rlCfg.Enabled || limiter == nil {
			return handler(ctx, req)
		}

		policy := ratelimit.PolicyForMethod(info.FullMethod)
		if policy.Phase != phase {
			return handler(ctx, req)
		}

		checks := buildRateLimitChecks(ctx, req, info.FullMethod, keyBuilder, policy)
		if len(checks) == 0 {
			return handler(ctx, req)
		}

		for _, check := range checks {
			result, err := limiter.Allow(ctx, check.Key, policy)
			if err != nil {
				if rlCfg.FailOpen {
					return handler(ctx, req)
				}

				return nil, status.Error(codes.Unavailable, "rate limiter unavailable")
			}

			if !result.Allowed {
				_ = grpc.SetHeader(ctx, metadata.Pairs(
					"x-ratelimit-limit", strconv.Itoa(result.Limit),
					"x-ratelimit-remaining", strconv.Itoa(result.Remaining),
					"x-ratelimit-retry-after", strconv.FormatInt(int64(result.RetryAfter.Seconds()), 10),
				))

				return nil, status.Errorf(
					codes.ResourceExhausted,
					"rate limit exceeded for %s, retry after %s",
					check.Scope,
					result.RetryAfter.String(),
				)
			}
		}

		return handler(ctx, req)
	}
}

type rateLimitCheck struct {
	Scope ratelimit.KeyScope
	Key   string
}

func buildRateLimitChecks(
	ctx context.Context,
	req any,
	method string,
	keyBuilder ratelimit.KeyBuilder,
	policy ratelimit.Policy,
) []rateLimitCheck {
	checks := make([]rateLimitCheck, 0, len(policy.Scopes))

	for _, scope := range policy.Scopes {
		value := valueForRateLimitScope(ctx, req, scope)
		if value == "" {
			continue
		}

		checks = append(checks, rateLimitCheck{
			Scope: scope,
			Key: keyBuilder.Build(
				policy.Name,
				scope,
				value,
			),
		})
	}

	return checks
}

func valueForRateLimitScope(ctx context.Context, req any, scope ratelimit.KeyScope) string {
	switch scope {
	case ratelimit.ScopeIP:
		return extractClientIP(ctx)

	case ratelimit.ScopeIdentifier:
		return identifierFromRequest(req)

	case ratelimit.ScopeEmail:
		return emailFromRequest(req)

	case ratelimit.ScopeUser:
		return userIDFromContext(ctx)

	default:
		return ""
	}
}

func identifierFromRequest(req any) string {
	r, ok := req.(interface{ GetIdentifier() string })
	if !ok {
		return ""
	}

	return strings.ToLower(strings.TrimSpace(r.GetIdentifier()))
}

func extractClientIP(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if forwarded := firstMetadataValue(md, "x-forwarded-for"); forwarded != "" {
			ip := strings.TrimSpace(strings.Split(forwarded, ",")[0])
			if ip != "" {
				return ip
			}
		}

		if realIP := firstMetadataValue(md, "x-real-ip"); realIP != "" {
			ip := strings.TrimSpace(realIP)
			if ip != "" {
				return ip
			}
		}
	}

	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		host, _, err := net.SplitHostPort(p.Addr.String())
		if err == nil {
			return host
		}

		return p.Addr.String()
	}

	return ""
}

func firstMetadataValue(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func emailFromRequest(req any) string {
	r, ok := req.(interface {
		GetEmail() string
	})
	if !ok {
		return ""
	}

	return strings.ToLower(strings.TrimSpace(r.GetEmail()))
}

func userIDFromContext(ctx context.Context) string {
	principal, ok := grpcctx.GetPrincipal(ctx)
	if !ok || principal == nil {
		return ""
	}

	if principal.UserID == "" {
		return ""
	}

	return principal.UserID
}
