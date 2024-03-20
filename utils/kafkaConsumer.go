package utils

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func StartKafkaConsumer(brokers, groupID, topic string) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}

	c.SubscribeTopics([]string{topic}, nil)

	for {
		msg, err := c.ReadMessage(-1)
		if err == nil {
			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
			// 在这里处理你的逻辑
		} else {
			// 错误处理
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}
}
