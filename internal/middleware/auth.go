package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/starfall-warsong/sws/pkg/auth"
	"github.com/starfall-warsong/sws/pkg/response"
)

func JWTAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Unauthorized(c, "缺少认证令牌")
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "认证格式错误")
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "令牌无效或已过期")
			c.Abort()
			return
		}

		if claims.TokenType != "access" {
			response.Unauthorized(c, "请使用访问令牌")
			c.Abort()
			return
		}

		c.Set("account_id", claims.AccountID)
		c.Set("phone", claims.Phone)
		c.Next()
	}
}

func GetAccountID(c *gin.Context) int64 {
	id, _ := c.Get("account_id")
	return id.(int64)
}
