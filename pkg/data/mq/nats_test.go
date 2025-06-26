package mq

import (
	"context"
	"fmt"
	"goserver/gen/pb"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func TestPublisherSubscriber(t *testing.T) {
	url := "nats://@127.0.0.1:4222"
	err := InitNats(url)
	if err != nil {
		t.Error(err)
		return
	}

	topic := "activity.test1"
	consumerConfig := jetstream.ConsumerConfig{
		Name:          "betrank-gate-broadcast",
		FilterSubject: topic,
		DeliverPolicy: jetstream.DeliverNewPolicy,    // 只处理consumer创建后的新消息
		AckPolicy:     jetstream.AckNonePolicy,       // 消息确认:不确认
		ReplayPolicy:  jetstream.ReplayInstantPolicy, // 消息回放机制
		MaxDeliver:    1,                             // 消息投递的最大次数-1表示没ack就不停投递
		MaxAckPending: 0,                             // 多少条消息未投递成功则不再投递新消息
	}
	consumer, err := NatsJS.CreateOrUpdateConsumer(context.Background(), StreamActivity, consumerConfig)
	if err != nil {
		t.Error(err)
		return
	}
	var counter int32
	consumer.Consume(func(msg jetstream.Msg) {
		defer msg.Ack()
		data := msg.Data()
		fmt.Printf("%s consume: %d bytes -> %s \n", msg.Subject(), len(data), string(data))

		atomic.AddInt32(&counter, 1)
	})

	total := 10
	for i := 0; i < total; i++ {
		_, err = NatsJS.Publish(context.Background(), topic, []byte(fmt.Sprintf("hello jetstream %d", i)))
		if err != nil {
			t.Error(err)
			return
		}
		// fmt.Printf("publish: %#v\n", ack)
		time.Sleep(time.Millisecond * 300)
	}
	time.Sleep(time.Second)

	if int(counter) != total {
		t.Error("message lose")
	}
}

func TestNatsReqRsp(t *testing.T) {
	url := "nats://@127.0.0.1:4222"
	err := InitNats(url)
	if err != nil {
		t.Error(err)
		return
	}

	subject := "activity.betrank.daily"
	NatsConn.Subscribe(subject, func(m *nats.Msg) {
		fmt.Println("receive:", string(m.Data))
		r := nats.NewMsg(subject)
		r.Header.Add("error", "some error")
		r.Data = []byte("I'm fine")
		m.RespondMsg(r)
	})

	// publish 1 request/response
	for i := 0; i < 3; i++ {
		rsp, err := NatsConn.Request(subject, []byte(fmt.Sprintf("publish request %d", i)), time.Second*5)
		if err != nil {
			t.Error(err)
			return
		}
		if rsp.Header.Get("error") != "some error" {
			t.Error("response header lose")
			return
		}
		fmt.Println("response:", string(rsp.Data))
	}
	// publish 2
	err = NatsConn.Publish(subject, []byte("publish conn data"))
	if err != nil {
		return
	}
	// publish 3
	_, err = NatsJS.Publish(context.Background(), subject, []byte("publish jetstream data"))
	if err != nil {
		return
	}
	time.Sleep(2 * time.Second)
}

func TestJson(t *testing.T) {
	var a = "123456"
	bytes, err := sonic.Marshal(a)
	if err != nil {
		t.Error(err)
		return
	}
	var bb = func() (r *string) {
		var s string
		return &s
	}
	b := bb()
	err = sonic.Unmarshal(bytes, &b)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(*b)
}

func TestAsd(t *testing.T) {
	var rsp *pb.ActivityBetRankRewordRsp
	rspMsg := nats.NewMsg("1231232")
	err := natsMarshal(rsp, rspMsg)
	if err != nil {
		t.Error(err)
	}
}
