package handlers

import (
	"github.com/gin-gonic/gin"
	"star-go/internal/controllers"
)

// SMSHandler 短信处理器
type SMSHandler struct {
	controller *controllers.SMSController
}

// NewSMSHandler 创建短信处理器
func NewSMSHandler(controller *controllers.SMSController) *SMSHandler {
	return &SMSHandler{
		controller: controller,
	}
}

// SendCode 发送验证码
func (h *SMSHandler) SendCode(ctx *gin.Context) {
	h.controller.SendCode(ctx)
}

// VerifyCode 验证验证码
func (h *SMSHandler) VerifyCode(ctx *gin.Context) {
	h.controller.VerifyCode(ctx)
}
