package mail

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/kiry163/claw-pliers/internal/config"
)

var (
	cfg      *config.Config
	accounts []config.AccountConfig
)

func Init(mailCfg config.Config) error {
	cfg = &mailCfg
	accounts = mailCfg.Mail.Accounts
	return nil
}

func GetConfig() *config.Config {
	return cfg
}

func ListAccounts() []config.AccountConfig {
	return accounts
}

func FindAccount(email string) (config.AccountConfig, bool) {
	for _, acc := range accounts {
		if acc.Email == email {
			return acc, true
		}
	}
	return config.AccountConfig{}, false
}

type EmailSummary struct {
	From    string
	Subject string
	Date    string
	Preview string
}

func GetLatestEmails(accountEmail string, count int) ([]EmailSummary, error) {
	account, found := FindAccount(accountEmail)
	if !found {
		return nil, fmt.Errorf("account not found: %s", accountEmail)
	}

	imapHost, _ := getProviderSettings(account.Provider)
	if imapHost == "" {
		return nil, fmt.Errorf("unknown provider: %s", account.Provider)
	}

	addr := fmt.Sprintf("%s:993", imapHost)
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: imapHost})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to IMAP: %v", err)
	}
	defer conn.Close()

	c, err := client.New(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to create IMAP client: %v", err)
	}
	defer c.Logout()

	if err := c.Login(account.Email, account.AuthToken); err != nil {
		return nil, fmt.Errorf("login failed: %v", err)
	}
	defer c.Logout()

	mbox, err := c.Select("INBOX", false)
	if err != nil {
		return nil, fmt.Errorf("failed to select inbox: %v", err)
	}

	if mbox.Messages == 0 {
		return []EmailSummary{}, nil
	}

	fromSeqNum := mbox.Messages - uint32(count) + 1
	if fromSeqNum < 1 {
		fromSeqNum = 1
	}

	seqset := new(imap.SeqSet)
	seqset.AddRange(fromSeqNum, mbox.Messages)

	items := []imap.FetchItem{
		imap.FetchEnvelope,
	}

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.Fetch(seqset, items, messages)
	}()

	var results []EmailSummary
	for msg := range messages {
		summary := EmailSummary{}

		if len(msg.Envelope.From) > 0 {
			summary.From = msg.Envelope.From[0].PersonalName
			if summary.From == "" {
				summary.From = msg.Envelope.From[0].MailboxName + "@" + msg.Envelope.From[0].HostName
			}
		}

		summary.Subject = msg.Envelope.Subject
		summary.Date = msg.Envelope.Date.Format("2006-01-02 15:04:05")
		summary.Preview = "(body not fetched)"

		results = append(results, summary)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("fetch failed: %v", err)
	}

	return results, nil
}

func SendMail(fromEmail, to, subject, body string) error {
	account, found := FindAccount(fromEmail)
	if !found {
		return fmt.Errorf("account not found: %s", fromEmail)
	}

	_, smtpHost := getProviderSettings(account.Provider)
	if smtpHost == "" {
		return fmt.Errorf("unknown provider: %s", account.Provider)
	}

	addr := fmt.Sprintf("%s:465", smtpHost)
	host := smtpHost

	tlsConfig := &tls.Config{ServerName: host}
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP: %v", err)
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}
	defer c.Quit()

	auth := smtp.PlainAuth("", account.Email, account.AuthToken, host)
	if err := c.Auth(auth); err != nil {
		return fmt.Errorf("auth failed: %v", err)
	}

	if err := c.Mail(account.Email); err != nil {
		return fmt.Errorf("mail from failed: %v", err)
	}

	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt failed: %v", err)
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("data failed: %v", err)
	}
	defer w.Close()

	msg := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/plain; charset=utf-8\r\n"+
			"\r\n"+
			"%s\r\n",
		account.Email, to, subject, body)

	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("write failed: %v", err)
	}

	return nil
}

func TestConnection(email string) (int64, error) {
	account, found := FindAccount(email)
	if !found {
		return 0, fmt.Errorf("account not found: %s", email)
	}

	imapHost, _ := getProviderSettings(account.Provider)
	if imapHost == "" {
		return 0, fmt.Errorf("unknown provider: %s", account.Provider)
	}

	start := time.Now()

	addr := fmt.Sprintf("%s:993", imapHost)
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: imapHost})
	if err != nil {
		return 0, fmt.Errorf("failed to connect: %v", err)
	}
	defer conn.Close()

	c, err := client.New(conn)
	if err != nil {
		return 0, fmt.Errorf("failed to create IMAP client: %v", err)
	}
	defer c.Logout()

	if err := c.Login(account.Email, account.AuthToken); err != nil {
		return 0, fmt.Errorf("login failed: %v", err)
	}
	defer c.Logout()

	latency := time.Since(start).Milliseconds()
	return latency, nil
}

func getProviderSettings(provider string) (imapHost, smtpHost string) {
	switch provider {
	case "163":
		return "imap.163.com", "smtp.163.com"
	case "126":
		return "imap.126.com", "smtp.126.com"
	case "qq":
		return "imap.qq.com", "smtp.qq.com"
	case "gmail":
		return "imap.gmail.com", "smtp.gmail.com"
	case "outlook":
		return "outlook.office365.com", "smtp.office365.com"
	default:
		return "", ""
	}
}
