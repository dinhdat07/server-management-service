package main

import (
	"os"
	"os/signal"
	"server-management-service/internal/app/scheduler"
	"server-management-service/internal/shared/logger"
	"syscall"
)

func main() {
	app, err := scheduler.NewApp()
	if err != nil {
		logger.Log.Sugar().Fatalf("Failed to initialize Daily Scheduler: %v", err)
	}

	app.Start()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.Stop()
}
