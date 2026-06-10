package service

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"sync"

	"server-management-service/internal/modules/reporting/domain"
	"server-management-service/internal/modules/reporting/repository"
)

//go:embed templates/status_report.html
var statusReportTemplate string

type TemplateData struct {
	StartDate      string
	EndDate        string
	TotalServers   int64
	OnlineServers  int64
	OfflineServers int64
	UptimePercent  float64
}

type ReportingWorker interface {
	Start(ctx context.Context)
	Stop()
	EnqueueReport(req *domain.ReportRequest)
}

type reportingWorkerImpl struct {
	repo        repository.ReportingRepository
	uptimeCalc  domain.UptimeCalculator
	jobQueue    chan *domain.ReportRequest
	workerCount int
	notifier    domain.ReportNotifier
	wg          sync.WaitGroup
}

func NewReportingWorker(repo repository.ReportingRepository, uptimeCalc domain.UptimeCalculator, workerCount int, jobQueueSize int, notifier domain.ReportNotifier) ReportingWorker {
	if workerCount <= 0 {
		workerCount = 5 // Default to 5 concurrent workers
	}
	if jobQueueSize <= 0 {
		jobQueueSize = 100 // Default to 100 capacity
	}
	return &reportingWorkerImpl{
		repo:        repo,
		uptimeCalc:  uptimeCalc,
		jobQueue:    make(chan *domain.ReportRequest, jobQueueSize), // Buffered queue
		workerCount: workerCount,
		notifier:    notifier,
	}
}

func (w *reportingWorkerImpl) Start(ctx context.Context) {
	log.Printf("[ReportingWorker] Starting pool with %d workers", w.workerCount)
	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go w.worker(ctx, i)
	}
}

func (w *reportingWorkerImpl) Stop() {
	log.Println("[ReportingWorker] Stopping pool...")
	close(w.jobQueue)
	w.wg.Wait()
	log.Println("[ReportingWorker] All workers stopped.")
}

func (w *reportingWorkerImpl) EnqueueReport(req *domain.ReportRequest) {
	w.jobQueue <- req
}

func (w *reportingWorkerImpl) worker(ctx context.Context, id int) {
	defer w.wg.Done()
	log.Printf("[ReportingWorker-%d] Started", id)
	for req := range w.jobQueue {
		w.processReport(ctx, req, id)
	}
	log.Printf("[ReportingWorker-%d] Stopped", id)
}

func (w *reportingWorkerImpl) processReport(ctx context.Context, req *domain.ReportRequest, workerID int) {
	log.Printf("[ReportingWorker-%d] Processing report: %s", workerID, req.ID)

	err := w.repo.UpdateReportStatus(ctx, req.ID.String(), domain.ReportStatusProcessing)
	if err != nil {
		log.Printf("[ReportingWorker-%d] Failed to update status to PROCESSING: %v", workerID, err)
		return
	}

	err = w.doWork(ctx, req)

	finalStatus := domain.ReportStatusCompleted
	if err != nil {
		log.Printf("[ReportingWorker-%d] Report %s failed: %v", workerID, req.ID, err)
		finalStatus = domain.ReportStatusFailed
	}

	err = w.repo.UpdateReportStatus(ctx, req.ID.String(), finalStatus)
	if err != nil {
		log.Printf("[ReportingWorker-%d] Failed to update final status: %v", workerID, err)
	} else {
		log.Printf("[ReportingWorker-%d] Finished report: %s", workerID, req.ID)
	}
}

func (w *reportingWorkerImpl) doWork(ctx context.Context, req *domain.ReportRequest) error {
	// 1. Get Server Stats natively via ReportingRepo
	totalServers, err := w.repo.GetServerCountByStatus(ctx, "")
	if err != nil {
		return err
	}
	onlineServers, err := w.repo.GetServerCountByStatus(ctx, "ONLINE")
	if err != nil {
		return err
	}
	offlineServers, err := w.repo.GetServerCountByStatus(ctx, "OFFLINE")
	if err != nil {
		return err
	}

	// 2. Get Uptime from Elasticsearch
	uptimePercent, err := w.uptimeCalc.CalculateUptime(ctx, req.StartTime, req.EndTime)
	if err != nil {
		return err
	}

	// 3. Render and Send HTML Email using template
	tmpl, err := template.New("status_report").Parse(statusReportTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	data := TemplateData{
		StartDate:      req.StartTime.Format("2006-01-02"),
		EndDate:        req.EndTime.Format("2006-01-02"),
		TotalServers:   totalServers,
		OnlineServers:  onlineServers,
		OfflineServers: offlineServers,
		UptimePercent:  uptimePercent,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}
	htmlStr := buf.String()

	log.Printf("[ReportingWorker] Sending email to %s", req.RequestorEmail)

	// Call internal Notification Service
	if w.notifier != nil {
		err := w.notifier.SendReportEmail(ctx, req.RequestorEmail, "Server Status Report", htmlStr)
		if err != nil {
			log.Printf("[ReportingWorker] Error sending email: %v", err)
			return fmt.Errorf("failed to send email: %w", err)
		}
	}

	return nil
}
