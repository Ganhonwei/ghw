package gate

import (
	"context"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/algo"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"strconv"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (a *GateActor) SyncConfig(ctx actor.Context) {
	//同步配置
	arg := ctx.Message().(*pb.SyncConfig)
	// glog.Debugf("SyncConfig %#v", arg)
	handler.SyncConfig(arg, a.Name)
	a.pushNotice(arg)
}

func (a *GateActor) PayCurrency(ctx actor.Context) {
	//后台或充值同步到game房间
	arg := ctx.Message().(*pb.PayCurrency)
	glog.Debugf("PayCurrency %#v", arg)
	userid := arg.Userid
	if v, ok := a.online[userid]; ok {
		v.Pid.Tell(arg)
	} else if v, ok := a.offline[userid]; ok {
		v.Pid.Tell(arg)
	} else {
		//离线
		a.rolePid.Tell(arg)
	}
}

func (a *GateActor) ModifyCurrency(ctx actor.Context) {
	//后台修改货币数量同步到game房间
	arg := ctx.Message().(*pb.ModifyCurrency)
	glog.Debugf("ModifyCurrency %#v", arg)
	userid := arg.Userid
	if v, ok := a.online[userid]; ok {
		v.Pid.Tell(arg)
	} else if v, ok := a.offline[userid]; ok {
		v.Pid.Tell(arg)
	} else {
		//离线
		a.rolePid.Tell(arg)
	}
}

func (a *GateActor) ChangeCurrency(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ChangeCurrency)
	userid := arg.Userid
	if v, ok := a.online[userid]; ok {
		glog.Infof("ChangeCurrency %#v", arg)
		v.Pid.Tell(msg)
	} else if v, ok := a.offline[userid]; ok {
		glog.Infof("ChangeCurrency %#v", arg)
		v.Pid.Tell(msg)
	} else {
		glog.Infof("ChangeCurrency %#v", arg)
		//离线
		a.rolePid.Tell(msg)
	}
}

func (a *GateActor) WxpayCallback(ctx actor.Context) {
	// msg := ctx.Message()
	// arg := msg.(*pb.WxpayCallback)
	// glog.Debugf("WxpayCallback %#v", arg)
	// if !handler.WxpayVerify(arg) {
	// 	return
	// }
	// a.rolePid.Tell(arg)
}

func (a *GateActor) PayGoods(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.PayGoods)
	glog.Debugf("PayGoods: %v", arg)
	userid := arg.Userid
	if v, ok := a.online[userid]; ok {
		v.Pid.Tell(arg)
	} else if v, ok := a.offline[userid]; ok {
		v.Pid.Tell(msg)
	} else {
		glog.Errorf("PayGoods: %v", arg)
	}
}

func (a *GateActor) MarQueeNtf(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.MarQueeNtf)
	glog.Debugf("MarQueeNtf: %v", arg)
	userid := arg.Userid
	if v, ok := a.online[userid]; ok {
		v.Pid.Tell(arg)
	}
}

// 在线玩家信息
// func (a *GateActor) OnlineUser(ctx actor.Context) {
// 	msg := ctx.Message()
// 	datas := make([]data.OnlineUser, 0)
// 	for _, p := range a.online {
// 		d, err := p.RequestFuture(msg, 3*time.Second).Result()
// 		if err != nil {
// 			continue
// 		}
// 		if u, ok := d.(data.OnlineUser); ok && u.Id != "" {
// 			datas = append(datas, u)
// 		}
// 	}
// 	res := new(pb.OnlinedUser)
// 	// b, err := jsoniter.Marshal(datas)
// 	if err == nil {
// 		// res.Data = b
// 		ctx.Respond(res)
// 	}
// }

func (a *GateActor) GamePeopleNtf(ctx actor.Context) {
	arg := ctx.Message().(*pb.GamePeopleNtf)
	a.broadcast(arg)
}

// 节点广播消息
func (a *GateActor) broadcast(msg interface{}) {
	for _, v := range a.online {
		v.Pid.Tell(msg)
	}
}

// 节点广播消息推送
func (a *GateActor) pushNotice(arg *pb.SyncConfig) {
	switch arg.Type {
	case pb.CONFIG_NOTICE: //公告
		b := make(map[string]data.Notice)
		err = json.Unmarshal(arg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, arg.Data)
			return
		}
		glog.Debugf("pushNotice %#v", b)
		for _, v := range b {
			switch arg.Atype {
			case pb.CONFIG_DELETE:
			case pb.CONFIG_UPSERT:
				msg := new(pb.PushNoticeNtf)
				msg.Info = &pb.Notice{
					Time: utils.Time2LocalStr(v.Ctime),
					// Rtype:   int32(v.Rtype),
					// Acttype: int32(v.Acttype),
					Content: v.Content,
				}
				// if v.Userid == "" {
				// 	//广播消息通知玩家,只发送新消息
				// 	a.broadcast(msg)
				// } else {
				// 	s := utils.Split(v.Userid, ",")
				// 	for _, val := range s {
				// 		//玩家个人消息单独通知
				// 		if val, ok := a.online[val]; ok {
				// 			val.Tell(msg)
				// 		}
				// 	}
				// }
			}
		}
	case pb.CONFIG_CUSTOMERADDR: //客服地址
		b := make(map[string]data.ModifyCustomer)
		err = json.Unmarshal(arg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, arg.Data)
			return
		}
		glog.Debugf("pushCustomAddr %#v", b)
		ntf := new(pb.CustomerAddressNtf)
		for _, c := range b {
			ntf.Mail = c.Mail
			ntf.Telegram = c.Telegram
			ntf.WhatsApp = c.WhatsApp
			break
		}
		for _, p := range a.online {
			p.Pid.Tell(ntf)
		}
	case pb.CONFIG_SHARE_WAY: //分享配置
		b := make(map[string]data.ModifyShare)
		err = json.Unmarshal(arg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, arg.Data)
			return
		}
		glog.Debugf("pushShareWay %#v", b)
		ntf := new(pb.ShareWayNtf)
		for _, c := range b {
			ntf.Youtobe = c.Youtobe
			ntf.Ins = c.Ins
			ntf.Facebook = c.Facebook
			ntf.Telegram = c.Telegram
			ntf.CashPrize = c.CashPrize
			ntf.Whatsapp = c.WhatsApp
			ntf.X = c.X
			break
		}
		for _, p := range a.online {
			p.Pid.Tell(ntf)
		}
	}

}

func (a *GateActor) resetScratch() {
	if node != "1" {
		return
	}
	now := time.Now().In(location)
	newVersion := utils.Time2DayDate(now)

	// var oldtime int32
	str, err := client.Get(context.Background(), data.SCRATCHTICKET_VERSION).Result()
	if err != nil || str == "" {
		glog.Errorf("resetScratch: %v", err)
		// 重新设置
		// client.Set(context.Background(), data.SCRATCHTICKET_VERSION, newVersion, 0).Result()
		// return
	}
	oldVersion, _ := strconv.Atoi(str)

	if oldVersion == newVersion {
		return
	}

	// old := time.Unix(oldtime, 0).In(location)
	// now := time.Now().In(location)
	// if utils.Time2DayDate(old) == utils.Time2DayDate(now) {
	// 	return
	// }

	// 重新随机中奖牌
	client.Del(context.Background(), data.SCRATCHTICKET_POKER)
	cards := algo.ShuffleCards()
	pokers := cards[:5] // 5张奖牌
	for _, v := range pokers {
		client.LPush(context.Background(), data.SCRATCHTICKET_POKER, v)
	}
	// 生成版本号
	client.Set(context.Background(), data.SCRATCHTICKET_VERSION, newVersion, 0).Result()
	scratchTick.Poker = pokers
	scratchTick.Version = int32(newVersion)
	glog.Infof("Scratchticket pokers: %v", pokers)
}

// 打码排行榜更新
func (a *GateActor) PublishActivityBetRankUpdate(ctx actor.Context) {
	arg := ctx.Message().(*pb.PublishActivityBetRankUpdate)
	// 在线用户分批更新
	batch := 100
	var groups []map[string]*actor.PID
	group := make(map[string]*actor.PID, batch)
	for userid, role := range a.online {
		group[userid] = role.Pid
		if len(group) >= batch {
			groups = append(groups, group)
		}
	}
	if len(group) > 0 {
		groups = append(groups, group)
	}

	go func() {
		for _, group := range groups {
			var userids []string
			for userid := range group {
				userids = append(userids, userid)
			}
			rsp := &pb.ActivityBetRankUpdateNtfs{}
			req := &mq.RequestActivityBetRankUpdateArgs{RankType: arg.RankType, Userids: userids}
			if err := mq.NatsRequest(mq.RequestActivityBetRankUpdate, rsp, req); err != nil {
				glog.Errorf("request bet rank update error: %v", err)
			} else {
				for userid, ntf := range rsp.Ntfs {
					if rsPid, ok := group[userid]; ok {
						rsPid.Tell(ntf)
					}
				}
			}
		}
	}()
}

// 支付渠道更新
func (a *GateActor) PublishPayChannelUpdate(ctx actor.Context) {
	// 更新一下支付渠道
	getPayChannelShops(true)

	arg := ctx.Message().(*pb.PublishPayChannelUpdate)
	for _, role := range a.online {
		role.Pid.Tell(arg)
	}
}

// 客服回复消息
func (a *GateActor) ConsumerReplyMessage(ctx actor.Context) {
	arg := ctx.Message().(*pb.ConsumerReplyMessage)
	if role, ok := a.online[arg.Userid]; ok {
		role.Pid.Tell(arg)
	} else {
		a.rolePid.Tell(arg)
	}
}

// 客服结束对话
func (a *GateActor) PublishCustomerSessionOver(ctx actor.Context) {
	arg := ctx.Message().(*pb.PublishCustomerSessionOver)
	if role, ok := a.online[arg.Userid]; ok {
		role.Pid.Tell(arg)
	} else {
		a.rolePid.Tell(arg)
	}
}

// 客服撤回消息
func (a *GateActor) PublishCustomerMessageRevoke(ctx actor.Context) {
	arg := ctx.Message().(*pb.PublishCustomerMessageRevoke)
	if role, ok := a.online[arg.Userid]; ok {
		role.Pid.Tell(arg)
	}
}
