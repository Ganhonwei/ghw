package dbms

import (
	"fmt"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/config"
	"goserver/pkg/game/event"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
	jsoniter "github.com/json-iterator/go"
)

// web请求处理
func (a *RoleActor) handlerWeb(arg *pb.WebRequest,
	rsp *pb.WebResponse, ctx actor.Context) {
	switch arg.Code {
	case pb.WebOnline:
		msg1 := make([]string, 0)
		err1 := json.Unmarshal(arg.Data, &msg1)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		//响应
		resp := make(map[string]int)
		for _, v := range msg1 {
			if _, ok := a.online[v]; ok {
				resp[v] = 1
			} else {
				resp[v] = 0
			}
		}
		result, err2 := json.Marshal(resp)
		if err2 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err2)
			return
		}
		rsp.Result = result
	case pb.WebBuild:
		//后台设置绑定关系
		// msg2 := new(pb.SetAgentBuild)
		// err1 := msg2.Unmarshal(arg.Data)
		// if err1 != nil {
		// 	rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
		// 	return
		// }
		// err2 := a.setBuild(msg2)
		// if err2 != nil {
		// 	rsp.ErrMsg = fmt.Sprintf("msg err: %v", err2)
		// }
	case pb.WebGive:
		//后台货币赠送同步到game房间
		msg2 := new(pb.PayCurrency)
		err1 := msg2.Unmarshal(arg.Data)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		//消息
		a.msg2role(msg2)
	case pb.WebNumber:
		result, err2 := a.getNumber(ctx)
		if err2 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err2)
			return
		}
		rsp.Result = result
	case pb.WebRate:
		//后台设置区域奖励百分比
		// msg2 := new(pb.SetAgentProfitRate)
		// err1 := msg2.Unmarshal(arg.Data)
		// if err1 != nil {
		// 	rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
		// 	return
		// }
		// err2 := a.setRate(msg2)
		// if err2 != nil {
		// 	rsp.ErrMsg = fmt.Sprintf("msg err: %v", err2)
		// }
	case pb.WebState:
		//后台设置代理
		// msg2 := new(pb.SetAgentState)
		// err1 := msg2.Unmarshal(arg.Data)
		// if err1 != nil {
		// 	rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
		// 	return
		// }
		// err2 := a.setState(msg2)
		// if err2 != nil {
		// 	rsp.ErrMsg = fmt.Sprintf("msg err: %v", err2)
		// }
	case pb.WebVaild:
		//后台设置代理
		// msg2 := new(pb.AgentBuildUpdate)
		// err1 := msg2.Unmarshal(arg.Data)
		// if err1 != nil {
		// 	rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
		// 	return
		// }
		// a.agentBuildUpdate(msg2)
	case pb.WebBlack:
		msg2 := new(data.BlackList)
		err1 := jsoniter.Unmarshal(arg.Data, msg2)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		a.blackListOperate(msg2, arg.Atype)
	case pb.WebWithdraw:
		// 提现审核
		msg2 := new(data.WithdrawOpreate)
		err1 := jsoniter.Unmarshal(arg.Data, msg2)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		a.withdrawExamine(rsp, msg2)
	case pb.WebFeedBack:
		// 客服消息
		msg2 := new(data.ChatLog)
		err1 := jsoniter.Unmarshal(arg.Data, msg2)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		a.customerReply(rsp, msg2)
	case pb.WebModifyUser:
		// 修改玩家数据
		msg2 := new(data.ModifyUserData)
		err1 := jsoniter.Unmarshal(arg.Data, msg2)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		a.modifyUserData(rsp, msg2)
	case pb.WebPointControl:
		msg := new(data.PointControl)
		err := jsoniter.Unmarshal(arg.Data, msg)
		if err != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err)
			return
		}
		a.pointControl(rsp, msg)
	case pb.WebModifyNum:
		msg2 := new(pb.ModifyCurrency)
		err1 := jsoniter.Unmarshal(arg.Data, msg2)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		rolePid.Tell(msg2)
	case pb.WebSuperior:
		msg := new(pb.ChangeShareSuperior)
		err1 := jsoniter.Unmarshal(arg.Data, msg)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		err := a.changeShareSuperior(msg)
		if err != nil {
			rsp.ErrMsg = fmt.Sprintf("%v", err)
		}
	case pb.WebOnlineUser:
		// 查看在线玩家数据
		a.buildOnlineUser(rsp)
		// req := new(pb.OnlineUser)
		// datas := make([]data.OnlineUser, 0)
		// for k, role := range a.roles {
		// 	user := a.getUser(k)
		// 	if user == nil {
		// 		continue
		// 	}
		// 	bean := data.OnlineUser{
		// 		Id:         user.Userid,
		// 		Nanme:      user.Nickname,
		// 		Channel:    user.AD_BundleId,
		// 		Asset:      user.GetAsset(),
		// 		ShowAsset:  user.Diamond + user.Coin,
		// 		Cash:       user.Diamond + user.ShadowDiamond,
		// 		Bouns:      user.Coin,
		// 		OtherAsset: user.ShadowDiamond,
		// 		Recharge:   user.Money,
		// 		CashOut:    uint32(user.CashOut),
		// 		WinScore:   0,
		// 		GameId:     "0",
		// 		RoomId:     "N/A",
		// 	}

		// 	// 获取玩家当前位置判断是否在游戏中
		// 	body, err := role.Pid.RequestFuture(req, 1*time.Second).Result()
		// 	if err != nil {
		// 		datas = append(datas, bean)
		// 		glog.Error("[onlineuser] err：", err)
		// 		continue
		// 	}
		// 	if u, ok := body.(*pb.OnlinedUser); ok {
		// 		game := config.GetGame(u.GameId)
		// 		if game.Id != "" {
		// 			bean.GameId = utils.String(game.Gtype)
		// 			bean.RoomId = u.RoomId
		// 		}
		// 		datas = append(datas, bean)
		// 	}
		// }
		// result, err := json.Marshal(datas)
		// if err == nil {
		// 	rsp.Result = result
		// }
	case pb.WebGiveWithdraw:
		// 增加可提现金额
		msg := new(data.GiveWithdrawCash)
		err1 := jsoniter.Unmarshal(arg.Data, msg)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		if msg.Cash <= 0 {
			rsp.ErrMsg = fmt.Sprintf("cash value illegality: %d", msg.Cash)
			return
		}
		a.giveWithdraw(msg)
	case pb.WebPayCallback:
		// 支付手动回调
		msg := ""
		err1 := jsoniter.Unmarshal(arg.Data, &msg)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		a.payManualCallback(msg, rsp)
	case pb.WebTpStoryStockMin: // tp剧情库存最低使用值查询
		result, err2 := a.getTpStoryStockMin(ctx)
		if err2 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err2)
			return
		}
		rsp.Result = result
	case pb.WebUserPhoto:
		msg := new(data.UserCustomPhoto)
		err1 := jsoniter.Unmarshal(arg.Data, &msg)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		a.userCustomPhoto(msg)
	case pb.WebTurnAudit:
		// 转盘活动审核
		msg2 := new(data.ActivityTurnPrizeOperate)
		err1 := jsoniter.Unmarshal(arg.Data, msg2)
		if err1 != nil {
			rsp.ErrMsg = fmt.Sprintf("msg err: %v", err1)
			return
		}
		a.webTurnAudit(rsp, msg2)
	default:
		glog.Errorf("unknown message %v", arg)
	}
}

// 消息通知到玩家
func (a *RoleActor) msg2role(arg *pb.PayCurrency) {
	if v, ok := a.roles[arg.Userid]; ok && v != nil {
		if arg.Diamond > 0 {
			// gate addMoney 时不会及时同步dbms
			if user := a.getUser(arg.Userid); user != nil {
				user.AddMoney(uint32(arg.Diamond))
			}
		}
		//存活在节点中
		v.Pid.Tell(arg)
		return
	}
	a.syncCurrency(arg.Diamond, 0, arg.Give, 0, 0, arg.Type, arg.Userid, arg.Desc, "", false)
	if user := a.getUserById(arg.Userid); user != nil {
		user.AddMoney(uint32(arg.Diamond))
		user.UpdateMoney()
		// vip
		event.Event(user, event.VIP, &event.VIPEvent{Amount: arg.Diamond})
		user.UpdateVIP()
	}
	// 邮件
	if config.SettingIsOpen(4, data.MAILREPLY) {
		user := a.getUserById(arg.Userid)
		event.Event(user, event.GIVE_CASH, &event.GiveCash{Amount: uint32(arg.Diamond)})
		user.UpdateFeedBack()
	}
}

// 获取在线人数
func (a *RoleActor) getNumber(ctx actor.Context) ([]byte, error) {
	//响应1 机器人,2 玩家
	resp := make(map[int]int)
	for _, v := range a.online {
		if v.GetRobot() {
			resp[1]++
		} else {
			resp[2]++
		}
	}
	result, err2 := json.Marshal(resp)
	if err2 != nil {
		glog.Errorf("msg err: %v", err2)
		return nil, fmt.Errorf("msg err: %v", err2)
	}
	return result, nil
}

// 设置区域奖励百分比
// func (a *RoleActor) setRate(arg *pb.SetAgentProfitRate) error {
// 	agent := a.getUserById(arg.GetUserid())
// 	if agent == nil {
// 		return fmt.Errorf("userid %s not exist", arg.GetUserid())
// 	}
// 	if arg.GetRate() > 38 || arg.GetRate() == 0 {
// 		return fmt.Errorf("rate %d error", arg.GetRate())
// 	}
// 	if agent.ProfitRateSum != 0 {
// 		//return fmt.Errorf("AlreadySetRate")
// 	}
// 	if !handler.IsVaild(agent) {
// 		//return fmt.Errorf("ProfitLimit")
// 	}
// 	if !handler.IsAgent(agent) {
// 		return fmt.Errorf("NotAgent")
// 	}
// 	// a.agentProfitRate(arg)
// 	return nil
// }

// 后台设置绑定关系
// func (a *RoleActor) setBuild(arg *pb.SetAgentBuild) error {
// 	user := a.getUserById(arg.GetUserid())
// 	if user == nil {
// 		return fmt.Errorf("userid %s not exist", arg.GetUserid())
// 	}
// 	//TODO 限制条件,绑定数量统计和日志
// 	if arg.GetAgent() != "" {
// 		agent := a.getUserById(arg.GetAgent())
// 		if agent == nil {
// 			return fmt.Errorf("agent %s not exist", arg.GetAgent())
// 		}
// 		if !handler.IsAgent(agent) {
// 			return fmt.Errorf("NotAgent")
// 		}
// 	}
// 	if v, ok := a.roles[arg.Userid]; ok && v != nil {
// 		v.Pid.Tell(arg)
// 		//return
// 	}
// 	handler.SetAgentBuild(arg, user)
// 	user.UpdateAgent()
// 	return nil
// }

// 后台设置代理
// func (a *RoleActor) setState(arg *pb.SetAgentState) error {
// 	user := a.getUserById(arg.GetUserid())
// 	if user == nil {
// 		return fmt.Errorf("userid %s not exist", arg.GetUserid())
// 	}
// 	//TODO 限制条件
// 	if v, ok := a.roles[arg.Userid]; ok && v != nil {
// 		v.Pid.Tell(arg)
// 		//return
// 	}
// 	handler.SetAgentState(arg, user)
// 	user.UpdateAgentJoin()
// 	return nil
// }

// 在线玩家数据
func (a *RoleActor) buildOnlineUser(rsp *pb.WebResponse) error {
	// req := new(pb.OnlineUser)
	datas := make([]data.OnlineUser, 0)
	// for k, role := range a.roles {
	// 	user := a.getUser(k)
	// 	if user == nil {
	// 		continue
	// 	}
	// 	bean := data.OnlineUser{
	// 		Id:         user.Userid,
	// 		Nanme:      user.Nickname,
	// 		Channel:    user.AD_BundleId,
	// 		Asset:      user.GetAsset(),
	// 		ShowAsset:  user.Diamond + user.Coin,
	// 		Cash:       user.Diamond + user.ShadowDiamond,
	// 		Bouns:      user.Coin,
	// 		OtherAsset: user.ShadowDiamond,
	// 		Recharge:   user.Money,
	// 		CashOut:    uint32(user.CashOut),
	// 		WinScore:   0,
	// 		GameId:     "0",
	// 		RoomId:     "N/A",
	// 	}

	// 获取玩家当前位置判断是否在游戏中
	res := &pb.OnlineUserListRsp{}
	err := mq.NatsRequest(mq.RequestOnlineUserList, res, &mq.RequestEmptyArgs{})
	if err != nil {
		glog.Error("requestOnlineUserListReq", "error", err)
		return err
	}
	for _, o := range res.UserList {
		user := a.getUserById(o.Userid)
		if user == nil {
			continue
		}
		bean := data.OnlineUser{
			Id:         user.Userid,
			Nanme:      user.Nickname,
			Channel:    user.AD_BundleId,
			Asset:      user.GetAsset(),
			ShowAsset:  user.Diamond + user.Coin,
			Cash:       user.Diamond + user.ShadowDiamond,
			Bouns:      user.Coin,
			OtherAsset: user.ShadowDiamond,
			Recharge:   user.Money,
			CashOut:    uint32(user.CashOut),
			WinScore:   0,
			GameId:     "0",
			RoomId:     "N/A",
		}
		if o.Gameid != "" && o.Gameid != "0" {
			bean.GameId = o.Gameid
			bean.RoomId = o.Roomid
		}
		datas = append(datas, bean)
	}
	// body, err := role.Pid.RequestFuture(req, 50*time.Millisecond).Result()
	// if err != nil {
	// 	datas = append(datas, bean)
	// 	glog.Error("[onlineuser] err：,user:%s", err, user.Userid)
	// 	continue
	// }
	// if u, ok := body.(*pb.OnlinedUser); ok {
	// 	if !u.Online {
	// 		// 只统计在线玩家
	// 		continue
	// 	}
	// 	bean.RoomId = u.RoomId
	// 	bean.GameId = u.Gtype
	// 	// game := config.GetGame(u.GameId)
	// 	// if game.Id != "" {
	// 	// 	bean.GameId = utils.String(game.Gtype)
	// 	// }
	// 	datas = append(datas, bean)
	// }
	// }
	result, err := json.Marshal(datas)
	if err == nil {
		rsp.Result = result
	}
	return nil
}
