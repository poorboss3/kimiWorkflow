// Package response 统一响应
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"workflow-engine/internal/pkg/errors"
)

// Response 统一响应结构
type Response struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageResult 分页结果
type PageResult struct {
	List    interface{} `json:"list"`
	Total   int         `json:"total"`
	Page    int         `json:"page"`
	Size    int         `json:"size"`
	HasMore bool        `json:"hasMore"`
}

// NewPageResult 创建分页结果
func NewPageResult(list interface{}, total, page, size int) *PageResult {
	return &PageResult{
		List:    list,
		Total:   total,
		Page:    page,
		Size:    size,
		HasMore: total > page*size,
	}
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    "0",
		Message: "success",
		Data:    data,
	})
}

// SuccessWithPage 成功响应（分页）
func SuccessWithPage(c *gin.Context, list interface{}, total, page, size int) {
	c.JSON(http.StatusOK, Response{
		Code:    "0",
		Message: "success",
		Data:    NewPageResult(list, total, page, size),
	})
}

// Error 错误响应
func Error(c *gin.Context, code errors.ErrorCode, message string) {
	if message == "" {
		message = errors.ErrorMessages[code]
	}
	c.JSON(http.StatusOK, Response{
		Code:    string(code),
		Message: message,
	})
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(c *gin.Context, code errors.ErrorCode, message string, data interface{}) {
	if message == "" {
		message = errors.ErrorMessages[code]
	}
	c.JSON(http.StatusOK, Response{
		Code:    string(code),
		Message: message,
		Data:    data,
	})
}

// HTTPError HTTP状态码错误
func HTTPError(c *gin.Context, status int, message string) {
	c.JSON(status, Response{
		Code:    string(errors.ErrInternal),
		Message: message,
	})
}

// HandleError 处理错误
func HandleError(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		if appErr.Data != nil {
			ErrorWithData(c, appErr.Code, appErr.Message, appErr.Data)
		} else {
			Error(c, appErr.Code, appErr.Message)
		}
		return
	}
	Error(c, errors.ErrInternal, err.Error())
}
