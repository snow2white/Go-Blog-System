package ioc

import (
	"basic-go/webook/pkg/limiter"
	"os"
	"time"

	"basic-go/webook/internal/service/sms"
	"basic-go/webook/internal/service/sms/auth"
	"basic-go/webook/internal/service/sms/failover"
	"basic-go/webook/internal/service/sms/localsms"
	"basic-go/webook/internal/service/sms/tencent"

	"github.com/redis/go-redis/v9"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentSMS "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

// func InitSMSService(cmd redis.Cmdable) sms.Service {
// 	// interval time.Duration, rate int
// 	// func InitSMSService() sms.Service {
// 	return ratelimit.NewRateLimitSMSService(localsms.NewService(), limiter.NewRedisSlidingWindowLimiter(cmd, time.Second, 1000))

//		// return localsms.NewService()
//		// 如果有需要，就可以用这个
//		// return initTencentSMSService()
//	}
//
// InitSMSService 初始化 SMS 服务
func InitSMSService(cmd redis.Cmdable) sms.Service {
	// 初始化限流器
	rateLimiter := limiter.NewRedisSlidingWindowLimiter(cmd, time.Second, 1000)

	// 初始化多个 SMS 服务实例
	// tencentSvc := initTencentSMSService()
	localSvc := localsms.NewService()

	// 创建 TimeoutFailoverSMSService 实例
	// timeoutFailoverSvc := failover.NewTimeoutFailoverSMSService([]sms.Service{localSvc, tencentSvc}, 3, rateLimiter)
	timeoutFailoverSvc := failover.NewTimeoutFailoverSMSService([]sms.Service{localSvc}, 3, rateLimiter)

	// 创建 SMSService 实例
	var JWTKey = []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgK")
	// authSvc := auth.NewSMSService(timeoutFailoverSvc, JWTKey)
	return auth.NewSMSService(timeoutFailoverSvc, JWTKey)
	// return ratelimit.NewRateLimitSMSService(authSvc, rateLimiter)

}
func initTencentSMSService() sms.Service {
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		panic("找不到腾讯 SMS 的 secret id")
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")
	if !ok {
		panic("找不到腾讯 SMS 的 secret key")
	}
	c, err := tencentSMS.NewClient(
		common.NewCredential(secretId, secretKey),
		"ap-nanjing",
		profile.NewClientProfile(),
	)
	if err != nil {
		panic(err)
	}
	return tencent.NewService(c, "1400842696", "妙影科技")
}
