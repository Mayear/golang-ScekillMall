package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"seckill-demo/dao"
	"seckill-demo/model"

	"github.com/bits-and-blooms/bloom/v3"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

// ---------------- 1. 结构体与布隆过滤器定义 ----------------

type OrderMsg struct {
	UserID    int `json:"user_id"`
	ProductID int `json:"product_id"`
}

// ProductBloomFilter 用于在内存层面拦截非法请求
var ProductBloomFilter *bloom.BloomFilter

// ---------------- 2. 缓存预热逻辑 ----------------

func WarmUpCache() {
	var products []model.Product
	dao.DB.Find(&products)

	// 初始化布隆过滤器
	ProductBloomFilter = bloom.NewWithEstimates(10000, 0.01)

	for _, p := range products {
		// 1. 库存同步到 Redis
		key := fmt.Sprintf("seckill:stock:%d", p.ID)
		dao.Rdb.Set(dao.Ctx, key, p.Stock, 0)

		// 2. 将合法商品 ID 写入布隆过滤器
		idStr := strconv.Itoa(int(p.ID))
		ProductBloomFilter.AddString(idStr)
	}
	fmt.Printf("📦 业务层启动: 已预热 %d 个商品，并构建布隆过滤器与 Redis 缓存\n", len(products))
}

// ---------------- 3. 核心秒杀执行逻辑 (生产者) ----------------

func SeckillExecute(productID int, userID int) int {
	// 🌟 防线 1：布隆过滤器过滤 (保持不变)
	idStr := strconv.Itoa(productID)
	if !ProductBloomFilter.TestString(idStr) {
		fmt.Printf("🛡️ 布隆过滤器拦截: 非法商品 ID [%d]\n", productID)
		return -1
	}

	// 🌟 防线 2：Redis Lua 原子扣减 + 个人限购校验 (全面升级)
	luaScript := `
		local stockKey = KEYS[1]
		local userBoughtKey = KEYS[2]
		local limitAmount = tonumber(ARGV[1])

		-- 1. 判断商品是否存在
		if (redis.call('exists', stockKey) == 0) then 
			return -1 
		end

		-- 2. 判断商品库存是否充足
		local stock = tonumber(redis.call('get', stockKey))
		if (stock <= 0) then 
			return 0 
		end

		-- 3. 判断该用户是否已达限购上限
		-- 使用 get 获取已购数量，如果没有买过，会返回 false，我们给它个默认值 0
		local bought = tonumber(redis.call('get', userBoughtKey) or '0')
		if (bought >= limitAmount) then 
			return -2 -- 🌟 自定义状态码：-2 表示触发限购
		end

		-- 4. 执行扣库存和记录用户已购数量
		redis.call('decr', stockKey)
		redis.call('incr', userBoughtKey)
		
		-- 5. 可以给用户的已购记录设个过期时间 (比如 24 小时后解除限购)，防止占用太多内存
		redis.call('expire', userBoughtKey, 86400)

		return 1 -- 抢购成功！
	`

	// 组装 Keys 和 Args
	stockKey := fmt.Sprintf("seckill:stock:%d", productID)
	userBoughtKey := fmt.Sprintf("seckill:bought:%d:user:%d", productID, userID) // 🌟 新增：用户个人的购买计数器
	limitAmount := 2                                                             // 🌟 业务规则：每人限购 2 件 (未来可以做成动态配置读入)

	// 执行 Lua 脚本 (注意传入了 2 个 Key 和 1 个 Arg)
	result, err := dao.Rdb.Eval(dao.Ctx, luaScript, []string{stockKey, userBoughtKey}, limitAmount).Result()
	if err != nil {
		fmt.Printf("❌ Redis 异常: %v\n", err)
		return -1
	}

	res := int(result.(int64))

	// 🌟 第三步：抢购成功，发送消息到 RabbitMQ
	if res == 1 {
		msg := OrderMsg{UserID: userID, ProductID: productID}
		body, _ := json.Marshal(msg)

		err := dao.MQChannel.PublishWithContext(context.Background(),
			"", dao.QueueName, false, false,
			amqp.Publishing{
				ContentType:  "application/json",
				DeliveryMode: amqp.Persistent,
				Body:         body,
			})
		if err != nil {
			fmt.Printf("❌ MQ 发送失败: %v\n", err)
			// 注意：在真实的商用系统中，如果这里 MQ 连不上，需要调用 redis.call('incr', stockKey) 把库存退回去
		}
	} else if res == -2 {
		fmt.Printf("⚠️ 触发限购: 用户 [%d] 已达到商品 [%d] 的限购上限 [%d] 件\n", userID, productID, limitAmount)
	}

	return res
}

// ---------------- 4. 异步落库消费者 (消费者) ----------------

func StartConsumer() {
	// 注册消费者
	msgs, err := dao.MQChannel.Consume(
		dao.QueueName, // 监听队列
		"",            // consumer
		false,         // 🌟 手动确认消息 (auto-ack 设为 false)
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	if err != nil {
		fmt.Printf("❌ 启动消费者失败: %v\n", err)
		return
	}

	fmt.Println("🚀 RabbitMQ 消费者已就绪，正在监听订单消息...")

	for d := range msgs {
		var msg OrderMsg
		_ = json.Unmarshal(d.Body, &msg)

		// 执行 MySQL 事务落库
		err := dao.DB.Transaction(func(tx *gorm.DB) error {
			// 1. 更新库存
			res := tx.Model(&model.Product{}).Where("id = ? AND stock > 0", msg.ProductID).Update("stock", gorm.Expr("stock - 1"))
			if res.RowsAffected == 0 {
				return fmt.Errorf("库存不足")
			}
			// 2. 创建订单 (唯一索引保证幂等性)
			return tx.Create(&model.Order{UserID: int64(msg.UserID), ProductID: int64(msg.ProductID)}).Error
		})

		if err == nil {
			fmt.Printf("✅ 订单入库: 用户 [%d] 商品 [%d]\n", msg.UserID, msg.ProductID)
			// 🌟 核心：手动给 MQ 回复 ACK，MQ 才会真正删除这条消息
			d.Ack(false)
		} else {
			fmt.Printf("❌ 入库失败: %v，尝试重新入队\n", err)
			// 如果是网络抖动导致的失败，让消息重回队列 (requeue = true)
			d.Nack(false, true)
		}
	}
}
