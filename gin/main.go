package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()
	server.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello, World!")
	})
	server.POST("/post", func(c *gin.Context) {
		c.String(200, "POSTfangfa")
	})
	server.GET("/user/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.String(200, "参数路由!\n")
		c.String(200, name)
	})
	server.GET("/view/*html", func(c *gin.Context) {
		name := c.Param(".html")
		c.String(200, "通配符路由!\n")
		c.String(200, name)
	})
	server.GET("/order", func(ctx *gin.Context) {
		id := ctx.Query("id")
		ctx.String(http.StatusOK, "订单 ID 是 "+id)
	})
	server.Run(":8080")
}
