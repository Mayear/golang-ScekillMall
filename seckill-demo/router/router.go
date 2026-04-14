package router

import (
	"seckill-demo/controller"
	"seckill-demo/middleware" // 🌟 记得引入我们写的保安中间件

	"github.com/gin-gonic/gin"
)

// SetupRouter 初始化并装载所有的路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 1. 配置静态资源（前端页面）
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")    // B端管理大屏
	r.StaticFile("/user", "./static/user.html") // C端消费者商城

	// 2. 配置所有的 API 分组
	api := r.Group("/api")
	{
		// 🌟 用户认证接口 (公开)
		api.POST("/register", controller.Register)
		api.POST("/login", controller.Login)

		// 🌟 商品管理的 CRUD 接口 (公开，暂时不加锁方便你在管理端操作)
		api.GET("/products", controller.GetProducts)
		api.POST("/products", controller.CreateProduct)
		api.PUT("/products/:id", controller.UpdateProduct)
		api.DELETE("/products/:id", controller.DeleteProduct)

		// 🌟 核心秒杀接口 (加了 JWTAuth 锁！必须携带 Token 才能访问)
		api.POST("/seckill", middleware.JWTAuth(), controller.HandleSeckill)
	}

	return r
}
