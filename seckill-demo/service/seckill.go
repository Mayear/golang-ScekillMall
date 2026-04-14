package service

import (
	"fmt"
	"strconv"

	"seckill-demo/dao"
	"seckill-demo/model"

	"github.com/bits-and-blooms/bloom/v3"
	"gorm.io/gorm"
)

// ---------------- 1. 异步队列与布隆过滤器定义 ----------------
type OrderMsg struct {
	UserID    int
	ProductID int
}

var OrderQueue = make(chan OrderMsg, 1024)

// 🌟 新增：全局布隆过滤器实例
var ProductBloomFilter *bloom.BloomFilter

// ---------------- 2. 缓存预热逻辑 ----------------
func WarmUpCache() {
	var products []model.Product
	dao.DB.Find(&products)

	// 🌟 初始化布隆过滤器
	// 参数1：预计存放的数据量 (比如 10000 个商品)
	// 参数2：允许的误判率 (比如 0.01 表示 1% 的容错率)
	ProductBloomFilter = bloom.NewWithEstimates(10000, 0.01)

	for _, p := range products {
		// 1. 同步到 Redis
		key := fmt.Sprintf("seckill:stock:%d", p.ID)
		dao.Rdb.Set(dao.Ctx, key, p.Stock, 0)

		// 2. 🌟 将商品 ID 写入布隆过滤器 (需要转成 []byte)
		idStr := strconv.Itoa(int(p.ID))
		ProductBloomFilter.AddString(idStr)
	}
	fmt.Printf("📦 业务层启动: 已从数据库预热 %d 个商品到 Redis，并成功构建布隆过滤器\n", len(products))
}

// ---------------- 3. 核心秒杀执行逻辑 ----------------
func SeckillExecute(productID int, userID int) int {
	// 🌟 拦截器：布隆过滤器查岗
	// 如果布隆过滤器说这个商品 ID 不存在，那它绝对不存在！直接拦截！
	idStr := strconv.Itoa(productID)
	if !ProductBloomFilter.TestString(idStr) {
		fmt.Printf("🛡️ 布隆过滤器拦截非法请求: 非法商品 ID [%d]\n", productID)
		return -1 // 直接返回异常，不给 Redis 增加任何压力
	}

	// === 通过了布隆过滤器的验证，说明商品极大概率存在，去问 Redis ===
	luaScript := `
		if (redis.call('exists', KEYS[1]) == 1) then
			local stock = tonumber(redis.call('get', KEYS[1]))
			if (stock > 0) then
				redis.call('decr', KEYS[1])
				return 1
			end
			return 0
		end
		return -1
	`

	key := fmt.Sprintf("seckill:stock:%d", productID)

	// 执行 Lua 脚本
	result, err := dao.Rdb.Eval(dao.Ctx, luaScript, []string{key}).Result()
	if err != nil {
		fmt.Printf("❌ 秒杀逻辑执行异常: %v\n", err)
		return -1
	}

	res := int(result.(int64))

	// 如果 Redis 返回 1 (抢购成功)，就把数据丢进通道排队
	if res == 1 {
		OrderQueue <- OrderMsg{UserID: userID, ProductID: productID}
	}

	return res
}

// ---------------- 4. 异步落库消费者 ----------------
func StartConsumer() {
	for msg := range OrderQueue {
		// 开启 GORM 数据库事务
		err := dao.DB.Transaction(func(tx *gorm.DB) error {
			res := tx.Model(&model.Product{}).Where("id = ? AND stock > 0", msg.ProductID).Update("stock", gorm.Expr("stock - 1"))
			if res.RowsAffected == 0 {
				return fmt.Errorf("库存不足")
			}
			return tx.Create(&model.Order{UserID: int64(msg.UserID), ProductID: int64(msg.ProductID)}).Error
		})

		if err == nil {
			fmt.Printf("✅ 订单入库: 用户 [%d] 成功购买商品 [%d]\n", msg.UserID, msg.ProductID)
		} else {
			fmt.Printf("❌ 落库失败: 用户 [%d], 原因: %v\n", msg.UserID, err)
		}
	}
}
