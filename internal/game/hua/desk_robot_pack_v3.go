package hua

import (
	"goserver/pkg/data"
	"goserver/pkg/game/algo"
	"goserver/pkg/utils"
)

// 触发极限倍数后逻辑
func (t *Desk) handleRobotOverflowMultipeV3(seatid uint32, seat *data.DeskSeat, stakeAction int32, trace *[]int16) (action int32, actionSee bool) {
	*trace = append(*trace, 3001)
	// 触发极限倍数直接进pack逻辑
	return t.handleRobotPackV3(seatid, seat, true, trace)
}

// pack逻辑
func (t *Desk) handleRobotPackV3(seatid uint32, seat *data.DeskSeat, isMaxMultiple bool, trace *[]int16) (action int32, actionSee bool) {
	if isMaxMultiple { // 是否极限倍数的pack
		*trace = append(*trace, 3101)
		// 是否第一轮
		return t.handleRobotPackFirstRoundV3(seatid, seat, trace)
	} else {
		if t.isPlayerPack() { // 玩家是否还在
			*trace = append(*trace, 3102)
			// 改走 stake 流程
			return t.handleRobotStake0(seatid, seat, trace)
		} else {
			if seat.BiggerThanPlayer || t.isPlayerPack() { // 比玩家大
				*trace = append(*trace, 3103)
				// 改走 stake 流程
				return t.handleRobotStake0(seatid, seat, trace)
			} else {
				// 本局赢后总返奖率是否大于等于风控基础返奖率
				if !t.isPlayerWinOverRiskControlRate() {
					*trace = append(*trace, 3104)
					return t.handleRobotStake0(seatid, seat, trace)
				} else {
					if t.DeskAct.ActTimes == 0 { // 是否第一轮
						*trace = append(*trace, 3105)
						return ActionPack, false
					} else {
						alive, _, _ := t.getAliveNum()
						if alive != 2 { // 当前是否只有2人
							*trace = append(*trace, 3106)
							return ActionPack, false
						} else {
							if !t.isBeforeChaal(seatid) { // 之前是否有过 chaal 和 doublechaal
								*trace = append(*trace, 3107)
								return ActionPack, false
							} else {
								_, ok := t.isNextPlayerBeforeChaal(seatid)
								if ok { // 判断对手之前是否chaal或doublechaal
									*trace = append(*trace, 3108)
									return ActionPack, false
								} else {
									if algo.HuaType(seat.Cards) < algo.DuiZi { // 是否对子以上
										*trace = append(*trace, 3109)
										return ActionPack, false
									} else {
										*trace = append(*trace, 3110)
										overMax := t.isPlayerWinOverRiskControlMaxRate()
										var changedCard bool
										if overMax {
											*trace = append(*trace, 3111)
											// 如换牌失败, 上一层为否继续判断
											changedCard = t.handleChangeRobotBigger(seatid, seat)
										}
										if overMax && changedCard {
											*trace = append(*trace, 3112)
											return t.handleRobotStake0(seatid, seat, trace)
										} else {
											*trace = append(*trace, 3113)
											return t.handleRobotStake0(seatid, seat, trace)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

// pack逻辑是否第一轮 ->
func (t *Desk) handleRobotPackFirstRoundV3(seatid uint32, seat *data.DeskSeat, trace *[]int16) (action int32, actionSee bool) {
	if t.DeskAct.ActTimes == 0 { // 是否第一轮
		*trace = append(*trace, 3201)
		if !t.isBiggerCards(seatid, false) { // 是否最大牌
			*trace = append(*trace, 3202)
			return ActionPack, false
		} else {
			// 本局赢后总返奖率是否大于等于风控基础返奖率
			if t.isPlayerWinOverRiskControlRate() {
				*trace = append(*trace, 3203)
				return t.handleRobotStake0(seatid, seat, trace)
			} else {
				*trace = append(*trace, 3204)
				return ActionPack, false
			}
		}
	} else {
		alive, _, _ := t.getAliveNum()
		if alive != 2 { // 当前是否只有2人
			*trace = append(*trace, 3205)
			if !t.isBiggerCards(seatid, false) { // 是否最大牌
				*trace = append(*trace, 3206)
				// return ActionPack, false
				if !(algo.HuaTypeUpOrDown(seat.Cards) <= algo.DuiziUp) { // 非大对子以上
					*trace = append(*trace, 3207)
					return ActionPack, false
				} else {
					// 判断之前是否show被拒绝过
					if t.isBeforeSideShowBeRejected(seatid, seat) {
						*trace = append(*trace, 3208)
						return ActionPack, false
					} else {
						*trace = append(*trace, 3209)
						return ActionBi, false
					}
				}
			} else {
				if algo.HuaTypeUpOrDown(seat.Cards) <= algo.DuiziUp { // 大对子及以上
					*trace = append(*trace, 3210)
					return t.handleRobotStake0(seatid, seat, trace)
				} else {
					if algo.HuaTypeUpOrDown(seat.Cards) == algo.DuiziDown { // 是小对子
						*trace = append(*trace, 3211)
						return t.handleRobotStake0(seatid, seat, trace)
					} else {
						// 本局赢后总返奖率是否大于等于风控基础返奖率
						if t.isPlayerWinOverRiskControlRate() {
							*trace = append(*trace, 3212)
							return t.handleRobotStake0(seatid, seat, trace)
						} else {
							*trace = append(*trace, 3213)
							return ActionPack, false
						}
					}
				}
			}
		} else {
			if !t.isBeforeChaal(seatid) { // 之前是否有过 chaal 和 doublechaal
				*trace = append(*trace, 3214)
				if !t.isBiggerCards(seatid, false) { // 是否此时全场最大
					*trace = append(*trace, 3215)
					if !(algo.HuaTypeUpOrDown(seat.Cards) <= algo.DuiziDown) { // 非小对子及以上
						*trace = append(*trace, 3216)
						return ActionPack, false
					} else {
						// 本局赢后总返奖率是否大于等于风控基础返奖率
						if t.isPlayerWinOverRiskControlRate() {
							*trace = append(*trace, 3217)
							return ActionPack, false
						} else {
							*trace = append(*trace, 3218)
							return ActionBi, false
						}
					}
				} else {
					if !t.isPlayerWinOverRiskControlRate() { // 赢后超过风控
						*trace = append(*trace, 3219)
						if algo.HuaTypeUpOrDown(seat.Cards) <= algo.DuiziUp { // 大对子及以上
							*trace = append(*trace, 3220)
							return ActionBi, false
						} else {
							*trace = append(*trace, 3221)
							return ActionPack, false
						}
					} else {
						if algo.HuaTypeUpOrDown(seat.Cards) <= algo.DuiziUp { // 大对子及以上
							*trace = append(*trace, 3222)
							return t.handleRobotStake0(seatid, seat, trace)
						} else {
							*trace = append(*trace, 3223)
							return ActionBi, false
						}
					}
				}
			} else {
				// 判断对手之前是否有过 chaal
				rivalSeatid, chaal := t.isNextPlayerBeforeChaal(seatid)
				if chaal {
					*trace = append(*trace, 3224)
					return t.handleRobotPackFirstRoundRivalChaalV3(seatid, seat, rivalSeatid, trace)
				} else {
					*trace = append(*trace, 3225)
					return t.handleRobotPackFirstRoundRivalUnChaalV3(seatid, seat, rivalSeatid, trace)
				}
			}
		}
	}
}

// pack逻辑非第一轮 对手chaal过
func (t *Desk) handleRobotPackFirstRoundRivalChaalV3(seatid uint32, seat *data.DeskSeat, rivalSeatid uint32, trace *[]int16) (action int32, actionSee bool) {
	if t.isRobot(rivalSeatid) { // 对手非玩家
		*trace = append(*trace, 3301)
		return ActionBi, false
	} else {
		rivalSeat := t.getSeat(rivalSeatid)
		if !algo.HuaCompare(seat.Cards, rivalSeat.Cards) { // 是否比对手大
			*trace = append(*trace, 3302)
			// 人机比牌玩家赢后是否超过风控基础返奖率
			if t.isPlayerWinOverRiskControlRateWithAction(ActionBi, false, seatid) {
				*trace = append(*trace, 3303)
				return ActionPack, false
			} else {
				if algo.HuaTypeUpOrDown(seat.Cards) <= algo.DuiziDown { // 小对子以上
					*trace = append(*trace, 3304)
					return ActionBi, false
				} else {
					*trace = append(*trace, 3305)
					return ActionPack, false
				}
			}
		} else {
			// 取玩家对应牌型的vh和极限倍数的高值判断是否触发极限倍数
			var overflowMultiple bool
			// 玩家对于人机本次拿到的牌型是否有对应场次的vh记录
			vh, _, _, ok := t.getPlayerVHRecord(t.GetOnlyOnePlayer(), seat.Cards...)
			if ok {
				if vh > seat.MaxMultiple {
					*trace = append(*trace, 3306)
					overflowMultiple = t.isOverflowMultipleWithVH(seatid, seat, vh, true) // 用vh判断是否触发极限倍数
				} else {
					*trace = append(*trace, 3307)
					overflowMultiple = t.isOverflowMultiple(seatid, seat, 1) // 是否触发极限倍数
				}
			}
			if !overflowMultiple {
				*trace = append(*trace, 3308)
				return t.handleRobotStake0(seatid, seat, trace)
			} else {
				// 人机比牌玩家赢后是否超过风控基础返奖率
				if t.isPlayerWinOverRiskControlRateWithAction(ActionBi, false, seatid) {
					*trace = append(*trace, 3309)
					return t.handleRobotStake0(seatid, seat, trace)
				} else {
					if algo.HuaTypeUpOrDown(seat.Cards) <= algo.DuiziDown { // 小对子以上
						*trace = append(*trace, 3310)
						return ActionBi, false
					} else {
						*trace = append(*trace, 3311)
						return ActionPack, false
					}
				}
			}
		}
	}
}

// pack逻辑非第一轮 对手未chaal过
func (t *Desk) handleRobotPackFirstRoundRivalUnChaalV3(seatid uint32, seat *data.DeskSeat, rivalSeatid uint32, trace *[]int16) (action int32, actionSee bool) {
	rivalSeat := t.getSeat(rivalSeatid)
	if t.isRobot(rivalSeatid) { // 对手非玩家
		*trace = append(*trace, 3401)
		return ActionBi, false
	} else {
		if algo.HuaTypeUpOrDown(seat.Cards) <= algo.DuiziUp { // 大对子以上
			if !(seat.BiggerThanPlayer || t.isPlayerPack()) { // 比玩家大
				*trace = append(*trace, 3402)
				return ActionBi, false
			} else {
				if t.isOverflowMultiple(seatid, seat, 2) { // 是否触发2倍极限倍数
					*trace = append(*trace, 3403)
					return ActionBi, false
				} else {
					*trace = append(*trace, 3404)
					return t.handleRobotStake0(seatid, seat, trace)
				}
			}
		} else {
			if algo.HuaTypeUpOrDown(seat.Cards) == algo.DuiziDown { // 是小对子
				if !algo.HuaCompare(seat.Cards, rivalSeat.Cards) { // 是否比对手大
					*trace = append(*trace, 3405)
					return ActionBi, false
				} else {
					if !t.isPlayerWinOverRiskControlRate() { // 是否超过风控
						*trace = append(*trace, 3406)
						return ActionBi, false
					} else {
						if t.isOverflowMultiple(seatid, seat, 2) { // 是否触发两倍极限倍数
							*trace = append(*trace, 3407)
							return ActionBi, false
						} else {
							*trace = append(*trace, 3408)
							return t.handleRobotStake0(seatid, seat, trace)
						}
					}
				}
			} else {
				if algo.HuaCompare(seat.Cards, rivalSeat.Cards) { // 是否比对手大
					if !t.isPlayerWinOverRiskControlRate() { // 玩家赢风控
						*trace = append(*trace, 3409)
						return ActionPack, false
					} else {
						if !algo.GaoPaiAK(seat.Cards) { // 是否K或以上的高牌
							*trace = append(*trace, 3410)
							return ActionPack, false
						} else {
							// 50% show或stake
							if utils.RandWan(5000) {
								*trace = append(*trace, 3411)
								return ActionBi, false
							} else {
								*trace = append(*trace, 3412)
								return t.handleRobotStake0(seatid, seat, trace)
							}
						}
					}
				} else {
					if t.isPlayerWinOverRiskControlRate() { // 玩家赢风控
						*trace = append(*trace, 3413)
						return ActionPack, false
					} else {
						if algo.HuaTypeUpOrDown(seat.Cards) == algo.GaopaiDown { // 是小高牌
							*trace = append(*trace, 3414)
							return ActionPack, false
						} else {
							*trace = append(*trace, 3415)
							return ActionBi, false
						}
					}
				}
			}
		}
	}
}
