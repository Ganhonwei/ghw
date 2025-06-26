package gate

import (
	"fmt"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"goserver/pkg/utils"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (rs *RoleActor) OfflineStop(ctx actor.Context) {
	arg := ctx.Message().(*pb.OfflineStop)
	glog.Debugf("rs OfflineStop %s", ctx.Self().String())
	//停止,TODO 暂时使用离线方法
	rs.forceLeave(int32(arg.Ltype))
	//断开连接
	rs.CloseWs()
	//关闭
	rs.StopRs()
}

func (rs *RoleActor) ServeClose(ctx actor.Context) {
	glog.Debugf("rs ServeClose %s", ctx.Self().String())
	arg := new(pb.LoginOutNtf)
	arg.Rtype = int32(pb.LOGOUT_TYPE2) //停服
	rs.Send(arg)
	//断开连接
	rs.CloseWs()
	//停止,TODO 暂时使用离线方法
	rs.loginElse()
	//关闭
	rs.StopRs()
}

func (rs *RoleActor) ServeStop(ctx actor.Context) {
	glog.Debugf("rs ServeStop %s", ctx.Self().String())
	rs.stop(ctx)
	//响应
	//rsp := new(pb.ServeStarted)
	//ctx.Respond(rsp)
}

func (rs *RoleActor) ServeStart(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ServeStart)
	glog.Debugf("rs ServeStart %s, %#v", ctx.Self().String(), arg)
	//启动时钟
	go rs.ticker(ctx)
	//响应
	//rsp := new(pb.ServeStarted)
	//ctx.Respond(rsp)
}

// 时钟
func (rs *RoleActor) ticker(ctx actor.Context) {
	tick := time.Tick(time.Second)
	msg := new(pb.Tick)
	for {
		select {
		case <-rs.stopCh:
			glog.Info("rs ticker closed")
			return
		default: //防止阻塞
		}
		select {
		case <-rs.stopCh:
			glog.Info("rs ticker closed")
			return
		case <-tick:
			ctx.Self().Tell(msg)
		}
	}
}

// 30秒同步一次
func (rs *RoleActor) Tick(ctx actor.Context) {
	if rs.User == nil {
		return
	}
	rs.zeroReset(rs.User.LastResetTime)
	// 大富翁重置
	rs.resetScratchTicket()

	// ad上报
	rs.adReport()

	// 打码排行榜更新
	// rs.betRankUpdateTick()

	// 限时礼包
	if rs.LimitedGiftId != 0 && rs.LimitedGiftOverTime < utils.LocalTime().UnixMilli() {
		rs.OverlimitedGift = append(rs.OverlimitedGift, rs.LimitedGiftId)
		rs.LimitedGiftId = 0
		rs.LimitedGiftOverTime = 0
	}

	// 转盘时间结束重置检测
	nowSec := time.Now().Unix()
	if (rs.User.ActivityTurn != nil && rs.User.ActivityTurn.TurnEtime > 0 &&
		nowSec > rs.User.ActivityTurn.TurnEtime) || // 当前回合结束
		(rs.User.ActivityTurn != nil && rs.User.ActivityTurn.NextFreeDrawTime > 0 &&
			nowSec > rs.User.ActivityTurn.NextFreeDrawTime && // 免费抽奖时间到
			rs.User.ActivityTurn.DrawTimesFree == 0 &&
			rs.User.ActivityTurn.Score < rs.User.ActivityTurn.ScoreTarget) {
		r, err := rs.rolePid.RequestFuture(&pb.ActivityTurnReq{Userid: rs.Userid}, 3*time.Second).Result()
		if err != nil {
			glog.Error("activity turn etime check error: ", err)
		} else if rsp, ok := r.(*pb.ActivityTurnRsp); ok {
			rs.Send(&pb.ActivityTurnNtf{Turn: rsp})
		}
	}

	// if rs.online && rs.gamePid != nil {
	// 	rs.User.OnlineTime++
	// 	rs.User.TotalOnlineTime++
	// }

	if rs.timer_day++; rs.timer_day >= 86400 {
		rs.timer_day = 0
	}

	rs.timer++
	if rs.timer%60 == 0 {
		// 检测玩家状态
		rs.checkUserState()
		// 解除客服聊天频繁
		rs.customerResetFeedBackTimes()
	}
	if rs.timer%300 == 0 {
		// 自动发放优惠券
		rs.autoSendCoupon()
	}
	if rs.timer != 300 {
		return
	}
	rs.timer = 0
	if !rs.online {
		return
	}
	//同步数据
	rs.syncUser()
}

// 断线
func (rs *RoleActor) stop(ctx actor.Context) {
	glog.Infof("rs stop: %v, online=%v", ctx.Self().String(), rs.online)
	myactor.Logger().Tell(&pb.LogButtonClick2{
		Userid: rs.Userid,
		Typ:    fmt.Sprintf("玩家id:%s 时间戳 %d 断线", rs.Userid, time.Now().UnixMilli()),
	})

	//已经断开,在别处登录
	if !rs.online {
		glog.Warningf("user %s already offline", rs.Userid)
		return
	}
	//设置为离线状态
	rs.User.OnlineStatus = false
	rs.Vip.ResetTime = utils.BsonNow().Unix()
	rs.User.LogoutTime = utils.BsonNow()
	rs.status = true
	//关闭连接
	rs.CloseWs()
	//离开消息
	rs.leaveDesk()
	//回存数据
	rs.syncUser()
	//登出日志
	msg2 := &pb.LogLogout{
		Userid: rs.User.Userid,
		Event:  int32(pb.LOGOUT_TYPE1), //正常断开
	}
	myactor.Logger().Tell(msg2)
	//断开处理
	msg := &pb.Logout{
		Sender: ctx.Self(),
		Userid: rs.User.Userid,
		Token:  rs.token,
	}
	nodePid.Tell(msg)
	//表示已经断开
	rs.online = false
}

func (rs *RoleActor) adReport() {
	if len(rs.adQueue) <= 0 {
		return
	}
	glog.Infof("start report adjuest user:%s queuelen:%d  adid:%s", rs.Userid, len(rs.adQueue), rs.adid)
	for _, v := range rs.adQueue {
		msg := &pb.EventAdParam{
			Userid:                  rs.Userid,
			Phone:                   rs.Phone,
			AdId:                    rs.adid,
			AdKey:                   rs.AD_Key,
			AdS2SCode:               rs.AD_S2S_Code,
			AdEventCodeRegister:     rs.AD_Event_Code1,
			AdEventCodeLogin:        rs.AD_Event_Code2,
			AdEventCodeDeposit:      rs.AD_Event_Code3,
			AdEventCodeWithdraw:     rs.AD_Event_Code4,
			AdEventCodeFirstDeposit: rs.AD_Event_Code5,
			UserAgent:               rs.AD_User_Agent,
			Ip:                      rs.RegistIP,
			Fbc:                     rs.FB_Fbc,
			Fbp:                     rs.FB_Fbp,
		}

		if rs.AD_User_Agent == "" {
			msg.UserAgent = rs.AD_Fake_User_Agent
		}

		// rs.reportPid.Tell(msg)
		mq.NatsPublish(mq.TopicReportAdParam, msg)
		switch ad := v.(type) {
		case *pb.ReportRegister:
			msg1 := &pb.EventRegister{
				Userid:   rs.Userid,
				BundleId: rs.AD_BundleId,
			}
			// rs.reportPid.Tell(msg1)
			mq.NatsPublish(mq.TopicReportRegister, msg1)
			glog.Infof("report ad regist, user:%s", rs.Userid)

			// fb上报
			// appid中不包含.的都要上报到fb
			// if !strings.Contains(rs.AD_AppId, ".") {
			if ad.FbInfo != nil {
				glog.Infof("report fb regist, user:%s", rs.Userid)
				msg2 := &pb.FBEventRegister{
					Userid: rs.Userid,
					AppId:  rs.AD_AppId,
					// Agent:  msg.UserAgent,
					// Ip:     rs.RegistIP,
				}
				// rs.reportPid.Tell(msg2)
				mq.NatsPublish(mq.TopicReportFBRegister, msg2)
			}
		case *pb.ReportLogin:
			msg1 := &pb.EventLogin{
				Userid:   rs.Userid,
				BundleId: rs.AD_BundleId,
			}
			// rs.reportPid.Tell(msg1)
			mq.NatsPublish(mq.TopicReportLogin, msg1)
			glog.Infof("report ad login, user:%s", rs.Userid)
			// if msg.AdInfo != nil {
			// 	msg.AdInfo.ADID = rs.adid
			// 	rs.reportPid.Tell(msg)
			// }
		case *pb.ReportDeposit:
			msg1 := &pb.EventDeposit{
				Userid:   rs.Userid,
				BundleId: rs.AD_BundleId,
				Amount:   ad.Amount,
				First:    rs.Money-uint32(ad.Amount*100) <= 0 && utils.Time2DayDate(rs.Ctime) == utils.Time2DayDate(time.Now().Local()),
			}
			// rs.reportPid.Tell(msg1)
			mq.NatsPublish(mq.TopicReportDeposit, msg1)
			glog.Infof("report ad deposit, user:%s", rs.Userid)

			// fb上报
			// appid中不包含.的都要上报到fb
			// if !strings.Contains(rs.AD_AppId, ".") {
			if ad.FbInfo != nil {
				msg2 := &pb.FBEventDeposit{
					Userid: rs.Userid,
					AppId:  rs.AD_AppId,
					Amount: ad.Amount,
					// Agent:  msg.UserAgent,
					// Ip:     rs.RegistIP,
					First: rs.Money-uint32(ad.Amount*100) <= 0 && utils.Time2DayDate(rs.Ctime) == utils.Time2DayDate(time.Now().Local()),
				}
				// rs.reportPid.Tell(msg2)
				mq.NatsPublish(mq.TopicReportFBDeposit, msg2)
			}
			// case *pb.ReportFirstDeposit:
			// 	// 首充
			// 	msg1 := &pb.EventFirstDeposit{
			// 		Userid:   rs.Userid,
			// 		BundleId: rs.AD_BundleId,
			// 		Amount:   ad.Amount,
			// 	}
			// 	rs.reportPid.Tell(msg1)
			// 	glog.Infof("report ad first deposit, user:%s", rs.Userid)

			// 	// fb上报
			// 	// appid中不包含.的都要上报到fb
			// 	if !strings.Contains(rs.AD_AppId, ".") {
			// 		msg2 := &pb.FBEventFirstDeposit{
			// 			Userid: rs.Userid,
			// 			AppId:  rs.AD_AppId,
			// 			Amount: ad.Amount,
			// 			Agent:  msg.UserAgent,
			// 			Ip:     rs.RegistIP,
			// 		}
			// 		rs.reportPid.Tell(msg2)
			// 	}
		}
	}
	rs.adQueue = make([]any, 0)
}

// 停服踢出
func (rs *RoleActor) CloseServer(ctx actor.Context) {
	if rs.gamePid != nil || rs.cpPid != nil {
		return
	}
	// rs.pid.Tell(&pb.OfflineStop{Ltype: pb.LOGOUT_TYPE2})
	// rs.pid.Tell(new(pb.ServeClose))
	arg := new(pb.LoginOutNtf)
	glog.Debugf("CloseServer %s", rs.User.Userid)
	arg.Rtype = int32(pb.LOGOUT_TYPE2)
	rs.Send(arg)

	//  rs.stop(ctx)
}

// 语言包
func (rs *RoleActor) SendLanguage(id string, ctype int, param ...string) {
	ntf := handler.BuildLanguageMsg(id, ctype, param...)
	rs.Send(ntf)
}
