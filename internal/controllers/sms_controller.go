package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"star-go/internal/services"
	"star-go/pkg/utils"
)

type SMSController struct {
	smsService services.SMSService
}

func NewSMSController() *SMSController {
	return &SMSController{
		services.NewSMSService(),
	}
}

type SendCodeRequest struct {
	Biz   string `json:"biz" binding:"required"`
	Phone string `json:"phone" binding:"required,len=11"`
}

// SendCodeResponse 发送验证码响应
type SendCodeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// VerifyCodeRequest 验证码验证请求
type VerifyCodeRequest struct {
	Biz   string `json:"biz" binding:"required"`
	Phone string `json:"phone" binding:"required,len=11"`
	Code  string `json:"code" binding:"required,len=6"`
}

// VerifyCodeResponse 验证码验证响应
type VerifyCodeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (c *SMSController) SendCode(ctx *gin.Context) {
	var req = SendCodeRequest{}

	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		utils.FailWithMessage(ctx, utils.INVALID_PARAMS, err.Error(), nil)
		return
	}
	// 验证业务类型是否有效
	if !isValidBizType(req.Biz) {
		utils.FailWithMessage(ctx, utils.UNAUTHORIZED, "非需要使用的业务", nil)
		return
	}

	//发送验证码
	err := c.smsService.Send(ctx, req.Biz, req.Phone)
	if err != nil {
		utils.Fail(ctx, utils.ERROR, fmt.Errorf("验证码发送失败:%s", err.Error()))
		return
	}
	utils.Success(ctx, gin.H{
		"success": true,
		"message": "验证码发送成功",
	})
}

func (c *SMSController) VerifyCode(ctx *gin.Context) {
	var req VerifyCodeRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		utils.FailWithMessage(ctx, utils.INVALID_PARAMS, err.Error(), nil)
		return
	}
	if !isValidBizType(req.Biz) {
		utils.Fail(ctx, utils.INVALID_PARAMS, gin.H{
			"biz": req.Biz,
		})
		return
	}

	//验证验证码
	isVerify, err := c.smsService.Verify(ctx, req.Biz, req.Phone, req.Code)
	if err != nil {
		utils.FailWithMessage(ctx, utils.INVALID_PARAMS, err.Error(), gin.H{"code": "手机号不存在"})
		return
	}
	if !isVerify {
		utils.Fail(ctx, utils.INVALID_PARAMS, gin.H{
			"Fail":    true,
			"Message": "验证码错误",
		})
		return
	}
	utils.Success(ctx, gin.H{
		"success": true,
		"Message": "验证码验证成功",
	})
}

// 验证业务类型是否有效
func isValidBizType(biz string) bool {
	validBizTypes := map[string]bool{
		services.BizLogin:       true,
		services.BizRegister:    true,
		services.BizChangePhone: true,
		services.BizResetPwd:    true,
	}
	return validBizTypes[biz]
}
