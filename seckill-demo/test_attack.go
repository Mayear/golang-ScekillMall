package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	Host           = "http://localhost:8080"
	ConcurrentNums = 200 // 模拟 200 个真实用户同时抢购
	TargetProduct  = 1   // 你要压测的商品 ID (请确保数据库里有这个商品，且库存设置个 50 左右)
)

var (
	successCount int32 // 记录抢购成功数
	failCount    int32 // 记录抢购失败数
	errorCount   int32 // 记录网络报错数
)

func main() {
	fmt.Printf("🔫 准备生成 %d 个虚拟并发用户...\n", ConcurrentNums)
	tokens := prepareUsers(ConcurrentNums)

	if len(tokens) == 0 {
		fmt.Println("❌ 用户准备失败，压测终止。")
		return
	}

	fmt.Printf("✅ 成功获取 %d 个 JWT Token，准备发动并发攻击！\n", len(tokens))
	time.Sleep(2 * time.Second) // 给人一点心理准备时间

	// 🌟 核心技巧：发令枪机制
	// 使用一个 channel 阻塞所有协程，然后 close 它，让所有协程在同一毫秒同时开火！
	startGate := make(chan struct{})
	var wg sync.WaitGroup

	fmt.Println("🔥 3... 2... 1... 攻击开始！")
	startTime := time.Now()

	for i := 0; i < len(tokens); i++ {
		wg.Add(1)
		go func(token string) {
			defer wg.Done()

			<-startGate // 阻塞在这里，等待发令枪响

			doSeckill(token, TargetProduct)
		}(tokens[i])
	}

	// 嘭！关闭通道，所有阻塞的协程瞬间同时执行
	close(startGate)

	// 等待所有子弹飞完
	wg.Wait()

	costTime := time.Since(startTime)

	fmt.Println("\n================ 📊 压测结果报告 ================")
	fmt.Printf("⏱️ 总耗时: %v\n", costTime)
	fmt.Printf("🎯 QPS (每秒请求数): %.2f\n", float64(ConcurrentNums)/costTime.Seconds())
	fmt.Printf("✅ 抢购成功数: %d\n", successCount)
	fmt.Printf("😭 抢购失败数 (手慢无): %d\n", failCount)
	fmt.Printf("❌ 网络错误数: %d\n", errorCount)
	fmt.Println("=================================================")
}

// 发起秒杀请求
func doSeckill(token string, productID int) {
	reqBody, _ := json.Marshal(map[string]int{"product_id": productID})
	req, _ := http.NewRequest("POST", Host+"/api/seckill", bytes.NewBuffer(reqBody))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token) // 携带专属身份

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)

	if err != nil {
		atomic.AddInt32(&errorCount, 1)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 简单解析返回的 JSON，判断业务 code
	var res map[string]interface{}
	json.Unmarshal(body, &res)

	// 根据你的 controller 逻辑，code 200 是成功，400 是失败
	if int(res["code"].(float64)) == 200 {
		atomic.AddInt32(&successCount, 1)
	} else {
		atomic.AddInt32(&failCount, 1)
	}
}

// 自动注册并登录，返回 Token 列表
func prepareUsers(num int) []string {
	var tokens []string
	for i := 0; i < num; i++ {
		username := fmt.Sprintf("testuser_%d_%d", time.Now().Unix(), i)
		password := "123456"

		// 1. 注册
		regBody, _ := json.Marshal(map[string]string{"username": username, "password": password})
		http.Post(Host+"/api/register", "application/json", bytes.NewBuffer(regBody))

		// 2. 登录拿 Token
		resp, err := http.Post(Host+"/api/login", "application/json", bytes.NewBuffer(regBody))
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var res map[string]interface{}
		json.Unmarshal(body, &res)

		if token, ok := res["token"].(string); ok {
			tokens = append(tokens, token)
		}
	}
	return tokens
}
