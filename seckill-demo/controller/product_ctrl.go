// 商品控制器：处理商品增删改查的 HTTP 请求
package controller

import (
	"fmt"
	"seckill-demo/dao"
	"seckill-demo/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "5"))
	if page < 1 {
		page = 1
	}

	var list []model.Product
	var total int64
	dao.DB.Model(&model.Product{}).Count(&total)
	offset := (page - 1) * size
	dao.DB.Offset(offset).Limit(size).Find(&list)

	c.JSON(200, gin.H{"list": list, "total": total})
}

func CreateProduct(c *gin.Context) {
	var p model.Product
	c.ShouldBindJSON(&p)
	dao.DB.Create(&p)
	dao.Rdb.Set(dao.Ctx, fmt.Sprintf("seckill:stock:%d", p.ID), p.Stock, 0)
	c.JSON(200, gin.H{"code": 200, "msg": "创建成功"})
}

func UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var p model.Product
	dao.DB.First(&p, id)
	c.ShouldBindJSON(&p)
	dao.DB.Save(&p)
	dao.Rdb.Set(dao.Ctx, fmt.Sprintf("seckill:stock:%d", p.ID), p.Stock, 0)
	c.JSON(200, gin.H{"code": 200, "msg": "更新成功"})
}

func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	dao.DB.Delete(&model.Product{}, id)
	dao.Rdb.Del(dao.Ctx, fmt.Sprintf("seckill:stock:%s", id))
	c.JSON(200, gin.H{"code": 200, "msg": "删除成功"})
}
