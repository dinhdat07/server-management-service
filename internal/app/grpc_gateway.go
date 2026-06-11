package app

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authv1 "server-management-service/gen/go/auth/v1"
	reportingv1 "server-management-service/gen/go/reporting/v1"
	server_managementv1 "server-management-service/gen/go/server_management/v1"
	"server-management-service/internal/infrastructure/gateway"
)

// setupGRPCGateway creates a gRPC-gateway mux and registers all service handlers.
func (a *App) setupGRPCGateway(ctx context.Context, grpcAddr string) (*runtime.ServeMux, error) {
	gwmux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(gateway.CustomIncomingMatcher),
		runtime.WithOutgoingHeaderMatcher(gateway.CustomOutgoingMatcher),
	)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := server_managementv1.RegisterServerManagementServiceHandlerFromEndpoint(ctx, gwmux, grpcAddr, opts); err != nil {
		return nil, fmt.Errorf("register server management gateway: %w", err)
	}
	if err := reportingv1.RegisterReportingServiceHandlerFromEndpoint(ctx, gwmux, grpcAddr, opts); err != nil {
		return nil, fmt.Errorf("register reporting gateway: %w", err)
	}
	if err := authv1.RegisterAuthServiceHandlerFromEndpoint(ctx, gwmux, grpcAddr, opts); err != nil {
		return nil, fmt.Errorf("register auth gateway: %w", err)
	}

	return gwmux, nil
}
