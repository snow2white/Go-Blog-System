package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"basic-go/webook/internal/repository"
	"basic-go/webook/internal/service/sms"
	// "basic-go/webook/internal/service/sms/ratelimit"
)

var ErrCodeSendTooMany = repository.ErrCodeSendTooMany
var ErrLimited = errors.New("触发限流")

// var ErrCodeSendTooMany = repository.ErrCodeSendTooMany
type CodeService interface {
	Send(ctx context.Context, biz, phone string) error
	Verify(ctx context.Context,
		biz, phone, inputCode string) (bool, error)
}
type codeService struct {
	repo repository.CodeRepository
	sms  sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &codeService{
		repo: repo,
		sms:  smsSvc,
	}
}

func (svc *codeService) Send(ctx context.Context, biz, phone string) error {

	code := svc.generate()
	err := svc.repo.Set(ctx, biz, phone, code)
	// 你在这儿，是不是要开始发送验证码了？
	if err != nil {
		return err
	}
	// 20天后过期，2025年1.29日生成的
	const codeTplId = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mzk5NzU3OTUsIlVpZCI6MiwiVXNlckFnZW50IjoiQXBpZm94LzEuMC4wIChodHRwczovL2FwaWZveC5jb20pIn0.AxqEdcit2f89dz3qsJUw32Pl3SBaSXytUf5H-p4r1jy03XyrMKRu-_ajUPYAnmOPv0I55E9xDPOlBttvAgGmLg"
	return svc.sms.Send(ctx, codeTplId, []string{code}, phone)
}

func (svc *codeService) Verify(ctx context.Context,
	biz, phone, inputCode string) (bool, error) {
	ok, err := svc.repo.Verify(ctx, biz, phone, inputCode)
	if err == repository.ErrCodeVerifyTooMany {
		// 相当于，我们对外面屏蔽了验证次数过多的错误，我们就是告诉调用者，你这个不对
		return false, nil
	}
	return ok, err
}

func (svc *codeService) generate() string {
	// 0-999999
	code := rand.Intn(1000000)
	// fmt.Println("生成的验证码是", code)
	return fmt.Sprintf("%06d", code)
}
