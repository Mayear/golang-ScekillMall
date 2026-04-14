package dao

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
)

var Rdb *redis.Client
var Ctx = context.Background()

func InitRedis() {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})

	_, err := Rdb.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("❌ Redis 连接失败: %v", err)
	}
	fmt.Println("✅ Redis 连接成功!")
}

// ================= 👇 新增核心防超卖逻辑 👇 =================
