package samarax

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"basic-go/webook/pkg/logger"

	"github.com/IBM/sarama"
)

type BatchHandler[T any] struct {
	fn func(msgs []*sarama.ConsumerMessage, ts []T) error
	l  logger.LoggerV1
}

func NewBatchHandler[T any](l logger.LoggerV1, fn func(msgs []*sarama.ConsumerMessage, ts []T) error) *BatchHandler[T] {
	return &BatchHandler[T]{fn: fn, l: l}
}

func (b *BatchHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 是 BatchHandler 结构体的方法，用于消费 Kafka 消息并将其分批处理。
// 该方法会从 Kafka 主题中消费消息，并将消息分批处理，每批最多包含 batchSize 条消息。
// 如果在指定的超时时间内无法凑够一批消息，则会提前处理当前批次。
//
// 参数:
//   - session: sarama.ConsumerGroupSession 对象，用于管理消费者组的会话状态。
//   - claim: sarama.ConsumerGroupClaim 对象，表示消费者组对某个分区的消费权。
//
// 返回值:
//   - error: 如果消费过程中发生错误，则返回相应的错误信息；否则返回 nil。
func (b *BatchHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	const batchSize = 10
	for {
		log.Println("一个批次开始")
		batch := make([]*sarama.ConsumerMessage, 0, batchSize)
		ts := make([]T, 0, batchSize)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
		var done = false
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				// 超时了
				done = true
			case msg, ok := <-msgs:
				fmt.Println("msg.Value: ", msg.Value)
				if !ok {
					cancel()
					return nil
				}
				fmt.Println("batch_handler: ", msg)
				batch = append(batch, msg)
				var t T
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					b.l.Error("反序列消息体失败",
						logger.String("topic", msg.Topic),
						logger.Int32("partition", msg.Partition),
						logger.Int64("offset", msg.Offset),
						logger.Error(err))
					continue
				}
				batch = append(batch, msg)
				ts = append(ts, t)
			}
		}
		cancel()
		// 凑够了一批，然后你就处理
		err := b.fn(batch, ts)
		if err != nil {
			b.l.Error("处理消息失败",
				// 把真个 msgs 都记录下来
				logger.Error(err))
		}
		for _, msg := range batch {
			session.MarkMessage(msg, "")
		}
	}
}
