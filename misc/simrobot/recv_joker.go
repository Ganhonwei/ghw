package main

import (
	"goserver/gen/pb"
	"goserver/pkg/game/algo"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
)

// hua

// 进入房间响应
func (r *Robot) recvJOKERCoinEnterRoomRsp(s2c *pb.JOKERCoinEnterRoomRsp) {
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
func (r *Robot) recvJOKERPushStateNtf(msg *pb.JOKERPushStateNtf) {
	if msg.State == int32(pb.STATE_DEALING) {
		r.see = false
		r.alive = true
		r.jokerstrategy = nil
		r.cards = []uint32{}
		r.state = 0
		r.sendJOKERReady2Req()
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

func (r *Robot) jokerSimAction(state int32) int {
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
func (r *Robot) jokerWeightAction(state int32, turn int32) int {
	if r.jokerstrategy == nil {
		return ActionPack
	}

	//回复比牌
	if state&int32(pb.ACT_REPLY_BI) != 0 {
		agree := utils.RandWan(r.jokerstrategy.AgreeBi)
		if agree {
			return ActionAgreeBi
		} else {
			return ActionRejectBi
		}
	}

	//看牌
	if !r.see {
		if utils.RandWan(r.jokerstrategy.SeeWeight[0]) {
			return ActionSee
		}
	}

	//弃牌，比牌，跟注，加注
	index := int(turn)
	if index >= 0 && index < len(r.jokerstrategy.ActionWeight) {
		aw := r.jokerstrategy.ActionWeight[index]

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

func (r *Robot) jokerRunAction(action int) {
	switch action {
	case ActionSee:
		r.sendJOKERCoinSeeReq()
	case ActionPack:
		r.sendJOKERCoinFoldReq()
	case ActionCall:
		r.sendJOKERCoinCallReq()
	case ActionRaise:
		r.sendJOKERCoinRaiseReq()
	case ActionBi:
		r.sendJOKERCoinBiReq()
	case ActionReplyBi:
		r.sendJOKERCoinReplyBiReq(utils.RandBool())
	case ActionAgreeBi:
		r.sendJOKERCoinReplyBiReq(true)
	case ActionRejectBi:
		r.sendJOKERCoinReplyBiReq(false)
	}
}

// 看牌
func (r *Robot) JOKERCoinSeeRsp(msg *pb.JOKERCoinSeeRsp) {
	if r.sim {
		r.cards = msg.Cards
		action := r.jokerSimAction(r.state)
		r.jokerRunAction(action)
	}
}

// 离开
func (r *Robot) JOKERLeaveRsp(msg *pb.JOKERLeaveRsp) {
	if msg.Error == pb.OK {
		rbs.PutRobot(r, int32(pb.JOKER))
	}
}

// 等待太久
func (r *Robot) JOKERCoinWaitTooLongNtf(msg *pb.JOKERCoinWaitTooLongNtf) {
	r.EmojiWaitLong(msg.Seat)
}

// 看牌推送
func (r *Robot) JOKERCoinSeeNtf(msg *pb.JOKERCoinSeeNtf) {
	if msg.Seat != r.seat {
		r.EmojiOtherSee(msg.Seat)
	}
}

// 加注推送
func (r *Robot) JOKERCoinRaiseNtf(msg *pb.JOKERCoinRaiseNtf) {
	if msg.Info.Seat != r.seat {
		r.EmojiOtherRaise(msg.Info.Seat)
	}
}

// 动作状态推送
func (r *Robot) recvJOKERPushActStateNtf(msg *pb.JOKERPushActStateNtf) {
	if r.alive && msg.Seat == r.seat { //自己操作
		if r.sim {
			r.state = msg.State
			action := r.jokerSimAction(msg.State)
			r.jokerRunAction(action)
		} else {
			action := r.jokerWeightAction(msg.State, msg.Turn)
			r.jokerRunAction(action)

			if action == ActionSee { //如果是看牌，还需要执行一次操作
				action := r.jokerWeightAction(msg.State, msg.Turn)
				r.jokerRunAction(action)
			}
		}
		// action := r.randomAction(msg.State)
		// if msg.State&int32(pb.ACT_REPLY_BI) == int32(pb.ACT_REPLY_BI) {
		// 	r.sendJOKERCoinReplyBiReq(false)
		// } else {
		// 	r.sendJOKERCoinCallReq()
		// }
		// if r.see {
		// 	r.sendJOKERCoinBiReq()
		// } else {
		// 	r.sendJOKERCoinCallReq()
		// }
		// r.sendJOKERCoinFoldReq()
		// r.sendJOKERCoinSeeReq()
		// r.sendJOKERCoinBiReq()
		// r.sendJOKERCoinCallReq()
		// r.sendJOKERCoinRaiseReq()
	} else { //其他人操作
		if r.alive && !r.see && !r.sim {
			if r.jokerstrategy != nil && utils.RandWan(r.jokerstrategy.SeeWeight[1]) {
				r.sendJOKERCoinSeeReq()
			}
		}
	}
}

// 比牌结果推送
func (r *Robot) recvJOKERCoinBiNtf(msg *pb.JOKERCoinBiNtf) {
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
func (r *Robot) JOKERCoinRobotStrategyNtf(msg *pb.JOKERCoinRobotStrategyNtf) {
	r.jokerstrategy = msg
	glog.Infof("robot strategy %#v", msg)
}

// 游戏结束
func (r *Robot) JOKERCoinGameoverNtf(msg *pb.JOKERCoinGameoverNtf) {
	if !r.sim && utils.RandWan(5000) {
		r.sendJOKERLeaveReq()
	}
}
