package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kiry163/claw-pliers/internal/config"
	"github.com/kiry163/claw-pliers/internal/service"
)

type MailHandler struct {
	cfg     *config.Config
	Service *service.MailService
}

func NewMailHandler(cfg *config.Config, svc *service.MailService) *MailHandler {
	return &MailHandler{cfg: cfg, Service: svc}
}

func (h *MailHandler) TestConnection(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		Error(c, http.StatusBadRequest, 10001, "email is required")
		return
	}

	latency, err := h.Service.TestConnection(email)
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

	err := h.Service.SendMail(req.From, req.To, req.Subject, req.Body)
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

	emails, err := h.Service.GetLatestEmails(email, count)
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
	accounts := h.Service.ListAccounts()
	OK(c, gin.H{
		"accounts": accounts,
	})
}
