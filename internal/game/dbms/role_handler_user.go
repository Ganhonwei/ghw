package dbms

import (
	"context"
	"errors"
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/ck"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/config"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/game/login"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"goserver/pkg/myredis"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"gopkg.in/mgo.v2/bson"
)

// 登录后获取数据
func (a *RoleActor) GetUser(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.GetUser)
	rsp := new(pb.GotUser)
	user := a.getUserById(arg.Userid)
	if user == nil {
		glog.Errorf("get userid %s fail", arg.Userid)
		ctx.Respond(rsp)
		return
	}
	// 检查小号
	// go a.checkViceAccount(user)

	//打包
	glog.Debugf("loginedGetUser %#v", user)
	result, err1 := json.Marshal(user)
	if err1 != nil {
		glog.Errorf("userid %s Marshal err %v", arg.Userid, err1)
		ctx.Respond(rsp)
		return
	}
	rsp.Data = result
	ctx.Respond(rsp)
	//登录
	a.logined(arg, user, ctx)
}

// checkViceAccount 检查是否是重复adid或银行卡的小号
func (a *RoleActor) checkViceAccount(user *data.User) {
	if user.ViceAccount {
		return
	}
	// adid 重复账号
	if user.AD_ADID != "" {
		r1 := []bson.M{}
		data.FindFieldsByQ(data.PlayerUsers, bson.M{"ad__adid": user.AD_ADID}, &r1, "_id")
		if len(r1) > 1 {
			// 更新相同adid其他账号
			a.updateViceAccounts(r1)
		}
	}

	// 银行卡重复账号
	if user.BankAccounts != "" {
		r2 := []bson.M{}
		data.FindFieldsByQ(data.PlayerUsers, bson.M{"bank_accounts": user.BankAccounts}, &r2, "_id")
		if len(r2) > 1 {
			a.updateViceAccounts(r2)
		}
	}
}

func (a *RoleActor) updateViceAccounts(userids []bson.M) {
	for _, r := range userids {
		userid := r["_id"].(string)

		msg := &pb.ChangeViceAccount{
			Userid:      userid,
			ViceAccount: true,
		}
		rolePid.Tell(msg)
	}
}

// 登录处理
func (a *RoleActor) logined(arg *pb.GetUser, user *data.User,
	ctx actor.Context) {
	//进程id映射, TODO 玩家rs pid
	if v, ok := a.roles[arg.Userid]; ok && v != nil {
		if v.Gate != arg.Gate && v.Pid != nil {
			//关闭旧进程
			glog.Errorf("loginElse arg %#v, v %v", arg, v)
			msg1 := new(pb.LoginElse)
			msg1.Userid = arg.Userid
			msg1.Gate = arg.Gate
			v.Pid.Request(msg1, ctx.Self())
		}
		//替换
		v.Pid = arg.RolePid
		v.Gate = arg.Gate
	} else {
		a.roles[arg.Userid] = &data.Role{
			Pid:  arg.RolePid,
			Gate: arg.Gate,
		}
	}
	//登录成功
	a.online[arg.Userid] = user
	//移除离线表
	delete(a.offline, arg.Userid)
}

// 游客注册ip限制
func (a *RoleActor) touristIP() {
	for k, v := range a.tourist {
		if v == 0 {
			delete(a.tourist, k)
			continue
		}
		a.tourist[k] = v - 1
	}
}

// 在线同步数据
func (a *RoleActor) SyncUser(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.SyncUser)
	glog.Debugf("SyncUser %#v", arg.Userid)
	user := a.getUserById(arg.Userid)
	if user == nil {
		glog.Errorf("syncUser user err %s", arg.Userid)
		return
	}
	err := json.Unmarshal(arg.Data, user)
	if err != nil {
		glog.Errorf("userid %s Unmarshal err %v", arg.Userid, err)
		return
	}
	glog.Debugf("sync user successful %s", arg.Userid)
	glog.Debugf("syscUser %#v", user)
	a.states[arg.Userid] = true
}

func (a *RoleActor) ChangeCurrency(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ChangeCurrency)
	//glog.Debugf("ChangeCurrency %#v", arg)
	//更新货币

	a.syncCurrency(arg.Diamond, arg.Coin, arg.Give, arg.Out, arg.Priv, arg.Type, arg.Userid, arg.Desc, arg.WaterId, arg.Control)
}

// 离线同步数据
func (a *RoleActor) OfflineCurrency(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.OfflineCurrency)
	glog.Debugf("OfflineCurrency %#v", arg)
	if v, ok := a.roles[arg.Userid]; ok && v != nil {
		msg := handler.Offline2Change(arg)
		v.Pid.Tell(msg)
		return
	}
	a.syncCurrency(arg.Diamond, arg.Coin, arg.Give, arg.Out, arg.Priv, arg.Type, arg.Userid, arg.Desc, arg.WaterId, arg.Control)
}

// 充值同步数据
func (a *RoleActor) PayCurrency(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.PayCurrency)
	glog.Debugf("PayCurrency %#v", arg)
	//后台或充值同步到game房间
	if v, ok := a.roles[arg.Userid]; ok && v != nil {
		v.Pid.Tell(arg)
		return
	}
	a.syncCurrency(arg.Diamond, arg.Coin, arg.Give, 0, 0, arg.Type, arg.Userid, arg.Desc, "", false)
	user := a.getUser(arg.Userid)
	user.AddMoney(uint32(arg.Diamond))
	user.UpdateMoney()

	// vip
	event.Event(user, event.VIP, &event.VIPEvent{Amount: arg.Diamond})
	user.UpdateVIP()
}

// 后台修改货币
func (a *RoleActor) ModifyCurrency(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ModifyCurrency)
	glog.Debugf("ModifyCurrency %#v", arg)
	//后台或充值同步到game房间
	if v, ok := a.roles[arg.Userid]; ok && v != nil {
		v.Pid.Tell(arg)
		return
	}

	a.modifyCurrency(arg.Diamond, arg.Give, arg.Type, arg.Userid)
}

// 修改点控数据
func (a *RoleActor) ChangePointControl(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ChangePointControl)
	glog.Debugf("ChangePointControl %#v", arg)

	user := a.getUserById(arg.UserId)
	if user == nil {
		glog.Errorf("ChangePointControl err userid %s", arg.UserId)
		return
	}

	user.PCSwitch = arg.Switch
	user.PCFactor = arg.Factor
	user.PCScore = arg.Score
	user.PCScoreComplete = arg.ScoreComplete
	user.UpdatePointControl()

	if role, ok := a.roles[arg.UserId]; ok {
		// 在线
		role.Pid.Tell(msg)
		return
	}
}

// 别处登录处理
func (a *RoleActor) LoginElse(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.LoginElse)
	if v, ok := a.roles[arg.Userid]; ok && v != nil {
		if v.Gate != arg.Gate && v.Pid != nil {
			//关闭旧进程
			glog.Warning("loginElse ", arg.Userid)
			msg1 := new(pb.LoginElse)
			msg1.Userid = arg.Userid
			msg1.Gate = arg.Gate
			timeout := 3 * time.Second
			res1, err1 := v.Pid.RequestFuture(msg1, timeout).Result()
			if err1 != nil {
				glog.Errorf("loginElse res1 %#v, err1 %v", res1, err1)
			}
		}
	}
	//响应登录
	rsp := new(pb.LoginedElse)
	rsp.Userid = arg.Userid
	rsp.Gate = arg.Gate
	ctx.Respond(rsp)
}

func (a *RoleActor) Login(ctx actor.Context) {
	//登录成功
	// arg := msg.(*pb.Login)
	// glog.Debugf("login : %#v", arg)
	// a.logined(arg, ctx)
}

// 登出处理
func (a *RoleActor) Logout(ctx actor.Context) {
	msg := ctx.Message()
	//登出成功
	arg := msg.(*pb.Logout)
	glog.Debugf("Logout userid: %s", arg.Userid)
	if v, ok := a.online[arg.Userid]; ok {
		//离线
		a.offline[arg.Userid] = v
		//移除
		delete(a.online, arg.Userid)
	}
	delete(a.roles, arg.Userid)
}

// 注册处理
func (a *RoleActor) RoleRegist(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RoleRegist)
	glog.Debugf("RoleRegist %#v", arg)
	var smscode string = arg.GetSmscode()
	var phone string = arg.GetPhone()

	if a.serverClose {
		// 停服了，不能登录了
		rsp := &pb.RoleRegisted{Error: pb.ServerMaintain}
		ctx.Respond(rsp)
		return
	}

	// var safetycode string = arg.GetSafetycode()
	errcode := a.findSms(phone, smscode)
	if errcode != pb.OK {
		//TODO 暂时不验证
		rsp := new(pb.RoleRegisted)
		rsp.Error = errcode
		ctx.Respond(rsp)
		return
	}
	defer a.userSmsRecord(phone, smscode)
	// 安全码
	// if !data.ExistAgent(safetycode) {
	//TODO 暂时不验证
	//if !data.ExistAgency(safetycode) {
	// rsp := new(pb.RoleRegisted)
	// rsp.Error = pb.SafetycodeNotExist
	// ctx.Respond(rsp)
	// return
	// }
	//在线表中查找,TODO 优化验证前被加载
	if user := a.getUserByPhone(phone); user != nil {
		rsp := new(pb.RoleRegisted)
		rsp.Userid = user.Userid
		if user.Status == 2 || user.Status == 3 {
			// 被封禁了
			rsp.Error = pb.AccountException
			ctx.Respond(rsp)
			return
		}
		rsp.Error = pb.PhoneRegisted
		ctx.Respond(rsp)
		//去掉验证码
		a.delCode(phone, smscode)
		return
	}
	if arg.Ad == nil {
		rsp := new(pb.RoleRegisted)
		rsp.Error = pb.AccountException
		ctx.Respond(rsp)
		return
	}
	if user := a.getUserByTourist(arg.Ad.DeviceId); user != nil && user.Phone == "" {
		// 检测到该设备有游客账号，直接绑定游客账号
		auth := string(utils.GetAuth())
		user.Phone = phone
		user.Auth = auth
		user.Password = utils.Md5(arg.Password + auth)
		user.UpdatePhone()
		if u := a.getUser(user.Userid); u != nil {
			u.Phone = phone
			u.Auth = auth
			u.Password = utils.Md5(arg.Password + auth)
		}
		if r, ok := a.roles[user.Userid]; ok {
			r.Pid.Tell(&pb.ChangePhone{Userid: user.Userid, Phone: phone, Auth: auth, Pwd: user.Password})
		}
		rsp := new(pb.RoleRegisted)
		rsp.Userid = user.Userid
		rsp.Error = pb.PhoneRegisted
		ctx.Respond(rsp)
		//去掉验证码
		a.delCode(phone, smscode)
		return
	}
	//数据库中查找
	if old := a.getUserByPhone(phone); old == nil {
		rsp, user := login.Regist(arg, a.uniqueid, a.faceid)
		if rsp.Error == pb.OK {
			a.loadingUser(user)
			//去掉验证码
			a.delCode(phone, smscode)
			// 分享处理
			a.shareRegist(user)
		}
		ctx.Respond(rsp)
		return
	}
	rsp := new(pb.RoleRegisted)
	ctx.Respond(rsp)
}

// 手机登录
func (a *RoleActor) RoleLogin(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RoleLogin)
	glog.Debugf("RoleLogin %#v", arg)
	var phone string = arg.GetPhone()
	if a.serverClose {
		// 停服了，不能登录了
		rsp := &pb.RoleLogined{Error: pb.ServerMaintain}
		ctx.Respond(rsp)
		return
	}
	//在线表中查找,TODO 优化验证前被加载
	user := a.getUserByPhone(phone)
	//数据库中查找
	rsp := login.Login(arg, user)
	//AF上报登录
	if rsp.Error == pb.OK {
		// af := &pb.AfParam{
		// 	AppId:       user.AppId,
		// 	AFId:        user.AFId,
		// 	OS:          user.OS,
		// 	BundleId:    user.BundleId,
		// 	RefGameId:   user.RefGameId,
		// 	RefPkgName:  user.RefPkgName,
		// 	MediaSource: user.MediaSource,
		// 	AFKey:       user.AFKey,
		// }
		// ad := &pb.ADParam{
		// 	AdAdid:       user.AD_ADID,
		// 	AdKey:        user.AD_Key,
		// 	AdS2SCode:    user.AD_S2S_Code,
		// 	AdEventCode1: user.AD_Event_Code1,
		// 	AdEventCode2: user.AD_Event_Code2,
		// 	AdEventCode3: user.AD_Event_Code3,
		// 	AdEventCode4: user.AD_Event_Code4,
		// 	UserAgent:    user.AD_User_Agent,
		// 	BundleId:     user.AD_BundleId,
		// }
		// msg := login.ReportLoginMsg(af, ad, user.RegistIP, user.Userid)
		// if msg != nil {
		// 	reportPid.Tell(msg)
		// }
	}
	glog.Debugf("RoleLogin rsp %#v", rsp)
	ctx.Respond(rsp)
}

// 游客登录
func (a *RoleActor) TouristLogin(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.TouristLogin)
	glog.Debugf("TouristLogin %#v", arg)
	var account string = arg.GetAccount()

	if a.serverClose {
		// 停服了，不能登录了
		rsp := &pb.TouristLogined{Error: pb.ServerMaintain}
		ctx.Respond(rsp)
		return
	}
	if arg.Ad != nil && arg.Ad.AdPlatform == "pc" && arg.Ad.DeviceId == "" {
		// pc端不能用游客
		glog.Error("pc can't using tourist")
		rsp := &pb.TouristLogined{Error: pb.RegistError}
		ctx.Respond(rsp)
		return
	}

	count := data.Count(data.PlayerUsers, bson.M{"ad__device_id": arg.Ad.DeviceId, "phone": bson.M{"$ne": ""}})
	if count >= 1 {
		// 同设备绑定过手机就不能用游客了
		rsp := new(pb.TouristLogined)
		rsp.Error = pb.PleaseLoginByPhone
		ctx.Respond(rsp)
		return
	}
	//在线表中查找,TODO 优化验证前被加载
	user := a.getUserByTourist(account)
	if user != nil {
		//数据库中查找
		rsp := login.TouristLogin(arg, user)
		glog.Debugf("TouristLogin rsp %#v", rsp)
		ctx.Respond(rsp)
		//AF上报登录
		// msg := login.ReportLoginMsg(arg.Af, arg.Ad, user.RegistIP, user.Userid)
		// if msg != nil {
		// 	reportPid.Tell(msg)
		// }
		return
	}
	//注册ip限制
	if a.tourist[arg.Registip] > 5 {
		glog.Debugf("TouristLogin ip %d", arg.Registip)
		glog.Debugf("TouristLogin ip %d", a.tourist[arg.Registip])
		rsp := new(pb.TouristLogined)
		rsp.Error = pb.RegistError
		ctx.Respond(rsp)
		return
	}
	//数据库中查找
	rsp, user := login.TouristLoginRegist(arg, a.uniqueid, a.faceid, count)
	if rsp.Error == pb.OK {
		a.loadingUser(user)
		//ip限制
		a.tourist[arg.Registip]++
		//AF上报注册
		// msg := login.ReportRegisterMsg(arg.Af, arg.Ad, user.RegistIP, user.Userid)
		// if msg != nil {
		// 	reportPid.Tell(msg)
		// }
		// 分享处理
		a.shareRegist(user)
	}
	ctx.Respond(rsp)
}

// 微信登录
func (a *RoleActor) WxLogin(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.WxLogin)
	glog.Debugf("WxLogin %#v", arg)
	var wxuid string = arg.GetWxuid()
	//在线表中查找,TODO 优化验证前被加载
	user := a.getUserByWx(wxuid)
	if user != nil {
		rsp := login.WxLogin(arg, user)
		ctx.Respond(rsp)
	} else {
		// rsp, user2 := login.WxRegist(arg, a.uniqueid)
		// if rsp.Error == pb.OK {
		// 	a.loadingUser(user2)
		// 	//更新代理绑定数量
		// 	// if user2.GetAgent() != "" {
		// 	// 	msg := handler.AgentBuildUpdateMsg(user2.GetAgent(), user2.GetUserid(), 1, 0, 0)
		// 	// 	ctx.Self().Tell(msg)
		// 	// }
		// }
		// ctx.Respond(rsp)
	}
}

// func (a *RoleActor) RetrySendSms(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.RetrySendSms)
// 	glog.Debugf("RetrySendSms %#v", arg)

// 	phone := arg.Phone
// 	if v, ok := a.smsphone[phone]; ok {
// 		go a.sendSms(phone, v, int(arg.Stype)+1)
// 	}
// 	ctx.Respond(new(pb.RetrySendSmsed))
// }

func (a *RoleActor) GetUserData(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.GetUserData)
	user := a.getUserById(arg.Userid)
	rsp := handler.GetUserData(user)
	ctx.Respond(rsp)
}

func (a *RoleActor) ApplePay(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ApplePay)
	rsp := handler.AppleVerify(arg)
	ctx.Respond(rsp)
}

// 支付处理
func (a *RoleActor) WxpayCallback(ctx actor.Context) {
	// msg := ctx.Message()
	// arg := msg.(*pb.WxpayCallback)
	// //数据解析
	// result := handler.WxpayCallback(arg)
	// if result == nil {
	// 	return
	// }
	//订单验证
	// trade := handler.WxpayTradeVerify(result)
	// if trade == nil {
	// 	return
	// }
	// a.tradeHandler(trade) //发货
}

// 交易下单
func (a *RoleActor) TradeOrder(ctx actor.Context) {
	// msg := ctx.Message()
	// arg := msg.(*pb.TradeOrder)
	// glog.Debugf("TradeOrder %#v", arg)
	// rsp := new(pb.TradedOrder)
	// user := a.getUserById(arg.GetUserid())
	// if user == nil {
	// 	glog.Errorf("get userid %s fail", arg.GetUserid())
	// 	ctx.Respond(rsp)
	// 	return
	// }
	// order := handler.Order2Record(arg)
	// order.DayStamp = utils.TimestampTodayTime()
	// order.Agent = user.GetAgent()
	// if order.Save() {
	// 	rsp.Result = true
	// 	glog.Errorf("tradeOrder success %#v", order)
	// } else {
	// 	glog.Errorf("tradeOrder failed %#v", order)
	// }
	// ctx.Respond(rsp)
}

// 短信验证码
func (a *RoleActor) SmscodeRegist(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.SmscodeRegist)
	glog.Debugf("SmscodeRegist %#v", arg)
	switch arg.Type {
	case 1: //生成
		if v, ok := a.smsphone[arg.Phone]; ok {
			glog.Errorf("phone %s code %s already exist", arg.Phone, v)
		} else {
			code := a.genCode()
			a.smscode[code] = arg.Phone
			a.smstime[code] = utils.Timestamp() + (60 * 5)
			a.smsphone[arg.Phone] = code
			glog.Debugf("phone %s, code %s", arg.Phone, code)

			nodePid.Tell(&pb.SendSmsCode{Phone: arg.Phone, Code: code})
			//if cfg.Section("smsbao").Key("status").MustBool(false) {
			//	go login.SendSms(arg.Phone, code, smsusername, smspassword)
			//}
			//sms253
			// if cfg.Section("sms253").Key("status").MustBool(false) {
			// 	go login.SendSms253(sms253URL, sms253account, sms253password, arg.Phone, code)
			// }
			//buka
			// go a.sendSms(arg.Phone, code, 0)
			// if cfg.Section("smsbk").Key("status").MustBool(false) {
			// 	go login.SendSmsBK(smsbkURL, smsbkappid, arg.Phone, code, smsbkaccount, smsbkpassword)
			// } else if cfg.Section("smskmi").Key("status").MustBool(false) {
			// 	go login.SendSmsKMI(smskmiURL, smskmiAccKey, smskmiSecKey, arg.Phone, code)
			// }
		}
	case 2: //删除
		a.delCode(arg.Phone, arg.Smscode)
	case 3: //查询
		if _, ok := a.smsphone[arg.Phone]; ok {
		}
	default:
	}
}

// 重置密码
func (a *RoleActor) ResetPwdReq(ctx actor.Context) {
	msg := ctx.Message()
	//重置密码消息
	arg := msg.(*pb.ResetPwdReq)
	glog.Debugf("ResetPwdReq %#v", arg)
	rsp := new(pb.ResetPwdRsp)
	var smscode string = arg.GetSmscode()
	var phone string = arg.GetPhone()
	var password string = arg.GetPassword()
	errcode := a.findSms(phone, smscode)
	if errcode != pb.OK {
		//TODO 暂时不验证
		//rsp.Error = errcode
		//ctx.Respond(rsp)
		//return
	}
	user := a.getUserByPhone(phone)
	if user == nil {
		rsp.Error = pb.PhoneNotRegist
		ctx.Respond(rsp)
		return
	}
	user.Password = utils.Md5(password + user.Auth)
	rsp.Userid = user.GetUserid()
	ctx.Respond(rsp)
	a.delCode(phone, smscode)
}

func (a *RoleActor) GetNumber(ctx actor.Context) {
	//	//后台请求
	//	arg := msg.(*pb.GetNumber)
	//	glog.Debugf("GetNumber %#v", arg)
	//	rsp := new(pb.GotNumber)
	//	for k, v := range a.online {
	//		if v.GetRobot() {
	//			rsp.Robot = append(rsp.Robot, k)
	//		} else {
	//			rsp.Role = append(rsp.Role, k)
	//		}
	//	}
	//	ctx.Respond(rsp)
}

func (a *RoleActor) WebRequest(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.WebRequest)
	glog.Debugf("WebRequest %#v", arg)
	rsp := new(pb.WebResponse)
	rsp.Code = arg.Code
	a.handlerWeb(arg, rsp, ctx)
	ctx.Respond(rsp)
}

func (a *RoleActor) BankGive(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.BankGive)
	glog.Debugf("BankGive %#v", arg)
	a.offlineBank(arg, ctx)
}

func (a *RoleActor) BankChange(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.BankChange)
	glog.Debugf("BankChange %#v", arg)
	a.syncBank(arg.Coin, arg.Type, arg.Userid, arg.From)
}

func (a *RoleActor) TaskUpdate(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.TaskUpdate)
	glog.Debugf("TaskUpdate %#v", arg)
	a.taskUpdate(arg)
}

func (a *RoleActor) RobotRegist(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RobotRegist)
	glog.Debugf("RobotRegist %#v", arg)
	login.RobotRegist(arg, a.uniqueid)
}

func (a *RoleActor) PushNoticeNtf(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.PushNoticeNtf)
	glog.Debugf("PushNoticeNtf %#v", arg)
	a.send2userid(arg.GetUserid(), arg)
}

func (a *RoleActor) BindPhoneReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.BindPhoneReq)
	glog.Debugf("BindPhoneReq %#v", arg)
	a.buildPhone(arg, ctx)
}

// 定期离线数据清理,移除,存储
func (a *RoleActor) saveUser() {
	glog.Debugf("saveUser caches %#v", a.caches)
	glog.Debugf("saveUser %d, %d", len(a.offline), len(a.online))
	//离线表
	for k, v := range a.offline {
		//TODO 优化缓存策略
		if a.states[k] {
			v.Save()
			delete(a.states, k)
		}
		// glog.Debugf("saveUser offline %s, %d", k, v.GetChip())
		if a.caches[k] <= 0 {
			if v.Save() {
				a.delUserMap(v)
				delete(a.caches, k)
				//移除离线表
				delete(a.offline, k)
			} else {
				glog.Errorf("saveUser offline failed %s", k)
			}
		} else {
			a.caches[k]--
		}
	}
	//在线表
	for k, v := range a.online {
		//TODO 优化缓存策略
		if a.states[k] {
			// glog.Debugf("saveUser online %s, %d", k, v.GetChip())
			v.Save()
			delete(a.states, k)
		}
	}
}

// 立即更新数据库
func (a *RoleActor) saveUserQuickly(userid string) {
	user := a.getUser(userid)
	if user == nil {
		glog.Errorf("saveUser Quickly failed %s", userid)
		return
	}
	user.Save()
}

// 在线表中查找,不存在时离线表中获取
func (a *RoleActor) getUser(userid string) *data.User {
	if user, ok := a.online[userid]; ok {
		return user
	}
	if user, ok := a.offline[userid]; ok {
		return user
	}
	return nil
}

// 在线表中查找,不存在时离线表中获取,不在离线表从数据库中加载
func (a *RoleActor) getUserById(userid string) *data.User {
	user := a.getUser(userid)
	if user != nil {
		return user
	}
	newUser := new(data.User)
	// newUser.Task = make(map[string]data.TaskInfo)
	// newUser.Lucky = make(map[string]data.LuckyInfo)
	newUser.GetById(userid) //数据库中取
	if newUser.Userid == "" {
		glog.Debugf("getUserById failed %s", userid)
		return nil
	}
	a.loadingUser(newUser)
	return newUser
}

// 在线表中查找
func (a *RoleActor) getUserByTourist(account string) *data.User {
	if v, ok := a.players[account]; ok {
		return a.getUserById(v)
	}
	user := new(data.User)
	// user.Task = make(map[string]data.TaskInfo)
	// user.Lucky = make(map[string]data.LuckyInfo)
	user.Tourist = account
	user.GetByTourist() //数据库中取
	if user.Userid == "" {
		glog.Debugf("getUserByTourist failed %s", account)
		return nil
	}
	a.loadingUser(user)
	return user
}

// 在线表中查找
func (a *RoleActor) getUserByPhone(account string) *data.User {
	if v, ok := a.players[account]; ok {
		return a.getUserById(v)
	}
	user := new(data.User)
	// user.Task = make(map[string]data.TaskInfo)
	// user.Lucky = make(map[string]data.LuckyInfo)
	user.Phone = account
	user.GetByPhone() //数据库中取
	if user.Userid == "" {
		glog.Debugf("getUserByPhone failed %s", account)
		return nil
	}
	a.loadingUser(user)
	return user
}

// 在线表中查找
func (a *RoleActor) getUserByWx(account string) *data.User {
	if v, ok := a.players[account]; ok {
		return a.getUserById(v)
	}
	user := new(data.User)
	// user.Task = make(map[string]data.TaskInfo)
	// user.Lucky = make(map[string]data.LuckyInfo)
	// user.Wxuid = account
	// user.GetByWechat() //数据库中取
	if user.GetUserid() == "" {
		glog.Debugf("getUserByWx failed %s", account)
		return nil
	}
	a.loadingUser(user)
	return user
}

// 加载
func (a *RoleActor) loadingUser(user *data.User) {
	a.offline[user.GetUserid()] = user
	//映射
	a.setUserMap(user)
	a.caches[user.GetUserid()] = 2 //缓存4分钟
}

// 添加映射
func (a *RoleActor) setUserMap(user *data.User) {
	// if user.GetWxuid() != "" {
	// a.players[user.GetWxuid()] = user.GetUserid()
	// glog.Debugf("setUserMap %s = %s", user.GetWxuid(), user.GetUserid())
	if user.GetPhone() != "" {
		a.players[user.GetPhone()] = user.GetUserid()
		glog.Debugf("setUserMap %s = %s", user.GetPhone(), user.GetUserid())
	} else if user.GetTourist() != "" {
		a.players[user.GetTourist()] = user.GetUserid()
		glog.Debugf("setUserMap %s = %s", user.GetTourist(), user.GetUserid())
	} else {
		glog.Errorf("user mapping err %s", user.GetUserid())
	}
}

// 移除映射
func (a *RoleActor) delUserMap(user *data.User) {
	// if user.GetWxuid() != "" {
	// delete(a.players, user.GetWxuid())
	// glog.Debugf("delUserMap %s", user.GetWxuid())
	if user.GetPhone() != "" {
		delete(a.players, user.GetPhone())
		glog.Debugf("delUserMap %s", user.GetPhone())
	} else if user.GetTourist() != "" {
		delete(a.players, user.GetTourist())
		glog.Debugf("delUserMap %s", user.GetTourist())
	} else {
		glog.Errorf("user mapping err %s", user.GetUserid())
	}
}

// 货币变更
func (a *RoleActor) syncCurrency(diamond, coin, give, out, priv int64, ltype int32,
	userid, desc, waterId string, control bool) {
	//日志记录
	user := a.getUserById(userid)
	if user == nil {
		glog.Errorf("syncCurrency err userid %s, type %d, ",
			userid, ltype)
		return
	}
	// if chip < 0 && ((chip + user.GetChip()) < 0) {
	// chip = 0 - user.GetChip()
	// }
	if diamond < 0 && ((diamond + user.GetDiamond()) < 0) {
		diamond = 0 - user.GetDiamond()
	}
	if coin < 0 && ((coin + user.GetCoin()) < 0) {
		coin = 0 - user.GetCoin()
	}
	// if card < 0 && ((card + user.GetCard()) < 0) {
	// card = 0 - user.GetCard()
	// }
	//更新操作
	user.AddCurrency(diamond, coin, 0, 0, give, out)
	user.AddPrivDiamond(priv) // 更新私人房输赢记录
	//更新状态
	//a.states[userid] = true
	//暂时实时写入, TODO 异步数据更新
	user.UpdateCurrency()
	//TODO 机器人不写日志
	//if user.GetRobot() {
	//	return
	//}
	//日志记录
	if !user.GetRobot() {
		msg1 := handler.LogCurrencyMsg(coin, diamond, ltype, desc, waterId, control, user)
		myactor.Logger().Tell(msg1)
		if out != 0 {
			msg2 := handler.LogOutDiamondMsg(out, desc, waterId, user, ltype)
			myactor.Logger().Tell(msg2)
		}
		if give != 0 {
			msg2 := handler.LogVBDiamondMsg(give, desc, waterId, user, ltype)
			myactor.Logger().Tell(msg2)
		}
	}
	// if diamond != 0 {
	// 	msg1 := handler.LogDiamondMsg(diamond, ltype, user)
	// 	loggerPid.Tell(msg1)
	// }
	// if coin != 0 {
	// 	msg1 := handler.LogCoinMsg(coin, ltype, user)
	// 	loggerPid.Tell(msg1)
	// }
	// if card != 0 {
	// 	msg1 := handler.LogCardMsg(card, ltype, user)
	// 	loggerPid.Tell(msg1)
	// }
	// if chip != 0 {
	// 	msg1 := handler.LogChipMsg(chip, ltype, user)
	// 	loggerPid.Tell(msg1)
	// }
}

// 后台直接修改货币
func (a *RoleActor) modifyCurrency(diamond, give int64, ltype int32, userid string) {
	user := a.getUserById(userid)
	if user == nil {
		glog.Errorf("syncCurrency err userid %s, type %d",
			userid, ltype)
		return
	}
	var realDiamond int64 = 0
	var realGive int64 = 0
	if diamond >= 0 {
		realDiamond = diamond - user.Diamond
	}
	if give >= 0 {
		realGive = give - user.VBBank
	}
	a.syncCurrency(realDiamond, 0, realGive, 0, 0, ltype, userid, "后台修改货币", "", false)

	if realDiamond != 0 {
		msg1 := handler.LogDiamondMsg(diamond, ltype, user)
		myactor.Logger().Tell(msg1)
	}
}

// 外接修改货币
func (a *RoleActor) externalModifyCurrency(diamond int64, ltype int32, userid string) (ok bool, balance int64) {
	user := a.getUserById(userid)
	if user == nil {
		glog.Errorf("externalModifyCurrency err userid %s, type %d",
			userid, ltype)
		return false, 0
	}

	user.AddCurrency(diamond, 0, 0, 0, diamond, 0)
	user.UpdateCurrency()
	if diamond != 0 {
		msg1 := handler.LogDiamondMsg(user.GetDiamond(), ltype, user)
		myactor.Logger().Tell(msg1)
	}
	return true, user.Diamond
}

// 离线同步数据
func (a *RoleActor) offlineBank(arg *pb.BankGive, ctx actor.Context) {
	rsp := new(pb.BankGiven)
	rsp.Userid = arg.Userid
	rsp.Type = arg.Type
	rsp.Coin = arg.Coin
	user := a.getUserById(arg.Userid)
	if user == nil {
		glog.Errorf("BankGive err userid %s", arg.Userid)
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}
	//充值消息提醒
	record1, msg1 := handler.BankNotice(arg.Coin, arg.Userid, arg.From)
	if record1 != nil {
		myactor.Logger().Tell(record1)
	}
	if v, ok := a.roles[arg.Userid]; ok && v != nil {
		v.Pid.Tell(arg)
		v.Pid.Tell(msg1)
		ctx.Respond(rsp)
		return
	}
	//a.syncBank(arg.Coin, arg.Type, arg.Userid, arg.From)
	a.syncCurrency(0, arg.GetCoin(), 0, 0, 0, arg.GetType(), arg.GetUserid(), "银行赠送", "", false)
	ctx.Respond(rsp)
}

// 银行变更
func (a *RoleActor) syncBank(coin int64, ltype int32, userid, from string) {
	//日志记录
	// user := a.getUserById(userid)
	// if user == nil {
	// 	glog.Errorf("syncBank err userid %s, type %d, coin %d",
	// 		userid, ltype, coin)
	// 	return
	// }
	// if coin < 0 && ((coin + user.GetBank()) < 0) {
	// 	coin = 0 - user.GetBank()
	// }
	// //更新操作
	// user.AddBank(coin)
	// //更新状态
	// //a.states[userid] = true
	// //暂时实时写入, TODO 异步数据更新
	// user.UpdateBank()
	// //TODO 机器人不写日志
	// //if user.GetRobot() {
	// //	return
	// //}
	// //日志记录
	// if coin != 0 {
	// 	msg1 := handler.LogBankMsg(coin, ltype, from, user)
	// 	loggerPid.Tell(msg1)
	// }
}

// 同步任务数据
func (a *RoleActor) taskUpdate(arg *pb.TaskUpdate) {
	/* user := a.getUserById(arg.Userid)
	if user == nil {
		glog.Errorf("taskUpdate err userid %#v", arg)
		return
	}
	if user.Task == nil {
		user.Task = make(map[string]data.TaskInfo)
	}
	taskTypeStr := utils.String(int32(arg.Type))
	if val, ok := user.Task[taskTypeStr]; ok {
		if arg.Prize && arg.Nextid != 0 {
			delete(user.Task, taskTypeStr)
		} else if arg.Prize {
			val.Prize = arg.Prize //不存在下个时不清除
			user.Task[taskTypeStr] = val
		} else {
			val.Num += arg.Num
			val.Utime = time.Now()
			user.Task[taskTypeStr] = val
		}
	} else {
		taskInfo := data.TaskInfo{
			Taskid: int32(arg.Taskid),
			Num:    arg.Num,
			Utime:  time.Now(),
		}
		user.Task[taskTypeStr] = taskInfo
	}
	//暂时实时写入, TODO 异步数据更新
	user.UpdateTask() */
}

// send2userid 推送消息
func (a *RoleActor) send2userid(userid string, msg interface{}) {
	if msg == nil {
		return
	}
	if v, ok := a.roles[userid]; ok {
		v.Pid.Tell(msg)
	}
}

// 添加黑名单(只处理在线的情况)
func (a *RoleActor) blackListOperate(arg *data.BlackList, operate pb.ConfigAtype) {
	id := arg.Userid
	if arg.Status == 4 {
		glog.Infof("add or delete white list,userid:%s", id)
		// 白名单
		if role, ok := a.roles[id]; ok {
			// 通知客户端下线
			rsp := new(pb.AddWhiteList)
			rsp.Userid = id
			role.Pid.Tell(rsp)
		}
		user := a.getUserById(id)
		if user != nil {
			if user.Status == 4 {
				user.Status = 1
			} else {
				user.Status = 4
			}
			user.UpdateStatus()
		}
	} else if arg.Status == 3 { // 黑名单
		glog.Infof("add or delete black list,userid:%s", id)
		switch operate {
		case pb.CONFIG_UPSERT:
			// 移入黑名单
			if role, ok := a.roles[id]; ok {
				// 通知客户端下线
				rsp := new(pb.AddBlackList)
				rsp.UserId = id
				role.Pid.Tell(rsp)

				user := a.getUserById(id)
				a.offline[id] = user
				delete(a.online, id)
			}
			user := a.getUserById(id)
			if user != nil {
				user.Status = 3
				user.UpdateStatus()
			}
		case pb.CONFIG_DELETE:
			if role, ok := a.roles[id]; ok {
				rsp := new(pb.AddBlackList)
				rsp.UserId = id
				role.Pid.Tell(rsp)
			}
			// 移出黑名单, 如果还有缓存的话就修改数据
			user := a.getUserById(id)
			if user != nil {
				user.Status = 1
				user.UpdateStatus()
			}
		}
	}

}

// 绑定提现手机号
func (rs *RoleActor) buildPhone(arg *pb.BindPhoneReq, ctx actor.Context) {
	rsp := new(pb.BindPhoneRsp)
	// 验证验证码
	if errcode := rs.findSms(arg.Phone, arg.Code); errcode != pb.OK {
		rsp.Error = errcode
		ctx.Respond(rsp)
		return
	}
	rsp.Phone = arg.Phone
	ctx.Respond(rsp)
	// 移除验证码
	rs.delCode(arg.Phone, arg.Code)
}

// 验证验证码
func (a *RoleActor) findSms(phone, smscode string) pb.ErrCode {
	//TODO 暂时不限制
	//return pb.OK
	if smscode == "8888" && env == "dev" { // 超级验证码,不用验证
		return pb.OK
	}
	ydphone := fmt.Sprintf("91%s", phone)
	mjlPhone := fmt.Sprintf("880%s", phone)
	bjsjphone := fmt.Sprintf("92%s", phone)
	if ydphone != a.smscode[smscode] && mjlPhone != a.smscode[smscode] && bjsjphone != a.smscode[smscode] {
		return pb.SmsCodeWrong
	}
	//验证码过期
	if a.smstime[smscode] <= utils.Timestamp() {
		return pb.SmsCodeExpired
	}
	return pb.OK
}

// 去掉验证码
func (a *RoleActor) delCode(phone, code string) {
	if v, ok := a.smscode[code]; ok {
		delete(a.smsphone, v)
	}
	delete(a.smstime, code)
	delete(a.smscode, code)
	delete(a.smsphone, phone)
}

// 过期检测
func (a *RoleActor) smsExpire() {
	now := utils.Timestamp()
	for k, v := range a.smsphone {
		if a.smstime[v] > now {
			continue
		}
		a.delCode(k, v)
	}
}

// 生成一个验证码,唯一
func (a *RoleActor) genCode() (s string) {
	s = utils.RandStr(4)
	//是否已经存在
	if _, ok := a.smscode[s]; ok {
		return a.genCode() //重复尝试,TODO:一定次数后放弃尝试
	}
	return
}

// 订单发货处理
func (a *RoleActor) tradeHandler(trade *data.TradeRecord) {
	user := a.getUserById(trade.Userid)
	//充值数量
	var diamond, coin, withdrawal int64 = handler.GetGoods(trade)
	//充值消息提醒
	// record, msg1 := handler.BuyNotice(coin, trade.Userid)
	// if record != nil {
	// 	loggerPid.Tell(record)
	// }
	//订单状态更新
	handler.SendGoods(trade, user)
	//充值赠送,TODO 区分日志记录
	// diamond2, coin2, msg2 := handler.FirstPay(trade.First, user)
	// diamond += diamond2
	// coin += coin2
	// //更新有效代理绑定
	// if msg2 != nil {
	// 	rolePid.Tell(msg2)
	// }
	money := trade.Amount
	if trade.RealAmount != 0 {
		money = trade.RealAmount
	}

	// 重复充值问题
	// diamond, coin, withdrawal = RepeatRecharge(user, trade, diamond, coin, withdrawal)
	// 优惠券充值
	if trade.ShopType == data.COUPON && !a.SubCoupon(trade.Userid, trade.ShopId) {
		// 没券了就给实际充值金额
		diamond = int64(trade.RealAmount)
		if diamond == 0 {
			diamond = int64(trade.Amount)
		}
	}
	//在线处理
	if v, ok := a.roles[trade.Userid]; ok {
		handler.SendGoodsOnline(money, user)
		//在线,发送给玩家pid处理
		msg := new(pb.PayGoods)
		msg.Userid = trade.Userid
		msg.Orderid = trade.OrderID
		msg.Money = money
		// msg.Diamond = int64(trade.Score)
		// msg.First = int32(trade.First)
		msg.Withdrawal = withdrawal
		msg.Diamond = diamond
		msg.Coin = coin
		msg.ShopId = trade.ShopId
		msg.ShopType = trade.ShopType
		msg.ShopName = trade.ShopName
		msg.ChannelId = trade.ChannelId
		msg.LastUtr = trade.LastUtr
		v.Pid.Tell(msg)
		// if msg1 != nil {
		// 	v.Pid.Tell(msg1)
		// }
		return
	}
	// 解锁提现
	if user.Money <= 0 {
		user.FirstChargeVal = int32(money)
		user.FirstChargeTime = utils.BsonNow().UnixMilli()
		user.WithdrawLock = handler.CalWithdrawLock(user)
		user.UpdateWithdrawLock()
	}
	// if user.VBBank < coin {
	// 	// vb银行不够了就不送
	// 	coin = 0
	// }

	//离线处理
	handler.SendGoodsOffline(money, user)
	// 活动充值后
	handler.PaySuccess(user, trade.ShopId, trade.ReportId, trade.ShopType, money, true)
	//货币变更及时同步
	//获取ltype
	ltype := handler.GetRechargeLtype(trade.ShopId, trade.ShopType)
	msg3 := handler.ChangeCurrencyMsg(diamond, 0,
		0, 0, coin, withdrawal, ltype, trade.Userid, trade.ShopName, "")

	// VB日志
	if coin != 0 {
		// VB记录日志
		user.VBankLog = append(user.VBankLog, data.VipBankLog{Date: utils.BsonNow().Unix(), LogType: ltype, Amount: coin})

		size := len(user.VBankLog)
		if size > 100 {
			user.VBankLog = user.VBankLog[size-100:]
		}
		user.UpdateVBankLog()
	}

	// 充值完成事件
	a.payEvent(trade, user)

	//上报AD充值
	msg1 := &pb.EventDeposit{
		Userid:   user.Userid,
		BundleId: user.AD_BundleId,
		Amount:   float64(money) / 100,
		First:    user.Money-money <= 0 && utils.Time2DayDate(user.Ctime) == utils.Time2DayDate(time.Now().Local()),
	}
	// reportPid.Tell(msg1)
	mq.NatsPublish(mq.TopicReportDeposit, msg1)

	glog.Infof("report ad deposit, user:%s", user.Userid)

	// fb上报
	// appid中不包含.的都要上报到fb
	// if !strings.Contains(user.AD_AppId, ".") {
	if user.FB_Fbc != "" {
		msg2 := &pb.FBEventDeposit{
			Userid: user.Userid,
			AppId:  user.AD_AppId,
			Amount: float64(money) / 100,
			Agent:  user.AD_User_Agent,
			Ip:     user.RegistIP,
			First:  user.Money-money <= 0 && utils.Time2DayDate(user.Ctime) == utils.Time2DayDate(time.Now().Local()),
		}
		// reportPid.Tell(msg2)
		mq.NatsPublish(mq.TopicReportFBDeposit, msg2)
	}
	// ad := &pb.ADParam{
	// 	AdAdid:       user.AD_ADID,
	// 	AdKey:        user.AD_Key,
	// 	AdS2SCode:    user.AD_S2S_Code,
	// 	AdEventCode1: user.AD_Event_Code1,
	// 	AdEventCode2: user.AD_Event_Code2,
	// 	AdEventCode3: user.AD_Event_Code3,
	// 	AdEventCode4: user.AD_Event_Code4,
	// 	UserAgent:    user.AD_User_Agent,
	// 	BundleId:     user.AD_BundleId,
	// }
	// report := login.ReportDepositMsg(nil, ad, float64(trade.Amount)/float64(100), user.RegistIP, user.Userid)
	// reportPid.Tell(report)

	rolePid.Tell(msg3)
}

func (a *RoleActor) payEvent(trade *data.TradeRecord, user *data.User) {
	diamond, coin, _ := handler.GetGoods(trade)

	// 充值任务事件
	event.Event(user, event.RECHARGE_TASK, &event.RechargeTaskEvent{Amount: trade.Amount})
	user.UpdateTask()

	// 状态检测事件
	event.Event(user, event.CHECK_STATE, &event.CheckStateEvent{})
	user.UpdateState()

	// 充值罐子
	event.Event(user, event.RECHARGE_JAR, &event.RechargeJarEvent{ShopType: int(trade.ShopType), ShopId: trade.ShopId, Give: coin})
	user.UpdateActivedPot()

	// vip
	event.Event(user, event.VIP, &event.VIPEvent{Amount: int64(trade.Amount)})
	// vip 升级奖励检查
	a.vipUpgradeReward(user)
	user.UpdateVIP()

	if user.RegistArea == 1 && (user.State == 1 || user.State == 3) {
		user.BGiveCash += diamond
		user.UpdateBGiveCash()
	}

	// 充值完成事件邮件
	if config.SettingIsOpen(4, data.MAILREPLY) {
		event.Event(user, event.RECHARGE_MAIL, &event.RechargeMailEvent{Amount: trade.Amount})
		user.UpdateFeedBack()
	}
	// 分享
	if user.ShareSuperior != "" {
		share := &pb.ShareRecharge{Money: int32(trade.Amount), Userid: user.Userid, Superior: user.ShareSuperior}
		rolePid.Tell(share)
	}

	user.CommonChannel = trade.ChannelId // 设置为常用渠道
	user.ToggleChannel = 0
	user.UpdateCommonChannel()

	// 日志
	if trade.ShopType == data.SHOP {
		bonus := handler.GetBonusByShop(user.RegistArea, trade.ShopId)
		a.BonusLog(trade.Userid, bonus)
	}

	// 订单回填UTR额外赠送
	if trade.LastUtr != "" {
		msg := &pb.PayUTRBackReword{
			Userid:     user.Userid,
			RegistArea: int32(user.RegistArea),
			Amount:     int64(utils.CaseElse(trade.RealAmount == 0, trade.Amount, trade.RealAmount)),
			OrderId:    trade.OrderID,
			LastUtr:    trade.LastUtr,
		}
		rolePid.Tell(msg)
	}

	if user.ShareSuperior != "" && !user.ShareAgent.ShareSuper {
		glog.Infof("user recharge, share superior: %s, userid: %s, amount: %d", user.ShareSuperior, user.Userid, trade.Amount)
		rolePid.Tell(&pb.ShareAgentPay{
			Userid:  user.Userid,
			SuperId: user.ShareSuperior,
			Amount:  int64(trade.Amount),
			OrderId: trade.OrderID,
		})
	}
}

// vip 升级奖励检查
func (a *RoleActor) vipUpgradeReward(user *data.User) {
	for i := 1; i <= user.Vip.Lv; i++ {
		vip := int32(i)
		vipRecord := table.GetTables().VipTable.Get(vip)
		if vipRecord.Upgrade <= 0 ||
			utils.SliceIn(vip, user.Vip.LevelReward...) { // 领过了
			continue
		}

		msg := handler.ChangeCurrencyMsg(0, 0, 0, 0, int64(vipRecord.Upgrade), 0,
			int32(pb.LOG_TYPE99), user.Userid, fmt.Sprintf("vip%d等级奖励", vipRecord.Id), fmt.Sprint(vipRecord.Id))
		// 在线时
		if v, ok := a.roles[user.Userid]; ok {
			v.Pid.Tell(msg)
		} else {
			// 离线时
			rolePid.Tell(msg)
		}
		user.Vip.LevelReward = append(user.Vip.LevelReward, vip)
	}
}

// 客服回复消息
func (a *RoleActor) ConsumerReplyMessage(ctx actor.Context) {
	arg := ctx.Message().(*pb.ConsumerReplyMessage)
	user := a.getUserById(arg.Userid)
	if user != nil {
		// 消息发送限制置 0
		user.FeedBackTimes = 0
	}
}

// 客服结束对话
func (a *RoleActor) PublishCustomerSessionOver(ctx actor.Context) {
	arg := ctx.Message().(*pb.PublishCustomerSessionOver)
	user := a.getUserById(arg.Userid)
	if user != nil {
		// 消息发送限制置 0
		user.FeedBackTimes = 0
	}
}

// 客服消息
func (a *RoleActor) FeedBackReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.FeedBackReq)
	// 将消息写入db
	user := a.getUserById(arg.UserId)
	if user == nil {
		glog.Errorf("user is nil, user:%s", arg.UserId)
		return
	}
	chat := &data.ChatLog{
		Content: arg.Content,
		Ctime:   utils.LocalTime(),
		Sender:  arg.UserId,
		Name:    user.Nickname,
	}
	log := &data.FeedBackLog{
		Userid: user.Userid,
	}
	log.Get()
	log.Logs = append(log.Logs, *chat)
	log.Utime = utils.BsonNow()
	log.Save()
}

// 客服发送消息
func (a *RoleActor) customerReply(rsp *pb.WebResponse, log *data.ChatLog) {
	user := a.getUserById(log.Receiver)
	if user == nil {
		rsp.ErrMsg = fmt.Sprintf("user is nil, id:%s", log.Receiver)
		return
	}
	uid := bson.NewObjectId().String()
	log.Ctime = utils.BsonNow()
	log.Uid = uid
	if role, ok := a.roles[log.Receiver]; ok {
		// 玩家在线
		msg := new(pb.FeedBackReply)
		bean := &pb.ChatLog{
			Uid:     uid,
			Userid:  log.Sender,
			Title:   log.Title,
			Content: log.Content,
			Ctime:   log.Ctime.Unix(),
		}
		msg.Log = bean
		role.Pid.Tell(msg)
		return
	}
	// 离线
	user.FeedBackTimes = 0
	chat := data.OfflineChatLog{
		Content: log.Content,
		Ctime:   utils.LocalTime().UnixMilli(),
		Sender:  "-1",
		Name:    "system",
	}
	user.OfflineChats = append(user.OfflineChats, chat)
	user.UpdateOfflineLog()
	// user.FeedBackLogMap[uid] = *log
	// user.UpdateFeedBack()
}

// change user data
func (a *RoleActor) modifyUserData(rsp *pb.WebResponse, data *data.ModifyUserData) {
	user := a.getUserById(data.UserId)
	if user == nil {
		rsp.ErrMsg = fmt.Sprintf("user is nil, id:%s", data.UserId)
		return
	}
	if role, ok := a.roles[data.UserId]; ok {
		// 在线
		body, err := json.Marshal(data)
		if err != nil {
			glog.Errorf("serialize fail")
			return
		}
		msg := &pb.ChangeUser{
			Data: body,
		}
		role.Pid.Tell(msg)
		return
	}
	// 离线
	handler.ModifyUserData(user, data)
	user.Save()
}

func (a *RoleActor) pointControl(rsp *pb.WebResponse, data *data.PointControl) {
	user := a.getUserById(data.UserId)
	if user == nil {
		rsp.ErrMsg = fmt.Sprintf("user is nil, id:%s", data.UserId)
		return
	}
	handler.PointControl(user, data)
	user.UpdatePointControl()

	if role, ok := a.roles[data.UserId]; ok {
		// 在线
		msg := &pb.PointControl{
			UserId: data.UserId,
			Switch: data.Switch,
			Factor: data.Factor,
			Score:  data.Score,
		}
		role.Pid.Tell(msg)
		return
	}
	// handler.PointControl(user, data)
	// user.Save()
}

// 改变玩家上级
func (a *RoleActor) changeShareSuperior(msg *pb.ChangeShareSuperior) error {
	if msg.Superior == msg.Userid {
		return errors.New("same userid")
	}
	superior := a.getUserById(msg.Superior)
	user := a.getUserById(msg.Userid)
	if superior == nil {
		return errors.New("no found superior by id")
	}
	if user == nil {
		return errors.New("no found user by id")
	}
	if role, ok := a.roles[msg.Userid]; ok {
		role.Pid.Tell(msg)
	} else {
		if msg.Superior == "" {
			user.ShareSuperior = ""
		} else {
			if user.ShareSuperior != "" {
				// 老的上级要移除当前玩家
				if role, ok := a.roles[user.ShareSuperior]; ok {
					re := &pb.RemoveShareJuntor{Superior: user.ShareSuperior, Userid: msg.Userid}
					role.Pid.Tell(re)
				} else {
					old := a.getUserById(user.ShareSuperior)
					delete(old.ShareBelow, msg.Userid)
					old.UpdateShare()
				}
			}
			user.ShareSuperior = msg.Superior
		}
		user.UpdateShare()
	}
	// 新的上级添加当前用户
	if role, ok := a.roles[msg.Superior]; ok {
		role.Pid.Tell(msg)
	} else {
		superior.ShareBelow[msg.Userid] = data.ShareData{
			UserId:    msg.Userid,
			Name:      msg.Name,
			LastLogin: utils.LocalTime().Unix(),
			Ctime:     utils.LocalTime().Unix(),
		}
		superior.UpdateShare()
	}
	return nil
}

// 跑马灯
func (a *RoleActor) marquee() {
	notice := config.GetNotices2()
	now := time.Now()
	for _, n := range notice {
		if n.Rtype != 0 {
			continue
		}
		if n.Etime.Before(now) {
			// 结束了
			continue
		}
		during := now.Unix() - n.LastSend
		if during >= (int64(n.Num) * 60) {
			// 给前端推送
			n.LastSend = now.Unix()
			ntf := &pb.MarQueeNtf{
				Content: n.Content,
			}
			for k, r := range a.roles {
				if _, ok := a.online[k]; ok {
					ntf.Userid = k
					r.Pid.Tell(ntf)
				}
			}
		}
	}
}

// 转发跑马灯
func (a *RoleActor) MarQueeNtf(ctx actor.Context) {
	for k, r := range a.roles {
		if u, ok := a.online[k]; ok {
			if !u.GetRobot() {
				r.Pid.Tell(ctx.Message())
			}
		}
	}
}

// 转发提现跑马灯
func (a *RoleActor) MarQueeWithdrawNtf(ctx actor.Context) {
	arg := ctx.Message().(*pb.MarQueeWithdrawNtf)
	msg := &pb.MarQueeWithdraw{Ntf: arg}
	for k, r := range a.roles {
		if u, ok := a.online[k]; ok {
			if !u.GetRobot() {
				r.Pid.Tell(msg)
			}
		}
	}
}

// 下级用户注册
func (a *RoleActor) ShareUserRegist(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareUserRegist)
	glog.Debugf("ShareUserRegist %#v", arg)
	user := a.getUserById(arg.Superior)
	if user == nil {
		// 用户不存在
		glog.Errorf("share regist user is nil, %s", arg.Superior)
		return
	}
	if user.ShareBelow == nil {
		user.ShareBelow = make(map[string]data.ShareData)
	}
	if _, ok := user.ShareBelow[arg.Userid]; ok {
		glog.Errorf("share regist is exist, %s", arg.Userid)
		// 已经有了，不加了
		return
	}

	// bean := table.GetTables().ShareConfigTable.Get()
	bean := config.GetShare(1)
	if bean.Id == 0 {
		glog.Errorf("share config is nil, %s", arg.Userid)
		return
	}
	if bean.MaxPeople <= int32(len(user.ShareBelow)) {
		glog.Errorf("share people limit, %s, len:%d", arg.Userid, len(user.ShareBelow))
		return
	}

	data := data.ShareData{
		UserId:    arg.Userid,
		Name:      arg.Name,
		LastLogin: utils.LocalTime().Unix(),
		Ctime:     utils.LocalTime().Unix(),
	}
	user.ShareBelow[arg.Userid] = data
	glog.Errorf("share regist success, superior:%s, user:%s", arg.Superior, arg.Userid)
	// 更新user
	user.UpdateShare()

	// 玩游戏抽奖
	event.Event(user, event.PLAY_SHARE, &event.PlayShareEvent{Ptype: 2, DeviceId: arg.DeviceId})
	user.UpdatePlayShare()
}

// 分享用户注册
func (a *RoleActor) shareRegist(user *data.User) {
	// if env == "dev" {
	// 	user.ShareSuperior = "120102" //test todo
	// }
	if user.ShareSuperior != "" && user.Userid != user.ShareSuperior {
		superior := a.getUserById(user.ShareSuperior)
		if superior != nil && superior.AD_ADID != user.AD_ADID {
			glog.Infof("share user %s", user.ShareSuperior)
			msg := &pb.ShareUserRegist{Userid: user.Userid, Name: user.Nickname, Superior: user.ShareSuperior, DeviceId: user.AD_DeviceId, ShareSource: user.ShareSource}
			if r, ok := a.roles[user.ShareSuperior]; ok {
				r.Pid.Tell(msg)
			} else {
				rolePid.Tell(msg)
			}
			// 分享日志记录
			data.ShareRegistLog(user.Userid, user.ShareSuperior)
			// 分享注册事件发布
			if err = mq.NatsPublish(mq.TopicShareSuperior, &pb.PublishShareRegist{
				Userid:        user.Userid,
				ShareSuperior: user.ShareSuperior,
				Adid:          user.AD_ADID,
				ShareSource:   user.ShareSource,
			}); err != nil {
				glog.Errorf("publish user share regist error: %s, %v", user.Userid, err)
			}

		} else {
			// 上级用户id错误
			glog.Errorf("share Superior id err:%s, adid:%s", user.ShareSuperior, user.AD_ADID)
			user.ShareSuperior = ""
			user.UpdateShare()
		}
	}
}

func (a *RoleActor) GiveAndOutCash(ctx actor.Context) {
	arg := ctx.Message().(*pb.GiveAndOutCash)
	user := a.getUser(arg.Userid)
	if user == nil {
		// 用户不存在
		glog.Errorf("GiveAndOutCash user is nil, %s", arg.Userid)
		return
	}
	// user.GiveDiamond += arg.Give
	user.OutDiamond += arg.Out
	// user.OutDiamond = int64(math.Min(float64(user.OutDiamond), float64(user.Diamond)))
	user.UpdateGiveAndOut()
	// 提现金日志
	if arg.Out != 0 {
		reason := int32(pb.LOG_TYPE45)
		if arg.Reason > 0 {
			reason = arg.Reason
		}
		log := handler.LogOutDiamondMsg(arg.Out, arg.Desc, arg.GameId, user, reason)
		myactor.Logger().Tell(log)
	}
	// if arg.Give != 0 {
	// 	// 赠送金日志
	// 	log := handler.LogVBDiamondMsg(arg.Give, arg.Desc, arg.GameId, user, int32(pb.LOG_TYPE45))
	// 	loggerPid.Tell(log)
	// }
}

func (a *RoleActor) BindLoginPhoneReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.BindLoginPhoneReq)
	user := a.getUser(arg.Userid)
	rsp := new(pb.BindLoginPhoneRsp)
	if user == nil {
		rsp.Error = pb.SmsCodeWrong
		ctx.Respond(rsp)
		return
	}
	if user.ExistsPhone(arg.Phone) {
		rsp.Error = pb.PhoneRegisted
		ctx.Respond(rsp)
		return
	}
	if a.findSms(arg.Phone, arg.Code) == pb.OK {
		user.Phone = arg.Phone
		user.UpdatePhone()
		ctx.Respond(rsp)
		a.delCode(arg.Phone, arg.Code)
		return
	}
	rsp.Error = pb.SmsCodeWrong
	ctx.Respond(rsp)
}

func (a *RoleActor) UpdateRegisReward(ctx actor.Context) {
	arg := ctx.Message().(*pb.UpdateRegisReward)
	user := a.getUserById(arg.Userid)
	if user == nil {
		return
	}
	user.RegistReward = arg.State
	user.UpdateRegistReward()
}

// 分享结算
func (a *RoleActor) shareSettlement() {
	list := make([]data.ShareBetAmount, 0)
	// shareBean := table.GetTables().ShareConfigTable.Get()
	shareBean := config.GetShare(1)
	if shareBean.Id == 0 {
		return
	}
	data.ListByQ(data.ShareAmounts, nil, &list)
	data.DeleteAll(data.ShareAmounts, nil)
	for _, sba := range list {
		score := sba.Score * int64(shareBean.FirstReward) / 10000 // 结算分数
		msg := &pb.ShareSettlement{
			Score:     score,
			Total:     sba.Score,
			Userid:    sba.Userid,
			Level:     1,
			Superior:  sba.Superior,
			Name:      sba.Name,
			LastLogin: sba.Ctime,
		}
		rolePid.Tell(msg)
	}
}

// 发送验证码
// func (a *RoleActor) sendSms(phone, code string, n int) {
// 	// if cfg.Section("smsbk").Key("status").MustBool(false) {
// 	// 	go login.SendSmsBK(smsbkURL, smsbkappid, arg.Phone, code, smsbkaccount, smsbkpassword)
// 	// } else
// 	success := false
// 	switch n {
// 	case 0:
// 		if cfg.Section("smskmi").Key("status").MustBool(false) {
// 			success = login.SendSmsKMI(config.SmskmiURL, config.SmskmiAccKey, config.SmskmiSecKey, phone, code)
// 		}
// 	case 1:
// 		if cfg.Section("smsmt").Key("status").MustBool(false) {
// 			success = login.SendSmsMT(config.SmsmtURL, config.SmsmtAppKey, config.SmsmtAppSecret, config.SmsmtAppCode, phone, code)
// 		}
// 	case 2:
// 		if cfg.Section("smsxxy").Key("status").MustBool(false) {
// 			success = login.SendSmsXXY(config.SmsxxyURL, config.SmsxxyAppKey, config.SmsxxyAppSecret, config.SmsxxyAppCode, phone, code)
// 		}
// 	case 3:
// 		if cfg.Section("smscl").Key("status").MustBool(false) {
// 			success = login.SendSmsCL(config.SmsclUrl, config.SmsclApiKey, config.SmsclPassword, phone, code)
// 		}
// 	default:
// 		a.delCode(phone, code)
// 		glog.Errorf("send smscode fail,phone:%s,code:%s", phone, code)
// 		return
// 	}
// 	if success {
// 		return
// 	}
// 	a.sendSms(phone, code, n+1)
// }

func (a *RoleActor) clearUserBankInfo(userid string) {
	if v, ok := a.roles[userid]; ok {
		v.Pid.Tell(new(pb.ClearBankInfo))
	} else {
		user := a.getUserById(userid)
		if user == nil {
			glog.Errorf("no found user by id:%s", userid)
			return
		}
		user.Bank = ""
		user.IFSC = ""
		user.BankAccounts = ""
		user.UpdateBank()
	}
}

func (a *RoleActor) giveWithdraw(msg *data.GiveWithdrawCash) {
	if role, ok := a.roles[msg.Userid]; ok {
		m := &pb.GiveWithdrawCash{
			OutCash: msg.Cash,
			Desc:    "后台赠送可提现金",
		}
		role.Pid.Tell(m)
		return
	}
	// 离线了
	user := a.getUserById(msg.Userid)
	if user == nil {
		return
	}

	var diamond int64 = 0
	if user.OutDiamond+msg.Cash > user.Diamond {
		diamond = user.OutDiamond + msg.Cash - user.Diamond
	}
	a.syncCurrency(diamond, 0, 0, msg.Cash, 0, int32(pb.LOG_TYPE83), msg.Userid, "后台赠送可提现金", "", false)
}

func (a *RoleActor) userSmsRecord(phone, code string) {
	msg := &pb.LogUserSmsRecord{
		Phone: phone,
		Code:  code,
	}
	myactor.Logger().Tell(msg)
}

// 网页上传端adid
func (a *RoleActor) WebPageAdid(ctx actor.Context) {
	arg := ctx.Message().(*pb.WebPageAdid)
	rsp := new(pb.WebPageAdided)
	if p, ok := a.roles[arg.Userid]; ok {
		req := &pb.AdjustAdidReq{Adid: arg.Adid}
		p.Pid.Tell(req)
		rsp.Code = 200
	} else {
		rsp.Code = -1
	}
	ctx.Respond(rsp)
}

// cdk兑换
func (a *RoleActor) CDKReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.CDKReq)
	glog.Debugf("CDKReq %#v", arg)
	rsp := new(pb.CDKRsp)

	c := data.GetCdkConfig(arg.Code)
	if c.Key == "" {
		rsp.Error = pb.InvalidCdk
		ctx.Respond(rsp)
		return
	}

	if c.UserId != "" && c.UserId != arg.Userid {
		rsp.Error = pb.InvalidCdk
		ctx.Respond(rsp)
		return
	}

	rsp.Cash = c.Diamond
	rsp.Give = c.GiveDiamond
	rsp.Out = c.OutDiamond
	rsp.Bouns = c.Bouns
	ctx.Respond(rsp)

	c.UpdateReceivePeople() // 增加领取人数

	if c.Bouns > 0 {
		// 加罐子
		user := a.getUserById(arg.Userid)
		if user != nil {
			event.Event(user, event.Normal_JAR, &event.NormalJarEvent{Give: c.Bouns, Multiple: 40})
			user.UpdateActivedPot()
			// 日志
			a.BonusLog(arg.Userid, c.Bouns)
		}
	}
}

func (a *RoleActor) BonusLog(userid string, bonus int64) {
	if bonus == 0 {
		return
	}

	log := &pb.BonusLog{
		Userid: userid,
		Bonus:  bonus,
	}

	myactor.Logger().Tell(log)
}

func (a *RoleActor) PlayshareTest(ctx actor.Context) {
	arg := ctx.Message().(*pb.PlayshareTest)
	glog.Debugf("PlayshareTest %#v", arg)
	if role, ok := a.roles[arg.Userid]; ok {
		role.Pid.Tell(arg)
	}
	ctx.Respond("")
}

func (a *RoleActor) UploadPhoto(ctx actor.Context) {
	arg := ctx.Message().(*pb.UploadPhotoLog)
	glog.Debugf("UploadPhotoLog %#v", arg)
	if role, ok := a.roles[arg.Userid]; ok {
		role.Pid.Tell(&pb.PhotoeAuditStatusNtf{Status: pb.AUDIT, Url: arg.Url})
	}
}

func (a *RoleActor) userCustomPhoto(msg *data.UserCustomPhoto) {
	log := data.GetPhotoLog(msg.Id)
	if log.UserId == "" {
		return
	}
	m := &pb.UserCustomPhoto{
		Url:    log.Url,
		Status: int32(msg.Result),
	}
	if r, ok := a.roles[log.UserId]; ok {
		r.Pid.Tell(m)
		return
	}

	user := a.getUserById(log.UserId)
	if user == nil {
		return
	}
	if msg.Result == 1 {
		user.Photo = log.Url
		user.UpdatePhoto()
		return
	}
	myredis.Redis().Del(context.Background(), data.GetCustomPhotoKey(log.UserId))
}

// 获取剧情模式最低使用库存
func (a *RoleActor) getTpStoryStockMin(ctx actor.Context) ([]byte, error) {
	resp := make(map[string]int32)
	amendRoom := table.GetTables().Tp2AmendRoomTable
	for _, room := range amendRoom.GetDataList() {
		resp[room.Id] = room.StoryStockMin
	}
	result, err2 := json.Marshal(resp)
	if err2 != nil {
		glog.Errorf("msg err: %v", err2)
		return nil, fmt.Errorf("msg err: %v", err2)
	}
	return result, nil
}

func (a *RoleActor) ChangeViceAccount(ctx actor.Context) {
	arg := ctx.Message().(*pb.ChangeViceAccount)
	// 不在线的下次登录时更新
	user := a.getUser(arg.Userid)
	if user != nil && !user.ViceAccount {
		// 标记为小号
		user.ViceAccount = true
		// 同步到网关
		a.send2userid(arg.Userid, arg)
		glog.Infof("update user vice account: %s", arg.Userid)
	}
}

// 新代理注册消息 ShareAgentRegist
func (a *RoleActor) ShareAgentRegist(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareAgentRegist)
	if arg.SuperId == "" {
		return
	}
	var superLv1, superLv2, superLv3 *data.User
	superLv1 = a.getUserById(arg.SuperId)
	if superLv1 != nil {
		a.shareAgentAdd(superLv1)
	} else {
		return
	}
	if superLv1 == nil || superLv1.ShareSuperior == "" {
		return
	}

	superLv2 = a.getUserById(superLv1.ShareSuperior)
	if superLv2 != nil {
		a.shareAgentAdd(superLv2)
	}
	if superLv2 == nil || superLv2.ShareSuperior == "" {
		return
	}

	superLv3 = a.getUserById(superLv2.ShareSuperior)
	if superLv3 != nil {
		a.shareAgentAdd(superLv3)
	}
}

// 新代理注册
func (a *RoleActor) shareAgentAdd(super *data.User) {
	super.ShareAgent.TeamAgents++
	nextTeamLv := super.ShareAgent.TeamLv + 1
	nextTeam := table.GetTables().ShareGroupTable.Get(nextTeamLv)
	if nextTeam != nil {
		if super.ShareAgent.TeamAgents >= nextTeam.NeedShareUsers &&
			super.ShareAgent.TeamBets >= int64(nextTeam.NeedShareBets) {
			super.ShareAgent.TeamLv = nextTeam.Level
		}
	}

	a.shareAgentSync(super)
}

// 新代理打码消息
func (a *RoleActor) ShareAgentBets(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareAgentBets)
	if arg.SuperId == "" {
		return
	}
	var superLv1, superLv2, superLv3 *data.User
	superLv1 = a.getUserById(arg.SuperId)
	if superLv1 != nil {
		a.shareAgentBets(arg, superLv1, 1)
	} else {
		return
	}
	if superLv1 == nil || superLv1.ShareSuperior == "" {
		return
	}

	superLv2 = a.getUserById(superLv1.ShareSuperior)
	if superLv2 != nil {
		a.shareAgentBets(arg, superLv2, 2)
	}
	if superLv2 == nil || superLv2.ShareSuperior == "" {
		return
	}

	superLv3 = a.getUserById(superLv2.ShareSuperior)
	if superLv3 != nil {
		a.shareAgentBets(arg, superLv3, 3)
	}
}

// 代理用户打码
func (a *RoleActor) shareAgentBets(arg *pb.ShareAgentBets, super *data.User, userLv int32) {
	super.ShareAgent.TeamBets += arg.Bets

	// 打码返佣
	team := table.GetTables().ShareGroupTable.Get(super.ShareAgent.TeamLv)
	if team != nil {
		shareRate, ok := utils.CaseWhen3(userLv, 1, team.ShareL1Rate, 2, team.ShareL2Rate, 3, team.ShareL3Rate)
		if ok {
			earningsBet := int64(float64(arg.Bets*int64(shareRate)*100) / 10000) // 厘
			if earningsBet > 0 {
				shareBetsPrizeType := table.GetTables().ShareTable.Get().ShareBetsPrizeType
				super.ShareAgent.EarningsBetPendding += earningsBet
				// 打码返佣日志
				now := time.Now().In(location)
				incomeRecord := &data.ShareAgentIncomeRecord{
					Id:         bson.NewObjectId().Hex(),
					Userid:     arg.Userid,
					SuperId:    super.Userid,
					Lv:         userLv,
					TeamLv:     super.ShareAgent.TeamLv,
					Itype:      1,
					Amount:     earningsBet,
					AmountType: shareBetsPrizeType,
					WaterId:    arg.WaterId,
					Gtype:      arg.Gtype,
					Bets:       arg.Bets,
					Idate:      now.Format(utils.FORMAT_DATE),
					Ctime:      now.UnixMilli(),
				}
				incomeRecord.Save()
			}
		}
	}

	// 团队升级
	nextTeamLv := super.ShareAgent.TeamLv + 1
	nextTeam := table.GetTables().ShareGroupTable.Get(nextTeamLv)
	if nextTeam != nil {
		if super.ShareAgent.TeamAgents >= nextTeam.NeedShareUsers &&
			super.ShareAgent.TeamBets >= int64(nextTeam.NeedShareBets) {
			super.ShareAgent.TeamLv = nextTeam.Level
		}
	}

	a.shareAgentSync(super)
}

// 新代理充值消息
func (a *RoleActor) ShareAgentPay(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareAgentPay)
	if arg.SuperId == "" {
		return
	}
	user := a.getUserById(arg.Userid)
	super := a.getUserById(arg.SuperId)
	if super == nil || user == nil || user.ShareSuperior == "" || user.ShareAgent.ShareSuper {
		return
	}

	share1PayMin := table.GetTables().ShareTable.Get().Share1PayMin
	if user.Money < uint32(share1PayMin) {
		return
	}
	// 标记用户已算作有效代理
	user.ShareAgent.ShareSuper = true
	a.shareAgentSync(user)

	// 计算累计人头奖
	super.ShareAgent.ShareUsers++
	shareTotalPrizeType := table.GetTables().ShareTable.Get().ShareTotalPrizeType
	for _, item := range table.GetTables().ShareUsersPrizeTable.GetDataList() {
		if super.ShareAgent.ShareUsers >= item.ShareUsers &&
			!utils.SliceIn(item.Id, super.ShareAgent.ShareUsersPrizes...) {
			super.ShareAgent.ShareUsersPrizes = append(super.ShareAgent.ShareUsersPrizes, item.Id)
			super.ShareAgent.EarningsUserPendding += int64(item.ShareUsersPrize)
			now := time.Now().In(location)
			record := &data.ShareAgentIncomeRecord{
				Id:         bson.NewObjectId().Hex(),
				Userid:     arg.Userid,
				SuperId:    arg.SuperId,
				Lv:         1,
				TeamLv:     super.ShareAgent.TeamLv,
				Itype:      3,
				Amount:     int64(item.ShareUsersPrize) * 100, // 单位厘
				AmountType: shareTotalPrizeType,
				WaterId:    arg.OrderId,
				Idate:      now.Format(utils.FORMAT_DATE),
				Ctime:      now.UnixMilli(),
			}
			record.Save()
		}
	}

	// 单个人头奖
	share1PrizeLimit := table.GetTables().ShareTable.Get().Share1PrizeLimit
	if super.ShareAgent.TodayEarningsUser < share1PrizeLimit {
		super.ShareAgent.TodayEarningsUser++
		// 邀请者奖金
		share1SuperPrize := table.GetTables().ShareEachUserPrizeTable.Get().Share1SuperPrize
		share1SuperPrizeType := table.GetTables().ShareTable.Get().Share1SuperPrizeType
		super.ShareAgent.EarningsUserPendding += int64(share1SuperPrize)
		now := time.Now().In(location)
		record := &data.ShareAgentIncomeRecord{
			Id:         bson.NewObjectId().Hex(),
			Userid:     arg.Userid,
			SuperId:    arg.SuperId,
			Lv:         1,
			TeamLv:     super.ShareAgent.TeamLv,
			Itype:      2,
			Amount:     int64(share1SuperPrize) * 100, // 单位厘
			AmountType: share1SuperPrizeType,
			WaterId:    arg.OrderId,
			Idate:      now.Format(utils.FORMAT_DATE),
			Ctime:      now.UnixMilli(),
		}
		record.Save()

		// 受邀者奖金
		share1SelfPrize := int64(table.GetTables().ShareEachUserPrizeTable.Get().Share1SelfPrize)
		share1SelfPrizeType := table.GetTables().ShareTable.Get().Share1SelfPrizeType
		selfRecord := &data.ShareAgentIncomeRecord{
			Id:         bson.NewObjectId().Hex(),
			Userid:     arg.Userid,
			SuperId:    arg.SuperId,
			Lv:         1,
			TeamLv:     super.ShareAgent.TeamLv,
			Itype:      4,
			Amount:     share1SelfPrize * 100, // 单位厘
			AmountType: share1SelfPrizeType,
			WaterId:    arg.OrderId,
			Idate:      now.Format(utils.FORMAT_DATE),
			Ctime:      now.UnixMilli(),
		}
		selfRecord.Save()

		// 受邀者奖金直接发放 1代表bonus,2代表cash,3代表withdrawable
		var bonus, diamond, outDiamond int64
		switch share1SelfPrizeType {
		case 1:
			bonus = share1SelfPrize
		case 2:
			diamond = share1SelfPrize
		case 3:
			diamond = share1SelfPrize
			outDiamond = share1SelfPrize
		}
		msg := handler.ChangeCurrencyMsg(diamond, 0, 0, 0, bonus, outDiamond,
			int32(pb.LOG_TYPE140), arg.Userid, "有效人头受邀者奖金", arg.OrderId)
		// 在线时
		if v, ok := a.roles[arg.Userid]; ok {
			v.Pid.Tell(msg)
		} else {
			// 离线时
			rolePid.Tell(msg)
		}
	}

	a.shareAgentSync(super)
}

// 代理信息同步gate
func (a *RoleActor) shareAgentSync(user *data.User) {
	if role, online := a.roles[user.Userid]; online {
		role.Pid.Tell(&pb.ShareAgentSync{
			TeamLv:                user.ShareAgent.TeamLv,
			TeamAgents:            user.ShareAgent.TeamAgents,
			TeamBets:              user.ShareAgent.TeamBets,
			ShareUsers:            user.ShareAgent.ShareUsers,
			ShareUsersPrizes:      user.ShareAgent.ShareUsersPrizes,
			ShareSuper:            user.ShareAgent.ShareSuper,
			EarningsBetTotal:      user.ShareAgent.EarningsBetTotal,
			EarningsBetUnclaimed:  user.ShareAgent.EarningsBetUnclaimed,
			EarningsBetPendding:   user.ShareAgent.EarningsBetPendding,
			EarningsUserTotal:     user.ShareAgent.EarningsUserTotal,
			EarningsUserUnclaimed: user.ShareAgent.EarningsUserUnclaimed,
			EarningsUserPendding:  user.ShareAgent.EarningsUserPendding,
			TodayEarningsUser:     user.ShareAgent.TodayEarningsUser,
			TodayTakeItypeDates:   user.ShareAgent.TodayTakeItypeDates,
		})
	}
}

// 领取所有打码返佣金额(earningsUnclaimed)
func (a *RoleActor) ShareOngoingEarningUnclaimedTackReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareOngoingEarningUnclaimedTackReq)
	rsp := new(pb.ShareOngoingEarningUnclaimedTackRsp)

	defer ctx.Respond(rsp)

	user := a.getUser(arg.Userid)
	if user == nil {
		rsp.Error = pb.Failed
		return
	}

	// 没有未领取金额
	if user.ShareAgent.EarningsBetUnclaimed <= 0 {
		rsp.Error = pb.Failed
		return
	}

	tacks, err := a.shareAgentIncomeTack(user, []int32{1}, "")
	if err != nil {
		rsp.Error = pb.Failed
		return
	}

	glog.Infof("tack all ongoing earnings: userid=%s, tacks=%d, unclaimed=%d", arg.Userid, tacks, user.ShareAgent.EarningsBetUnclaimed)
	user.ShareAgent.EarningsBetTotal += tacks
	user.ShareAgent.EarningsBetUnclaimed = 0
	a.shareAgentSync(user)

	rsp.EarningsTotal = user.ShareAgent.EarningsBetTotal
	rsp.EarningsUnclaimed = user.ShareAgent.EarningsBetUnclaimed
}

// 领取所有人头和任务返佣金额(earningsUnclaimed)
func (a *RoleActor) ShareInstantEarningUnclaimedTackReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareInstantEarningUnclaimedTackReq)
	rsp := new(pb.ShareInstantEarningUnclaimedTackRsp)

	defer ctx.Respond(rsp)

	user := a.getUser(arg.Userid)
	if user == nil {
		rsp.Error = pb.Failed
		return
	}
	// 没有未领取金额
	if user.ShareAgent.EarningsUserUnclaimed <= 0 {
		rsp.Error = pb.Failed
		return
	}

	tacks, err := a.shareAgentIncomeTack(user, []int32{2, 3}, "")
	if err != nil {
		rsp.Error = pb.Failed
		return
	}

	glog.Infof("tack all instant earnings: userid=%s, tacks=%d, unclaimed=%d", arg.Userid, tacks, user.ShareAgent.EarningsUserUnclaimed)
	user.ShareAgent.EarningsUserTotal += tacks
	user.ShareAgent.EarningsUserUnclaimed = 0
	a.shareAgentSync(user)

	rsp.EarningsTotal = user.ShareAgent.EarningsUserTotal
	rsp.EarningsUnclaimed = user.ShareAgent.EarningsUserUnclaimed
}

// 代理收入账单记录 打码返佣/人头奖励 unclaimed 奖励领取
func (a *RoleActor) ShareAgentIncomeTackReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareAgentIncomeTackReq)
	rsp := new(pb.ShareAgentIncomeTackRsp)

	defer ctx.Respond(rsp)

	if arg.Date == "" {
		rsp.Error = pb.Failed
		return
	}

	user := a.getUser(arg.Userid)
	if user == nil {
		rsp.Error = pb.Failed
		return
	}

	var itype []int32
	switch arg.Itype {
	case 1:
		itype = []int32{1}
	case 2:
		itype = []int32{2, 3}
	default:
		rsp.Error = pb.Failed
		return
	}

	tacks, err := a.shareAgentIncomeTack(user, itype, arg.Date)
	if err != nil {
		rsp.Error = pb.Failed
		return
	}

	var unclaimed int64
	if arg.Itype == 1 {
		unclaimed = user.ShareAgent.EarningsBetUnclaimed
		user.ShareAgent.EarningsBetTotal += tacks
		user.ShareAgent.EarningsBetUnclaimed -= tacks
		if user.ShareAgent.EarningsBetUnclaimed < 0 {
			user.ShareAgent.EarningsBetUnclaimed = 0
		}
	} else {
		unclaimed = user.ShareAgent.EarningsUserUnclaimed
		user.ShareAgent.EarningsUserTotal += tacks
		user.ShareAgent.EarningsUserUnclaimed -= tacks
		if user.ShareAgent.EarningsUserUnclaimed < 0 {
			user.ShareAgent.EarningsUserUnclaimed = 0
		}
	}
	a.shareAgentSync(user)
	glog.Infof("tack share income earnings: userid=%s, date=%s, itype=%d, tacks=%d, unclaimed=%d", arg.Userid, arg.Date, arg.Itype, tacks, unclaimed)
}

func (a *RoleActor) shareAgentIncomeTack(user *data.User, itype []int32, idate string) (tacks int64, err error) {
	userid := user.Userid
	now := time.Now().In(location)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	args := []any{userid, itype, todayStart.UnixMilli(), userid, itype}

	var filter string
	if idate != "" {
		filter = " AND t0.idate = ?"
		args = append(args, idate)
	}

	// 查询所有未领奖记录
	sql := `
		SELECT t0.idate idate, t0.itype itype, t0.amount_type amount_type, t0.amounts, (CASE WHEN t1.ctime > 0 THEN 1 ELSE 0 END) tacked
		FROM (
			SELECT idate, itype, amount_type, MAX(ctime) ctime 
			FROM game.col_activity_share_income_record_tack_log FINAL
			WHERE super_id = ? AND itype IN ?
			GROUP BY idate, itype, amount_type
		) t1 RIGHT JOIN (
			SELECT idate, itype, amount_type, SUM(amount) amounts
			FROM game.col_activity_share_income_record FINAL 
			WHERE ctime < ? AND super_id = ? AND itype IN ?
			GROUP BY idate, itype, amount_type
		) t0 ON t0.idate = t1.idate AND t0.itype = t1.itype
		WHERE tacked = 0 %s
		ORDER BY t0.idate DESC
	`
	var results []map[string]any
	if err = ck.Select(&results, fmt.Sprintf(sql, filter), args...); err != nil {
		return
	}

	// 1.打码奖励,2人数人头奖励,3.累计任务人头奖励,4.受邀者人头奖励(直接发放不用在这领取)
	var ltypePrizes = map[int32][]int64{
		1: []int64{0, 0, 0}, // 0:bonus,1:cash,2:withdrawable
		2: []int64{0, 0, 0},
		3: []int64{0, 0, 0},
	}
	// var bonus, diamond, outDiamond int64
	for _, r := range results {
		idate := r["idate"].(string)
		itype := int32(utils.ToInt64(r["itype"]))
		amount_type := utils.ToInt64(r["amount_type"])
		amounts := utils.ToInt64(r["amounts"]) / 100 // 厘
		if amounts == 0 {
			continue
		}

		// 今日已领取过 itype, idate
		takeKey := fmt.Sprintf("%d-%s", itype, idate)
		if utils.SliceIn(takeKey, user.ShareAgent.TodayTakeItypeDates...) {
			glog.Errorf("share agent prize already taked: %s, %d, %s, amounts=%v", userid, itype, idate, amounts)
			continue
		}
		// 记录今日已领取
		user.ShareAgent.TodayTakeItypeDates = append(user.ShareAgent.TodayTakeItypeDates, takeKey)

		log := &data.ShareAgentIncomeRecordTackLog{
			Id:         bson.NewObjectId().Hex(),
			SuperId:    userid,
			Amount:     amounts,
			AmountType: int32(amount_type),
			Idate:      idate,
			Itype:      itype,
			Ctime:      now.UnixMilli(),
		}
		log.Save()

		tacks += amounts
		// 1代表bonus,2代表cash,3代表withdrawable
		switch amount_type {
		case 1:
			ltypePrizes[itype][0] += amounts
		case 2:
			ltypePrizes[itype][1] += amounts
		case 3:
			ltypePrizes[itype][1] += amounts
			ltypePrizes[itype][2] += amounts
		}
	}

	if tacks > 0 {
		for itype, amounts := range ltypePrizes {
			ltype, _ := utils.CaseWhen3(itype, 1, int32(pb.LOG_TYPE142), 2, int32(pb.LOG_TYPE141), 3, int32(pb.LOG_TYPE144))
			ltypeName, _ := utils.CaseWhen3(itype, 1, "打码", 2, "人头", 3, "任务")
			all := utils.CaseElse(idate == "", "所有", "")
			desc := fmt.Sprintf("领取%s代理%s返佣金额", all, ltypeName)
			msg := handler.ChangeCurrencyMsg(amounts[1], 0, 0, 0, amounts[0], amounts[2],
				ltype, userid, desc, idate)
			// 在线时
			if v, ok := a.roles[userid]; ok {
				v.Pid.Tell(msg)
			} else {
				// 离线时
				rolePid.Tell(msg)
			}
		}
	}
	return
}

// 代理零点重置
func (a *RoleActor) ShareAgentZeroReset(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareAgentZeroReset)
	user := a.getUser(arg.Userid)
	if user == nil {
		return
	}
	user.ShareAgent.EarningsBetUnclaimed += user.ShareAgent.EarningsBetPendding
	user.ShareAgent.EarningsBetPendding = 0
	user.ShareAgent.EarningsUserUnclaimed += user.ShareAgent.EarningsUserPendding
	user.ShareAgent.EarningsUserPendding = 0
	user.ShareAgent.TodayEarningsUser = 0
	user.ShareAgent.TodayTakeItypeDates = []string{}
	a.shareAgentSync(user)
}

// 礼包码领取
func (rs *RoleActor) GiftPackCodeTackReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.GiftPackCodeTackReq)
	rsp := new(pb.GiftPackCodeTackRsp)
	defer ctx.Respond(rsp)

	user := rs.getUser(arg.Userid)
	if user == nil {
		rsp.Error = pb.Failed
		return
	}

	gift := &data.GiftPackCode{Code: arg.Code}
	gift.Get()

	// 礼包码不存在
	if gift.Score <= 0 {
		rsp.Error = pb.GiftPackCodeInvalid
		return
	}

	// 已结束
	nowSec := time.Now().Unix()
	if gift.Status != 1 || gift.FinishTime > 0 ||
		nowSec < gift.TimeOpen || nowSec > gift.TimeClose {
		rsp.Error = pb.GiftPackCodeFinish
		return
	}

	// 已领取过
	if utils.SliceIn(arg.Userid, gift.Userids...) {
		rsp.Error = pb.GiftPackCodeTacked
		return
	}

	// 是否满足领取用户类型
	chargeType := handler.GetChargeType(user) + 10
	var matchUtypes []int32
	if utils.SliceIn(int32(chargeType), gift.Utypes...) {
		matchUtypes = append(matchUtypes, int32(chargeType))
	}
	if utils.SliceIn(int32(user.RegistArea), gift.Utypes...) {
		matchUtypes = append(matchUtypes, int32(user.RegistArea))
	}
	if len(matchUtypes) == 0 {
		rsp.Error = pb.GiftPackCodeUnqualified
		return
	}

	// 计算领取金额
	var scoreTack int64
	var finished bool
	// 1.随机码 2.等额码
	if gift.GiftType == 1 {
		pieceMax := (gift.Score - gift.ScoreReceived) - int64(gift.Pieces-gift.PiecesReceived) // 剩余金额-剩余人数，最少每个人能分到1分
		pieceMax = utils.Min(pieceMax, gift.PieceMax)
		pieceMin := utils.Min(pieceMax, gift.PieceMin)
		if pieceMin >= pieceMax {
			scoreTack = pieceMin
		} else {
			scoreTack = int64(utils.RandMN(int(pieceMin), int(pieceMax)))
		}
	} else {
		scoreTack = gift.Score / int64(gift.Pieces)
	}

	// 最后一份领取剩下所有的
	if gift.PiecesReceived+1 == gift.Pieces {
		scoreTack = gift.Score - gift.ScoreReceived
	}

	if scoreTack+gift.ScoreReceived > gift.Score {
		scoreTack = gift.Score - gift.ScoreReceived
	}

	// 是否领完
	finished = scoreTack+gift.ScoreReceived >= gift.Score || gift.PiecesReceived+1 >= gift.Pieces
	if finished {
		gift.FinishTime = nowSec
	}
	gift.ScoreReceived += scoreTack
	gift.PiecesReceived++
	gift.Userids = append(gift.Userids, arg.Userid)
	gift.Save()

	log := &data.GiftPackCodeTackLog{
		Id:          bson.NewObjectId().Hex(),
		Code:        gift.Code,
		Userid:      arg.Userid,
		GiftType:    gift.GiftType,
		ScoreTack:   scoreTack,
		Score:       gift.Score,
		ScoreType:   gift.ScoreType,
		ScoreBefore: gift.Score - gift.ScoreReceived + scoreTack,
		ScoreAfter:  gift.Score - gift.ScoreReceived,
		Utypes:      matchUtypes,
		Ctime:       nowSec,
	}
	log.Save()

	rsp.ScoreTack = scoreTack
	rsp.ScoreType = gift.GiftType
	rsp.GiftType = gift.GiftType

	if scoreTack > 0 {
		// 发送金额 1代表bonus,2代表cash,3代表withdrawable
		var bonus, diamond, outDiamond int64
		switch gift.ScoreType {
		case 1:
			bonus = scoreTack
		case 2:
			diamond = scoreTack
		case 3:
			diamond = scoreTack
			outDiamond = scoreTack
		}

		msg := handler.ChangeCurrencyMsg(diamond, 0, 0, 0, bonus, outDiamond,
			int32(pb.LOG_TYPE143), arg.Userid, "礼包码领取", gift.Code)
		// 在线时
		if v, ok := rs.roles[arg.Userid]; ok {
			v.Pid.Tell(msg)
		} else {
			// 离线时
			rolePid.Tell(msg)
		}
	}
}

// 客服聊天玩家上传文件
func (a *RoleActor) CustomerUploadFile(ctx actor.Context) {
	arg := ctx.Message().(*pb.CustomerUploadFile)
	glog.Debugf("CustomerUploadFile %#v", arg)

	if role, ok := a.roles[arg.Userid]; ok {
		role.Pid.Tell(arg)
		return
	}

	// 离线处理
	if arg.Error != "" {
		glog.Error("upload file error: ", arg.Error)
		return
	}

	user := a.getUserById(arg.Userid)
	user.FeedBackTimes++

	req := &pb.CustomerSendReq{
		Userid:   arg.Userid,
		Ctype:    arg.Ctype,
		Content:  arg.Url,
		Filename: arg.FileName,
		Username: user.Nickname,
		Filesize: arg.FileSize,
	}
	if err := mq.NatsPublish(mq.RequestCustomerSend, req); err != nil {
		glog.Errorf("CustomerUploadFile msg publish error: %v", err)
	}
}
