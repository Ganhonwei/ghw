package main

import (
	"goserver/gen/pb"
	"goserver/pkg/game/algo"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
)

// hua

// 进入房间响应
func (r *Robot) recvAK47CoinEnterRoomRsp(s2c *pb.AK47CoinEnterRoomRsp) {
	var errcode = s2c.GetError()
	switch errcode {
	case pb.OK:
	default:
		glog.Errorf("comein err -> %d", errcode)
		rbs.PutRobot(r, int32(pb.HUA))
		return
	}
	roominfo := s2c.GetRoominfo()
	r.gtype = roominfo.Gtype
	r.rtype = roominfo.Rtype
	r.dtype = roominfo.Dtype
	r.roomid = roominfo.Roomid
	userinfo := s2c.GetUserinfo()
	for _, v := range userinfo {
		//只返回坐下玩家
		if v.Userid == r.data.Userid {
			glog.Debugf("comein user info -> %s", v.Userid)
			r.seat = v.Seat
			break
		} else { //记录其他玩家座位号
			r.seats = append(r.seats, v.Seat)
		}
	}
	//上桌发送表情
	r.EmojiDeskSend(r.GetRandSeat())
	// r.gameStart(s2c.Roominfo.State)
}

// 房间状态推送
func (r *Robot) recvAK47PushStateNtf(msg *pb.AK47PushStateNtf) {
	if msg.State == int32(pb.STATE_DEALING) {
		r.see = false
		r.alive = true
		r.ak47strategy = nil
		r.cards = []uint32{}
		r.state = 0
		r.sendAK47Ready2Req()
	} else if msg.State == int32(pb.STATE_BET) { //下注
		r.EmojiGameStart(r.GetRandSeat())
	}
}

// const (
// 	ActionNull     int = iota
// 	ActionSee          //看牌
// 	ActionPack         //弃牌
// 	ActionCall         //跟注
// 	ActionRaise        //加注
// 	ActionBi           //发起比牌
// 	ActionReplyBi      //回复比牌
// 	ActionAgreeBi      //同意比牌
// 	ActionRejectBi     //拒绝比牌
// )

// 随机操作
// func (r *Robot) randomAction(state int32) int {
// 	//如果是回复比牌动作，必须回复
// 	if state&int32(pb.ACT_REPLY_BI) != 0 {
// 		return ActionReplyBi
// 	}

// 	actions := []int{}
// 	if !r.see {
// 		actions = append(actions, ActionSee)
// 	}

// 	actions = append(actions, ActionPack)
// 	actions = append(actions, ActionCall)
// 	actions = append(actions, ActionRaise)

// 	if state&int32(pb.ACT_SIDESHOW) != 0 {
// 		actions = append(actions, ActionBi)
// 	} else if state&int32(pb.ACT_SHOW) != 0 {
// 		actions = append(actions, ActionBi)
// 	}

// 	len := len(actions)
// 	idx := utils.RandIntN(len)

// 	return actions[idx]
// }

func (r *Robot) ak47SimAction(state int32) int {
	if !r.see {
		return ActionSee
	}

	if len(r.cards) == 0 {
		return ActionPack
	}

	if state&int32(pb.ACT_REPLY_BI) != 0 {
		if algo.HuaType(r.cards) >= algo.TongHua {
			return ActionAgreeBi
		} else {
			return ActionRejectBi
		}
	}

	if algo.HuaType(r.cards) >= algo.TongHua {
		var choices []utils.Choice
		choices = append(choices, utils.Choice{Weight: 50, Item: ActionCall})

		if state&int32(pb.ACT_SIDESHOW) != 0 || state&int32(pb.ACT_SHOW) != 0 {
			choices = append(choices, utils.Choice{Weight: 50, Item: ActionBi})
		}
		c, _ := utils.WeightedChoice(choices)
		return c.Item.(int)
	} else {
		var choices []utils.Choice
		choices = append(choices, utils.Choice{Weight: 50, Item: ActionCall})
		choices = append(choices, utils.Choice{Weight: 50, Item: ActionPack})
		c, _ := utils.WeightedChoice(choices)
		return c.Item.(int)
	}
}

// 权重操作
func (r *Robot) ak47WeightAction(state int32, turn int32) int {
	if r.ak47strategy == nil {
		return ActionPack
	}

	//回复比牌
	if state&int32(pb.ACT_REPLY_BI) != 0 {
		agree := utils.RandWan(r.ak47strategy.AgreeBi)
		if agree {
			return ActionAgreeBi
		} else {
			return ActionRejectBi
		}
	}

	//看牌
	if !r.see {
		if utils.RandWan(r.ak47strategy.SeeWeight[0]) {
			return ActionSee
		}
	}

	//弃牌，比牌，跟注，加注
	index := int(turn)
	if index >= 0 && index < len(r.ak47strategy.ActionWeight) {
		aw := r.ak47strategy.ActionWeight[index]

		var choices []utils.Choice
		choices = append(choices, utils.Choice{Weight: int(aw.Values[0]), Item: ActionPack})
		choices = append(choices, utils.Choice{Weight: int(aw.Values[2]), Item: ActionCall})
		choices = append(choices, utils.Choice{Weight: int(aw.Values[3]), Item: ActionRaise})

		if state&int32(pb.ACT_SIDESHOW) != 0 || state&int32(pb.ACT_SHOW) != 0 {
			choices = append(choices, utils.Choice{Weight: int(aw.Values[1]), Item: ActionBi})
		}
		c, _ := utils.WeightedChoice(choices)
		return c.Item.(int)

	} else {
		return ActionPack
	}
}

func (r *Robot) ak47RunAction(action int) {
	switch action {
	case ActionSee:
		r.sendAK47CoinSeeReq()
	case ActionPack:
		r.sendAK47CoinFoldReq()
	case ActionCall:
		r.sendAK47CoinCallReq()
	case ActionRaise:
		r.sendAK47CoinRaiseReq()
	case ActionBi:
		r.sendAK47CoinBiReq()
	case ActionReplyBi:
		r.sendAK47CoinReplyBiReq(utils.RandBool())
	case ActionAgreeBi:
		r.sendAK47CoinReplyBiReq(true)
	case ActionRejectBi:
		r.sendAK47CoinReplyBiReq(false)
	}
}

// 看牌
func (r *Robot) AK47CoinSeeRsp(msg *pb.AK47CoinSeeRsp) {
	if r.sim {
		r.cards = msg.Cards
		action := r.ak47SimAction(r.state)
		r.ak47RunAction(action)
	}
}

// 离开
func (r *Robot) AK47LeaveRsp(msg *pb.AK47LeaveRsp) {
	if msg.Error == pb.OK {
		rbs.PutRobot(r, int32(pb.AK47))
	}
}

// 等待太久
func (r *Robot) AK47CoinWaitTooLongNtf(msg *pb.AK47CoinWaitTooLongNtf) {
	r.EmojiWaitLong(msg.Seat)
}

// 看牌推送
func (r *Robot) AK47CoinSeeNtf(msg *pb.AK47CoinSeeNtf) {
	if msg.Seat != r.seat {
		r.EmojiOtherSee(msg.Seat)
	}
}

// 加注推送
func (r *Robot) AK47CoinRaiseNtf(msg *pb.AK47CoinRaiseNtf) {
	if msg.Info.Seat != r.seat {
		r.EmojiOtherRaise(msg.Info.Seat)
	}
}

// 动作状态推送
func (r *Robot) recvAK47PushActStateNtf(msg *pb.AK47PushActStateNtf) {
	if r.alive && msg.Seat == r.seat { //自己操作
		if r.sim {
			r.state = msg.State
			action := r.ak47SimAction(msg.State)
			r.ak47RunAction(action)
		} else {
			action := r.ak47WeightAction(msg.State, msg.Turn)
			r.ak47RunAction(action)

			if action == ActionSee { //如果是看牌，还需要执行一次操作
				action := r.ak47WeightAction(msg.State, msg.Turn)
				r.ak47RunAction(action)
			}
		}
		// action := r.randomAction(msg.State)
		// if msg.State&int32(pb.ACT_REPLY_BI) == int32(pb.ACT_REPLY_BI) {
		// 	r.sendAK47CoinReplyBiReq(false)
		// } else {
		// 	r.sendAK47CoinCallReq()
		// }
		// if r.see {
		// 	r.sendAK47CoinBiReq()
		// } else {
		// 	r.sendAK47CoinCallReq()
		// }
		// r.sendAK47CoinFoldReq()
		// r.sendAK47CoinSeeReq()
		// r.sendAK47CoinBiReq()
		// r.sendAK47CoinCallReq()
		// r.sendAK47CoinRaiseReq()
	} else { //其他人操作
		if r.alive && !r.see && !r.sim {
			if r.ak47strategy != nil && utils.RandWan(r.ak47strategy.SeeWeight[1]) {
				r.sendAK47CoinSeeReq()
			}
		}
	}
}

// 比牌结果推送
func (r *Robot) recvAK47CoinBiNtf(msg *pb.AK47CoinBiNtf) {
	if r.alive && msg.Loseseat == r.seat {
		r.alive = false
	}

	//人机表情判断
	if msg.Winseat != r.seat && msg.Loseseat != r.seat { //没参与
		r.EmojiOtherBi(msg.Winseat, msg.Loseseat)
	} else if msg.Winseat == r.seat { //自己赢
		r.EmojiSelfBiWin(msg.Loseseat)
	}
}

// 人机策略推送
func (r *Robot) AK47CoinRobotStrategyNtf(msg *pb.AK47CoinRobotStrategyNtf) {
	r.ak47strategy = msg
	glog.Infof("robot strategy %#v", msg)
}

// 游戏结束
func (r *Robot) AK47CoinGameoverNtf(msg *pb.AK47CoinGameoverNtf) {
	if !r.sim && utils.RandWan(5000) {
		r.sendAK47LeaveReq()
	}
}
