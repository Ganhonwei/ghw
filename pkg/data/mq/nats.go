package mq

import (
	"context"
	"errors"
	"fmt"
	"goserver/pkg/glog"
	"runtime/debug"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

var (
	NatsConn *nats.Conn
	NatsJS   jetstream.JetStream
)

func InitNats(natsUrl string) (err error) {
	options := nats.Options{
		Url: natsUrl,
		// 断线重连配置
		RetryOnFailedConnect: true,
		AllowReconnect:       true,
		MaxReconnect:         -1,
		ReconnectWait:        time.Second * 2,
		ReconnectBufSize:     10 * 1024 * 1024, // 10M
	}

	NatsConn, err = options.Connect()
	if err != nil {
		return
	}
	NatsJS, err = jetstream.New(NatsConn)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err = initNatsStream(ctx)
	if err != nil {
		return
	}
	return
}

// initNatsStream 初始化 jetstream 消息流
func initNatsStream(ctx context.Context) (err error) {
	for _, ss := range initialStream {
		if len(ss) < 2 {
			err = fmt.Errorf("initial jetstream %v config error", ss)
			return
		}
		err = NatsCreateStream(ctx, ss[0], ss[1:]...)
		if err != nil {
			return
		}
	}
	return
}

// NatsCreateStream 创建或更新jetstream消息流
func NatsCreateStream(ctx context.Context, streamName string, subjects ...string) (err error) {
	// 活动消息流 https://docs.nats.io/nats-concepts/jetstream/streams#configuration
	streamConfig := jetstream.StreamConfig{
		Name:              streamName,
		Storage:           jetstream.FileStorage,
		Subjects:          subjects,       // 要绑定的topics
		Replicas:          1,              // 消息副本数量
		MaxAge:            time.Hour * 24, // 流中任何消息的最大年龄
		MaxBytes:          -1,             // 存储的所有消息最大字节数 -1无限制
		MaxMsgs:           -1,             // 流中存储的所有消息最大消息数
		MaxMsgSize:        -1,             // 单个消息最大字节数
		MaxConsumers:      -1,             // Stream 定义的最大消费者数量
		NoAck:             false,
		Retention:         jetstream.LimitsPolicy, // 声明流的保留策略
		Discard:           jetstream.DiscardOld,   // stream消息数量达到limit处理策略
		MaxMsgsPerSubject: -1,                     // stream保留每个主题最大消息数
		Duplicates:        time.Minute * 2,        // 跟踪重复消息的窗口
		AllowRollup:       false,                  // 允许使用标题Nats-Rollup将流的所有内容或流中的主题替换为单个新消息
		DenyDelete:        true,                   // 不允许删除消息
		DenyPurge:         true,                   // 不允许清空消息
	}
	_, err = NatsJS.CreateOrUpdateStream(ctx, streamConfig)
	if err != nil {
		return
	}
	return
}

// NatsCreateConsumer 创建或更新 jetstream consumer
// consumerName 多个相同name只消费一次,多个不同name消费分别消费一次
func NatsCreateConsumer[T any](consumerName, streamName, topic string, inst func() T, handler func(T) (err error), options ...NatsConsumerOption) {
	if NatsJS == nil {
		panic(errors.New("nats js not connect"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	consumerName = strings.ReplaceAll(fmt.Sprintf("%s-%s", consumerName, topic), ".", "-")
	// https://docs.nats.io/nats-concepts/jetstream/consumers#general
	consumerConfig := jetstream.ConsumerConfig{
		Durable:       consumerName,
		Name:          consumerName,
		FilterSubject: topic,
		DeliverPolicy: jetstream.DeliverNewPolicy,    // 只处理consumer创建后的新消息
		AckPolicy:     jetstream.AckExplicitPolicy,   // 消息确认:不确认高qps
		ReplayPolicy:  jetstream.ReplayInstantPolicy, // 消息回放机制
		MaxDeliver:    -1,                            // 消息投递的最大次数-1表示没ack就不停投递
		MaxAckPending: -1,                            // 多少条消息未投递成功则不再投递新消息-1不限制
	}
	for _, opt := range options {
		opt(&consumerConfig)
	}
	consumer, err := NatsJS.Consumer(ctx, streamName, consumerName)
	if err != nil {
		consumer, err = NatsJS.CreateConsumer(ctx, streamName, consumerConfig)
		if err != nil {
			// consumer 配置修改
			if err == jetstream.ErrConsumerExists {
				consumer, err = NatsJS.UpdateConsumer(ctx, streamName, consumerConfig)
				if err != nil {
					// NatsJS.DeleteConsumer(ctx, streamName, consumerName)
					panic(err)
				}
			} else {
				panic(err)
			}
		}
	}
	consumer.Consume(func(msg jetstream.Msg) {
		defer func() {
			if r := recover(); r != nil {
				glog.Errorf("nats consumer recover error: consumer=%s, stream=%s, topic=%s, %v", consumerName, streamName, topic, r)
				debug.PrintStack()
			}
		}()

		defer msg.Ack()

		var data T
		if inst != nil {
			data = inst()
			if err := natsJsUnmarshal(data, msg); err != nil {
				glog.Errorf("unmarshal data %T error, %s, %v", data, msg.Subject(), err)
				return
			}
		}
		err = handler(data)
		if err != nil {
			glog.Error(err)
			debug.PrintStack()
		}
	})
}

func NatsCreateConsumerExtend[T any](consumerName, streamName, topic string, inst func() T, handler func(T, *jetstream.MsgMetadata) (ack bool, delay time.Duration, err error), options ...NatsConsumerOption) {
	if NatsJS == nil {
		panic(errors.New("nats js not connect"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	consumerName = strings.ReplaceAll(fmt.Sprintf("%s-%s", consumerName, topic), ".", "-")
	// https://docs.nats.io/nats-concepts/jetstream/consumers#general
	consumerConfig := jetstream.ConsumerConfig{
		Durable:       consumerName,
		Name:          consumerName,
		FilterSubject: topic,
		DeliverPolicy: jetstream.DeliverNewPolicy,    // 只处理consumer创建后的新消息
		AckPolicy:     jetstream.AckExplicitPolicy,   // 消息确认:不确认高qps
		ReplayPolicy:  jetstream.ReplayInstantPolicy, // 消息回放机制
		MaxDeliver:    -1,                            // 消息投递的最大次数-1表示没ack就不停投递
		MaxAckPending: -1,                            // 多少条消息未投递成功则不再投递新消息-1不限制
	}
	for _, opt := range options {
		opt(&consumerConfig)
	}
	consumer, err := NatsJS.Consumer(ctx, streamName, consumerName)
	if err != nil {
		consumer, err = NatsJS.CreateConsumer(ctx, streamName, consumerConfig)
		if err != nil {
			// consumer 配置修改
			if err == jetstream.ErrConsumerExists {
				consumer, err = NatsJS.UpdateConsumer(ctx, streamName, consumerConfig)
				if err != nil {
					// NatsJS.DeleteConsumer(ctx, streamName, consumerName)
					panic(err)
				}
			} else {
				panic(err)
			}
		}
	}
	consumer.Consume(func(msg jetstream.Msg) {
		defer func() {
			if r := recover(); r != nil {
				glog.Errorf("nats consumer recover error: consumer=%s, stream=%s, topic=%s, %v", consumerName, streamName, topic, r)
				debug.PrintStack()
			}
		}()

		var data T
		if inst != nil {
			data = inst()
			if err := natsJsUnmarshal(data, msg); err != nil {
				glog.Errorf("unmarshal data %T error, %s, %v", data, msg.Subject(), err)
				msg.Ack() // 解析失败直接确认，避免无限重试
				return
			}
		}

		// 获取消息元数据
		metadata, err := msg.Metadata()
		if err != nil {
			glog.Errorf("get message metadata error: %v", err)
			msg.Ack() // 获取元数据失败直接确认
			return
		}

		// 调用handler并处理结果
		ack, delay, err := handler(data, metadata)
		if err != nil {
			glog.Error(err)
			debug.PrintStack()
		}

		// 根据返回值决定确认或重试
		if ack {
			msg.Ack()
		} else {
			if err := msg.NakWithDelay(delay); err != nil {
				glog.Errorf("nak with delay error: %v", err)
				msg.Nak() // 如果设置延迟失败，直接Nak
			}
		}
	})
}

// NatsSubscribe nats 请求响应注册
func NatsSubscribe[T, V any](subject string, inst func() T, handler func(req T) (rsp V, err error)) (s *nats.Subscription) {
	if NatsConn == nil {
		panic(errors.New("nats not connect"))
	}
	s, err := NatsConn.Subscribe(subject, func(msg *nats.Msg) {
		defer func() {
			if r := recover(); r != nil {
				glog.Errorf("nats subject recover error: topic=%s, %v", subject, r)
				debug.PrintStack()
			}
		}()
		defer msg.Ack()

		var req T
		if inst != nil {
			req = inst()
			if err := natsUnmarshal(req, msg); err != nil {
				glog.Errorf("unmarshal data %T error, %s", req, msg.Subject)
				return
			}
		}

		rsp, err := handler(req)

		rspMsg := nats.NewMsg(subject)
		if err != nil {
			rspMsg.Header.Add("error", err.Error())
		} else {
			if err := natsMarshal(rsp, rspMsg); err != nil {
				glog.Errorf("marshal rsp data %T error, %s", rsp, msg.Subject)
				return
			}
		}
		msg.RespondMsg(rspMsg)
	})
	if err != nil {
		panic(err)
	}
	return
}

// NatsRequest nats发送请求
func NatsRequest(subject string, rsp, req any, options ...NatsRequestOption) (err error) {
	if NatsConn == nil {
		return errors.New("nats not connect")
	}
	c := &NatsRequestOptions{Timeout: time.Second * 10}
	for _, option := range options {
		option(c)
	}
	reqMsg := nats.NewMsg(subject)
	if err = natsMarshal(req, reqMsg); err != nil {
		return
	}
	rspMsg, err := NatsConn.RequestMsg(reqMsg, c.Timeout)
	if err != nil {
		return
	}
	if e := rspMsg.Header.Get("error"); e != "" {
		err = errors.New(e)
		return
	}
	if err = natsUnmarshal(rsp, rspMsg); err != nil {
		return
	}
	return
}

// NatsPublish nats 发布消息
func NatsPublish(subject string, data any) (err error) {
	if NatsJS == nil {
		return errors.New("nats js not connect")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	msg := nats.NewMsg(subject)
	if err = natsMarshal(data, msg); err != nil {
		return
	}
	_, err = NatsJS.PublishMsg(ctx, msg)
	if err != nil {
		return
	}
	return
}

func natsMarshal(data any, msg *nats.Msg) (err error) {
	mt := "1" // 序列化1json,2protobuf
	if marshaler, ok := data.(proto.Marshaler); ok {
		if bytes, err := marshaler.Marshal(); err == nil {
			mt = "2"
			msg.Data = bytes
		}
	}
	if mt == "1" {
		msg.Data, err = sonic.Marshal(data)
		if err != nil {
			return
		}
	}
	msg.Header.Add("mt", mt)
	return
}

func natsUnmarshal(data any, msg *nats.Msg) (err error) {
	mt := msg.Header.Get("mt")
	if mt == "2" {
		unmarshaler, ok := data.(proto.Unmarshaler)
		if !ok {
			return fmt.Errorf("%T is not protobuf type", data)
		}
		return unmarshaler.Unmarshal(msg.Data)
	}
	// mt=1 json
	err = sonic.Unmarshal(msg.Data, data)
	if err != nil {
		return
	}
	return
}

func natsJsUnmarshal(data any, msg jetstream.Msg) (err error) {
	mt := msg.Headers().Get("mt")
	if mt == "2" {
		unmarshaler, ok := data.(proto.Unmarshaler)
		if !ok {
			return fmt.Errorf("%T is not protobuf type", data)
		}
		return unmarshaler.Unmarshal(msg.Data())
	}
	// mt=1 json
	err = sonic.Unmarshal(msg.Data(), data)
	if err != nil {
		return
	}
	return
}
