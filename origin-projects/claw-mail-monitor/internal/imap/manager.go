package imap

import (
	"context"
	"errors"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/emersion/go-imap"
	id "github.com/emersion/go-imap-id"
	"github.com/emersion/go-imap/client"

	"github.com/kiry163/claw-mail-monitor/internal/config"
	"github.com/kiry163/claw-mail-monitor/internal/parser"
	"github.com/kiry163/claw-mail-monitor/internal/webhook"
)

type Manager struct {
	cfg          *config.Config
	webhook      *webhook.Client
	logs         *LogStore
	pollInterval time.Duration

	mu       sync.Mutex
	running  map[string]context.CancelFunc
	started  map[string]struct{}
	stopOnce sync.Once
}

func NewManager(cfg *config.Config, webhookClient *webhook.Client) *Manager {
	pollInterval := 30 * time.Second
	if cfg != nil && cfg.Monitoring.PollInterval != "" {
		if v, err := time.ParseDuration(cfg.Monitoring.PollInterval); err == nil && v > 0 {
			pollInterval = v
		} else if err != nil {
			slog.Warn("invalid poll interval, fallback to default", "value", cfg.Monitoring.PollInterval, "error", err)
		}
	}

	return &Manager{
		cfg:          cfg,
		webhook:      webhookClient,
		logs:         NewLogStore(200),
		pollInterval: pollInterval,
		running:      make(map[string]context.CancelFunc),
		started:      make(map[string]struct{}),
	}
}

func (m *Manager) StartAll(ctx context.Context) {
	for _, account := range m.cfg.Accounts {
		if account.Enabled {
			m.StartAccount(ctx, account)
		}
	}
}

func (m *Manager) StartAccount(ctx context.Context, account config.Account) {
	if account.Email == "" || account.IMAPHost == "" {
		return
	}

	m.mu.Lock()
	if _, ok := m.running[account.Email]; ok {
		m.mu.Unlock()
		return
	}

	childCtx, cancel := context.WithCancel(ctx)
	m.running[account.Email] = cancel
	m.started[account.Email] = struct{}{}
	m.mu.Unlock()

	slog.Info("imap monitoring starting", "email", account.Email)

	go m.monitorLoop(childCtx, account)
}

func (m *Manager) StopAccount(email string) {
	m.mu.Lock()
	cancel, ok := m.running[email]
	if ok {
		delete(m.running, email)
	}
	m.mu.Unlock()

	if ok && cancel != nil {
		slog.Info("imap monitoring stopping", "email", email)
		cancel()
	}
}

func (m *Manager) IsMonitoring(email string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.running[email]
	return ok
}

func (m *Manager) MonitoringCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.running)
}

func (m *Manager) Logs(limit int) ([]EmailLog, int) {
	return m.logs.List(limit), m.logs.Total()
}

func (m *Manager) GetLatestEmails(ctx context.Context, account config.Account, count int) ([]EmailContent, error) {
	if account.IMAPHost == "" || account.Email == "" || account.AuthToken == "" {
		return nil, errors.New("imap connection info is incomplete")
	}

	c, err := client.DialTLS(account.IMAPHost, nil)
	if err != nil {
		return nil, err
	}
	defer c.Logout()

	if err := c.Login(account.Email, account.AuthToken); err != nil {
		return nil, err
	}

	// 发送 IMAP ID (重要：163需要特定值)
	idClient := id.NewClient(c)
	if _, err := idClient.ID(id.ID{
		id.FieldName:    "IMAPClient",
		id.FieldVersion: "3.1.0",
	}); err != nil {
		slog.Warn("imap id command failed", "email", account.Email, "error", err)
	}

	mailbox, err := c.Select("INBOX", false)
	if err != nil {
		return nil, err
	}

	if mailbox.Messages == 0 {
		slog.Debug("no emails found", "email", account.Email, "messages", mailbox.Messages)
		return []EmailContent{}, nil
	}

	// 使用序号获取最新的邮件
	totalMessages := int(mailbox.Messages)
	if count > totalMessages {
		count = totalMessages
	}

	// 获取最后 count 封邮件的序号范围
	startSeq := totalMessages - count + 1
	endSeq := totalMessages

	seqset := new(imap.SeqSet)
	seqset.AddRange(uint32(startSeq), uint32(endSeq))

	slog.Debug("fetching emails by sequence", "email", account.Email, "start_seq", startSeq, "end_seq", endSeq, "count", count)

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchUid, section.FetchItem()}

	messages := make(chan *imap.Message, 10)
	errCh := make(chan error, 1)
	go func() {
		errCh <- c.Fetch(seqset, items, messages)
	}()

	var emails []EmailContent
	for msg := range messages {
		if msg == nil {
			continue
		}
		body := msg.GetBody(section)
		parsed, err := parser.ParseMessage(msg, body)
		if err != nil {
			slog.Warn("parse email failed", "error", err)
			continue
		}
		emails = append(emails, EmailContent{
			Account: account.Email,
			From:    parsed.From,
			To:      parsed.To,
			Subject: parsed.Subject,
			Date:    parsed.Date,
			Body:    parsed.Body,
			UID:     parsed.UID,
		})
	}

	select {
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
	default:
	}

	if emails == nil {
		emails = []EmailContent{}
	}

	return emails, nil
}

func (m *Manager) GetLatestEmailsSince(ctx context.Context, account config.Account, count int, since time.Duration) ([]EmailContent, error) {
	if count <= 0 {
		count = 1
	}
	perAccount := count * 2
	if perAccount < count {
		perAccount = count
	}

	emails, err := m.GetLatestEmails(ctx, account, perAccount)
	if err != nil {
		return nil, err
	}

	emails = sortEmails(emails)
	if since > 0 {
		emails = filterBySince(emails, since)
	}
	if len(emails) > count {
		emails = emails[:count]
	}
	return emails, nil
}

func (m *Manager) GetLatestEmailsAll(ctx context.Context, count int) ([]EmailContent, error) {
	if count <= 0 {
		count = 1
	}

	perAccount := count * 2
	if perAccount < count {
		perAccount = count
	}

	accounts := make([]config.Account, 0, len(m.cfg.Accounts))
	for _, acct := range m.cfg.Accounts {
		if acct.Enabled {
			accounts = append(accounts, acct)
		}
	}

	var combined []EmailContent
	for _, acct := range accounts {
		emails, err := m.GetLatestEmails(ctx, acct, perAccount)
		if err != nil {
			slog.Warn("fetch latest emails failed", "email", acct.Email, "error", err)
			continue
		}
		combined = append(combined, emails...)
	}

	if len(combined) == 0 {
		return []EmailContent{}, nil
	}

	combined = sortEmails(combined)

	if len(combined) > count {
		combined = combined[:count]
	}

	return combined, nil
}

func (m *Manager) GetLatestEmailsAllSince(ctx context.Context, count int, since time.Duration) ([]EmailContent, error) {
	emails, err := m.GetLatestEmailsAll(ctx, count*2)
	if err != nil {
		return nil, err
	}
	if since > 0 {
		emails = filterBySince(emails, since)
	}
	if len(emails) > count {
		emails = emails[:count]
	}
	return emails, nil
}

func sortEmails(emails []EmailContent) []EmailContent {
	if len(emails) <= 1 {
		return emails
	}
	sort.Slice(emails, func(i, j int) bool {
		a := emails[i]
		b := emails[j]
		if !a.Date.IsZero() && !b.Date.IsZero() {
			return a.Date.After(b.Date)
		}
		if !a.Date.IsZero() {
			return true
		}
		if !b.Date.IsZero() {
			return false
		}
		return a.UID > b.UID
	})
	return emails
}

func filterBySince(emails []EmailContent, since time.Duration) []EmailContent {
	if since <= 0 {
		return emails
	}
	cutoff := time.Now().Add(-since)
	filtered := make([]EmailContent, 0, len(emails))
	for _, email := range emails {
		if email.Date.IsZero() {
			continue
		}
		if email.Date.Before(cutoff) {
			continue
		}
		filtered = append(filtered, email)
	}
	return filtered
}

func (m *Manager) TestConnection(ctx context.Context, account config.Account) (time.Duration, error) {
	if account.IMAPHost == "" || account.Email == "" || account.AuthToken == "" {
		return 0, errors.New("imap connection info is incomplete")
	}

	start := time.Now()
	c, err := client.DialTLS(account.IMAPHost, nil)
	if err != nil {
		return 0, err
	}
	defer c.Logout()

	if err := c.Login(account.Email, account.AuthToken); err != nil {
		slog.Warn("imap test login failed", "email", account.Email, "error", err)
		return 0, err
	}

	idClient := id.NewClient(c)
	_, _ = idClient.ID(id.ID{
		id.FieldName:    "claw-mail-monitor",
		id.FieldVersion: "0.1.0",
	})

	if _, err := c.Select("INBOX", false); err != nil {
		slog.Warn("imap test select failed", "email", account.Email, "error", err)
		return 0, err
	}

	latency := time.Since(start)
	slog.Info("imap test connection ok", "email", account.Email, "latency_ms", latency.Milliseconds())
	return latency, nil
}

func (m *Manager) monitorLoop(ctx context.Context, account config.Account) {
	slog.Info("imap monitoring started", "email", account.Email)

	for {
		if err := m.monitorOnce(ctx, account); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			slog.Warn("imap monitor error", "email", account.Email, "error", err)
			slog.Info("imap monitoring retry", "email", account.Email, "after", "30s")
			select {
			case <-ctx.Done():
				return
			case <-time.After(30 * time.Second):
			}
		}
	}
}

func (m *Manager) monitorOnce(ctx context.Context, account config.Account) error {
	c, err := client.DialTLS(account.IMAPHost, nil)
	if err != nil {
		return err
	}
	defer c.Logout()

	if err := c.Login(account.Email, account.AuthToken); err != nil {
		return err
	}

	idClient := id.NewClient(c)
	if _, err := idClient.ID(id.ID{
		id.FieldName:    "claw-mail-monitor",
		id.FieldVersion: "0.1.0",
	}); err != nil {
		slog.Warn("imap id command failed", "email", account.Email, "error", err)
	}

	mailbox, err := c.Select("INBOX", false)
	if err != nil {
		return err
	}

	slog.Info("imap connected", "email", account.Email, "host", account.IMAPHost)

	lastUID, err := m.initLastUID(c, account, mailbox)
	if err != nil {
		slog.Warn("imap init last uid failed", "email", account.Email, "error", err)
		lastUID = 0
	}

	for {
		if err := m.fetchNewMessages(ctx, c, account, &lastUID); err != nil {
			return err
		}

		slog.Info("imap polling", "email", account.Email, "interval", m.pollInterval.String())
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(m.pollInterval):
		}
	}
}

func (m *Manager) fetchNewMessages(ctx context.Context, c *client.Client, account config.Account, lastUID *uint32) error {
	if c == nil {
		return errors.New("imap client is nil")
	}
	if lastUID == nil {
		return errors.New("last uid pointer is nil")
	}

	startUID := *lastUID + 1
	seqset := new(imap.SeqSet)
	if startUID <= 1 {
		seqset.AddRange(1, 0)
	} else {
		seqset.AddRange(startUID, 0)
	}

	slog.Debug("imap fetch start", "email", account.Email, "last_uid", *lastUID, "start_uid", startUID)

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchUid, section.FetchItem()}

	messages := make(chan *imap.Message, 10)
	errCh := make(chan error, 1)
	go func() {
		errCh <- c.UidFetch(seqset, items, messages)
	}()

	fetched := 0
	skipped := 0
	parsedOK := 0
	delivered := 0
	for msg := range messages {
		if msg == nil || msg.Uid == 0 {
			continue
		}
		fetched++
		if msg.Uid <= *lastUID {
			slog.Debug("imap skip old message", "email", account.Email, "uid", msg.Uid, "last_uid", *lastUID)
			skipped++
			continue
		}
		body := msg.GetBody(section)
		parsed, err := parser.ParseMessage(msg, body)
		if err != nil {
			slog.Warn("parse email failed", "email", account.Email, "error", err)
			continue
		}
		parsedOK++

		if parsed.Date.IsZero() {
			parsed.Date = time.Now()
		}

		slog.Info("new email received", "email", account.Email, "uid", parsed.UID, "from", parsed.From, "subject", parsed.Subject)

		if m.webhook != nil {
			notification := webhook.FormatNotification(parsed)
			if err := m.webhook.SendNotification(ctx, notification); err != nil {
				slog.Warn("webhook send failed", "email", account.Email, "uid", parsed.UID, "error", err)
			} else {
				slog.Info("webhook delivered", "email", account.Email, "uid", parsed.UID)
				delivered++
			}
		}

		m.logs.Add(EmailLog{
			From:       parsed.From,
			Subject:    parsed.Subject,
			ReceivedAt: parsed.Date,
			Summary:    parsed.Summary,
		})

		if msg.Uid > *lastUID {
			*lastUID = msg.Uid
		}
	}

	if fetched > 0 {
		slog.Info("imap fetch summary", "email", account.Email, "fetched", fetched, "parsed", parsedOK, "delivered", delivered, "skipped", skipped, "last_uid", *lastUID)
	} else {
		slog.Debug("imap fetch done", "email", account.Email, "fetched", fetched, "last_uid", *lastUID)
	}

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	default:
	}

	return nil
}

func (m *Manager) initLastUID(c *client.Client, account config.Account, mailbox *imap.MailboxStatus) (uint32, error) {
	var uidNext uint32
	var messages uint32
	if mailbox != nil {
		uidNext = mailbox.UidNext
		messages = mailbox.Messages
	}

	if uidNext > 1 {
		baseline := uidNext - 1
		slog.Info("imap baseline uid", "email", account.Email, "uidnext", uidNext, "messages", messages, "baseline_uid", baseline)
		return baseline, nil
	}

	if messages == 0 {
		slog.Info("imap baseline uid", "email", account.Email, "uidnext", uidNext, "messages", messages, "baseline_uid", 0)
		return 0, nil
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(messages)
	items := []imap.FetchItem{imap.FetchUid}

	msgCh := make(chan *imap.Message, 1)
	errCh := make(chan error, 1)
	go func() {
		errCh <- c.Fetch(seqset, items, msgCh)
	}()

	var baseline uint32
	for msg := range msgCh {
		if msg != nil && msg.Uid > 0 {
			baseline = msg.Uid
		}
	}

	select {
	case err := <-errCh:
		if err != nil {
			return 0, err
		}
	default:
	}

	slog.Info("imap baseline uid", "email", account.Email, "uidnext", uidNext, "messages", messages, "baseline_uid", baseline)
	return baseline, nil
}
