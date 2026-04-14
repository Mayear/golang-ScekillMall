package controller

import (
	"seckill-demo/dao"
	"seckill-demo/model"
	"seckill-demo/utils"

	"github.com/gin-gonic/gin"
)

// Register 用户注册
func Register(c *gin.Context) {
	var u model.User
	c.ShouldBindJSON(&u)
	if u.Username == "" || u.Password == "" {
		c.JSON(200, gin.H{"code": 400, "msg": "用户名或密码不能为空"})
		return
	}

	var count int64
	dao.DB.Model(&model.User{}).Where("username = ?", u.Username).Count(&count)
	if count > 0 {
		c.JSON(200, gin.H{"code": 400, "msg": "用户名已被注册"})
		return
	}

	dao.DB.Create(&u)
	c.JSON(200, gin.H{"code": 200, "msg": "注册成功，请登录！"})
}

// Login 用户登录
func Login(c *gin.Context) {
	var input model.User
	c.ShouldBindJSON(&input)

	var user model.User
	dao.DB.Where("username = ? AND password = ?", input.Username, input.Password).First(&user)

	if user.ID == 0 {
		c.JSON(200, gin.H{"code": 400, "msg": "账号或密码错误"})
		return
	}

	// 🌟 生成 JWT Token
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(200, gin.H{"code": 500, "msg": "Token 生成失败"})
		return
	}

	c.JSON(200, gin.H{
		"code":     200,
		"msg":      "登录成功",
		"user_id":  user.ID,
		"username": user.Username,
		"token":    token, // 把 Token 颁发给前端
	})
}
