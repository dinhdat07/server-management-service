package smtp

import (
	"context"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"server-management-service/internal/modules/notification/domain"
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

func TestSMTPClient_Ping(t *testing.T) {
	addr := startDummySMTPServer(t)
	host, port, _ := net.SplitHostPort(addr)

	err := Ping(context.Background(), host, port)
	assert.NoError(t, err)

	err = Ping(context.Background(), host, "99999") // invalid port
	assert.Error(t, err)
}

func TestSMTPClient_Send(t *testing.T) {
	addr := startDummySMTPServer(t)
	host, port, _ := net.SplitHostPort(addr)

	cfg := Config{
		Host:     host,
		Port:     port,
		UseAuth:  true,
		UseTLS:   false,
		Username: "user",
		Password: "password",
		From:     "from@test.com",
		FromName: "Test Sender",
	}
	mailer := NewMailer(cfg)

	msg := domain.Message{
		To:       "to@test.com",
		Subject:  "Subject",
		HTMLBody: "<h1>HTML</h1>",
	}

	err := mailer.Send(context.Background(), msg)
	assert.NoError(t, err)
}

func TestSMTPClient_Send_ErrorDial(t *testing.T) {
	cfg := Config{
		Host: "127.0.0.1",
		Port: "99999",
	}
	mailer := NewMailer(cfg)
	err := mailer.Send(context.Background(), domain.Message{To: "a@b.com"})
	assert.Error(t, err)
}

func TestMimeHeaderEncode(t *testing.T) {
	enc := mimeHeaderEncode("Test Name Nguyễn")
	assert.Equal(t, "Test Name Nguyễn", enc)
}
