package interceptor

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": []string{"missing authorization token"},
			})
			return
		}

		// 验证accesstoken
		// 1. 正常直接放行
		// 2. 篡改/过期等情况返回对应状态码

		// 验证refreshToken
		// 1. 正常刷新双token返回并更新redis中refreshToken
		// 2. 过期强制重新登录
		// 3. 其他异常情况返回对应状态码

	}
}