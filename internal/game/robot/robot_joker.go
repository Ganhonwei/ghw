package robot

import (
	"goserver/gen/pb"
	"goserver/pkg/game/algo"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 进入或匹配桌子
func (rs *RoleActor) enterJokerMatchDesk(arg *pb.RobotMsg) *pb.MatchDesk {
	//获取游戏服节点或者房间进程
	msg := new(pb.MatchDesk)
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.joker").Name()
	msg.Gtype = int32(pb.JOKER)
	msg.Roomid = arg.Roomid
	msg.Gameid = arg.Gameid
	return msg
}

// 进入房间响应
func (r *RoleActor) recvJOKERCoinEnterRoomRsp(ctx actor.Context) {
	s2c := ctx.Message().(*pb.JOKERCoinEnterRoomRsp)
	var errcode = s2c.GetError()
	switch errcode {
	case pb.OK:
	default:
		// 关闭节点
		glog.Errorf("enter ak47 fail user:%s,err:%v ", r.Userid, errcode)
		r.closeRs()
		return
	}
	roominfo := s2c.GetRoominfo()
	r.gtype = roominfo.Gtype
	// r.rtype = roominfo.Rtype
	// r.dtype = roominfo.Dtype
	// r.roomid = roominfo.Roomid
	userinfo := s2c.GetUserinfo()
	for _, v := range userinfo {
		//只返回坐下玩家
		if v.Userid == r.Userid {
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
func (r *RoleActor) recvJOKERPushStateNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.JOKERPushStateNtf)
	if msg.State == int32(pb.STATE_DEALING) { //发牌
		r.alive = true
		r.sendJOKERReady2Req()
	} else if msg.State == int32(pb.STATE_BET) { //下注
		r.EmojiGameStart(r.GetRandSeat())
	} else if msg.State == int32(pb.STATE_OVER) { //游戏结束
		r.see = false
		r.jokerstrategy = nil
		r.cards = []uint32{}
		r.state = 0
		r.core = false
		r.seeAndPack = false
		r.identity = ""
	}
}

func (r *RoleActor) jokerIdentityABAction(msg *pb.JOKERPushActStateNtf) int {
	if !msg.IsCharge { //没有充值
		if msg.Condition == 1 {
			// 拒绝比牌
			if msg.State&int32(pb.ACT_REPLY_BI) != 0 {
				return ActionRejectBi
			}

			//看牌
			if !r.see && msg.Turn < 4 {
				if !utils.RandWan(6000) {
					return ActionSee
				}
			}

			var choices []utils.Choice
			choices = append(choices, utils.Choice{Weight: 75, Item: ActionCall})  //下注
			choices = append(choices, utils.Choice{Weight: 25, Item: ActionRaise}) //加注

			c, _ := utils.WeightedChoice(choices)
			return c.Item.(int)
		} else if msg.Condition == 2 {
			//拒绝比牌
			if msg.State&int32(pb.ACT_REPLY_BI) != 0 {
				return ActionRejectBi
			}

			//看牌
			if !r.see && msg.Turn < 4 {
				if !utils.RandWan(6000) {
					return ActionSee
				}
			}

			return ActionCall //100%下注

		} else if msg.Condition == 3 {
			//同意比牌
			if msg.State&int32(pb.ACT_REPLY_BI) != 0 {
				return ActionAgreeBi
			}

			//看牌
			if !r.see && msg.Turn < 4 {
				if !utils.RandWan(6000) {
					return ActionSee
				}
			}

			if msg.Turn+1 >= 0 && !msg.IsPrevPlayer { //第0轮以后上家不是真人
				var choices []utils.Choice
				choices = append(choices, utils.Choice{Weight: 50, Item: ActionCall}) //下注

				if msg.State&int32(pb.ACT_SIDESHOW) != 0 || msg.State&int32(pb.ACT_SHOW) != 0 {
					choices = append(choices, utils.Choice{Weight: 50, Item: ActionBi}) //比牌
				}

				c, _ := utils.WeightedChoice(choices)
				return c.Item.(int)
			} else {
				return ActionCall
			}

		} else if msg.Condition == 4 {
			// 同意比牌
			if msg.State&int32(pb.ACT_REPLY_BI) != 0 {
				return ActionAgreeBi
			}

			//看牌
			if !r.see && msg.Turn < 4 {
				if !utils.RandWan(6000) {
					return ActionSee
				}
			}

			if msg.Turn < 6 {
				var choices []utils.Choice
				choices = append(choices, utils.Choice{Weight: 75, Item: ActionCall})  //下注
				choices = append(choices, utils.Choice{Weight: 25, Item: ActionRaise}) //加注

				c, _ := utils.WeightedChoice(choices)
				return c.Item.(int)
			} else {
				var choices []utils.Choice
				choices = append(choices, utils.Choice{Weight: 40, Item: ActionCall})  //下注
				choices = append(choices, utils.Choice{Weight: 20, Item: ActionRaise}) //加注

				if msg.State&int32(pb.ACT_SIDESHOW) != 0 || msg.State&int32(pb.ACT_SHOW) != 0 {
					choices = append(choices, utils.Choice{Weight: 40, Item: ActionBi}) //比牌
				}

				c, _ := utils.WeightedChoice(choices)
				return c.Item.(int)
			}
		}

	} else { //充值了

		if msg.Bigger {
			//同意比牌
			if msg.State&int32(pb.ACT_REPLY_BI) != 0 {
				return ActionAgreeBi
			}

			// 看牌
			if !r.see && msg.Turn < 4 {
				if !utils.RandWan(6000) {
					return ActionSee
				}
			}

			var choices []utils.Choice
			choices = append(choices, utils.Choice{Weight: 50, Item: ActionCall})  //下注
			choices = append(choices, utils.Choice{Weight: 10, Item: ActionRaise}) //加注

			if msg.State&int32(pb.ACT_SIDESHOW) != 0 || msg.State&int32(pb.ACT_SHOW) != 0 {
				choices = append(choices, utils.Choice{Weight: 40, Item: ActionBi}) //比牌
			}

			c, _ := utils.WeightedChoice(choices)
			return c.Item.(int)
		} else {
			if msg.Condition == 1 {
				//同意比牌
				if msg.State&int32(pb.ACT_REPLY_BI) != 0 {
					return ActionAgreeBi
				}

				//看牌
				if !r.see && msg.Turn < 4 {
					if !utils.RandWan(6000) {
						return ActionSee
					}
				}

				var choices []utils.Choice
				choices = append(choices, utils.Choice{Weight: 30, Item: ActionCall}) //下注
				choices = append(choices, utils.Choice{Weight: 50, Item: ActionPack}) //弃牌

				if msg.State&int32(pb.ACT_SIDESHOW) != 0 || msg.State&int32(pb.ACT_SHOW) != 0 {
					choices = append(choices, utils.Choice{Weight: 20, Item: ActionBi}) //比牌
				}

				c, _ := utils.WeightedChoice(choices)
				return c.Item.(int)

			} else if msg.Condition == 2 {

				// 同意比牌
				if msg.State&int32(pb.ACT_REPLY_BI) != 0 {
					return ActionAgreeBi
				}

				//看牌
				if !r.see && msg.Turn < 4 {
					if !utils.RandWan(6000) {
						return ActionSee
					}
				}

				// 100%弃牌
				return ActionPack

				// if r.identity == "a" {
				// 	//同意比牌
				// 	if msg.State&int32(pb.ACT_REPLY_BI) != 0 {
				// 		return ActionAgreeBi
				// 	}

				// 	//看牌
				// 	if !r.see && msg.Turn < 4 {
				// 		if !utils.RandWan(6000) {
				// 			return ActionSee
				// 		}
				// 	}

				// 	var choices []utils.Choice
				// 	choices = append(choices, utils.Choice{Weight: 30, Item: ActionCall}) //下注
				// 	choices = append(choices, utils.Choice{Weight: 50, Item: ActionPack}) //下注

				// 	if msg.State&int32(pb.ACT_SIDESHOW) != 0 || msg.State&int32(pb.ACT_SHOW) != 0 {
				// 		choices = append(choices, utils.Choice{Weight: 20, Item: ActionBi}) //比牌
				// 	}

				// 	c, _ := utils.WeightedChoice(choices)
				// 	return c.Item.(int)

				// } else if r.identity == "b" {
				// 	// 同意比牌
				// 	if msg.State&int32(pb.ACT_REPLY_BI) != 0 {
				// 		return ActionAgreeBi
				// 	}

				// 	//看牌
				// 	if !r.see && msg.Turn < 4 {
				// 		if !utils.RandWan(6000) {
				// 			return ActionSee
				// 		}
				// 	}

				// 	//100%弃牌
				// 	return ActionPack

				// }
			}
		}

		// } else if msg.Condition == 3 {
		// 	// 同意比牌
		// 	if msg.State&int32(pb.ACT_REPLY_BI) != 0 {
		// 		return ActionAgreeBi
		// 	}

		// 	//看牌
		// 	if !r.see && msg.Turn < 4 {
		// 		if !utils.RandWan(6000) {
		// 			return ActionSee
		// 		}
		// 	}

		// 	//100%弃牌
		// 	return ActionPack
		// }
	}
	return ActionPack
}

func (r *RoleActor) jokerIdentityCDAction(msg *pb.JOKERPushActStateNtf) int {
	if !msg.IsCharge { //未充值
		//同意比牌
		if msg.State&int32(pb.ACT_REPLY_BI) != 0 {
			return ActionAgreeBi
		}

		//看牌
		if !r.see {
			if utils.RandWan(5000) {
				return ActionSee
			}
		}

		var choices []utils.Choice
		// if msg.Turn == 0 {
		// 	choices = append(choices, utils.Choice{Weight: 50, Item: ActionCall}) //下注
		// 	choices = append(choices, utils.Choice{Weight: 50, Item: ActionPack}) //弃牌
		// } else if msg.Turn == 1 {
		// 	choices = append(choices, utils.Choice{Weight: 25, Item: ActionCall}) //下注
		// 	choices = append(choices, utils.Choice{Weight: 75, Item: ActionPack}) //弃牌
		// } else {
		// 	choices = append(choices, utils.Choice{Weight: 100, Item: ActionPack}) //弃牌
		// }
		if msg.Turn < 4 { //前4轮
			choices = append(choices, utils.Choice{Weight: 50, Item: ActionCall}) //下注
			choices = append(choices, utils.Choice{Weight: 50, Item: ActionPack}) //弃牌
		} else {
			choices = append(choices, utils.Choice{Weight: 100, Item: ActionPack}) //弃牌
		}

		c, _ := utils.WeightedChoice(choices)
		return c.Item.(int)

	} else { //充值了
		if msg.Condition == 1 {
			// 同意比牌
			if msg.State&int32(pb.ACT_REPLY_BI) != 0 {
				return ActionAgreeBi
			}

			//看牌
			if !r.see {
				if utils.RandWan(5000) {
					return ActionSee
				}
			}

			// var choices []utils.Choice
			// choices = append(choices, utils.Choice{Weight: 50, Item: ActionCall}) //下注
			// choices = append(choices, utils.Choice{Weight: 50, Item: ActionPack}) //弃牌
			// c, _ := utils.WeightedChoice(choices)
			// return c.Item.(int)
			return ActionPack

		} else if msg.Condition == 2 || msg.Condition == 3 {
			// 同意比牌
			if msg.State&int32(pb.ACT_REPLY_BI) != 0 {
				return ActionAgreeBi
			}
			//看牌
			if !r.see {
				if utils.RandWan(5000) {
					return ActionSee
				}
			}

			//100%弃牌
			return ActionPack
		}
	}
	return ActionPack
}

// 主机僚机操作
func (r *RoleActor) jokerIdentityAction(msg *pb.JOKERPushActStateNtf) int {
	if r.identity == "a" || r.identity == "b" {
		return r.jokerIdentityABAction(msg)
	} else if r.identity == "c" || r.identity == "d" {
		return r.jokerIdentityCDAction(msg)
	}
	return ActionPack
}

// 看后即弃
func (r *RoleActor) jokerSeeAndPackAction(state int32, turn int32) int {
	//拒绝比牌
	if state&int32(pb.ACT_REPLY_BI) != 0 {
		return ActionRejectBi
	}

	//先看牌
	if !r.see {
		return ActionSee
	}
	//看后即弃
	return ActionPack
}

// 核心人机操作
func (r *RoleActor) jokerCoreAction(state int32, turn int32) int {
	//回复比牌
	if state&int32(pb.ACT_REPLY_BI) != 0 {
		if turn < 5 {
			return ActionRejectBi
		} else { //第6轮开始后有30%概率同意下家比牌
			agree := utils.RandWan(3000)
			if agree {
				return ActionAgreeBi
			} else {
				return ActionRejectBi
			}
		}
	}

	//看牌
	if !r.see {
		if utils.RandWan(6000) {
			return ActionSee
		}
	}

	var choices []utils.Choice
	choices = append(choices, utils.Choice{Weight: 40, Item: ActionCall})  //下注
	choices = append(choices, utils.Choice{Weight: 50, Item: ActionRaise}) //加注

	if state&int32(pb.ACT_SIDESHOW) != 0 || state&int32(pb.ACT_SHOW) != 0 {
		if algo.HuaType(r.cards) != algo.BaoZi && turn >= 10 && utils.RandWan(1000) {
			choices = append(choices, utils.Choice{Weight: 10, Item: ActionBi})
		}
	}
	c, _ := utils.WeightedChoice(choices)
	return c.Item.(int)
}

// 权重操作
func (r *RoleActor) jokerWeightAction(msg *pb.JOKERPushActStateNtf) int {
	if r.jokerstrategy == nil {
		return ActionPack
	}

	getWeightAction := func() int {
		// 回复比牌
		if msg.State&int32(pb.ACT_REPLY_BI) != 0 {
			agree := utils.RandWan(r.jokerstrategy.AgreeBi)
			if agree {
				return ActionAgreeBi
			} else {
				return ActionRejectBi
			}
		}

		// 看牌
		if !r.see {
			if utils.RandWan(r.jokerstrategy.SeeWeight[0]) {
				return ActionSee
			}
		}

		// 弃牌，比牌，跟注，加注
		index := int(msg.Turn)
		if index >= 0 && index < len(r.jokerstrategy.ActionWeight) {
			aw := r.jokerstrategy.ActionWeight[index]

			var choices []utils.Choice
			choices = append(choices, utils.Choice{Weight: int(aw.Values[0]), Item: ActionPack})
			choices = append(choices, utils.Choice{Weight: int(aw.Values[2]), Item: ActionCall})
			choices = append(choices, utils.Choice{Weight: int(aw.Values[3]), Item: ActionRaise})

			if msg.State&int32(pb.ACT_SIDESHOW) != 0 || msg.State&int32(pb.ACT_SHOW) != 0 {
				choices = append(choices, utils.Choice{Weight: int(aw.Values[1]), Item: ActionBi})
			}
			c, _ := utils.WeightedChoice(choices)
			return c.Item.(int)

		} else {
			return ActionPack
		}
	}
	action := getWeightAction()
	if action == ActionPack {
		if msg.IsChaalOper && msg.IsPlayerBlind && msg.IsOnlyTwo && msg.IsAbovePair {
			var choices []utils.Choice
			choices = append(choices, utils.Choice{Weight: 50, Item: ActionCall})
			if msg.State&int32(pb.ACT_SIDESHOW) != 0 || msg.State&int32(pb.ACT_SHOW) != 0 {
				choices = append(choices, utils.Choice{Weight: 50, Item: ActionBi})
			}
			c, _ := utils.WeightedChoice(choices)
			action = c.Item.(int)
		}
	}
	return action
}

func (r *RoleActor) jokerRunAction(action int) {
	r.prevAction = action
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

// 离开
func (r *RoleActor) JOKERLeaveRsp(ctx actor.Context) {
	msg := ctx.Message().(*pb.JOKERLeaveRsp)
	glog.Debugf("AK47LeaveRsp %v", msg)
	if msg.Userid == r.Userid {
		// 关闭节点
		stop := new(pb.ServeStop)
		r.pid.Tell(stop)
	}
}

// 离开
func (r *RoleActor) JOKERLeaveNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.JOKERLeaveNtf)
	glog.Debugf("JOKERLeaveNtf %v", msg)
	if msg.Userid == r.Userid {
		// 关闭节点
		stop := new(pb.ServeStop)
		r.pid.Tell(stop)
	}
}

func (a *RoleActor) JOKERCoinEnterRoomRsp(ctx actor.Context) {
	res := ctx.Message().(*pb.JOKERCoinEnterRoomRsp)
	glog.Debugf("JOKERCoinEnterRoomRsp %v", res)
	if res.Error != pb.OK {
		glog.Infof("JOKERCoinEnterRoomRsp enter err:%s", res.Error.String())
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

// 等待太久
func (r *RoleActor) JOKERCoinWaitTooLongNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.JOKERCoinWaitTooLongNtf)
	r.EmojiWaitLong(msg.Seat)
}

// 看牌推送
func (r *RoleActor) JOKERCoinSeeNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.JOKERCoinSeeNtf)
	if msg.Seat != r.seat {
		r.EmojiOtherSee(msg.Seat)
	}
}

// 看牌响应
func (r *RoleActor) JOKERCoinSeeRsp(ctx actor.Context) {
	msg := ctx.Message().(*pb.JOKERCoinSeeRsp)
	r.see = true
	r.cards = msg.Cards
}

// 加注推送
func (r *RoleActor) JOKERCoinRaiseNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.JOKERCoinRaiseNtf)
	if msg.Info.Seat != r.seat {
		r.EmojiOtherRaise(msg.Info.Seat)
	}
}

// 核心人机
func (r *RoleActor) jokerRunCore(msg *pb.JOKERPushActStateNtf) {
	action := r.jokerCoreAction(msg.State, msg.Turn)
	r.jokerRunAction(action)

	if action == ActionSee {
		action := r.jokerCoreAction(msg.State, msg.Turn)
		r.jokerRunAction(action)
	}
}

// 看后即弃人机
func (r *RoleActor) jokerRunSeeAndPack(msg *pb.JOKERPushActStateNtf) {
	action := r.jokerSeeAndPackAction(msg.State, msg.Turn)
	r.jokerRunAction(action)

	if action == ActionSee {
		action := r.jokerSeeAndPackAction(msg.State, msg.Turn)
		r.jokerRunAction(action)
	}
}

// 主机僚机
func (r *RoleActor) jokerRunIdentity(msg *pb.JOKERPushActStateNtf) {
	action := r.jokerIdentityAction(msg)
	r.jokerRunAction(action)

	if action == ActionSee {
		action := r.jokerIdentityAction(msg)
		r.jokerRunAction(action)
	}
}

// 默认人机
func (r *RoleActor) jokerRunDefalut(msg *pb.JOKERPushActStateNtf) {
	action := r.jokerWeightAction(msg)
	r.jokerRunAction(action)

	if action == ActionSee { //如果是看牌，还需要执行一次操作
		action := r.jokerWeightAction(msg)
		r.jokerRunAction(action)
	}
}

// 动作状态推送
func (r *RoleActor) recvJOKERPushActStateNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.JOKERPushActStateNtf)
	if r.alive && msg.Seat == r.seat { //自己操作
		if r.core {
			r.jokerRunCore(msg)
		} else if r.seeAndPack {
			r.jokerRunSeeAndPack(msg)
		} else if r.identity != "" {
			r.jokerRunIdentity(msg)
		} else {
			r.jokerRunDefalut(msg)
		}
	} else { //其他人操作
		if r.alive && !r.see {
			if r.core {
				if utils.RandWan(3000) {
					r.sendJOKERCoinSeeReq()
				}
			} else if r.seeAndPack {
			} else if r.identity != "" {
			} else {
				if r.jokerstrategy != nil {
					if utils.RandWan(r.jokerstrategy.SeeWeight[1]) {
						r.sendJOKERCoinSeeReq()
					}
				}
			}
		}
	}
}

// 比牌结果推送
func (r *RoleActor) recvJOKERCoinBiNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.JOKERCoinBiNtf)
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
func (r *RoleActor) JOKERCoinRobotStrategyNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.JOKERCoinRobotStrategyNtf)
	r.jokerstrategy = msg
	r.core = msg.Core
	r.seeAndPack = msg.SeeAndPack
	r.identity = msg.Identity

	if msg.BetChange {
		r.cards = msg.ChangeCards
	}

	glog.Infof("robot strategy %#v", msg)
}

// 游戏结束
func (r *RoleActor) JOKERCoinGameoverNtf(ctx actor.Context) {
	if r.chargeInGame {
		// 假装局内充值了,回合结束离开
		glog.Infof("robot charge in game then leave: %v,%v,%v,%v", r.gtype, r.gameId, r.roomId, r.Userid)
		r.sendJOKERLeaveReq()
		return
	}
	if utils.RandWan(5000) {
		r.sendJOKERLeaveReq()
	}
}

// 准备请求
func (c *RoleActor) sendJOKERReady2Req() {
	c2s := new(pb.JOKERReady2Req)
	c.Sender(c2s)
}

// 看牌请求
func (c *RoleActor) sendJOKERCoinSeeReq() {
	c2s := new(pb.JOKERCoinSeeReq)
	c.SendDefer(c2s)
	c.see = true
}

// 跟注请求
func (c *RoleActor) sendJOKERCoinCallReq() {
	c2s := new(pb.JOKERCoinCallReq)
	c.SendDefer(c2s)
}

// 加注请求
func (c *RoleActor) sendJOKERCoinRaiseReq() {
	c2s := new(pb.JOKERCoinRaiseReq)
	c.SendDefer(c2s)
}

// 弃牌请求
func (c *RoleActor) sendJOKERCoinFoldReq() {
	c.alive = false
	if utils.RandWan(100) {
		return
	}
	c2s := new(pb.JOKERCoinFoldReq)
	c.SendDefer(c2s)
}

// 比牌请求
func (c *RoleActor) sendJOKERCoinBiReq() {
	c2s := new(pb.JOKERCoinBiReq)
	c.SendDefer(c2s)
}

// 回复比牌请求
func (c *RoleActor) sendJOKERCoinReplyBiReq(agress bool) {
	c2s := new(pb.JOKERCoinReplyBiReq)
	c2s.Agree = agress
	c.SendDefer(c2s)
}

// 机器人离开
func (c *RoleActor) sendJOKERLeaveReq() {
	c2s := new(pb.JOKERLeaveReq)
	c.SendDefer(c2s)
}
