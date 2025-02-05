// package ratelimit

// import (
// 	"basic-go/webook/pkg/limiter"
// 	_ "embed"
// 	"fmt"
// 	"log"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// )

// type Builder struct {
// 	prefix  string
// 	limiter limiter.Limiter
// }

// func NewBuilder(l limiter.Limiter) *Builder {
// 	return &Builder{
// 		prefix:  "ip-limiter",
// 		limiter: l,
// 	}
// }

// func (b *Builder) Prefix(prefix string) *Builder {
// 	b.prefix = prefix
// 	return b
// }

//	func (b *Builder) Build() gin.HandlerFunc {
//		return func(ctx *gin.Context) {
//			limited, err := b.limiter.Limit(ctx, fmt.Sprintf("%s:%s", b.prefix, ctx.ClientIP()))
//			if err != nil {
//				log.Println(err)
//				// 这一步很有意思，就是如果这边出错了
//				// 要怎么办？
//				// 保守做法：因为借助于 Redis 来做限流，那么 Redis 崩溃了，为了防止系统崩溃，直接限流
//				ctx.AbortWithStatus(http.StatusInternalServerError)
//				// 激进做法：虽然 Redis 崩溃了，但是这个时候还是要尽量服务正常的用户，所以不限流
//				// ctx.Next()
//				return
//			}
//			if limited {
//				log.Println(err)
//				ctx.AbortWithStatus(http.StatusTooManyRequests)
//				return
//			}
//			ctx.Next()
//		}
//	}
package ratelimit

import (
	"basic-go/webook/pkg/limiter"
	_ "embed"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Builder 是一个构建限流中间件的结构体。
// 它通过指定的 Limiter 实现来控制请求的速率。
type Builder struct {
	prefix  string          // 前缀，用于标识不同类型的限流规则
	limiter limiter.Limiter // 限流器实例，用于执行限流逻辑
}

// NewBuilder 创建一个新的 Builder 实例。
// 参数：
// - l: 限流器实例
// 返回：
// - *Builder: Builder 实例
func NewBuilder(l limiter.Limiter) *Builder {
	return &Builder{
		prefix:  "ip-limiter", // 默认前缀
		limiter: l,
	}
}

// Prefix 设置限流规则的前缀。
// 这个前缀可以用来区分不同的限流规则。
// 参数：
// - prefix: 前缀字符串
// 返回：
// - *Builder: Builder 实例，支持链式调用
func (b *Builder) Prefix(prefix string) *Builder {
	b.prefix = prefix
	return b
}

// Build 构建并返回一个 Gin 中间件函数。
// 该中间件会在每个请求到达时检查是否超过限流阈值。
// 如果超过阈值，会返回 429 Too Many Requests 状态码。
// 如果限流器操作失败，可以选择保守或激进的做法：
// - 保守做法：因为借助于 Redis 来做限流，那么 Redis 崩溃了，为了防止系统崩溃，直接限流
// - 激进做法：虽然 Redis 崩溃了，但是这个时候还是要尽量服务正常的用户，所以不限流
// 返回：
// - gin.HandlerFunc: Gin 中间件函数
func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 构建唯一的限流键
		key := fmt.Sprintf("%s:%s", b.prefix, ctx.ClientIP())
		// 检查是否超过限流阈值
		limited, err := b.limiter.Limit(ctx, key)
		if err != nil {
			log.Println(err)
			// 保守做法：因为借助于 Redis 来做限流，那么 Redis 崩溃了，为了防止系统崩溃，直接限流
			ctx.AbortWithStatus(http.StatusInternalServerError)
			// 激进做法：虽然 Redis 崩溃了，但是这个时候还是要尽量服务正常的用户，所以不限流
			// ctx.Next()
			return
		}
		if limited {
			log.Println("Request is limited")
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}
