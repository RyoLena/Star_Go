// pkg/utils/response.go
package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 响应码
const (
	SUCCESS        = 200
	ERROR          = 500
	INVALID_PARAMS = 400
	UNAUTHORIZED   = 401
	FORBIDDEN      = 403
	NOT_FOUND      = 404
)

// MsgFlags 响应消息
var MsgFlags = map[int]string{
	SUCCESS:        "成功",
	ERROR:          "服务器内部错误",
	INVALID_PARAMS: "请求参数错误",
	UNAUTHORIZED:   "未授权访问",
	FORBIDDEN:      "禁止访问",
	NOT_FOUND:      "资源不存在",
}

// 获取响应消息
func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}

	return MsgFlags[ERROR]
}

// Response 标准响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    SUCCESS,
		Message: GetMsg(SUCCESS),
		Data:    data,
	})
}

// SuccessWithMessage 成功响应带自定义消息
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    SUCCESS,
		Message: message,
		Data:    data,
	})
}

// Fail 失败响应
func Fail(c *gin.Context, code int, data interface{}) {
	c.JSON(getHttpStatusByCode(code), Response{
		Code:    code,
		Message: GetMsg(code),
		Data:    data,
	})
}

// FailWithMessage 失败响应带自定义消息
func FailWithMessage(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(getHttpStatusByCode(code), Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// 根据业务码获取HTTP状态码
func getHttpStatusByCode(code int) int {
	switch code {
	case INVALID_PARAMS:
		return http.StatusBadRequest
	case UNAUTHORIZED:
		return http.StatusUnauthorized
	case FORBIDDEN:
		return http.StatusForbidden
	case NOT_FOUND:
		return http.StatusNotFound
	case ERROR:
		return http.StatusInternalServerError
	default:
		return http.StatusOK
	}
}
