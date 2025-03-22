// pkg/middleware/logger.go
package middleware

import (
	"bytes"
	"fmt"
	"io"
	"star-go/pkg/config"
	"time"

	"github.com/gin-gonic/gin"
)

// ANSI颜色代码
const (
	Reset      = "\033[0m"
	Red        = "\033[31m"
	Green      = "\033[32m"
	Yellow     = "\033[33m"
	Blue       = "\033[34m"
	Magenta    = "\033[35m"
	Cyan       = "\033[36m"
	White      = "\033[37m"
	BoldRed    = "\033[1;31m"
	BoldGreen  = "\033[1;32m"
	BoldYellow = "\033[1;33m"
	BoldBlue   = "\033[1;34m"
)

// 获取HTTP状态码的颜色
func statusCodeColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return BoldGreen // 成功状态码为绿色
	case code >= 300 && code < 400:
		return BoldBlue // 重定向状态码为蓝色
	case code >= 400 && code < 500:
		return BoldYellow // 客户端错误状态码为黄色
	default:
		return BoldRed // 服务器错误状态码为红色
	}
}

// 获取HTTP方法的颜色
func methodColor(method string) string {
	switch method {
	case "GET":
		return Blue
	case "POST":
		return Green
	case "PUT":
		return Yellow
	case "DELETE":
		return Red
	case "PATCH":
		return Cyan
	case "HEAD":
		return Magenta
	case "OPTIONS":
		return White
	default:
		return Reset
	}
}

// Logger 日志中间件 - 记录请求和响应信息
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.GetConfig()
		// 开始时间
		startTime := time.Now()

		// 记录请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 创建自定义响应写入器
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()
		// 执行时间
		latencyTime := endTime.Sub(startTime)

		// 请求方法
		reqMethod := c.Request.Method
		// 请求路由
		reqUri := c.Request.RequestURI
		// 状态码
		statusCode := c.Writer.Status()
		// 请求IP
		clientIP := c.ClientIP()
		if cfg.Server.EnableRequestLog {
			// 获取状态码和方法的颜色
			statusColor := statusCodeColor(statusCode)
			methodColor := methodColor(reqMethod)

			// 格式化延迟时间，保留3位小数
			latencyStr := fmt.Sprintf("%.3fms", float64(latencyTime.Microseconds())/1000.0)
			// 日志格式 - 带颜色和对齐
			fmt.Printf("[%s][GIN] %s%-7s%s | %s%3d%s | %13s | %15s | %s\n",
				time.Now().Format("2006-01-02 15:04:05"),
				methodColor, reqMethod, Reset,
				statusColor, statusCode, Reset,
				latencyStr,
				clientIP,
				reqUri,
			)

			// 如果是错误状态码，记录请求体和响应体
			if statusCode >= 400 {
				fmt.Printf("%s[ERROR]%s Request: %s\n", BoldRed, Reset, string(requestBody))
				fmt.Printf("%s[ERROR]%s Response: %s\n", BoldRed, Reset, blw.body.String())
			}
		}
	}
}

// 自定义响应写入器
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// 重写Write方法
func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
