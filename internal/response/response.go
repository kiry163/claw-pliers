package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorCode 定义错误码
const (
	CodeSuccess       = 0
	CodeUnauthorized  = 10001
	CodeNotFound      = 10002
	CodeGone          = 10003
	CodeInvalidParam  = 10004
	CodeInternalError = 19999
)

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    CodeSuccess,
		"message": "success",
		"data":    data,
	})
}

// SuccessWithMessage 返回带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    CodeSuccess,
		"message": message,
		"data":    data,
	})
}

// Error 返回错误响应
func Error(c *gin.Context, status int, code int, message string) {
	c.JSON(status, gin.H{
		"code":    code,
		"message": message,
	})
}

// Message 返回简单消息响应
func Message(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{
		"code":    CodeSuccess,
		"message": message,
	})
}

// Page 返回分页响应
func Page(c *gin.Context, items interface{}, total int64) {
	c.JSON(http.StatusOK, gin.H{
		"code":    CodeSuccess,
		"message": "success",
		"data": gin.H{
			"total": total,
			"items": items,
		},
	})
}

// BadRequest 返回 400 错误
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, CodeInvalidParam, message)
}

// Unauthorized 返回 401 错误
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, CodeUnauthorized, message)
}

// NotFound 返回 404 错误
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, CodeNotFound, message)
}

// Gone 返回 410 错误
func Gone(c *gin.Context, message string) {
	Error(c, http.StatusGone, CodeGone, message)
}

// InternalError 返回 500 错误
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, CodeInternalError, message)
}
