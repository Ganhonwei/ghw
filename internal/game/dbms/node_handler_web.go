package dbms

import (
	"fmt"
	"strings"

	"goserver/gen/pb"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// web请求处理
func (a *DBMSActor) handlerWeb(arg *pb.WebRequest,
	rsp *pb.WebResponse, ctx actor.Context) {
	switch arg.Code {
	case pb.WebShop:
		//更新配置
		// msg2 := handler.SyncConfig2(pb.CONFIG_SHOP, arg.Atype, arg.Data)
		// err1 := handler.SyncConfig(msg2, a.Name)
		// if err1 != nil {
		// 	rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
		// 	return
		// }
		// //广播所有节点,主动通知同步配置,只同步修改数据
		// a.broadcast(msg2, ctx)
	case pb.WebEnv:
		//更新配置
		msg2 := handler.SyncConfig2(pb.CONFIG_ENV, arg.Atype, arg.Data)
		err1 := handler.SyncConfig(msg2, a.Name)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		//广播所有节点,主动通知同步配置,只同步修改数据
		a.broadcast(msg2, ctx)
	case pb.WebNotice:
		//更新配置
		msg2 := handler.SyncConfig2(pb.CONFIG_NOTICE, arg.Atype, arg.Data)
		err1 := handler.SyncConfig(msg2, a.Name)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		//广播所有节点,主动通知同步配置,只同步修改数据
		a.broadcast(msg2, ctx)
	case pb.WebGame:
		//更新配置
		msg2 := handler.SyncConfig2(pb.CONFIG_GAMES, arg.Atype, arg.Data)
		err1 := handler.SyncConfig(msg2, a.Name)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		//广播所有节点,主动通知同步配置,只同步修改数据
		a.broadcast(msg2, ctx)
	case pb.WebVip:
		//更新配置
		msg2 := handler.SyncConfig2(pb.CONFIG_VIP, arg.Atype, arg.Data)
		err1 := handler.SyncConfig(msg2, a.Name)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		//广播所有节点,主动通知同步配置,只同步修改数据
		a.broadcast(msg2, ctx)
	case pb.WebTask:
		//更新配置
		msg2 := handler.SyncConfig2(pb.CONFIG_TASK, arg.Atype, arg.Data)
		err1 := handler.SyncConfig(msg2, a.Name)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		//广播所有节点,主动通知同步配置,只同步修改数据
		a.broadcast(msg2, ctx)
	case pb.WebLogin:
		//更新配置
		msg2 := handler.SyncConfig2(pb.CONFIG_LOGIN, arg.Atype, arg.Data)
		err1 := handler.SyncConfig(msg2, a.Name)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		//广播所有节点,主动通知同步配置,只同步修改数据
		a.broadcast(msg2, ctx)
	case pb.WebStat:
		//测试接口
		// msg2 := new(pb.AgentActivityStat)
		// err1 := msg2.Unmarshal(arg.Data)
		// if err1 != nil {
		// 	rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
		// 	return
		// }
		// if a.statPid != nil {
		// 	a.statPid.Tell(msg2)
		// }
	case pb.WebSetWithDraw:
		msg2 := handler.SyncConfig2(pb.CONFIG_WITHDRAW, arg.Atype, arg.Data)
		err1 := handler.SyncConfig(msg2, a.Name)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		//广播所有节点,主动通知同步配置,只同步修改数据
		a.broadcast(msg2, ctx)
	case pb.WebSwitch:
		msg2 := handler.SyncConfig2(pb.CONFIG_SWITCH, arg.Atype, arg.Data)
		err1 := handler.SyncConfig(msg2, a.Name)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		//广播所有节点,主动通知同步配置,只同步修改数据
		a.broadcast(msg2, ctx)
	case pb.WebUploadConfig:
		msg2 := handler.SyncConfig2(pb.CONFIG_UPLOAD_CONFIG, arg.Atype, arg.Data)
		err := handler.SyncConfig(msg2, a.Name)
		if err != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err)
			return
		}
		a.broadcast(msg2, ctx)
		a.broadcastRobot(msg2, ctx)
	case pb.WebReloadConfig:
		msg2 := handler.SyncConfig3(pb.CONFIG_RELOAD)
		err := handler.SyncConfig(msg2, a.Name)
		if err != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err)
			return
		}
		a.broadcast(msg2, ctx)
		a.broadcastRobot(msg2, ctx)
	case pb.WebCloseServer:
		msg := &pb.CloseServer{}
		a.broadcast(msg, ctx)
		rolePid.Tell(msg)
	case pb.WebPayChannel:
		msg2 := handler.SyncConfig2(pb.CONFIG_PAY_CHANNEL, arg.Atype, arg.Data)
		err1 := handler.SyncConfig(msg2, a.Name)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		//广播所有节点,主动通知同步配置,只同步修改数据
		a.broadcast(msg2, ctx)
	case pb.WebIpWhite:
		// ip白名单
		msg2 := handler.SyncConfig2(pb.CONFIG_IPWHITE, arg.Atype, arg.Data)
		err1 := handler.SyncConfig(msg2, a.Name)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		a.broadcastgate(msg2, ctx)
	case pb.WebRechareBlack:
		// 充值限制
		msg2 := handler.SyncConfig2(pb.CONFIG_RECHARGELIMIT, arg.Atype, arg.Data)
		err1 := handler.SyncConfig(msg2, a.Name)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		a.broadcastgate(msg2, ctx)
	case pb.WebChannel:
		// 渠道包名
		msg2 := handler.SyncConfig2(pb.CONFIG_CHANNEL, arg.Atype, arg.Data)
		err1 := handler.SyncConfig(msg2, a.Name)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		a.broadcastgate(msg2, ctx)
	case pb.WebServerWhite:
		// ip白名单
		msg2 := handler.SyncConfig2(pb.CONFIG_SEVERWHITE, arg.Atype, arg.Data)
		err1 := handler.SyncConfig(msg2, a.Name)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		a.broadcast(msg2, ctx)
	case pb.WebModifyCustomer:
		// 客服地址
		msg2 := handler.SyncConfig2(pb.CONFIG_CUSTOMERADDR, arg.Atype, arg.Data)
		err1 := handler.SyncConfig(msg2, a.Name)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		a.broadGateAndLogin(msg2, ctx)
	case pb.WebDeviceBlack:
		// 设备码黑名单
		msg2 := handler.SyncConfig2(pb.CONFIG_DEVICELIMIT, arg.Atype, arg.Data)
		err1 := handler.SyncConfig(msg2, a.Name)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		a.broadcastgate(msg2, ctx)
	case pb.WebModifyShare:
		// 分享配置
		msg2 := handler.SyncConfig2(pb.CONFIG_SHARE_WAY, arg.Atype, arg.Data)
		err1 := handler.SyncConfig(msg2, a.Name)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		a.broadGateAndLogin(msg2, ctx)
	default:
		glog.Errorf("unknown message %v", arg)
	}
}

// 广播所有节点,游戏逻辑服,dbms
func (a *DBMSActor) broadcast(msg interface{}, ctx actor.Context) {
	for k, v := range a.serve {
		if strings.Contains(k, "gate.") ||
			strings.Contains(k, "game.") {
			v.Tell(msg)
		}
	}
}

// 广播所有节点,游戏逻辑服,dbms
func (a *DBMSActor) broadcastgate(msg interface{}, ctx actor.Context) {
	for k, v := range a.serve {
		if strings.Contains(k, "gate.") {
			v.Tell(msg)
		}
	}
}

// 广播所有节点,游戏逻辑服,dbms
func (a *DBMSActor) broadGateAndLogin(msg interface{}, ctx actor.Context) {
	for k, v := range a.serve {
		if strings.Contains(k, "gate.") ||
			strings.Contains(k, "login") {
			v.Tell(msg)
		}
	}
}

// 广播所有机器人节点
func (a *DBMSActor) broadcastRobot(msg interface{}, ctx actor.Context) {
	for k, v := range a.serve {
		if strings.Contains(k, "robot.") {
			v.Tell(msg)
		}
	}
}
