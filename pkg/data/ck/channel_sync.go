package ck

import (
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/zlog"
	"runtime/debug"

	"github.com/bytedance/sonic"
)

type ChannelSync struct {
	topicChannels map[string]chan []byte
}

func InitChannelProducerConsumer() {
	c := &ChannelSync{topicChannels: make(map[string]chan []byte, 16)}
	for _, topic := range topics {
		c.topicChannels[topic] = make(chan []byte, 4096)
	}
	c.initConsumer()
	data.DataProducers = c
}

func (c *ChannelSync) PublishDatas(topic string, datas []interface{}) {
	defer func() {
		if r := recover(); r != nil {
			glog.Error("PublishDatas error:", r)
			zlog.Error("PublishDatas error:", r)
			debug.PrintStack()
		}
	}()
	if len(datas) == 0 {
		return
	}
	if c.topicChannels == nil {
		return
	}
	channel, ok := c.topicChannels[topic]
	if !ok {
		glog.Errorf("topic channel not exists: %s", topic)
		zlog.Errorf("topic channel not exists: %s", topic)
		return
	}

	var msgs [][]byte
	for _, data := range datas {
		msg, err := sonic.Marshal(data)
		if err != nil {
			glog.Error("publish data marshal error: %s, %v", err, topic, err)
			zlog.Error("publish data marshal error: %s, %v", err, topic, err)
			return
		}
		msgs = append(msgs, msg)
	}

	for _, msg := range msgs {
		select {
		case channel <- msg:
		default:
			glog.Errorf("topic channel full: %s", topic)
			zlog.Errorf("topic channel full: %s", topic)
		}
	}
}

func (c *ChannelSync) initConsumer() {
	for topic, ch := range c.topicChannels {
		go func(topic string, channel chan []byte) {
			zlog.Infof("channel start consumer topic %s", topic)
			for {
				var msgs [][]byte
				msgs = append(msgs, <-channel)

			waiting:
				for i := 0; i < 1000; i++ {
					select {
					case msg := <-channel:
						msgs = append(msgs, msg)
					default:
						break waiting
					}
				}
				err := handleConsumerMessage(topic, msgs)
				if err != nil {
					glog.Errorf("message consumer error: topic %s, %v", topic, err)
					zlog.Errorf("message consumer error: topic %s, %v", topic, err)
				}
			}
		}(topic, ch)
	}
}
