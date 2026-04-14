package middleware

import (
	"strings"

	"seckill-demo/utils"

	"github.com/gin-gonic/gin"
)

// JWTAuth 鉴权中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从请求头获取 Authorization 字段
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(200, gin.H{"code": 401, "msg": "请求未携带 Token，请先登录"})
			c.Abort() // 拦截请求，不再往下执行
			return
		}

		// 2. 按约定的 "Bearer xxxx" 格式提取 Token
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(200, gin.H{"code": 401, "msg": "Token 格式错误"})
			c.Abort()
			return
		}

		// 3. 解析 Token
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			c.JSON(200, gin.H{"code": 401, "msg": "Token 无效或已过期，请重新登录"})
			c.Abort()
			return
		}

		// 4. 解析成功，将 UserID 存入 Gin 的上下文 (Context) 中，供后续 Controller 使用
		c.Set("user_id", claims.UserID)
		c.Next() // 放行
	}
}
