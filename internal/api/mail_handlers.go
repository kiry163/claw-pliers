package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kiry163/claw-pliers/internal/config"
	"github.com/kiry163/claw-pliers/internal/mail"
)

type MailHandler struct {
	cfg *config.Config
}

func NewMailHandler(cfg *config.Config) *MailHandler {
	return &MailHandler{cfg: cfg}
}

func (h *MailHandler) TestConnection(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		Error(c, http.StatusBadRequest, 10001, "email is required")
		return
	}

	latency, err := mail.TestConnection(email)
	if err != nil {
		Error(c, http.StatusInternalServerError, 10002, err.Error())
		return
	}

	OK(c, gin.H{
		"email":   email,
		"latency": latency,
		"status":  "ok",
	})
}

type SendMailRequest struct {
	From    string `json:"from" binding:"required"`
	To      string `json:"to" binding:"required"`
	Subject string `json:"subject" binding:"required"`
	Body    string `json:"body" binding:"required"`
}

func (h *MailHandler) SendMail(c *gin.Context) {
	var req SendMailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 10001, "invalid request body")
		return
	}

	err := mail.SendMail(req.From, req.To, req.Subject, req.Body)
	if err != nil {
		Error(c, http.StatusInternalServerError, 10002, err.Error())
		return
	}

	OK(c, gin.H{
		"status":  "ok",
		"message": "email sent successfully",
	})
}

func (h *MailHandler) GetLatestEmails(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		Error(c, http.StatusBadRequest, 10001, "email is required")
		return
	}

	count := 5
	if countStr := c.Query("count"); countStr != "" {
		_ = countStr // Just to avoid unused warning, use default count
	}

	emails, err := mail.GetLatestEmails(email, count)
	if err != nil {
		Error(c, http.StatusInternalServerError, 10002, err.Error())
		return
	}

	OK(c, gin.H{
		"emails": emails,
		"count":  len(emails),
	})
}

func (h *MailHandler) ListAccounts(c *gin.Context) {
	accounts := mail.ListAccounts()
	OK(c, gin.H{
		"accounts": accounts,
	})
}
