# Go-Seckill 高并发秒杀系统

基于 Golang 构建的高并发秒杀系统，专注于解决瞬间洪峰流量下的超卖、数据库击穿以及系统稳定性问题。

## 🛠️ 技术栈

- **语言与框架：** Golang, Gin, GORM
- **数据存储：** MySQL 8.0, Redis 6.0
- **消息队列：** RabbitMQ (AMQP)
- **安全与风控：** JWT,  Bloom Filter

## 🚀 核心架构流转

1. **安全校验网关：** 请求进入系统，由中间件校验 JWT Token 拦截非法用户与机器脚本。
2. **前置拦截防御：** 本地 Bloom Filter 校验商品 ID，若判定为伪造 ID 则直接丢弃，防止缓存穿透。
3. **内存原子扣减：** 请求到达 Redis，通过 Lua 脚本原子执行库存扣减与限购校验，成功则发放“通行证”。
4. **异步削峰填谷：** 获取“通行证”的订单数据序列化后投入 RabbitMQ，直接响应前端。
5. **平滑持久落库：** 后台消费者协程以平稳速率从 RabbitMQ 消费数据，开启 GORM 事务安全写入 MySQL，并手动执行 ACK 确认。

## ✨ 亮点功能

- **动态限购：** 摆脱 MySQL 唯一索引的死板限制，在 Redis Lua 层面实现针对单个用户的灵活限购。
- **毒药消息防范：** 消费者端精确识别业务级死错误（如库存不足、重复数据），主动打断死循环。
- **开箱即用：** 极简的代码结构，严格遵循 Controller-Service-DAO 分层规范。

## 🏁 快速开始

### 1. 基础设施准备 (推荐使用 Docker Compose)
确保已安装 MySQL, Redis, RabbitMQ。

### 2. 运行服务
\`\`\`bash
# 下载依赖
go mod tidy
# 启动
go run main.go
\`\`\`
服务默认运行在 `http://localhost:8080`
用户端运行在 `http://localhost:8080/user`
