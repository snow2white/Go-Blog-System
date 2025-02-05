package auth

import (
	"context"
	"fmt"

	"basic-go/webook/internal/service/sms"

	"github.com/golang-jwt/jwt/v5"
)

// type SMSService interface {
// 	Send(ctx context.Context, tplId string, args []string, numbers ...string) error
// }

// SMSService 是一个包装了 SMS 服务的结构体，用于在发送短信之前验证 JWT token。
type SMSService struct {
	svc sms.Service // 原始的 SMS 服务接口
	key []byte      // 用于解析 JWT token 的密钥
}

func NewSMSService(timeoutFailoverSvc sms.Service, key []byte) *SMSService {
	return &SMSService{
		svc: timeoutFailoverSvc,
		key: key,
	}
}

// Send 方法用于发送短信。它首先解析传入的模板令牌 (tplToken)，然后使用原始的 SMS 服务发送短信。
// 如果 tplToken 解析失败，则返回错误。
func (s *SMSService) Send(ctx context.Context, tplToken string, args []string, numbers ...string) error {
	var claims SMSClaims // 定义一个 SMSClaims 类型的变量来存储解析后的声明

	// 解析 JWT token 并提取其中的声明
	fmt.Println("tplToken:", tplToken)
	token, err := jwt.ParseWithClaims(tplToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil // 使用提供的密钥解析 JWT token
	})
	fmt.Println(token)
	if err != nil {
		// 如果解析失败，返回错误
		return err
	}

	// 使用解析后的模板 ID 和其他参数调用原始的 SMS 服务发送短信
	return s.svc.Send(ctx, claims.Tpl, args, numbers...)
}

// SMSClaims 表示 JWT token 中的声明，包含标准的注册声明和自定义的模板 ID 字段。
type SMSClaims struct {
	jwt.RegisteredClaims        // 标准的 JWT 注册声明
	Tpl                  string // 自定义字段：模板 ID
	// 额外加字段（如果需要）
}
