package ratelimit

import (
	"basic-go/webook/internal/service"
	"basic-go/webook/pkg/limiter"
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrLimited = errors.New("触发限流")

type RateLimitSMSService struct {
	// authsvc auth.SMSService
	codeSvc service.CodeService
	limiter limiter.Limiter

	key string
}

func NewRateLimitSMSService(codesvc service.CodeService, cmd redis.Cmdable) *RateLimitSMSService {
	return &RateLimitSMSService{
		codeSvc: codesvc,
		limiter: limiter.NewRedisSlidingWindowLimiter(cmd, time.Second, 1000),
		key:     "sms-limiter",
	}
}

func (r *RateLimitSMSService) Send(ctx context.Context, biz, phone string) error {
	limited, err := r.limiter.Limit(ctx, "sms-limiter")
	// return ErrLimited

	if err != nil {
		// 系统错误可以限流，也可以不限流
		// 可以限流：保守策略，你的下游很坑
		// 不可以限流：激进策略，你的下游很强，业务可用性要求很高就不限流
		return err
	}
	if limited {
		return ErrLimited
	}
	return r.codeSvc.Send(ctx, biz, phone)
}

// package ratelimit

// import (
// 	"context"
// 	"errors"

// 	"basic-go/webook/internal/service/sms"
// 	"basic-go/webook/internal/service/sms/auth"
// 	"basic-go/webook/pkg/limiter"
// )

// var errLimited = errors.New("触发限流")

// // var _ sms.Service = &RateLimitSMSService{}
// // var _ auth.SMSService = &RateLimitSMSService{}

// type RateLimitSMSService struct {
// 	authsvc auth.SMSService
// 	limiter limiter.Limiter
// 	key     string
// }

// type RateLimitSMSServiceV1 struct {
// 	sms.Service
// 	limiter limiter.Limiter
// 	key     string
// }

// func NewRateLimitSMSService(authsvc *auth.SMSService, l limiter.Limiter) *RateLimitSMSService {
// 	return &RateLimitSMSService{
// 		authsvc: *authsvc,
// 		limiter: l,
// 		key:     "sms-limiter",
// 	}
// }

// func (r *RateLimitSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
// 	limited, err := r.limiter.Limit(ctx, r.key)
// 	// return errLimited

// 	if err != nil {
// 		// 系统错误可以限流，也可以不限流
// 		// 可以限流：保守策略，你的下游很坑
// 		// 不可以限流：激进策略，你的下游很强，业务可用性要求很高就不限流
// 		return err
// 	}
// 	if limited {
// 		return errLimited
// 	}
// 	return r.authsvc.Send(ctx, tplId, args, numbers...)
// }

// // func NewRateLimitSMSService(svc sms.Service,
// // 	l limiter.Limiter) *RateLimitSMSService {
// // 	return &RateLimitSMSService{
// // 		svc:     svc,
// // 		limiter: l,
// // 		key:     "sms-limiter",
// // 	}
// // }
