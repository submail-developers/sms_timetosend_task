package kafka_consumer

import (
	"sms_timetosend_task/conf"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	logs "github.com/astaxie/beego/logs"
)

var logger *logs.BeeLogger

func init() {
}

type KAKFAPartitionConsumer struct {
	Host     string
	Topic    string
	GroupID  string
	Messages chan []byte
}

func NewConsumer(topic, group string) *KAKFAPartitionConsumer {
	kfk := &KAKFAPartitionConsumer{}
	kfk.Host = conf.KAFKA_HOST
	kfk.Topic = topic
	kfk.GroupID = group
	kfk.Messages = make(chan []byte)
	go kfk.Subscribe()
	return kfk
}

func (k *KAKFAPartitionConsumer) Subscribe() {
	config := sarama.NewConfig()
	config.Consumer.Offsets.AutoCommit.Enable = true                     // 开启自动 commit offset
	config.Consumer.Offsets.AutoCommit.Interval = 100 * time.Millisecond // 自动 commit时间间隔
	var wg sync.WaitGroup
	client, err := sarama.NewClient([]string{k.Host}, config)
	if err != nil {
		logger.Error("NewClient err: %s", err.Error())
	}
	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		logger.Error("Failed to start consumer: %s", err.Error())
		return
	}
	offsetManager, _ := sarama.NewOffsetManagerFromClient(k.GroupID, client)
	partitionList, err := consumer.Partitions(k.Topic) //获得该topic所有的分区
	if err != nil {
		logger.Error("Failed to get the list of partition err:%s ", err.Error())
		return
	}

	for partition := range partitionList {
		partitionOffsetManager, _ := offsetManager.ManagePartition(k.Topic, int32(partition))
		nextOffset, _ := partitionOffsetManager.NextOffset()
		pc, err := consumer.ConsumePartition(k.Topic, int32(partition), nextOffset)
		if err != nil {
			logger.Error("Failed to start consumer for partition %d: %s", partition, err)
			return
		}
		wg.Add(1)
		go func(sarama.PartitionConsumer) { //为每个分区开一个go协程去取值
			for msg := range pc.Messages() { //阻塞直到有值发送过来，然后再继续等待
				logger.Debug("Partition:%d, Offset:%d, key:%s, value:%s", msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
				k.Messages <- msg.Value
				partitionOffsetManager.MarkOffset(msg.Offset+1, "modified metadata")
			}
			defer pc.AsyncClose()
			wg.Done()
		}(pc)
	}
	wg.Wait()
}
