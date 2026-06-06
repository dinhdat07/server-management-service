package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"

	"server-management-service/internal/modules/reporting/repository/impl"
	"server-management-service/internal/modules/reporting/service"
	"server-management-service/internal/shared/config"
	"server-management-service/internal/shared/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Database
	db, err := database.GetInstance(cfg.DBUrl)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	txManager := impl.NewGormTxManager(db)
	outboxRepo := impl.NewGormOutboxRepository(db)

	// Initialize Reporting Service
	reportingService := service.NewReportingService(outboxRepo, txManager)

	// Admin email to receive reports
	adminEmail := config.GetEnvDefault("ADMIN_EMAIL", "admin@portal.local")

	// Initialize Cron
	c := cron.New(cron.WithLocation(time.Local))

	// Schedule at midnight (00:00) every day
	_, err = c.AddFunc("0 0 * * *", func() {
		log.Println("Running daily report scheduler...")

		ctx := context.Background()

		// For the previous day
		now := time.Now()
		yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

		err := reportingService.RequestReport(ctx, adminEmail, yesterday, yesterday)
		if err != nil {
			log.Printf("Failed to request daily report: %v", err)
		} else {
			log.Println("Successfully scheduled daily report")
		}
	})

	if err != nil {
		log.Fatalf("Failed to add cron job: %v", err)
	}

	c.Start()
	log.Println("Daily Scheduler started, waiting for cron jobs...")

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Daily Scheduler...")
	c.Stop()
}
