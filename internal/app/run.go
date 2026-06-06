package app

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	
	reportingv1 "server-management-service/gen/go/reporting/v1"
	server_managementv1 "server-management-service/gen/go/server_management/v1"
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

	// TODO: Replace with proper GRPC server initialization in Phase 2 (Auth Interceptor)
	a.GRPCServer = grpc.NewServer()
	
	if a.ServerHandler != nil {
		server_managementv1.RegisterServerManagementServiceServer(a.GRPCServer, a.ServerHandler)
	}
	
	if a.ReportingHandler != nil {
		reportingv1.RegisterReportingServiceServer(a.GRPCServer, a.ReportingHandler)
	}

	// TODO: Replace with proper gateway mux in Phase 2
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

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
