package main

import (
	"net/http"
	"strings"
	"time"

	"basic-go/webook/internal/repository"
	"basic-go/webook/internal/repository/cache"
	"basic-go/webook/internal/repository/dao"
	"basic-go/webook/internal/service"
	"basic-go/webook/internal/web"
	"basic-go/webook/internal/web/middleware"
	"basic-go/webook/ioc"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db := initDB()
	server := initWebServer()
	initUserHdl(db, server)
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello，启动成功了！")
	})
	server.Run(":8080")
}

func initUserHdl(db *gorm.DB, server *gin.Engine) {

	cmdable := ioc.InitRedis()
	userCache := cache.NewUserCache(cmdable)
	ud := dao.NewUserDAO(db)

	ur := repository.NewUserRepository(ud, userCache)
	us := service.NewUserService(ur)
	hdl := web.NewUserHandler(us)
	hdl.RegisterRoutes(server)

}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initWebServer() *gin.Engine {
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		//AllowAllOrigins: true,
		//AllowOrigins:     []string{"http://localhost:3000"},
		AllowCredentials: true,

		// ExposeHeaders: []string{"x-jwt-token"},
		AllowHeaders:  []string{"Content-Type", "Authorization"},
		ExposeHeaders: []string{"x-jwt-token"},
		//AllowHeaders: []string{"content-type"},
		//AllowMethods: []string{"POST"},
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				//if strings.Contains(origin, "localhost") {
				return true
			}
			return strings.Contains(origin, "your_company.com")
		},
		MaxAge: 12 * time.Hour,
	}), func(ctx *gin.Context) {
		println("这是我的 Middleware")
	})

	// redisClient := redis.NewClient(&redis.Options{
	// 	Addr: config.Config.Redis.Addr,
	// })

	// server.Use(ratelimit.NewBuilder(redisClient,
	// 	time.Second, 1).Build())

	useJWT(server)

	// useSession(server)
	return server
}

func useJWT(server *gin.Engine) {
	login := middleware.LoginJWTMiddlewareBuilder{}
	server.Use(login.CheckLogin())
}

func useSession(server *gin.Engine) {
	// 这一部分是session
	login := &middleware.LoginMiddlewareBuilder{}
	// 存储数据的，也就是你 userId 存哪里
	// 直接存 cookie
	store := cookie.NewStore([]byte("secret"))
	// 多个服务器可以共享同一个 Redis 实例，具有分布式特性
	// store, err := redis.NewStore(16, "tcp", "localhost:6379", "", []byte("mgyn8y9abnd7q4zkq2m73yw8tu9j5ixm"),
	// 	[]byte("o6jdlg2cb9f9pb6h46fimllw481ldebi"))
	// if err != nil {
	// 	panic(err)
	// }
	server.Use(sessions.Sessions("ssid", store), login.CheckLogin())
}
