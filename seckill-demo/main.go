package main

import (
	"fmt"
	"seckill-demo/dao"
	"seckill-demo/model"
	"seckill-demo/router"
	"seckill-demo/service"
)

func main() {
	// 🌟 1. 初始化所有基础设施 (必须放在最前面！)
	dao.InitMySQL()
	dao.InitRedis()
	dao.InitRabbitMQ() // 👈 罪魁祸首就是缺了这行，或者顺序放错了！

	// 自动迁移表结构
	dao.DB.AutoMigrate(&model.User{}, &model.Product{}, &model.Order{})

	// 2. 启动服务与缓存预热
	service.WarmUpCache()

	// 🌟 只有在 InitRabbitMQ 成功后，才能启动消费者！
	go service.StartConsumer()

	// 3. 装载路由
	r := router.SetupRouter()

	// 4. 启动服务器
	fmt.Println("🚀 全栈分布式秒杀系统已就绪: http://localhost:8080")
	r.Run(":8080")
}
