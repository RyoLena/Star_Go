// Package middleware pkg/middleware/recovery.go
package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery 恢复中间件 - 处理程序崩溃，避免整个服务器宕机
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录堆栈信息
				stack := debug.Stack()
				log.Printf("[PANIC RECOVER] 错误: %v\n堆栈: %s", err, string(stack))

				// 向客户端返回500错误
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "服务器内部错误",
				})

				// 终止后续中间件
				c.Abort()
			}
		}()

		c.Next()
	}
}
