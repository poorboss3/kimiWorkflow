package middleware

import (
	"github.com/gin-gonic/gin"
	"workflow-engine/api/response"
)

// AuthUserKey 用户ID上下文键
const AuthUserKey = "userID"

// Auth 认证中间件（模拟）
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取用户ID（实际应该从token解析）
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			userID = "user_001" // 默认用户，用于测试
		}

		c.Set(AuthUserKey, userID)
		c.Next()
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) string {
	userID, _ := c.Get(AuthUserKey)
	if id, ok := userID.(string); ok {
		return id
	}
	return ""
}

// RequireAuth 需要认证
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			response.Error(c, "60001", "未登录")
			c.Abort()
			return
		}
		c.Set(AuthUserKey, userID)
		c.Next()
	}
}
