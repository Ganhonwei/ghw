package gate

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/ck"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/algo"
	"goserver/pkg/game/config"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"goserver/pkg/table"
	"goserver/pkg/utils"

	"slices"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"golang.org/x/sync/singleflight"
	"gopkg.in/mgo.v2/bson"
)

func (rs *RoleActor) PingReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.PingReq)
	// glog.Debugf("PingReq %s", rs.Userid)
	rsp := handler.Ping(arg)
	err = mq.NatsPublish(mq.TopicOnlineUser, &pb.OnlineUserTTL{
		Userid: rs.Userid,
		GameId: utils.String(rs.gtype),
		RoomId: rs.roomId,
	})
	if err != nil {
		glog.Error(err)
	}
	rs.Send(rsp)
}

func (rs *RoleActor) GetGameListReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.GetGameListReq)
	glog.Debugf("GetGameListReq %#v", arg)
	rs.roomPid.Request(arg, ctx.Self())
}

// func (rs *RoleActor) NoticeReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.NoticeReq)
// 	glog.Debugf("NoticeReq %#v", arg)
// 	arg.Userid = rs.User.GetUserid()
// 	rs.dbmsPid.Request(arg, ctx.Self())
// }

// func (rs *RoleActor) NoticeRsp(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.NoticeRsp)
// 	glog.Debugf("NoticeRsp %#v", arg)
// 	handler.PackNotice(arg)
// 	rs.Send(arg)
// }

// func (rs *RoleActor) ActivityReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.ActivityReq)
// 	glog.Debugf("ActivityReq %#v", arg)
// 	arg.Userid = rs.User.GetUserid()
// 	rs.dbmsPid.Request(arg, ctx.Self())
// }

// func (rs *RoleActor) ActivityRsp(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.ActivityRsp)
// 	glog.Debugf("ActivityRsp %#v", arg)
// 	handler.PackActivity(arg)
// 	//glog.Debugf("ActivityRsp %#v, userid %s", arg, rs.User.GetUserid())
// 	rs.Send(arg)
// }

// func (rs *RoleActor) JoinActivityReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.JoinActivityReq)
// 	glog.Debugf("JoinActivityReq %#v", arg)
// 	rs.joinActivity(arg, ctx)
// }

func (rs *RoleActor) GetCurrencyReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.GetCurrencyReq)
	glog.Debugf("GetCurrencyReq %#v", arg)
	//响应
	rsp := handler.GetCurrency(arg, rs.User)
	rs.Send(rsp)
}

// func (rs *RoleActor) BuyReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.BuyReq)
// 	glog.Debugf("BuyReq %#v", arg)
// 	// //优化
// 	// rsp, diamond, coin := handler.Buy(arg, rs.User)
// 	// //同步兑换
// 	// rs.addCurrency(diamond, coin, 0, 0, 0, int32(pb.LOG_TYPE18), "商城购买", "")
// 	// //响应
// 	// rs.Send(rsp)
// 	// record, msg2 := handler.BuyNotice(coin, rs.User.GetUserid())
// 	// if record != nil {
// 	// 	myactor.Logger().Tell(record)
// 	// }
// 	// if msg2 != nil {
// 	// 	rs.Send(msg2)
// 	// }
// }

var (
	// 支付渠道信息
	payChannelShopCache           *data.PayChannelShopResponse
	payChannelShopCacheSec        int64
	payChannelShopCacheRefreshSec int64 = 30 // 缓存刷新时间
	payChannelShopCacheRefreshSF  singleflight.Group
)

// 获取支付渠道信息
func getPayChannelShops(force bool) []data.PayChannelShop {
	nowSec := time.Now().Unix()
	if force || payChannelShopCache == nil || payChannelShopCacheSec+payChannelShopCacheRefreshSec < nowSec {
		payChannelShopCacheRefreshSF.Do("payChannelShop", func() (interface{}, error) {
			// 支付渠道信息查询 http
			response := new(data.PayChannelShopResponse)
			// request := new(data.PayChannelShopRequest)
			// body, _ := sonic.Marshal(request)
			body := []byte("{}")
			if err := handler.SubmitPayNode(payurl+"/channelshop", body, response); err != nil {
				glog.Error("getPayChannelShops error: ", err)
				return nil, err
			}
			payChannelShopCache = response
			payChannelShopCacheSec = nowSec
			return response, nil
		})
	}

	if payChannelShopCache == nil {
		return nil
	}
	return payChannelShopCache.Channels
}

func (rs *RoleActor) PublishPayChannelUpdate(ctx actor.Context) {
	// 商城充值
	shopRsp := packShopRsp(rs.User)
	rs.Send(&pb.ShopNtf{Data: shopRsp})

	// 商城提现
	withdrawRsp := rs.resWithDrawData()
	rs.Send(&pb.GetWithDrawDataNtf{Data: withdrawRsp})
}

func (rs *RoleActor) ShopReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ShopReq)
	glog.Debugf("ShopReq %#v", arg)
	//响应
	rsp := packShopRsp(rs.User)
	//优惠券
	rs.Send(rsp)
}

func packShopRsp(user *data.User) (rsp *pb.ShopRsp) {
	rsp = handler.Shop(user)
	if len(rsp.List) == 0 {
		return
	}

	// 过滤支付渠道支持的商品
	shops, payMin, payMax := handleMatchPayShops(rsp.List,
		func(s *pb.Shop) int64 { return int64(s.Price) },
		func(s *pb.Shop, options []int32) { s.PaymentOptions = options },
		func(s *pb.Shop, apps []int32) { s.PaymentApps = apps })

	// 查询均单价
	if len(shops) > 0 {
		var payAvg int64
		if user.PayAvg != 0 {
			payAvg = user.PayAvg
		} else {
			r := make(map[string]any)
			err := ck.Select(&r, `
				SELECT AVG(amount) amount_avg FROM game.col_trade_record FINAL WHERE userid = ? AND order_status = 4
			`, user.Userid)
			if err != nil {
				glog.Error(err)
			} else {
				payAvg = utils.ToInt64(r["amount_avg"])
			}

			user.PayAvg = payAvg
			if payAvg == 0 {
				user.PayAvg = -1
			}
		}

		if payAvg <= 0 {
			rsp.ShopDefault = shops[0].Price
		} else {
			for _, shop := range shops {
				if shop.Price >= uint32(payAvg) {
					rsp.ShopDefault = shop.Price
					break
				}
			}
		}
	}

	rsp.List = shops
	rsp.MinRecharge = uint32(payMin)
	rsp.MaxRecharge = uint32(payMax)
	return
}

func handleMatchPayShops[T any](shops []T,
	priceGetter func(T) int64,
	payOptionsSetter func(T, []int32),
	payAppsSetter func(T, []int32)) (matchShops []T, payMin, payMax int64) {

	// get payChannels
	payChannels := getPayChannelShops(false)

	var payOptions, payApps = make(map[int32]bool), make(map[int32]bool)
	for _, shop := range shops {
		clear(payOptions)
		clear(payApps)
		// 找匹配的渠道和支付方式
		var matchChannel bool
		for _, c := range payChannels {
			if !c.Payable {
				continue
			}
			if payMin == 0 || payMin > c.PayMin {
				payMin = c.PayMin
			}
			if payMax == 0 || payMax < c.PayMax {
				payMax = c.PayMax
			}
			price := priceGetter(shop)
			// price := int64(shop.Price)
			if price >= c.PayMin && price <= c.PayMax {
				matchChannel = true
				if c.UtrRequired { // 这一项的优先级高于支持的支付方式中配置的内容
					payOptions[3] = true
				} else {
					for _, opt := range c.PayOptions {
						payOptions[opt] = true
					}
				}
				for _, app := range c.PayApps {
					payApps[app] = true
				}
			}
		}
		if !matchChannel {
			continue
		}
		var options, apps []int32
		for opt := range payOptions {
			options = append(options, opt)
		}
		for app := range payApps {
			apps = append(apps, app)
		}
		slices.Sort(options)
		slices.Sort(apps)

		payOptionsSetter(shop, options)
		payAppsSetter(shop, apps)
		matchShops = append(matchShops, shop)
	}
	return
}

func (rs *RoleActor) BankGive(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.BankGive)
	glog.Debugf("BankGive %#v", arg)
	//rs.addBank(arg.Coin, arg.Type, arg.From)
	// rs.addCurrency(0, arg.GetCoin(), 0, 0, 0, arg.GetType(), "银行赠送", "")
	// if rs.gamePid != nil {
	// 	rs.gamePid.Tell(arg)
	// }
}

// func (rs *RoleActor) BankReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.BankReq)
// 	glog.Debugf("BankReq %#v", arg)
// 	// rs.bank(arg)
// }

// func (rs *RoleActor) RankReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.RankReq)
// 	glog.Debugf("RankReq %#v", arg)
// 	rs.dbmsPid.Request(arg, ctx.Self())
// }

// func (rs *RoleActor) BankLogReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.BankLogReq)
// 	glog.Debugf("BankLogReq %#v", arg)
// 	arg.Userid = rs.User.GetUserid()
// 	rs.dbmsPid.Request(arg, ctx.Self())
// }

func (rs *RoleActor) TaskUpdate(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.TaskUpdate)
	glog.Debugf("TaskUpdate %#v", arg)
	rs.taskUpdate(arg)
}

// func (rs *RoleActor) TaskReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.TaskReq)
// 	glog.Debugf("TaskReq %#v", arg)
// 	rs.task()
// }

func (rs *RoleActor) LuckyUpdate(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.LuckyUpdate)
	glog.Debugf("LuckyUpdate %#v", arg)
	rs.luckyUpdate(arg)
}

// func (rs *RoleActor) LuckyReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.LuckyReq)
// 	glog.Debugf("LuckyReq %#v", arg)
// 	rs.lucky()
// }

// func (rs *RoleActor) TaskPrizeReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.TaskPrizeReq)
// 	glog.Debugf("TaskPrizeReq %#v", arg)
// 	rs.taskPrize(arg.Type)
// }

// func (rs *RoleActor) LoginPrizeReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.LoginPrizeReq)
// 	glog.Debugf("LoginPrizeReq %#v", arg)
// 	rs.loginPrize(arg)
// }

func (rs *RoleActor) SignatureReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.SignatureReq)
	glog.Debugf("SignatureReq %#v", arg)
	rs.setSign(arg)
}

func (rs *RoleActor) PhotoReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.PhotoReq)
	glog.Debugf("PhotoReq %#v", arg)
	rs.setPhoto(arg)
}

func (rs *RoleActor) NickNameReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.NickNameReq)
	glog.Debugf("NickNameReq %#v", arg)
	rs.setNickName(arg)
}

// func (rs *RoleActor) LatLngReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.LatLngReq)
// 	glog.Debugf("LatLngReq %#v", arg)
// 	rs.setLatLng(arg)
// }

// func (rs *RoleActor) RoomRecordReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.RoomRecordReq)
// 	glog.Debugf("RoomRecordReq %#v", arg)
// 	msg1 := &pb.GetRoomRecord{
// 		Gtype:  arg.Gtype,
// 		Page:   arg.Page,
// 		Userid: rs.User.GetUserid(),
// 	}
// 	rs.dbmsPid.Request(msg1, ctx.Self())
// }

func (rs *RoleActor) UserDataReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.UserDataReq)
	glog.Debugf("UserDataReq %#v", arg)
	userid := arg.GetUserid()
	if userid == "" {
		userid = rs.User.GetUserid()
	}
	if userid != rs.User.GetUserid() {
		msg1 := new(pb.GetUserData)
		msg1.Userid = userid
		rs.rolePid.Request(msg1, ctx.Self())
	} else {
		if rs.Token == "" {
			// 生成一个token
			token, _ := handler.Sign(rs.Userid)
			rs.Token = token
			rs.status = true
		}
		// 头像处理
		photo, _ := strconv.Atoi(rs.Photo)
		if rs.Vip.Lv < 4 && photo > 30 {
			rs.Photo = utils.String(utils.RandInt32N(30) + 1)
		} else if rs.Vip.Lv >= 4 && rs.Vip.Lv < 8 && photo > 42 {
			rs.Photo = utils.String(utils.RandInt32N(42) + 1)
		} else if rs.Vip.Lv >= 8 && photo > 52 {
			rs.Photo = utils.String(utils.RandInt32N(52) + 1)
		}
		//TODO 添加房间数据返回
		rsp := handler.GetUserDataMsg(arg, rs.User)
		if rs.gamePid != nil {
			rsp.Game = true
		}
		if rs.gameId != "" && rs.roomId != "" {
			rsp.Gameid = rs.gameId
			rsp.Roomid = rs.roomId
			game := config.GetGame(rs.gameId)
			rsp.GameType = rs.gtype
			if game.Id != "" {
				rsp.GameType = game.Gtype
				rsp.RoomType = int32(game.RoomType)
			} else {
				switch rs.gtype {
				case int32(pb.REDBLACK):
					gameId, _ := strconv.Atoi(rs.gameId)
					game := table.GetTables().RbRoomStockTable.Get(int32(gameId))
					if game != nil {
						rsp.RoomType = game.Rtype
					}
				case int32(pb.CRASH), int32(pb.PLANE), int32(pb.SEVEN):
					rsp.RoomType = 0
				}
			}
			rsp.Code = rs.roomCode
		} else {
			// 前端提示一下所在私人房已解散
			if rs.privRoomDismissed {
				rsp.PrivRoomDismissed = true
				rs.privRoomDismissed = false
			}
		}
		// 本次登录token
		rsp.LoginToken = rs.token
		rsp.Data.HasFailutr = rs.HasFailUtr(ctx)
		// if rs.BankPhone != "" {
		// 	rsp.Bank = true
		// }
		rs.Send(rsp)

		//初始化消息
		rs.sendSystem()
		// 登录上报
		if err = mq.NatsPublish(mq.TopicGameLogin, &pb.PublishGameLogin{
			UserPid: rs.pid,
			Userid:  rs.Userid,
			// Reconnect: reconnect,
			Ts: time.Now().Unix(),
		}); err != nil {
			glog.Error("publish user login error", err)
		}
		// 登录后到大厅上报（外接退出后会走刷新重新连接）
		if !rs.InGame() {
			if err = mq.NatsPublish(mq.TopicGameToLobby, &pb.PublishGameToLobby{
				UserPid: rs.pid,
				Userid:  rs.Userid,
				Ts:      time.Now().Unix(),
				// Reconnect: reconnect,
				Login: true,
			}); err != nil {
				glog.Error("publish user login to lobby error", err)
			}
		}
		//登录成功
		rs.online = true
	}
}

func (rs *RoleActor) GotUserData(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.GotUserData)
	glog.Debugf("GotUserData %#v", arg)
	rsp := handler.UserDataMsg(arg)
	rs.Send(rsp)
}

// func (rs *RoleActor) BroadcastNtf(ctx actor.Context) {
// 	msg := ctx.Message()
// 	// 跑马灯
// 	arg := msg.(*pb.BroadcastNtf)
// 	glog.Debugf("BroadcastNtf %#v", arg)
// 	rs.Send(arg)
// }

// func (rs *RoleActor) FirstRechargeReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	// 请求首冲活动
// 	arg := msg.(*pb.FirstRechargeReq)
// 	glog.Debugf("FirstRechargeReq %#v", arg)
// 	res := config.GetFirstRecharge(rs.User)
// 	rs.Send(res)
// }

func (rs *RoleActor) RechargeActivityReq(ctx actor.Context) {
	msg := ctx.Message()
	// 请求充值活动
	arg := msg.(*pb.RechargeActivityReq)
	glog.Debugf("RechargeActivityReq %#v", arg)
	res := config.GetRechargeActivity(rs.User)
	rs.Send(res)
}

func (rs *RoleActor) WeeklyCardReq(ctx actor.Context) {
	msg := ctx.Message()
	// 请求周卡
	arg := msg.(*pb.WeeklyCardReq)
	glog.Debugf("WeeklyCardReq %#v", arg)
	// res := config.GetWeeklyCard(rs.User)
	res := new(pb.WeeklyCardRsp)
	if rs.WeeklyCardMap == nil {
		rs.WeeklyCardMap = make(map[int32]*data.WeekCard)
	}
	tables := table.GetTables().GiftRechargeTable.GetDataList()
	for _, t := range tables {
		if t.GiftType != 3 {
			continue
		}
		// 兼容老的
		idstr := t.Id[len(t.Id)-1:]
		id, _ := strconv.Atoi(idstr)
		var dailyGive, lastDayGive int32
		for _, v := range t.DailyGive[0].Nums {
			dailyGive += v
		}
		for _, v := range t.DailyGive[len(t.DailyGive)-1].Nums {
			lastDayGive += v
		}
		bean := &pb.WeeklyCardData{
			Id:            int32(id),
			Price:         t.Price,
			Number:        t.Price,
			DuringDay:     int32(len(t.DailyGive)),
			DailyReward:   int32(int64(t.Price) * int64(dailyGive) / 10000),
			LastDayReward: int32(int64(t.Price) * int64(lastDayGive) / 10000),
			State:         1,
		}
		if da, ok := rs.WeeklyCardMap[int32(id)]; ok {
			if da.OverTime == -1 {
				res.Data = append(res.Data, bean)
				continue
			}
			if utils.LocalTime().Unix() >= da.OverTime {
				bean.State = 3
				// 下次可领取时间
				bean.NextTime = utils.TimestampTomorrow(location)
				res.Data = append(res.Data, bean)
				continue
			}
			if da.Get {
				bean.State = 2
			} else {
				bean.State = 3
				// 下次可领取时间
				bean.NextTime = utils.TimestampTomorrow(location)
			}
		}
		res.Data = append(res.Data, bean)
	}

	// 过滤支付渠道支持的商品
	res.Data, _, _ = handleMatchPayShops(res.Data,
		func(s *pb.WeeklyCardData) int64 { return int64(s.Price) },
		func(s *pb.WeeklyCardData, options []int32) { s.PaymentOptions = options },
		func(s *pb.WeeklyCardData, apps []int32) { s.PaymentApps = apps })

	rs.Send(res)
}

func (rs *RoleActor) DailySignDataReq(ctx actor.Context) {
	msg := ctx.Message()
	// 请求签到数据
	arg := msg.(*pb.DailySignDataReq)
	glog.Debugf("DailySignDataReq %#v", arg)
	res := config.GetDailSignData(rs.User)
	rs.Send(res)
}

func (rs *RoleActor) DailySignReceiveReq(ctx actor.Context) {
	msg := ctx.Message()
	// 请求领取签到奖励
	arg := msg.(*pb.DailySignReceiveReq)
	glog.Debugf("DailySignReceiveReq %#v", arg)
	res := rs.GetDailSignReward(arg, rs.User)
	rs.Send(res)
}

func (rs *RoleActor) BindPhoneReq(ctx actor.Context) {
	msg := ctx.Message()
	// 绑定手机
	arg := msg.(*pb.BindPhoneReq)
	glog.Debugf("RoleBuild %#v", arg)
	rs.Send(rs.bindPhone2(arg, ctx))
}

// func (rs *RoleActor) FirstRechargeRewardReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	// 首充领奖
// 	arg := msg.(*pb.FirstRechargeRewardReq)
// 	glog.Debugf("FirstRechargeRewardReq %#v", arg)
// 	rs.getFirstRechargeReward(arg)
// }

func (rs *RoleActor) WeeklyCardReceiveReq(ctx actor.Context) {
	msg := ctx.Message()
	// 周卡领奖
	arg := msg.(*pb.WeeklyCardReceiveReq)
	glog.Debugf("WeeklyCardReceiveReq %#v", arg)
	rs.getWeekCardDaily(arg)
}

func (rs *RoleActor) OnlineRewardDataReq(ctx actor.Context) {
	msg := ctx.Message()
	// 请求在线领奖数据
	arg := msg.(*pb.OnlineRewardDataReq)
	glog.Debugf("OnlineRewardDataReq %#v", arg)
	rs.onlineRewardData(arg)
}

func (rs *RoleActor) OnlineRewardReceiveReq(ctx actor.Context) {
	msg := ctx.Message()
	// 在线领奖
	arg := msg.(*pb.OnlineRewardReceiveReq)
	glog.Debugf("OnlineRewardReceiveReq %#v", arg)
	rs.getOnlineReward()
}

func (rs *RoleActor) OnlineRewardTimeReq(ctx actor.Context) {
	msg := ctx.Message()
	// 在线领奖
	arg := msg.(*pb.OnlineRewardTimeReq)
	glog.Debugf("OnlineRewardTimeReq %#v", arg)
	rs.getOnlineTime()
}

// 提现配置
func (a *RoleActor) GetWithDrawDataReq(ctx actor.Context) {
	msg := ctx.Message()
	// 提请配置
	arg := msg.(*pb.GetWithDrawDataReq)
	glog.Debugf("GetWithDrawDataReq %#v", arg)
	rsp := a.resWithDrawData()
	a.Send(rsp)
}

// 提现日志
func (a *RoleActor) WithdrawLogReq(ctx actor.Context) {
	glog.Debugf("WithdrawLogReq")
	if a.User.WithdrawLogMap == nil {
		a.User.WithdrawLogMap = make(map[string]*data.WithdrawLog)
	}
	rsp := new(pb.WithdrawLogRsp)
	logs := make([]*pb.WithdrawLog, 0)
	for _, v := range a.User.WithdrawLogMap {
		bean := &pb.WithdrawLog{
			Id:     v.Id,
			Time:   v.Ctime,
			Amount: v.Amount,
			Status: pb.WithdrawLog_State(v.Status - 1),
		}
		logs = append(logs, bean)
	}
	sort.Slice(logs, func(i, j int) bool { return logs[i].Time > logs[j].Time })
	rsp.Logs = logs
	a.Send(rsp)
}

// 事件传递
func (a *RoleActor) EventPost(ctx actor.Context) {
	arg := ctx.Message().(*pb.EventPost)
	glog.Debugf("EventPost %#v", arg)
	rsp := event.EventPost(a.User, arg.EventId, arg.Data)

	if ntf, ok := rsp.(*pb.GiftWelfareNtf); ok {
		if ntf.Gtype == 2 {
			// 破产礼包触发前检查是否触发波动返水
			r, err := a.rolePid.RequestFuture(&pb.VolatilitySubsidyCheck{Userid: a.Userid}, 3*time.Second).Result()
			if err != nil {
				glog.Error(err)
			} else {
				rsp, ok := r.(*pb.VolatilitySubsidyChecked)
				if !ok {
					glog.Errorf("type error: %#v", r)
				} else if rsp.Trigger {
					glog.Infof("user %s trigger pbgs gift, also trigger volatility subsidy", a.Userid)
					return
				}
			}
		}

		// 过滤支付渠道支持商品
		gifts, payMin, payMax := handleMatchPayShops(ntf.Gifts,
			func(s *pb.GiftWelfare) int64 { return int64(s.Price) },
			func(s *pb.GiftWelfare, options []int32) { s.PaymentOptions = options },
			func(s *pb.GiftWelfare, apps []int32) { s.PaymentApps = apps })
		ntf.Gifts = gifts
		ntf.MinRecharge = int32(payMin)
		ntf.MaxRecharge = int32(payMax)
	}

	if rsp != nil {
		a.Send(rsp)
	}
	a.status = true
}

// 客服消息
func (a *RoleActor) FeedBackReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.FeedBackReq)
	glog.Debugf("FeedBackReq %#v", arg)
	a.feedBackReq(arg)
}

// 读取客服消息
func (a *RoleActor) ReadFeedBackLogReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ReadFeedBackLogReq)
	glog.Debugf("ReadFeedBackLogReq %#v", arg)
	a.readFeedBackLogReq(arg)
}

// 客服回复
func (a *RoleActor) FeedBackReply(ctx actor.Context) {
	arg := ctx.Message().(*pb.FeedBackReply)
	glog.Debugf("FeedBackReply %#v", arg)
	a.feedBackReply(arg)
}

// 玩家上传文件
func (rs *RoleActor) CustomerUploadFile(ctx actor.Context) {
	arg := ctx.Message().(*pb.CustomerUploadFile)
	glog.Debugf("CustomerUploadFile %#v", arg)
	ntf := new(pb.CustomerUploadFileNtf)
	defer rs.Send(ntf)

	if arg.Error != "" {
		ntf.Error = pb.Failed
		glog.Error("upload file error: ", arg.Error)
		return
	}

	req := &pb.CustomerSendReq{
		Userid:   rs.User.Userid,
		Username: rs.User.Nickname,
		Ctype:    arg.Ctype,
		Content:  arg.Url,
		Filename: arg.FileName,
		Filesize: arg.FileSize,
	}

	ret := new(pb.CustomerMessage)
	err := mq.NatsRequest(mq.RequestCustomerSend, ret, req,
		mq.WithRequestTimeout(3*time.Second))
	if err != nil {
		glog.Error(err)
		ntf.Error = pb.Failed
		return
	}

	rs.User.FeedBackTimes++
	rs.status = true
	ntf.Msg = ret
}

// 玩家发送消息
func (rs *RoleActor) CustomerSendReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CustomerSendReq)
	glog.Debugf("CustomerSendReq %#v", arg)
	rsp := new(pb.CustomerSendRsp)

	customerTable := table.GetTables().CustomerTable.Get()
	// 内容检测
	if arg.Ctype == 0 && len(arg.Content) > int(customerTable.MessageLength) {
		rsp.Error = pb.FeedBackFail
		rs.Send(rsp)
		return
	}
	// 发送次数
	times := rs.User.FeedBackTimes
	if times >= customerTable.ReplySendLimit {
		rsp.Error = pb.CustomerSendLimited
		rs.Send(rsp)
		return
	}

	defer rs.Send(rsp)

	arg.Userid = rs.User.Userid
	arg.Username = rs.User.Nickname
	ret := new(pb.CustomerMessage)
	err := mq.NatsRequest(mq.RequestCustomerSend, ret, arg,
		mq.WithRequestTimeout(3*time.Second))
	if err != nil {
		glog.Error(err)
		rsp.Error = pb.Failed
		return
	}

	rs.User.FeedBackTimes++
	rs.status = true
	rsp.Msg = ret
}

// 重置客服聊天发送次数 解除频繁
func (rs *RoleActor) customerResetFeedBackTimes() {
	if rs.User.FeedBackTimes != 0 {
		rs.User.FeedBackTimes = 0
		rs.status = true
	}
}

// 客服回复消息
func (rs *RoleActor) ConsumerReplyMessage(ctx actor.Context) {
	arg := ctx.Message().(*pb.ConsumerReplyMessage)
	rs.customerResetFeedBackTimes()

	ntf := new(pb.ConsumerReplyNtf)
	ntf.Msgs = append(ntf.Msgs, arg.Msg)
	rs.Send(ntf)
}

// 客服结束对话
func (rs *RoleActor) PublishCustomerSessionOver(ctx actor.Context) {
	arg := ctx.Message().(*pb.PublishCustomerSessionOver)
	rs.customerResetFeedBackTimes()

	// 通知评分
	ntf := new(pb.CustomerScoreNtf)
	ntf.SessionId = arg.SessionId
	rs.Send(ntf)
}

// 客服撤回消息
func (rs *RoleActor) PublishCustomerMessageRevoke(ctx actor.Context) {
	arg := ctx.Message().(*pb.PublishCustomerMessageRevoke)
	// 通知评分
	ntf := new(pb.CustomerMessageRevokeNtf)
	ntf.MessageId = arg.MessageId
	rs.Send(ntf)
}

// 客服消息已读
func (rs *RoleActor) CustomerReadAck2Req(ctx actor.Context) {
	arg := ctx.Message().(*pb.CustomerReadAck2Req)
	glog.Debugf("CustomerReadAck2Req: %v", arg)
	rsp := new(pb.CustomerReadAck2Rsp)
	defer rs.Send(rsp)

	if len(arg.ReadReplyMsgs) == 0 {
		return
	}

	var msgs []int64
	for _, msg := range arg.ReadReplyMsgs {
		msgId, err := strconv.ParseInt(msg, 10, 64)
		if err != nil {
			glog.Errorf("msgId parse error: %s, %s", rs.User.Userid, msg)
			continue
		}
		msgs = append(msgs, msgId)
	}
	err := mq.NatsPublish(mq.TopicGameCustomerRead, &pb.PublishGameCustomerRead{
		Userid: rs.User.Userid,
		Msgs:   msgs,
	})
	if err != nil {
		glog.Error(err)
		return
	}
}

// 客服消息记录
func (rs *RoleActor) CustomerHistoryReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CustomerHistoryReq)
	rsp := new(pb.CustomerHistoryRsp)
	defer rs.Send(rsp)

	prevLastId, _ := strconv.ParseInt(arg.PrevLastId, 10, 64)
	sql_session_filter := ""
	if prevLastId != 0 {
		sql_session_filter = fmt.Sprintf("and id < %d", prevLastId)
	}
	// 查ck聊天会话
	sql_session := `
		select id, score, ctime, etime from game.col_customer_chat_sessions final 
		where userid = ? and ctime > ? %s
		order by id desc limit 2
	`
	var session_datas []map[string]any
	err := ck.Select(&session_datas, fmt.Sprintf(sql_session, sql_session_filter), rs.User.Userid, rs.User.Ctime.UnixMilli())
	if err != nil {
		glog.Error(err)
		rsp.Error = pb.Failed
		return
	}
	sessions := len(session_datas)
	rsp.HasMore = sessions > 1

	var sesssion_close bool
	_ = sesssion_close
	// 欢迎语/问题列表
	defer func() {
		nowMs := time.Now().UnixMilli()
		var session_time = nowMs
		if len(session_datas) > 0 {
			ctime := utils.ToInt64(session_datas[0]["ctime"])
			if ctime > 0 {
				session_time = ctime
			}
		}
		if len(table.GetTables().CustomerQuestionTable.GetDataList()) > 0 {
			mayAskMsg := &pb.CustomerMessage{
				Reply: true,
				Ctype: 12,
				Ctime: time.UnixMilli(session_time).In(location).Format(utils.FORMAT),
			}
			rsp.Msgs = append(rsp.Msgs, mayAskMsg)
		}
		wellcome := table.GetTables().CustomerTable.Get().WellcomeText
		if wellcome != "" {
			wellcomeMsg := &pb.CustomerMessage{
				Reply:   true,
				Ctype:   11,
				Content: "Customer service BigWin is at your service",
				Ctime:   time.UnixMilli(session_time).In(location).Format(utils.FORMAT),
			}
			rsp.Msgs = append(rsp.Msgs, wellcomeMsg)
		}

	}()

	if sessions == 0 {
		return
	}
	// 查ck聊天记录
	sessionId := utils.ToInt64(session_datas[0]["id"])
	score := utils.ToInt64(session_datas[0]["score"])
	etime := utils.ToInt64(session_datas[0]["etime"])

	sesssion_close = etime > 0
	rsp.LastId = fmt.Sprint(sessionId)

	// 已评分消息
	if score > 0 {
		rsp.Msgs = append(rsp.Msgs, &pb.CustomerMessage{
			Ctype: 10,
			Score: int32(score),
		})
	}

	var msg_datas []map[string]any
	err = ck.Select(&msg_datas, `
		select id, reply, ctype, content, ctime, read, filename
		from game.col_customer_chat_messages final 
		where userid = ? and session_id = ? and ctime > ? and revoke = 0
		order by ctime desc
	`, rs.User.Userid, sessionId, rs.User.Ctime.UnixMilli())
	if err != nil {
		glog.Error(err)
		rsp.Error = pb.Failed
		return
	}
	var unreadMsgs []int64
	for _, data := range msg_datas {
		id := utils.ToInt64(data["id"])
		reply := utils.ToInt64(data["reply"])
		ctype := utils.ToInt64(data["ctype"])
		content := data["content"].(string)
		ctime := utils.ToInt64(data["ctime"])
		read := utils.ToInt64(data["read"])
		filename := data["filename"].(string)

		msg := &pb.CustomerMessage{
			MessageId: fmt.Sprint(id),
			Reply:     reply == 1,
			Ctype:     int32(ctype),
			Content:   content,
			Ctime:     time.UnixMilli(ctime).In(location).Format(utils.FORMAT),
			Read:      read == 1,
			Filename:  filename,
		}
		rsp.Msgs = append(rsp.Msgs, msg)

		// 首次打开窗口
		if prevLastId == 0 {
			if msg.Reply && !msg.Read {
				rsp.Unread++
				unreadMsgs = append(unreadMsgs, id)
			}
		}
	}

	if len(unreadMsgs) > 0 {
		err := mq.NatsPublish(mq.TopicGameCustomerRead, &pb.PublishGameCustomerRead{
			Userid: rs.User.Userid,
			Msgs:   unreadMsgs,
		})
		if err != nil {
			glog.Error(err)
			return
		}
	}
}

// 客户会话评分
func (rs *RoleActor) CustomerScoreReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CustomerScoreReq)
	rsp := new(pb.CustomerScoreRsp)
	defer rs.Send(rsp)

	arg.Userid = rs.User.Userid
	err := mq.NatsPublish(mq.TopicGameCustomerScore, arg)
	if err != nil {
		glog.Error(err)
		rsp.Error = pb.Failed
		return
	}

	// 评分消息
	rsp.Msg = &pb.CustomerMessage{
		Ctype: 10,
		Score: arg.Score,
	}
}

func (a *RoleActor) PointControl(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.PointControl)
	glog.Debugf("PointControl %#v", arg)
	data := &data.PointControl{
		UserId: arg.UserId,
		Switch: arg.Switch,
		Factor: arg.Factor,
		Score:  arg.Score,
	}
	handler.PointControl(a.User, data)
	a.status = true
	if a.gamePid != nil {
		a.gamePid.Tell(msg)
	}
}

func (a *RoleActor) ChangePointControl(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ChangePointControl)
	glog.Debugf("ChangePointControl %#v", arg)
	a.PCSwitch = arg.Switch
	a.PCFactor = arg.Factor
	a.PCScore = arg.Score
	a.PCScoreComplete = arg.ScoreComplete
	a.status = true
}

// 改变用户信息
func (a *RoleActor) ChangeUser(ctx actor.Context) {
	arg := ctx.Message().(*pb.ChangeUser)
	glog.Debugf("ChangeUser %#v", arg)
	data := &data.ModifyUserData{}
	err := json.Unmarshal(arg.Data, data)
	if err != nil {
		return
	}
	user := a.User
	handler.ModifyUserData(user, data)
	a.status = true
	// 通知客户端
	ntf := &pb.UserDataNtf{
		Userid:     a.Userid,
		NickName:   user.Nickname,
		Photo:      user.Photo,
		Phone:      user.Phone2,
		BankName:   user.Bank,
		BankNumber: user.BankAccounts,
		Ifsc:       user.IFSC,
	}
	a.Send(ntf)
}

func (a *RoleActor) OnlineUser(ctx actor.Context) {
	arg := ctx.Message().(*pb.OnlineUser)
	glog.Debugf("OnlineUser %#v", arg)
	// if a.Robot {
	// 	ctx.Respond(data.OnlineUser{})
	// 	return
	// }
	res := new(pb.OnlinedUser)
	if !a.online {
		ctx.Respond(res)
		return
	}
	res.Online = true
	res.GameId = a.gameId
	res.RoomId = a.roomId
	res.Gtype = utils.String(a.gtype)
	ctx.Respond(res)
}

// 分享活动数据
func (a *RoleActor) ShareDataReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareDataReq)
	glog.Debugf("ShareDataReq %#v", arg)
	a.shareData(arg)
}

// 分享领奖
func (a *RoleActor) ShareWithdrawReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareWithdrawReq)
	glog.Debugf("ShareWithdrawReq %#v", arg)
	a.shareWithdraw(arg)
}

// 分享下线信息
func (a *RoleActor) ShareReferralReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareReferralReq)
	glog.Debugf("ShareReferralReq %#v", arg)
	a.shareReferralReq(arg)
}

// 分享下线信息
func (a *RoleActor) ShareRewardReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareRewardReq)
	glog.Debugf("ShareRewardReq %#v", arg)
	a.shareRewardReq(arg)
}

// 分享详情
func (a *RoleActor) ShareRewardDetailReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareRewardDetailReq)
	glog.Debugf("ShareRewardDetailReq %#v", arg)
	a.shareRewardDetailReq(arg)
}

// 引导详情
func (a *RoleActor) GameGuidReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.GameGuidReq)
	glog.Debugf("GameGuidReq %#v", arg)
	a.gameGuidReq(arg)
}

// 引导详情
func (a *RoleActor) BugFeedbackReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.BugFeedbackReq)
	glog.Debugf("BugFeedbackReq %#v", arg)
	rsp := new(pb.BugFeedbackRsp)
	if a.BugFeedback >= 3 {
		rsp.Error = pb.BugFeedTimesLimit
		a.Send(rsp)
		return
	}
	a.BugFeedback++
	a.status = true
	a.Send(rsp)

	arg.Userid = a.Userid
	a.rolePid.Tell(arg)
}

// 按钮点击
func (a *RoleActor) ButtonClickReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ButtonClickReq)
	// glog.Debugf("ButtonClickReq %#v", arg)
	rsp := new(pb.ButtonClickRsp)
	a.Send(rsp)

	log := &pb.LogButtonClick{
		Userid: a.Userid,
		Typ:    arg.Typ,
	}
	myactor.Logger().Tell(log)
}

// 按钮点击（每个按钮都记录的）
func (a *RoleActor) ButtonClick2Req(ctx actor.Context) {
	arg := ctx.Message().(*pb.ButtonClick2Req)
	// glog.Debugf("ButtonClick2Req %#v", arg)
	rsp := new(pb.ButtonClick2Rsp)
	a.Send(rsp)

	log := &pb.LogButtonClick2{
		Userid: a.Userid,
		Typ:    arg.Typ,
	}
	myactor.Logger().Tell(log)
}

// 充值时间埋点
func (a *RoleActor) RechargeReportReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.RechargeReportReq)
	glog.Debugf("RechargeReportReq %#v", arg)
	rsp := new(pb.RechargeReportRsp)
	a.Send(rsp)

	log := &pb.LogRechargeReport{
		Uid:     arg.ReportId,
		Code:    int32(arg.Code),
		Content: arg.Body,
	}
	myactor.Logger().Tell(log)
}

// 修改银行卡信息
func (a *RoleActor) SaveBankInfoReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.SaveBankInfoReq)
	glog.Debugf("SaveBankInfoReq %#v", arg)
	rsp := new(pb.SaveBankInfoRsp)

	arg.BlankNumber = strings.ReplaceAll(arg.BlankNumber, " ", "")
	arg.IFSC = strings.ReplaceAll(arg.IFSC, " ", "")

	if arg.PayWay == 1 {
		a.saveBank3(arg)
		rsp.PayWay = arg.PayWay
		rsp.BlankNumber = arg.BlankNumber
		a.Send(rsp)
		return
	}

	//检查卡号
	if arg.BlankAccount == "" || arg.BlankNumber == "" {
		rsp.Error = pb.CheckBankInfo
		a.Send(rsp)
		return
	}
	a.saveBank2(arg)
	rsp.BlankAccount = arg.BlankAccount
	rsp.BlankNumber = arg.BlankNumber
	rsp.IFSC = arg.IFSC
	rsp.BlankAccountHolder = arg.BlankAccountHolder
	a.Send(rsp)
}

// 绑定adid
func (a *RoleActor) AdjustAdidReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.AdjustAdidReq)
	glog.Debugf("AdjustAdidReq %#v", arg)
	rsp := new(pb.AdjustAdidRsp)
	a.AD_ADID = arg.Adid
	a.rolePid.Tell(&pb.SetAdid{Userid: a.Userid, Adid: arg.Adid})

	if a.AD_RefGameId == "" && arg.RefGameId != "" && a.adid == "" {
		// 分享绑定
		a.AD_RefGameId = arg.RefGameId
		shareStr := utils.Split(arg.RefGameId, "-")

		if len(shareStr) > 1 && strings.Contains(shareStr[0], "share") {
			superId := shareStr[1]
			shareSource := 0
			if shares := strings.Split(superId, "_"); len(shares) > 1 {
				superId = shares[0]
				shareSource, _ = strconv.Atoi(shares[1])
			}
			if a.Userid != superId {
				if _, ok := a.ShareBelow[superId]; !ok {
					// 绑的不是自己的下级
					a.ShareSuperior = superId
					a.ShareSource = int32(shareSource)
					a.rolePid.Tell(&pb.ShareRegist{Superior: superId, Userid: a.Userid, ShareSource: int32(shareSource)})
				}
			}
		}
	}

	if a.AD_User_Agent == "" {
		// 生成假的ua
		a.AD_Fake_User_Agent = handler.GenUserAgent()
		glog.Warning("gen fake ua param:", a.AD_Fake_User_Agent)
	}

	glog.Infof("%s receive adid %s", a.Userid, arg.Adid)
	a.status = true
	a.Send(rsp)

	// 上报事件
	a.adid = arg.Adid
}

// 公告
func (a *RoleActor) AnnouncementReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.AnnouncementReq)
	glog.Debugf("AnnouncementReq %#v", arg)
	rsp := new(pb.AnnouncementRsp)

	notices := config.GetNotices2()
	for _, n := range notices {
		if n.Rtype != 1 {
			continue
		}
		rsp.Content = n.Content
		break
	}

	a.Send(rsp)
}

// cdk
func (a *RoleActor) CDKReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CDKReq)
	glog.Debugf("CDKReq %#v", arg)

	if a.User.CdkCodeMap == nil {
		a.User.CdkCodeMap = make(map[string]string)
	}

	// 是否领过
	if _, ok := a.User.CdkCodeMap[arg.Code]; ok {
		rsp := &pb.CDKRsp{Error: pb.InvalidCdk}
		a.Send(rsp)
		return
	}

	arg.Userid = a.Userid
	rsp, err := a.rolePid.RequestFuture(arg, 5*time.Second).Result()
	if err != nil {
		rsp := &pb.CDKRsp{Error: pb.InvalidCdk}
		a.Send(rsp)
		return
	}

	if r, ok := rsp.(*pb.CDKRsp); ok {
		if r.Error == pb.OK {
			if r.Bouns > 0 {
				// 加罐子
				ntf := event.Event(a.User, event.Normal_JAR, &event.NormalJarEvent{Give: r.Bouns, Multiple: 40})
				if ntf != nil {
					a.Send(ntf)
				}
				// 日志
				a.BonusLog(a.Userid, r.Bouns)
			} else {
				// 发奖
				var diamond int64 = r.Cash
				if r.Give > 0 {
					diamond = r.Give
				}
				// CDKReq
				if r.Out > 0 {
					diamond = r.Out
				}
				a.sendGood(diamond, 0, r.Give, r.Out, 0, 0, int32(pb.LOG_TYPE85), fmt.Sprint("CDK:", arg.Code))
			}
			a.User.CdkCodeMap[arg.Code] = ""
			a.status = true
		}
		a.Send(r)
		return
	}

	r := &pb.CDKRsp{Error: pb.InvalidCdk}
	a.Send(r)
}

func (a *RoleActor) GetUserInfoReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.GetUserInfoReq)
	glog.Debugf("GetUserInfoReq %#v", arg)

	rsp := &pb.GetUserInfoRsp{}
	rsp.NickName = a.User.Nickname
	rsp.Balance = a.User.Diamond
	rsp.Photo = a.User.Photo
	rsp.Recharge = int64(a.User.Money)
	rsp.UserType = int32(a.User.RegistArea)

	ctx.Respond(rsp)
}

func (a *RoleActor) ExternalBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ExternalBetReq)
	glog.Debugf("ExternalBetReq %#v", arg)

	if a.gamePid != nil {
		rsp := &pb.ExternalBetRsp{Err: "already in the game", ErrCode: 500}
		ctx.Respond(rsp)
		return
	}

	if arg.Amount > 0 && arg.Amount > a.User.Diamond {
		rsp := &pb.ExternalBetRsp{Err: "insufficient balance", ErrCode: 30005}
		ctx.Respond(rsp)
		return
	}

	// a.externalModifyCurrency(-arg.Amount, int32(pb.LOG_TYPE94))
	a.addCurrency(-arg.Amount, 0, 0, 0, 0, 0, 0, int32(pb.LOG_TYPE94), "外接下注", arg.RoundId)
	a.vbFlow(arg.Amount)
	// a.User.Diamond -= arg.Amount
	// a.status = true

	id, _ := strconv.ParseUint(a.User.Userid, 10, 32)

	rsp := &pb.ExternalBetRsp{
		MerchantOrderNo: strconv.FormatInt(handler.GenerateOrderId(uint32(id)), 10),
		OrderNo:         arg.OrderNo,
		Balance:         a.User.Diamond,
	}
	ctx.Respond(rsp)

	//日志
	betLog := &pb.NsqLogExternalBet{
		Uid:             a.User.Userid,
		GameId:          arg.GameId,
		RoundId:         arg.RoundId,
		MerchantOrderNo: rsp.MerchantOrderNo,
		OrderNo:         rsp.OrderNo,
		Amount:          arg.Amount,
	}
	body, _ := proto.Marshal(betLog)

	log := &pb.NsqLog{
		Typ:  pb.ExternalBet,
		Body: body,
	}
	// body2, _ := proto.Marshal(log)
	// producer.Publish(data.TopicLog, body2)
	mq.NatsPublish(mq.TopicReportLog, log)
}

func (a *RoleActor) ExternalRewardReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ExternalRewardReq)
	glog.Debugf("ExternalRewardReq %#v", arg)

	if a.gamePid != nil {
		rsp := &pb.ExternalBetRsp{Err: "already in the game", ErrCode: 500}
		ctx.Respond(rsp)
		return
	}

	// a.externalModifyCurrency(arg.RewardAmount, int32(pb.LOG_TYPE94))

	if arg.RewardAmount != 0 {
		var out, give int64 = 0, 0
		if arg.RewardAmount > 0 {
			out += arg.RewardAmount
		}
		a.addCurrency(arg.RewardAmount, 0, 0, 0, give, out, 0, int32(pb.LOG_TYPE95), "外接结算", arg.RoundId)
	}
	if a.OutDiamond > a.Diamond {
		a.addCurrency(0, 0, 0, 0, 0, a.Diamond-a.OutDiamond, 0, int32(pb.LOG_TYPE95), "外接扣提现金", arg.RoundId)
	}
	// a.User.Diamond += arg.RewardAmount
	a.status = true

	id, _ := strconv.ParseUint(a.User.Userid, 10, 32)

	rsp := &pb.ExternalRewardRsp{
		MerchantOrderNo: strconv.FormatInt(handler.GenerateOrderId(uint32(id)), 10),
		OrderNo:         arg.OrderNo,
		Balance:         a.User.Diamond,
	}
	ctx.Respond(rsp)

	//日志
	rewardLog := &pb.NsqLogExternalReward{
		Uid:             a.User.Userid,
		GameId:          arg.GameId,
		RoundId:         arg.RoundId,
		MerchantOrderNo: rsp.MerchantOrderNo,
		OrderNo:         rsp.OrderNo,
		RewardAmount:    arg.RewardAmount,
	}

	body, _ := proto.Marshal(rewardLog)

	log := &pb.NsqLog{
		Typ:  pb.ExternalReward,
		Body: body,
	}
	// body2, _ := proto.Marshal(log)
	// producer.Publish(data.TopicLog, body2)
	mq.NatsPublish(mq.TopicReportLog, log)
}

// 视讯交易
func (a *RoleActor) EfiTransactionReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.EfiTransactionReq)
	glog.Debugf("EfiTransactionReq %#v", arg)
	rsp := new(pb.EfiTransactionRsp)

	beforeBalance := float64(a.GetDiamond()) / 100.0
	var amountTotal float64
	for _, trans := range arg.Transactions {
		amountTotal += trans.TransactionAmount
	}
	amount := int64(amountTotal * 100.0)
	if amount > 0 && amount > a.GetDiamond() {
		rsp := &pb.EfiTransactionRsp{Err: "insufficient balance"}
		ctx.Respond(rsp)
		return
	}
	desc := fmt.Sprintf("视讯-%s", arg.TransTypeName)
	// 先加再减, 避免不够扣
	for _, trans := range arg.Transactions {
		if trans.TransactionAmount > 0 {
			amount := int64(trans.TransactionAmount * 100)
			a.addCurrency(amount, 0, 0, 0, 0, 0, 0, int32(pb.LOG_TYPE105), desc, trans.TransactionID)
		}
	}
	for _, trans := range arg.Transactions {
		if trans.TransactionAmount < 0 {
			amount := int64(trans.TransactionAmount * 100)
			a.addCurrency(amount, 0, 0, 0, 0, 0, 0, int32(pb.LOG_TYPE105), desc, trans.TransactionID)
		}
	}

	rsp.BeforeBalance = beforeBalance
	rsp.Balance = float64(a.GetDiamond()) / 100.0
	ctx.Respond(rsp)
}

// 破产礼包
func (a *RoleActor) BreakingGiftReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.BreakingGiftReq)
	glog.Debugf("BreakingGiftReq %#v", arg)
	rsp := new(pb.BreakingGiftRsp)

	bean := config.GetGame(arg.GameId)
	if bean.Id == "" || handler.IsFreeGame(bean.Gtype) {
		glog.Errorf("no found game config ,gtype:%d,id:%s", arg.Gtype, arg.GameId)
		rsp.Error = pb.Failed
		a.Send(rsp)
		return
	}

	for _, g := range a.User.BreakingGift {
		if g.Gtype == arg.Gtype && g.GameId == arg.GameId {
			if bean.BreakingTimes <= g.BuyTimes {
				rsp.Error = pb.Failed
				a.Send(rsp)
				return
			}
		}
	}

	a.BreakGtype = arg.Gtype
	a.BreakGameId = arg.GameId

	rsp.Cash = int64(bean.BreakingPrice)
	rsp.Price = int32(bean.BreakingPrice)
	rsp.Give = int64(bean.BreakingGive)
	// if a.VBBank >= int64(bean.BreakingGive) {
	// }
	a.Send(rsp)
}

// 商城罐子
func (a *RoleActor) ShopPotRewardReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShopPotRewardReq)
	glog.Debugf("ShopPotRewardReq %#v", arg)

	var pot data.ShopPot
	var index int
	rsp := new(pb.ShopPotRewardRsp)

	defer a.Send(rsp)

	for i, p := range a.ActivatedPot {
		if p.Id == arg.Id {
			pot = p
			index = i
			break
		}
	}

	if pot.Id == "" {
		rsp.Error = pb.ShopPotNoFound
		return
	}

	if a.PotFlow < pot.FlowWater {
		rsp.Error = pb.ShopPotFlowUnenough
		return
	}

	a.sendGood(pot.Number, 0, 0, 0, 0, 0, int32(pb.LOG_TYPE96), "罐子")

	a.PotFlow -= pot.FlowWater
	a.ActivatedPot = append(a.ActivatedPot[:index], a.ActivatedPot[index+1:]...)

	rsp.ShopPots = handler.BuildShopPotDataMsg(a.User)
	rsp.PotsFlow = a.PotFlow

	// 日志
	a.BonusLog(a.Userid, -pot.Number)

	a.status = true
}

// VIP
func (a *RoleActor) VipRewardReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.VipRewardReq)
	glog.Debugf("VipRewardReq %#v", arg)

	user := a.User
	rsp := new(pb.VipRewardRsp)

	defer a.Send(rsp)

	vipBean := table.GetTables().VipTable.Get(arg.Id)
	switch arg.Rtype {
	case 1:
		// 每日
		// if user.Vip.Daily || vipBean.DailySign <= 0 {
		// 	rsp.Error = pb.AlreadyAward
		// 	return
		// }
		// a.sendGood(vipBean.DailySign, 0, -vipBean.DailySign, 0, 0, 0, int32(pb.LOG_TYPE97), fmt.Sprintf("vip%d每日签到释放%.2fVB金", vipBean.Id, float64(vipBean.DailySign)/100))
		// user.Vip.Daily = true
	case 2:
		// 每周
		if user.Vip.Weekly || vipBean.WeeklySign <= 0 {
			rsp.Error = pb.AlreadyAward
			//fix 老弹窗
			user.Vip.Weekly = true
			return
		}
		if user.Vip.WeeklyTime <= 0 {
			nowTime := time.Now().In(location)
			weekday := nowTime.Weekday()
			offset := int(-weekday + 1)
			if weekday == time.Sunday {
				offset = -6
			}
			nextMonday := nowTime.AddDate(0, 0, offset+7)
			nextMondayZero := time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(), 0, 0, 0, 0, nowTime.Location())
			user.Vip.WeeklyTime = nextMondayZero.Unix()
			user.Vip.Weekday = int(weekday)
			// zeroTime := utils.TimestampZeroTime(utils.LocalTime(), time.Now().Local().Location())
			// user.Vip.WeeklyTime = zeroTime.AddDate(0, 0, 7).Unix()
			// user.Vip.Weekday = int(zeroTime.Weekday())
		}
		a.sendGood(0, 0, int64(vipBean.WeeklySign), 0, 0, 0, int32(pb.LOG_TYPE98), fmt.Sprintf("vip%d每周签到%.2f", vipBean.Id, float64(vipBean.WeeklySign)/100))
		user.Vip.Weekly = true
	case 3:
		// 等级奖励
		for _, v := range user.Vip.LevelReward {
			if v == arg.Id {
				// 领过了
				rsp.Error = pb.AlreadyAward
				return
			}
		}

		a.sendGood(0, 0, int64(vipBean.Upgrade), 0, 0, 0, int32(pb.LOG_TYPE99), fmt.Sprintf("vip%d等级奖励", vipBean.Id))
		user.Vip.LevelReward = append(user.Vip.LevelReward, arg.Id)
	}
	a.status = true

	rsp.Daily = user.Vip.Daily
	rsp.Weekly = user.Vip.Weekly
	rsp.LvReward = user.Vip.LevelReward
	rsp.WeeklyTime = user.Vip.WeeklyTime
}

// 领取注册奖励
func (a *RoleActor) RegistRewardReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.RegistRewardReq)
	glog.Debugf("RegistRewardReq %#v", arg)
	user := a.User
	rsp := new(pb.RegistRewardRsp)
	if user.RegistReward || user.RegistMode == 0 {
		rsp.Error = pb.AlreadyAward
		a.Send(rsp)
		return
	}
	bean := table.GetTables().NewbieTable.Get()

	reward := handler.BuildRegistReward(a.RegistArea)
	_ = reward

	// 游戏引导和赠送金
	game := bean.GameId[0]
	if a.RegistArea == 1 {
		game = bean.GameId[1]
	}

	// var give int64 = 0
	switch user.RegistMode {
	case 1:
		// 模式1
		// if a.RegistArea == 1 {
		// 	// B类给赠送金
		// 	give = int64(reward[0])
		// }
		// a.sendGood(int64(reward[0]), 0, 0, 0, 0, 0, int32(pb.LOG_TYPE1), "模式1注册奖励")
		rsp.Gameid = fmt.Sprintf("%d", game.Nums[1])
	case 2:
		// 模式2
		// if a.RegistArea == 1 {
		// 	// B类给赠送金
		// 	give = int64(reward[1])
		// }
		// a.sendGood(int64(reward[1]), 0, 0, 0, 0, 0, int32(pb.LOG_TYPE1), "模式2注册奖励")
		rsp.Gameid = fmt.Sprintf("%d", game.Nums[1])
	}
	user.RegistReward = true
	rsp.Gtype = game.Nums[0]
	if user.Phone == "" {
		rsp.Gameid = fmt.Sprintf("%d", game.Nums[1])
	}
	a.status = true
	a.Send(rsp)
	msg := &pb.UpdateRegisReward{
		Userid: a.Userid,
		State:  true,
	}
	a.rolePid.Tell(msg)

	a.ET_RegistReward()
}

// 绑定登录账号
func (a *RoleActor) BindLoginPhoneReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.BindLoginPhoneReq)
	glog.Debugf("BindLoginPhoneReq %#v", arg)
	user := a.User
	res := new(pb.BindLoginPhoneRsp)
	if user.Phone != "" {
		glog.Errorf("user has bind login phone, user:%s", a.Userid)
		return
	}
	arg.Userid = user.Userid
	rsp, err := ctx.RequestFuture(a.rolePid, arg, 3*time.Second).Result()
	if err != nil {
		res.Error = pb.BindPhoneFail
		return
	}
	var ok bool = false
	if res, ok = rsp.(*pb.BindLoginPhoneRsp); ok && res.Error == pb.OK {
		res.Userid = a.Userid
		user.Phone = arg.Phone
		a.status = true
		// 发奖
		if user.RegistMode != 1 {
			a.Send(res)
			return
		}
		if !user.RegistReward && user.RegistMode == 1 {
			user.RegistMode = 2 // 绑定手机后切换为模式2
			a.Send(res)
			return
		}
		// bean := table.GetTables().NewbieTable.Get()
		// var give int64 = 0
		// switch user.RegistMode {
		// case 1:
		// 	reward = int64(bean.RegistMode1)
		// case 2:
		// 	reward = int64(bean.RegistMode2)
		// }
		rewards := handler.BuildRegistReward(a.RegistArea)

		reward := int64(rewards[1] - rewards[0])
		// if user.State != 1 { // 新手玩家
		// 	give = reward
		// }
		if reward > 0 {
			a.sendGood(reward, 0, 0, 0, 0, 0, int32(pb.LOG_TYPE81), "绑定登录手机号")
		}
	}
	a.Send(res)
}

// 关闭提现弹窗
func (a *RoleActor) CloseWithdrawReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CloseWithdrawReq)
	glog.Debugf("CloseWithdrawReq %#v", arg)
	rsp := new(pb.CloseWithdrawRsp)
	if arg.Ctype == 1 {
		a.User.WithdrawWindow = true
	} else if arg.Ctype == 2 {
		a.User.WithdrawWindow2 = true
	}
	rsp.FirstWithdraw = true
	a.status = true
	a.Send(rsp)
}

// 下级注册
func (a *RoleActor) ShareUserRegist(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareUserRegist)
	// share := table.GetTables().ShareConfigTable.Get()
	share := config.GetShare(1)
	if share.Id == 0 {
		return
	}
	user := a.User
	if user.ShareBelow == nil {
		user.ShareBelow = make(map[string]data.ShareData)
	}
	if _, ok := user.ShareBelow[arg.Userid]; ok {
		glog.Errorf("below user has exist, user:%s, superior:%s", arg.Userid, arg.Superior)
		return
	}
	if len(user.ShareBelow) >= int(share.MaxPeople) {
		glog.Errorf("below people upper limit, user:%s", arg.Superior)
		return
	}
	s := data.ShareData{
		UserId:    arg.Userid,
		Name:      arg.Name,
		LastLogin: utils.LocalTime().Unix(),
	}
	user.ShareBelow[arg.Userid] = s

	// 玩游戏抽奖
	ntf := event.Event(user, event.PLAY_SHARE, &event.PlayShareEvent{Ptype: 2, DeviceId: arg.DeviceId})
	if ntf != nil {
		a.Send(ntf)
	}

	a.status = true

	glog.Errorf("share regist success, superior:%s, user:%s", arg.Superior, arg.Userid)
	// 更新分享信息
	a.sharedataNtf()
}

/*
func (rs *RoleActor) addPrize(rtype, ltype, amount int32) {
	switch uint32(rtype) {
	case data.DIAMOND:
		rs.addCurrency(amount, 0, 0, 0, ltype)
	case data.COIN:
		rs.addCurrency(0, amount, 0, 0, ltype)
	case data.CARD:
		rs.addCurrency(0, 0, amount, 0, ltype)
	case data.CHIP:
		rs.addCurrency(0, 0, 0, amount, ltype)
	}
}

//消耗钻石
func (rs *RoleActor) expend(cost uint32, ltype int32) {
	diamond := -1 * int64(cost)
	rs.addCurrency(diamond, 0, 0, 0, ltype)
}
*/

// 后台直接修改货币
func (rs *RoleActor) modifyCurrency(diamond, give int64, ltype int32) {
	user := rs.User
	var realDiamond int64 = 0
	var realGive int64 = 0
	if diamond >= 0 {
		realDiamond = diamond - user.Diamond
	}
	if give >= 0 {
		realGive = give - user.VBBank
	}
	rs.sendGood(realDiamond, 0, realGive, 0, 0, 0, ltype, "后台修改货币")

	// rs.User.AddCurrency(realDiamond, 0, realGive, 0, realDiamond, 0)
	// //货币变更及时同步
	// msg2 := handler.ChangeCurrencyMsg(realDiamond, realCoin,
	// 	0, 0, realDiamond, 0, ltype, rs.Userid, "后台修改货币", "")
	// rs.rolePid.Tell(msg2)
	// //消息
	// msg := handler.PushCurrencyMsg(realDiamond, realCoin,
	// 	0, 0, rs.OutDiamond, ltype)

	// msg.Data.Coin2 = rs.User.GetCoin()
	// msg.Data.Diamond2 = rs.User.GetDiamond()

	// rs.Send(msg)
}

// 外接修改货币
// func (rs *RoleActor) externalModifyCurrency(diamond int64, ltype int32) {
// 	rs.User.AddCurrency(diamond, 0, 0, 0, diamond, 0)
// 	//货币变更及时同步
// 	msg2 := handler.ChangeCurrencyMsg(diamond, 0,
// 		0, 0, diamond, 0, ltype, rs.Userid, "外接修改货币", "")
// 	rs.rolePid.Tell(msg2)
// 	//消息
// 	msg := handler.PushCurrencyMsg(diamond, 0,
// 		0, 0, rs.OutDiamond, ltype)

// 	msg.Data.Coin2 = rs.User.GetCoin()
// 	msg.Data.Diamond2 = rs.User.GetDiamond()

// 	rs.Send(msg)
// }

// 奖励发放
func (rs *RoleActor) addCurrency(diamond, coin, card, chip, give, out, priv int64, ltype int32, desc, water string) {
	rs.addCurrency1(diamond, coin, give, out, priv, ltype, desc, water, false)
}

func (rs *RoleActor) addCurrency1(diamond, coin, give, out, priv int64, ltype int32, desc, water string, control bool) {
	if rs.User == nil {
		glog.Errorf("add currency user err: %d", ltype)
		return
	}
	//日志记录
	if diamond < 0 && ((rs.User.GetDiamond() + diamond) < 0) {
		diamond = 0 - rs.User.GetDiamond()
	}
	// if chip < 0 && ((rs.User.GetChip() + chip) < 0) {
	// chip = 0 - rs.User.GetChip()
	// }
	if coin < 0 && ((rs.User.GetCoin() + coin) < 0) {
		coin = 0 - rs.User.GetCoin()
	}
	// if card < 0 && ((rs.User.GetCard() + card) < 0) {
	// card = 0 - rs.User.GetCard()
	// }
	rs.User.AddCurrency(diamond, coin, 0, 0, give, out)
	rs.User.AddPrivDiamond(priv) // 私人房输赢记录
	//货币变更及时同步
	msg2 := handler.ChangeCurrencyMsg1(diamond, coin,
		0, 0, give, out, ltype, rs.User.GetUserid(), desc, water, control)
	msg2.Priv = priv
	rs.rolePid.Tell(msg2)
	//消息
	msg := handler.PushCurrencyMsg(diamond, coin,
		0, give, rs.OutDiamond, ltype)
	msg.Data.Coin2 = rs.User.GetCoin()
	msg.Data.Diamond2 = rs.User.GetDiamond()
	rs.Send(msg)

	if give != 0 {
		// VB记录日志
		vb := data.VipBankLog{Date: utils.BsonNow().Unix(), LogType: ltype, Amount: give}
		rs.VBankLog = append(rs.VBankLog, vb)

		size := len(rs.VBankLog)
		if size > 100 {
			rs.VBankLog = rs.VBankLog[size-100:]
		}
		// 弹窗
		bean := &pb.VipBankLog{
			Date:   vb.Date,
			Ltype:  handler.GetBonusChangeType(vb.LogType, rs.VBBank),
			Amount: int32(vb.Amount),
			Vb:     rs.VBBank,
		}
		ntf := &pb.VipBankLogAddNtf{Log: bean}
		rs.Send(ntf)
	}
}

// 同步数据
func (rs *RoleActor) syncUser() {
	if rs.User == nil {
		return
	}
	if rs.rolePid == nil {
		return
	}
	if !rs.status { //有变更才同步
		return
	}
	rs.status = false
	msg := new(pb.SyncUser)
	msg.Userid = rs.User.GetUserid()
	// glog.Debugf("syscUser %#v", rs.User)
	result, err := json.Marshal(rs.User)
	if err != nil {
		glog.Errorf("user %s Marshal err %v", rs.User.GetUserid(), err)
		return
	}
	msg.Data = result
	rs.rolePid.Tell(msg)
}

//'银行

// 银行发放
func (rs *RoleActor) addBank(coin int64, ltype int32, from string) {
	// if rs.User == nil {
	// 	glog.Errorf("add addBank user err: %d", ltype)
	// 	return
	// }
	// //日志记录
	// if coin < 0 && ((rs.User.GetBank() + coin) < 0) {
	// 	coin = 0 - rs.User.GetBank()
	// }
	// rs.User.AddBank(coin)
	// //银行变动及时同步
	// msg2 := handler.BankChangeMsg(coin,
	// 	ltype, rs.User.GetUserid(), from)
	// rs.rolePid.Tell(msg2)
}

// 1存入,2取出,3赠送
// func (rs *RoleActor) bank(arg *pb.BankReq) {
// 	msg := new(pb.BankRsp)
// 	rtype := arg.GetRtype()
// 	amount := int64(arg.GetAmount())
// 	userid := arg.GetUserid()
// 	coin := rs.User.GetCoin()
// 	switch rtype {
// 	case pb.BankDeposit: //存入
// 		if (coin - amount) < data.BANKRUPT {
// 			msg.Error = pb.NotEnoughCoin
// 		} else if amount <= 0 {
// 			msg.Error = pb.DepositNumberError
// 		} else {
// 			rs.addCurrency(0, -1*amount, 0, 0, int32(pb.LOG_TYPE12), "银行存入", "")
// 			rs.addBank(amount, int32(pb.LOG_TYPE12), "")
// 		}
// 	case pb.BankDraw: //取出
// 		if amount < data.DRAW_MONEY_LOW {
// 			msg.Error = pb.DrawMoneyNumberError
// 		} else {
// 			rs.addCurrency(0, amount, 0, 0, int32(pb.LOG_TYPE13), "银行取出", "")
// 			rs.addBank(-1*amount, int32(pb.LOG_TYPE13), "")
// 		}
// 	case pb.BankGift: //赠送
// 		/* if rs.User.BankPhone == "" {
// 			msg.Error = pb.BankNotOpen
// 		} else if arg.GetPassword() != rs.User.BankPassword {
// 			msg.Error = pb.PwdError
// 			//} else if amount > rs.User.GetBank() {
// 		} else */if amount > rs.User.GetCoin() { //修改成赠送bank外面的
// 			msg.Error = pb.NotEnoughCoin
// 		} else if amount < data.DRAW_MONEY {
// 			msg.Error = pb.GiveNumberError
// 		} else if userid == "" {
// 			msg.Error = pb.GiveUseridError
// 		} else {
// 			msg1 := handler.GiveBankMsg(amount, int32(pb.LOG_TYPE15), userid, rs.User.GetUserid())
// 			if rs.bank2give(msg1) {
// 				//rs.addBank(-1*amount, int32(pb.LOG_TYPE15), userid)
// 				rs.addCurrency(0, -1*amount, 0, 0, int32(pb.LOG_TYPE15), "银行赠送", "")
// 				//充值消息提醒
// 				record1, msg1 := handler.GiveNotice(amount, rs.User.GetUserid(), userid)
// 				if record1 != nil {
// 					myactor.Logger().Tell(record1)
// 				}
// 				rs.Send(msg1)
// 			} else {
// 				msg.Error = pb.GiveUseridError
// 			}
// 		}
// 	case pb.BankOpen: //开通
// 		/* if rs.User.BankPhone != "" {
// 			msg.Error = pb.BankAlreadyOpen
// 		} else  */if !utils.PhoneValidate(arg.GetPhone()) {
// 			msg.Error = pb.PhoneNumberError
// 		} else if len(arg.GetPassword()) != 32 {
// 			msg.Error = pb.PwdError
// 		} else if len(arg.GetSmscode()) != 6 {
// 			msg.Error = pb.SmsCodeWrong
// 		} else {
// 			msg.Error = rs.bankCheck(arg)
// 			if msg.Error == pb.OK {
// 				//奖励发放
// 				rs.addCurrency(0, 666, 0, 0, int32(pb.LOG_TYPE56), "银行开通", "")
// 				//消息提醒
// 				record, msg2 := handler.BankOpenNotice(666, rs.User.GetUserid())
// 				if record != nil {
// 					myactor.Logger().Tell(record)
// 				}
// 				if msg2 != nil {
// 					rs.Send(msg2)
// 				}
// 			}
// 		}
// 	case pb.BankResetPwd: //重置密码
// 		/* if rs.User.BankPhone == "" {
// 			msg.Error = pb.BankNotOpen
// 		} else if rs.User.BankPhone != arg.GetPhone() {
// 			msg.Error = pb.PhoneNumberError
// 		} else  */if len(arg.GetPassword()) != 32 {
// 			msg.Error = pb.PwdError
// 		} else if len(arg.GetSmscode()) != 6 {
// 			msg.Error = pb.SmsCodeWrong
// 		} else {
// 			msg.Error = rs.bankCheck(arg)
// 		}
// 	}
// 	msg.Rtype = rtype
// 	msg.Amount = arg.GetAmount()
// 	msg.Userid = userid
// 	// msg.Balance = rs.User.GetBank()
// 	rs.Send(msg)
// }

// 银行赠送
func (rs *RoleActor) bank2give(msg1 interface{}) bool {
	timeout := 3 * time.Second
	res1, err1 := rs.rolePid.RequestFuture(msg1, timeout).Result()
	if err1 != nil {
		glog.Errorf("bank give failed: %v", err1)
		return false
	}
	if response1, ok := res1.(*pb.BankGiven); ok {
		if response1.Error == pb.OK {
			return true
		}
		glog.Errorf("BankGiven err %#v", response1)
		return false
	}
	return false
}

// 银行重置密码, 银行开通
// func (rs *RoleActor) bankCheck(arg *pb.BankReq) pb.ErrCode {
// 	msg1 := &pb.BankCheck{
// 		Userid:   rs.User.GetUserid(),
// 		Phone:    arg.GetPhone(),
// 		Password: arg.GetPassword(),
// 		Smscode:  arg.GetSmscode(),
// 	}
// 	timeout := 3 * time.Second
// 	res1, err1 := rs.rolePid.RequestFuture(msg1, timeout).Result()
// 	if err1 != nil {
// 		glog.Errorf("bank check failed: %v", err1)
// 		return pb.OperateError
// 	}
// 	if response1, ok := res1.(*pb.BankChecked); ok {
// 		if response1.Error == pb.OK {
// 			// rs.User.BankPhone = arg.GetPhone()
// 			// rs.User.BankPassword = arg.GetPassword()
// 			return response1.Error
// 		}
// 		glog.Errorf("bankCheck err %#v", response1)
// 		return response1.Error
// 	}
// 	return pb.OperateError
// }

//.

//'任务

// 任务信息,TODO next任务不显示和重置当日任务
func (rs *RoleActor) task() {
	/* rs.taskInit()
	msg := new(pb.TaskRsp)
	list := config.GetOrderTasks()
	m := make(map[int32]bool)
	for _, v := range list {
		if val, ok := rs.User.Task[utils.String(v.Type)]; ok {
			if val.Prize {
				continue
			}
			if val.Taskid != v.Taskid {
				continue
			}
		}
		if _, ok := m[v.Type]; ok {
			continue
		}
		msg2 := &pb.Task{
			Taskid:  v.Taskid,
			Type:    v.Type,
			Name:    v.Name,
			Count:   v.Count,
			Coin:    v.Coin,
			Diamond: v.Diamond,
		}
		if val, ok := rs.User.Task[utils.String(v.Type)]; ok {
			msg2.Num = val.Num
		}
		m[v.Type] = true
		msg.List = append(msg.List, msg2)
	}
	rs.Send(msg) */
}

// 任务奖励领取
func (rs *RoleActor) taskPrize(taskType int32) {
	/* rs.taskInit()
	glog.Debugf("task prize type %d, task %#v", taskType, rs.User.Task)
	msg := new(pb.TaskPrizeRsp)
	if val, ok := rs.User.Task[utils.String(taskType)]; ok {
		task := config.GetTask(val.Taskid)
		if val.Num < task.Count || task.Taskid != val.Taskid {
			msg.Error = pb.AwardFaild
			rs.Send(msg)
			glog.Errorf("task prize err %d, val %#v", taskType, val)
			return
		}
		//奖励发放
		rs.addCurrency(task.Diamond, task.Coin,
			0, 0, int32(pb.LOG_TYPE46))
		//消息提醒
		record2, msg2 := handler.TaskNotice(task.Coin, task.Name, rs.User.GetUserid())
		if record2 != nil {
			myactor.Logger().Tell(record2)
		}
		if msg2 != nil {
			rs.Send(msg2)
		}
		val.Prize = true
		rs.User.Task[utils.String(taskType)] = val
		//响应消息
		msg.Type = taskType
		msg.Coin = task.Coin
		msg.Diamond = task.Diamond
		//添加新任务
		rs.nextTask(taskType, task.Nextid, msg)
		//日志记录
		record := &pb.LogTask{
			Userid: rs.User.GetUserid(),
			Taskid: val.Taskid,
			Type:   taskType,
		}
		myactor.Logger().Tell(record)
	} else {
		msg.Error = pb.AwardFaild
		glog.Errorf("task prize err type %d", taskType)
	}
	rs.Send(msg) */
}

// func (rs *RoleActor) nextTask(taskType, nextid int32, msg *pb.TaskPrizeRsp) {
// 	/* rs.taskInit()
// 	//TODO 任务完成日志
// 	msg2 := handler.TaskUpdateMsg(0, pb.TaskType(taskType),
// 		rs.User.GetUserid())
// 	msg2.Prize = true //移除标识
// 	msg2.Nextid = nextid
// 	rs.rolePid.Tell(msg2)
// 	if nextid == 0 {
// 		return
// 	}
// 	//存在下个任务
// 	delete(rs.User.Task, utils.String(taskType)) //移除
// 	//查找
// 	task := config.GetTask(nextid)
// 	if task.Taskid != nextid {
// 		return
// 	}
// 	msg.Next = &pb.Task{
// 		Taskid:  task.Taskid,
// 		Type:    task.Type,
// 		Name:    task.Name,
// 		Count:   task.Count,
// 		Coin:    task.Coin,
// 		Diamond: task.Diamond,
// 	}
// 	//添加新任务
// 	taskInfo := data.TaskInfo{
// 		Taskid: task.Taskid,
// 		Utime:  time.Now(),
// 	}
// 	rs.User.Task[utils.String(task.Type)] = taskInfo
// 	msg3 := handler.TaskUpdateMsg(0, pb.TaskType(task.Type),
// 		rs.User.GetUserid())
// 	msg3.Taskid = task.Taskid
// 	rs.rolePid.Tell(msg3) */
// }

// 更新任务数据
func (rs *RoleActor) taskUpdate(arg *pb.TaskUpdate) {
	/* rs.taskInit()
	taskTypeStr := utils.String(int32(arg.Type))
	if val, ok := rs.User.Task[taskTypeStr]; ok {
		if val.Prize {
			return
		}
		//数值超出不再更新
		task := config.GetTask(val.Taskid)
		if val.Num >= task.Count {
			return
		}
		val.Num += arg.Num
		val.Utime = time.Now()
		rs.User.Task[taskTypeStr] = val
		rs.rolePid.Tell(arg)
	} else {
		list := config.GetOrderTasks()
		for _, v := range list {
			if v.Type != int32(arg.Type) {
				continue
			}
			taskInfo := data.TaskInfo{
				Taskid: v.Taskid,
				Num:    arg.Num,
				Utime:  time.Now(),
			}
			rs.User.Task[taskTypeStr] = taskInfo
			rs.rolePid.Tell(arg)
			break
		}
	} */
}

func (rs *RoleActor) taskInit() {
	/* if rs.User.Task == nil {
		rs.User.Task = make(map[string]data.TaskInfo)
	} */
}

func (rs *RoleActor) luckyInit() {
	/* if rs.User.Lucky == nil {
		rs.User.Lucky = make(map[string]data.LuckyInfo)
	} */
}

// lucky信息
func (rs *RoleActor) lucky() {
	/* rs.luckyInit()
	msg := new(pb.LuckyRsp)
	list := config.GetLuckys()
	for _, v := range list {
		msg2 := &pb.Lucky{
			Luckyid: v.Luckyid,
			Name:    v.Name,
			Count:   v.Count,
			Coin:    v.Coin,
			Diamond: v.Diamond,
			Gtype:   v.Gtype,
		}
		if val, ok := rs.User.Lucky[utils.String(v.Luckyid)]; ok {
			msg2.Num = val.Num
		}
		msg.List = append(msg.List, msg2)
	}
	rs.Send(msg) */
}

// 更新lucky数据
func (rs *RoleActor) luckyUpdate(arg *pb.LuckyUpdate) {
	/* rs.luckyInit()
	luckyidStr := utils.String(int32(arg.GetLuckyid()))
	lucky := config.GetLucky(arg.GetLuckyid())
	if lucky.Luckyid != arg.GetLuckyid() {
		return
	}
	if lucky.Luckyid == 0 {
		return
	}
	if val, ok := rs.User.Lucky[luckyidStr]; ok {
		//数值超出不再更新
		if val.Num >= lucky.Count {
			//return
		}
		val.Num += arg.Num
		rs.User.Lucky[luckyidStr] = val
		if val.Num == lucky.Count {
			//奖励发放
			//rs.addCurrency(lucky.Diamond, lucky.Coin, 0, 0, int32(pb.LOG_TYPE51))
			//消息提醒
			//record, msg2 := handler.LuckyNotice(lucky.Coin, lucky.Name, arg.Userid)
			//if record != nil {
			//	myactor.Logger().Tell(record)
			//}
			//if msg2 != nil {
			//	rs.Send(msg2)
			//}
		}
	} else {
		luckyInfo := data.LuckyInfo{
			Luckyid: arg.GetLuckyid(),
			Num:     arg.Num,
		}
		rs.User.Lucky[luckyidStr] = luckyInfo
	}
	rs.rolePid.Tell(arg) */
}

//.

//'签到

// 更新连续登录奖励
func (rs *RoleActor) loginPrizeInit() {
	//连续登录
	/* glog.Debugf("userid %s, LoginTime %s", rs.User.GetUserid(),
		utils.Time2Str(rs.User.LoginTime.Local()))
	glog.Debugf("userid %s, LoginTimes %d, LoginPrize %d",
		rs.User.GetUserid(), rs.User.LoginTimes, rs.User.LoginPrize)
	//rs.User.LoginTime = utils.Stamp2Time(utils.TimestampToday() - 10)
	handler.SetLoginPrize(rs.User)
	glog.Debugf("userid %s, LoginTime %s", rs.User.GetUserid(),
		utils.Time2Str(rs.User.LoginTime.Local()))
	glog.Debugf("userid %s, LoginTimes %d, LoginPrize %d",
		rs.User.GetUserid(), rs.User.LoginTimes, rs.User.LoginPrize)
	rs.User.LoginTime = utils.BsonNow().Local()
	msg := handler.LoginPrizeUpdateMsg(rs.User)
	rs.rolePid.Tell(msg) */
}

// 连续登录奖励处理
// func (rs *RoleActor) loginPrize(arg *pb.LoginPrizeReq) {
// 	msg := new(pb.LoginPrizeRsp)
// 	msg.Type = arg.Type
// 	// switch arg.Type {
// 	// case pb.LoginPrizeSelect:
// 	// 	msg.List = handler.LoginPrizeInfo(rs.User)
// 	// case pb.LoginPrizeDraw:
// 	// 	coin, diamond, ok := handler.GetLoginPrize(arg.Day, rs.User)
// 	// 	msg.Error = ok
// 	// 	if ok == pb.OK {
// 	// 		//奖励发放
// 	// 		rs.addCurrency(diamond, coin, 0, 0, 0, int32(pb.LOG_TYPE47), "连续登录奖励", "")
// 	// 		msg.List = handler.LoginPrizeInfo(rs.User)
// 	// 		msg := handler.LoginPrizeUpdateMsg(rs.User)
// 	// 		rs.rolePid.Tell(msg)
// 	// 	}
// 	// }
// 	rs.Send(msg)
// }

//.

// 设置个性签名
func (rs *RoleActor) setSign(arg *pb.SignatureReq) {
	msg := new(pb.SignatureRsp)
	if len(arg.GetContent()) > 1024 {
		msg.Error = pb.SignTooLong
		rs.Send(msg)
		return
	}
	msg.Userid = rs.User.GetUserid()
	msg.Content = arg.GetContent()
	rs.User.SetSign(arg.GetContent())
	arg.Userid = rs.User.GetUserid()
	rs.rolePid.Tell(arg)
	rs.Send(msg)
}

// 设置头像
func (rs *RoleActor) setPhoto(arg *pb.PhotoReq) {
	msg := new(pb.PhotoRsp)
	if len(arg.GetPhoto()) > 1024 {
		msg.Error = pb.PhotoTooLong
		rs.Send(msg)
		return
	}
	msg.Userid = rs.User.GetUserid()
	msg.Photo = arg.GetPhoto()
	rs.User.SetPhoto(arg.GetPhoto())
	arg.Userid = rs.User.GetUserid()
	rs.rolePid.Tell(arg)
	rs.Send(msg)
}

// 设置昵称
func (rs *RoleActor) setNickName(arg *pb.NickNameReq) {
	msg := new(pb.NickNameRsp)
	if arg.Type == pb.ChangeInfoTypeNickName {
		if len(arg.GetName()) > 1024 || !utils.LegalName(arg.GetName(), 7) {
			glog.Errorf("name too long error %s", arg.GetName)
			msg.Error = pb.NickNameTooLong
			rs.Send(msg)
			return
		}
		msg.Userid = rs.User.GetUserid()
		msg.Name = arg.GetName()
		rs.User.SetNickname(arg.GetName())
		arg.Userid = rs.User.GetUserid()
		rs.rolePid.Tell(arg)
		rs.Send(msg)
	} else if arg.Type == pb.ChangeInfoTypeEmail {
		var IsValidEmail = func(email string) bool {
			// 邮箱正则表达式模式
			pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
			// 编译正则表达式
			reg := regexp.MustCompile(pattern)
			// 检查邮箱是否匹配正则表达式
			return reg.MatchString(email)
		}
		if !IsValidEmail(arg.GetName()) {
			msg.Error = pb.EmailFormatError
			rs.Send(msg)
			return
		}
		msg.Userid = rs.User.GetUserid()
		msg.Name = arg.GetName()
		rs.User.SetEmail(arg.GetName())
		arg.Userid = rs.User.GetUserid()
		rs.rolePid.Tell(arg)
		rs.Send(msg)
	} else if arg.Type == pb.ChangeInfoTypeBankName {
		if arg.GetName() == "" {
			msg.Error = pb.CheckBankInfo
			rs.Send(msg)
			return
		}
		msg.Userid = rs.User.GetUserid()
		msg.Name = arg.GetName()
		rs.User.SetBankName(arg.GetName())
		arg.Userid = rs.User.GetUserid()
		rs.rolePid.Tell(arg)
		rs.Send(msg)
	} else if arg.Type == pb.ChangeInfoTypeBankAccountHolder {
		if arg.GetName() == "" {
			msg.Error = pb.CheckBankInfo
			rs.Send(msg)
			return
		}
		msg.Userid = rs.User.GetUserid()
		msg.Name = arg.GetName()
		rs.User.SetBankAccountHolder(arg.GetName())
		arg.Userid = rs.User.GetUserid()
		rs.rolePid.Tell(arg)
		rs.Send(msg)
	} else if arg.Type == pb.ChangeInfoTypeAccountNumber {
		arg.Name = strings.ReplaceAll(arg.GetName(), " ", "")
		if arg.GetName() == "" {
			msg.Error = pb.CheckBankInfo
			rs.Send(msg)
			return
		}
		msg.Userid = rs.User.GetUserid()
		msg.Name = arg.GetName()
		rs.User.SetAccountNumber(arg.GetName())
		arg.Userid = rs.User.GetUserid()
		rs.rolePid.Tell(arg)
		rs.Send(msg)
	} else if arg.Type == pb.ChangeInfoTypeIFSC {
		arg.Name = strings.ReplaceAll(arg.GetName(), " ", "")
		if arg.GetName() == "" {
			msg.Error = pb.CheckBankInfo
			rs.Send(msg)
			return
		}
		msg.Userid = rs.User.GetUserid()
		msg.Name = arg.GetName()
		rs.User.SetIFSC(arg.GetName())
		arg.Userid = rs.User.GetUserid()
		rs.rolePid.Tell(arg)
		rs.Send(msg)
	}
}

// 设置经纬度
// func (rs *RoleActor) setLatLng(arg *pb.LatLngReq) {
// 	msg := new(pb.LatLngRsp)
// 	rs.User.Lat = arg.GetLat()
// 	rs.User.Lng = arg.GetLng()
// 	rs.User.Address = arg.GetAddress()
// 	rs.Send(msg)
// 	arg.Userid = rs.User.GetUserid()
// 	rs.rolePid.Tell(arg)
// }

// // join activity
// func (rs *RoleActor) joinActivity(arg *pb.JoinActivityReq, ctx actor.Context) {
// 	// if handler.IsNotAgent(rs.User) {
// 	// 	rsp := new(pb.JoinActivityRsp)
// 	// 	rsp.Error = pb.NotAgent
// 	// 	rs.Send(rsp)
// 	// 	return
// 	// }
// 	arg.Selfid = rs.User.GetUserid()
// 	act := config.GetActivity(arg.GetActid())
// 	if act.Id != arg.GetActid() {
// 		msg := new(pb.JoinActivityRsp)
// 		msg.Error = pb.ActidError
// 		rs.Send(msg)
// 		return
// 	}
// 	rs.dbmsPid.Request(arg, ctx.Self())
// }

// 计算活动统一领奖
func (rs *RoleActor) getActivityReceive(receive bool) uint32 {
	var bouns uint32 = 0
	// 首充
	/* if rs.User.FirstRecharge != "" && !rs.User.Receive && rs.User.ReceiveDay < 5 {
		actData := config.GetActivity(rs.User.FirstRecharge)
		if actData.Id != "" {
			switch rs.User.ReceiveDay {
			case 1:
				bouns += actData.F_secondDay
			case 2:
				bouns += actData.F_thirdDay
			case 3:
				bouns += actData.F_forthDay
			case 4:
				bouns += actData.F_fifthDay
			}
			if receive {
				rs.User.Receive = true
				rs.User.ReceiveDay++
				rs.status = true
			}
		}
	} */
	return bouns
}

// 领取签到奖励
func (rs *RoleActor) GetDailSignReward(arg *pb.DailySignReceiveReq, user *data.User) (res *pb.DailySignReceiveRsp) {
	res = new(pb.DailySignReceiveRsp)
	day := arg.Day
	if user.LoginTimes < day {
		// 登录天数不够
		res.Code = pb.SignException
		return
	}
	if user.SignDay&(1<<day) != 0 {
		res.Code = pb.AlreadyAward
		return
	}
	user.SignDay |= (1 << day)
	// 发奖励
	act := config.GetDailySign(arg.GetDay())
	if act.Id < 0 {
		res.Code = pb.ActidError
		return
	}
	if act.Price > 0 {
		glog.Error("sign need recharge,user：", rs.Userid)
		res.Code = pb.SignException
		return
	}
	rs.sendGood(int64(act.Number), 0, int64(act.Number), 0, 0, 0, int32(pb.LOG_TYPE61), fmt.Sprintf("签到第%d天", day))
	// switch act.Ctype {
	// case int32(data.DIAMOND):
	// 	// case int32(data.COIN):
	// 	// 	rs.sendGood(0, int64(act.Number), 0, 0, int32(pb.LOG_TYPE61), fmt.Sprintf("签到第%d天", day))
	// }
	res.Number = act.Number
	// 通知dbms
	rs.rolePid.Tell(arg)
	return
}

// 玩家跨天逻辑
func (rs *RoleActor) zeroReset(oldTime int64) {
	//补丁
	old := time.UnixMilli(oldTime).In(location)
	now := time.Now().In(location)
	if utils.Time2DayDate(now) == utils.Time2DayDate(rs.ResetTime) {
		// 今天领过的不再reset
		rs.LastResetTime = time.Now().UnixMilli()
		return
	}

	if utils.Time2DayDate(old) == utils.Time2DayDate(now) {
		return
	}
	if !rs.online {
		// glog.Debugf("user %s was logout", rs.Userid)
		return
	}
	// 不是同一天
	glog.Infof("user %s zeroreset, old:%s, now:%s", rs.User.Userid, old.Format("2006-01-02 15:04:05"), now.Format("2006-01-02 15:04:05"))
	//VB银行利息, 这个需要在vip重置之前
	rs.VBankInterest()
	rs.SkipRank = false
	rs.SkipTurntable = false
	rs.SkipGiftCode = false
	rs.User.FirstLoginToday = true
	rs.User.LoginTimes++
	// rs.User.ResetTime = time.Now().Local()
	rs.LastResetTime = time.Now().UnixMilli()
	rs.User.OnlineTime = 0
	// 金银铜卡
	rs.updateWeekCard()
	rs.firstRechargeReset()
	// 任务
	rs.taskReset()
	// vip bank 任务
	rs.vbTaskReset()
	// vb 输分补偿次数重置
	rs.User.VBOutDiamondDayTimes = 0
	rs.User.OnlineReward = nil
	// 充值破产礼包
	rs.BreakingGift = make([]*data.BreakingGift, 0)
	// vip
	rs.vipReset()
	// 礼包
	rs.giftPopReset()
	//tp
	rs.TpReset()
	// rummy双人
	rs.rm2Reset()
	// playshare
	rs.playShareReset()
	// 罐子
	rs.shopPotReset()
	//重置tp触发次数
	rs.ResetTPTiggerTimes()
	//龙虎重置
	rs.LHDReset()
	//crash重置
	rs.CrashReset()
	//Aviator重置
	rs.AviatorReset()
	//7UP重置
	rs.UPReset()
	// bug反馈次数
	rs.BugFeedback = 0
	// 代理零点重置
	rs.rolePid.Tell(&pb.ShareAgentZeroReset{Userid: rs.Userid})
	//同步游戏
	if rs.gamePid != nil {
		rs.gamePid.Tell(&pb.TPTriggeTimes{Isreset: true})
	}
	//重置105kick
	rs.Kick105Flag = false
	rs.Kick200Flag = false
	//B类必赢策略重置
	rs.FreeWelfare = 0
	rs.Novice200StrategyCount = 0
	//重置转单补贴金
	rs.TransferCash = 0

	rs.status = true
}

// 初始化用户历史数据
func (rs *RoleActor) initHistory() {
	// 查用户充值次数
	if rs.User.HistoryVersion == 0 {
		var chargeTimes int64
		err := ck.Select(&chargeTimes, `
			SELECT COUNT(*) pay_times FROM game.col_trade_record FINAL WHERE userid = ? AND order_status = 4
		`, rs.Userid)
		if err != nil {
			glog.Errorf("init history charge_times error: %s, %v", rs.Userid, err)
		} else {
			rs.User.ChargeTimes = int32(chargeTimes)
			rs.User.HistoryVersion++
		}
	}
}

// 首充跨天
func (rs *RoleActor) firstRechargeReset() {
	user := rs.User
	if user.FirstRecharge == "" || user.ReceiveDay >= 5 {
		return
	}
	user.Receive = false
}

// 周卡跨天更新
func (rs *RoleActor) updateWeekCard() {
	cardMap := rs.User.WeeklyCardMap
	// remove := make([]int32, 0)
	for k, v := range cardMap {
		if k == 0 {
			continue
		}
		cardId := fmt.Sprintf("lbzk%d", k)
		act := table.GetTables().GiftRechargeTable.Get(cardId)
		if act == nil || v.ReceiveTimes >= int32(len(act.DailyGive)) || v.OverTime <= utils.LocalTime().Unix() { // 活动不存在了||领完了
			glog.Infof("user %s card timeover", rs.Userid)
			// remove = append(remove, k)
			v.OverTime = -1
			continue
		}
		v.Get = true
		// if (v.CanGet << (v.ReceiveTimes + 1)) <= (1 << int32(len(act.Give))) {
		// 	// 累计领天数没到上限
		// 	v.CanGet <<= 1
		// }
	}
	// for _, id := range remove {
	// 	delete(rs.User.WeeklyCardMap, id)
	// 	if len(rs.User.WeeklyCardMap) <= 0 {
	// 		rs.User.WeeklyCardMap[0] = new(data.WeekCard)
	// 	}
	// }
}

func (rs *RoleActor) shopPotReset() {
	if rs.ActivatedPot == nil {
		return
	}

	pots := make([]data.ShopPot, 0)
	now := utils.LocalTime().Unix()
	for _, pot := range rs.ActivatedPot {
		if pot.OverTime > now {
			pots = append(pots, pot)
		}
	}

	rs.ActivatedPot = pots
}

func (rs *RoleActor) VBankInterest() {
	if rs.State == data.NoveiceState || rs.State == data.ExceptionState || rs.LastResetTime <= 0 {
		return
	}

	line := rs.LastResetTime % (24 * 60 * 60 * 1000)
	newLine := time.Now().UnixMilli() % (24 * 60 * 60 * 1000)

	startTime := time.UnixMilli(rs.LastResetTime)
	endTime := time.Now()

	duration := endTime.Sub(startTime)
	day := int32(duration.Seconds()) / (24 * 60 * 60)
	if line > newLine {
		// 多一天
		day++
	}

	vipbean := table.GetTables().VipTable.Get(int32(rs.Vip.Lv))
	if vipbean == nil {
		return
	}

	if day > vipbean.Reset {
		day = vipbean.Reset
	}

	conf := table.GetTables().VipInterestTable.Get(int32(rs.Vip.Lv))

	var base float64 = 10000
	for i := 0; i < int(day); i++ {
		base = base * (10000 + conf.InterestRate[rs.RegistArea]) / 10000
	}

	interest := rs.OutDiamond * int64((base-10000)*100) / 10000 / 100
	if interest > 0 {
		rs.sendGood(0, 0, interest, 0, 0, 0, int32(pb.LOG_TYPE121), fmt.Sprintf("%.2f本金利息增加%.2fVB金", float64(rs.OutDiamond)/100, float64(interest)/100))
	}
	// if conf.Mod == 2 && interest > 0 && interest <= rs.VBBank {
	// 	// 释放
	// 	rs.sendGood(interest, 0, -interest, 0, 0, 0, int32(pb.LOG_TYPE122), fmt.Sprintf("%.2f本金利息释放%.2fVB金", float64(rs.OutDiamond)/100, float64(interest)/100))
	// }
	// if conf.Mod == 1 && interest > 0 {
	// 	// 增长
	// 	rs.sendGood(0, 0, interest, 0, 0, 0, int32(pb.LOG_TYPE121), fmt.Sprintf("%.2f本金利息增加%.2fVB金", float64(rs.OutDiamond)/100, float64(interest)/100))
	// }
}

// 绑定提现手机号
func (rs *RoleActor) bindPhone2(arg *pb.BindPhoneReq, ctx actor.Context) (res *pb.BindPhoneRsp) {
	res = new(pb.BindPhoneRsp)
	var ok bool
	rsp, err := ctx.RequestFuture(rs.rolePid, arg, 3*time.Second).Result()
	if err != nil {
		res.Error = pb.BindPhoneFail
		return
	}
	if res, ok = rsp.(*pb.BindPhoneRsp); ok && res.Error == pb.OK {
		res.Phone = arg.GetPhone()
		rs.User.Phone2 = arg.Phone
		rs.User.RealName = arg.RealName
		rs.status = true
	}
	return
}

// 首充领奖
// func (rs *RoleActor) getFirstRechargeReward(arg *pb.FirstRechargeRewardReq) {
// 	rsp := new(pb.FirstRechargeRewardRsp)
// 	var bouns uint32 = 0
// 	if rs.User.FirstRecharge != "" && !rs.User.Receive && rs.User.ReceiveDay < 5 {
// 		actData := config.GetActivity(rs.User.FirstRecharge)
// 		if actData.Id != "" {
// 			switch rs.User.ReceiveDay {
// 			case 1:
// 				bouns += actData.F_secondDay
// 			case 2:
// 				bouns += actData.F_thirdDay
// 			case 3:
// 				bouns += actData.F_forthDay
// 			case 4:
// 				bouns += actData.F_fifthDay
// 			}
// 			rs.User.Receive = true
// 			rs.User.ReceiveDay++
// 			rs.sendGood(int64(bouns), 0, 0, 0, 0, 0, int32(pb.LOG_TYPE57), fmt.Sprintf("首充第%d天", rs.User.ReceiveDay))
// 			// 更新活动数据
// 			rsp.Id = actData.Id
// 			rsp.UserData = &pb.FisrtRecharge{
// 				Recharge: rs.User.FirstRecharge != "",
// 				Receive:  rs.User.Receive,
// 				Day:      rs.User.ReceiveDay,
// 			}
// 			rs.Send(rsp)
// 			// 同步数据
// 			msg := &pb.FirstRecharged{
// 				Userid:  rs.Userid,
// 				Receive: rs.User.Receive,
// 				Day:     rs.User.ReceiveDay,
// 			}
// 			rs.rolePid.Tell(msg)
// 		}
// 		glog.Errorf("activity missing, user:%s, id:%s", rs.Userid, rs.User.FirstRecharge)
// 	}
// }

// 领取金银铜卡每日奖励
func (rs *RoleActor) getWeekCardDaily(arg *pb.WeeklyCardReceiveReq) {
	rsp := new(pb.WeeklyCardReceiveRsp)
	card := rs.User.WeeklyCardMap[arg.Id]
	if card == nil {
		rsp.Code = pb.NoBuyWeekCard
		rs.Send(rsp)
		return
	}
	if !card.Get {
		rsp.Code = pb.AlreadyAward
		rs.Send(rsp)
		return
	}
	if utils.LocalTime().Unix() >= card.OverTime {
		// 过期了
		rsp.Code = pb.WeeklyCardOverTime
		rs.Send(rsp)
		return
	}
	cardId := fmt.Sprintf("lbzk%d", arg.Id)
	act := table.GetTables().GiftRechargeTable.Get(cardId)
	if act == nil {
		rsp.Code = pb.ActidError
		rs.Send(rsp)
		return
	}
	if card.ReceiveTimes >= int32(len(act.DailyGive)) {
		rsp.Code = pb.NoBuyWeekCard
		rs.Send(rsp)
		return
	}
	give, cash, withdrawable := rs.getWeeklyReward(arg.Id)
	glog.Infof("receive %d card daily reward bouns:%d", act.Id, cash)
	rs.sendGood(cash, 0, give, withdrawable, 0, 0, int32(pb.LOG_TYPE59), fmt.Sprintf("周卡%d每日", arg.Id)) // desc的每日不要删,后台经济日报统计!!!
	card.Get = false
	card.ReceiveTimes++
	rs.status = true
	// 更新
	rsp.Card = config.BuildWeekCardData(act.Id, rs.User, location)
	rs.Send(rsp)
	// 同步
	rs.rolePid.Tell(arg)
}

// 在线奖励数据
func (rs *RoleActor) onlineRewardData(arg *pb.OnlineRewardDataReq) {
	msg := new(pb.OnlineRewardDataRsp)
	// onlineMap := config.GetOnlineReward2()
	onlineMap := table.GetTables().OnlineRewardTable.GetDataList()
	for _, v := range onlineMap {
		build := &pb.OnlineReward{
			Id:         v.Id,
			DuringTime: int32(v.OnlineTime),
			// OnlineTime:  v.OnlineTime,
			// ReceiveTime: utils.BsonNow().Unix() + (v.OnlineTime - int64(rs.User.OnlineTime)),
		}
		for _, r := range v.Rewards {
			if len(r.Nums) < 2 {
				glog.Errorf("online reward length error id:%d", v.Id)
				continue
			}
			item := &pb.Item{
				Itype:  r.Nums[0],
				Number: int64(r.Nums[1]),
			}
			build.Items = append(build.Items, item)
		}
		msg.Onlines = append(msg.Onlines, build)
	}
	sort.Slice(msg.Onlines, func(i, j int) bool {
		return msg.Onlines[i].Id < msg.Onlines[j].Id
	})
	rs.Send(msg)
}

func (rs *RoleActor) getOnlineTime() {
	rsp := &pb.OnlineRewardTimeRsp{
		OnlineTime: int64(rs.OnlineTime),
	}
	rs.Send(rsp)
}

// 在线奖励领取
func (rs *RoleActor) getOnlineReward() {
	rsp := new(pb.OnlineRewardReceiveRsp)
	var id int32 = 1
	// 能不能领
	if len(rs.User.OnlineReward) > 0 {
		id = rs.User.OnlineReward[len(rs.User.OnlineReward)-1] + 1
	}
	// onlineData := config.GetOnlineReward(id)
	onlineData := table.GetTables().OnlineRewardTable.Get(int(id))
	if onlineData.Id == 0 || onlineData.OnlineTime > int32(rs.User.OnlineTime) {
		rsp.Error = pb.OnlineIndexFail
		rs.Send(rsp)
		return
	}
	glog.Infof("getOnlineReward user:%s ,time:%d", rs.Userid, rs.User.OnlineTime)
	// 随机一个奖励
	rewards := make([][]int32, 0)
	for _, v := range onlineData.Rewards {
		rewards = append(rewards, []int32{int32(v.Nums[0]), int32(v.Nums[1]), int32(v.Nums[2])})
	}
	index, err := handler.RandomIndex(rewards)
	if err != nil {
		rsp.Error = pb.OnlineIndexFail
		rs.Send(rsp)
		return
	}
	// 发送奖励
	reward := onlineData.Rewards[index]
	// rs.sendGoodByType(uint32(reward[0]), int64(reward[1]), 0, int32(pb.LOG_TYPE62), "在线奖励")
	rs.sendGood(0, 0, int64(reward.Nums[1]), 0, 0, 0, int32(pb.LOG_TYPE62), "在线奖励")
	rs.User.OnlineReward = append(rs.User.OnlineReward, id)
	rsp.Index = index
	rsp.Id = id
	rs.Send(rsp)
	// 通知dbms
	msg := new(pb.OnlineRewardRecord)
	msg.UserId = rs.Userid
	msg.Id = id
	msg.Otime = int64(rs.User.OnlineTime)
	rs.rolePid.Tell(msg)
}

func (rs *RoleActor) resWithDrawData() (rsp *pb.GetWithDrawDataRsp) {
	rsp = new(pb.GetWithDrawDataRsp)
	// shop := config.GetShop()
	shop := table.GetTables().WithdrawConfigTable.Get()
	if shop == nil {
		return
	}

	// get payChannels
	payChannels := getPayChannelShops(false)
	AppsBanks := make(map[string]struct{})
	AppsWallets := make(map[string]struct{})
	// 渠道最大最小提现额
	var withdrawMin, withdrawMax int64
	for _, c := range payChannels {
		if !c.Withdrawable {
			continue
		}
		if withdrawMin == 0 || withdrawMin > c.WithdrawMin {
			withdrawMin = c.WithdrawMin
		}
		if withdrawMax == 0 || withdrawMax < c.WithdrawMax {
			withdrawMax = c.WithdrawMax
		}

		// 提现app配置 巴基斯坦
		for _, bank := range c.WithdrawBanks {
			AppsBanks[bank] = struct{}{}
		}
		for _, w := range c.WithdrawWallets {
			AppsWallets[w] = struct{}{}
		}
	}

	rsp.PkData = rs.BuildPKWithdrawData(AppsBanks, AppsWallets)

	// 配置表最大最小提现额
	if withdrawMin < int64(shop.WithdrawInterval[0]) {
		withdrawMin = int64(shop.WithdrawInterval[0])
	}
	if withdrawMax > int64(shop.WithdrawInterval[1]) {
		withdrawMax = int64(shop.WithdrawInterval[1])
	}
	rsp.MinWithdraw = int32(withdrawMin)
	rsp.MaxWithdraw = int32(withdrawMax)

	for i, w := range shop.DefaultWithdraw {
		bean := &pb.WithDrawData{
			Id:          int32(i),
			CostDiamond: uint32(w),
			GetCash:     uint32(w),
			Index:       int32(i),
			// Titile:        w.Name,
			// Limit:         w.Limit,
			// RechargeLimit: w.RechargeLimit,
			// FlowingLimit:  w.FlowingLimit,
			// Commission:    w.Commission,
			// Count:         countMap[w.Id],
		}
		// C类用户
		// if rs.RegistArea == 2 {
		// 	// 不显示105
		// 	if w.CostDiamond == 10500 {
		// 		continue
		// 	}
		// 	// 手续费10%
		// 	bean.GetCash = w.CostDiamond * 90 / 100
		// 	bean.Commission = w.CostDiamond * 10 / 100
		// }

		cash := int64(bean.GetCash)
		if cash >= withdrawMin && cash <= withdrawMax {
			rsp.Data = append(rsp.Data, bean)
		}
	}
	return
}

// 系统功能开关
func (rs *RoleActor) systemSwitch() {
	ntf := new(pb.FunctionStateNtf)
	switchMap := config.GetSettingMap()
	for _, v := range switchMap {
		if v.Stype == 0 {
			// 游戏
			// bean := &pb.GameSwitch{
			// 	Open:  v.Status == 1,
			// 	Wight: int32(v.SortId),
			// }
			// switch v.Name {
			// case "TeenPatti经典":
			// 	bean.Gtype = pb.HUA
			// case "TeenPattiAK47":
			// 	bean.Gtype = pb.AK47
			// case "TeenPattiJoker":
			// 	bean.Gtype = pb.JOKER
			// case "Rummy":
			// 	bean.Gtype = pb.RUMMY
			// case "7UPDOWN":
			// 	bean.Gtype = pb.SEVEN
			// case "DRAGON TIGER":
			// 	bean.Gtype = pb.LHD
			// case "CRASH":
			// 	bean.Gtype = pb.CRASH
			// case "ANDARBAHAR":
			// 	bean.Gtype = pb.ABAR
			// case "彩票":
			// 	bean.Gtype = pb.LOTTERY
			// case "飞机":
			// 	bean.Gtype = pb.PLANE
			// case "Rummy双人":
			// 	bean.Gtype = pb.RUMMY2
			// case "红黑大战":
			// 	bean.Gtype = pb.REDBLACK
			// }
			// ntf.Game = append(ntf.Game, bean)
		} else if v.Stype == 1 {
			// 活动功能
			bean := &pb.FunctionSwitch{
				Open:      v.Status == 1,
				Wight:     int32(v.SortId),
				Condition: int32(v.PayStatus),
			}
			switch v.Name {
			case "首充":
				bean.Gtype = pb.FirstRecharge
			case "入门礼包":
				bean.Gtype = pb.NoviceGift
			case "在线奖励":
				bean.Gtype = pb.OnlineRewrad
			case "每日签到":
				bean.Gtype = pb.DailySign
			case "任务活动":
				bean.Gtype = pb.DailyTask
			case "牌型活动":
				bean.Gtype = pb.PokerHands
			case "金银铜卡":
				bean.Gtype = pb.MetalCard
			case "分享活动":
				bean.Gtype = pb.Share
			case "包赔活动":
				bean.Gtype = pb.Compensation
			case "BUG有奖":
				bean.Gtype = pb.BugFeedback
			case "限时活动":
				bean.Gtype = pb.LimitedGiftAct
			case "Slot开关":
				bean.Gtype = pb.Slot
			case "对战房开关":
				bean.Gtype = pb.PVPRoom
			case "商城罐子":
				bean.Gtype = pb.ShopJar
			case "VIP":
				bean.Gtype = pb.VIPFUNC
			case "真人视讯":
				bean.Gtype = pb.LiveVideo
			case "排行榜开关":
				bean.Gtype = pb.RANK
			case "小米手机活动开关":
				bean.Gtype = pb.LuckyDraw
			case "大富翁":
				bean.Gtype = pb.Monopoly
			case "拼多多":
				bean.Gtype = pb.PDD
			case "社媒关注奖金活动":
				bean.Gtype = pb.FocesOnBouns
			case "VB活动任务":
				bean.Gtype = pb.VBGameTask
			case "利息获取":
				bean.Gtype = pb.VBInterestFunc
			case "输分补偿":
				bean.Gtype = pb.VBCompensation
			}
			ntf.Func = append(ntf.Func, bean)
		}
	}
	rs.Send(ntf)
}

// 客服消息
func (rs *RoleActor) feedBackChats() {
	user := rs.User
	if user.FeedBackLogMap == nil {
		user.FeedBackLogMap = make(map[string]data.ChatLog)
	}
	if len(user.FeedBackLogMap) <= 0 {
		return
	}
	ntf := new(pb.FeedBackLogNtf)
	logs := make([]*pb.ChatLog, 0)
	for k, cl := range user.FeedBackLogMap {
		bean := &pb.ChatLog{
			Uid:   k,
			Title: cl.Title,
			Ctime: cl.Ctime.Unix(),
			Read:  cl.Read,
		}
		logs = append(logs, bean)
	}
	sort.Slice(logs, func(i, j int) bool { return logs[i].Ctime >= logs[j].Ctime })
	ntf.Logs = logs
	rs.Send(ntf)
}

// 向客服发送消息
func (rs *RoleActor) feedBackReq(arg *pb.FeedBackReq) {
	rsp := new(pb.FeedBackRsp)
	rsp.UserId = rs.Userid
	user := rs.User
	// 内容检测
	if !handler.CheckContent(arg.Content) {
		rsp.Error = pb.FeedBackFail
		rs.Send(rsp)
		return
	}
	// 发送次数
	times := user.FeedBackTimes
	if times >= 10 {
		rsp.Error = pb.FeedBackLimited
		rs.Send(rsp)
		return
	}
	if arg.ReqId == 0 {
		user.FeedBackTimes++
	}
	rs.Send(rsp)
	rs.status = true
	// 保存db
	arg.Name = user.Nickname
	myactor.Logger().Tell(arg)
}

// 读客服消息
func (rs *RoleActor) readFeedBackLogReq(arg *pb.ReadFeedBackLogReq) {
	uid := arg.Uid
	user := rs.User
	rsp := new(pb.ReadFeedBackLogRsp)
	if log, ok := user.FeedBackLogMap[uid]; ok {
		log.Read = true
		user.FeedBackLogMap[uid] = log
		rs.status = true
		rsp.Content = log.Content
		rsp.Uid = uid
		rs.Send(rsp)
		return
	}
	rsp.Error = pb.FeedBackUid
	rs.Send(rsp)
}

// 客服回复消息
func (rs *RoleActor) feedBackReply(arg *pb.FeedBackReply) {
	user := rs.User

	if !rs.online {
		user.FeedBackTimes = 0
		chat := data.OfflineChatLog{
			Content: arg.Log.Content,
			Ctime:   utils.LocalTime().UnixMilli(),
			Sender:  "-1",
			Name:    "system",
		}
		user.OfflineChats = append(user.OfflineChats, chat)
		rs.status = true
		return
	}

	ntf := new(pb.ConsumerLogNtf)
	log := &pb.ConsumerLog{
		Userid:  arg.Log.Userid,
		Content: arg.Log.Content,
		Ctime:   arg.Log.Ctime,
	}
	ntf.Chats = append(ntf.Chats, log)
	rs.Send(ntf)

	user.FeedBackTimes = 0
	rs.status = true
}

// 活动初始化数据
func (rs *RoleActor) sendActivityData() {
	ntf := &pb.ActivityDataNtf{
		OnlineData:  handler.BuildOnlineRewadMsg(),
		NewbiewGift: handler.BuildRechargeActMsg(rs.User),
		Share:       handler.BuildShareDataMsg(rs.User),
		// Task:               handler.BuildTaskDataMsg(rs.User),
		PokerHands: handler.BuildPokerHandsDataMsg(rs.User),
		// LimitGift:          handler.BuildLimitedGiftDataMsg(rs.User),
		WeekCard:           handler.BuildWeeklyCardDataMsg(rs.User, location),
		Sign:               handler.BuildDailySignDataMsg(rs.User),
		ShopPots:           handler.BuildShopPotDataMsg(rs.User),
		OnlineTime:         int64(rs.OnlineTime),
		PotsFlow:           rs.PotFlow,
		Scratch:            rs.buildScratchTicketDataMsg(),
		PlayShare:          handler.BuildPlayShareDataMsg(rs.User),
		VbTask:             handler.BuildVBTaskDataMsg(rs.User),         // vip bank 任务数据
		VbLoseCompensation: handler.BuildVBLoseCompensationMsg(rs.User), // vb输分补偿
	}
	rs.Send(ntf)
}

// 大富翁
func (rs *RoleActor) buildScratchTicketDataMsg() (bean *pb.ScratchTicket) {
	// if rs.Money <= 0 {
	// 	return
	// }
	bean = new(pb.ScratchTicket)
	tickets := config.GetScratchTicketConfig()
	ticket := rs.ScratchTicket // 玩家身上的刮刮乐数据
	for _, t := range tickets {
		info := &pb.ScratchCell{
			Cost:   t.Cost,
			Id:     int32(t.Id),
			NextId: int32(t.NextId),
		}
		if s, ok := ticket.Tickets[t.Id]; ok {
			info.Cell = s.Scratch
		} else {
			info.Cell = make([]uint32, t.Cell)
		}
		bean.Info = append(bean.Info, info)
	}
	// 今日中奖牌
	value, err := client.LRange(context.Background(), data.SCRATCHTICKET_POKER, 0, -1).Result()
	if err != nil {
		glog.Error("scratchticket data err, user:", rs.Userid)
		return
	}
	for _, v := range value {
		val, _ := strconv.Atoi(v)
		bean.Jackpots = append(bean.Jackpots, uint32(val))
	}
	bean.Rewards = tickets[0].Jackpot
	return
}

// 分享初始化数据
func (rs *RoleActor) sharedataNtf() {
	// share := table.GetTables().ShareConfigTable.Get()
	share := config.GetShare(1)
	if share.Id == 0 {
		return
	}
	if rs.User.ShareBelow == nil {
		rs.User.ShareBelow = make(map[string]data.ShareData)
	}
	if rs.Userid == rs.ShareSuperior {
		rs.ShareSuperior = ""
		delete(rs.ShareBelow, rs.Userid)
		rs.status = true
	}

	link := share.Url
	// addr := config.GetShareAddr(rs.AD_BundleId)
	// addr := table.GetTables().ShareAddrTable.Get(rs.AD_BundleId)
	addr := config.GetShareAddr(rs.AD_BundleId)
	if addr.Id != "" {
		link = addr.Link
	}
	// 是否pc端
	if rs.Platform == "pc" {
		link = share.PCUrl
	}

	ntf := new(pb.ShareDataNtf)
	ntf.Url = fmt.Sprintf("%sadj_label=share-%s", link, rs.Userid)
	ntf.Url1 = fmt.Sprintf("%sadj_label=share-%s_%d", link, rs.Userid, 1) // 转盘
	ntf.Url2 = fmt.Sprintf("%sadj_label=share-%s_%d", link, rs.Userid, 2) // 代理
	rs.Send(ntf)
}

// 领取分享的邀请奖励
func (rs *RoleActor) shareWithdraw(arg *pb.ShareWithdrawReq) {
	user := rs.User
	rsp := new(pb.ShareWithdrawRsp)
	var withdraw int64 = 0
	if arg.Ctype == 1 {
		//打码量提取
		withdraw = user.ShareWithdraw
		if withdraw <= 0 {
			rsp.Error = pb.UnWithdrawable
			rs.Send(rsp)
			return
		}
		user.ShareWithdraw = 0
		rs.sendGood(0, 0, withdraw, 0, 0, 0, int32(pb.LOG_TYPE82), "分享提取")
	} else {
		// 充值奖励
		// sharebean := table.GetTables().ShareConfigTable.Get()
		sharebean := config.GetShare(1)
		if sharebean.Id == 0 || user.State != 2 {
			rsp.Error = pb.UnWithdrawable
			rs.Send(rsp)
			return
		}
		for k := range user.ShareBelow {
			sd := user.ShareBelow[k]
			if sd.Receive {
				continue
			}
			if sd.Recharge < int64(sharebean.Recharge) {
				continue
			}
			withdraw += int64(sharebean.FirstRecharge)
			sd.Receive = true
			user.ShareBelow[k] = sd
		}
		if withdraw <= 0 {
			rsp.Error = pb.UnWithdrawable
			rs.Send(rsp)
			return
		}
		rs.sendGood(0, 0, withdraw, 0, 0, 0, int32(pb.LOG_TYPE82), "分享提取")
	}
	rsp.Ctype = arg.Ctype
	rs.Send(rsp)
	// 日志
	log := &pb.LogShareWithDraw{
		Userid: rs.Userid,
		Adid:   user.AD_BundleId,
		Gtype:  arg.Ctype,
		Cash:   withdraw,
	}
	myactor.Logger().Tell(log)
}

func (rs *RoleActor) shareReferralReq(arg *pb.ShareReferralReq) {
	user := rs.User
	rsp := new(pb.ShareReferralRsp)
	for _, sd := range user.ShareBelow {
		bean := &pb.ShareReferral{
			Id:        sd.UserId,
			Name:      sd.Name,
			BetAmount: sd.BetAmount,
			LastLogin: sd.LastLogin,
		}
		rsp.Referral = append(rsp.Referral, bean)
	}
	sort.Slice(rsp.Referral, func(i, j int) bool { return rsp.Referral[j].LastLogin < rsp.Referral[i].LastLogin })
	rs.Send(rsp)
}

func (rs *RoleActor) shareRewardReq(arg *pb.ShareRewardReq) {
	user := rs.User
	rsp := new(pb.ShareRewardRsp)
	for _, sal := range user.ShareBetLog {
		bean := &pb.ShareReward{
			Id:        sal.Id,
			BetAmount: sal.BetAmount,
			Date:      sal.Ctime,
		}
		for _, sd := range sal.Detail {
			bean.Income += sd.Income
		}
		rsp.Reward = append(rsp.Reward, bean)
	}
	sort.Slice(rsp.Reward, func(i, j int) bool { return rsp.Reward[j].Date < rsp.Reward[i].Date })
	rs.Send(rsp)
}

func (rs *RoleActor) shareRewardDetailReq(arg *pb.ShareRewardDetailReq) {
	uid := arg.Uid
	rsp := new(pb.ShareRewardDetailRsp)
	for _, sal := range rs.User.ShareBetLog {
		if sal.Id == uid {
			for _, sd := range sal.Detail {
				detail := &pb.ShareRewardDetail{
					Id:        sd.UserId,
					Name:      sd.Name,
					BetAmount: sd.BetAmount,
					Income:    sd.Income,
					Date:      sal.Ctime,
				}
				rsp.Detail = append(rsp.Detail, detail)
			}
		}
	}
	rs.Send(rsp)
}

func (rs *RoleActor) shareData(arg *pb.ShareDataReq) {
	rsp := new(pb.ShareDataRsp)
	rsp.Referrals = int32(len(rs.ShareBelow))
	if rs.State != 2 {
		rs.Send(rsp)
		return
	}
	rsp.TotalCash = rs.ShareTotal
	rsp.Withdrawable = rs.ShareWithdraw
	for _, sd := range rs.ShareBelow {
		sharebean := config.GetShare(1)
		if sharebean.Id == 0 {
			break
		}
		if sd.Receive {
			continue
		}
		if sharebean.Recharge <= int32(sd.Recharge) {
			rsp.ReRecharge += int64(sharebean.FirstRecharge)
		}
	}
	rs.Send(rsp)
}

func (rs *RoleActor) GiveAndOutCash(ctx actor.Context) {
	arg := ctx.Message().(*pb.GiveAndOutCash)
	user := rs.User
	// user.GiveDiamond += arg.Give
	user.OutDiamond += arg.Out
	// user.OutDiamond = int64(math.Min(float64(user.OutDiamond), float64(user.Diamond)))
	rs.rolePid.Tell(arg)
	// 通知前端
	msg := handler.PushCurrencyMsg(0, 0, 0, 0, rs.OutDiamond, 0)
	rs.Send(msg)

	rs.ET_NewbieCondition()

	// 关闭输分补偿
	// rs.checkVBOutDiamond(arg.Out)

	// if rs.OutDiamond >= 1000*100 && rs.Money == 0 { //可提现平民条件
	// 	rs.State = 3
	// }
	// if rs.OutDiamond >= 1000*100 && rs.Money == 0 && rs.RegistArea == 1 && !rs.Withdrawable200 { //B类可提现平民条件
	// 	rs.Withdrawable200 = true
	// 	// B类到达200后，再玩10分钟或者5局变平民
	// 	rs.SurplusGame = 5
	// 	rs.SurplusGameTime = 60 * 10
	// }
	rs.status = true

	// 同步彩票房间
	if rs.cpPid != nil {
		rs.cpPid.Tell(&pb.ChangeCurrency{Userid: rs.Userid, Out: arg.Out})
	}
	// 需要同步游戏
	if arg.SyncGame && rs.gamePid != nil {
		rs.gamePid.Tell(&pb.ChangeCurrency{Userid: rs.Userid, Out: arg.Out})
	}
}

// VB可提现分数输分补偿累计
func (rs *RoleActor) checkVBOutDiamond(outDiamond int64) {
	if outDiamond >= 0 {
		return
	}
	// 未充值不加输分补偿
	if rs.User.GetMoney() <= 0 {
		return
	}
	loseCom := table.GetTables().LoseCompensationTable.Get()
	if loseCom == nil {
		return
	}
	vbOut := (-outDiamond) * int64(loseCom.Compensate) / 10000
	rs.User.VBOutDiamond += vbOut
	// 变化通知
	if vbOut != 0 {
		msg := &pb.ActivityDataNtf{
			VbLoseCompensation: handler.BuildVBLoseCompensationMsg(rs.User), // vb输分补偿
		}
		rs.Send(msg)
	}
}

// 检测玩家当前状态
func (rs *RoleActor) checkUserState() {
	if rs.gamePid != nil || rs.cpPid != nil {
		return
	}
	rs.NewBiewConvertNormal()
}

func (rs *RoleActor) NewBiewConvertNormal() bool {
	user := rs.User
	if rs.State == 1 || rs.State == 3 {
		// 新手或平民，检测是否要切换为正常状态
		// bean := table.GetTables().NewbieTable.Get()
		// if bean == nil {
		// 	return false
		// }
		if user.Money <= 0 {
			return false
		}
		// 满足条件了切换为正常或泡沫状态
		// rs.State = rs.NormalOrFrothState()
		rs.State = data.NormalState
		// rs.toggleNormalOrFroth()
		// 赠送金转换
		if user.BGiveCash == 0 {
			user.BGiveCash = int64(rs.Money)
		}
		outdaimond := -rs.OutDiamond
		diamond := user.Diamond - user.BGiveCash
		vb := diamond
		if vb < 0 {
			vb = user.Diamond
		}
		rs.sendGood(-diamond, 0, vb, outdaimond, 0, 0, int32(pb.LOG_TYPE103), "转为正常玩家")

		// 事件
		event.Event(rs.User, event.LHD_Reset, &event.LHDStrategyResetEvent{Strategys: []int{3}})

		// if rs.RegistArea == 0 {
		// 	add := user.Diamond - int64(user.Money) - user.GiveDiamond // 增加的赠送金
		// 	if add > 0 {
		// 		// rs.sendGood(0, 0, add, 0, 0, 0, int32(pb.LOG_TYPE11), "转为正常玩家")
		// 	}
		// } else {

		// 	// 扣的钱进罐子
		// 	// ntf := event.Event(rs.User, event.RECHARGE_JAR, &event.RechargeJarEvent{Give: diamond})
		// 	// if ntf != nil {
		// 	// 	rs.Send(ntf)
		// 	// }
		// 	// 日志
		// 	// rs.BonusLog(rs.Userid, diamond)
		// }

		return true
	}

	// 泡沫状态切换正常状态
	if handler.FrothToNormalState(rs.User) {
		rs.State = data.NormalState
		return true
	}
	return false
}

// 新手领取金币埋点
func (rs *RoleActor) ET_RegistReward() {
	log := &pb.LogEventTrack{
		Typ:    pb.ETT_RegistReward,
		Userid: rs.Userid,
	}
	myactor.Logger().Tell(log)
}

// 第一局、第二局游戏埋点
func (rs *RoleActor) ET_FirstAndSecondRound(arg *pb.FreeSetRecord) {
	if rs.Round == 1 {
		log := &pb.LogEventTrack{
			Typ:    pb.ETT_FirstRoundGame,
			Gtype:  arg.Gtype,
			Userid: rs.Userid,
		}
		myactor.Logger().Tell(log)
	} else if rs.Round == 2 {
		log := &pb.LogEventTrack{
			Typ:    pb.ETT_SecondRoundGame,
			Gtype:  arg.Gtype,
			Userid: rs.Userid,
		}
		myactor.Logger().Tell(log)
	}
}

// 新手引导埋点1
func (rs *RoleActor) ET_NewbieGuide1(arg *pb.FreeSetRecord) {
	if rs.NewbieGuid {
		return
	}

	if rs.Round == 1 && (arg.Gtype == int32(pb.HUA) || arg.Gtype == int32(pb.HUA2)) {
		log := &pb.LogEventTrack{
			Typ:    pb.ETT_PlayTP,
			Userid: rs.Userid,
		}
		myactor.Logger().Tell(log)
	} else if rs.Round == 2 && (arg.Gtype == int32(pb.HUA) || arg.Gtype == int32(pb.HUA2)) {
		log := &pb.LogEventTrack{
			Typ:    pb.ETT_ContinuePlayTP,
			Userid: rs.Userid,
		}
		myactor.Logger().Tell(log)
	}
}

// 新手引导埋点2
func (rs *RoleActor) ET_NewbieGuide2(arg *pb.LeftDesk) {
	if rs.NewbieGuid {
		return
	}

	if (arg.Gtype == int32(pb.HUA) || arg.Gtype == int32(pb.HUA2)) && arg.Reason == 1 { //tp主动离开
		rs.NewbieGuid = true
		rs.status = true
		if rs.Round == 1 {
			log := &pb.LogEventTrack{
				Typ:    pb.ETT_LeaveTP,
				Userid: rs.Userid,
			}
			myactor.Logger().Tell(log)
		}
	}
}

// 新手状态埋点
func (rs *RoleActor) ET_NewbieCondition() {
	if !rs.NewbieCondition1 {
		if rs.OutDiamond >= 105*100 && rs.OutDiamond < 200*100 {
			rs.NewbieCondition1 = true
			rs.status = true
			log := &pb.LogEventTrack{
				Typ:    pb.ETT_NewbieCondition1,
				Userid: rs.Userid,
			}
			myactor.Logger().Tell(log)
		}
	}

	if !rs.NewbieCondition2 {
		if rs.OutDiamond >= 200*100 {
			rs.NewbieCondition2 = true
			rs.status = true
			log := &pb.LogEventTrack{
				Typ:    pb.ETT_NewbieCondition2,
				Userid: rs.Userid,
			}
			myactor.Logger().Tell(log)
		}
	}
}

// 分享结算
func (a *RoleActor) ShareSettlement(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareSettlement)
	glog.Debugf("ShareSettlement %#v", arg)
	user := a.User
	if b, ok := user.ShareBelow[arg.Userid]; ok {
		b.BetAmount += arg.Total
		b.LastLogin = arg.LastLogin
		user.ShareBelow[arg.Userid] = b
	}
	user.ShareTotal += arg.Score
	user.ShareWithdraw += arg.Score
	// 奖励日志
	dateStr := utils.DateLocalStr()
	detail := data.ShareDetail{
		UserId:    arg.Userid,
		Name:      arg.Name,
		Income:    arg.Score,
		BetAmount: arg.Total,
	}
	if log, ok := user.ShareBetLog[dateStr]; ok {
		log.BetAmount += arg.Total
		log.Detail = append(log.Detail, detail)
		user.ShareBetLog[dateStr] = log
	} else {
		log := data.ShareAmountLog{
			Id:        bson.NewObjectId().String(),
			BetAmount: arg.Total,
			Ctime:     utils.LocalTime().Unix(),
		}
		log.Detail = append(log.Detail, detail)
		user.ShareBetLog[dateStr] = log
	}
	a.status = true
	// 再发送给上级
	if user.ShareSuperior == "" {
		return
	}
	// shareBean := table.GetTables().ShareConfigTable.Get()
	// if shareBean == nil || arg.Level > 1 {
	// 	return
	// }
	sharebean := config.GetShare(1)
	if sharebean.Id == 0 || arg.Level > 1 {
		return
	}
	arg.Score = arg.Total * int64(sharebean.SecondReward) / 10000
	arg.Superior = user.ShareSuperior
	arg.Level = 2
	a.rolePid.Tell(arg)
}

func (a *RoleActor) ChangeShareSuperior(ctx actor.Context) {
	arg := ctx.Message().(*pb.ChangeShareSuperior)
	glog.Debugf("ChangeShareSuperior %#v", arg)
	if arg.Superior == a.Userid {
		a.ShareBelow[arg.Userid] = data.ShareData{
			UserId:    arg.Userid,
			LastLogin: utils.LocalTime().Unix(),
			Name:      arg.Name,
		}
	}
	if arg.Userid == a.Userid {
		if arg.Superior == "" {
			a.ShareSuperior = ""
		} else {
			a.ShareSuperior = arg.Superior
		}
	}
	a.status = true
}

func (a *RoleActor) RemoveShareJuntor(ctx actor.Context) {
	arg := ctx.Message().(*pb.RemoveShareJuntor)
	glog.Debugf("RemoveShareJuntor %#v", arg)
	user := a.User
	delete(user.ShareBelow, arg.Userid)
	a.status = true
}

func (a *RoleActor) ChangePhone(ctx actor.Context) {
	arg := ctx.Message().(*pb.ChangePhone)
	glog.Debugf("ChangePhone %#v", arg)
	a.User.Phone = arg.Phone
	a.User.Auth = arg.Auth
	a.Password = arg.Pwd
}

func (a *RoleActor) ChangeViceAccount(ctx actor.Context) {
	arg := ctx.Message().(*pb.ChangeViceAccount)
	a.User.ViceAccount = arg.ViceAccount
	if a.gamePid != nil {
		a.gamePid.Tell(arg)
	}
}

func (a *RoleActor) ClearBankInfo(ctx actor.Context) {
	arg := ctx.Message().(*pb.ClearBankInfo)
	glog.Debugf("ClearBankInfo %#v", arg)
	a.Bank = ""
	a.IFSC = ""
	a.BankAccounts = ""
	a.status = true
}

func (a *RoleActor) GiveWithdrawCash(ctx actor.Context) {
	arg := ctx.Message().(*pb.GiveWithdrawCash)
	glog.Debugf("GiveWithdrawCash %#v", arg)
	var diamond int64 = 0 // 需要加的钻石
	if a.OutDiamond+arg.OutCash > a.Diamond {
		diamond = a.OutDiamond + arg.OutCash - a.Diamond
	}
	a.sendGood(diamond, 0, 0, arg.OutCash, 0, 0, int32(pb.LOG_TYPE83), arg.Desc)
}

// 完成游戏引导
func (rs *RoleActor) gameGuidReq(arg *pb.GameGuidReq) {
	user := rs.User
	rsp := new(pb.GameGuidRsp)
	guid := &data.GameGuid{Gtype: -1}
	if arg.Ftype != 0 {
		guid = &data.GameGuid{Ftype: -1}
	}
	for _, gg := range user.GameGuildSlice {
		if gg.Gtype > 0 && gg.Gtype == arg.Gtype ||
			(gg.Ftype > 0 && gg.Ftype == arg.Ftype) {
			guid = gg
			break
		}
	}
	if guid.Gtype == -1 || guid.Ftype == -1 {
		// 当前游戏没有引导过，初始化
		guid = &data.GameGuid{Gtype: arg.Gtype, Ftype: arg.Ftype}
		user.GameGuildSlice = append(user.GameGuildSlice, guid)
	}
	exists := false
	for _, v := range guid.GuidId {
		if v == int32(arg.GuidId) {
			// 已经存在了
			exists = true
			break
		}
	}
	if !exists {
		guid.GuidId = append(guid.GuidId, int32(arg.GuidId))
	}
	rs.status = true
	rsp.Gtype = arg.Gtype
	rsp.Ftype = arg.Ftype
	rsp.GuidId = arg.GuidId
	rs.Send(rsp)
}

// 通过时间去计算领那一天的周卡奖励(兼容已经买过周卡的玩家)
func (rs *RoleActor) getWeeklyReward(id int32) (int64, int64, int64) {
	cardId := fmt.Sprintf("lbzk%d", id)
	bean := table.GetTables().GiftRechargeTable.Get(cardId)

	now := utils.LocalTime()
	c := rs.User.WeeklyCardMap[id]
	ovTime := utils.Stamp2Time(c.OverTime)
	ovTime = time.Date(ovTime.Year(), ovTime.Month(), ovTime.Day(), 0, 0, 0, 0, time.Local)
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	dur := ovTime.Sub(now)
	days := int(dur.Hours() / 24)
	maxDay := len(bean.DailyGive)
	if days > maxDay {
		return 0, 0, 0
	}
	if maxDay-days >= len(bean.DailyGive) {
		return 0, 0, 0
	}
	return int64(bean.DailyGive[maxDay-days].Nums[0]) * int64(bean.Price) / 10000,
		int64(bean.DailyGive[maxDay-days].Nums[1]) * int64(bean.Price) / 10000,
		int64(bean.DailyGive[maxDay-days].Nums[2]) * int64(bean.Price) / 10000
}

func (rs *RoleActor) sendOfflineMsg() {
	if len(rs.OfflineChats) <= 0 {
		return
	}
	ntf := new(pb.ConsumerLogNtf)
	for _, chat := range rs.OfflineChats {
		log := &pb.ConsumerLog{
			Userid:  chat.Sender,
			Ctime:   chat.Ctime,
			Content: chat.Content,
		}
		ntf.Chats = append(ntf.Chats, log)
	}
	rs.Send(ntf)
	// 清空消息
	rs.OfflineChats = make([]data.OfflineChatLog, 0)
}

// 发送客服地址的通知
func (rs *RoleActor) customerAddr() {
	ntf := new(pb.CustomerAddressNtf)
	addrs := config.GetCustomerAddressMap()
	for _, addr := range addrs {
		ntf.Mail = addr.Mail
		ntf.Telegram = addr.Telegram
		ntf.WhatsApp = addr.WhatsApp
		break
	}
	rs.Send(ntf)
}

// 分享配置信息
func (rs *RoleActor) shareWayNtf() {
	ntf := new(pb.ShareWayNtf)
	shares := config.GetShareWayMap()
	for _, share := range shares {
		ntf.Youtobe = share.Youtobe
		ntf.Ins = share.Ins
		ntf.Facebook = share.Facebook
		ntf.Telegram = share.Telegram
		ntf.CashPrize = share.CashPrize
		ntf.Whatsapp = share.WhatsApp
		ntf.X = share.X
		break
	}
	rs.Send(ntf)
}

func (rs *RoleActor) photoNtf() {
	url, err := client.Get(context.Background(), data.GetCustomPhotoKey(rs.Userid)).Result()
	if err != nil || url == "" {
		return
	}
	ntf := &pb.PhotoeAuditStatusNtf{
		Status: pb.AUDIT,
		Url:    url,
	}
	rs.Send(ntf)
}

// 限时礼包
// func (rs *RoleActor) limitedGiftNtf() {
// 	if rs.LimitedGiftId == 0 {
// 		return
// 	}
// 	if utils.LocalTime().UnixMilli() > rs.LimitedGiftOverTime {
// 		rs.OverlimitedGift = append(rs.OverlimitedGift, rs.LimitedGiftId)
// 		rs.LimitedGiftId = 0
// 		return
// 	}
// 	bean := config.GetLimitedGift(rs.LimitedGiftId)
// 	ntf := &pb.LimitedGiftNtf{
// 		Id:       bean.Id,
// 		Price:    bean.Price,
// 		Cash:     int32(bean.Reward),
// 		OverTime: rs.LimitedGiftOverTime,
// 	}
// 	rs.Send(ntf)
// }

// pvpRoom 对战房配置信息
func (rs *RoleActor) pvpRoom() {
	pvpRoom := config.GetPvpRoom()
	if pvpRoom == nil {
		glog.Error("未上传对战房基本配置")
		return
	}

	// 组装配置信息给前端
	ntf := &pb.PvpRoomNtf{}
	// TP
	hua := &pb.PvpRoomGame{Gtype: pb.HUA, Modes: &pb.PvpRoomMode{}}
	ntf.Games = append(ntf.Games, hua)
	for _, round := range pvpRoom.TpGoldRound {
		hua.Modes.Cash = append(hua.Modes.Cash, &pb.PvpRoomRoundFee{
			Round: uint32(round),
			Fee:   0,
		})
	}
	for i, round := range pvpRoom.TpFunRoundCost[0] {
		hua.Modes.Entertainment = append(hua.Modes.Entertainment, &pb.PvpRoomRoundFee{
			Round: uint32(round),
			Fee:   uint32(pvpRoom.TpFunRoundCost[1][i]),
		})
	}
	// RUMMY
	rummy := &pb.PvpRoomGame{Gtype: pb.RUMMY, Modes: &pb.PvpRoomMode{}}
	ntf.Games = append(ntf.Games, rummy)
	for _, round := range pvpRoom.RmGoldRound {
		rummy.Modes.Cash = append(rummy.Modes.Cash, &pb.PvpRoomRoundFee{
			Round: uint32(round),
			Fee:   0,
		})
	}
	for i, round := range pvpRoom.RmFunRoundCost[0] {
		rummy.Modes.Entertainment = append(rummy.Modes.Entertainment, &pb.PvpRoomRoundFee{
			Round: uint32(round),
			Fee:   uint32(pvpRoom.RmFunRoundCost[1][i]),
		})
	}
	// AB
	ab := &pb.PvpRoomGame{Gtype: pb.ABAR, Modes: &pb.PvpRoomMode{}}
	ntf.Games = append(ntf.Games, ab)
	for _, round := range pvpRoom.AbGoldRound {
		ab.Modes.Cash = append(ab.Modes.Cash, &pb.PvpRoomRoundFee{
			Round: uint32(round),
			Fee:   0,
		})
	}
	for i, round := range pvpRoom.AbFunRoundCost[0] {
		ab.Modes.Entertainment = append(ab.Modes.Entertainment, &pb.PvpRoomRoundFee{
			Round: uint32(round),
			Fee:   uint32(pvpRoom.RmFunRoundCost[1][i]),
		})
	}
	// 对战房配置...
	games := config.GetGames()
	for _, game := range games {
		if game.RoomType != 1 { // 非对战房
			continue
		}
		ante := &pb.PvpRoomAntes{RoomId: game.Id, MinAccess: int64(game.Min_Access)}
		switch game.Gtype {
		case int32(pb.HUA):
			ante.Ante = uint32(game.TP.Bottom)
			hua.CashAntes = append(hua.CashAntes, ante)
		case int32(pb.RUMMY):
			ante.Ante = uint32(game.RM.Bottom)
			rummy.CashAntes = append(rummy.CashAntes, ante)
		case int32(pb.ABAR):
			ante.Ante = game.AB.Bottom
			ab.CashAntes = append(ab.CashAntes, ante)
		}
	}
	sort.Slice(hua.CashAntes, func(i, j int) bool {
		return hua.CashAntes[i].Ante < hua.CashAntes[j].Ante
	})
	sort.Slice(rummy.CashAntes, func(i, j int) bool {
		return rummy.CashAntes[i].Ante < rummy.CashAntes[j].Ante
	})
	sort.Slice(ab.CashAntes, func(i, j int) bool {
		return ab.CashAntes[i].Ante < ab.CashAntes[j].Ante
	})
	rs.Send(ntf)
}

// 当前是什么状态
// func (rs *RoleActor) NormalOrFrothState() int {
// 	if rs.IsB() {
// 		return data.NormalState
// 	}
// 	bean := table.GetTables().NewbieTable.Get()
// 	if bean == nil {
// 		return data.NormalState
// 	}

// 	if rs.Money >= uint32(bean.FrothMaxRecharge) {
// 		// 充值超过泡沫充值最大值
// 		return data.NormalState
// 	}

// 	winScroe := rs.Diamond + int64(rs.CashOut) - int64(rs.Money) // 当前赢分
// 	// 当前赢分/已充值金额 > 30%
// 	if rs.Money > 0 && winScroe*10000/int64(rs.Money) > int64(bean.FrothGiftRate) {
// 		// 进入泡沫状态
// 		return data.FrothState
// 	}
// 	return data.NormalState
// }

func (rs *RoleActor) vipInitDataNtf() {
	// if rs.Money <= 0 {
	// 	// 清空vip
	// 	rs.Vip = data.UserVip{}
	// }
	// 老号初始化BeforeLv
	if !rs.Vip.Init {
		rs.Vip.BeforeLv = rs.Vip.Lv
		rs.Vip.Init = true
	}
	handler.CheckVipLv(rs.User)

	//  fix 老号 vip = 0 时weekly_sign配置成0的问题
	vipBean := table.GetTables().VipTable.Get(int32(rs.Vip.Lv))
	if vipBean.WeeklySign <= 0 {
		rs.Vip.Weekly = true
	}

	nowTime := time.Now().In(location)
	now := nowTime.Unix()
	if vipBean.WeeklySign > 0 && rs.User.Vip.WeeklyTime > 0 && now >= rs.Vip.WeeklyTime {
		rs.Vip.Weekly = false
	}
	if rs.User.Vip.WeeklyTime == 0 || !rs.Vip.Weekly {
		// 下周一零点再重置
		weekday := nowTime.Weekday()
		offset := int(-weekday + 1)
		if weekday == time.Sunday {
			offset = -6
		}
		nextMonday := nowTime.AddDate(0, 0, offset+7)
		nextMondayZero := time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(), 0, 0, 0, 0, nowTime.Location())
		rs.Vip.WeeklyTime = nextMondayZero.Unix()
	}

	ntf := &pb.VipDataNtf{
		Vip:                 handler.BuildVipData(rs.User),
		Progress:            int32(rs.Vip.Exp),
		Lv:                  int32(rs.Vip.Lv),
		Daily:               rs.Vip.Daily,
		Weekly:              rs.Vip.Weekly,
		LvReward:            rs.Vip.LevelReward,
		DailyWithdraw:       int32(rs.Vip.WithdrawCount),
		WeeklyTime:          rs.Vip.WeeklyTime,
		BeforeLv:            int32(rs.Vip.BeforeLv),
		DailyWithdrawAmount: int32(rs.Vip.WithdrawAmount),
		ResetDays:           vipBean.Reset,
		ResetLv:             vipBean.ResetLv,
	}
	if !ntf.Weekly {
		ntf.WeeklySign = int64(vipBean.WeeklySign)
	}
	rs.Send(ntf)
}

func (rs *RoleActor) vipReset() {
	nowTime := time.Now().In(location)
	// nowMonth := utils.Time2MonthDate(nowTime)
	// if rs.Vip.Version == 0 {
	// 	rs.Vip.Version = nowMonth
	// }
	// if nowMonth != rs.Vip.Version {
	// 	// rs.Vip = data.UserVip{Version: nowMonth}
	// }

	rs.Vip.Daily = false
	now := nowTime.Unix()
	weekday := nowTime.Weekday()
	if rs.Vip.WeeklyTime > 0 && now >= rs.Vip.WeeklyTime {
		rs.Vip.Weekly = false

		// 计算下周周一的时间
		offset := int(-weekday + 1)
		if weekday == time.Sunday {
			offset = -6
		}
		nextMonday := nowTime.AddDate(0, 0, offset+7)
		nextMondayZero := time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(), 0, 0, 0, 0, nowTime.Location())
		rs.Vip.WeeklyTime = nextMondayZero.Unix()

		// monday := nowTime.AddDate(0, 0, -int(weekday)+1)
		// rs.Vip.WeeklyTime = monday.AddDate(0, 0, 7+rs.Vip.Weekday).Unix()
	}
	//提现次数更新
	rs.Vip.WithdrawCount = 0
	//提现金额更新
	rs.Vip.WithdrawAmount = 0
}

func (rs *RoleActor) TpReset() {
	rs.TpUserTodayChargeAmount = 0
	rs.TpUserTodayChargeNum = 0
	rs.TpUserTodayWithdrawAmount = 0
	rs.TpUserTodayWithdrawNum = 0
	rs.TpUserTodayGameRound = 0
	rs.TpUserTodayControlStrategyNum = make(map[int32]int32)
}

func (rs *RoleActor) rm2Reset() {
	// rm roi天触发策略
	rs.User.RmRoiDayLimits = make(map[string]int32)
	if rs.gamePid != nil && rs.gtype == int32(pb.RUMMY2) {
		// 同步到 game
		msg := &pb.RMRoiRecordsync{
			Userid:      rs.Userid,
			ResetDayRoi: true,
		}
		rs.gamePid.Tell(msg)
	}
}

func (rs *RoleActor) playShareReset() {
	if rs.PlayShareData.Over {
		return
	}

	// c := config.GetPlayShareConfig(1)
	c := table.GetTables().PlayShareTable.Get()
	if c == nil {
		return
	}

	// 判断活动开没开
	if !config.SettingIsOpenByName(1, "拼多多") {
		return
	}

	if rs.PlayShareData.Config.Id == 0 {
		rs.PlayShareData.Config = data.PlayAndDrawActivity{
			Id:           1,
			ValidityTime: c.During,
			MinDrawTimes: c.DrawTimes,
			MaxPlayDraw:  c.PlayMaxTimes,
			MaxShareDraw: c.ShareMaxTimes,
			Reward:       c.Reward,
			RandSpan:     float64(c.RandomSpan),
			JumpRate:     float64(c.JumpRate),
			DownRate:     float64(c.DecliningRate),
			PlayRounds:   c.PlayTimesDraw,
			ShareFriends: c.ShareTimesDraw,
		}
	}
	// 使用玩家身上的配置
	// c = rs.PlayShareData.Config

	if rs.PlayShareData.OverTime == 0 {
		rs.PlayShareData.OverTime = utils.BsonNow().Unix() + int64(rs.PlayShareData.Config.ValidityTime*60*60)
	}

	if rs.PlayShareData.OverTime < utils.BsonNow().Unix() {
		rs.PlayShareData.Over = true
		return
	}
}

func (rs *RoleActor) RankWithdrawReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RankWithdrawReq)
	glog.Debugf("RankWithdrawReq %#v", arg)
	rs.roomPid.Request(arg, ctx.Self())
}

func (rs *RoleActor) LuckyDrawReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.LuckyDrawReq)
	glog.Debugf("LuckyDrawReq %#v", arg)
	arg.Userid = rs.Userid
	rs.rolePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) LuckyDrawHistoryReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.LuckyDrawHistoryReq)
	glog.Debugf("LuckyDrawHistoryReq %#v", arg)
	arg.Userid = rs.Userid
	rs.rolePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) LuckyDrawTakeReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.LuckyDrawTakeReq)
	glog.Debugf("LuckyDrawTakeReq %#v", arg)
	rsp := &pb.LuckyDrawTakeRsp{}

	req := &pb.LuckyDrawTake{Userid: rs.Userid, Round: arg.Round}
	result, err := ctx.RequestFuture(rs.rolePid, req, time.Second*3).Result()
	if err != nil {
		glog.Errorf("LuckyDrawTake failed: %v", err)
		rsp.Error = pb.Failed
		rs.Send(rsp)
		return
	}
	taked, ok := result.(*pb.LuckyDrawTaked)
	if !ok {
		glog.Errorf("LuckyDrawTake type failed: %v", err)
		rsp.Error = pb.Failed
		rs.Send(rsp)
		return
	}
	if taked.Error != pb.OK {
		glog.Errorf("LuckyDrawTake type error: %v", taked.Error)
		rsp.Error = taked.Error
		rs.Send(rsp)
		return
	}
	// 发奖
	rs.sendGood(taked.RewordAmount, 0, 0, 0, 0, 0, int32(pb.LOG_TYPE104),
		fmt.Sprintf("小米手机活动%d期%d等奖号码%s", taked.Round, taked.RewordId, taked.Number))
	rsp.RewordAmount = taked.RewordAmount
	rs.Send(rsp)
}

func (rs *RoleActor) LuckyDrawNumberMail(ctx actor.Context) {
	arg := ctx.Message().(*pb.LuckyDrawNumberMail)
	var content string
	if arg.Mtype == 2 { // 中奖邮件
		content = handler.BuildLuckyDrawWin(rs.Nickname, int(arg.Level), int(arg.Round), arg.Amount/100)
	} else { // 参与邮件
		content = handler.BuildLuckyDraw(rs.Nickname, arg.Numbers, arg.Amount/100)
	}
	chat := data.ChatLog{
		Uid:      bson.NewObjectId().String(),
		Receiver: rs.Userid,
		Title:    "System Message",
		Content:  content,
		Ctime:    utils.LocalTime(),
		Sender:   "-1",
		Name:     "System",
	}
	rs.FeedBackLogMap[chat.Uid] = chat

	ntf := new(pb.FeedBackLogNtf)
	ntf.Logs = append(ntf.Logs, &pb.ChatLog{Uid: chat.Uid, Userid: chat.Sender, Title: chat.Title, Ctime: chat.Ctime.Unix(), Content: chat.Content})
	rs.Send(ntf)
}

func (rs *RoleActor) ScratchTicketRewardReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ScratchTicketRewardReq)
	glog.Debugf("ScratchTicketRewardReq %#v", arg)

	rsp := new(pb.ScratchTicketRewardRsp)
	defer rs.Send(rsp)
	// 玩家刮开过的格子
	tickets := rs.ScratchTicket.Tickets

	// 使用过的牌
	usedPokerMap := make(map[uint32]string)
	configs := config.GetScratchTicketConfigMap()
	bean := configs[1]
	for {
		if t, ok := tickets[bean.Id]; ok {
			used := true
			for _, poker := range t.Scratch {
				if poker == 0 {
					used = false
					break
				}

				usedPokerMap[poker] = ""
			}
			if used {
				bean = configs[bean.NextId]
			} else {
				break
			}
		} else {
			break
		}

		if bean.Id <= 0 {
			break
		}
	}

	if bean.Id <= 0 {
		rsp.Error = pb.AwardFaild
		return
	}

	// 获取bonus
	pots := rs.ActivatedPot
	if len(pots) <= 0 {
		rsp.Error = pb.NotEnoughBonus
		return
	}

	sort.Slice(pots, func(i, j int) bool {
		return pots[i].Number > pots[j].Number
	})

	var allBonus int64
	var removeIndex int = -1
	for _, pot := range pots {
		allBonus += pot.Number
	}

	if allBonus < bean.Cost {
		rsp.Error = pb.NotEnoughBonus
		return
	}

	// 扣bonus
	cost := bean.Cost
	for i, pot := range pots {
		if pot.Number > cost {
			multiple := pot.FlowWater / pot.Number
			pot.Number -= cost
			pot.FlowWater = multiple * pot.Number
			pots[i] = pot
			break
		}
		cost -= pot.Number
		removeIndex = i
	}

	if removeIndex > -1 {
		pots = pots[removeIndex+1:]
	}
	rs.ActivatedPot = pots

	// 获取中奖范围配置
	ruleId := bean.Rule[len(bean.Rule)-1]
	for i, v := range bean.Scope {
		if rs.ScratchTicket.CostBonus <= v {
			ruleId = bean.Rule[i]
			break
		}
	}
	rule := bean.RuleConfig[0]
	for _, r := range bean.RuleConfig {
		if r.Id == ruleId {
			rule = r
			break
		}
	}
	rs.ScratchTicket.CostBonus += bean.Cost

	// 中奖日志
	log := &pb.ScratchTicketLog{
		Userid: rs.Userid,
		Cost:   bean.Cost,
	}

	// 奖金
	var bonus int64
	detail := data.ScratchTicketDetail{
		Id: bean.Id,
	}
	// 开牌
	for i := 0; i < bean.Cell; i++ {
		poker := rs.ScratchTicket.Pokers[0]
		rs.ScratchTicket.Pokers = rs.ScratchTicket.Pokers[1:]
		index, err := utils.ChoiceIntIndex(rule.WeightPro)
		if err != nil {
			glog.Error("scratch ticket error", err)
			index = 5
		}

		for i := 0; i < len(rs.ScratchTicket.Pokers); i++ {
			success := true
			for _, v := range scratchTick.Poker {
				if v == poker {
					// 换一张
					poker = rs.ScratchTicket.Pokers[0]
					rs.ScratchTicket.Pokers = rs.ScratchTicket.Pokers[1:]
					success = false
					break
				}
			}
			if success {
				break
			}
		}

		if index >= 5 {
			// 没中奖
			usedPokerMap[poker] = ""
			detail.Scratch = append(detail.Scratch, poker)
			continue
		}
		// 中奖了
		// sup := scratchTick.Poker[index]
		r, _ := client.LIndex(context.Background(), data.SCRATCHTICKET_POKER, int64(index)).Result()
		s, _ := strconv.Atoi(r)
		sup := uint32(s)
		// 如果中奖牌开过了就不开了
		if _, ok := usedPokerMap[sup]; ok {
			usedPokerMap[poker] = ""
			detail.Scratch = append(detail.Scratch, poker)
		} else {
			// 累计奖金
			if index > 2 {
				// 中了前3等奖,找客服领
				bonus += bean.Jackpot[index]
			}
			usedPokerMap[sup] = ""
			detail.Scratch = append(detail.Scratch, sup)
			log.Level = append(log.Level, int32(index+1))
			log.Bonus = append(log.Bonus, bean.Jackpot[index])
		}
	}

	rs.ScratchTicket.Tickets[bean.Id] = detail
	if bonus > 0 {
		rs.sendGood(bonus, 0, 0, 0, 0, 0, int32(pb.LOG_TYPE118), "大富翁")
	}
	rsp.Scratch = rs.buildScratchTicketDataMsg()

	// 重置罐子流水上限
	rs.checkShopPotFlowMax()
	// 罐子活动更新
	ntf := &pb.ActivityDataNtf{
		ShopPots: handler.BuildShopPotDataMsg(rs.User),
	}
	rs.Send(ntf)

	rs.status = true

	myactor.Logger().Tell(log)

	// bonus日志
	rs.BonusLog(rs.Userid, -bean.Cost)
}

// 大富翁重置检测
func (rs *RoleActor) resetScratchTicket() {
	if rs.ScratchTicket.Version == scratchTick.Version {
		return
	}

	// 重置
	configs := config.GetScratchTicketConfig()
	glog.Infof("reset scratch ticket, user:%s,version:%d,newversion:%d", rs.Userid, rs.ScratchTicket.Version, scratchTick.Version)
	rs.ScratchTicket.Version = scratchTick.Version
	rs.ScratchTicket.Pokers = algo.ShuffleCards()
	rs.ScratchTicket.Tickets = make(map[int]data.ScratchTicketDetail)
	for _, c := range configs {
		rs.ScratchTicket.Tickets[c.Id] = data.ScratchTicketDetail{Id: c.Id, Scratch: make([]uint32, c.Cell)}
	}
	rs.status = true

	// 通知
	ntf := &pb.ActivityDataNtf{
		Scratch: rs.buildScratchTicketDataMsg(),
	}
	rs.Send(ntf)
}

func (rs *RoleActor) ActivityMonitorReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ActivityMonitorReq)
	glog.Debugf("ActivityMonitorReq %#v", arg)
	rsp := &pb.ActivityMonitorRsp{}

	myactor.Logger().Tell(arg)
	defer rs.Send(rsp)
}

// 重置罐子流水
func (rs *RoleActor) checkShopPotFlowMax() {
	var maxFlow int64
	for _, pot := range rs.ActivatedPot {
		maxFlow += pot.FlowWater
	}
	rs.PotFlow = int64(math.Min(float64(rs.PotFlow), float64(maxFlow)))
}

func (rs *RoleActor) PlayShareDrawReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.PlayShareDrawReq)
	glog.Debugf("PlayShareDrawReq %#v", arg)
	rsp := &pb.PlayShareDrawRsp{}

	defer rs.Send(rsp)

	// c := config.GetPlayShareConfig(1)
	t := table.GetTables().PlayShareTable.Get()
	if t == nil {
		rsp.Error = pb.Failed
		return
	}

	// 是否过期
	if rs.PlayShareData.OverTime < time.Now().Unix() || rs.PlayShareData.Over {
		rsp.Error = pb.ActivityOver
		return
	}

	if rs.PlayShareData.Config.Id == 0 {
		rs.PlayShareData.Config = data.PlayAndDrawActivity{
			Id:           1,
			ValidityTime: t.During,
			MinDrawTimes: t.DrawTimes,
			MaxPlayDraw:  t.PlayMaxTimes,
			MaxShareDraw: t.ShareMaxTimes,
			Reward:       t.Reward,
			RandSpan:     float64(t.RandomSpan),
			JumpRate:     float64(t.JumpRate),
			DownRate:     float64(t.DecliningRate),
			PlayRounds:   t.PlayTimesDraw,
			ShareFriends: t.ShareTimesDraw,
		}
	}
	// 使用玩家身上的配置
	c := rs.PlayShareData.Config

	// 优先使用玩游戏的
	ctype := 1 // 1：玩游戏的 2:分享的
	d := rs.PlayShareData
	if d.PlayDraws > 0 {
		rs.PlayShareData.PlayDraws--
		rs.PlayShareData.PlayUsed++
	} else if d.ShareDraws > 0 {
		ctype = 2
		rs.PlayShareData.ShareDraws--
		rs.PlayShareData.ShareUsed++
	} else {
		rsp.Error = pb.NotEnoughDrawTimes
	}

	drawTimes := rs.PlayShareData.ShareUsed + rs.PlayShareData.PlayUsed

	var w, jackpot int64 = 0, 0
	u := c.JumpRate
	z := c.MaxPlayDraw
	for _, v := range c.Reward {
		w += int64(v)
	}
	y := float64(w) * u

	// 开始抽奖
	// if ctype == 1 {
	if drawTimes <= int(z) {
		min := y / (c.RandSpan * float64(z))
		max := y / float64(z)
		jackpot = utils.RandInt64N(int64(max-min)) + int64(min)
		rs.PlayShareData.PlayGet += jackpot
	} else {
		if drawTimes-int(z) == 1 {
			jackpot = int64(y - float64(rs.PlayShareData.PlayGet))
		} else {
			if drawTimes >= int(c.MinDrawTimes) {
				jackpot = w - rs.PlayShareData.Rewards
			} else {
				jackpot = int64(c.DownRate * float64(w-rs.PlayShareData.Rewards))
			}
		}
	}

	if w < rs.PlayShareData.Rewards+jackpot {
		jackpot = w - rs.PlayShareData.Rewards
	}

	if jackpot <= 0 {
		glog.Warning("jackpot <= 0, jackpot:%d", jackpot)
	}
	jackpot = int64(math.Max(float64(jackpot), float64(0)))
	rs.PlayShareData.Rewards += jackpot

	rs.status = true

	rsp.Allreward = rs.PlayShareData.Rewards
	rsp.Reward = jackpot
	rsp.Times = int32(rs.PlayShareData.PlayDraws + rs.PlayShareData.ShareDraws)

	if rs.PlayShareData.PlayUsed >= int(c.MaxPlayDraw) {
		ntf := &pb.ActivityDataNtf{PlayShare: handler.BuildPlayShareDataMsg(rs.User)}
		rs.Send(ntf)
	}

	glog.Infof("user %s draw reward %d, ctype %d,drawTimes:%d", rs.Userid, jackpot, ctype, drawTimes)
	// 日志记录
	log := &pb.PlayShareDrawLog{
		Userid: rs.Userid,
		Dtype:  int32(ctype),
		Reward: jackpot,
	}
	myactor.Logger().Tell(log)
}

func (rs *RoleActor) PlayShareRewardReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.PlayShareRewardReq)
	glog.Debugf("PlayShareRewardReq %#v", arg)
	rsp := &pb.PlayShareRewardRsp{}

	defer rs.Send(rsp)

	// c := config.GetPlayShareConfig(1)
	t := table.GetTables().PlayShareTable.Get()
	if t == nil {
		glog.Error("no found config, user:", rs.Userid)
		rsp.Error = pb.CanNotReceive
		return
	}

	if rs.PlayShareData.Config.Id == 0 {
		rs.PlayShareData.Config = data.PlayAndDrawActivity{
			Id:           1,
			ValidityTime: t.During,
			MinDrawTimes: t.DrawTimes,
			MaxPlayDraw:  t.PlayMaxTimes,
			MaxShareDraw: t.ShareMaxTimes,
			Reward:       t.Reward,
			RandSpan:     float64(t.RandomSpan),
			JumpRate:     float64(t.JumpRate),
			DownRate:     float64(t.DecliningRate),
			PlayRounds:   t.PlayTimesDraw,
			ShareFriends: t.ShareTimesDraw,
		}
	}
	// 使用玩家身上的配置
	c := rs.PlayShareData.Config

	// 是否过期
	if rs.PlayShareData.OverTime < time.Now().Unix() || rs.PlayShareData.Over {
		rsp.Error = pb.ActivityOver
		return
	}

	if rs.PlayShareData.PlayUsed+rs.PlayShareData.ShareUsed < int(c.MinDrawTimes) {
		glog.Errorf("can not receive reward,user:%s, draw times not enough:%d, min:%d", rs.Userid, rs.PlayShareData.PlayUsed+rs.PlayShareData.ShareUsed, c.MinDrawTimes)
		rsp.Error = pb.CanNotReceive
		return
	}

	var maxReward int64
	for _, v := range c.Reward {
		maxReward += int64(v)
	}

	if rs.PlayShareData.Rewards < maxReward {
		glog.Errorf("can not receive reward,user:%s,rewards:%d, max:%d", rs.Userid, rs.PlayShareData.Rewards, maxReward)
		rsp.Error = pb.CanNotReceive
		return
	}

	rs.sendGood(int64(c.Reward[0]+c.Reward[1]), 0, 0, int64(c.Reward[1]), 0, 0, int32(pb.LOG_TYPE120), "玩游戏抽奖")

	if len(c.Reward) >= 3 && c.Reward[2] > 0 {
		// 加罐子
		ntf := event.Event(rs.User, event.Normal_JAR, &event.NormalJarEvent{Give: int64(c.Reward[2]), Multiple: 40})
		if ntf != nil {
			rs.Send(ntf)
		}
		// 日志
		rs.BonusLog(rs.Userid, int64(c.Reward[2]))
	}
	rs.PlayShareData.Over = true
	rs.status = true

	// 日志记录
	log := &pb.PlayShareDrawLog{
		Userid: rs.Userid,
		Dtype:  3,
		Reward: maxReward,
	}
	myactor.Logger().Tell(log)
}

func (rs *RoleActor) PlayshareTest() {
	// 玩游戏抽奖
	ntf := event.Event(rs.User, event.PLAY_SHARE, &event.PlayShareEvent{Ptype: 2, DeviceId: rs.AD_DeviceId})
	if ntf != nil {
		rs.Send(ntf)
	}

	rs.status = true
}

func (rs *RoleActor) BonusLog(userid string, bonus int64) {
	if bonus == 0 {
		return
	}

	log := &pb.BonusLog{
		Userid: userid,
		Bonus:  bonus,
	}

	myactor.Logger().Tell(log)
}

// 推送提现跑马灯
func (rs *RoleActor) MarQueeWithdraw(ctx actor.Context) {
	arg := ctx.Message().(*pb.MarQueeWithdraw)
	// 游戏中不推送
	if rs.gamePid != nil {
		return
	}
	rs.Send(arg.Ntf)
}

// 用户自定义头像
func (rs *RoleActor) UserCustomPhoto(ctx actor.Context) {
	arg := ctx.Message().(*pb.UserCustomPhoto)
	if arg.Status == 1 {
		rs.Photo = arg.Url
		rs.status = true
	}

	ntf := &pb.PhotoeAuditStatusNtf{
		Status: pb.PhotoeAuditStatusNtf_Code(arg.Status),
		Url:    arg.Url,
	}
	rs.Send(ntf)
	client.Del(context.Background(), data.GetCustomPhotoKey(rs.Userid))
}

// vip bank 任务奖励领取
func (rs *RoleActor) VBTaskRewardReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.VBTaskRewardReq)
	rsp := new(pb.VBTaskRewardRsp)
	for _, task := range rs.User.VBTask {
		if task.Id == arg.TaskId {
			if task.Prize != 2 {
				switch task.Prize {
				case 1:
					rsp.Error = pb.NotWinning
				case 3:
					rsp.Error = pb.AlreadyPrize
				default:
					rsp.Error = pb.Failed
				}
				rs.Send(rsp)
				return
			}

			// if task.RewardMode == 2 && rs.User.VBBank < int64(task.Reward) {
			// 	rsp.Error = pb.NotEnoughCoin
			// 	rs.Send(rsp)
			// 	return
			// }

			task.Prize = 3
			// 增长
			rs.sendGood(0, 0, int64(task.Reward), 0, 0, 0, int32(pb.LOG_TYPE123), fmt.Sprintf("完成游戏任务id%d", task.Id))
			// 发放奖励
			// if task.RewardMode == 2 {
			// 	// 释放
			// 	desc := fmt.Sprintf("完成游戏任务id%d", task.Id)
			// 	rs.sendGood(int64(task.Reward), 0, int64(-task.Reward), 0, 0, 0, int32(pb.LOG_TYPE124), desc)
			// 	if rs.gamePid != nil {
			// 		// todo 在游戏中也能领取, 在游戏中同步到游戏
			// 		msg := handler.ChangeCurrencyMsg1(int64(task.Reward), 0,
			// 			0, 0, int64(-task.Reward), 0, int32(pb.LOG_TYPE124), rs.User.GetUserid(), desc, "", false)
			// 		rs.gamePid.Tell(msg)
			// 	}
			// } else {
			// 	// 增长
			// 	rs.sendGood(0, 0, int64(task.Reward), 0, 0, 0, int32(pb.LOG_TYPE123), fmt.Sprintf("完成游戏任务id%d", task.Id))
			// }
		}
	}

	rsp.VbTask = rs.getVBTasks()
	rs.Send(rsp)
}

// GetVBTaskReq 获取vb任务列表, 在游戏中则只获取游戏关联的
func (rs *RoleActor) GetVBTaskReq(ctx actor.Context) {
	rsp := new(pb.GetVBTaskRsp)
	rsp.VbTask = rs.getVBTasks()
	rs.Send(rsp)
}

// getVBTasks 获取vb任务列表, 在游戏中则只获取游戏关联的
func (rs *RoleActor) getVBTasks() []*pb.VBTaskData {
	tasks := handler.BuildVBTaskDataMsg(rs.User)
	if rs.gamePid != nil {
		gtype := rs.gtype
		var ftasks []*pb.VBTaskData
		for _, task := range tasks {
			var types []*pb.VBTaskType
			for _, taskType := range task.TaskTypes {
				if utils.SliceIn(gtype, taskType.Gtype...) {
					types = append(types, taskType)
				}
			}
			if len(types) > 0 {
				task.TaskTypes = types
				ftasks = append(ftasks, task)
			}
		}
		tasks = ftasks
	}
	return tasks
}

// 获取vb银行日志
func (rs *RoleActor) VipBankLogNtf() {
	rsp := new(pb.VipBankLogNtf)

	end := utils.BsonNow().Unix() - 30*24*3600
	for i := len(rs.VBankLog) - 1; i >= 0; i-- {
		vb := rs.VBankLog[i]
		if vb.Date < end {
			break
		}
		bean := &pb.VipBankLog{
			Date:   vb.Date,
			Ltype:  vb.LogType,
			Amount: int32(vb.Amount),
		}
		rsp.Logs = append(rsp.Logs, bean)
	}
	// for _, vb := range rs.VBankLog {
	// 	bean := &pb.VipBankLog{
	// 		Date:   vb.Date,
	// 		Ltype:  vb.LogType,
	// 		Amount: int32(vb.Amount),
	// 	}
	// 	rsp.Logs = append(rsp.Logs, bean)
	// }

	rs.Send(rsp)
}

// 输分补偿查询
func (rs *RoleActor) VBLoseCompensationReq(ctx actor.Context) {
	rsp := new(pb.VBLoseCompensationRsp)
	loseCom := table.GetTables().LoseCompensationTable.Get()

	nextReceive := (int64(loseCom.RewardInterval)*60 - (utils.BsonNow().Unix() - rs.User.VBOutDiamondRewardTime))
	if nextReceive < 0 {
		nextReceive = 0
	}
	nextReceive = 0
	info := &pb.VBLoseCompensation{
		VbOutDiamond: int32(rs.User.VBOutDiamond),
		LoseRate:     loseCom.Compensate,
		Withdrawal:   loseCom.Withdrawal,
		CanReward:    rs.User.VBOutDiamond >= int64(loseCom.Withdrawal) && rs.User.VBOutDiamondDayTimes < loseCom.DayLimit && nextReceive <= 0,
		NextReceive:  nextReceive, // 下次领取时间
	}
	rsp.Info = info
	rs.Send(rsp)
}

// 领取输分补偿
func (rs *RoleActor) VBLoseCompensationRewardReq(ctx actor.Context) {
	rsp := new(pb.VBLoseCompensationRewardRsp)

	loseCom := table.GetTables().LoseCompensationTable.Get()
	// 每次领100
	reward := int64(loseCom.Withdrawal)
	// 累计输分不足
	if rs.User.VBOutDiamond < reward {
		rsp.Error = pb.NotEnoughCoin
		rs.Send(rsp)
		return
	}
	// 每日领取次数
	if rs.User.VBOutDiamondDayTimes >= loseCom.DayLimit {
		rsp.Error = pb.CanNotReceive
		rs.Send(rsp)
		return
	}

	now := utils.BsonNow().Unix()
	minute := (now - rs.User.VBOutDiamondRewardTime) / 60
	if minute > 0 && minute < int64(loseCom.RewardInterval) {
		rsp.Error = pb.CanNotReceive
		rs.Send(rsp)
		return
	}

	// 释放VB金不足
	// if loseCom.Mod == 2 && rs.User.VBBank < reward {
	// 	rsp.Error = pb.NotEnoughCoin
	// 	rs.Send(rsp)
	// 	return
	// }

	// 累计到配置额度从VB奖池取出
	if loseCom.Mod == 2 {
		// 释放
		rs.sendGood(reward, 0, -reward, 0, 0, 0, int32(pb.LOG_TYPE126), "输分补偿释放VB")
	} else {
		// 增长
		rs.sendGood(0, 0, reward, 0, 0, 0, int32(pb.LOG_TYPE125), "输分补偿增长VB")
	}
	rs.User.VBOutDiamond -= reward
	rs.User.VBOutDiamondDayTimes++       // 领取次数+1
	rs.User.VBOutDiamondRewardTime = now // 记录最后领取时间
	rs.status = true

	// 前端同步
	msg := &pb.ActivityDataNtf{
		VbLoseCompensation: handler.BuildVBLoseCompensationMsg(rs.User),
	}
	rs.Send(msg)

	nextReceive := (int64(loseCom.RewardInterval)*60 - (utils.BsonNow().Unix() - rs.User.VBOutDiamondRewardTime))
	if nextReceive < 0 {
		nextReceive = 0
	}
	nextReceive = 0
	rsp.Info = &pb.VBLoseCompensation{
		VbOutDiamond: int32(rs.User.VBOutDiamond),
		LoseRate:     loseCom.Compensate,
		Withdrawal:   loseCom.Withdrawal,
		CanReward:    rs.User.VBOutDiamond >= int64(loseCom.Withdrawal) && rs.User.VBOutDiamondDayTimes < loseCom.DayLimit && nextReceive <= 0,
		NextReceive:  nextReceive,
	}
	rs.Send(rsp)
}

func (rs *RoleActor) WithdrawalPopReq(ctx actor.Context) {
	rsp := new(pb.WithdrawalPopRsp)
	if ok, p := handler.CheckNewbieWithdrawFlag(rs.User); !ok &&
		utils.BsonNow().Unix() >= rs.WithdrawPOPTime+int64(table.GetTables().TransferDemoCashTable.Get().Cooldown)*60 {
		rs.GetWithdrawFlag()
		rsp.Amount = p
	}
	rs.Send(rsp)
}

func (rs *RoleActor) VipBonusReq(ctx actor.Context) {
	rsp := new(pb.VipBonusRsp)

	defer rs.Send(rsp)
	bean := table.GetTables().VipBonusTable.Get(int32(rs.Vip.Lv))
	unlock := bean.OnceAmount[rs.RegistArea]
	if unlock <= 0 {
		return
	}

	if rs.UnlockBonus < int64(unlock) {
		rsp.Error = pb.CanNotReceive
		return
	}

	// 全部领
	receive := (rs.UnlockBonus / int64(unlock)) * int64(unlock)
	// 不能全部领，每次只领一个
	// receive := int64(unlock)

	rs.sendGood(int64(receive), 0, -int64(receive), 0, 0, 0, int32(pb.LOG_TYPE132), fmt.Sprintf("vip:%d vb提取", rs.Vip.Lv))

	rs.UnlockBonus -= receive
	rs.CashOutBonus += receive

	rsp.Bonus = rs.UnlockBonus
	rsp.CashOutBonus = rs.CashOutBonus
	rs.status = true
}

func (rs *RoleActor) VipPopReq(ctx actor.Context) {
	rs.Vip.BeforeLv = rs.Vip.Lv
	rs.status = true
	rs.Send(new(pb.VipPopRsp))
}

func (rs *RoleActor) vbFlow(score int64) {
	if score == 0 || rs.Vip.Lv <= 0 {
		return
	}

	score = int64(math.Abs(float64(score)))

	bean := table.GetTables().VipBonusTable.Get(int32(rs.Vip.Lv))
	coverRate := bean.ConverRate[rs.RegistArea]
	cover := score * int64(coverRate) / 10000
	rs.UnlockBonus += cover
	rs.checkUnlockCash()
	rs.status = true
}

func (rs *RoleActor) checkUnlockCash() {
	if rs.Vip.Lv <= 0 {
		return
	}
	bean := table.GetTables().VipBonusTable.Get(int32(rs.Vip.Lv))
	if rs.UnlockBonus > rs.VBBank {
		rs.UnlockBonus = rs.VBBank
	}
	msg := &pb.VipBonusNtf{
		VbCash: rs.UnlockBonus,
	}
	// 提醒
	onceAmount := int64(bean.OnceAmount[rs.RegistArea])
	if rs.UnlockBonus >= onceAmount {
		// msg.Times = rs.UnlockBonus / onceAmount
		// receive := onceAmount
		msg.Times = 1
		receive := (rs.UnlockBonus / onceAmount) * onceAmount
		msg.Pop = receive
	}
	rs.Send(msg)
}

// vb bank日志查询
func (rs *RoleActor) VipBankLogRecordReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.VipBankLogRecordReq)
	rsp := new(pb.VipBankLogRecordRsp)

	userid := rs.Userid
	pageSize := arg.PageSize
	if pageSize <= 0 {
		pageSize = 100
	} else if pageSize > 200 {
		pageSize = 200
	}

	if arg.Month == "" {
		rsp.Error = pb.Failed
		rs.Send(rsp)
		return
	}

	// 筛选日期
	month, err := time.ParseInLocation(utils.FORMAT_MONTH, arg.Month, location)
	if err != nil {
		rsp.Error = pb.Failed
		rs.Send(rsp)
		return
	}
	rsp.Month = arg.Month
	stime := month
	etime := month.AddDate(0, 1, 0)

	args := []any{stime.Unix(), etime.Unix(), userid}

	sql := `
		SELECT id, ctime, reason, vb, after_out_cash
		FROM game.col_log_vbdiamond FINAL WHERE ctime >= ? AND ctime < ? AND userid = ? %s AND vb != 0
		ORDER BY ctime DESC LIMIT ?
	`

	cond := ""
	if arg.PrevLastId != "" {
		cond += " AND ctime < (SELECT ctime FROM game.col_log_vbdiamond FINAL WHERE id = ?)"
		args = append(args, arg.PrevLastId)
	}

	// change: 0.All,1.Income,2.Expense
	switch arg.Change {
	case 1:
		cond += " AND vb > 0"
	case 2:
		cond += " AND vb < 0"
	}

	if arg.Reason != 0 {
		ltypes, notLtypes, sub := handler.GetBonusChangeType2Ltypes(arg.Reason)
		if len(ltypes) > 0 {
			cond += " AND reason IN ?"
			args = append(args, ltypes)
		} else if len(notLtypes) > 0 {
			cond += " AND reason NOT IN ?"
			args = append(args, notLtypes)

			switch sub {
			case 1:
				cond += " AND vb > 0"
			case 2:
				cond += " AND vb < 0"
			}
		}
	}

	args = append(args, pageSize)

	sql = fmt.Sprintf(sql, cond)

	var results []map[string]any
	if err := ck.Select(&results, sql, args...); err != nil {
		glog.Errorf("select user vbbank records error: %s, %v", userid, err)
		rsp.Error = pb.Failed
		rs.Send(rsp)
		return
	}
	for _, r := range results {
		id := r["id"].(string)
		ctime := utils.ToInt64(r["ctime"])
		reason := int32(utils.ToInt64(r["reason"]))
		vb := utils.ToInt64(r["vb"])
		after_out_cash := utils.ToInt64(r["after_out_cash"])

		ctimestr := time.Unix(ctime, 0).In(location).Format(utils.FORMAT)

		record := &pb.VipBankLogRecord{
			Id:         id,
			Date:       ctimestr,
			Reason:     handler.GetBonusChangeType(reason, vb),
			Score:      vb,
			AfterScore: after_out_cash,
		}
		rsp.Records = append(rsp.Records, record)
	}

	rs.Send(rsp)
}

// 打码量排行榜活动榜单全量数据请求
func (rs *RoleActor) ActivityBetRankReq(ctx actor.Context) {
	req := &mq.RequestActivityBetRankListArgs{Userid: rs.Userid, RegistArea: rs.RegistArea}
	rsp := &pb.ActivityBetRankRsp{}
	err := mq.NatsRequest(mq.RequestActivityBetRankList, rsp, req)
	if err != nil {
		rsp.Error = pb.Failed
	}
	rs.Send(rsp)
}

// 排行榜历史请求
func (rs *RoleActor) ActivityBetRankHistoryReq(ctx actor.Context) {
	req := &mq.RequestUserArgs{Userid: rs.Userid}
	rsp := &pb.ActivityBetRankHistoryRsp{}
	err := mq.NatsRequest(mq.RequestActivityBetRankHistory, rsp, req)
	if err != nil {
		rsp.Error = pb.Failed
	}
	rs.Send(rsp)
}

// 个人领取记录请求
func (rs *RoleActor) ActivityBetRankMyRecordReq(ctx actor.Context) {
	req := &mq.RequestUserArgs{Userid: rs.Userid}
	rsp := &pb.ActivityBetRankMyRecordRsp{}
	err := mq.NatsRequest(mq.RequestActivityBetRankMyRecord, rsp, req)
	if err != nil {
		rsp.Error = pb.Failed
	}
	rs.Send(rsp)
}

// 排行榜上榜报喜弹窗 Skip for today
func (rs *RoleActor) ActivityBetRankNewRankSkipTodayReq(ctx actor.Context) {
	req := &mq.RequestActivityBetRankSkipWindowArgs{Userid: rs.Userid, WType: 1}
	rsp := &pb.ActivityBetRankNewRankSkipTodayRsp{}
	err := mq.NatsRequest(mq.RequestActivityBetRankSkipWindow, rsp, req)
	if err != nil {
		rsp.Error = pb.Failed
	}
	rs.Send(rsp)
}

// 快上榜提醒弹窗 Skip for today
func (rs *RoleActor) ActivityBetRankWillRankSkipTodayReq(ctx actor.Context) {
	req := &mq.RequestActivityBetRankSkipWindowArgs{Userid: rs.Userid, WType: 2}
	rsp := &pb.ActivityBetRankWillRankSkipTodayRsp{}
	err := mq.NatsRequest(mq.RequestActivityBetRankSkipWindow, rsp, req)
	if err != nil {
		rsp.Error = pb.Failed
	}
	rs.Send(rsp)
}

// 掉榜弹窗 Skip for today
func (rs *RoleActor) ActivityBetRankLoseRankSkipTodayReq(ctx actor.Context) {
	req := &mq.RequestActivityBetRankSkipWindowArgs{Userid: rs.Userid, WType: 3}
	rsp := &pb.ActivityBetRankLoseRankSkipTodayRsp{}
	err := mq.NatsRequest(mq.RequestActivityBetRankSkipWindow, rsp, req)
	if err != nil {
		rsp.Error = pb.Failed
	}
	rs.Send(rsp)
}

// 排行榜奖励领取
func (rs *RoleActor) ActivityBetRankRewordReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ActivityBetRankRewordReq)
	req := &mq.RequestActivityBetRankRewordArgs{Id: arg.Id, Userid: rs.Userid}
	rsp := &pb.ActivityBetRankRewordRsp{}
	err := mq.NatsRequest(mq.RequestActivityBetRankReword, rsp, req)
	if err != nil {
		rsp.Error = pb.Failed
		rs.Send(rsp)
		return
	}

	// 发送奖励 1代表bonus,2代表cash,3代表withdrawable
	var bonus, diamond, outDiamond int64
	switch rsp.PrizeType {
	case 1:
		bonus = rsp.Prize
	case 2:
		diamond = rsp.Prize
	case 3:
		diamond = rsp.Prize
		outDiamond = rsp.Prize
	}
	rs.sendGood(diamond, 0, bonus, outDiamond, 0, 0, int32(pb.LOG_TYPE134), fmt.Sprintf("打码排行榜奖励领取: %s", req.Id))
	rs.Send(rsp)
}

// 更新favorite
func (rs *RoleActor) UpdateFavoriteReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.UpdateFavoriteReq)
	rs.User.Favorite = arg.Gtype
	rs.status = true
	rsp := &pb.UpdateFavoriteRsp{}
	rs.Send(rsp)
}

// ActivityTurnSync 转盘用户状态同步
func (rs *RoleActor) ActivityTurnSync(ctx actor.Context) {
	arg := ctx.Message().(*pb.ActivityTurnSync)
	if rs.User.ActivityTurn == nil {
		rs.User.ActivityTurn = new(data.ActivityTurn)
	}
	rs.User.ActivityTurn.TurnTime = arg.TurnTime
	rs.User.ActivityTurn.TurnEtime = arg.TurnEtime
	rs.User.ActivityTurn.GiveSelected = arg.GiveSelected
	rs.User.ActivityTurn.GiveAmount = arg.GiveAmount
	rs.User.ActivityTurn.GiveAmount3 = arg.GiveAmount3
	rs.User.ActivityTurn.GiveIndex = arg.GiveIndex
	rs.User.ActivityTurn.DrawedTimes = arg.DrawedTimes
	rs.User.ActivityTurn.DrawTimesFree = arg.DrawTimesFree
	rs.User.ActivityTurn.DrawTimesInvite = arg.DrawTimesInvite
	rs.User.ActivityTurn.NextFreeDrawTime = arg.NextFreeDrawTime
	rs.User.ActivityTurn.Score = arg.Score
	rs.User.ActivityTurn.ScoreTarget = arg.ScoreTarget
	rs.User.ActivityTurn.ScoreType = arg.ScoreType
	rs.User.ActivityTurn.TackedPrize = arg.TackedPrize
	rs.User.ActivityTurn.DrawedTimesFree = arg.DrawedTimesFree
	rs.User.ActivityTurn.DrawedTimesInvite = arg.DrawedTimesInvite
	rs.User.ActivityTurn.DrawTimesLucky = arg.DrawTimesLucky
	rs.User.ActivityTurnPrizeTimes = arg.ActivityTurnPrizeTimes
}

// 转盘主界面请求
func (rs *RoleActor) ActivityTurnReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ActivityTurnReq)
	arg.Userid = rs.Userid
	rs.rolePid.Request(arg, ctx.Self())
}

// 点击序幕礼盒、点击invite按钮记录
func (rs *RoleActor) ActivityTurnClickLogReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ActivityTurnClickLogReq)
	arg.Userid = rs.Userid
	rs.rolePid.Request(arg, ctx.Self())
}

// 转盘抽奖请求
func (rs *RoleActor) ActivityTurnDrawReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ActivityTurnDrawReq)
	arg.Userid = rs.Userid
	rs.rolePid.Request(arg, ctx.Self())
}

// 本轮抽奖记录 滚动分页请求
func (rs *RoleActor) ActivityTurnDrawLogReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ActivityTurnDrawLogReq)
	arg.Userid = rs.Userid
	rs.rolePid.Request(arg, ctx.Self())
}

// 转盘奖励领取
func (rs *RoleActor) ActivityTurnTackPrizeReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ActivityTurnTackPrizeReq)
	arg.Userid = rs.Userid
	rs.rolePid.Request(arg, ctx.Self())
}

// 转盘奖金审核记录
func (rs *RoleActor) ActivityTurnPrizesReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ActivityTurnPrizesReq)
	arg.Userid = rs.Userid
	rs.rolePid.Request(arg, ctx.Self())
}

var (
	// reason: 0全部,1充值,2提现,3游戏返奖,4下注,5奖金福利,7.取消下注
	ltypeReason1 = []int32{57, 58, 59, 74, 80, 84, 86, 96, 119, 129, 139, 152}                                   // 1.充值
	ltypeReason2 = []int32{13, 63, 136, 137, 138}                                                                // 2.提现
	ltypeReason3 = []int32{45, 67, 68, 70, 71, 73, 88, 91, 95, 101, 107, 111, 113, 115, 117, 128, 131, 150, 154} // 3.游戏返奖
	ltypeReason4 = []int32{5, 75, 76, 79, 87, 90, 94, 100, 106, 110, 112, 114, 116, 127, 130, 148, 149, 153}     // 4.下注
	ltypeReason6 = []int32{69, 72, 89, 92, 109}                                                                  // 6.扣税,抽水
	ltypeReason7 = []int32{146, 147, 151}                                                                        // 7.取消下注

	ltypeReason1M = map[int32]bool{57: true, 58: true, 59: true, 74: true, 80: true, 84: true, 86: true, 96: true, 119: true, 129: true, 139: true, 152: true}                                                                             // 1充值
	ltypeReason2M = map[int32]bool{13: true, 63: true, 136: true, 137: true, 138: true}                                                                                                                                                    // 2提现
	ltypeReason3M = map[int32]bool{45: true, 67: true, 68: true, 70: true, 71: true, 73: true, 88: true, 91: true, 95: true, 101: true, 107: true, 111: true, 113: true, 115: true, 117: true, 128: true, 131: true, 150: true, 154: true} // 3游戏返奖
	ltypeReason4M = map[int32]bool{5: true, 75: true, 76: true, 79: true, 87: true, 90: true, 94: true, 100: true, 106: true, 110: true, 112: true, 114: true, 116: true, 127: true, 130: true, 148: true, 149: true, 153: true}           // 4下注
	ltypeReason6M = map[int32]bool{69: true, 72: true, 89: true, 92: true, 109: true}                                                                                                                                                      // 6.扣税,抽水
	ltypeReason7M = map[int32]bool{146: true, 147: true, 151: true}                                                                                                                                                                        // 7.取消下注
)

// BalanceRecordsReq 玩家cash账变记录
func (rs *RoleActor) BalanceRecordsReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.BalanceRecordsReq)
	rsp := new(pb.BalanceRecordsRsp)

	userid := rs.Userid
	pageSize := arg.PageSize
	if pageSize <= 0 {
		pageSize = 100
	} else if pageSize > 200 {
		pageSize = 200
	}

	sql := `
		SELECT id, ctime, ltype, add_diamond, old_diamond, now_diamond
		FROM game.col_log_water FINAL WHERE userid = ? %s AND add_diamond != 0
		ORDER BY id DESC LIMIT ?
	`
	args := []any{userid}

	cond := ""
	if arg.PrevLastId != "" {
		cond += " AND id < ?"
		args = append(args, arg.PrevLastId)
	}
	// 筛选日期
	if arg.Date != "" {
		sdate := fmt.Sprintf("%s 00:00:00", arg.Date)
		stime, err := time.ParseInLocation(utils.FORMAT, sdate, location)
		if err != nil {
			rsp.Error = pb.Failed
			rs.Send(rsp)
			return
		}
		etime := stime.AddDate(0, 0, 1)
		cond += " AND ctime >= ? AND ctime < ?"
		args = append(args, stime, etime)
	}

	// change: 0.All,1.Income,2.Expense
	switch arg.Change {
	case 1:
		cond += " AND add_diamond > 0"
	case 2:
		cond += " AND add_diamond < 0"
	}

	// reason: 0全部,1充值,2提现,3游戏返奖,4下注,5奖金福利,6抽水,7.CancelBet
	switch arg.Reason {
	case 1:
		cond += " AND ltype IN ?"
		args = append(args, ltypeReason1)
	case 2:
		cond += " AND ltype IN ?"
		args = append(args, ltypeReason2)
	case 3:
		cond += " AND ltype IN ?"
		args = append(args, append(ltypeReason3, ltypeReason6...))
	case 4:
		cond += " AND ltype IN ?"
		args = append(args, append(ltypeReason4, ltypeReason7...))
	case 5:
		var ltypes []int32
		ltypes = append(ltypes, ltypeReason1...)
		ltypes = append(ltypes, ltypeReason2...)
		ltypes = append(ltypes, ltypeReason3...)
		ltypes = append(ltypes, ltypeReason4...)
		ltypes = append(ltypes, ltypeReason6...)
		ltypes = append(ltypes, ltypeReason7...)
		cond += " AND ltype NOT IN ?"
		args = append(args, ltypes)
	}

	args = append(args, pageSize)

	sql = fmt.Sprintf(sql, cond)

	var results []map[string]any
	if err := ck.Select(&results, sql, args...); err != nil {
		glog.Errorf("select user balance records error: %s, %v", userid, err)
		rsp.Error = pb.Failed
		rs.Send(rsp)
		return
	}
	for _, r := range results {
		id := r["id"].(string)
		ltype := int32(utils.ToInt64(r["ltype"]))
		add_diamond := utils.ToInt64(r["add_diamond"])
		old_diamond := utils.ToInt64(r["old_diamond"])
		now_diamond := utils.ToInt64(r["now_diamond"])

		var ctimestr string
		if ctime, ok := r["ctime"].(time.Time); ok {
			ctimestr = ctime.In(location).Format(utils.FORMAT)
		}

		var reason int32
		switch true {
		case ltypeReason1M[ltype]:
			reason = 1
		case ltypeReason2M[ltype]:
			reason = 2
		case ltypeReason3M[ltype]:
			reason = 3
		case ltypeReason4M[ltype]:
			reason = 4
		case ltypeReason6M[ltype]:
			reason = 6
		case ltypeReason7M[ltype]:
			reason = 7
		default:
			reason = 5
		}

		record := &pb.BalanceRecord{
			Id:         id,
			Ctime:      ctimestr,
			Reason:     reason,
			AddDiamond: add_diamond,
			OldDiamond: old_diamond,
			NowDiamond: now_diamond,
		}
		rsp.Records = append(rsp.Records, record)
	}

	rsp.HasMore = len(results) >= int(pageSize)
	rs.Send(rsp)
}

// 今日跳过充值礼包
func (rs *RoleActor) SkipGiftWelfareReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.SkipGiftWelfareReq)
	// _ = arg
	if arg.Rtype == 1 {
		rs.User.SkipGiftWelfare = true
	} else if arg.Rtype == 2 {
		rs.User.SkipTurntable = true
	} else if arg.Rtype == 3 {
		rs.User.SkipRank = true
	} else if arg.Rtype == 4 {
		rs.User.SkipGiftCode = true
	}
	rs.status = true
	rsp := &pb.SkipGiftWelfareRsp{}
	rs.Send(rsp)
}

// 点邀请按钮日志
func (rs *RoleActor) LaunchInviteLogReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.LaunchInviteLogReq)
	rsp := new(pb.LaunchInviteLogRsp)
	log := &pb.LaunchInviteLog{
		Userid:     rs.Userid,
		Ltype:      arg.Ltype,
		ChannelId:  rs.AD_BundleId,
		RegistArea: int32(rs.RegistArea),
		Rtime:      rs.Ctime.UnixMilli(),
		Ctime:      time.Now().UnixMilli(),
	}
	myactor.Logger().Tell(log)
	rs.Send(rsp)
}

// 代理信息同步
func (rs *RoleActor) ShareAgentSync(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareAgentSync)
	rs.ShareAgent.TeamLv = arg.TeamLv
	rs.ShareAgent.TeamAgents = arg.TeamAgents
	rs.ShareAgent.TeamBets = arg.TeamBets
	rs.ShareAgent.ShareUsers = arg.ShareUsers
	rs.ShareAgent.ShareUsersPrizes = arg.ShareUsersPrizes
	rs.ShareAgent.ShareSuper = arg.ShareSuper

	rs.ShareAgent.EarningsBetTotal = arg.EarningsBetTotal
	rs.ShareAgent.EarningsBetUnclaimed = arg.EarningsBetUnclaimed
	rs.ShareAgent.EarningsBetPendding = arg.EarningsBetPendding
	rs.ShareAgent.EarningsUserTotal = arg.EarningsUserTotal
	rs.ShareAgent.EarningsUserUnclaimed = arg.EarningsUserUnclaimed
	rs.ShareAgent.EarningsUserPendding = arg.EarningsUserPendding

	rs.ShareAgent.TodayEarningsUser = arg.TodayEarningsUser
	rs.ShareAgent.TodayTakeItypeDates = arg.TodayTakeItypeDates
	// 推送更新
	ntf := new(pb.ShareAgentSyncNtf)
	rs.Send(ntf)
}

// 代理持续受益界面请求
func (rs *RoleActor) ShareOngoingEarningReq(ctx actor.Context) {
	// arg := ctx.Message().(*pb.ShareOngoingEarningReq)
	rsp := new(pb.ShareOngoingEarningRsp)

	rsp.TeamTier = rs.ShareAgent.TeamLv
	rsp.CriteriaSubAgentsNumber = rs.ShareAgent.TeamAgents
	rsp.CriteriaTeamBettingAmount = rs.ShareAgent.TeamBets
	rsp.EarningsTotal = rs.ShareAgent.EarningsBetTotal
	rsp.EarningsUnclaimed = rs.ShareAgent.EarningsBetUnclaimed

	for _, data := range table.GetTables().ShareGroupTable.GetDataList() {
		rsp.TeamLvs = append(rsp.TeamLvs, &pb.ShareAgentRuleTeamLv{
			TeamTier:         data.Level,
			SubAgentNumber:   data.NeedShareUsers,
			SubAgentTurnover: int64(data.NeedShareBets),
			RebateRateLv1:    fmt.Sprintf("%g%%", float64(data.ShareL1Rate)*100/10000),
			RebateRateLv2:    fmt.Sprintf("%g%%", float64(data.ShareL2Rate)*100/10000),
			RebateRateLv3:    fmt.Sprintf("%g%%", float64(data.ShareL3Rate)*100/10000),
		})
	}

	// 当前团队等级
	// shareGroupTable := table.GetTables().ShareGroupTable.GetDataList()
	// var group *tb.ShareShareGroupRecord
	// if rs.ShareAgent.TeamLv == 0 {
	// 	group = shareGroupTable[0]
	// } else {
	// 	group = table.GetTables().ShareGroupTable.Get(rs.ShareAgent.TeamLv)
	// 	if group == nil {
	// 		group = shareGroupTable[len(shareGroupTable)-1]
	// 	}
	// }

	// rsp.CriteriaSubAgentsNumberRequire = group.NeedShareUsers
	// rsp.CriteriaTeamBettingRequire = int64(group.NeedShareBets)
	// rsp.CriteriaSubAgentsNumberAchieved = rsp.CriteriaSubAgentsNumber >= rsp.CriteriaSubAgentsNumberRequire
	// rsp.CriteriaTeamBettingAmountAchieved = rsp.CriteriaTeamBettingAmount >= rsp.CriteriaTeamBettingRequire
	// rsp.RebaseAgents = &pb.ShareAgentRuleRebateRate{
	// 	TeamTier: rs.ShareAgent.TeamLv,
	// 	Lv1:      fmt.Sprintf("%g%%", float64(group.ShareL1Rate)*100/10000),
	// 	Lv2:      fmt.Sprintf("%g%%", float64(group.ShareL2Rate)*100/10000),
	// 	Lv3:      fmt.Sprintf("%g%%", float64(group.ShareL3Rate)*100/10000),
	// }

	rs.Send(rsp)
}

// 领取打码返佣金额
func (rs *RoleActor) ShareOngoingEarningUnclaimedTackReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareOngoingEarningUnclaimedTackReq)
	arg.Userid = rs.Userid
	rs.rolePid.Request(arg, ctx.Self())
}

// 代理一次性收益页面请求
func (rs *RoleActor) ShareInstantBonusReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareInstantBonusReq)
	rsp := new(pb.ShareInstantBonusRsp)
	_ = arg
	rsp.YouCanGet = int64(table.GetTables().ShareEachUserPrizeTable.Get().Share1SuperPrize)
	rsp.TheInviteeCanGet = int64(table.GetTables().ShareEachUserPrizeTable.Get().Share1SelfPrize)

	rsp.EarningsTotal = rs.ShareAgent.EarningsUserTotal
	rsp.EarningsUnclaimed = rs.ShareAgent.EarningsUserUnclaimed

	rsp.ShareUsers = rs.ShareAgent.ShareUsers
	for _, data := range table.GetTables().ShareUsersPrizeTable.GetDataList() {
		rsp.Milestones = append(rsp.Milestones, &pb.ShareInstantBonusMilestone{
			ShareUsers:      data.ShareUsers,
			ShareUsersPrize: int64(data.ShareUsersPrize),
		})
	}
	rs.Send(rsp)
}

// 领取人头和任务返佣金额
func (rs *RoleActor) ShareInstantEarningUnclaimedTackReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareInstantEarningUnclaimedTackReq)
	arg.Userid = rs.Userid
	rs.rolePid.Request(arg, ctx.Self())
}

// 代理收入账单记录 打码返佣 Betting Commission 页面请求
func (rs *RoleActor) ShareAgentIncomeRecordBettingCommissionReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareAgentIncomeRecordBettingCommissionReq)
	rsp := new(pb.ShareAgentIncomeRecordBettingCommissionRsp)

	month, err := time.ParseInLocation(utils.FORMAT_MONTH, arg.Month, location)
	if err != nil {
		rsp.Error = pb.Failed
		rs.Send(rsp)
		return
	}
	rsp.Month = arg.Month

	start := month
	end := month.AddDate(0, 1, 0)

	sql_total := `
		SELECT SUM(amount) amount_sum
		FROM game.col_activity_share_income_record FINAL 
		WHERE ctime >= ? AND ctime < ? AND super_id = ? AND itype = 1
	`
	var args_total = []any{start.UnixMilli(), end.UnixMilli(), rs.Userid}
	var r_total = make(map[string]any)
	if err := ck.Select(&r_total, sql_total, args_total...); err != nil {
		rsp.Error = pb.Failed
		rs.Send(rsp)
		return
	}
	rsp.Total = utils.ToInt64(r_total["amount_sum"])
	rsp.Total /= 100

	sql := `
		SELECT t0.super_id userid, t0.idate idate, t0.itype itype, t0.day_amount, (CASE WHEN t1.ctime > 0 THEN 1 ELSE 0 END) tacked
		FROM (
			SELECT super_id, idate, itype, MAX(ctime) ctime 
			FROM game.col_activity_share_income_record_tack_log FINAL
			WHERE super_id = ? AND itype = 1
			GROUP BY super_id, idate, itype
		) t1 RIGHT JOIN (
			SELECT super_id, idate, itype, SUM(amount) day_amount 
			FROM game.col_activity_share_income_record FINAL 
			WHERE ctime >= ? AND ctime < ? AND super_id = ? AND itype = 1
			GROUP BY super_id, idate, itype
		) t0 ON t0.super_id = t1.super_id AND t0.idate = t1.idate AND t0.itype = t1.itype
		ORDER BY t0.idate DESC
	`
	var args = []any{rs.Userid, start.UnixMilli(), end.UnixMilli(), rs.Userid}
	var results []map[string]any
	if err := ck.Select(&results, sql, args...); err != nil {
		rsp.Error = pb.Failed
		rs.Send(rsp)
		return
	}

	curDate := time.Now().In(location).Format(utils.FORMAT_DATE)
	for _, r := range results {
		idate := r["idate"].(string)
		day_amount := utils.ToInt64(r["day_amount"]) / 100
		tacked := utils.ToInt64(r["tacked"])
		if day_amount <= 0 {
			continue
		}
		record := &pb.ShareAgentIncomeRecordBettingCommission{
			Date:   idate,
			Amount: day_amount,
		}
		if tacked == 1 {
			record.State = 3
		} else if idate == curDate { // 当天结束数据确定后再领取
			record.State = 1
		} else {
			record.State = 2
		}
		rsp.Records = append(rsp.Records, record)
	}

	rs.Send(rsp)
}

// 代理收入账单记录 人头奖励 Referral Bonus 页面请求
func (rs *RoleActor) ShareAgentIncomeRecordReferralBonusReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareAgentIncomeRecordReferralBonusReq)
	rsp := new(pb.ShareAgentIncomeRecordReferralBonusRsp)

	month, err := time.ParseInLocation(utils.FORMAT_MONTH, arg.Month, location)
	if err != nil {
		rsp.Error = pb.Failed
		rs.Send(rsp)
		return
	}
	rsp.Month = arg.Month

	start := month
	end := month.AddDate(0, 1, 0)
	var args_total = []any{start.UnixMilli(), end.UnixMilli(), rs.Userid}
	var args = []any{rs.Userid, start.UnixMilli(), end.UnixMilli(), rs.Userid}

	var filter string
	if utils.SliceIn(arg.Stype, 1, 2) {
		filter = " AND itype = ?"
		args_total = append(args_total, arg.Stype+1)
		args = append(args, arg.Stype+1)
	}

	sql_total := `
		SELECT SUM(amount) amount_sum
		FROM game.col_activity_share_income_record FINAL 
		WHERE ctime >= ? AND ctime < ? AND super_id = ? AND itype IN (2,3) %s
	`
	var r_total = make(map[string]any)
	if err := ck.Select(&r_total, fmt.Sprintf(sql_total, filter), args_total...); err != nil {
		rsp.Error = pb.Failed
		rs.Send(rsp)
		return
	}
	rsp.Total = utils.ToInt64(r_total["amount_sum"])
	rsp.Total /= 100

	sql := `
		SELECT t0.super_id userid, t0.idate idate, t0.itype itype, t0.day_amount, (CASE WHEN t1.ctime > 0 THEN 1 ELSE 0 END) tacked
		FROM (
			SELECT super_id, idate, itype, MAX(ctime) ctime 
			FROM game.col_activity_share_income_record_tack_log FINAL
			WHERE super_id = ? AND itype IN (2,3)
			GROUP BY super_id, idate, itype
		) t1 RIGHT JOIN (
			SELECT super_id, idate, itype, SUM(amount) day_amount 
			FROM game.col_activity_share_income_record FINAL 
			WHERE ctime >= ? AND ctime < ? AND super_id = ? AND itype IN (2,3) %s
			GROUP BY super_id, idate, itype
		) t0 ON t0.super_id = t1.super_id AND t0.idate = t1.idate AND t0.itype = t1.itype
		ORDER BY t0.idate DESC
	`
	var results []map[string]any
	if err := ck.Select(&results, fmt.Sprintf(sql, filter), args...); err != nil {
		rsp.Error = pb.Failed
		rs.Send(rsp)
		return
	}

	curDate := time.Now().In(location).Format(utils.FORMAT_DATE)
	for _, r := range results {
		idate := r["idate"].(string)
		day_amount := utils.ToInt64(r["day_amount"]) / 100
		tacked := utils.ToInt64(r["tacked"])
		itype := utils.ToInt64(r["itype"])
		if day_amount <= 0 {
			continue
		}
		record := &pb.ShareAgentIncomeRecordReferralBonus{
			Date:   idate,
			Amount: day_amount,
			Reason: int32(itype - 1),
		}
		if tacked == 1 {
			record.State = 3
		} else if idate == curDate { // 当天结束数据确定后再领取
			record.State = 1
		} else {
			record.State = 2
		}
		rsp.Records = append(rsp.Records, record)
	}

	rs.Send(rsp)
}

// 代理收入账单记录 打码返佣/人头奖励 unclaimed 奖励领取
func (rs *RoleActor) ShareAgentIncomeTackReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareAgentIncomeTackReq)
	arg.Userid = rs.Userid
	rs.rolePid.Request(arg, ctx.Self())
}

// 代理收入账单记录 代理明细 Agent Details 页面请求
func (rs *RoleActor) ShareAgentIncomeRecordAgentDetailsReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareAgentIncomeRecordAgentDetailsReq)
	rsp := new(pb.ShareAgentIncomeRecordAgentDetailsRsp)

	month, err := time.ParseInLocation(utils.FORMAT_MONTH, arg.Month, location)
	if err != nil {
		rsp.Error = pb.Failed
		rs.Send(rsp)
		return
	}

	rsp.TeamTier = rs.ShareAgent.TeamLv
	rsp.TeamMembers = rs.ShareAgent.TeamAgents

	now := time.Now().In(location)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekday := now.Weekday()
	offset := int(weekday) - 1
	if weekday == time.Sunday {
		offset = 6
	}
	weekStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -offset)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// 团队人数
	sql_lv_total := `
		WITH RECURSIVE share_cte AS (
			SELECT userid, ctime, 1 lv FROM game.col_user FINAL WHERE share_superior = ?
			UNION ALL
			SELECT t.userid, t.ctime, cte.lv+1 lv
			FROM game.col_user t FINAL 
			JOIN share_cte cte ON t.share_superior = cte.userid AND cte.lv < 3
		)
		SELECT lv, count(*) total,
			SUM(CASE WHEN ctime >= ? THEN 1 ELSE 0 END) today,
			SUM(CASE WHEN ctime >= ? THEN 1 ELSE 0 END) week,
			SUM(CASE WHEN ctime >= ? THEN 1 ELSE 0 END) month
		FROM share_cte GROUP BY lv
	`
	args_lv_total := []any{rs.Userid, todayStart, weekStart, monthStart}
	var results_lv_total []map[string]any
	if err := ck.Select(&results_lv_total, sql_lv_total, args_lv_total...); err != nil {
		rsp.Error = pb.Failed
		rs.Send(rsp)
		return
	}
	for _, r := range results_lv_total {
		lv := utils.ToInt64(r["lv"])
		total := int32(utils.ToInt64(r["total"]))
		today := int32(utils.ToInt64(r["today"]))
		week := int32(utils.ToInt64(r["week"]))
		month := int32(utils.ToInt64(r["month"]))
		rsp.AddMembersToday = today
		rsp.AddMembersThisWeek = week
		rsp.AddMembersMonth = month
		switch lv {
		case 1:
			rsp.Lv1Count = total
		case 2:
			rsp.Lv2Count = total
		case 3:
			rsp.Lv3Count = total
		}
	}

	// 升降序, 分, 月份参数
	start := month
	end := month.AddDate(0, 1, 0)
	var sort, asc = "t0.ctime", "DESC"
	if arg.Sort == 2 {
		sort = "t1.commission"
	}
	if arg.Asc {
		asc = "ASC"
	}
	orderBy := fmt.Sprintf("%s %s", sort, asc)

	args := []any{rs.Userid, arg.Lv, start, end, rs.Userid}
	var limitFilter string
	if arg.PrevLastUserid != "" {
		if arg.Sort == 1 {
			limitFilter = fmt.Sprintf(" AND t0.ctime %s (SELECT ctime FROM game.col_user FINAL WHERE userid = ?)", utils.CaseElse(arg.Asc, ">", "<"))
			args = append(args, arg.PrevLastUserid)
		} else {
			limitFilter = fmt.Sprintf(" AND t1.commission %s (SELECT SUM(amount) commission FROM game.col_activity_share_income_record FINAL where super_id = ? and userid = ?)", utils.CaseElse(arg.Asc, ">", "<"))
			args = append(args, rs.Userid, arg.PrevLastUserid)
		}
	}
	args = append(args, arg.PageSize)

	sql := `
		SELECT t0.userid, t0.nickname, t0.photo, t0.money, t0.ctime, t0.lv, t1.commission FROM (
			WITH RECURSIVE share_cte AS (
				SELECT userid, nickname, photo, money, ctime, 1 lv FROM game.col_user FINAL WHERE share_superior = ?
				UNION ALL
				SELECT t.userid, t.nickname, t.photo, t.money, t.ctime, cte.lv+1 lv
				FROM game.col_user t FINAL 
				JOIN share_cte cte ON t.share_superior = cte.userid AND cte.lv < 3
			)
			SELECT userid, nickname, photo, money, ctime, lv FROM share_cte WHERE lv = ? AND ctime >= ? AND ctime < ?
		) t0 LEFT JOIN (
			SELECT userid, SUM(amount) commission FROM game.col_activity_share_income_record FINAL 
			WHERE super_id = ? AND itype = 1
			GROUP BY userid
		) t1 ON t0.userid = t1.userid
		WHERE 1 = 1 %s
		ORDER BY %s
		LIMIT ?
	`
	var results []map[string]any
	if err := ck.Select(&results, fmt.Sprintf(sql, limitFilter, orderBy), args...); err != nil {
		rsp.Error = pb.Failed
		rs.Send(rsp)
		return
	}
	for _, r := range results {
		userid := r["userid"].(string)
		nickname := r["nickname"].(string)
		money := utils.ToInt64(r["money"])
		lv := utils.ToInt64(r["lv"])
		commission := utils.ToInt64(r["commission"]) / 100

		record := &pb.ShareAgentIncomeRecordAgentDetails{
			Userid:        userid,
			Username:      nickname,
			Lv:            int32(lv),
			Commission:    commission,
			DepositAmount: money,
		}
		if ctime, ok := r["ctime"].(time.Time); ok {
			record.JoinTime = ctime.In(location).Format(utils.FORMAT)
		}
		rsp.Records = append(rsp.Records, record)
	}

	rs.Send(rsp)
}

// 代理活动规则请求
func (rs *RoleActor) ShareAgentRulesReq(ctx actor.Context) {
	// arg := ctx.Message().(*pb.ShareAgentRulesReq)
	rsp := new(pb.ShareAgentRulesRsp)

	// 团队等级和返佣比例
	for _, data := range table.GetTables().ShareGroupTable.GetDataList() {
		rsp.RebateRates = append(rsp.RebateRates, &pb.ShareAgentRuleRebateRate{
			TeamTier: data.Level,
			Lv1:      fmt.Sprintf("%g%%", float64(data.ShareL1Rate*100)/10000),
			Lv2:      fmt.Sprintf("%g%%", float64(data.ShareL2Rate*100)/10000),
			Lv3:      fmt.Sprintf("%g%%", float64(data.ShareL3Rate*100)/10000),
		})
		// 团队等级条件表
		rsp.TeamLvs = append(rsp.TeamLvs, &pb.ShareAgentRuleTeamLv{
			TeamTier:         data.Level,
			SubAgentNumber:   data.NeedShareUsers,
			SubAgentTurnover: int64(data.NeedShareBets),
		})
	}

	rsp.Share1PayMin = int64(table.GetTables().ShareTable.Get().Share1PayMin)
	rsp.Share1SuperPrize = int64(table.GetTables().ShareEachUserPrizeTable.Get().Share1SuperPrize)
	rsp.Share1SelfPrize = int64(table.GetTables().ShareEachUserPrizeTable.Get().Share1SelfPrize)

	// 累计人头奖金金额表
	var milestoneRewordsTotal int64
	for _, data := range table.GetTables().ShareUsersPrizeTable.GetDataList() {
		rsp.MilestoneRewords = append(rsp.MilestoneRewords, &pb.ShareAgentRuleMilestoneReword{
			InvitedPayingUsers: data.ShareUsers,
			MilestoneRewords:   int64(data.ShareUsersPrize),
		})
		milestoneRewordsTotal += int64(data.ShareUsersPrize)
	}
	rsp.MilestoneRewordsTotal = milestoneRewordsTotal
	rs.Send(rsp)
}

// 礼包码领取
func (rs *RoleActor) GiftPackCodeTackReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.GiftPackCodeTackReq)
	arg.Userid = rs.Userid
	rs.rolePid.Request(arg, ctx.Self())
}

// 波动返水转盘请求
func (rs *RoleActor) VolatilitySubsidyWheelReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.VolatilitySubsidyWheelReq)
	arg.Userid = rs.Userid
	rs.rolePid.Request(arg, ctx.Self())
}

// 波动返水同步
func (rs *RoleActor) VolatilitySubsidySync(ctx actor.Context) {
	arg := ctx.Message().(*pb.VolatilitySubsidySync)
	rs.VolatilitySubsidyTimes = arg.VolatilitySubsidyTimes
	rs.VolatilitySubsidyAmounts = arg.VolatilitySubsidyAmounts
	rs.VolatilitySubsidyFirstTime = arg.VolatilitySubsidyFirstTime
}

// 客服配置推送
func (rs *RoleActor) customerSetting() {
	customer := table.GetTables().CustomerTable.Get()
	ntf := &pb.CustomerSettingNtf{
		MessageLength: customer.MessageLength,
		FileLength:    customer.FileLength,
		AllowFiles:    customer.AllowFiles,
	}
	for _, q := range table.GetTables().CustomerQuestionTable.GetDataList() {
		ntf.MayAsks = append(ntf.MayAsks, q.Question)
	}
	rs.Send(ntf)
}
