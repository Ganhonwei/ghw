package dbms

import (
	"context"
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/config"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"goserver/pkg/myredis"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"sort"
	"strconv"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/gogo/protobuf/proto"
	"gopkg.in/mgo.v2/bson"
)

// 同步首充奖励
func (a *RoleActor) FirstRecharged(ctx actor.Context) {
	// 首充领奖同步数据
	arg := ctx.Message().(*pb.FirstRecharged)
	user := a.getUserById(arg.Userid)
	if user == nil {
		glog.Errorf("no found user,id:%s", arg.Userid)
		return
	}
	user.Receive = arg.Receive
	user.ReceiveDay = arg.Day
	user.UpdateFirstRecharge()
}

// 周卡数据同步
func (a *RoleActor) WeeklyCardReceiveReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.WeeklyCardReceiveReq)
	glog.Debugf("WeeklyCardReceiveReq %#v", arg)
	user := a.getUserById(arg.GetUserid())
	if user == nil {
		glog.Errorf("user is nil, id:%s", arg.Userid)
		return
	}
	card := user.WeeklyCardMap[arg.GetId()]
	card.Get = false
	card.ReceiveTimes++
	user.UpdateWeekCard()
}

// 在线奖励
func (a *RoleActor) OnlineRewardRecord(ctx actor.Context) {
	arg := ctx.Message().(*pb.OnlineRewardRecord)
	glog.Debugf("OnlineRewardRecord %#v", arg)
	user := a.getUserById(arg.UserId)
	if user == nil {
		glog.Errorf("user is nil, id:%s", arg.UserId)
		return
	}
	user.OnlineReward = append(user.OnlineReward, arg.Id)
	user.OnlineTime = uint64(arg.Otime)
	user.UpdateOnlineReward()
}

// 签到数据
func (a *RoleActor) DailySignReceiveReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.DailySignReceiveReq)
	glog.Debugf("DailySignReceiveReq %#v", arg)
	user := a.getUserById(arg.UserId)
	if user == nil {
		glog.Errorf("user is nil, id:%s", arg.UserId)
		return
	}
	user.SignDay |= (1 << arg.Day)
	user.UpdateDailySign()
}

// 分享结算
func (a *RoleActor) ShareSettlement(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareSettlement)
	glog.Debugf("ShareSettlement %#v", arg)
	if r, ok := a.roles[arg.Superior]; ok {
		r.Pid.Tell(arg)
		return
	}
	user := a.getUserById(arg.Superior)
	if user == nil {
		return
	}
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
	// 同步db
	user.UpdateShare()
	// 再发送给上级
	if user.ShareSuperior == "" {
		return
	}
	// shareBean := table.GetTables().ShareConfigTable.Get()
	shareBean := config.GetShare(1)
	if shareBean.Id == 0 || arg.Level > 1 {
		return
	}
	arg.Score = arg.Total * int64(shareBean.SecondReward) / 10000
	arg.Superior = user.ShareSuperior
	rolePid.Tell(arg)
}

// share充值
func (a *RoleActor) ShareRecharge(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ShareRecharge)
	glog.Debugf("ShareRecharge %#v", arg)
	if r, ok := a.roles[arg.Superior]; ok {
		r.Pid.Tell(arg)
		return
	}
	user := a.getUserById(arg.Superior)
	if user == nil {
		return
	}
	if b, ok := user.ShareBelow[arg.Userid]; ok {
		b.Recharge += int64(arg.Money)
		user.ShareBelow[arg.Userid] = b
		user.UpdateShare()
	}
}

// 分享注册
func (a *RoleActor) ShareRegist(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ShareRegist)
	glog.Debugf("ShareRegist %#v", arg)

	user := a.getUser(arg.Userid)
	if user == nil {
		glog.Errorf("no found online user, user:%s", arg.Userid)
		return
	}
	user.ShareSuperior = arg.Superior
	user.ShareSource = arg.ShareSource
	a.shareRegist(user)
}

func (a *RoleActor) GetUserInfoReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.GetUserInfoReq)
	glog.Debugf("GetUserInfoReq %#v", arg)

	if v, ok := a.roles[arg.Uid]; ok {
		res, err := v.Pid.RequestFuture(arg, 5*time.Second).Result()
		if err != nil {
			rsp := &pb.GetUserInfoRsp{Err: err.Error()}
			ctx.Respond(rsp)
			return
		}
		ctx.Respond(res)
	} else {
		rsp := &pb.GetUserInfoRsp{Err: "player is not online"}
		ctx.Respond(rsp)
		return
		// user := a.getUserById(arg.Uid)
		// rsp := &pb.GetUserInfoRsp{}
		// if user == nil {
		// 	rsp.Err = "player is not exists"
		// 	ctx.Respond(rsp)
		// 	return
		// }
		// rsp.NickName = user.Nickname
		// rsp.Balance = user.Diamond
		// rsp.Photo = user.Photo
		// rsp.Recharge = int64(user.Money)
		// rsp.UserType = int32(user.RegistArea)
		// ctx.Respond(rsp)
		// return
	}
}

func (a *RoleActor) ExternalBetReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ExternalBetReq)
	glog.Debugf("ExternalBetRsp %#v", arg)

	// 打码量日志
	if arg.Amount != 0 {
		otype := handler.GetExternalGameType(int(arg.GameId))
		log := handler.GameFlowWaterLog(arg.Amount, int(arg.GameId), otype, 0, arg.Uid)
		myactor.Logger().Tell(log)

		// 打码上报
		if role := a.getUser(arg.Uid); role != nil {
			pubBet := &pb.PublishGameBets{
				Userid:     arg.Uid,
				Gtype:      arg.GameId,
				Otype:      int32(otype),
				Ts:         time.Now().Unix(),
				Bets:       arg.Amount,
				Robot:      role.Robot,
				Username:   role.Nickname,
				Photo:      role.Photo,
				RegistArea: int32(role.RegistArea),
				VipLv:      int32(role.Vip.Lv),
				SuperId:    role.ShareSuperior,
				WaterId:    arg.RoundId,
			}
			if v, ok := a.roles[arg.Uid]; ok {
				pubBet.UserPid = v.Pid
			}
			if err = mq.NatsPublish(mq.TopicGameBets, pubBet); err != nil {
				glog.Error("publish user external bets error", err)
			}

			// 外接打码上报
			if err = mq.NatsPublish(mq.TopicExternalBet, &pb.PublishExternalBet{
				Userid:     arg.Uid,
				RoundId:    arg.RoundId,
				Bets:       arg.Amount,
				RegistArea: int32(role.RegistArea),
			}); err != nil {
				glog.Error("publish external bets error", err)
			}
		}
	}

	if v, ok := a.roles[arg.Uid]; ok {
		res, err := v.Pid.RequestFuture(arg, 5*time.Second).Result()
		if err != nil {
			rsp := &pb.ExternalBetRsp{ErrCode: 500, Err: err.Error()}
			ctx.Respond(rsp)
			return
		}
		ctx.Respond(res)
	} else {
		// rsp := &pb.ExternalBetRsp{Err: "player is not online"}
		// ctx.Respond(rsp)
		a.ExternalOfflineBet(ctx)
		return
	}
}

func (a *RoleActor) ExternalRewardReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ExternalRewardReq)
	glog.Debugf("ExternalRewardReq %#v", arg)

	defer func() {
		if arg.RewardAmount != 0 {
			// 外接返奖上报
			if role := a.getUser(arg.Uid); role != nil {
				if err = mq.NatsPublish(mq.TopicExternalReword, &pb.PublishExternalReward{
					Userid:     arg.Uid,
					RoundId:    arg.RoundId,
					Amount:     arg.RewardAmount,
					RegistArea: int32(role.RegistArea),
				}); err != nil {
					glog.Error("publish external rewords error", err)
				}
			}
		}
	}()

	if v, ok := a.roles[arg.Uid]; ok {
		res, err := v.Pid.RequestFuture(arg, 1*time.Second).Result()
		if err != nil {
			rsp := &pb.ExternalRewardRsp{Err: err.Error()}
			ctx.Respond(rsp)
			return
		}
		ctx.Respond(res)

		// 外接玩游戏抽奖
		user := a.getUserById(arg.Uid)
		ntf := event.Event(user, event.PLAY_SHARE, &event.PlayShareEvent{Ptype: 1})
		user.UpdatePlayShare()
		if ntf != nil {
			v.Pid.Tell(ntf)
		}
	} else {
		// rsp := &pb.ExternalRewardRsp{Err: "player is not online"}
		// ctx.Respond(rsp)
		a.ExternalOfflineReward(ctx)

		// 外接玩游戏抽奖
		user := a.getUserById(arg.Uid)
		event.Event(user, event.PLAY_SHARE, &event.PlayShareEvent{Ptype: 1})
		user.UpdatePlayShare()
		return
	}
}

func (a *RoleActor) ExternalCancelReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ExternalCancelReq)
	glog.Debugf("ExternalCancelReq %#v", arg)

	defer func() {
		if arg.OrderAmount != 0 {
			// 外接撤单上报
			if role := a.getUser(arg.Uid); role != nil {
				if err = mq.NatsPublish(mq.TopicExternalCancel, &pb.PublishExternalCancel{
					Userid:     arg.Uid,
					RoundId:    arg.RoundId,
					Type:       arg.Type,
					Amount:     arg.OrderAmount,
					RegistArea: int32(role.RegistArea),
				}); err != nil {
					glog.Error("publish external rewords error", err)
				}
			}
		}
	}()

	if arg.Type == 1 { //投注撤单
		//判断订单状态
		l := &data.NsqLogExternalBet{}
		l.UserId = arg.Uid
		l.GameId = arg.GameId
		l.RoundId = arg.RoundId
		l.Get(arg.CancelOrderNo, arg.CancelPlatOrderNo)
		if l.Id == "" {
			rsp := &pb.ExternalCancelRsp{Err: "the order cannot be found"}
			ctx.Respond(rsp)
			return
		}

		var balance int64
		//玩家在线，请求gate处理
		if v, ok := a.roles[arg.Uid]; ok && v != nil {
			res, err := v.Pid.RequestFuture(&pb.ExternalModifyCurrencyReq{Diamond: arg.OrderAmount, Type: int32(pb.LOG_TYPE93)}, 5*time.Second).Result()
			if err != nil {
				rsp := &pb.ExternalCancelRsp{Err: "modify user currency error"}
				ctx.Respond(rsp)
				return
			}
			rsp := res.(*pb.ExternalCancelRsp)
			if rsp.Err != "" {
				rsp := &pb.ExternalCancelRsp{Err: "modify user currency error"}
				ctx.Respond(rsp)
				return
			}
			balance = rsp.Balance
		} else {
			//玩家不在线，直接修改db
			var ok bool
			if ok, balance = a.externalModifyCurrency(arg.OrderAmount, int32(pb.LOG_TYPE93), l.UserId); !ok {
				rsp := &pb.ExternalCancelRsp{Err: "modify user currency error"}
				ctx.Respond(rsp)
				return
			}
		}

		//修改订单状态
		l.CancelBet()

		//响应消息
		id, _ := strconv.ParseUint(l.UserId, 10, 32)
		rsp := &pb.ExternalCancelRsp{
			MerchantOrderNo: strconv.FormatInt(handler.GenerateOrderId(uint32(id)), 10),
			OrderNo:         arg.OrderNo,
			Balance:         balance,
		}
		ctx.Respond(rsp)
		//日志
		cancelLog := &pb.NsqLogExternalCancel{
			Type:              arg.Type,
			Uid:               arg.Uid,
			GameId:            arg.GameId,
			RoundId:           arg.RoundId,
			MerchantOrderNo:   rsp.MerchantOrderNo,
			OrderNo:           rsp.OrderNo,
			CancelOrderNo:     arg.CancelOrderNo,
			CancelPlatOrderNo: arg.CancelPlatOrderNo,
			OrderAmount:       arg.OrderAmount,
			OrderDesc:         arg.OrderDesc,
		}
		body, _ := proto.Marshal(cancelLog)
		log := &pb.NsqLog{
			Typ:  pb.ExternalCancel,
			Body: body,
		}
		// body2, _ := proto.Marshal(log)
		// producer.Publish(data.TopicLog, body2)
		mq.NatsPublish(mq.TopicReportLog, log)

	} else if arg.Type == 2 { //派奖撤单
		//判断订单状态
		l := &data.NsqLogExternalReward{}
		l.UserId = arg.Uid
		l.GameId = arg.GameId
		l.RoundId = arg.RoundId
		l.Get(arg.CancelOrderNo, arg.CancelPlatOrderNo)
		if l.Id == "" {
			rsp := &pb.ExternalCancelRsp{Err: "the order cannot be found"}
			ctx.Respond(rsp)
			return
		}

		var balance int64
		//玩家在线，请求gate处理
		if v, ok := a.roles[arg.Uid]; ok && v != nil {
			res, err := v.Pid.RequestFuture(&pb.ExternalModifyCurrencyReq{Diamond: -arg.OrderAmount, Type: int32(pb.LOG_TYPE93)}, 5*time.Second).Result()
			if err != nil {
				rsp := &pb.ExternalCancelRsp{Err: "modify user currency error"}
				ctx.Respond(rsp)
				return
			}
			rsp := res.(*pb.ExternalCancelRsp)
			if rsp.Err != "" {
				rsp := &pb.ExternalCancelRsp{Err: "modify user currency error"}
				ctx.Respond(rsp)
				return
			}
			balance = rsp.Balance
		} else {
			// 玩家不在线，直接修改db
			var ok bool
			if ok, balance = a.externalModifyCurrency(-arg.OrderAmount, int32(pb.LOG_TYPE93), l.UserId); !ok {
				rsp := &pb.ExternalCancelRsp{Err: "modify user currency error"}
				ctx.Respond(rsp)
				return
			}
		}

		//修改订单状态
		l.CancelReward()

		// 响应消息
		id, _ := strconv.ParseUint(l.UserId, 10, 32)
		rsp := &pb.ExternalCancelRsp{
			MerchantOrderNo: strconv.FormatInt(handler.GenerateOrderId(uint32(id)), 10),
			OrderNo:         arg.OrderNo,
			Balance:         balance,
		}
		ctx.Respond(rsp)
		//日志
		cancelLog := &pb.NsqLogExternalCancel{
			Type:              arg.Type,
			Uid:               arg.Uid,
			GameId:            arg.GameId,
			RoundId:           arg.RoundId,
			MerchantOrderNo:   rsp.MerchantOrderNo,
			OrderNo:           rsp.OrderNo,
			CancelOrderNo:     arg.CancelOrderNo,
			CancelPlatOrderNo: arg.CancelPlatOrderNo,
			OrderAmount:       arg.OrderAmount,
			OrderDesc:         arg.OrderDesc,
		}
		body, _ := proto.Marshal(cancelLog)
		log := &pb.NsqLog{
			Typ:  pb.ExternalCancel,
			Body: body,
		}
		// body2, _ := proto.Marshal(log)
		// producer.Publish(data.TopicLog, body2)
		mq.NatsPublish(mq.TopicReportLog, log)
	} else {
		rsp := &pb.ExternalCancelRsp{Err: "error cancel type"}
		ctx.Respond(rsp)
		return
	}

	// if v, ok := a.roles[arg.Uid]; ok {
	// 	res, err := v.Pid.RequestFuture(arg, 5*time.Second).Result()
	// 	if err != nil {
	// 		rsp := &pb.ExternalRewardRsp{Err: err.Error()}
	// 		ctx.Respond(rsp)
	// 		return
	// 	}
	// 	ctx.Respond(res)
	// } else {
	// 	rsp := &pb.ExternalRewardRsp{Err: "player is not online"}
	// 	ctx.Respond(rsp)
	// 	return
	// }
}

var (
	efiTransactionLogBeginSec int64
	efiTransKeyCur            int       = 0
	efiTransKeys              [2]string = [2]string{"efi:trans_a", "efi:trans_b"}
	efiTransKeyCleanSec       int64
)

// 如果TransactionID不为空且在运营方系统中存在，请返回重复交易
// TransactionID redis cache
// func efiTransactionIdRepeat(arg *pb.EfiTransactionReq) (repeat bool) {
// 	nowSec := time.Now().Unix()
// 	if efiTransactionLogBeginSec == 0 {
// 		efiTransactionLogBeginSec = nowSec
// 		efiTransKeyCleanSec = int64((time.Hour * 12).Seconds())
// 		// efiTransKeyCleanSec = int64((time.Minute).Seconds())
// 	}

// 	ctx, cencel := context.WithTimeout(context.Background(), time.Second*5)
// 	defer cencel()

// 	// 隔12个小时清一下
// 	if nowSec-efiTransactionLogBeginSec > efiTransKeyCleanSec {
// 		efiTransactionLogBeginSec = nowSec
// 		efiTransKeyCur = (efiTransKeyCur + 1) % len(efiTransKeys)
// 		_, err := client.Del(ctx, efiTransKeys[efiTransKeyCur]).Result()
// 		if err != nil {
// 			glog.Errorf("efi trans repeat check clean error: %v", err)
// 		}
// 	}

// 	if len(arg.Transactions) == 0 {
// 		return
// 	}
// 	var transIdTimes []interface{}
// 	var transIds []string
// 	for _, trans := range arg.Transactions {
// 		transIdTimes = append(transIdTimes, trans.TransactionID, "1")
// 		transIds = append(transIds, trans.TransactionID)
// 	}
// 	for i := 0; i < len(efiTransKeys); i++ {
// 		times, err := client.HMGet(ctx, efiTransKeys[i], transIds...).Result()
// 		if err != nil {
// 			glog.Errorf("efi trans repeat check %d error: %v, %v", i, err, transIds)
// 			return
// 		}
// 		for _, v := range times {
// 			if v != nil {
// 				return true
// 			}
// 		}
// 	}

// 	// 记录到 redis
// 	_, err = client.HMSet(ctx, efiTransKeys[efiTransKeyCur], transIdTimes...).Result()
// 	if err != nil {
// 		glog.Errorf("efi transaction repeat check redis save error: %v, %v", err, transIds)
// 		return
// 	}
// 	return
// }

// func (a *RoleActor) EfiTransactionReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.EfiTransactionReq)
// 	glog.Debugf("EfiTransactionReq %#v", arg)

// 	// 重复交易判断
// 	if efiTransactionIdRepeat(arg) {
// 		glog.Errorf("transaction repeat: %v", arg)
// 		rsp := &pb.EfiTransactionRsp{ErrCode: 1003, Err: "API Duplicate Transaction"}
// 		ctx.Respond(rsp)
// 		return
// 	}

// 	var transType string
// 	switch arg.TransType {
// 	case pb.EfiPlaceBet:
// 		transType = "投注"
// 	case pb.EfiGameResult:
// 		transType = "游戏结果"
// 	case pb.EfiRollback:
// 		transType = "回滚"
// 	case pb.EfiCancelBet:
// 		transType = "取消下注"
// 	case pb.EfiBonus:
// 		transType = "奖金/红利"
// 	case pb.EfiJackpot:
// 		transType = "奖池"
// 	case pb.EfiBuyIn:
// 		transType = "投注"
// 	case pb.EfiBuyOut:
// 		transType = "买断"
// 	default:
// 		transType = "交易"
// 	}
// 	arg.TransTypeName = transType

// 	uid := arg.MemberName
// 	if v, ok := a.roles[uid]; ok {
// 		res, err := v.Pid.RequestFuture(arg, 5*time.Second).Result()
// 		if err != nil {
// 			rsp := &pb.EfiTransactionRsp{Err: err.Error()}
// 			ctx.Respond(rsp)
// 			return
// 		}
// 		ctx.Respond(res)
// 	} else {
// 		// offline transaction
// 		user := a.getUserById(uid)
// 		if user == nil {
// 			rsp := &pb.EfiTransactionRsp{Err: "invalid user"}
// 			ctx.Respond(rsp)
// 			return
// 		}
// 		beforeBalance := float64(user.GetDiamond()) / 100.0
// 		var amountTotal float64
// 		for _, trans := range arg.Transactions {
// 			amountTotal += trans.TransactionAmount
// 		}
// 		amount := int64(amountTotal * 100.0)
// 		if amount > 0 && amount > user.GetDiamond() {
// 			rsp := &pb.EfiTransactionRsp{Err: "insufficient balance"}
// 			ctx.Respond(rsp)
// 			return
// 		}
// 		desc := fmt.Sprintf("视讯-%s", arg.TransTypeName)
// 		// 先加再减, 避免不够扣
// 		for _, trans := range arg.Transactions {
// 			if trans.TransactionAmount > 0 {
// 				amount := int64(trans.TransactionAmount * 100)
// 				a.syncCurrency(amount, 0, 0, 0, 0, int32(pb.LOG_TYPE105), uid, desc, trans.TransactionID, false)
// 			}
// 		}
// 		for _, trans := range arg.Transactions {
// 			if trans.TransactionAmount < 0 {
// 				amount := int64(trans.TransactionAmount * 100)
// 				a.syncCurrency(amount, 0, 0, 0, 0, int32(pb.LOG_TYPE105), uid, desc, trans.TransactionID, false)
// 			}
// 		}
// 		rsp := &pb.EfiTransactionRsp{
// 			BeforeBalance: beforeBalance,
// 			Balance:       float64(user.GetDiamond()) / 100.0,
// 		}
// 		ctx.Respond(rsp)
// 	}

// 	//日志
// 	body, _ := proto.Marshal(arg)
// 	log := &pb.NsqLog{
// 		Typ:  pb.EfiTransactionLog,
// 		Body: body,
// 	}
// 	// body2, _ := proto.Marshal(log)
// 	// producer.Publish(data.TopicLog, body2)
// 	mq.NatsPublish(mq.TopicReportLog, log)
// }

func (a *RoleActor) BugFeedbackReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.BugFeedbackReq)
	glog.Debugf("BugFeedbackReq %#v", arg)
	if len(arg.Id) <= 0 {
		return
	}
	log := &pb.LogBugFeedback{
		Userid: arg.Userid,
		Bugid:  arg.Id,
	}
	myactor.Logger().Tell(log)
}

// 外接离线下注
func (a *RoleActor) ExternalOfflineBet(ctx actor.Context) {
	arg := ctx.Message().(*pb.ExternalBetReq)
	glog.Debugf("ExternalBetReq %#v", arg)

	user := a.getUserById(arg.Uid)
	if user == nil {
		rsp := &pb.ExternalBetRsp{Err: "invalid user", ErrCode: 30008}
		ctx.Respond(rsp)
		return
	}
	if arg.Amount > 0 && arg.Amount > user.Diamond {
		rsp := &pb.ExternalBetRsp{Err: "insufficient balance", ErrCode: 30005}
		ctx.Respond(rsp)
		return
	}

	// a.externalModifyCurrency(-arg.Amount, int32(pb.LOG_TYPE94))
	if arg.Amount != 0 {
		a.syncCurrency(-arg.Amount, 0, 0, 0, 0, int32(pb.LOG_TYPE94), arg.Uid, "外接下注", arg.RoundId, false)
		if user.Vip.Lv > 0 {
			// vb
			bean := table.GetTables().VipBonusTable.Get(int32(user.Vip.Lv))
			coverRate := bean.ConverRate[user.RegistArea]
			cover := arg.Amount * int64(coverRate) / 10000
			user.UnlockBonus += cover
			user.UpdateUnlockBonus()
		}
	}
	// a.User.Diamond -= arg.Amount
	// a.status = true

	id, _ := strconv.ParseUint(user.Userid, 10, 32)

	rsp := &pb.ExternalBetRsp{
		MerchantOrderNo: strconv.FormatInt(handler.GenerateOrderId(uint32(id)), 10),
		OrderNo:         arg.OrderNo,
		Balance:         user.Diamond,
	}
	ctx.Respond(rsp)

	//日志
	betLog := &pb.NsqLogExternalBet{
		Uid:             user.Userid,
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

func (a *RoleActor) ExternalOfflineReward(ctx actor.Context) {
	arg := ctx.Message().(*pb.ExternalRewardReq)
	glog.Debugf("ExternalRewardReq %#v", arg)

	user := a.getUserById(arg.Uid)
	if user == nil {
		rsp := &pb.ExternalBetRsp{Err: "invalid user", ErrCode: 30008}
		ctx.Respond(rsp)
		return
	}
	// a.externalModifyCurrency(arg.RewardAmount, int32(pb.LOG_TYPE94))

	if arg.RewardAmount != 0 {
		var out, give int64 = 0, 0
		if arg.RewardAmount > 0 {
			out += arg.RewardAmount
		}
		a.syncCurrency(arg.RewardAmount, 0, give, out, 0, int32(pb.LOG_TYPE95), arg.Uid, "外接结算", arg.RoundId, false)
	}
	if user.OutDiamond > user.Diamond {
		a.syncCurrency(0, 0, 0, user.Diamond-user.OutDiamond, 0, int32(pb.LOG_TYPE95), arg.Uid, "外接扣提现金", arg.RoundId, false)
	}
	// a.User.Diamond += arg.RewardAmount
	// a.status = true

	id, _ := strconv.ParseUint(user.Userid, 10, 32)

	rsp := &pb.ExternalRewardRsp{
		MerchantOrderNo: strconv.FormatInt(handler.GenerateOrderId(uint32(id)), 10),
		OrderNo:         arg.OrderNo,
		Balance:         user.Diamond,
	}
	ctx.Respond(rsp)

	//日志
	rewardLog := &pb.NsqLogExternalReward{
		Uid:             user.Userid,
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

// =================== 汇总计算外接返奖
const (
	gameExternalRound       = "game:external"
	gameExternalRoundExpire = 20 * time.Minute
)

// 外接打码(多次打码)
func handleExternalBet(msg *pb.PublishExternalBet) (err error) {
	ctx, cencel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cencel()
	key := fmt.Sprintf("%s:%s-%s", gameExternalRound, msg.Userid, msg.RoundId)

	if _, err = myredis.Redis().IncrBy(ctx, key, -msg.Bets).Result(); err != nil {
		glog.Errorf("external round incr round bet error: %#v, %v", msg, err)
		return
	}

	if err := myredis.Redis().Expire(ctx, key, gameExternalRoundExpire).Err(); err != nil {
		glog.Errorf("external round set expire error: %#v, %v", msg, err)
	}
	return nil
}

// 外接撤销
func handleExternalCancel(msg *pb.PublishExternalCancel) (err error) {
	ctx, cencel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cencel()
	key := fmt.Sprintf("%s:%s-%s", gameExternalRound, msg.Userid, msg.RoundId)

	var score int64
	switch msg.Type { // 1下注撤销,2返奖撤销
	case 1:
		score = msg.Amount
	case 2:
		score = -msg.Amount
	default:
		glog.Errorf("unknown external cancel type: %#v", msg)
		return
	}
	if _, err = myredis.Redis().IncrBy(ctx, key, score).Result(); err != nil {
		glog.Errorf("external round incr round cancel error: %#v, %v", msg, err)
		return
	}

	return nil
}

// 外接返奖(多次返奖)
func handleExternalReward(msg *pb.PublishExternalReward) (err error) {
	ctx, cencel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cencel()
	key := fmt.Sprintf("%s:%s-%s", gameExternalRound, msg.Userid, msg.RoundId)

	score, err := myredis.Redis().IncrBy(ctx, key, 0).Result()
	if err != nil {
		glog.Errorf("external round get round score error: %#v, %v", msg, err)
		return
	}

	curScore := score + msg.Amount
	// 玩家本局亏损
	if curScore <= 0 {
		if _, err := myredis.Redis().IncrBy(ctx, key, msg.Amount).Result(); err != nil {
			glog.Errorf("external round incr round score error: %#v, %v", msg, err)
		}
		return
	}
	// 本局盈亏清零
	if _, err := myredis.Redis().IncrBy(ctx, key, msg.Amount-curScore).Result(); err != nil {
		glog.Errorf("external round incr round score error: %#v, %v", msg, err)
	}

	// 波动返水返奖
	glog.Infof("volatility external reward: %s, %s, %d", msg.Userid, msg.RoundId, curScore)
	volatilitySubsidyReward(msg.Userid, msg.RegistArea, curScore)

	return nil
}

func (a *RoleActor) ActivityPop(userid string, ptype int32) {
	user := a.getUserById(userid)
	if user == nil {
		return
	}

	// 活动弹窗
	list := table.GetTables().ActivityPopTable.GetDataList()
	if len(list) == 0 {
		return
	}
	// 从小到大排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].Index < list[j].Index
	})

	ntf := new(pb.ActivityPopNtf)
	// 登录
	for _, v := range list {
		// 是否跳过
		if user.SkipGiftWelfare && v.Id == 1 { // 充值礼包
			continue
		}
		if user.SkipTurntable && v.Id == 3 { // 转盘
			continue
		}
		if user.SkipRank && v.Id == 4 { // 排行榜
			continue
		}
		if user.SkipGiftCode && v.Id == 5 { // 礼包码
			continue
		}
		if ptype == 1 {
			// 首次登录
			if user.LogoutTime.IsZero() {
				if v.FirstLogin == 1 {
					ntf.Ptype = append(ntf.Ptype, v.Id)
					// 冷却
					if v.LoginLobby > 0 {
						myredis.Redis().SetNX(context.Background(), fmt.Sprintf(data.ActivityLoginPopKey, userid, v.Id), 1, time.Duration(v.LoginLobby)*time.Minute)
					}
				}
			}
			// 非首次登录
			if v.LoginLobby >= 0 {
				success, _ := myredis.Redis().SetNX(context.Background(), fmt.Sprintf(data.ActivityLoginPopKey, userid, v.Id), 1, time.Duration(v.LoginLobby)*time.Minute).Result()
				if success {
					ntf.Ptype = append(ntf.Ptype, v.Id)
				}
			}

		} else if ptype == 2 {
			if v.GameLobby >= 0 {
				success, _ := myredis.Redis().SetNX(context.Background(), fmt.Sprintf(data.ActivityLobbyPopKey, userid, v.Id), 1, time.Duration(v.GameLobby)*time.Minute).Result()
				if success {
					ntf.Ptype = append(ntf.Ptype, v.Id)
				}
			}
		}
	}
	if len(ntf.Ptype) > 0 {
		ntf.Userid = userid
		rolePid.Tell(ntf)
	}
}

func (a *RoleActor) ActivityPopNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.ActivityPopNtf)
	if r, ok := a.roles[msg.Userid]; ok {
		r.Pid.Tell(msg)
	}
}
