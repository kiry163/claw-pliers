package smtp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/kiry163/claw-mail-monitor/internal/config"
)

type SendRequest struct {
	To      []string `json:"to"`
	Cc      []string `json:"cc,omitempty"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
}

func Send(ctx context.Context, account config.Account, req SendRequest) error {
	if len(req.To) == 0 {
		return errors.New("recipient list is empty")
	}
	if account.Email == "" || account.AuthToken == "" || account.SMTPHost == "" {
		return errors.New("smtp account info is incomplete")
	}

	host, _, err := net.SplitHostPort(account.SMTPHost)
	if err != nil {
		host = account.SMTPHost
	}

	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", account.SMTPHost, &tls.Config{ServerName: host})
	if err != nil {
		return fmt.Errorf("smtp dial failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp client init failed: %w", err)
	}
	defer client.Quit()

	if err := client.Auth(smtp.PlainAuth("", account.Email, account.AuthToken, host)); err != nil {
		return fmt.Errorf("smtp auth failed: %w", err)
	}

	if err := client.Mail(account.Email); err != nil {
		return fmt.Errorf("smtp mail from failed: %w", err)
	}

	for _, addr := range append(req.To, req.Cc...) {
		if err := client.Rcpt(strings.TrimSpace(addr)); err != nil {
			return fmt.Errorf("smtp rcpt failed: %w", err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data failed: %w", err)
	}
	defer writer.Close()

	subject := mime.QEncoding.Encode("utf-8", req.Subject)
	headers := []string{
		fmt.Sprintf("From: %s", account.Email),
		fmt.Sprintf("To: %s", strings.Join(req.To, ", ")),
		fmt.Sprintf("Subject: %s", subject),
		fmt.Sprintf("Date: %s", time.Now().Format(time.RFC1123Z)),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=utf-8",
		"Content-Transfer-Encoding: 8bit",
	}
	if len(req.Cc) > 0 {
		headers = append(headers, fmt.Sprintf("Cc: %s", strings.Join(req.Cc, ", ")))
	}

	if _, err := fmt.Fprintf(writer, "%s\r\n\r\n%s", strings.Join(headers, "\r\n"), req.Body); err != nil {
		return fmt.Errorf("smtp write failed: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}
