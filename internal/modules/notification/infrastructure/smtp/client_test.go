package smtp

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"server-management-service/internal/modules/notification/domain"

	"github.com/stretchr/testify/assert"
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

func TestPing_Success(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("Failed to start dummy tcp server")
	}
	defer ln.Close()

	host, port, _ := net.SplitHostPort(ln.Addr().String())

	err = Ping(context.Background(), host, port)
	assert.NoError(t, err)
}

func TestPing_Error(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err := Ping(ctx, "127.0.0.255", "1") // unreachable IP
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

func startFailingSMTPServer(t *testing.T, failCmd string) string {
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
					} else if strings.HasPrefix(req, failCmd) {
						c.Write([]byte("500 Error\r\n"))
						return
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

func TestSMTPClient_Send_Failures(t *testing.T) {
	failCmds := []string{"AUTH", "MAIL", "RCPT", "DATA", "QUIT"}
	for _, cmd := range failCmds {
		t.Run("Fail_"+cmd, func(t *testing.T) {
			addr := startFailingSMTPServer(t, cmd)
			host, port, _ := net.SplitHostPort(addr)

			cfg := Config{
				Host:     host,
				Port:     port,
				UseAuth:  true,
				Username: "user",
				Password: "password",
				From:     "from@test.com",
			}
			mailer := NewMailer(cfg)

			err := mailer.Send(context.Background(), domain.Message{To: "to@test.com"})
			assert.Error(t, err)
		})
	}
}

func TestSMTPClient_Send_ErrorNewClient(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	go func() {
		conn, _ := l.Accept()
		if conn != nil {
			conn.Close() // Close immediately to cause NewClient to fail
		}
	}()

	host, port, _ := net.SplitHostPort(l.Addr().String())
	cfg := Config{Host: host, Port: port}
	mailer := NewMailer(cfg)
	err = mailer.Send(context.Background(), domain.Message{To: "to@test.com"})
	assert.Error(t, err)
}

func TestSMTPClient_Send_TLS(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	go func() {
		conn, _ := l.Accept()
		defer conn.Close()
		conn.Write([]byte("220 dummy ESMTP\r\n"))
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			req := string(buf[:n])
			if strings.HasPrefix(req, "EHLO") {
				conn.Write([]byte("250-dummy\r\n250 STARTTLS\r\n"))
			} else if strings.HasPrefix(req, "STARTTLS") {
				conn.Write([]byte("500 StartTLS Failed\r\n"))
				return
			}
		}
	}()

	host, port, _ := net.SplitHostPort(l.Addr().String())
	cfg := Config{
		Host:     host,
		Port:     port,
		UseTLS:   true, 
		From:     "from@test.com",
	}
	mailer := NewMailer(cfg)

	err = mailer.Send(context.Background(), domain.Message{To: "to@test.com"})
	assert.Error(t, err)
}

