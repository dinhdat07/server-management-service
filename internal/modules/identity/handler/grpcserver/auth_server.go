package grpcserver

import (
	"context"
	"errors"
	"fmt"

	authv1 "server-management-service/gen/go/auth/v1"
	commonv1 "server-management-service/gen/go/common/v1"
	"server-management-service/internal/infrastructure/security"
	"server-management-service/internal/shared/grpcctx"

	"server-management-service/internal/modules/identity/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	authv1.UnimplementedAuthServiceServer
	authService service.AuthService
}

func NewAuthServer(authService service.AuthService) *AuthServer {
	return &AuthServer{authService: authService}
}

func getClientIP(ctx context.Context) string {
	if p, ok := peer.FromContext(ctx); ok {
		return p.Addr.String()
	}
	return ""
}

func getUserAgent(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if agents := md.Get("grpcgateway-user-agent"); len(agents) > 0 {
			return agents[0]
		}
		if agents := md.Get("user-agent"); len(agents) > 0 {
			return agents[0]
		}
	}
	return ""
}

func (s *AuthServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if req == nil || req.Identifier == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "identifier and password are required")
	}

	ipAddress := getClientIP(ctx)
	userAgent := getUserAgent(ctx)

	result, err := s.authService.Login(ctx, req.Identifier, req.Password, ipAddress, userAgent)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "failed to login")
	}

	csrfManager := security.NewCSRFManager()
	csrfToken, err := csrfManager.GenerateCSRFToken()
	if err == nil {
		_ = grpc.SetHeader(ctx, metadata.Pairs(
			"Set-Cookie-Access-Token", result.AccessToken,
			"Set-Cookie-Refresh-Token", result.RefreshToken,
			"Set-Cookie-Csrf-Token", csrfToken,
		))
	}

	return &authv1.LoginResponse{
		AccessToken: result.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   result.ExpiresIn,
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *authv1.LogoutRequest) (*commonv1.MessageResponse, error) {
	principal, ok := grpcctx.GetPrincipal(ctx)
	if ok && principal != nil && principal.SessionID != "" {
		_ = s.authService.Logout(ctx, principal.SessionID)
	}

	_ = grpc.SetHeader(ctx, metadata.Pairs(
		"Clear-Cookie", "access_token",
		"Clear-Cookie", "refresh_token",
		"Clear-Cookie", "csrf_token",
	))

	return &commonv1.MessageResponse{
		Message: "Logout successful",
	}, nil
}

func (s *AuthServer) RefreshToken(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	if req == nil || req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	result, err := s.authService.Refresh(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRefreshToken) {
			return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
		}
		return nil, status.Error(codes.Internal, "failed to refresh token")
	}

	csrfManager := security.NewCSRFManager()
	csrfToken, err := csrfManager.GenerateCSRFToken()
	if err == nil {
		_ = grpc.SetHeader(ctx, metadata.Pairs(
			"Set-Cookie-Access-Token", result.AccessToken,
			"Set-Cookie-Refresh-Token", result.RefreshToken,
			"Set-Cookie-Csrf-Token", csrfToken,
		))
	}

	return &authv1.RefreshResponse{
		AccessToken: result.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   result.ExpiresIn,
	}, nil
}

func (s *AuthServer) LogoutAll(ctx context.Context, req *authv1.LogoutAllRequest) (*commonv1.MessageResponse, error) {
	principal, ok := grpcctx.GetPrincipal(ctx)
	if ok && principal != nil && principal.UserID != "" {
		var userID uint
		fmt.Sscanf(principal.UserID, "%d", &userID)
		if userID > 0 {
			_ = s.authService.LogoutAll(ctx, userID)
		}
	}

	_ = grpc.SetHeader(ctx, metadata.Pairs(
		"Clear-Cookie", "access_token",
		"Clear-Cookie", "refresh_token",
		"Clear-Cookie", "csrf_token",
	))

	return &commonv1.MessageResponse{
		Message: "Logout all successful",
	}, nil
}
