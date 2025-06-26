package gate

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 在别的节点登录
func (a *GateActor) loginElse(ctx actor.Context) {
	arg := ctx.Message().(*pb.LoginElse)
	glog.Debugf("LoginElse %#v", arg)
	userid := arg.GetUserid()
	if p, ok := a.online[userid]; ok {
		p.Pid.Tell(arg)
		//直接关闭
		p.Pid.Tell(&pb.OfflineStop{Ltype: pb.LOGOUT_TYPE4})
	} else if p, ok := a.offline[userid]; ok {
		p.Pid.Tell(arg)
		//直接关闭
		p.Pid.Tell(&pb.OfflineStop{Ltype: pb.LOGOUT_TYPE4})
	}
	glog.Debugf("LoginElse userid: %s", userid)
	//移除
	delete(a.online, userid)
	delete(a.offline, userid)
	delete(a.offtime, userid)
}

func (a *GateActor) LoginedElse(ctx actor.Context) {
	//在别的节点登录
	arg := ctx.Message().(*pb.LoginedElse)
	glog.Debugf("LoginedElse %#v", arg)
}

func (a *GateActor) SetLogin(ctx actor.Context) {
	arg := ctx.Message().(*pb.SetLogin)
	glog.Debugf("SetLogin %#v", arg)
	rsp := &pb.SetLogined{
		RolePid: a.rolePid,
	}
	glog.Infof("SetLogin %s", arg.Sender.String())
	ctx.Respond(rsp)
}

// 登出成功
func (a *GateActor) Logout(ctx actor.Context) {
	arg := ctx.Message().(*pb.Logout)
	glog.Debugf("Logout %#v", arg)
	userid := arg.GetUserid()
	glog.Debugf("Logout userid: %s", userid)
	//离线 arg.Sender == a.online[userid]
	a.offline[userid] = &data.LoginRole{Pid: arg.Sender, Token: arg.Token}
	a.offtime[userid] = 10 //缓存5分钟
	//移除
	delete(a.online, userid)
	//不能离开,因为缓存了数据,离开会出现数据不同步
	//a.rolePid.Tell(arg)
	//a.roomPid.Tell(arg)
}

// 下线离线玩家
func (a *GateActor) offlineStop(ctx actor.Context) {
	for k, v := range a.offline {
		if a.offtime[k] <= 0 {
			v.Pid.Tell(&pb.OfflineStop{Ltype: pb.LOGOUT_TYPE1})
			delete(a.offline, k)
			delete(a.offtime, k)
			//正式下线消息
			arg := new(pb.Logout)
			arg.Userid = k
			arg.Sender = v.Pid
			a.rolePid.Tell(arg)
			a.roomPid.Tell(arg)
			continue
		}
		a.offtime[k]--
	}
}

////断开其它节点连接
//func (a *GateActor) logoutOther(userid string, ctx actor.Context) {
//	//TODO 数据一致性,防止数据覆盖
//	msg1 := new(pb.LoginHall)
//	msg1.Userid = userid
//	msg1.NodeName = a.Name
//	a.rolePid.Tell(msg1)
//}

// 登录成功查询
func (a *GateActor) SelectGate(ctx actor.Context) {
	//登录成功
	arg := ctx.Message().(*pb.SelectGate)
	glog.Debugf("SelectGate %#v", arg)
	userid := arg.GetUserid()
	glog.Debugf("SelectGate userid: %s", userid)
	//断开其它节点连接
	//a.logoutOther(userid, ctx)
	//在线表查询
	if p, ok := a.online[userid]; ok {
		if p.Token != arg.Token {
			//不是断线重连,断开当前节点旧连接
			msg1 := new(pb.LoginElse)
			msg1.Userid = userid
			msg1.Gate = a.Name
			p.Token = arg.Token
			p.Pid.Request(msg1, ctx.Self())
			glog.Debugf("SelectGate loginelse userid: %s,oldtoken:%s,newtoken:%s %s", userid, p.Token, arg.Token, p.Pid.String())
		}
		glog.Debugf("SelectGate online userid: %s, %s", userid, p.Pid.String())
		if arg.Token == "" {
			token, _ := handler.SignByTime(userid)
			p.Token = token
		}
		//响应登录
		rsp := new(pb.SelectedGate)
		rsp.Role = p.Pid
		rsp.Token = p.Token
		ctx.Respond(rsp)
		return
	}
	//离线表查找
	if p, ok := a.offline[userid]; ok {
		//切换到在线表
		a.online[userid] = p
		delete(a.offline, userid)
		delete(a.offtime, userid)
		glog.Debugf("SelectGate offline userid: %s, %s", userid, p.Pid.String())
		//响应登录
		if arg.Token == "" || arg.Token != p.Token {
			token, _ := handler.SignByTime(userid)
			p.Token = token
		}
		rsp := new(pb.SelectedGate)
		rsp.Role = p.Pid
		rsp.Token = p.Token
		ctx.Respond(rsp)
		return
	}
	//新玩家
	rolePid := a.spawnRole(userid)
	//添加
	token, _ := handler.SignByTime(userid)
	a.online[userid] = &data.LoginRole{Pid: rolePid, Token: token}
	//响应登录,不存在
	rsp := new(pb.SelectedGate)
	rsp.Role = rolePid
	rsp.Token = token
	ctx.Respond(rsp)
}

// 新玩家
func (a *GateActor) spawnRole(userid string) *actor.PID {
	newRole := NewRole()
	newRole.dbmsPid = a.dbmsPid
	newRole.roomPid = a.roomPid
	newRole.rolePid = a.rolePid
	// newRole.loggerPid = a.loggerPid
	// newRole.reportPid = a.reportPid
	rolePid := newRole.initRs()
	newRole.pid = rolePid
	msg1 := &pb.ServeStart{
		Message: userid,
	}
	rolePid.Tell(msg1)
	return rolePid
}
