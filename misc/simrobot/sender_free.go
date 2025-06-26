package main

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
)

// lhd

// sendLHDEntryRoom 进入房间
func (c *Robot) sendLHDEntryRoom() {
	glog.Debugf("enter roomid %s", c.roomid)
	c2s := new(pb.LHFreeEnterRoomReq)
	c2s.Roomid = c.roomid
	c.Sender(c2s)
}

// 下注
func (c *Robot) lhdFreeBet() {
	// glog.Debugf("lhdFreeBet %s", c.roomid)
	if c.lh.state != int32(pb.STATE_BET) {
		// 不在下注状态
		return
	}
	if c.sim {
		// 模拟人机
		chips := []int{500, 1000, 10000, 50000, 100000}
		c.data.LHNowSeat = c.getLHSeat()
		chip, _ := utils.ChoiceInt(chips)
		c.data.LHNowChip = uint32(chip)
		c.lhBet(false)
		return
	}
	if c.lh.lhstrategy == nil {
		return
	}
	if c.data.LHNowChip > 0 {
		// 没下完继续下
		// 拆分筹码,然后下注
		c.lhBet(true)
		return
	}
	// 随机下龙虎和
	c.data.LHNowSeat = c.getLHSeat()
	// 下注金额
	c.data.LHNowChip = c.getLHChip()
	if c.data.LHNowChip <= 0 {
		utils.SleepRand(3)
		c.lhdFreeBet()
		return
	}
	if uint32(c.data.Diamond+c.data.Coin) < c.data.LHNowChip {
		// 筹码不够
		return
	}
	glog.Infof("user %s lh bet seat:%d score:%d sim:%s", c.data.Userid, c.data.LHNowSeat, c.data.LHNowChip, c.sim)
	// 拆分筹码,然后下注
	c.lhBet(false)
}

// 拆分筹码下注
func (c *Robot) lhBet(old bool) {
	seat := c.data.LHNowSeat
	chips := []uint32{500, 1000, 10000, 50000, 100000}
	if !c.sim {
		chips = c.lh.lhstrategy.Chips
	}
	if len(chips) <= 0 {
		return
	}
	if c.data.LHNowChip < chips[0] {
		c.data.LHNowChip = chips[0]
	}
	for i := len(chips) - 1; i >= 0; i-- {
		if c.data.LHNowChip >= chips[i] {
			c.data.LHNowChip -= chips[i]
			// 下注
			c2s := new(pb.LHFreeBetReq)
			c2s.Seat = seat
			c2s.Value = chips[i]
			if !old {
				// 间隔
				c.SendDefer(c2s)
			} else {
				// 0.5ms
				c.SendDefer2(c2s, 500)
			}
			break
		}
	}
}

// 根据人机策略选择下注
func (c *Robot) getLHSeat() uint32 {
	var choices []utils.Choice
	weights := []int32{45, 45, 10}
	if !c.sim {
		if c.lh.lhstrategy == nil || len(c.lh.lhstrategy.BetWeight) == 0 {
			return 1
		}
		weights = c.lh.lhstrategy.BetWeight
	}
	choices = append(choices, utils.Choice{Weight: int(weights[0]), Item: Dragon}) // 龙
	choices = append(choices, utils.Choice{Weight: int(weights[1]), Item: Tiger})  // 虎
	choices = append(choices, utils.Choice{Weight: int(weights[2]), Item: Tie})    // 和
	choice, err := utils.WeightedChoice(choices)
	if err != nil {
		return Dragon
	}
	return choice.Item.(uint32)
}

// 根据人机策略选择筹码
func (c *Robot) getLHChip() uint32 {
	// 根据概率判断下不下注TODO
	pro := []int32{25, 75}
	if c.lh.lhstrategy != nil {
		pro = c.lh.lhstrategy.BetPro
	}
	var choices []utils.Choice
	choices = append(choices, utils.Choice{Weight: int(pro[0]), Item: 0}) // 不下注
	choices = append(choices, utils.Choice{Weight: int(pro[1]), Item: 1}) // 下注
	choice, err := utils.WeightedChoice(choices)
	if err == nil && choice.Item == 0 {
		return 0
	}
	// 选择筹码
	betArea := c.lh.lhstrategy.BetArea
	base := c.lh.lhstrategy.Min
	bet := utils.RandInt32N(betArea[1]/base-betArea[0]/base) + betArea[0]/base
	return uint32(bet * base)
}

//.

//' niu

//· 7up

// sendUPEntryRoom 进入房间
func (c *Robot) sendUPEntryRoom() {
	glog.Debugf("enter roomid %s", c.roomid)
	c2s := new(pb.UPFreeEnterRoomReq)
	c2s.Roomid = c.roomid
	c.Sender(c2s)
}

// 下注
func (c *Robot) upFreeBet() {
	// glog.Debugf("lhdFreeBet %s", c.roomid)
	if c.lh.state != int32(pb.STATE_BET) {
		// 不在下注状态
		return
	}
	if c.sim {
		// 模拟人机
		chips := []int{500, 1000, 10000, 50000, 100000}
		c.data.LHNowSeat = c.getUPSeat()
		chip, _ := utils.ChoiceInt(chips)
		c.data.LHNowChip = uint32(chip)
		c.upBet(false)
		return
	}
	if c.data.LHNowChip > 0 {
		// 没下完继续下
		// 拆分筹码,然后下注
		c.upBet(true)
		return
	}
	// 随机下龙虎和
	c.data.LHNowSeat = c.getUPSeat()
	// 下注金额
	c.data.LHNowChip = c.getUPChip()
	if c.data.LHNowChip <= 0 {
		utils.SleepRand(3)
		c.upFreeBet()
		return
	}
	if uint32(c.data.Diamond+c.data.Coin) < c.data.LHNowChip {
		// 筹码不够
		return
	}
	glog.Infof("user %s 7up bet seat:%d score:%d", c.data.Userid, c.data.LHNowSeat, c.data.LHNowChip)
	// 拆分筹码,然后下注
	c.upBet(false)
}

// 拆分筹码下注
func (c *Robot) upBet(old bool) {
	seat := c.data.LHNowSeat
	chips := []uint32{500, 1000, 10000, 50000, 100000}
	if !c.sim {
		chips = c.up.lhstrategy.Chips
	}
	if len(chips) <= 0 {
		return
	}
	if c.data.LHNowChip < chips[0] {
		c.data.LHNowChip = chips[0]
	}
	for i := len(chips) - 1; i >= 0; i-- {
		if c.data.LHNowChip >= chips[i] {
			c.data.LHNowChip -= chips[i]
			// 下注
			c2s := new(pb.UPFreeBetReq)
			c2s.Seat = seat
			c2s.Value = chips[i]
			if !old {
				// 间隔
				c.SendDefer(c2s)
			} else {
				// 0.5ms
				c.SendDefer2(c2s, 500)
			}
			break
		}
	}

	// 延迟时间下注
	// utils.Sleep(utils.RandIntN(2) + 3) //随机
	// c.lhdFreeBet()
}

// 根据人机策略选择下注
func (c *Robot) getUPSeat() uint32 {
	var choices []utils.Choice
	weights := []int32{45, 45, 0}
	if !c.sim {
		if c.up.lhstrategy == nil || len(c.up.lhstrategy.BetWeight) == 0 {
			return 1
		}
		weights = c.up.lhstrategy.BetWeight
	}
	// if c.up.lhstrategy == nil || len(c.up.lhstrategy.BetWeight) == 0 {
	// 	return 1
	// }
	// weights := c.up.lhstrategy.BetWeight
	choices = append(choices, utils.Choice{Weight: int(weights[0]), Item: Dragon}) // 龙
	choices = append(choices, utils.Choice{Weight: int(weights[1]), Item: Tiger})  // 虎
	choices = append(choices, utils.Choice{Weight: int(weights[2]), Item: Tie})    // 和
	choice, err := utils.WeightedChoice(choices)
	if err != nil {
		return Dragon
	}
	return choice.Item.(uint32)
}

// 根据人机策略选择筹码
func (c *Robot) getUPChip() uint32 {
	// 根据概率判断下不下注TODO
	pro := c.up.lhstrategy.BetPro
	var choices []utils.Choice
	choices = append(choices, utils.Choice{Weight: int(pro[0]), Item: 0}) // 不下注
	choices = append(choices, utils.Choice{Weight: int(pro[1]), Item: 1}) // 下注
	choice, err := utils.WeightedChoice(choices)
	if err == nil && choice.Item == 0 {
		return 0
	}
	// 选择筹码
	betArea := c.up.lhstrategy.BetArea
	base := c.up.lhstrategy.Min
	bet := utils.RandInt32N(betArea[1]/base-betArea[0]/base) + betArea[0]/base
	return uint32(bet * base)
}

//· crash

// sendUPEntryRoom 进入房间
func (c *Robot) sendCRASHEntryRoom() {
	glog.Debugf("enter roomid %s", c.roomid)
	c2s := new(pb.CRASHEnterRoomReq)
	c2s.Roomid = c.roomid
	c.Sender(c2s)
}

// 下注
func (c *Robot) crashFreeBet() {
	// glog.Debugf("lhdFreeBet %s", c.roomid)
	if c.crash.state != int32(pb.STATE_BET) {
		// 不在下注状态
		return
	}
	if c.sim {
		// 模拟人机
		chips := []int{500, 1000, 10000, 50000, 100000}
		chip, _ := utils.ChoiceInt(chips)
		c.data.LHNowChip = uint32(chip)
		c.crashBet(false)
		return
	}
	if c.data.LHNowChip > 0 {
		// 没下完继续下
		// 拆分筹码,然后下注
		c.crashBet(true)
		return
	}
	// 下注金额
	c.data.LHNowChip = c.getCRASHChip()
	if c.data.LHNowChip <= 0 {
		utils.SleepRand(3)
		c.crashFreeBet()
		return
	}
	if uint32(c.data.Diamond+c.data.Coin) < c.data.LHNowChip {
		// 筹码不够
		return
	}
	glog.Infof("user %s crash bet seat:%d score:%d", c.data.Userid, c.data.LHNowSeat, c.data.LHNowChip)
	// 拆分筹码,然后下注
	c.crashBet(false)
}

// 拆分筹码下注
func (c *Robot) crashBet(old bool) {
	chips := []uint32{500, 1000, 10000, 50000, 100000}
	if !c.sim {
		chips = c.crash.cashstrategy.Chips
	}
	if len(chips) <= 0 {
		return
	}
	if c.data.LHNowChip < chips[0] {
		c.data.LHNowChip = chips[0]
	}
	for i := len(chips) - 1; i >= 0; i-- {
		if c.data.LHNowChip >= chips[i] {
			c.data.LHNowChip -= chips[i]
			// 下注
			c2s := new(pb.CRASHBetReq)
			c2s.Value = chips[i]
			if !old {
				// 间隔
				c.SendDefer(c2s)
			} else {
				// 0.5ms
				c.SendDefer2(c2s, 500)
			}
			break
		}
	}

	// 延迟时间下注
	// utils.Sleep(utils.RandIntN(2) + 3) //随机
	// c.lhdFreeBet()
}

// 根据人机策略选择筹码
func (c *Robot) getCRASHChip() uint32 {
	// 根据概率判断下不下注TODO
	pro := c.crash.cashstrategy.BetPro
	var choices []utils.Choice
	choices = append(choices, utils.Choice{Weight: int(pro[0]), Item: 0}) // 不下注
	choices = append(choices, utils.Choice{Weight: int(pro[1]), Item: 1}) // 下注
	choice, err := utils.WeightedChoice(choices)
	if err == nil && choice.Item == 0 {
		return 0
	}
	// 选择筹码
	betArea := c.crash.cashstrategy.BetArea
	base := c.crash.cashstrategy.Min
	bet := utils.RandInt32N(betArea[1]/base-betArea[0]/base) + betArea[0]/base
	return uint32(bet * base)
}
