package middlewares

import (
	"context"

	"server-management-service/internal/infrastructure/security"
	"server-management-service/internal/shared/grpcctx"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func PermissionInterceptor(authorizer *security.Authorizer, methodRoles map[string]string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if authorizer == nil {
			return nil, status.Error(codes.Internal, "internal server error")
		}

		requiredRole, hasRule := methodRoles[info.FullMethod]
		// If no specific rule is configured for this method, allow access by default.
		if !hasRule {
			return handler(ctx, req)
		}

		principal, ok := grpcctx.GetPrincipal(ctx)
		if !ok || principal == nil {
			return nil, status.Error(codes.Unauthenticated, "unauthorized")
		}

		allowed := authorizer.HasRole(ctx, principal, requiredRole)
		if !allowed {
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}

		return handler(ctx, req)
	}
}
