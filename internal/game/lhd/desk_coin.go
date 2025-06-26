package lhd

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"goserver/pkg/utils"
	"math"
	"strconv"
	"time"
)

// 更新点控
func (t *Desk) sendPointControl(userid string) {
	if v, ok := t.roles[userid]; ok && v != nil {
		msg := &pb.ChangePointControl{
			UserId:        v.Userid,
			Switch:        v.PCSwitch,
			Factor:        v.PCFactor,
			Score:         v.PCScore,
			ScoreComplete: v.PCScoreComplete,
		}
		t.rolePid.Tell(msg)
	}
}

// 更新货币
func (t *Desk) sendCurrency(userid string, coin, diamond int64, ltype int32, desc string) {
	if coin == 0 && diamond == 0 {
		return
	}
	//玩家在线
	if v, ok := t.roles[userid]; ok && v != nil {
		v.User.AddCoin(coin)
		v.User.AddDiamond(diamond)
		//在线
		if !v.Offline {
			//货币变更及时同步
			msg := &pb.ChangeCurrency{
				Userid:  userid,
				Coin:    coin,
				Diamond: diamond,
				Type:    ltype,
				Desc:    desc,
				WaterId: t.GameId,
				Control: t.Dtype == int32(pb.DESK_TYPE_POINTCONTROL),
			}
			v.Pid.Tell(msg)
			return
		}
	}
	glog.Infof("sendCoin userid %s, coin %d, diamond %d ltype %d",
		userid, coin, diamond, ltype)
	//TODO 检测是否在其它房间内,如果在则通过房间同步,否则正常同步
	msg := &pb.OfflineCurrency{
		Userid:  userid,
		Coin:    coin,
		Diamond: diamond,
		Type:    ltype,
		Desc:    desc,
		WaterId: t.GameId,
		Control: t.Dtype == int32(pb.DESK_TYPE_POINTCONTROL),
	}
	//通过大厅通知其它节点
	t.rolePid.Tell(msg)
}

// '更新金币
func (t *Desk) sendCoin(userid string, num int64, ltype int32) {
	if num == 0 {
		return
	}
	//玩家在线
	if v, ok := t.roles[userid]; ok && v != nil {
		v.User.AddCoin(num)
		//在线
		if !v.Offline {
			//货币变更及时同步
			msg := &pb.ChangeCurrency{
				Userid: userid,
				Coin:   num,
				Type:   ltype,
			}
			v.Pid.Tell(msg)
			return
		}
	}
	//离线同步数据
	glog.Infof("sendCoin userid %s, num %d, ltype %d",
		userid, num, ltype)
	//TODO 检测是否在其它房间内,如果在则通过房间同步,否则正常同步
	msg := &pb.OfflineCurrency{
		Userid: userid,
		Coin:   num,
		Type:   ltype,
	}
	//通过大厅通知其它节点
	t.rolePid.Tell(msg)
}

func (t *Desk) sendDiamond(userid string, num int64, ltype int32) {
	if num == 0 {
		return
	}
	//玩家在线
	if v, ok := t.roles[userid]; ok && v != nil {
		v.User.AddDiamond(num)
		//在线
		if !v.Offline {
			//货币变更及时同步
			msg := &pb.ChangeCurrency{
				Userid:  userid,
				Diamond: num,
				Type:    ltype,
			}
			v.Pid.Tell(msg)
			return
		}
	}
	//离线同步数据
	glog.Infof("sendDiamond userid %s, num %d, ltype %d",
		userid, num, ltype)
	msg := &pb.OfflineCurrency{
		Userid:  userid,
		Diamond: num,
		Type:    ltype,
	}
	//通过大厅通知其它节点
	t.rolePid.Tell(msg)
}

//.

// '召唤机器人
func (t *Desk) loadRobot() {
	t.robotTime++
	if t.robotTime < 10 {
		return
	}
	t.robotTime = 0
	t.callRobot()
}

func (t *Desk) callRobot() {
	// if len(t.roles) >= 5 {
	// 	return
	// }
	if len(t.roles) >= int(t.Game.Count) {
		return
	}
	r, n := t.roleCountNum()
	if n >= int(t.Game.LHD.Robot.Num) {
		return
	}
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0):
		if r >= 3 || n >= 2 {
			return
		}
	case int32(pb.ROOM_TYPE1):
		if !t.DeskData.Pub {
			return
		}
		if r >= 2 || n >= 2 {
			return
		}
	case int32(pb.ROOM_TYPE2):
	}
	msg := new(pb.RobotMsg)
	msg.Roomid = t.DeskData.Rid
	msg.Code = t.DeskData.Code
	msg.Rtype = t.DeskData.Rtype
	msg.Ltype = t.DeskData.Ltype
	msg.Gtype = t.DeskData.Gtype
	msg.EnvBet = int32(t.DeskData.Multiple)
	msg.Min = t.Game.LHD.Robot.InitScore[0]
	msg.Max = t.Game.LHD.Robot.InitScore[1]
	msg.Num = 1
	t.dbmsPid.Tell(msg)
}

func (t *Desk) roleCountNum() (r, n int) {
	for _, v := range t.roles {
		if v.User.GetRobot() {
			n++
		} else {
			r++
		}
	}
	return
}

// 能否下注
func (t *Desk) canBet(role *data.DeskRole, num int64) bool {
	user := role.User
	cash := t.getCashOrCoinBet(role, num, true)
	if user.Diamond < cash || user.Coin < num-cash {
		// 彩金不够下注 || 奖励金不够下注
		glog.Errorf("no enough coin d:%d,c:%d, num:%d", user.Diamond, user.Coin, num)
		return false
	}
	return true
}

// 下注的奖励金和彩金(彩金向下取整，奖励金向上取整)
func (t *Desk) getCashOrCoinBet(role *data.DeskRole, num int64, cash bool) int64 {
	cashPro := handler.GetCashProportion(role.User) // 每次下注都算
	c := int64(math.Floor(float64(int64(cashPro)*num) / 10000))
	if cash {
		return c
	}
	return num - c
}

// 位置上玩家数据
func (t *Desk) coinRoleMsg(userid string) (msg *pb.LHRoomUser) {
	if v, ok := t.roles[userid]; ok {
		if v.Seat == 0 {
			return //没有坐下不广播
		}
		msg = handler.PackLHCoinUser(v.User)
		msg.Seat = v.Seat
		msg.Offline = v.Offline
		if val, ok := t.seats[v.Seat]; ok {
			msg.Dealer = val.BeDealer
			msg.Bet = val.Bet
			msg.Num = val.DealerN
			msg.Niu = val.Niu
			msg.Ready = val.Ready
		}
		if t.DeskPriv != nil {
			msg.Score = t.DeskPriv.PrivScore[userid]
		}
	}
	return
}

// 获取平均系数
func (t *Desk) getAverageFactor() float64 {
	var total float64
	for _, v := range t.roles {
		// 没下注跳过
		bet := t.DeskFree.Bets[v.Userid]
		if bet <= 0 || v.GetRobot() {
			continue
		}
		if _, ok := t.roles[v.Userid]; ok {
			userfactor := t.UserFactorMap[v.Userid]
			f, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(userfactor*(float64(bet)/float64(t.DeskGame.CashBets+t.DeskGame.CoinBets)))), 64)
			total += f
		}
	}
	return total
}

// 获取最终系数
func (t *Desk) getFinalFactor() float64 {
	if t.Game.Algo_Switch == 2 {
		// 新算法
		return t.getNewFinalFactor()
	}

	msg := &pb.GetRoomFactor{}
	msg.GameId = t.Game.Id

	res := t.reqRoom(msg)
	if res == nil {
		return t.getFinalFactor()
	}
	var response *pb.GetedRoomFactor
	var ok bool

	if response, ok = res.(*pb.GetedRoomFactor); !ok {
		glog.Errorf("get  room factor failed: %#v", res)
	}

	ret := (2 - t.getAverageFactor() + response.Factor) / 2
	if ret < 0 {
		ret = 0
	} else if ret > 2 {
		ret = 2
	}
	return ret
}

// 最终系数新算法
func (t *Desk) getNewFinalFactor() float64 {
	ret := 2 - t.getAverageFactor()
	if ret < 0 {
		ret = 0
	} else if ret > 2 {
		ret = 2
	}
	return ret
}

// 获取点控系数
func (t *Desk) getControlFactor() float64 {
	for _, v := range t.roles {
		if v.GetRobot() {
			continue
		}
		if !v.Robot && v.PCSwitch {
			return float64(v.PCFactor) / 100
		}
	}
	return t.getFinalFactor()
}

// 收税
func (t *Desk) getTax(diamond, coin int64, m bool) (int64, int64) {
	if m {
		// 明税
		return int64(math.Ceil(float64(diamond) * float64(t.DeskData.Game.LHD.LHMTax) / 10000)), int64(math.Floor(float64(coin) * float64(t.DeskData.Game.LHD.LHMTax) / 10000))
	}
	return t.calAnTax(diamond), t.calAnTax(coin)
}

// 计算暗税
func (t *Desk) calAnTax(num int64) int64 {
	socre := math.Abs(float64(num))
	real := int64(math.Round(socre * float64(10000-t.DeskData.Game.LHD.LHATax) / 10000))
	return int64(socre) - real
}

// 改变库存
func (t *Desk) changeStock(changeCash, changeCoin, mcashTaxStock, mcoinTaxStock, acashTaxStock, acoinTaxStock int64, id string) {
	if changeCash == 0 && changeCoin == 0 && mcashTaxStock == 0 &&
		mcoinTaxStock == 0 && acashTaxStock == 0 && acoinTaxStock == 0 {
		return
	}
	msg := new(pb.ChangeStock)
	msg.GameId = id
	msg.Real = t.Dtype != int32(pb.DESK_TYPE_NEWBIEW)
	msg.Gtype = int32(pb.LHD)
	msg.CashStock = changeCash
	msg.BonusStock = changeCoin
	msg.CashMingTax = mcashTaxStock
	msg.BonusMingTax = mcoinTaxStock
	msg.CashAnTax = acashTaxStock
	msg.BonusAnTax = acoinTaxStock
	// t.roomPid.Tell(msg)
	res := t.reqRoom(msg)
	if _, ok := res.(*pb.ChangedStock); !ok {
		glog.Errorf("change stock failed: %#v", res)
	}

	if t.Dtype == int32(pb.DESK_TYPE_NEWBIEW) {
		// 新手放库存记录
		userid := t.GetOnlyOnePlayer()
		if userid == "" {
			return
		}
		bet := t.Bets[userid]
		if bet == 0 {
			return
		}

		role := t.roles[userid]
		if role.RoundGames == nil {
			role.RoundGames = make(map[int32]int32)
		}

		m := &pb.NewbieStock{
			Gtype:     int32(pb.LHD),
			GameTime:  time.Now().Unix() - t.BeginTime,
			CashStock: changeCash,
			CashMing:  mcashTaxStock,
			CashAn:    acashTaxStock,
			First:     role.RoundGames[int32(pb.LHD)] <= 1,
		}
		t.roomPid.Tell(m)
	}
}

// 获取玩家位置上的下注
func (t *Desk) getBetNum(userid string, seat uint32) int64 {
	items := t.DeskFree.LHSeatRoleBets[seat]
	if items == nil {
		return 0
	}
	if c, ok := items[userid]; ok {
		return c.GetSum()
	}
	return 0
}

// 输赢结果
func (t *Desk) getWinReslut(score int64) string {
	if score == 0 {
		return "平"
	} else if score > 0 {
		return "赢"
	}
	return "输"
}

// 开和输赢
func (t *Desk) getTieWin() int64 {
	// 当前总下注
	betPool := t.DeskGame.CashBets
	tie := t.DeskFree.SeatCashBets[Tie]
	dragon := t.DeskFree.SeatCashBets[Dragon]
	tiger := t.DeskFree.SeatCashBets[Tiger]
	odds := t.DeskFree.LHDeskFree.OddsMap[Tie]
	return betPool - int64(odds)*tie - dragon/2 - tiger/2
}

// 选择赢家
func (t *Desk) choiceWinner(lhtMap map[uint32]int64, otherMap map[uint32]int64, playerwin bool) uint32 {
	if len(lhtMap) == 1 {
		if _, ok := lhtMap[Tie]; ok && playerwin {
			//玩家赢，并且只能开和
			//有1/3的概率开到和
			if utils.RandInt32N(3)+1 <= 1 {
				return Tie
			} else {
				res := uint32(utils.RandInt32N(2) + 1)
				score := otherMap[res]
				for k, v := range otherMap {
					if v < score {
						score = v
						res = k
					}
				}
				glog.Infof("player only open tie, final seat:%d, gameid:%s", res, t.GameId)
				return res
			}
		}
		// 只有一种结果就直接开
		for k, _ := range lhtMap {
			return k
		}
	}

	switch len(lhtMap) {
	case 2:
		if _, ok := lhtMap[Tie]; ok {
			// 有和，开和概率为3/51
			if utils.RandInt32N(52)+1 <= 3 {
				return Tie
			}
		}
		if playerwin {
			// 玩家赢，但是开的结果都是平台赢,直接开玩家输得少的
			var score int64 = 0
			var res uint32 = 3
			for k, v := range lhtMap {
				if v > score {
					score = v
					res = k
				}
			}
			if res < 3 {
				return res
			}
		}
		if _, ok := lhtMap[Tie]; ok {
			// 有和，开和概率为3/51
			if utils.RandInt32N(52)+1 <= 3 {
				return Tie
			} else {
				if _, ok := lhtMap[Dragon]; ok {
					return Dragon
				} else {
					return Tiger
				}
			}
		} else {
			// 没有和,赢分多的概率为70%，赢分少的概率为30%。赢分一样多，各50%
			choices := make([]utils.Choice, 0)
			var score int64 = 0
			for k, v := range lhtMap {
				if v > score {
					choices = append(choices, utils.Choice{Weight: 70, Item: int(k)})
					if len(choices) > 1 {
						choices[0].Weight = 30
					}
					score = v
				} else if v == score {
					choices = append(choices, utils.Choice{Weight: 50, Item: int(k)})
					if len(choices) > 1 {
						choices[0].Weight = 50
					}
					score = v
				} else {
					choices = append(choices, utils.Choice{Weight: 30, Item: int(k)})
					score = v
				}
			}
			c, err := utils.WeightedChoice(choices)
			if err != nil {
				glog.Errorf("choice winner error, err:%s", err)
			}
			return uint32(c.Item.(int))
		}
	case 3: // 三个都能开
		choices := make([]utils.Choice, 0)
		if playerwin {
			// 玩家赢，但是开的结果都是平台赢
			var lose bool = false
			for _, v := range lhtMap {
				if v > 0 {
					lose = true
					break
				}
			}
			if lose {
				var score int64 = 0
				for k, v := range lhtMap {
					if k == 0 {
						continue
					}
					if v > score {
						choices = append(choices, utils.Choice{Weight: 15, Item: int(k)})
						if len(choices) > 1 {
							choices[0].Weight = 33
						}
					} else if v == score {
						choices = append(choices, utils.Choice{Weight: 24, Item: int(k)})
						if len(choices) > 1 {
							choices[0].Weight = 24
						}
					} else {
						choices = append(choices, utils.Choice{Weight: 33, Item: int(k)})
					}
					score = v
				}
				choices = append(choices, utils.Choice{Weight: 3, Item: 0})
				c, err := utils.WeightedChoice(choices)
				if err != nil {
					glog.Errorf("choice winner error, err:%s", err)
				}
				return uint32(c.Item.(int))
			}
		}

		// if _, ok := lhtMap[Tie]; ok {
		// 	// 有和，开和概率为3/51
		// 	if utils.RandInt32N(52)+1 <= 3 {
		// 		return Tie
		// 	}
		// }
		delete(lhtMap, Tie)
		//
		var score int64 = 0
		for k, v := range lhtMap {
			if v > score {
				choices = append(choices, utils.Choice{Weight: 33, Item: int(k)})
				if len(choices) > 1 {
					choices[0].Weight = 15
				}
			} else if v == score {
				choices = append(choices, utils.Choice{Weight: 24, Item: int(k)})
				if len(choices) > 1 {
					choices[0].Weight = 24
				}
			} else {
				choices = append(choices, utils.Choice{Weight: 15, Item: int(k)})
			}
			score = v
		}
		choices = append(choices, utils.Choice{Weight: 3, Item: 0})
		c, err := utils.WeightedChoice(choices)
		if err != nil {
			glog.Errorf("choice winner error, err:%s", err)
		}
		return uint32(c.Item.(int))
	}

	glog.Errorf("no find seat result, open dragon")
	return Dragon
}

// 获取新手胜率
func (t *Desk) getNewBiewWinning() int32 {
	// beginner := table.GetTables().NewbieTable.Get()
	// if beginner.Id == 0 {
	// 	glog.Errorf("[lhd] no find beginner config")
	// 	return 0
	// }
	interval := t.Game.LHD.NewbiewMode.OutCashInterval
	winning := t.Game.LHD.NewbiewMode.Winning[len(interval)-1]
	for _, r := range t.roles {
		if r.State == data.FrothState {
			bets := t.Bets[r.Userid]
			winning = handler.GetFreeFrothWinning(bets)
			break
		}
		for i := 1; i < len(interval); i++ {
			if r.OutDiamond < int64(interval[i]) {
				winning = t.Game.LHD.NewbiewMode.Winning[i-1]
				break
			}
		}
	}
	return winning
}

func (t *Desk) shareAmount(userid string, score int64) {
	if role, ok := t.roles[userid]; ok {
		if role.ShareSuperior == "" || role.State != 2 {
			return
		}
		msg := &pb.ShareBetAmount{
			Score:    score,
			Userid:   userid,
			Superior: role.ShareSuperior,
			Name:     role.Nickname,
		}
		t.roomPid.Tell(msg)
	}
}

func (t *Desk) GameTime(userid string) {
	if role, ok := t.roles[userid]; ok {
		if role.Robot {
			return
		}

		msg := &pb.LogGameTime{
			Userid: userid,
			Gtype:  int32(pb.LHD),
			Time:   time.Now().Unix() - t.BeginTime,
		}
		myactor.Logger().Tell(msg)
		t.send2userid(userid, msg)
	}
}

func (t *Desk) flowWater(userid string, score int64) {
	if score == 0 {
		return
	}
	// 事件
	bean := &event.ShopPotFlowEvent{
		Score: score,
	}
	t.eventPost(userid, event.POT_FLOW, bean) // 商城打码量
}

func (t *Desk) getMaxLose() int32 {
	for _, r := range t.roles {
		if r.RegistArea == 1 && handler.FluctuateLine(r.User) >= 500 {
			// B类起伏线超过5000不限制
			return 0
		}
	}
	return t.DeskData.Game.LHD.LoseLimited
}

//.

// vim: set foldmethod=marker foldmarker=//',//.:
