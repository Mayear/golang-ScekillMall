# ⚡ Go-Seckill-Fullstack | 高并发全栈秒杀系统

这是一个基于 **Go (Gin)** 和 **Redis** 构建的工业级高并发秒杀系统。项目涵盖了从前端用户抢购到后端异步入库、安全防护、缓存优化等全链路实战场景。

## 🚀 核心架构设计

项目采用 **"内存预减 + 异步落库"** 的高性能架构：

1. **流量过滤：** 本地 **布隆过滤器** 拦截非法 ID。
2. **安全验证：** **JWT** 中间件校验身份 + **图形验证码** 防机器刷单。
3. **并发扣减：** **Redis Lua 脚本** 实现原子库存扣减。
4. **异步削峰：** **Go Channel** 异步消费订单，平滑写入 **MySQL**。



## 🛠️ 技术栈

- **后端：** Golang, Gin Web Framework, GORM (ORM)
- **缓存/中间件：** Redis (Lua Scripting), Go Channels
- **安全：** JWT (JSON Web Token), Base64 Captcha, Bloom Filter
- **前端：** Vue 3, Element Plus, Axios
- **存储：** MySQL 8.0

## 📂 项目结构

```text
seckill-demo/
├── controller/     # 控制层：处理 HTTP 请求与参数校验
├── service/        # 业务层：核心秒杀逻辑、布隆过滤器、异步队列
├── dao/            # 数据访问层：MySQL 与 Redis 初始化
├── middleware/     # 中间件：JWT 鉴权、Cors 等
├── model/          # 模型层：数据库表结构定义
├── router/         # 路由层：API 路径管理与中间件挂载
├── utils/          # 工具类：JWT 生成、图形验证码、加密等
├── static/         # 前端：B端大屏 (index.html) 与 C端商城 (user.html)
└── main.go         # 入口：程序初始化与服务启动
