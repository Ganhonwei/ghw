package gate

import (
	"fmt"
	"sort"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/game/login"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"goserver/pkg/table"
	"goserver/pkg/utils"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (rs *RoleActor) RegistReq(ctx actor.Context) {
	msg := ctx.Message()
	//注册消息
	arg := msg.(*pb.RegistReq)
	glog.Debugf("RegistReq %#v", arg)
	//重复登录
	stoc := new(pb.RegistRsp)
	stoc.Error = pb.RepeatLogin
	rs.Send(stoc)
}

func (rs *RoleActor) LoginReq(ctx actor.Context) {
	msg := ctx.Message()
	//登录消息
	arg := msg.(*pb.LoginReq)
	glog.Debugf("LoginReq %#v", arg)
	//重复登录
	stoc := new(pb.LoginRsp)
	stoc.Error = pb.RepeatLogin
	rs.Send(stoc)
}

// func (rs *RoleActor) WxLoginReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	//登录消息
// 	arg := msg.(*pb.WxLoginReq)
// 	glog.Debugf("WxLoginReq %#v", arg)
// 	//重复登录
// 	stoc := new(pb.WxLoginRsp)
// 	stoc.Error = pb.RepeatLogin
// 	rs.Send(stoc)
// }

func (rs *RoleActor) ResetPwdReq(ctx actor.Context) {
	msg := ctx.Message()
	//重置密码消息
	arg := msg.(*pb.ResetPwdReq)
	glog.Debugf("ResetPwdReq %#v", arg)
	//重复登录
	stoc := new(pb.ResetPwdRsp)
	stoc.Error = pb.RepeatLogin
	rs.Send(stoc)
}

func (rs *RoleActor) TouristReq(ctx actor.Context) {
	msg := ctx.Message()
	//登录消息
	arg := msg.(*pb.TouristReq)
	glog.Debugf("TouristReq %#v", arg)
	//重复登录
	stoc := new(pb.TouristRsp)
	stoc.Error = pb.RepeatLogin
	rs.Send(stoc)
}

func (rs *RoleActor) LoginElse(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.LoginElse)
	glog.Debugf("LoginElse %#v", arg)
	rs.loginElse() //别处登录
	//响应登录
	rsp := new(pb.LoginedElse)
	rsp.Userid = arg.Userid
	rsp.Gate = arg.Gate
	ctx.Respond(rsp)
}

func (rs *RoleActor) LoginSuccess(ctx actor.Context) {
	msg := ctx.Message()
	//登录成功处理
	arg := msg.(*pb.LoginSuccess)
	glog.Debugf("LoginSuccess %#v", arg)
	rs.logined(arg, ctx)
}

func (rs *RoleActor) AddBlackList(ctx actor.Context) {
	msg := ctx.Message()
	// 黑名单
	glog.Debugf("AddBlackList %#v", msg)
	if rs.User.Status == 3 {
		rs.status = true
		rs.User.Status = 1
		rs.User.StatusTime = time.Now().Local()
		return
	}
	rs.User.Status = 3
	rs.User.StatusTime = time.Now().Local()
	if rs.gamePid != nil {
		// 在游戏中,踢出去
		rs.gamePid.Tell(msg)
	}
	rs.forceLeave(int32(pb.LOGOUT_TYPE3))
}

func (rs *RoleActor) AddWhiteList(ctx actor.Context) {
	msg := ctx.Message()
	// 白名单
	glog.Debugf("AddWhiteList %#v", msg)
	if rs.User.Status == 4 {
		rs.User.Status = 1
	} else {
		rs.User.Status = 4
	}
	rs.User.StatusTime = time.Now().Local()
	rs.status = true
}

// 别处登录
func (rs *RoleActor) loginElse() {
	rs.forceLeave(int32(pb.LOGOUT_TYPE4))
	rs.wsPid.Tell(&pb.LoginElse{WsPid: rs.wsPid})
}

// 强制离线
func (rs *RoleActor) forceLeave(ltype int32) {
	if rs.User == nil {
		return
	}
	//同步数据
	rs.syncUser()
	//已经断开
	if !rs.online {
		return
	}
	arg := new(pb.LoginOutNtf)
	glog.Debugf("LoginOutNtf %s, ltype:%d", rs.User.Userid, ltype)
	arg.Rtype = ltype
	rs.Send(arg)
	//断开连接,改为客户端断开连接
	// rs.CloseWs()
	//离开游戏消息
	rs.leaveDesk()
	//登出日志
	msg3 := &pb.LogLogout{
		Userid: rs.User.Userid,
		Event:  ltype,
	}
	myactor.Logger().Tell(msg3)
	//表示已经断开
	// rs.online = false

}

// 离开游戏处理
func (rs *RoleActor) leaveDesk() {
	//离线
	//msg2 := new(pb.OfflineDesk)
	//if rs.User != nil {
	//	msg2.Userid = rs.User.GetUserid()
	//}
	//rs.gamePid.Tell(msg2)
	//下线
	msg3 := new(pb.LeaveDesk)
	if rs.User != nil {
		msg3.Userid = rs.User.GetUserid()
	}
	//rs.gamePid.Tell(msg3)
	timeout := 3 * time.Second
	if rs.cpPid != nil {
		rs.cpPid.RequestFuture(msg3, time.Second*3)
		rs.cpPid = nil
	}

	if rs.gamePid == nil {
		return
	}
	res1, err1 := rs.gamePid.RequestFuture(msg3, timeout).Result()
	if err1 != nil {
		glog.Errorf("leave desk failed: %v", err1)
	}
	if response1, ok := res1.(*pb.LeftDesk); ok {
		glog.Debugf("left desk response1: %#v", response1)
		if response1.Error != pb.OK {
			glog.Errorf("leave desk failed: %#v", res1)
		} else {
			//不同游戏可能规则不同，由游戏控制是否离开
			rs.gamePid = nil
			rs.gameId = ""
			rs.gtype = 0
			rs.roomId = ""
			rs.roomCode = ""
		}
	} else {
		glog.Errorf("leave desk failed: %#v", res1)
	}
}

// 登录成功处理
func (rs *RoleActor) logined(arg *pb.LoginSuccess, ctx actor.Context) {
	if !rs.loginedGetUser(arg.Userid, ctx) {
		glog.Errorf("logined fail %s, %s", ctx.Self().String(), arg.Userid)
		rs.loginFailed(ctx, arg.Userid)
		return
	}
	reconnect := rs.token == arg.Token
	glog.Infof("user %s,reconnect: %t, login token: %s", rs.Userid, reconnect, arg.Token)
	//断开旧连接
	rs.CloseWs()
	//新连接
	rs.wsPid = arg.WsPid
	rs.token = arg.Token
	//头像
	rs.setHeadImag(arg.IsRegist, ctx)
	//vb活动任务初始化
	rs.setVbTask()
	//计算已提现vb
	rs.setCashOutBounds()
	//日志
	rs.loginedLog(arg)
	//登录成功
	// rs.online = true
	rs.adid = rs.AD_ADID
	// rs.RegistArea = 1 // 全部改为B类
	rs.User.OnlineStatus = true
	rs.User.FirstLoginToday = false
	//设置登录时间
	rs.User.LoginTime = utils.BsonNow()
	//检查是否需要跨天
	rs.zeroReset(rs.User.LastResetTime)
	if rs.OutDiamond < 0 {
		rs.OutDiamond = 0
	}
	// rs.OutDiamond = int64(math.Min(float64(rs.OutDiamond), float64(rs.Diamond)))
	if rs.User.RegistMode == 0 {
		rs.User.RegistReward = true
	}
	// 初始化历史数据
	rs.initHistory()
	rs.status = true

	// ad上报
	if arg.IsRegist {
		// 注册
		msg := login.ReportRegisterMsg(rs.User)
		if msg != nil {
			rs.adQueue = append(rs.adQueue, msg)
		}
	} else {
		// 登录
		if !reconnect {
			// 不是断线重连才上报
			msg := login.ReportLoginMsg(rs.User)
			if msg != nil {
				rs.adQueue = append(rs.adQueue, msg)
			}
		}
	}
}

func (rs *RoleActor) sendSystem() {
	if rs.User.State <= 0 {
		rs.User.State = 2
	}
	// 功能开关
	rs.systemSwitch()
	// 游戏列表
	rs.gameList()
	// 客服消息
	rs.feedBackChats()
	// 活动数据
	rs.sendActivityData()
	// 离线消息
	rs.sendOfflineMsg()
	// 客服地址
	rs.customerAddr()
	// 分享信息
	rs.sharedataNtf()
	// 限时礼包
	// rs.limitedGiftNtf()
	// 对战房配置信息
	rs.pvpRoom()
	// VIP等级重置
	rs.vipLvReset()
	// vip初始化消息
	rs.vipInitDataNtf()
	// VIP升级奖励补发
	rs.vipUpgradeRewardReissue()
	// vbbank日志 -> VipBankLogRecordReq
	// rs.VipBankLogNtf()
	// 大厅游戏人数
	rs.roomPid.Request(new(pb.SyncGamePeople), rs.pid)
	// 分享配置
	rs.shareWayNtf()
	// 自定义头像
	rs.photoNtf()
	// 破冰礼包
	rs.pbGift()
	// unlockcash
	rs.checkUnlockCash()
	// 波动返水补贴检查
	rs.rolePid.Tell(&pb.VolatilitySubsidyToLobby{Userid: rs.Userid})
	// 优惠券
	rs.autoSendCoupon()
	// 客服配置推送
	rs.customerSetting()
}

func (rs *RoleActor) vipLvReset() {
	// vip等级重置
	if rs.Vip.Lv > 0 && rs.Vip.ResetTime > 0 {
		pop := false
		// vipbean := table.GetTables().VipTable.Get(int32(rs.Vip.Lv))
		duringDay := (utils.BsonNow().Unix() - rs.Vip.ResetTime) / 3600 / 24
		for {
			if rs.Vip.Lv <= 0 {
				break
			}
			vipbean := table.GetTables().VipTable.Get(int32(rs.Vip.Lv))
			if duringDay > int64(vipbean.Reset) && vipbean.Reset > 0 {
				pop = true
				rs.Vip.Lv = int(vipbean.ResetLv)
				rs.Vip.Exp = int64(vipbean.Recharge)
				if rs.Vip.Lv > rs.Vip.MaxLv {
					rs.Vip.MaxLv = rs.Vip.Lv
				}
				duringDay -= int64(vipbean.Reset)
				continue
			}
			break
		}
		if pop {
			ntf := &pb.VipLevelResetNtf{Lv: int32(rs.Vip.Lv), Exp: rs.Vip.Exp, Day: duringDay}
			rs.Send(ntf)
		}
		rs.status = true
	}
}

// vip 升级奖励补发
func (rs *RoleActor) vipUpgradeRewardReissue() {
	rewardLvs := rs.vipUpgradeReward()
	for _, vip := range rewardLvs {
		// 前端弹窗
		msg := &pb.UpdateVipExpNtf{
			Lv:       vip,
			Exp:      rs.User.Vip.Exp,
			BeforeLv: vip - 1,
		}
		rs.Send(msg)
	}
}

// vip 升级奖励补发
func (rs *RoleActor) vipUpgradeReward() (rewardLvs []int32) {
	user := rs.User
	for i := 1; i <= rs.Vip.Lv; i++ {
		vip := int32(i)
		vipRecord := table.GetTables().VipTable.Get(vip)
		if vipRecord.Upgrade <= 0 ||
			utils.SliceIn(vip, user.Vip.LevelReward...) { // 领过了
			continue
		}

		rs.sendGood(0, 0, int64(vipRecord.Upgrade), 0, 0, 0, int32(pb.LOG_TYPE99), fmt.Sprintf("vip%d等级奖励", vipRecord.Id))
		user.Vip.LevelReward = append(user.Vip.LevelReward, vip)
		rs.status = true

		rewardLvs = append(rewardLvs, vip)
	}
	return
}

func (rs *RoleActor) gameList() {
	type gameWeight struct {
		gtype  string
		weight int
	}

	ntf := new(pb.GameDistributionNtf)
	games := make(map[int][]gameWeight)
	switchMap := config.GetSettingMap()
	for _, v := range switchMap {
		if v.Stype != 0 {
			continue
		}
		for _, t := range v.Tab {
			if g, ok := games[t]; ok {
				g = append(g, gameWeight{gtype: utils.String(v.Gtype), weight: v.SortId})
				games[t] = g
			} else {
				games[t] = []gameWeight{{gtype: utils.String(v.Gtype), weight: v.SortId}}
			}
		}
	}

	for k, weights := range games {
		sort.Slice(weights, func(i, j int) bool {
			return weights[i].weight < weights[j].weight
		})
		distribution := &pb.GameDistribution{Dtype: int32(k)}
		for _, w := range weights {
			distribution.Gtype = append(distribution.Gtype, w.gtype)
		}
		ntf.Data = append(ntf.Data, distribution)
	}

	ntf.Data = append(ntf.Data, &pb.GameDistribution{Dtype: 4, Gtype: rs.User.Favorite})

	rs.Send(ntf)
}

// 登录后获取数据
func (rs *RoleActor) loginedGetUser(userid string, ctx actor.Context) bool {
	//关闭旧进程
	rs.loginElseGate(userid)
	//已经存在，TODO 优化,SelectGate时返回结果
	if rs.User != nil {
		return true
	}
	//登录
	msg1 := &pb.GetUser{
		Userid:  userid,
		Gate:    cfg.Section(nodeName).Name(),
		RolePid: ctx.Self(),
	}
	timeout := 3 * time.Second
	res1, err1 := rs.rolePid.RequestFuture(msg1, timeout).Result()
	if err1 != nil {
		glog.Errorf("logined GetUser failed: %v", err1)
		return false
	}
	if response1, ok := res1.(*pb.GotUser); ok {
		user := new(data.User)
		err2 := json.Unmarshal(response1.Data, user)
		if err2 != nil {
			glog.Errorf("user Unmarshal err %v", err2)
			return false
		}
		if user.GetUserid() == "" {
			return false
		}
		rs.User = user
		glog.Debugf("loginedGetUser %#v", rs.User)
		return true
	}
	return false
}

// 别处登录处理
func (rs *RoleActor) loginElseGate(userid string) {
	msg1 := new(pb.LoginElse)
	msg1.Userid = userid
	msg1.Gate = cfg.Section(nodeName).Name()
	timeout := 3 * time.Second
	res1, err1 := rs.rolePid.RequestFuture(msg1, timeout).Result()
	if err1 != nil {
		glog.Errorf("loginElseGate res1 %#v, err1 %v", res1, err1)
	}
}

// 登录失败 TODO 优化
func (rs *RoleActor) loginFailed(ctx actor.Context, userid string) {
	ctx.Self().Tell(&pb.OfflineStop{Ltype: pb.LOGOUT_TYPE1})
	//正式下线消息
	arg := new(pb.Logout)
	arg.Userid = userid
	arg.Sender = rs.pid
	rs.rolePid.Tell(arg)
	rs.roomPid.Tell(arg)
}

// 默认头像
func (rs *RoleActor) setHeadImag(isRegist bool, ctx actor.Context) {
	if !isRegist {
		//return
	}
	if rs.User == nil {
		return
	}
	if rs.User.GetPhoto() != "" {
		return
	}
	if !rs.User.GetRobot() {
		return
	}
	if len(HeadImagList) == 0 {
		return
	}
	head := cfg.Section("domain").Key("headimag").Value()
	if head == "" {
		return
	}
	i := utils.RandIntN(len(HeadImagList))
	rs.User.Photo = head + "/" + HeadImagList[i].Photo
}

// 登录成功日志处理
func (rs *RoleActor) loginedLog(arg *pb.LoginSuccess) {
	rs.User.LoginIP = arg.Ip
	//连续登录
	rs.loginPrizeInit()
	if arg.IsRegist {
		//注册ip
		// rs.User.RegistIP = arg.Ip
		if !rs.User.Robot && !rs.User.SimRobot {
			//注册奖励发放
			// var diamond int64 = int64(config.GetEnv(data.ENV1))
			// var coin int64 = int64(config.GetEnv(data.ENV2))
			// var chip int64 = int64(config.GetEnv(data.ENV3))
			// var card int64 = int64(config.GetEnv(data.ENV4))
			// bean := table.GetTables().NewbieTable.Get()
			// rs.addCurrency(int64(bean.GiveCash), int64(bean.GiveBouns), 0, 0, 0, 0, int32(pb.LOG_TYPE1), "注册奖励", "")
			//注册日志
			msg1 := &pb.LogRegist{
				Userid:   rs.User.Userid,
				Nickname: rs.User.Nickname,
				Ip:       arg.Ip,
			}
			myactor.Logger().Tell(msg1)
		}
	}
	//登录日志
	if !rs.Robot {
		msg2 := &pb.LogLogin{
			Userid: rs.User.Userid,
			Ip:     arg.Ip,
		}
		myactor.Logger().Tell(msg2)
	}
	//TODO test
	//rs.loginedLog2()
	//登录检测区域奖励发放
	// msg3, msg5 := handler.AgentProfitMonthSendCheck(rs.User)
	// if msg3 != nil {
	// 	myactor.Logger().Tell(msg3)
	// }
	// if msg5 != nil {
	// 	rs.rolePid.Tell(msg5)
	// }
}

// test
func (rs *RoleActor) loginedLog2() {
	var diamond, coin int64
	if rs.User.GetDiamond() < 10000 {
		diamond = 10000
	}
	if rs.User.GetCoin() < 5000000 {
		coin = 5000000
	}
	glog.Debugf("loginedLog2 userid %s, diamond %d, coin %d",
		rs.User.GetUserid(), diamond, coin)
	if diamond == 0 && coin == 0 {
		return
	}
	rs.addCurrency(diamond, coin, 0, 0, diamond, 0, 0, int32(pb.LOG_TYPE11), "loginedLog2", "")
}

// Send 发送消息
func (rs *RoleActor) Send(msg interface{}) {
	//glog.Debugf("Send %#v", msg)
	if rs.stopCh == nil {
		glog.Errorf("rs msg channel closed %v", msg)
		return
	}
	if rs.wsPid == nil {
		glog.Errorf("ws pid stoped %v", msg)
		return
	}
	//glog.Debugf("send message %s", rs.wsPid.String())
	select {
	case <-rs.stopCh:
		return
	default:
	}
	select {
	case <-rs.stopCh:
		return
	default:
		//glog.Debugf("send message %#v", msg)
		rs.wsPid.Tell(msg)
	}
}

// StopRs 关闭
func (rs *RoleActor) StopRs() {
	select {
	case <-rs.stopCh:
		return
	default:
		//停止发送消息
		close(rs.stopCh)
	}
	//停止
	rs.pid.Stop()
}

// CloseWs 关闭连接
func (rs *RoleActor) CloseWs() {
	if rs.wsPid == nil {
		return
	}
	glog.Debugf("CloseWs userid: %s", rs.wsPid.String())
	msg1 := new(pb.ServeStop)
	//关闭连接
	rs.wsPid.Tell(msg1)
	//断开
	rs.wsPid = nil
}
