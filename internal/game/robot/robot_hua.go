package robot

import (
	"goserver/gen/pb"
	"goserver/pkg/game/algo"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	"github.com/AsynkronIT/protoactor-go/actor"
)

const (
	ActionNull     int = iota
	ActionSee          //看牌
	ActionPack         //弃牌
	ActionCall         //跟注
	ActionRaise        //加注
	ActionBi           //发起比牌
	ActionReplyBi      //回复比牌
	ActionAgreeBi      //同意比牌
	ActionRejectBi     //拒绝比牌
)

// 准备请求
func (r *RoleActor) sendJHReady2Req() {
	c2s := new(pb.JHReady2Req)
	r.Sender(c2s)
}

// 看牌请求
func (c *RoleActor) sendJHCoinSeeReq(delay ...int64) {
	c.see = true
	c2s := new(pb.JHCoinSeeReq)
	if len(delay) == 0 {
		c.SendDefer(c2s)
	} else {
		c.SendDefer2(c2s, delay[0])
	}
}

// 跟注请求
func (c *RoleActor) sendJHCoinCallReq(delay ...int64) {
	c2s := new(pb.JHCoinCallReq)
	if len(delay) == 0 {
		c.SendDefer(c2s)
	} else {
		c.SendDefer2(c2s, delay[0])
	}
}

// 加注请求
func (c *RoleActor) sendJHCoinRaiseReq(delay ...int64) {
	c2s := new(pb.JHCoinRaiseReq)
	if len(delay) == 0 {
		c.SendDefer(c2s)
	} else {
		c.SendDefer2(c2s, delay[0])
	}
}

// 弃牌请求
func (c *RoleActor) sendJHCoinFoldReq() {
	c.alive = false
	if utils.RandWan(100) {
		return
	}
	c2s := new(pb.JHCoinFoldReq)
	c.SendDefer(c2s)
}

// 弃牌请求
func (c *RoleActor) sendJHCoinFoldReq1(delay ...int64) {
	c.alive = false
	c2s := new(pb.JHCoinFoldReq)
	if len(delay) == 0 {
		c.SendDefer(c2s)
	} else {
		c.SendDefer2(c2s, delay[0])
	}
}

// 比牌请求
func (c *RoleActor) sendJHCoinBiReq(delay ...int64) {
	c2s := new(pb.JHCoinBiReq)
	if len(delay) == 0 {
		c.SendDefer(c2s)
	} else {
		c.SendDefer2(c2s, delay[0])
	}
}

// 回复比牌请求
func (c *RoleActor) sendJHCoinReplyBiReq(agress bool, delay ...int64) {
	c2s := new(pb.JHCoinReplyBiReq)
	c2s.Agree = agress
	if len(delay) == 0 {
		c.SendDefer(c2s)
	} else {
		c.SendDefer2(c2s, delay[0])
	}
}

// 机器人离开
func (c *RoleActor) sendJHLeaveReq(delay ...int64) {
	c2s := new(pb.JHLeaveReq)
	if len(delay) == 0 {
		c.SendDefer(c2s)
	} else {
		c.SendDefer2(c2s, delay[0])
	}
}

// 进入或匹配桌子
func (rs *RoleActor) enterJHMatchDesk(arg *pb.RobotMsg) *pb.MatchDesk {
	//获取游戏服节点或者房间进程
	msg := new(pb.MatchDesk)
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.hua").Name()
	msg.Gtype = int32(pb.HUA) //金花
	msg.Roomid = arg.Roomid
	msg.Gameid = arg.Gameid
	return msg
}

func (rs *RoleActor) JHLeaveNtf(ctx actor.Context) {
	ntf := ctx.Message().(*pb.JHLeaveNtf)
	if ntf.Userid == rs.Userid {
		// 关闭节点
		glog.Debugf("JHLeaveNtf %v", ntf)
		stop := new(pb.ServeStop)
		rs.pid.Tell(stop)
	}
}

func (rs *RoleActor) JHLeaveRsp(ctx actor.Context) {
	ntf := ctx.Message().(*pb.JHLeaveRsp)
	if ntf.Userid == rs.Userid {
		// 关闭节点
		glog.Debugf("JHLeaveRsp %v", ntf)
		stop := new(pb.ServeStop)
		rs.pid.Tell(stop)
	}
}

func (a *RoleActor) JHCoinEnterRoomRsp(ctx actor.Context) {
	res := ctx.Message().(*pb.JHCoinEnterRoomRsp)
	glog.Debugf("JHCoinEnterRoomRsp %v", res)
	if res.Error != pb.OK {
		glog.Infof("JHCoinEnterRoomRsp enter err:%s", res.Error.String())
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

// 房间状态推送
func (r *RoleActor) recvJHPushStateNtf(ctx actor.Context) {
	switch r.gtype {
	case int32(pb.HUA2):
		r.recvJHPushStateNtf2(ctx)
	default:
		r.recvJHPushStateNtf1(ctx)
	}
}

// 房间状态推送
func (r *RoleActor) recvJHPushStateNtf1(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHPushStateNtf)
	glog.Debugf("JHPushStateNtf %#v", msg)
	if msg.State == int32(pb.STATE_DEALING) { //发牌
		r.alive = true
		r.sendJHReady2Req()
	} else if msg.State == int32(pb.STATE_BET) { //下注
		r.EmojiGameStart(r.GetRandSeat())
	} else if msg.State == int32(pb.STATE_OVER) { //游戏结束
		r.see = false
		r.jhstrategy = nil
		r.cards = []uint32{}
		r.state = 0
		r.core = false
		r.seeAndPack = false
		r.identity = ""
	}
}

// 核心人机
func (r *RoleActor) huaRunCore(msg *pb.JHPushActStateNtf) {
	action := r.huaCoreAction(msg.State, msg.Turn)
	r.huaRunAction(action)

	if action == ActionSee {
		action := r.huaCoreAction(msg.State, msg.Turn)
		r.huaRunAction(action)
	}
}

// 看后即弃人机
func (r *RoleActor) huaRunSeeAndPack(msg *pb.JHPushActStateNtf) {
	action := r.huaSeeAndPackAction(msg.State, msg.Turn)
	r.huaRunAction(action)

	if action == ActionSee {
		action := r.huaSeeAndPackAction(msg.State, msg.Turn)
		r.huaRunAction(action)
	}
}

// 主机僚机
func (r *RoleActor) huaRunIdentity(msg *pb.JHPushActStateNtf) {
	action := r.huaIdentityAction(msg)
	r.huaRunAction(action)

	if action == ActionSee {
		action := r.huaIdentityAction(msg)
		r.huaRunAction(action)
	}
}

// 默认人机
func (r *RoleActor) huaRunDefalut(msg *pb.JHPushActStateNtf) {
	action := r.huaWeightAction(msg)
	r.huaRunAction(action)

	if action == ActionSee { //如果是看牌，还需要执行一次操作
		action := r.huaWeightAction(msg)
		r.huaRunAction(action)
	}
}

// 动作状态推送
func (r *RoleActor) recvJHPushActStateNtf(ctx actor.Context) {
	switch r.gtype {
	case int32(pb.HUA2):
		r.recvJHPushActStateNtf2(ctx)
	default:
		r.recvJHPushActStateNtf1(ctx)
	}
}

// 动作状态推送
func (r *RoleActor) recvJHPushActStateNtf1(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHPushActStateNtf)
	if r.alive && msg.Seat == r.seat { //自己操作
		r.cancelSeeingInOther() // 取消其他玩家回合看牌操作

		// 局内充值后操作
		if msg.RAction.ActionCharge {
			r.prevRAction = msg.RAction
			return
		}

		// 执行动作
		r.huaRunRAction(msg.RAction)
	}
}

// 动作状态推送
// func (r *RoleActor) recvJHPushActStateNtf1(ctx actor.Context) {
// 	msg := ctx.Message().(*pb.JHPushActStateNtf)
// 	if r.alive && msg.Seat == r.seat { //自己操作
// 		if r.core { //核心人机
// 			r.huaRunCore(msg)
// 		} else if r.seeAndPack { //看后即弃
// 			r.huaRunSeeAndPack(msg)
// 		} else if r.identity != "" { //主机僚机
// 			r.huaRunIdentity(msg)
// 		} else { //默认
// 			r.huaRunDefalut(msg)
// 		}
// 	} else { //其他人操作
// 		if r.alive && !r.see {
// 			if r.core {
// 				if utils.RandWan(3000) {
// 					r.sendJHCoinSeeReq()
// 				}
// 			} else if r.seeAndPack {
// 			} else if r.identity != "" {
// 			} else {
// 				if r.jhstrategy != nil {
// 					if utils.RandWan(r.jhstrategy.SeeWeight[1]) {
// 						r.sendJHCoinSeeReq()
// 					}
// 				}
// 			}
// 		}
// 	}
// }

// 人机策略推送
func (r *RoleActor) JHCoinRobotStrategyNtf(ctx actor.Context) {
	switch r.gtype {
	case int32(pb.HUA2):
		r.JHCoinRobotStrategyNtf2(ctx)
	default:
		r.JHCoinRobotStrategyNtf1(ctx)
	}
}

// 人机策略推送
func (r *RoleActor) JHCoinRobotStrategyNtf1(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHCoinRobotStrategyNtf)
	r.jhstrategy = msg
	r.core = msg.Core
	r.seeAndPack = msg.SeeAndPack
	r.identity = msg.Identity

	//局中换牌重置牌
	if msg.BetChange {
		r.cards = msg.ChangeCards
	}
	// glog.Infof("robot strategy %#v", msg)
}

// 游戏结束
func (r *RoleActor) JHCoinGameoverNtf(ctx actor.Context) {
	switch r.gtype {
	case int32(pb.HUA2):
		r.JHCoinGameoverNtf2(ctx)
	default:
		r.JHCoinGameoverNtf1(ctx)
	}
}

// 游戏结束, 等RobotLeaveNtf通知离开
func (r *RoleActor) JHCoinGameoverNtf1(ctx actor.Context) {
	// msg := ctx.Message().(*pb.JHCoinGameoverNtf)
	// for i, seatid := range msg.Lr {
	// 	if r.seat == seatid {
	// 		// 人机离开
	// 		r.sendJHLeaveReq(int64(msg.Lrms[i]))
	// 	}
	// }
}

// 看牌推送
func (r *RoleActor) JHCoinSeeNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHCoinSeeNtf)
	if msg.Seat != r.seat {
		r.EmojiOtherSee(msg.Seat)
	}
}

// 看牌响应
func (r *RoleActor) JHCoinSeeRsp(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHCoinSeeRsp)
	r.see = true
	r.cards = msg.Cards
}

// 如果延迟时长到了之后，处于自己操作回合，则原定要执行see的操作取消
func (r *RoleActor) cancelSeeingInOther() {
	r.seeingInOther.Store(false)
}

// 通知人机看牌
func (r *RoleActor) JHRobotSeeInOtherNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHRobotSeeInOtherNtf)
	if msg.Seat == r.seat {
		if !r.seeingInOther.CompareAndSwap(false, true) { // 正在看牌中...
			return
		}
		// 发送看牌请求
		c2s := new(pb.JHCoinSeeReq)
		delayCall := func() bool {
			defer r.seeingInOther.Store(false)
			if r.see { // 已经看牌
				return false
			}
			if !r.seeingInOther.Load() { // 被取消
				return false
			}
			r.see = true
			return true
		}
		r.SendDefer3(c2s, int64(msg.Delay), delayCall)
	}
}

// 其他玩家跟注
func (r *RoleActor) JHCoinCallNtf(ctx actor.Context) {
	// msg := ctx.Message().(*pb.JHCoinCallNtf)
	// if msg.Info.Seat != r.seat {

	// }
}

// 其他玩家加注
func (r *RoleActor) JHCoinRaiseNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHCoinRaiseNtf)
	if msg.Info.Seat != r.seat {
		r.EmojiOtherRaise(msg.Info.Seat)
	}
}

// 比牌结果推送
func (r *RoleActor) recvJHCoinBiNtf(ctx actor.Context) {
	switch r.gtype {
	case int32(pb.HUA2):
		r.recvJHCoinBiNtf2(ctx)
	default:
		r.recvJHCoinBiNtf1(ctx)
	}
}

// 比牌结果推送
func (r *RoleActor) recvJHCoinBiNtf1(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHCoinBiNtf)
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

func (r *RoleActor) recvJHCoinWaitTooLongNtf(ctx actor.Context) {
	switch r.gtype {
	case int32(pb.HUA2):
		r.recvJHCoinWaitTooLongNtf2(ctx)
	default:
		r.recvJHCoinWaitTooLongNtf1(ctx)
	}
}

func (r *RoleActor) recvJHCoinWaitTooLongNtf1(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHCoinWaitTooLongNtf)
	r.EmojiWaitLong(msg.Seat)
}

func (r *RoleActor) recvJHCoinEnterRoomRsp(ctx actor.Context) {
	switch r.gtype {
	case int32(pb.HUA2):
		r.recvJHCoinEnterRoomRsp2(ctx)
	default:
		r.recvJHCoinEnterRoomRsp1(ctx)
	}
}

func (r *RoleActor) recvJHCoinEnterRoomRsp1(ctx actor.Context) {
	s2c := ctx.Message().(*pb.JHCoinEnterRoomRsp)
	var errcode = s2c.GetError()
	switch errcode {
	case pb.OK:
	default:
		// 关闭节点
		glog.Errorf("enter hua fail user:%s,err:%v ", r.Userid, errcode)
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

func (r *RoleActor) huaIdentityABAction(msg *pb.JHPushActStateNtf) int {
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

func (r *RoleActor) huaIdentityCDAction(msg *pb.JHPushActStateNtf) int {
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
func (r *RoleActor) huaIdentityAction(msg *pb.JHPushActStateNtf) int {
	if r.identity == "a" || r.identity == "b" {
		return r.huaIdentityABAction(msg)
	} else if r.identity == "c" || r.identity == "d" {
		return r.huaIdentityCDAction(msg)
	}
	return ActionPack
}

// 看后即弃
func (r *RoleActor) huaSeeAndPackAction(state int32, turn int32) int {
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
func (r *RoleActor) huaCoreAction(state int32, turn int32) int {
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

// 权重操作(默认)
func (r *RoleActor) huaWeightAction(msg *pb.JHPushActStateNtf) int {
	if r.jhstrategy == nil {
		return ActionPack
	}

	getWeightAction := func() int {
		// 回复比牌
		if msg.State&int32(pb.ACT_REPLY_BI) != 0 {
			agree := utils.RandWan(r.jhstrategy.AgreeBi)
			if agree {
				return ActionAgreeBi
			} else {
				return ActionRejectBi
			}
		}

		// 看牌
		if !r.see {
			if utils.RandWan(r.jhstrategy.SeeWeight[0]) {
				return ActionSee
			}
		}

		// 弃牌，比牌，跟注，加注
		index := int(msg.Turn)
		if index >= len(r.jhstrategy.ActionWeight) {
			index = len(r.jhstrategy.ActionWeight) - 1
		}

		if index >= 0 && index < len(r.jhstrategy.ActionWeight) {
			aw := r.jhstrategy.ActionWeight[index]

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

func (r *RoleActor) huaRunRAction(rAction *pb.RAction) {
	if rAction.ActionSee { // 看牌后操作
		r.huaRunAction1(ActionSee, int64(rAction.SeeDelay))
		if rAction.Action != int32(ActionNull) {
			r.huaRunAction1(int(rAction.Action), int64(rAction.Delay)) // +rAction.SeeDelay
		}
	} else {
		r.huaRunAction1(int(rAction.Action), int64(rAction.Delay))
	}
}

func (r *RoleActor) huaRunAction(action int) {
	r.prevAction = action
	switch action {
	case ActionSee:
		r.sendJHCoinSeeReq()
	case ActionPack:
		r.sendJHCoinFoldReq()
	case ActionCall:
		r.sendJHCoinCallReq()
	case ActionRaise:
		r.sendJHCoinRaiseReq()
	case ActionBi:
		r.sendJHCoinBiReq()
	case ActionReplyBi:
		r.sendJHCoinReplyBiReq(utils.RandBool())
	case ActionAgreeBi:
		r.sendJHCoinReplyBiReq(true)
	case ActionRejectBi:
		r.sendJHCoinReplyBiReq(false)
	}
}

func (r *RoleActor) huaRunAction1(action int, delay int64) {
	r.prevAction = action
	switch action {
	case ActionSee:
		r.sendJHCoinSeeReq(delay)
	case ActionPack:
		r.sendJHCoinFoldReq1(delay)
	case ActionCall:
		r.sendJHCoinCallReq(delay)
	case ActionRaise:
		r.sendJHCoinRaiseReq(delay)
	case ActionBi:
		r.sendJHCoinBiReq(delay)
	case ActionReplyBi:
		r.sendJHCoinReplyBiReq(utils.RandBool())
	case ActionAgreeBi:
		r.sendJHCoinReplyBiReq(true, delay)
	case ActionRejectBi:
		r.sendJHCoinReplyBiReq(false, delay)
	}
}
