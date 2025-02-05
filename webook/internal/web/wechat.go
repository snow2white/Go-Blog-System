package web

import (
	"basic-go/webook/internal/service"
	"basic-go/webook/internal/service/oauth2/wechat"
	ijwt "basic-go/webook/internal/web/jwt"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
)

// OAuth2WechatHandler 处理微信OAuth2授权的HTTP请求
type OAuth2WechatHandler struct {
	svc             wechat.Service      // 微信服务
	userSvc         service.UserService // 用户服务
	ijwt.Handler                        // JWT处理器
	key             []byte              // JWT密钥
	stateCookieName string              // 用于存储state的cookie名称
}

// NewOAuth2WechatHandler 创建一个新的OAuth2WechatHandler实例
func NewOAuth2WechatHandler(svc wechat.Service,
	hdl ijwt.Handler,
	userSvc service.UserService) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:             svc,
		userSvc:         userSvc,
		key:             []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgB"), // 硬编码的JWT密钥，建议从环境变量中获取
		stateCookieName: "jwt-state",
		Handler:         hdl,
	}
}

// RegisterRoutes 注册微信OAuth2相关的路由
func (o *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", o.Auth2URL)  // 获取微信授权URL
	g.Any("/callback", o.Callback) // 处理微信授权回调
}

// Auth2URL 生成微信授权URL并返回给客户端
func (o *OAuth2WechatHandler) Auth2URL(ctx *gin.Context) {
	state := uuid.New() // 生成一个唯一的state
	val, err := o.svc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "构造跳转URL失败",
			Code: 5,
		})
		return
	}
	err = o.setStateCookie(ctx, state) // 设置state到cookie中
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "服务器异常",
			Code: 5,
		})
	}
	ctx.JSON(http.StatusOK, Result{
		Data: val,
	})
}

// Callback 处理微信授权回调，验证授权码并登录用户
func (o *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	err := o.verifyState(ctx) // 验证state是否匹配
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "非法请求",
			Code: 4,
		})
		return
	}
	code := ctx.Query("code")                      // 获取授权码
	wechatInfo, err := o.svc.VerifyCode(ctx, code) // 验证授权码并获取用户信息
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "授权码有误",
			Code: 4,
		})
		return
	}
	u, err := o.userSvc.FindOrCreateByWechat(ctx, wechatInfo) // 根据微信信息查找或创建用户
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
		return
	}
	err = o.SetLoginToken(ctx, u.Id) // 设置登录token
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

// verifyState 验证state是否匹配
func (o *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")              // 获取state参数
	ck, err := ctx.Cookie(o.stateCookieName) // 获取state cookie
	if err != nil {
		return fmt.Errorf("无法获得 cookie %w", err)
	}
	var sc StateClaims
	_, err = jwt.ParseWithClaims(ck, &sc, func(token *jwt.Token) (interface{}, error) {
		return o.key, nil
	})
	if err != nil {
		return fmt.Errorf("解析 token 失败 %w", err)
	}
	if state != sc.State {
		return fmt.Errorf("state 不匹配")
	}
	return nil
}

// setStateCookie 设置state到cookie中
func (o *OAuth2WechatHandler) setStateCookie(ctx *gin.Context, state string) error {
	claims := StateClaims{
		State: state,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString(o.key)
	if err != nil {
		return err
	}
	ctx.SetCookie(o.stateCookieName, tokenStr,
		600, "/oauth2/wechat/callback",
		"", false, true)
	return nil
}

// StateClaims 自定义JWT声明结构体，包含state字段
type StateClaims struct {
	jwt.RegisteredClaims
	State string
}
