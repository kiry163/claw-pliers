package httpapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kiry163/claw-mail-monitor/internal/account"
	"github.com/kiry163/claw-mail-monitor/internal/config"
	"github.com/kiry163/claw-mail-monitor/internal/imap"
	"github.com/kiry163/claw-mail-monitor/internal/smtp"
)

type Server struct {
	addr     string
	cfg      *config.Config
	accounts *account.Manager
	monitor  *imap.Manager
	version  string
	engine   *gin.Engine
}

func NewServer(addr string, cfg *config.Config, monitor *imap.Manager, version string) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(requestLogger())

	s := &Server{
		addr:     addr,
		cfg:      cfg,
		accounts: account.NewManager(cfg),
		monitor:  monitor,
		version:  version,
		engine:   engine,
	}
	s.registerRoutes()
	return s
}

func (s *Server) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:              s.addr,
		Handler:           s.engine,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) registerRoutes() {
	s.engine.GET("/health", s.handleHealth)
	s.engine.GET("/status", s.handleStatus)
	s.engine.GET("/accounts", s.handleListAccounts)
	s.engine.POST("/accounts", s.handleAddAccount)
	s.engine.DELETE("/accounts/:email", s.handleRemoveAccount)
	s.engine.POST("/send", s.handleSend)
	s.engine.POST("/test-connection", s.handleTestConnection)
	s.engine.GET("/latest", s.handleLatest)
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "version": s.version})
}

func (s *Server) handleStatus(c *gin.Context) {
	accounts := s.accounts.List()
	monitoring := s.monitor.MonitoringCount()
	status := "stopped"
	if monitoring > 0 {
		status = "running"
	}

	configPath := s.cfg.ConfigPath
	if configPath == "" {
		if resolved, err := config.DefaultConfigPath(); err == nil {
			configPath = resolved
		}
	}

	logFile := config.ExpandPath(s.cfg.Logging.File)

	c.JSON(http.StatusOK, gin.H{
		"monitoring":  monitoring,
		"total":       len(accounts),
		"status":      status,
		"version":     s.version,
		"config_path": configPath,
		"log_file":    logFile,
	})
}

func (s *Server) handleListAccounts(c *gin.Context) {
	accounts := s.accounts.List()
	result := make([]gin.H, 0, len(accounts))
	for _, acct := range accounts {
		status := "disabled"
		if acct.Enabled && s.monitor.IsMonitoring(acct.Email) {
			status = "monitoring"
		}
		result = append(result, gin.H{
			"provider": acct.Provider,
			"email":    acct.Email,
			"enabled":  acct.Enabled,
			"status":   status,
		})
	}

	c.JSON(http.StatusOK, gin.H{"accounts": result, "total": len(accounts)})
}

type addAccountInput struct {
	Provider  string `json:"provider"`
	Email     string `json:"email"`
	AuthToken string `json:"auth_token"`
}

func (s *Server) handleAddAccount(c *gin.Context) {
	var input addAccountInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	acct := config.Account{
		Provider:  strings.TrimSpace(input.Provider),
		Email:     strings.TrimSpace(input.Email),
		AuthToken: strings.TrimSpace(input.AuthToken),
		Enabled:   true,
	}
	s.cfg.ApplyDefaults(&acct)

	if _, err := s.monitor.TestConnection(c.Request.Context(), acct); err != nil {
		respondError(c, http.StatusBadRequest, fmt.Errorf("imap connection failed: %w", err))
		return
	}

	if err := s.accounts.Add(acct); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	s.monitor.StartAccount(c.Request.Context(), acct)
	slog.Info("account added", "email", acct.Email, "provider", acct.Provider)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("Account %s added successfully", acct.Email)})
}

func (s *Server) handleRemoveAccount(c *gin.Context) {
	email := c.Param("email")
	if email != "" {
		if decoded, err := url.PathUnescape(email); err == nil {
			email = decoded
		}
	}

	if strings.TrimSpace(email) == "" {
		respondError(c, http.StatusBadRequest, errors.New("email is required"))
		return
	}

	s.monitor.StopAccount(email)
	if err := s.accounts.Remove(email); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	slog.Info("account removed", "email", email)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("Account %s removed successfully", email)})
}

func (s *Server) handleSend(c *gin.Context) {
	var req smtp.SendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	account, ok := s.accounts.FirstEnabled()
	if !ok {
		respondError(c, http.StatusBadRequest, errors.New("no enabled account"))
		return
	}

	if err := smtp.Send(c.Request.Context(), account, req); err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "mail sent"})
}

type testConnectionInput struct {
	Email string `json:"email"`
}

func (s *Server) handleTestConnection(c *gin.Context) {
	var input testConnectionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	email := strings.TrimSpace(input.Email)
	account, ok := s.accounts.Find(email)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"success": false, "status": "not_found", "latency_ms": 0})
		return
	}

	latency, err := s.monitor.TestConnection(c.Request.Context(), account)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "status": "failed", "latency_ms": latency.Milliseconds()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "status": "connected", "latency_ms": latency.Milliseconds()})
}

func (s *Server) handleLatest(c *gin.Context) {
	email := strings.TrimSpace(c.Query("email"))
	count := 1
	if v := c.Query("count"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			count = parsed
		}
	}

	if count <= 0 {
		count = 1
	}
	if count > 10 {
		count = 10
	}

	since := time.Duration(0)
	if v := strings.TrimSpace(c.Query("since")); v != "" {
		parsed, err := time.ParseDuration(v)
		if err != nil {
			respondError(c, http.StatusBadRequest, err)
			return
		}
		since = parsed
	}

	if email == "" {
		if since > 0 {
			emails, err := s.monitor.GetLatestEmailsAllSince(c.Request.Context(), count, since)
			if err != nil {
				respondError(c, http.StatusInternalServerError, err)
				return
			}
			c.JSON(http.StatusOK, gin.H{"emails": emails, "count": len(emails), "email": "all"})
			return
		}
		emails, err := s.monitor.GetLatestEmailsAll(c.Request.Context(), count)
		if err != nil {
			respondError(c, http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"emails": emails, "count": len(emails), "email": "all"})
		return
	}

	account, ok := s.accounts.Find(email)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"emails": []imap.EmailContent{}, "count": 0, "email": email})
		return
	}

	var emails []imap.EmailContent
	var err error
	if since > 0 {
		emails, err = s.monitor.GetLatestEmailsSince(c.Request.Context(), account, count, since)
	} else {
		emails, err = s.monitor.GetLatestEmails(c.Request.Context(), account, count)
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"emails": emails, "count": len(emails), "email": email})
}

func respondError(c *gin.Context, status int, err error) {
	slog.Warn("http request failed", "method", c.Request.Method, "path", c.FullPath(), "status", status, "error", err)
	c.JSON(status, gin.H{"error": err.Error()})
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path

		if len(c.Errors) > 0 {
			slog.Warn("http request error", "method", method, "path", path, "status", status, "latency_ms", latency.Milliseconds(), "error", c.Errors.String())
			return
		}

		slog.Info("http request", "method", method, "path", path, "status", status, "latency_ms", latency.Milliseconds())
	}
}
