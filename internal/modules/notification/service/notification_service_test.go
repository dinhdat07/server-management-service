package service

import (
	"context"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"server-management-service/internal/modules/notification/infrastructure/smtp"
)

func startDummySMTPServer(t *testing.T) string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				c.Write([]byte("220 dummy ESMTP\r\n"))
				buf := make([]byte, 2048)
				for {
					n, err := c.Read(buf)
					if err != nil {
						return
					}
					req := string(buf[:n])
					if strings.HasPrefix(req, "EHLO") || strings.HasPrefix(req, "HELO") {
						c.Write([]byte("250-dummy\r\n250 AUTH PLAIN\r\n"))
					} else if strings.HasPrefix(req, "AUTH") {
						c.Write([]byte("235 2.7.0 Authentication successful\r\n"))
					} else if strings.HasPrefix(req, "MAIL") {
						c.Write([]byte("250 2.1.0 Ok\r\n"))
					} else if strings.HasPrefix(req, "RCPT") {
						c.Write([]byte("250 2.1.5 Ok\r\n"))
					} else if strings.HasPrefix(req, "DATA") {
						c.Write([]byte("354 End data with <CR><LF>.<CR><LF>\r\n"))
					} else if strings.HasPrefix(req, "QUIT") {
						c.Write([]byte("221 2.0.0 Bye\r\n"))
						return
					} else if strings.Contains(req, "\r\n.\r\n") {
						c.Write([]byte("250 2.0.0 Ok: queued\r\n"))
					}
				}
			}(conn)
		}
	}()
	return l.Addr().String()
}

func TestNotificationService_SendReportEmail(t *testing.T) {
	addr := startDummySMTPServer(t)
	host, port, _ := net.SplitHostPort(addr)

	cfg := smtp.Config{
		Host:     host,
		Port:     port,
		UseAuth:  false,
		UseTLS:   false,
		From:     "noreply@test.com",
	}
	mailer := smtp.NewMailer(cfg)
	svc := NewNotificationService(mailer)

	err := svc.SendReportEmail(context.Background(), "user@test.com", "Report", "<p>content</p>")
	assert.NoError(t, err)
}
