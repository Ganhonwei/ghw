package service

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/glog"

	"github.com/nsqio/go-nsq"
	"gopkg.in/ini.v1"
)

var (
	producer *nsq.Producer
)

func InitNsq(cfg *ini.File) {
	InitNsqProducer(cfg)
}

func InitNsqProducer(cfg *ini.File) {
	var err error
	nsqAddr := cfg.Section("nsq").Key("addr").Value()
	producer, err = nsq.NewProducer(nsqAddr, nsq.NewConfig())
	if err != nil {
		panic(err)
	}
}

// 提现成功排行榜事件
func publishRankWithdraw(order *data.PayNotify) {
	glog.Debugf("publishRankWithdraw: %#v", order)
	// order, err := GetWithDrawOrder(orderId)
	// if err != nil {
	// 	glog.Error("publish rank withdraw get order fail: ", err)
	// 	return
	// }
	record, err := GetWithDrawRecord(order.MerOrderID)
	if err != nil {
		glog.Error("publish rank withdraw get order fail: ", err)
		return
	}
	player, err := GetPlayerUser(record.Userid)
	if err != nil {
		glog.Error("publish rank withdraw get player fail: ", err)
		return
	}
	event := &pb.EventRankWithdraw{
		Robot:    false,
		Userid:   player.Userid,
		Username: player.Nickname,
		Avatar:   player.Photo,
		Amount:   int32(record.Score),
		VipLv:    int32(player.Vip.Lv),
	}

	body, _ := event.Marshal()
	producer.Publish(data.TopicRankWithdraw, body)
}
