package model

import "time"

// Product 商品表模型
type Product struct {
	ID    int64  `gorm:"primaryKey;column:id" json:"id"` // 增加 json 标签
	Title string `gorm:"column:title" json:"title"`      // 前端传来的 "title" 会自动填入这里
	Stock int    `gorm:"column:stock" json:"stock"`      // 前端传来的 "stock" 会自动填入这里
}

// 绑定 MySQL 表名
func (Product) TableName() string {
	return "sk_product"
}

// Order 订单表模型
type Order struct {
	ID         int64     `gorm:"primaryKey;column:id" json:"id"`
	UserID     int64     `gorm:"column:user_id" json:"user_id"`
	ProductID  int64     `gorm:"column:product_id" json:"product_id"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"create_time"`
}

func (Order) TableName() string {
	return "sk_order"
}

// User 用户表模型
type User struct {
	ID       int    `gorm:"primaryKey;column:id;autoIncrement" json:"id"`
	Username string `gorm:"column:username;unique" json:"username"`
	Password string `gorm:"column:password" json:"password"`
}

func (User) TableName() string {
	return "sk_user"
}
