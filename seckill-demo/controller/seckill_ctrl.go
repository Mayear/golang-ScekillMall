// 秒杀控制器：处理抢购的 HTTP 请求
package controller

import (
	"fmt"
	"seckill-demo/dao"
	"seckill-demo/model"
	"seckill-demo/service" // 这里调用你的 service 层

	"github.com/gin-gonic/gin"
)

func HandleSeckill(c *gin.Context) {

	type Req struct {
		ProductID int `json:"product_id"`
		// UserID 字段被删掉了！前端不允许传，直接从 Token 里取！
	}
	var req Req
	c.ShouldBindJSON(&req)

	// 🌟 从中间件(Context)中安全地获取 user_id
	userID := c.GetInt("user_id")

	res := service.SeckillExecute(req.ProductID, userID)

	if res == 1 {
		c.JSON(200, gin.H{"code": 200, "msg": "🎉 抢到啦！正在排队入库..."})
	} else if res == 0 {
		c.JSON(200, gin.H{"code": 400, "msg": "😭 被抢光了"})
	} else if res == -2 {
		c.JSON(200, gin.H{"code": 400, "msg": "您已达到该商品的限购上限！把机会留给别人吧~"})
	} else {
		c.JSON(200, gin.H{"code": 500, "msg": "系统繁忙"})
	}
}

func warmUpCache() {
	var products []model.Product
	dao.DB.Find(&products)
	for _, p := range products {
		key := fmt.Sprintf("seckill:stock:%d", p.ID)
		dao.Rdb.Set(dao.Ctx, key, p.Stock, 0)
	}
	fmt.Printf("📦 已从数据库预热 %d 个商品到 Redis\n", len(products))
}
