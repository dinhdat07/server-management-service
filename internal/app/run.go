package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	
	reportingv1 "server-management-service/gen/go/reporting/v1"
	server_managementv1 "server-management-service/gen/go/server_management/v1"
	"server-management-service/internal/infrastructure/security"
	"server-management-service/internal/shared/middlewares"
)

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	grpcAddr := ":" + a.Config.GRPCPort
	httpAddr := ":" + a.Config.HTTPPort

	authenticator := security.NewAuthenticator(a.Config.JWTSecret, a.RedisClient)
	authorizer := security.NewAuthorizer()

	publicMethods := map[string]bool{
		"/grpc.health.v1.Health/Check": true,
	}

	methodRoles := map[string]string{
		"/server_management.v1.ServerManagementService/CreateServer":  "ADMIN",
		"/server_management.v1.ServerManagementService/UpdateServer":  "ADMIN",
		"/server_management.v1.ServerManagementService/DeleteServer":  "ADMIN",
		"/server_management.v1.ServerManagementService/ImportServers": "ADMIN",
		"/server_management.v1.ServerManagementService/ExportServers": "ADMIN",
		"/server_management.v1.ServerManagementService/ViewServers":   "", // allow any logged-in user
		"/reporting.v1.ReportingService/RequestReport":                "ADMIN",
	}

	a.GRPCServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middlewares.AuthenticationInterceptor(authenticator, publicMethods),
			middlewares.PermissionInterceptor(authorizer, methodRoles),
		),
	)
	
	if a.ServerHandler != nil {
		server_managementv1.RegisterServerManagementServiceServer(a.GRPCServer, a.ServerHandler)
	}
	
	if a.ReportingHandler != nil {
		reportingv1.RegisterReportingServiceServer(a.GRPCServer, a.ReportingHandler)
	}

	if a.ReportingWorker != nil {
		a.ReportingWorker.Start(ctx)
	}

	gwmux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := server_managementv1.RegisterServerManagementServiceHandlerFromEndpoint(ctx, gwmux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register server management gateway: %w", err)
	}

	if err := reportingv1.RegisterReportingServiceHandlerFromEndpoint(ctx, gwmux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register reporting gateway: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	// Mount the gRPC gateway to the root of the HTTP server
	mux.Handle("/", gwmux)

	a.HTTPServer = &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	errCh := make(chan error, 2)

	go func() {
		log.Printf("grpc listening on %s", grpcAddr)
		errCh <- a.runGRPCServer(grpcAddr)
	}()

	go func() {
		log.Printf("gateway listening on %s", httpAddr)
		errCh <- a.runHTTPServer()
	}()

	select {
	case <-ctx.Done():
		log.Println("shutdown signal received")
		return a.Shutdown(context.Background())

	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func (a *App) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	done := make(chan struct{})

	go func() {
		if a.GRPCServer != nil {
			log.Println("stopping grpc server")
			a.GRPCServer.GracefulStop()
		}
		close(done)
	}()

	select {
	case <-done:
	case <-shutdownCtx.Done():
		if a.GRPCServer != nil {
			log.Println("grpc graceful stop timeout, forcing stop")
			a.GRPCServer.Stop()
		}
	}

	if a.HTTPServer != nil {
		log.Println("stopping gateway server")
		if err := a.HTTPServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
	}

	if a.ReportingWorker != nil {
		a.ReportingWorker.Stop()
	}

	log.Println("application stopped gracefully")
	return nil
}

func (a *App) runHTTPServer() error {
	return a.HTTPServer.ListenAndServe()
}

func (a *App) runGRPCServer(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return a.GRPCServer.Serve(lis)
}
