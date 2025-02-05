package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	initViperV1()
	initLogger()
	server := InitWebServer()
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello，启动成功了！")
	})
	server.Run(":8080")
}
func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}

func initViperV1() {
	cfile := pflag.String("config",
		"./config/config.yaml", "配置文件路径")
	// 这一步之后，cfile 里面才有值
	pflag.Parse()
	//viper.Set("db.dsn", "localhost:3306")
	// 所有的默认值放好s
	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfile)
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	val := viper.Get("test.key")
	log.Println(val)
}
func initViper() {
	viper.SetConfigName("dev")
	viper.SetConfigType("yaml")
	// 当前工作目录的 config 子目录
	viper.AddConfigPath("config")
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	val := viper.Get("test.key")
	log.Println(val)
}

// func useSession(server *gin.Engine) {
// 	// 这一部分是session
// 	login := &middleware.LoginMiddlewareBuilder{}
// 	// 存储数据的，也就是你 userId 存哪里
// 	// 直接存 cookie
// 	store := cookie.NewStore([]byte("secret"))
// 	// 多个服务器可以共享同一个 Redis 实例，具有分布式特性
// 	// store, err := redis.NewStore(16, "tcp", "localhost:6379", "", []byte("mgyn8y9abnd7q4zkq2m73yw8tu9j5ixm"),
// 	// 	[]byte("o6jdlg2cb9f9pb6h46fimllw481ldebi"))
// 	// if err != nil {
// 	// 	panic(err)
// 	// }
// 	server.Use(sessions.Sessions("ssid", store), login.CheckLogin())
// }
