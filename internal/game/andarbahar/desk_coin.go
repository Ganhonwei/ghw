package andarbahar

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
	if t.robotTime < 20 {
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
	if n >= int(t.Game.AB.Robot.Num) {
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
	msg.Min = t.Game.AB.Robot.InitScore[0]
	msg.Max = t.Game.AB.Robot.InitScore[1]
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
func (t *Desk) coinRoleMsg(userid string) (msg *pb.ABRoomUser) {
	if v, ok := t.roles[userid]; ok {
		if v.Seat == 0 {
			return //没有坐下不广播
		}
		msg = handler.PackABCoinUser(v.User)
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
func (t *Desk) getFinalFactor() (float64, int64) {
	var stock int64
	resp := t.reqRoom(&pb.GetGameStock{Gtype: int32(pb.LOTTERY)})
	if resp == nil {
		glog.Errorf("stock is nil, repeat")
	} else {
		if rsp, ok := resp.(*pb.GetGamedStock); ok {
			stock = rsp.Stock
		}
	}

	if t.Game.Algo_Switch == 2 {
		// 新算法
		return t.getNewFinalFactor(), stock
	}
	// 老算法
	msg := &pb.GetRoomFactor{}
	msg.GameId = t.Game.Id

	res := t.reqRoom(msg)
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
		return int64(math.Ceil(float64(diamond) * float64(t.DeskData.Game.AB.MTax) / 10000)), int64(math.Floor(float64(coin) * float64(t.DeskData.Game.AB.MTax) / 10000))
	}
	return t.calAnTax(diamond), t.calAnTax(coin)
}

// 计算暗税
func (t *Desk) calAnTax(num int64) int64 {
	socre := math.Abs(float64(num))
	real := int64(math.Round(socre * float64(10000-t.DeskData.Game.AB.ATax) / 10000))
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
	msg.Gtype = int32(pb.ABAR)
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
			Gtype:     int32(pb.ABAR),
			GameTime:  time.Now().Unix() - t.BeginTime,
			CashStock: changeCash,
			CashMing:  mcashTaxStock,
			CashAn:    acashTaxStock,
			First:     role.RoundGames[int32(pb.ABAR)] <= 1,
		}
		t.roomPid.Tell(m)
	}
}

// 获取玩家位置上的下注
func (t *Desk) getBetNum(userid string, seat uint32) int64 {
	items := t.DeskFree.ABSeatRoleBets[seat]
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
// func (t *Desk) getTieWin() int64 {
// 	// 当前总下注
// 	betPool := t.DeskGame.CashBets
// 	tie := t.DeskFree.SeatCashBets[Tie]
// 	dragon := t.DeskFree.SeatCashBets[Dragon]
// 	tiger := t.DeskFree.SeatCashBets[Tiger]
// 	odds := t.DeskFree.ABDeskFree.OddsMap[Tie]
// 	return betPool - int64(odds)*tie - dragon/2 - tiger/2
// }

// 获取新手牌型
func (t *Desk) getNewBiewCardType() int32 {
	state := data.NoveiceState
	bean := t.Game.AB.NewbiewMode
	interval := bean.CanWithdrawRange // 可提取间
	winRate := bean.WinRate           // 对应胜率
	for _, role := range t.roles {
		if role.State == data.FrothState {
			// 泡沫模式
			return int32(bean.FrothCardType)
		}

		if role.State == data.ExceptionState {
			state = data.ExceptionState
			interval = bean.CivilianCanWithdrawRange
			winRate = bean.CivilianWinRate
		}
	}

	winning := winRate[len(winRate)-1] // 胜率
	for _, r := range t.roles {
		for i := 0; i < len(interval); i++ {
			if r.OutDiamond < int64(interval[i]) {
				winning = winRate[i]
				break
			}
		}
	}

	var weights, cardTypes []int
	if utils.RandWan(int32(winning)) {
		// 赢
		if state == data.NoveiceState {
			weights = bean.WinWeight
			cardTypes = bean.WinCardType
		} else {
			weights = bean.CivilianWinWeight
			cardTypes = bean.CivilianWinCardType
		}
	} else {
		// 输
		if state == data.NoveiceState {
			weights = bean.LoseWeight
			cardTypes = bean.LoseCardType
		} else {
			weights = bean.CivilianLoseWeight
			cardTypes = bean.CivilianLoseCardType
		}
	}

	index, err := utils.ChoiceIntIndex(weights)
	if err != nil {
		glog.Errorf("config err in newbiew room")
		return 0
	}
	return int32(cardTypes[index])
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
			Gtype:  int32(pb.ABAR),
			Time:   time.Now().Unix() - t.BeginTime,
		}
		myactor.Logger().Tell(msg)
		t.send2userid(userid, msg)
	}
}

// 计算16种结果输赢分
func (t *Desk) calWinningScoreOf16() []DrawResult {
	bahar := t.SeatCashBets[Bahar] // bahar下注
	andar := t.SeatCashBets[Andar] // andar下注

	// 输赢分
	wins := make([]DrawResult, 0)
	for i := 3; i < 11; i++ {
		seatBet := t.SeatCashBets[uint32(i)]          // 边注
		other := t.CashBets - bahar - andar - seatBet // 其他位置总下注
		// 当前边注倍数
		odds := t.ABDeskFree.OddsMap[uint32(i)]

		// 结果为andar输赢分
		winScore := bahar + other - andar*(int64(t.ABDeskFree.OddsMap[Andar]-100))/100 - seatBet*(int64(odds)-100)/100
		wins = append(wins, DrawResult{Winner: Andar, SideWinner: uint32(i), Score: winScore})

		// 结果为bahar输赢分
		winScore = andar + other - bahar*(int64(t.ABDeskFree.OddsMap[Bahar]-100))/100 - seatBet*(int64(odds)-100)/100
		wins = append(wins, DrawResult{Winner: Bahar, SideWinner: uint32(i), Score: winScore})
	}

	return wins
}

// 获取玩家位置对应下注
func (t *Desk) getSeatBetsOfUserid(userid string) map[string]int64 {
	seatBetMap := make(map[string]int64)
	for k, bets := range t.ABSeatRoleBets {
		if b, ok := bets[userid]; ok {
			seatBetMap[utils.String(k)] += b.GetSum()
		}
	}
	return seatBetMap
}

// 是否游戏进行中
func (t *Desk) IsGaming() bool {
	if t.state == int32(pb.STATE_DEALING) || t.state == int32(pb.STATE_BET) || t.state == int32(pb.STATE_PAUSE) || t.state == int32(pb.STATE_CHARGE) {
		return true
	} else {
		return false
	}
}

// 库存、明税、暗税按比例拆分为彩金、奖励金
func (t *Desk) calcStockAndTax(score int64, ratio float64, options ...int64) (int64, *data.ActStock) {
	score_final, stock, ming, an := t._calcStockAndTax(score, options...)
	//库存拆分
	cash_stock := int64(math.Round(ratio * float64(stock)))
	bonus_stock := stock - cash_stock
	//明税拆分
	cash_ming := int64(math.Round(ratio * float64(ming)))
	bonus_ming := ming - cash_ming
	//暗税拆分
	cash_an := int64(math.Round(ratio * float64(an)))
	bonus_an := an - cash_an

	return score_final, &data.ActStock{
		CashStock:  cash_stock,
		BonusStock: bonus_stock,
		CashMing:   cash_ming,
		BonusMing:  bonus_ming,
		CashAn:     cash_an,
		BonusAn:    bonus_an,
	}
}

// 计算单个玩家库存、明税、暗税
// 输入：score-输赢分
// 输出：score1-最终输赢分 stock-库存变动 ming-明税 an-暗税
// 明税：赢才有
// 暗税：输赢都有
func (t *Desk) _calcStockAndTax(score int64, options ...int64) (score1, stock, ming, an int64) {
	if score == 0 {
		return
	}
	ming_tax := t.Game.AB.MTax
	an_tax := t.Game.AB.ATax

	if score > 0 { //赢分
		if ming_tax != 0 { //明税只对赢有效
			ming = int64(math.Round(float64(score) * float64(ming_tax) / float64(10000))) //明税
		}
		score1 = score - ming //最终赢分
		stock = -score1       //减库存初始值

		if an_tax != 0 { //计算暗税
			an = int64(math.Round(float64(score1) * float64(an_tax) / float64(10000))) //暗税
		}

		stock -= an //赢分库存多减暗税
	} else { //输分
		var score_deduction int64
		if len(options) != 0 {
			score_deduction = options[0]
		}
		score1 = score                   //最终输分
		stock = -score + score_deduction //加库存初始值
		if an_tax != 0 {
			an = int64(math.Round(float64(-score1) * float64(an_tax) / float64(10000))) //暗税
		}

		stock -= an //输分库存少加暗税
	}
	return
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
	return t.DeskData.Game.AB.LoseLimited
}

//.

// vim: set foldmethod=marker foldmarker=//',//.:
