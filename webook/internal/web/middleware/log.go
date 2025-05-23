// package middleware

// import (
// 	"bytes"
// 	"context"
// 	"io"
// 	"time"

// 	"github.com/gin-gonic/gin"
// )

// type LogMiddlewareBuilder struct {
// 	logFn         func(ctx context.Context, l AccessLog)
// 	allowReqBody  bool
// 	allowRespBody bool
// }

// func NewLogMiddlewareBuilder(logFn func(ctx context.Context, l AccessLog)) *LogMiddlewareBuilder {
// 	return &LogMiddlewareBuilder{
// 		logFn: logFn,
// 	}
// }

// func (l *LogMiddlewareBuilder) AllowReqBody() *LogMiddlewareBuilder {
// 	l.allowReqBody = true
// 	return l
// }

// func (l *LogMiddlewareBuilder) AllowRespBody() *LogMiddlewareBuilder {
// 	l.allowRespBody = true
// 	return l
// }

// func (l *LogMiddlewareBuilder) Build() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		path := ctx.Request.URL.Path
// 		if len(path) > 1024 {
// 			path = path[:1024]
// 		}
// 		method := ctx.Request.Method
// 		al := AccessLog{
// 			Path:   path,
// 			Method: method,
// 		}
// 		if l.allowReqBody {
// 			// Request.Body 是一个 Stream 对象，只能读一次
// 			body, _ := ctx.GetRawData()
// 			if len(body) > 2048 {
// 				al.ReqBody = string(body[:2048])
// 			} else {
// 				al.ReqBody = string(body)
// 			}
// 			// 放回去
// 			ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
// 			//ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
// 		}

// 		start := time.Now()

// 		if l.allowRespBody {
// 			ctx.Writer = &responseWriter{
// 				ResponseWriter: ctx.Writer,
// 				al:             &al,
// 			}
// 		}

// 		defer func() {
// 			al.Duration = time.Since(start)
// 			//duration := time.Now().Sub(start)
// 			l.logFn(ctx, al)
// 		}()

// 		// 直接执行下一个 middleware...直到业务逻辑
// 		ctx.Next()
// 		// 在这里，你就拿到了响应
// 	}
// }

// type AccessLog struct {
// 	Path     string        `json:"path"`
// 	Method   string        `json:"method"`
// 	ReqBody  string        `json:"req_body"`
// 	Status   int           `json:"status"`
// 	RespBody string        `json:"resp_body"`
// 	Duration time.Duration `json:"duration"`
// }

// type responseWriter struct {
// 	gin.ResponseWriter
// 	al *AccessLog
// }

// func (w *responseWriter) Write(data []byte) (int, error) {
// 	w.al.RespBody = string(data)
// 	return w.ResponseWriter.Write(data)
// }

//	func (w *responseWriter) WriteHeader(statusCode int) {
//		w.al.Status = statusCode
//		w.ResponseWriter.WriteHeader(statusCode)
//	}
package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

type LogMiddlewareBuilder struct {
	logFn         func(ctx context.Context, l AccessLog)
	allowReqBody  bool
	allowRespBody bool
}

func NewLogMiddlewareBuilder(logFn func(ctx context.Context, l AccessLog)) *LogMiddlewareBuilder {
	return &LogMiddlewareBuilder{
		logFn: logFn,
	}
}

func (l *LogMiddlewareBuilder) AllowReqBody() *LogMiddlewareBuilder {
	l.allowReqBody = true
	return l
}

func (l *LogMiddlewareBuilder) AllowRespBody() *LogMiddlewareBuilder {
	l.allowRespBody = true
	return l
}

func (l *LogMiddlewareBuilder) maskSensitiveData(data map[string]interface{}) {
	for key, _ := range data {
		switch key {
		case "Email":
			data[key] = "******"
			// data[key] = strings.ReplaceAll(value.(string), "@", "[at]")
		case "Password":
			data[key] = "******"
		}
	}
}

func (l *LogMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if len(path) > 1024 {
			path = path[:1024]
		}
		method := ctx.Request.Method
		al := AccessLog{
			Path:   path,
			Method: method,
		}
		if l.allowReqBody {
			// Request.Body 是一个 Stream 对象，只能读一次
			body, _ := io.ReadAll(ctx.Request.Body)
			if len(body) > 2048 {
				al.ReqBody = string(body[:2048])
			} else {
				al.ReqBody = string(body)
			}
			// 放回去
			ctx.Request.Body = io.NopCloser(bytes.NewReader(body))

			// 脱敏处理请求体
			var reqData map[string]interface{}
			if err := json.Unmarshal([]byte(al.ReqBody), &reqData); err == nil {
				l.maskSensitiveData(reqData)
				if marshalled, err := json.Marshal(reqData); err == nil {
					al.ReqBody = string(marshalled)
				}
			}
		}

		start := time.Now()

		if l.allowRespBody {
			ctx.Writer = &responseWriter{
				ResponseWriter: ctx.Writer,
				al:             &al,
			}
		}

		defer func() {
			// 脱敏处理响应体
			if l.allowRespBody {
				var respData map[string]interface{}
				if err := json.Unmarshal([]byte(al.RespBody), &respData); err == nil {
					l.maskSensitiveData(respData)
					if marshalled, err := json.Marshal(respData); err == nil {
						al.RespBody = string(marshalled)
					}
				}
			}

			al.Duration = time.Since(start)
			l.logFn(ctx, al)
		}()

		// 直接执行下一个 middleware...直到业务逻辑
		ctx.Next()
	}
}

type AccessLog struct {
	Path     string        `json:"path"`
	Method   string        `json:"method"`
	ReqBody  string        `json:"req_body"`
	Status   int           `json:"status"`
	RespBody string        `json:"resp_body"`
	Duration time.Duration `json:"duration"`
}

type responseWriter struct {
	gin.ResponseWriter
	al *AccessLog
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.al.RespBody = string(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.al.Status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
