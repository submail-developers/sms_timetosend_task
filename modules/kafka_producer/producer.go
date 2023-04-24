package kafka_producer

import (
	"sms_timetosend_task/conf"
	"sms_timetosend_task/log"
	"time"

	"github.com/Shopify/sarama"
)

var _producer *Produce

func init() {
	_producer = NewProducer()
}

type Produce struct {
	AsyncProducer sarama.AsyncProducer
}

func NewProducer() *Produce {
	config := sarama.NewConfig()
	config.Version = sarama.V3_3_1_0
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Errors = true                     // 接收producer的error，下面细说用法
	config.Producer.Compression = sarama.CompressionLZ4      // 压缩方式，如果kafka版本大于1.2，推荐使用zstd压缩
	config.Producer.Flush.Messages = 10                      // 缓存条数
	config.Producer.Flush.Frequency = 500 * time.Millisecond // 缓存时间
	producer, err := sarama.NewAsyncProducer([]string{conf.KAFKA_HOST}, config)
	if err != nil {
		log.Logger.Error("producer_test create producer error :%s\n", err.Error())
	}
	pd := &Produce{AsyncProducer: producer}
	// 接受结束信号
	go func() {
		for {
			select {
			case suc := <-producer.Successes():
				log.Logger.Info("offset: %d,  timestamp: %s", suc.Offset, suc.Timestamp.String())
			case err := <-producer.Errors():
				log.Logger.Error("kafka producer send error %s", err.Err.Error())
			}
		}
	}()
	return pd
}

func (p *Produce) Send(topic, message string) {
	p.AsyncProducer.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}
}

func GetProducer() *Produce {
	return _producer
}
