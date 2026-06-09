package app

import (
	"buf.build/go/protovalidate"
	"google.golang.org/grpc"

	authv1 "server-management-service/gen/go/auth/v1"
	reportingv1 "server-management-service/gen/go/reporting/v1"
	server_managementv1 "server-management-service/gen/go/server_management/v1"
	"server-management-service/internal/infrastructure/ratelimit"
	"server-management-service/internal/infrastructure/security"
	authgrpc "server-management-service/internal/modules/identity/handler/grpcserver"
	reportinggrpc "server-management-service/internal/modules/reporting/handler/grpcserver"
	servergrpc "server-management-service/internal/modules/server_management/handler/grpcserver"
	"server-management-service/internal/shared/config"
	"server-management-service/internal/shared/middlewares"
)

type GRPCServerDeps struct {
	Validator           protovalidate.Validator
	Authenticator       *security.Authenticator
	Authorizer          *security.Authorizer
	CSRFManager         *security.CSRFManager
	Auth                *authgrpc.AuthServer
	Reporting           *reportinggrpc.ReportingGrpcHandler
	ServerManagement    *servergrpc.ServerManagementServer
	RateLimiter         ratelimit.Limiter
	RateLimitKeyBuilder ratelimit.KeyBuilder
	RateLimitConfig     *config.RateLimitConfig
}

func NewGRPCServer(deps GRPCServerDeps) *grpc.Server {
	publicMethods := buildGRPCPublicMethods()
	methodPermissions := buildGRPCMethodPermissions()

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middlewares.RecoveryInterceptor(),
			middlewares.PreAuthRateLimitInterceptor(deps.RateLimiter, deps.RateLimitKeyBuilder, deps.RateLimitConfig),
			middlewares.ValidationInterceptor(deps.Validator),
			middlewares.CSRFInterceptor(deps.CSRFManager),
			middlewares.AuthenticationInterceptor(deps.Authenticator, publicMethods),
			middlewares.PostAuthRateLimitInterceptor(deps.RateLimiter, deps.RateLimitKeyBuilder, deps.RateLimitConfig),
			middlewares.PermissionInterceptor(deps.Authorizer, methodPermissions),
		),
	)

	if deps.ServerManagement != nil {
		server_managementv1.RegisterServerManagementServiceServer(s, deps.ServerManagement)
	}
	if deps.Reporting != nil {
		reportingv1.RegisterReportingServiceServer(s, deps.Reporting)
	}
	if deps.Auth != nil {
		authv1.RegisterAuthServiceServer(s, deps.Auth)
	}

	return s
}

func buildGRPCPublicMethods() map[string]bool {
	return map[string]bool{
		"/grpc.health.v1.Health/Check":             true,
		"/portal.auth.v1.AuthService/Login":        true,
		"/portal.auth.v1.AuthService/RefreshToken": true,
	}
}

func buildGRPCMethodPermissions() map[string]security.PermissionCode {
	return map[string]security.PermissionCode{
		"/server_management.v1.ServerManagementService/CreateServer":  security.PermServerCreate,
		"/server_management.v1.ServerManagementService/UpdateServer":  security.PermServerUpdate,
		"/server_management.v1.ServerManagementService/DeleteServer":  security.PermServerDelete,
		"/server_management.v1.ServerManagementService/ImportServers": security.PermServerImport,
		"/server_management.v1.ServerManagementService/ExportServers": security.PermServerExport,
		"/server_management.v1.ServerManagementService/ViewServers":   security.PermServerRead,
		"/reporting.v1.ReportingService/RequestReport":                security.PermReportRequest,
	}
}
