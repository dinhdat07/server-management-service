package service

import (
	"context"
	
	"server-management-service/internal/modules/notification/domain"
	"server-management-service/internal/modules/notification/infrastructure/smtp"
)

type NotificationService struct {
	mailer *smtp.SMTPMailer
}

func NewNotificationService(mailer *smtp.SMTPMailer) *NotificationService {
	return &NotificationService{mailer: mailer}
}

func (s *NotificationService) SendReportEmail(ctx context.Context, toEmail string, subject string, htmlBody string) error {
	msg := domain.Message{
		To:       toEmail,
		Subject:  subject,
		HTMLBody: htmlBody,
	}
	return s.mailer.Send(ctx, msg)
}
