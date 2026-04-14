package dao

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

var MQConn *amqp.Connection
var MQChannel *amqp.Channel
var QueueName = "seckill_order_queue"

func InitRabbitMQ() {
	var err error
	// 连接 RabbitMQ (默认账号密码都是 guest，端口 5672)
	MQConn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("❌ RabbitMQ 连接失败: %v", err)
	}

	MQChannel, err = MQConn.Channel()
	if err != nil {
		log.Fatalf("❌ 打开 RabbitMQ Channel 失败: %v", err)
	}

	// 声明一个队列（如果不存在会自动创建）
	_, err = MQChannel.QueueDeclare(
		QueueName,
		true,  // durable: 队列持久化（重启不丢数据）
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		log.Fatalf("❌ 队列声明失败: %v", err)
	}
	fmt.Println("✅ RabbitMQ 连接成功并声明队列!")
}
