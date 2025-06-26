package lottery

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/algo"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"goserver/pkg/table"
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

		//在线
		if !v.Offline {
			v.User.AddCoin(coin)
			v.User.AddDiamond(diamond)

			//货币变更及时同步
			msg := &pb.ChangeCurrency{
				Userid:    userid,
				Coin:      coin,
				Diamond:   diamond,
				Type:      ltype,
				Desc:      desc,
				WaterId:   t.GameId,
				Control:   t.Dtype == int32(pb.DESK_TYPE_POINTCONTROL),
				SyncGame:  true,
				SyncCpOff: true, // 不用再次同步彩票
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
	if n >= int(t.Game.LOTTERY.Robot.Num) {
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
	msg.Min = t.Game.LOTTERY.Robot.InitScore[0]
	msg.Max = t.Game.LOTTERY.Robot.InitScore[1]
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

// 下注的奖励金和彩金(彩金向下取整，奖励金向上取整)
func (t *Desk) getCashOrCoinBet(role *data.DeskRole, num int64, cash bool) int64 {
	cashPro := handler.GetCashProportion(role.User) // 每次下注都算
	c := int64(math.Floor(float64(int64(cashPro)*num) / 10000))
	if cash {
		return c
	}
	return num - c
}

// 获取平均系数
func (t *Desk) getAverageFactor() float64 {
	var total float64
	r, _ := t.roleCountNum()
	if r <= 0 {
		return 1
	}
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
func (t *Desk) getFinalFactor() (float64, int64) {
	var stock int64
	res := t.reqRoom(&pb.GetGameStock{Gtype: int32(pb.LOTTERY)})
	if res == nil {
		glog.Errorf("stock is nil, repeat")
	} else {
		if rsp, ok := res.(*pb.GetGamedStock); ok {
			stock = rsp.Stock
		}
	}

	if t.Game.Algo_Switch == 2 {
		// 新算法
		return t.getNewFinalFactor(), stock
	}

	msg := &pb.GetRoomFactor{}
	msg.GameId = t.Game.Id

	res = t.reqRoom(msg)
	if res == nil {
		return 0, 0
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
	return ret, stock
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
func (t *Desk) getControlFactor() (float64, int64) {
	for _, v := range t.roles {
		if v.GetRobot() {
			continue
		}
		if !v.Robot && v.PCSwitch {
			return float64(v.PCFactor) / 100, 0
		}
	}
	return t.getFinalFactor()
}

// 收税
func (t *Desk) getTax(diamond, coin int64, m bool) (int64, int64) {
	if m {
		// 明税
		return int64(math.Ceil(float64(diamond) * float64(t.DeskData.Game.LOTTERY.MTax) / 10000)), int64(math.Floor(float64(coin) * float64(t.DeskData.Game.LOTTERY.MTax) / 10000))
	}
	return t.calAnTax(diamond), t.calAnTax(coin)
}

// 计算暗税
func (t *Desk) calAnTax(num int64) int64 {
	socre := math.Abs(float64(num))
	real := int64(math.Round(socre * float64(10000-t.DeskData.Game.LOTTERY.ATax) / 10000))
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
	msg.Gtype = int32(pb.LOTTERY)
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
			Gtype:     int32(pb.LOTTERY),
			GameTime:  time.Now().Unix() - t.BeginTime,
			CashStock: changeCash,
			CashMing:  mcashTaxStock,
			CashAn:    acashTaxStock,
			First:     role.RoundGames[int32(pb.LOTTERY)] <= 1,
		}
		t.roomPid.Tell(m)
	}
}

// 获取玩家位置上的下注
func (t *Desk) getBetNum(userid string, seat uint32) int64 {
	items := t.LSeatRoleBets[seat]
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

// 获取新手胜率
func (t *Desk) getNewBiewWinning() int32 {
	// beginner := table.GetTables().NewbieTable.Get()
	// if beginner.Id == 0 {
	// 	glog.Errorf("[cp] no find beginner config")
	// 	return 0
	// }
	interval := t.Game.LOTTERY.NewbiewMode.OutCashInterval
	winning := t.Game.LOTTERY.NewbiewMode.Winning[len(interval)-1]
	for _, r := range t.roles {
		if r.State == data.FrothState {
			bets := t.Bets[r.Userid]
			winning = handler.GetFreeFrothWinning(bets)
			break
		}
		for i := 1; i < len(interval); i++ {
			if r.OutDiamond < int64(interval[i]) {
				winning = t.Game.LOTTERY.NewbiewMode.Winning[i-1]
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
			Gtype:  int32(pb.LOTTERY),
			Time:   time.Now().Unix() - t.BeginTime,
		}
		myactor.Logger().Tell(msg)
		t.send2userid(userid, msg)
	}
}

// 计算可开的牌型
func (t *Desk) drawCardType() []uint32 {
	cardTypes := make([]uint32, 0)
	// exclude := make([]uint32, 0)
	betPool := t.CashBets // 总下注

	newbieMustWin := t.isNewbieMustWin()
	maxwin := t.getMaxWinScore() // 玩家赢分上限

	// 每个位置赔付
	for k, v := range t.DeskFree.Multiple {
		// if k == algo.BaoZi {
		// 	continue
		// }
		var bets int64 = 0
		betMap := t.LSeatRoleBets[k]
		for userid, b := range betMap {
			role := t.roles[userid]
			if role == nil || role.Robot {
				continue
			}
			bets += b.Diamond
		}
		if bets <= 0 && !newbieMustWin {
			cardTypes = append(cardTypes, k)
			continue
		}

		// 赔付分数，小于零平台赢
		lose := bets*v - betPool
		if k == algo.BaoZi {
			var allBets int64 = 0 // 人机+真人下注
			for _, v := range betMap {
				allBets += v.GetSum()
			}
			//JACKPOT总值*0.2÷（人机+真人本轮在此总下注金额）*所有真人下注金额-6个点位总下注
			lose = t.CPJackpot*2000/allBets*bets/10000 - betPool
		}

		if (lose >= 0 && lose <= maxwin) || lose < 0 {
			// 没超过输分上限
			cardTypes = append(cardTypes, k)
		}
	}

	//
	return cardTypes
}

// 奖池开奖
func (t *Desk) drawJackpot(winner uint32) int64 {
	if winner != algo.BaoZi {
		return 0
	}
	pool := t.CPJackpot * 2000 / 10000
	roles := t.LSeatRoleBets[algo.BaoZi]
	if roles == nil {
		return 0
	}
	t.CPJackpot -= pool
	return pool
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
	// for _, r := range t.roles {
	// 	if r.RegistArea == 1 && handler.FluctuateLine(r.User) >= 500 {
	// 		// B类起伏线超过5000不限制
	// 		return 0
	// 	}
	// }
	return t.DeskData.Game.LOTTERY.LoseLimited
}

func (t *Desk) getMaxWinScore() int64 {
	begginner := table.GetTables().NewbieTable.Get()
	if begginner == nil {
		return 0
	}
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return 0
	}

	role := t.roles[userid]

	if role.Money <= 0 {
		return t.Game.LOTTERY.NewbieMaxWin - role.Diamond
	}
	rate := t.Game.LOTTERY.WinnabilityRate
	factor := t.UserFactorMap[userid]
	if role.PCSwitch {
		// 点控玩家
		rate = role.PCFactor
	}
	// 系数差值
	// m := math.Pow(begginner.FactorSeed, float64(role.Money)/10000)
	// diff := (float64(rate)/10000 - factor) * float64(role.Money)
	// glog.Infof("m:%.6f,rate:%d,factor:%f,diff:%f", m, rate, factor, diff)
	return int64((float64(rate)/10000 - factor) * float64(role.Money))
}

func (t *Desk) isNewbieMustWin() bool {
	if t.Dtype != int32(pb.DESK_TYPE_NEWBIEW) {
		return false
	}

	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return false
	}

	role := t.roles[userid]
	bet := t.Bets[userid]
	if bet <= 0 {
		return false
	}

	if bet+role.Diamond < t.Game.LOTTERY.NewbieMinWin {
		return utils.RandWan(t.Game.LOTTERY.NewbieGiveRate)
	}
	return false
}

func (t *Desk) jackpotExceed() bool {
	var bet int64
	// userid := t.GetOnlyOnePlayer()
	for k, role := range t.roles {
		if role.Robot {
			continue
		}
		bet += t.Bets[k]
	}

	if bet > 0 {
		return false
	}

	if t.CPJackpot > t.Game.LOTTERY.JackpotExceed && utils.RandWan(3000) {
		return true
	}

	return false
}

// 计算玩家输得最多的位置
func (t *Desk) getUserLoseMostSeat(userid string) uint32 {
	allBet := t.Bets[userid]
	var loseMost int64
	var seat uint32 = 1
	for s, betmap := range t.LSeatRoleBets {
		odds := t.DeskFree.Multiple[s]
		if bet, ok := betmap[userid]; ok {
			win := bet.GetSum()*odds/100 - allBet
			if win < loseMost {
				loseMost = win
				if seat > s {
					seat = s
				}
			}
		}
	}
	return seat
}

//.

// vim: set foldmethod=marker foldmarker=//',//.:
