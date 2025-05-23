package sarama

import (
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
)

var addr = []string{"localhost:9094"}

func TestSyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(addr, cfg)
	cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	//cfg.Producer.Partitioner = sarama.NewRandomPartitioner
	//cfg.Producer.Partitioner = sarama.NewHashPartitioner
	//cfg.Producer.Partitioner = sarama.NewManualPartitioner
	//cfg.Producer.Partitioner = sarama.NewConsistentCRCHashPartitioner
	//cfg.Producer.Partitioner = sarama.NewCustomPartitioner()
	// kafka-console-consumer --bootstrap-server=localhost:9094 --topic=test_topic --group=my_unique_group --from-beginning
	assert.NoError(t, err)
	for i := 0; i < 100; i++ {
		_, _, err = producer.SendMessage(&sarama.ProducerMessage{
			Topic: "test_topic",
			Value: sarama.StringEncoder("这是一条消息"),
			// 会在生产者和消费者之间传递的
			Headers: []sarama.RecordHeader{
				{
					Key:   []byte("key1"),
					Value: []byte("value1"),
				},
			},
			Metadata: "这是 metadata",
		})
	}
}

func TestAsyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true
	producer, err := sarama.NewAsyncProducer(addr, cfg)
	assert.NoError(t, err)
	msgs := producer.Input()
	msgs <- &sarama.ProducerMessage{
		Topic: "test_topic",
		Value: sarama.StringEncoder("这是一条消息"),
		// 会在生产者和消费者之间传递的
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("key1"),
				Value: []byte("value1"),
			},
		},
		Metadata: "这是 metadata",
	}

	select {
	case msg := <-producer.Successes():
		t.Log("发送成功", string(msg.Value.(sarama.StringEncoder)))
	case err := <-producer.Errors():
		t.Log("发送失败", err.Err, err.Msg)
	}
}
