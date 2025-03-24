package services

import (
	"context"
	"fmt"
	"math/rand/v2"
	"star-go/pkg/cache"
)

// 业务类型常量
const (
	BizLogin       = "login"
	BizRegister    = "register"
	BizResetPwd    = "reset_pwd"
	BizChangePhone = "change_phone"
)

type SMSService interface {
	Send(ctx context.Context, biz, phone string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

type smsService struct {
	cache cache.SMSCodeCache

	//添加一个真正发送短信的客户端
	// smsClient SMSClient
}

func NewSMSService() SMSService {
	return &smsService{
		cache: cache.NewRedisCodeCache(),
	}
}

func (s *smsService) Send(ctx context.Context, biz, phone string) error {
	code := generateCode()

	//将代码存到缓存中
	err := s.cache.Set(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 调用短信发送SDK发送短信
	// err = s.smsClient.Send(phone, fmt.Sprintf("您的验证码是: %s, 有效期5分钟", code))
	// if err != nil {
	//     s.logger.Error("Failed to send SMS", "error", err, "phone", phone, "biz", biz)
	//     return "", err
	// }
	fmt.Printf("sms:%s", code)

	return nil
}

func (s *smsService) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	return s.cache.Verify(ctx, biz, phone, code)
}

// 生成一个随机数
func generateCode() string {
	num := rand.IntN(1000000)
	//不够6位数的，加上前导0
	return fmt.Sprintf("%06d", num)
}
