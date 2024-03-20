package utils

import "github.com/confluentinc/confluent-kafka-go/kafka"

// NewKafkaProducer 初始化Kafka生产者
func NewKafkaProducer(brokers string) *kafka.Producer {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": brokers})
	if err != nil {
		panic(err)
	}
	return p
}

// SendVoteRequest 发送消息
func SendVoteRequest(producer *kafka.Producer, topic string, message []byte) error {
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          message,
	}

	err := producer.Produce(msg, nil)
	if err != nil {
		return err
	}
	producer.Flush(15 * 1000) // 等待所有消息发送完成
	return nil
}
