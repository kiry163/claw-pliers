package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kiry163/claw-pliers/internal/config"
	"github.com/kiry163/claw-pliers/internal/logger"
)

type Handler struct {
	Config *config.Config
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()

		log := logger.Get()
		if status >= 400 {
			log.Error().
				Str("method", method).
				Str("path", path).
				Int("status", status).
				Dur("latency", latency).
				Str("ip", clientIP).
				Msg("request failed")
		} else {
			log.Info().
				Str("method", method).
				Str("path", path).
				Int("status", status).
				Dur("latency", latency).
				Str("ip", clientIP).
				Msg("request completed")
		}
	}
}

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			_ = tokenString
			c.Set("user", "bearer")
			c.Next()
			return
		}

		localKey := c.GetHeader("X-Local-Key")
		if localKey != "" && localKey == cfg.Auth.LocalKey {
			c.Set("user", "local")
			c.Next()
			return
		}

		Error(c, http.StatusUnauthorized, 10001, "unauthorized")
		c.Abort()
	}
}

func Error(c *gin.Context, status int, code int, message string) {
	c.JSON(status, gin.H{
		"code":    code,
		"message": message,
	})
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"data": data,
	})
}

func Message(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"message": msg,
	})
}

func getUser(c *gin.Context) string {
	if user, exists := c.Get("user"); exists {
		return user.(string)
	}
	return "unknown"
}
