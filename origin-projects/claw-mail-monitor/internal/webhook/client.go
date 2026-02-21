package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/kiry163/claw-mail-monitor/internal/config"
)

const (
	defaultName    = "EmailMonitor"
	defaultChannel = "feishu"
)

type Client struct {
	cfg        *config.Webhook
	httpClient *http.Client
}

type payload struct {
	Message    string `json:"message"`
	Name       string `json:"name"`
	Deliver    bool   `json:"deliver"`
	Channel    string `json:"channel"`
	To         string `json:"to,omitempty"`
	SessionKey string `json:"session_key,omitempty"`
}

func NewClient(cfg *config.Webhook) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) SendNotification(ctx context.Context, message string) error {
	if c == nil || c.cfg == nil {
		return errors.New("webhook client not configured")
	}
	if !c.cfg.Enable {
		return nil
	}
	if c.cfg.URL == "" {
		return errors.New("webhook url is empty")
	}

	customPayload := strings.TrimSpace(c.cfg.CustomPayload)
	var body []byte
	if customPayload != "" {
		if !json.Valid([]byte(customPayload)) {
			slog.Warn("invalid custom webhook payload, fallback to default")
		} else {
			body = []byte(customPayload)
			slog.Info("using custom webhook payload")
		}
	}

	if len(body) == 0 {
		var err error
		body, err = json.Marshal(payload{
			Message:    message,
			Name:       defaultName,
			Deliver:    true,
			Channel:    defaultChannel,
			To:         c.cfg.To,
			SessionKey: c.cfg.SessionKey,
		})
		if err != nil {
			return fmt.Errorf("marshal webhook payload failed: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build webhook request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.Token)
	}

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Warn("webhook request failed", "error", err)
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Warn("webhook response error", "status", resp.Status)
		return fmt.Errorf("webhook returned status %s", resp.Status)
	}

	latency := time.Since(start)
	slog.Info("webhook request ok", "status", resp.Status, "latency_ms", latency.Milliseconds())

	return nil
}
