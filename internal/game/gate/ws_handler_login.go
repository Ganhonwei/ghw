package gate

import (
	"net"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/game/login"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 游客登录
func (ws *WSConn) TouristReq(ctx actor.Context) {
	//登录消息
	arg := ctx.Message().(*pb.TouristReq)
	glog.Debugf("TouristReq %#v", arg)
	//检测参数
	// key := cfg.Section("gate").Key("tourist").Value()
	// stoc := login.TouristLoginCheck(arg, key)
	// if stoc.Error != pb.OK {
	// 	ws.Send(stoc)
	// 	return
	// }
	stoc := new(pb.TouristRsp)

	if arg.Account == "" || arg.Ad == nil {
		glog.Infof("[TouristReq] account is nil")
		stoc.Error = pb.LoginError
		ws.Send(stoc)
		return
	}

	// 检测服务器状态
	stoc.Error = ws.checkServerStatus("")
	if stoc.Error != pb.OK {
		// 服务器维护
		stoc.NotifyContent = config.GetNoticesByType(2)
		ws.Send(stoc)
		return
	}

	deviceMap := config.GetDeviceBlackMap()
	if _, ok := deviceMap[arg.Ad.DeviceId]; ok {
		glog.Infof("[TouristReq] device not white list device:%s", arg.Ad.DeviceId)
		stoc.Error = pb.RegistError
		ws.Send(stoc)
		return
	}

	if handler.ChannelInterceptor(arg.Ad) && env != "dev" {
		glog.Infof("[TouristReq] channel package err")
		stoc.Error = pb.RegistError
		ws.Send(stoc)
		return
	}

	if ipInterceptor(ws.GetIPAddr(), 0) {
		glog.Infof("[TouristReq] ip not white list ip:%s", ws.GetIPAddr())
		stoc.Error = pb.LoginError
		ws.Send(stoc)
		return
	}

	//重复登录
	if ws.online {
		stoc.Error = pb.RepeatLogin
		ws.Send(stoc)
		return
	}
	msg1 := new(pb.TouristLogin)
	msg1.Account = arg.GetAccount()
	msg1.Password = arg.GetPassword()
	msg1.Registip = ws.GetIPAddr()
	msg1.Af = arg.GetAf()
	msg1.Ad = arg.GetAd()
	msg1.Fb = arg.Fb
	//登录
	res1 := ws.reqRole(msg1, ctx)
	var response1 *pb.TouristLogined
	var ok bool
	if response1, ok = res1.(*pb.TouristLogined); ok {
		if response1.Error != pb.OK {
			glog.Errorf("TouristReq fail %d", response1.Error)
			stoc.Error = response1.Error
			ws.Send(stoc)
			return
		}
	} else {
		glog.Error("TouristReq fail")
		stoc.Error = pb.LoginError
		ws.Send(stoc)
		return
	}

	// 地区拦截(oppo)
	if ws.AreaInterceptor(response1.Appid) {
		glog.Infof("ip is not allow to login in this city ip:%s,appid:%s,", ws.GetIPAddr(), response1.Appid)
		stoc.Error = pb.CityRestrictions
		ws.Send(stoc)
		return
	}

	userid := response1.GetUserid()
	glog.Debugf("tourist login successfully %s", userid)
	succ, t := ws.logining(userid, arg.Token, ctx)
	if !succ {
		glog.Debugf("login failed %s", ctx.Self())
		stoc.Error = pb.LoginError
		ws.Send(stoc)
		return
	}
	stoc.Userid = userid
	stoc.Mode = response1.Mode
	// stoc.Reward = handler.BuildRegistReward(user)
	glog.Debugf("tourist login successfully %s", userid)
	ws.Send(stoc)
	//成功后处理
	ws.logined(userid, t, response1.IsRegist, ctx)
}

func (ws *WSConn) ResetPwdReq(ctx actor.Context) {
	//重置密码消息
	arg := ctx.Message().(*pb.ResetPwdReq)
	glog.Debugf("ResetPwdReq %#v", arg)
	stoc := login.RestPwdCheck(arg)
	if stoc.Error != pb.OK {
		ws.Send(stoc)
		return
	}
	//已经登录
	if ws.online {
		stoc.Error = pb.RepeatLogin
		ws.Send(stoc)
		return
	}
	//重置
	res1 := ws.reqRole(arg, ctx)
	var response1 *pb.ResetPwdRsp
	var ok bool
	if response1, ok = res1.(*pb.ResetPwdRsp); ok {
		if response1.Error != pb.OK {
			glog.Errorf("RegistReq fail %d", response1.Error)
			stoc.Error = response1.Error
			ws.Send(stoc)
			return
		}
	} else {
		glog.Error("ResetPwdReq fail")
		stoc.Error = pb.ResetPwdFaild
		ws.Send(stoc)
		return
	}
	userid := response1.GetUserid()
	glog.Debugf("ResetPwdReq successfully %s", userid)
	succ, t := ws.logining(userid, arg.Token, ctx)
	if !succ {
		glog.Debugf("ResetPwdReq failed %s", ctx.Self())
		stoc.Error = pb.ResetPwdFaild
		ws.Send(stoc)
		return
	}
	stoc.Userid = userid
	glog.Debugf("ResetPwdReq successfully %s", userid)
	ws.Send(stoc)
	//成功后处理
	ws.logined(userid, t, true, ctx)
}

// TODO 注册ip限制
func (ws *WSConn) RegistReq(ctx actor.Context) {
	//注册消息
	arg := ctx.Message().(*pb.RegistReq)
	glog.Debugf("RegistReq %#v", arg)
	stoc := login.RegistCheck(arg)
	if stoc.Error != pb.OK {
		ws.Send(stoc)
		return
	}

	stoc.Error = ws.checkServerStatus(arg.Phone)
	if stoc.Error != pb.OK {
		// 服务器维护
		stoc.NotifyContent = config.GetNoticesByType(2)
		ws.Send(stoc)
		return
	}

	if arg.Ad == nil {
		glog.Infof("[RegistReq] Ad is nil")
		stoc.Error = pb.RegistError
		ws.Send(stoc)
		return
	}

	deviceMap := config.GetDeviceBlackMap()
	if _, ok := deviceMap[arg.Ad.DeviceId]; ok {
		glog.Infof("[RegistReq] device not white list device:%s", arg.Ad.DeviceId)
		stoc.Error = pb.RegistError
		ws.Send(stoc)
		return
	}

	if handler.ChannelInterceptor(arg.Ad) && env != "dev" {
		glog.Infof("[RegistReq] channel package err")
		stoc.Error = pb.RegistError
		ws.Send(stoc)
		return
	}
	if ipInterceptor(ws.GetIPAddr(), 0) {
		glog.Infof("[RegistReq] ip not white list ip:%s", ws.GetIPAddr())
		stoc.Error = pb.RegistError
		ws.Send(stoc)
		return
	}

	//重复登录
	if ws.online {
		stoc.Error = pb.RepeatLogin
		ws.Send(stoc)
		return
	}
	msg1 := new(pb.RoleRegist)
	msg1.Phone = arg.GetPhone()
	msg1.Nickname = arg.GetNickname()
	msg1.Password = arg.GetPassword()
	msg1.Smscode = arg.GetSmscode()
	msg1.Safetycode = arg.GetSafetycode()
	msg1.Af = arg.GetAf()
	msg1.Ad = arg.GetAd()
	msg1.Ip = ws.GetIPAddr()
	msg1.Fb = arg.Fb

	//注册
	res1 := ws.reqRole(msg1, ctx)
	var response1 *pb.RoleRegisted
	var ok, regist bool = false, true
	if response1, ok = res1.(*pb.RoleRegisted); ok {
		if response1.Error != pb.OK && response1.Error != pb.PhoneRegisted {
			glog.Errorf("RegistReq fail %d", response1.Error)
			stoc.Error = response1.Error
			ws.Send(stoc)
			return
		}
		if response1.Error == pb.PhoneRegisted {
			regist = false
		} else {
			stoc.Mode = response1.Mode
		}
	} else {
		glog.Error("RegistReq fail")
		stoc.Error = pb.RegistError
		ws.Send(stoc)
		return
	}
	userid := response1.GetUserid()
	glog.Debugf("regist successfully %s", userid)
	succ, t := ws.logining(userid, arg.Token, ctx)
	if !succ {
		glog.Debugf("regist failed %s", ctx.Self())
		stoc.Error = pb.RegistError
		ws.Send(stoc)
		return
	}
	stoc.Userid = userid
	// stoc.Reward = handler.BuildRegistReward()
	glog.Debugf("regist successfully %s", userid)
	ws.Send(stoc)
	//成功后处理
	ws.logined(userid, t, regist, ctx)
}

func (ws *WSConn) LoginReq(ctx actor.Context) {
	//登录消息
	arg := ctx.Message().(*pb.LoginReq)
	glog.Debugf("LoginReq %#v", arg)
	//检测参数
	stoc := login.LoginCheck(arg)
	if stoc.Error != pb.OK {
		ws.Send(stoc)
		return
	}

	// 检查服务器状态
	stoc.Error = ws.checkServerStatus(arg.Phone)
	if stoc.Error != pb.OK {
		// 服务器维护
		stoc.NotifyContent = config.GetNoticesByType(2)
		ws.Send(stoc)
		return
	}

	//重复登录
	if ws.online {
		stoc.Error = pb.RepeatLogin
		ws.Send(stoc)
		return
	}
	msg1 := new(pb.RoleLogin)
	msg1.Phone = arg.GetPhone()
	msg1.Password = arg.GetPassword()
	//登录
	res1 := ws.reqRole(msg1, ctx)
	var response1 *pb.RoleLogined
	var ok bool
	if response1, ok = res1.(*pb.RoleLogined); ok {
		if response1.Error != pb.OK {
			glog.Errorf("LoginReq fail %d", response1.Error)
			stoc.Error = response1.Error
			ws.Send(stoc)
			return
		}
	} else {
		glog.Error("LoginReq fail")
		stoc.Error = pb.LoginError
		ws.Send(stoc)
		return
	}
	userid := response1.GetUserid()
	// ip拦截
	// 非白名单用户
	if ipInterceptor(ws.GetIPAddr(), response1.Status) {
		glog.Infof("ip or userid not white list ip:%s, userid:%s ,status:%d,", ws.GetIPAddr(), userid, response1.Status)
		stoc.Error = pb.LoginError
		ws.Send(stoc)
		return
	}

	// 地区拦截(oppo)
	if ws.AreaInterceptor(response1.Appid) {
		glog.Infof("ip is not allow to login in this city ip:%s, userid:%s ,appid:%s,", ws.GetIPAddr(), userid, response1.Appid)
		stoc.Error = pb.CityRestrictions
		ws.Send(stoc)
		return
	}

	glog.Debugf("login successfully %s", userid)
	succ, t := ws.logining(userid, arg.Token, ctx)
	if !succ {
		glog.Debugf("login failed %s", ctx.Self())
		stoc.Error = pb.LoginError
		ws.Send(stoc)
		return
	}
	stoc.Userid = userid
	glog.Debugf("login successfully %s", userid)
	ws.Send(stoc)
	//成功后处理
	ws.logined(userid, t, false, ctx)
}

// 微信
// func (ws *WSConn) WxLoginReq(ctx actor.Context) {
// 	//登录消息
// 	arg := ctx.Message().(*pb.WxLoginReq)
// 	glog.Debugf("WxLoginReq %#v", arg)
// 	stoc, wxdata := login.WxLoginCheck(arg)
// 	if stoc.Error != pb.OK {
// 		ws.Send(stoc)
// 		return
// 	}
// 	//重复登录
// 	if ws.online {
// 		stoc.Error = pb.RepeatLogin
// 		ws.Send(stoc)
// 		return
// 	}
// 	msg1 := new(pb.WxLogin)
// 	msg1.Wxuid = wxdata.UnionId // Unionid or OpenId
// 	msg1.Nickname = wxdata.Nickname
// 	msg1.Photo = wxdata.HeadImagUrl
// 	msg1.Sex = uint32(wxdata.Sex)
// 	msg1.OpenId = wxdata.OpenId
// 	msg1.UnionId = wxdata.UnionId
// 	//登录
// 	res1 := ws.reqRole(msg1, ctx)
// 	var response1 *pb.WxLogined
// 	var ok bool
// 	if response1, ok = res1.(*pb.WxLogined); ok {
// 		if response1.Error != pb.OK {
// 			glog.Errorf("WxLoginReq fail %d", response1.Error)
// 			stoc.Error = response1.Error
// 			ws.Send(stoc)
// 			return
// 		}
// 	} else {
// 		glog.Error("WxLoginReq fail")
// 		stoc.Error = pb.GetWechatUserInfoFail
// 		ws.Send(stoc)
// 		return
// 	}
// 	userid := response1.GetUserid()
// 	glog.Debugf("weixin login successfully %s", userid)
// 	if !ws.logining(userid, ctx) {
// 		stoc.Error = pb.GetWechatUserInfoFail
// 		ws.Send(stoc)
// 		return
// 	}
// 	stoc.Userid = userid
// 	glog.Debugf("weixin login successfully %s", userid)
// 	ws.Send(stoc)
// 	//成功后处理
// 	ws.logined(userid, response1.IsRegist, ctx)
// }

// 登录成功处理
func (ws *WSConn) logined(userid, token string, isRegist bool,
	ctx actor.Context) {
	//登录成功消息
	msg := new(pb.LoginSuccess)
	msg.IsRegist = isRegist
	msg.Ip = ws.GetIPAddr()
	msg.Userid = userid
	msg.WsPid = ctx.Self()
	msg.Token = token
	//pid已经切换为rsPid
	ws.pid.Tell(msg)
	//登录成功
	ws.online = true
	//成功
	ctx.SetReceiveTimeout(0) //login Successfully, timeout off
}

// 登录流程处理
func (ws *WSConn) logining(userid, token string, ctx actor.Context) (bool, string) {
	if userid == "" {
		glog.Debugf("logining failed %s", userid)
		return false, ""
	}
	//当前节点中查询
	succ, t := ws.selectGate(userid, token, ctx)
	if !succ {
		////节点中不存在新建一个
		//if !ws.loginUser(userid, ctx) {
		//	glog.Debugf("logining loginUser failed %s", userid)
		//	return false
		//}
		return false, ""
	}
	return true, t
}

/*
//登录成功数据处理
func (ws *WSConn) loginUser(userid string, ctx actor.Context) bool {
	msg4 := new(pb.Login)
	msg4.Userid = userid
	//节点名称
	msg4.Gate = cfg.Section(nodeName).Name()
	msg4.RolePid = ws.pid
	//请求
	res4 := ws.reqRole(msg4, ctx)
	var response4 *pb.Logined
	var ok bool
	if response4, ok = res4.(*pb.Logined); !ok {
		glog.Debugf("loginUser failed %s", userid)
		return false
	}
	//glog.Debugf("response4: %#v", response4)
	msg1 := new(pb.Login2Gate)
	//msg1.WsPid = ws.pid
	msg1.Userid = userid
	msg1.Data = response4.Data //数据
	//在节点中spawn一个玩家进程
	if !ws.loginGate(msg1, ctx) {
		glog.Debugf("loginGate failed %s", userid)
		return false
	}
	return true
}

//登录节点
func (ws *WSConn) loginGate(msg2 *pb.Login2Gate, ctx actor.Context) bool {
	timeout := 3 * time.Second
	res2, err2 := nodePid.RequestFuture(msg2, timeout).Result()
	if err2 != nil {
		glog.Errorf("LoginGate err: %v", err2)
		return false
	}
	glog.Debugf("res2: %#v", res2)
	var response2 *pb.Logined2Gate
	var ok bool
	if response2, ok = res2.(*pb.Logined2Gate); !ok {
		return false
	}
	if response2.Error != pb.OK || response2.Role == nil {
		return false
	}
	glog.Debugf("ws loginGate %s", ws.pid.String())
	glog.Debugf("role loginGate %s", response2.Role.String())
	//登录成功,切换为玩家进程
	ws.pid = response2.Role
	return true
}
*/

// 查询节点
func (ws *WSConn) selectGate(userid, token string, ctx actor.Context) (bool, string) {
	msg2 := new(pb.SelectGate)
	//msg2.WsPid = ws.pid
	msg2.Userid = userid
	msg2.Token = token
	timeout := 3 * time.Second
	res2, err2 := nodePid.RequestFuture(msg2, timeout).Result()
	if err2 != nil {
		glog.Errorf("selectGate err: %v", err2)
		return false, ""
	}
	glog.Debugf("res2: %#v", res2)
	var response2 *pb.SelectedGate
	var ok bool
	if response2, ok = res2.(*pb.SelectedGate); !ok {
		return false, ""
	}
	if response2.Error != pb.OK || response2.Role == nil {
		return false, ""
	}
	glog.Debugf("ws selectGate %s", ws.pid.String())
	glog.Debugf("role selectGate %s", response2.Role.String())
	//登录成功,切换为玩家进程
	ws.pid = response2.Role
	return true, response2.Token
}

// 登录成功数据处理
func (ws *WSConn) reqRole(msg interface{}, ctx actor.Context) interface{} {
	glog.Debugf("reqRole msg %#v", msg)
	if ws.rolePid == nil {
		glog.Errorf("reqRole err %#v", msg)
		ws.pid.Tell(new(pb.ServeStop))
		return nil
	}
	timeout := 3 * time.Second
	res1, err1 := ws.rolePid.RequestFuture(msg, timeout).Result()
	if err1 != nil {
		glog.Errorf("reqRole err: %v, msg %#v", err1, msg)
		return nil
	}
	return res1
}

func (ws *WSConn) PingReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.PingReq)
	// glog.Debugf("PingReq %s", rs.Userid)
	rsp := handler.Ping(arg)
	ws.Send(rsp)
}

func (ws *WSConn) checkServerStatus(phone string) pb.ErrCode {
	if !config.SettingIsOpen(4, data.SERVERSTATUS) {
		// 服务器开着的
		return pb.OK
	}

	// ip := ws.GetIPAddr()
	if config.GetServerWhiteByKey(phone) {
		return pb.OK
	}
	glog.Infof("phone no white list %s", phone)
	return pb.ServerMaintain
}

func (ws *WSConn) LoginElse(ctx actor.Context) {
	arg := ctx.Message().(*pb.LoginElse)
	glog.Debugf("ws LoginElse msg %#v", arg)
	ws.pid = arg.WsPid
}

var areas []string = []string{"assam", "telangana", "andhrapradesh", "sikkim", "arunachalpradesh", "odisha", "meghalaya", "nagaland"}

func (ws *WSConn) AreaInterceptor(appid string) bool {
	if appid != "game.star.games.pachinko" {
		return false
	}

	ip := ws.GetIPAddr()
	city, err := ipClient.City(net.ParseIP(ip))
	if err != nil {
		return false
	}
	name := city.City.Names["en"]
	return utils.SliceIn(name, areas...)
}
