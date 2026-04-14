package main

import (
	"fmt"
	"seckill-demo/dao"
	"seckill-demo/model"
	"seckill-demo/router"
	"seckill-demo/service"
)

func main() {
	// 1. 初始化基础设施
	dao.InitMySQL()
	dao.InitRedis()
	dao.DB.AutoMigrate(&model.User{}, &model.Product{}, &model.Order{})

	// 2. 启动服务与缓存预热
	service.WarmUpCache()
	go service.StartConsumer() // 消费者丢去 service 层启动

	// 3. 装载路由
	r := router.SetupRouter()

	// 4. 启动服务器
	fmt.Println("🚀 全栈秒杀系统已就绪: http://localhost:8080")
	r.Run(":8080")
}
