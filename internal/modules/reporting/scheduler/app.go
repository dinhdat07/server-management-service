package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"server-management-service/internal/modules/notification/infrastructure/smtp"
	notificationsvc "server-management-service/internal/modules/notification/service"
	"server-management-service/internal/modules/reporting/repository/impl"
	"server-management-service/internal/modules/reporting/service"
	"server-management-service/internal/infrastructure/elasticsearch"
	"server-management-service/internal/shared/config"
	"server-management-service/internal/shared/database"
)

type App struct {
	cron             *cron.Cron
	reportingService service.ReportingService
	reportingWorker  service.ReportingWorker
	adminEmail       string
}

func NewApp() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	db, err := database.GetInstance(cfg.DBUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	esCfg := config.LoadElasticsearchConfig()
	esClient, err := database.NewElasticsearchClient(context.Background(), []string{esCfg.URL})
	if err != nil {
		return nil, fmt.Errorf("elasticsearch connection failed: %w", err)
	}

	reportingRepo := impl.NewGormReportingRepository(db)

	smtpConfig := smtp.Config{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		UseAuth:  cfg.SMTP.UseAuth,
		UseTLS:   cfg.SMTP.UseTLS,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
		From:     cfg.SMTP.From,
		FromName: cfg.SMTP.FromName,
	}
	smtpMailer := smtp.NewMailer(smtpConfig)
	notificationService := notificationsvc.NewNotificationService(smtpMailer)

	uptimeCalc := elasticsearch.NewESUptimeCalculator(esClient, esCfg.ServerIndex)
	reportingWorker := service.NewReportingWorker(reportingRepo, uptimeCalc, cfg.Reporting.WorkerCount, cfg.Reporting.JobQueueSize, notificationService)
	reportingService := service.NewReportingService(reportingRepo, reportingWorker)

	adminEmail := config.GetEnvDefault("ADMIN_EMAIL", "admin@portal.local")

	c := cron.New(cron.WithLocation(time.Local))

	app := &App{
		cron:             c,
		reportingService: reportingService,
		reportingWorker:  reportingWorker,
		adminEmail:       adminEmail,
	}

	err = app.setupCronJobs()
	if err != nil {
		return nil, fmt.Errorf("failed to setup cron jobs: %w", err)
	}

	return app, nil
}

func (a *App) setupCronJobs() error {
	// Schedule at midnight (00:00) every day
	_, err := a.cron.AddFunc("0 0 * * *", func() {
		log.Println("Running daily report scheduler...")

		ctx := context.Background()

		// For the previous day
		now := time.Now()
		yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

		err := a.reportingService.RequestReport(ctx, a.adminEmail, yesterday, yesterday)
		if err != nil {
			log.Printf("Failed to request daily report: %v", err)
		} else {
			log.Println("Successfully scheduled daily report")
		}
	})

	return err
}

func (a *App) Start() {
	a.reportingWorker.Start(context.Background())
	a.cron.Start()
	log.Println("Daily Scheduler started, waiting for cron jobs...")
}

func (a *App) Stop() {
	log.Println("Shutting down Daily Scheduler...")
	a.cron.Stop()
	a.reportingWorker.Stop()
	log.Println("Daily Scheduler stopped.")
}

