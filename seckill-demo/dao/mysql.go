package dao

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DB 暴露给全局使用
var DB *gorm.DB

// InitMySQL 初始化 MySQL 连接
func InitMySQL() {
	// DSN (Data Source Name): 用户名:密码@tcp(IP:端口)/数据库名?...
	dsn := "root:root@tcp(127.0.0.1:3306)/flash_sale?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ MySQL 连接失败: %v", err)
	}

	DB = db
	fmt.Println("✅ MySQL 连接成功!")
}
