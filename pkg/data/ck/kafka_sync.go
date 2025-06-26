package ck

import (
	"context"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/zlog"
	"runtime/debug"

	"github.com/bytedance/sonic"
	"github.com/segmentio/kafka-go"
	"gopkg.in/ini.v1"
)

var (
	kafkaEnable    bool
	topicProducers = make(map[string]*kafka.Writer, 16)
	topicConsumers = make(map[string]*kafka.Reader, 16)
)

var (
	topics = []string{
		// data.TopicUser,
		// data.TopicUserFinance,
		// data.TopicTradeRecord,
		// data.TopicWithdrawRecord,
		// data.TopicWithdrawRecord_transfer_order_detail,
		// data.TopicLogWater,
		// data.TopicDetail,
		// data.TopicLogLogin,
		// data.TopicExternalBet,
		// data.TopicExternalReward,
		// data.TopicExternalCancel,
		// data.TopicActivityBetRankPrize,
		// data.TopicActivityTurnDrawLog,
		// data.TopicActivityTurnPrizeLog,
		// data.TopicLaunchInviteLog,
		// data.TopicShareAgentIncomeRecord,
		// data.TopicShareAgentIncomeRecordTackLog,
	}
)

func InitKafkaProducer(cfg *ini.File) {
	kafkaEnable = cfg.Section("kafka").Key("enable").MustBool(false)
	if !kafkaEnable {
		glog.Infof("kafka disabled")
		zlog.Infof("kafka disabled")
		return
	} else {
		glog.Infof("kafka enabled")
		zlog.Infof("kafka enabled")
	}
	addr := cfg.Section("kafka").Key("addr").Value()

	for _, topic := range topics {
		w := &kafka.Writer{
			Addr:                   kafka.TCP(addr),
			Topic:                  topic,            // Writer中的Topic和Message中的Topic是互斥的，同一时刻有且只能设置一处。
			AllowAutoTopicCreation: true,             // 自动创建topic
			Balancer:               &kafka.Hash{},    // 用hash对msg key路由
			RequiredAcks:           kafka.RequireAll, // ack模式
			Async:                  true,             // 异步
		}
		// stats := w.Stats()
		// glog.Infof("topic %s writer status: writes=%d, messages=%d, errors=%d, bytes=%d", topic, stats.Writes, stats.Messages, stats.Errors, stats.Bytes)
		topicProducers[topic] = w
	}

	data.DataProducers = &KafkaDataProducer{}
}

func InitKafkaConsumer(cfg *ini.File) {
	kafkaEnable = cfg.Section("kafka").Key("enable").MustBool(false)
	if !kafkaEnable {
		glog.Infof("kafka disabled")
		zlog.Infof("kafka disabled")
		return
	} else {
		glog.Infof("kafka enabled")
		zlog.Infof("kafka enabled")
	}
	addr := cfg.Section("kafka").Key("addr").Value()

	for _, topic := range topics {
		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers:   []string{addr},
			Topic:     topic,
			GroupID:   "consumer-group-1",
			Partition: 0,
			MinBytes:  10e3, // 10KB
			MaxBytes:  10e6, // 10MB
		})
		topicConsumers[topic] = r
	}

	for t, r := range topicConsumers {
		go func(topic string, reader *kafka.Reader) {
			defer func() {
				if r := recover(); r != nil {
					glog.Error("Consumer error:", r)
					zlog.Error("Consumer error:", r)
					debug.PrintStack()
				}
			}()

			glog.Infof("kafka start consumer topic %s", topic)
			zlog.Infof("kafka start consumer topic %s", topic)
			for {
				msg, err := reader.FetchMessage(context.Background())
				// m, err := r.ReadMessage(context.Background()) // read 自动提交
				if err != nil {
					glog.Errorf("message fetch error: topic %s, %v", topic, err)
					zlog.Errorf("message fetch error: topic %s, %v", topic, err)
					break
				}

				err = handleConsumerMessage(topic, [][]byte{msg.Value})
				if err != nil {
					// 不commit
					glog.Errorf("message consumer error: topic %s, %v", topic, err)
					zlog.Errorf("message consumer error: topic %s, %v", topic, err)
					continue
				}

				// fetch 后不 comment 会重复推送
				if err := reader.CommitMessages(context.Background(), msg); err != nil {
					glog.Errorf("failed to commit messages:  topic %s, offset %d, key %s, %v", topic, msg.Offset, string(msg.Key), err)
					zlog.Errorf("failed to commit messages:  topic %s, offset %d, key %s, %v", topic, msg.Offset, string(msg.Key), err)
				}
				// fmt.Printf("message at offset %d: %s = %s\n", msg.Offset, string(msg.Key), string(msg.Value))
			}
		}(t, r)
	}
}

func handleConsumerMessage(topic string, msgs [][]byte) (err error) {
	// var messages []any
	// for _, msg := range msgs {
	// 	var message any
	// 	switch topic {
	// 	case data.TopicUser:
	// 		message = new(data.User)
	// 	case data.TopicUserFinance:
	// 		message = new(data.UserFinance)
	// 	case data.TopicTradeRecord:
	// 		message = new(data.TradeRecord)
	// 	case data.TopicWithdrawRecord:
	// 		message = new(data.WithdrawRecord)
	// 	case data.TopicWithdrawRecord_transfer_order_detail:
	// 		message = new(data.TransferOrderDetail)
	// 	case data.TopicLogWater:
	// 		message = new(data.LogWater)
	// 	case data.TopicDetail:
	// 		message = new(data.Detail)
	// 	case data.TopicLogLogin:
	// 		message = new(data.LogLogin)
	// 	case data.TopicExternalBet:
	// 		message = new(data.NsqLogExternalBet)
	// 	case data.TopicExternalReward:
	// 		message = new(data.NsqLogExternalReward)
	// 	case data.TopicExternalCancel:
	// 		message = new(data.NsqLogExternalCancel)
	// 	case data.TopicActivityBetRankPrize:
	// 		message = new(data.ActivityBetRankPrize)
	// 	case data.TopicActivityTurnDrawLog:
	// 		message = new(data.ActivityTurnDrawLog)
	// 	case data.TopicActivityTurnPrizeLog:
	// 		message = new(data.ActivityTurnPrizeLog)
	// 	case data.TopicLaunchInviteLog:
	// 		message = new(data.LaunchInviteLog)
	// 	case data.TopicShareAgentIncomeRecord:
	// 		message = new(data.ShareAgentIncomeRecord)
	// 	case data.TopicShareAgentIncomeRecordTackLog:
	// 		message = new(data.ShareAgentIncomeRecordTackLog)
	// 	default:
	// 		glog.Errorf("unknown topic: %s", topic)
	// 		continue
	// 	}
	// 	err = sonic.Unmarshal(msg, message)
	// 	if err != nil {
	// 		glog.Errorf("unmarshal topic %s msg error: %v", topic, err)
	// 		continue
	// 	}
	// 	messages = append(messages, message)
	// }
	// if len(messages) == 0 {
	// 	return
	// }
	// err = ParseAndInsert(messages)
	// if err != nil {
	// 	return
	// }
	return
}

type KafkaDataProducer struct{}

func (KafkaDataProducer) PublishDatas(topic string, datas []interface{}) {
	defer func() {
		if r := recover(); r != nil {
			glog.Error("kafka PublishDatas error:", r)
			zlog.Error("kafka PublishDatas error:", r)
			debug.PrintStack()
		}
	}()

	if !kafkaEnable {
		return
	}
	if len(datas) == 0 {
		return
	}
	writer, ok := topicProducers[topic]
	if !ok {
		glog.Errorf("topic producer not exists: %s", topic)
		zlog.Errorf("topic producer not exists: %s", topic)
		return
	}

	var msgs []kafka.Message
	for _, data := range datas {
		msg, err := sonic.Marshal(data)
		if err != nil {
			glog.Errorf("publish data marshal error: %s, %v", err, topic, err)
			zlog.Errorf("publish data marshal error: %s, %v", err, topic, err)
			return
		}
		msgs = append(msgs, kafka.Message{Key: []byte(topic), Value: msg})
	}

	err := writer.WriteMessages(context.Background(), msgs...)
	if err != nil {
		glog.Errorf("topic producer write message error: %s, %v", topic, err)
		zlog.Errorf("topic producer write message error: %s, %v", topic, err)
		return
	}
}
