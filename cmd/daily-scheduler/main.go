package main

import (
	"log"
	"os"
	"os/signal"
	"server-management-service/internal/app/scheduler"
	"syscall"
)

func main() {
	app, err := scheduler.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize Daily Scheduler: %v", err)
	}

	app.Start()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.Stop()
}
