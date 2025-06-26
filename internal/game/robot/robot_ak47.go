package robot

import (
	"goserver/gen/pb"
	"goserver/pkg/game/algo"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 进入或匹配桌子
func (rs *RoleActor) enterAK47MatchDesk(arg *pb.RobotMsg) *pb.MatchDesk {
	//获取游戏服节点或者房间进程
	msg := new(pb.MatchDesk)
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.ak47").Name()
	msg.Gtype = int32(pb.AK47) //
	msg.Roomid = arg.Roomid
	msg.Gameid = arg.Gameid
	return msg
}

// 进入房间响应
func (r *RoleActor) recvAK47CoinEnterRoomRsp(ctx actor.Context) {
	s2c := ctx.Message().(*pb.AK47CoinEnterRoomRsp)
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
func (r *RoleActor) recvAK47PushStateNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.AK47PushStateNtf)
	if msg.State == int32(pb.STATE_DEALING) {
		r.alive = true
		r.sendAK47Ready2Req()
	} else if msg.State == int32(pb.STATE_BET) { //下注
		r.EmojiGameStart(r.GetRandSeat())
	} else if msg.State == int32(pb.STATE_OVER) { //游戏结束
		r.see = false
		r.ak47strategy = nil
		r.cards = []uint32{}
		r.ccards = []uint32{}
		r.state = 0
		r.core = false
		r.seeAndPack = false
		r.identity = ""
	}
}

func (r *RoleActor) ak47SimAction(state int32) int {
	if !r.see {
		return ActionSee
	}

	if len(r.cards) == 0 {
		return ActionPack
	}

	if state&int32(pb.ACT_REPLY_BI) != 0 {
		if algo.HuaType(r.ccards) >= algo.TongHua {
			return ActionAgreeBi
		} else {
			return ActionRejectBi
		}
	}

	if algo.HuaType(r.ccards) >= algo.TongHua {
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

func (r *RoleActor) ak47IdentityABAction(msg *pb.AK47PushActStateNtf) int {
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

func (r *RoleActor) ak47IdentityCDAction(msg *pb.AK47PushActStateNtf) int {
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
func (r *RoleActor) ak47IdentityAction(msg *pb.AK47PushActStateNtf) int {
	if r.identity == "a" || r.identity == "b" {
		return r.ak47IdentityABAction(msg)
	} else if r.identity == "c" || r.identity == "d" {
		return r.ak47IdentityCDAction(msg)
	}
	return ActionPack
}

// 看后即弃
func (r *RoleActor) ak47SeeAndPackAction(state int32, turn int32) int {
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
func (r *RoleActor) ak47CoreAction(state int32, turn int32) int {
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
		if algo.HuaType(r.ccards) != algo.BaoZi && turn >= 10 && utils.RandWan(1000) {
			choices = append(choices, utils.Choice{Weight: 10, Item: ActionBi})
		}
	}
	c, _ := utils.WeightedChoice(choices)
	return c.Item.(int)
}

// 权重操作
func (r *RoleActor) ak47WeightAction(msg *pb.AK47PushActStateNtf) int {
	if r.ak47strategy == nil {
		return ActionPack
	}

	getWeightAction := func() int {
		// 回复比牌
		if msg.State&int32(pb.ACT_REPLY_BI) != 0 {
			agree := utils.RandWan(r.ak47strategy.AgreeBi)
			if agree {
				return ActionAgreeBi
			} else {
				return ActionRejectBi
			}
		}

		// 看牌
		if !r.see {
			if utils.RandWan(r.ak47strategy.SeeWeight[0]) {
				return ActionSee
			}
		}

		// 弃牌，比牌，跟注，加注
		index := int(msg.Turn)
		if index >= 0 && index < len(r.ak47strategy.ActionWeight) {
			aw := r.ak47strategy.ActionWeight[index]

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

func (r *RoleActor) ak47RunAction(action int) {
	r.prevAction = action
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

// 离开
func (r *RoleActor) AK47LeaveRsp(ctx actor.Context) {
	ntf := ctx.Message().(*pb.AK47LeaveRsp)
	glog.Debugf("AK47LeaveRsp %v", ntf)
	if ntf.Userid == r.Userid {
		// 关闭节点
		stop := new(pb.ServeStop)
		r.pid.Tell(stop)
	}
}

// 离开
func (r *RoleActor) AK47LeaveNtf(ctx actor.Context) {
	ntf := ctx.Message().(*pb.AK47LeaveNtf)
	glog.Debugf("AK47LeaveNtf %v", ntf)
	if ntf.Userid == r.Userid {
		// 关闭节点
		stop := new(pb.ServeStop)
		r.pid.Tell(stop)
	}
}

func (a *RoleActor) AK47CoinEnterRoomRsp(ctx actor.Context) {
	res := ctx.Message().(*pb.AK47CoinEnterRoomRsp)
	glog.Debugf("AK47CoinEnterRoomRsp %v", res)
	if res.Error != pb.OK {
		glog.Infof("AK47CoinEnterRoomRsp enter err:%s", res.Error.String())
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

// 等待太久
func (r *RoleActor) AK47CoinWaitTooLongNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.AK47CoinWaitTooLongNtf)
	r.EmojiWaitLong(msg.Seat)
}

// 看牌推送
func (r *RoleActor) AK47CoinSeeNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.AK47CoinSeeNtf)
	if msg.Seat != r.seat {
		r.EmojiOtherSee(msg.Seat)
	}
}

// 看牌响应
func (r *RoleActor) AK47CoinSeeRsp(ctx actor.Context) {
	msg := ctx.Message().(*pb.AK47CoinSeeRsp)
	r.see = true
	r.cards = msg.Cards
	r.ccards = msg.Changecards
}

// 加注推送
func (r *RoleActor) AK47CoinRaiseNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.AK47CoinRaiseNtf)
	if msg.Info.Seat != r.seat {
		r.EmojiOtherRaise(msg.Info.Seat)
	}
}

// 核心人机
func (r *RoleActor) ak47RunCore(msg *pb.AK47PushActStateNtf) {
	action := r.ak47CoreAction(msg.State, msg.Turn)
	r.ak47RunAction(action)

	if action == ActionSee {
		action := r.ak47CoreAction(msg.State, msg.Turn)
		r.ak47RunAction(action)
	}
}

// 看后即弃人机
func (r *RoleActor) ak47RunSeeAndPack(msg *pb.AK47PushActStateNtf) {
	action := r.ak47SeeAndPackAction(msg.State, msg.Turn)
	r.ak47RunAction(action)

	if action == ActionSee {
		action := r.ak47SeeAndPackAction(msg.State, msg.Turn)
		r.ak47RunAction(action)
	}
}

// 主机僚机
func (r *RoleActor) ak47RunIdentity(msg *pb.AK47PushActStateNtf) {
	action := r.ak47IdentityAction(msg)
	r.ak47RunAction(action)

	if action == ActionSee {
		action := r.ak47IdentityAction(msg)
		r.ak47RunAction(action)
	}
}

// 默认人机
func (r *RoleActor) ak47RunDefalut(msg *pb.AK47PushActStateNtf) {
	action := r.ak47WeightAction(msg)
	r.ak47RunAction(action)

	if action == ActionSee { //如果是看牌，还需要执行一次操作
		action := r.ak47WeightAction(msg)
		r.ak47RunAction(action)
	}
}

// 动作状态推送
func (r *RoleActor) recvAK47PushActStateNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.AK47PushActStateNtf)
	if r.alive && msg.Seat == r.seat { //自己操作
		if r.core { //核心人机
			r.ak47RunCore(msg)
		} else if r.seeAndPack { //看后即弃
			r.ak47RunSeeAndPack(msg)
		} else if r.identity != "" { //主机僚机
			r.ak47RunIdentity(msg)
		} else { //默认
			r.ak47RunDefalut(msg)
		}
	} else { //其他人操作
		if r.alive && !r.see {
			if r.core {
				if utils.RandWan(3000) {
					r.sendAK47CoinSeeReq()
				}
			} else if r.seeAndPack {
			} else if r.identity != "" {
			} else {
				if r.ak47strategy != nil {
					if utils.RandWan(r.ak47strategy.SeeWeight[1]) {
						r.sendAK47CoinSeeReq()
					}
				}
			}
		}
	}
}

// 比牌结果推送
func (r *RoleActor) recvAK47CoinBiNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.AK47CoinBiNtf)
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
func (r *RoleActor) AK47CoinRobotStrategyNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.AK47CoinRobotStrategyNtf)
	r.ak47strategy = msg
	r.core = msg.Core
	r.seeAndPack = msg.SeeAndPack
	r.identity = msg.Identity

	if msg.BetChange {
		r.cards = msg.ChangeCards
		r.ccards = msg.ChangeCcards
	}

	glog.Infof("robot strategy %#v", msg)
}

// 游戏结束
func (r *RoleActor) AK47CoinGameoverNtf(ctx actor.Context) {
	if r.chargeInGame {
		// 假装局内充值了,回合结束离开
		glog.Infof("robot charge in game then leave: %v,%v,%v,%v", r.gtype, r.gameId, r.roomId, r.Userid)
		r.sendAK47LeaveReq()
		return
	}
	if utils.RandWan(5000) {
		r.sendAK47LeaveReq()
	}
}

// 准备请求
func (c *RoleActor) sendAK47Ready2Req() {
	c2s := new(pb.AK47Ready2Req)
	c.Sender(c2s)
}

// 看牌请求
func (c *RoleActor) sendAK47CoinSeeReq() {
	c2s := new(pb.AK47CoinSeeReq)
	c.SendDefer(c2s)
	c.see = true
}

// 跟注请求
func (c *RoleActor) sendAK47CoinCallReq() {
	c2s := new(pb.AK47CoinCallReq)
	c.SendDefer(c2s)
}

// 加注请求
func (c *RoleActor) sendAK47CoinRaiseReq() {
	c2s := new(pb.AK47CoinRaiseReq)
	c.SendDefer(c2s)
}

// 弃牌请求
func (c *RoleActor) sendAK47CoinFoldReq() {
	c.alive = false
	if utils.RandWan(100) {
		return
	}
	c2s := new(pb.AK47CoinFoldReq)
	c.SendDefer(c2s)
}

// 比牌请求
func (c *RoleActor) sendAK47CoinBiReq() {
	c2s := new(pb.AK47CoinBiReq)
	c.SendDefer(c2s)
}

// 回复比牌请求
func (c *RoleActor) sendAK47CoinReplyBiReq(agress bool) {
	c2s := new(pb.AK47CoinReplyBiReq)
	c2s.Agree = agress
	c.SendDefer(c2s)
}

// 机器人离开
func (c *RoleActor) sendAK47LeaveReq() {
	c2s := new(pb.AK47LeaveReq)
	c.SendDefer(c2s)
}
