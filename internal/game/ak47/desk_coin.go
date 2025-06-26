package ak47

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/algo"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"math"
	"strings"
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
func (t *Desk) sendCurrency(userid string, score int64, ltype int32, desc string) {
	if score == 0 {
		return
	}
	//玩家在线
	if v, ok := t.roles[userid]; ok && v != nil {
		diamond := int64(math.Round((float64(score) * v.Ratio)))
		coin := score - diamond

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
			}
			v.Pid.Tell(msg)
			return
		}
		glog.Infof("sendCurrency userid %s, coin %d, diamond %d ltype %d",
			userid, coin, diamond, ltype)
		//TODO 检测是否在其它房间内,如果在则通过房间同步,否则正常同步
		msg := &pb.OfflineCurrency{
			Userid:  userid,
			Coin:    coin,
			Diamond: diamond,
			Type:    ltype,
			Desc:    desc,
			WaterId: "",
		}
		//通过大厅通知其它节点
		t.rolePid.Tell(msg)
	}
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

// 是否游戏进行中
func (t *Desk) IsGaming() bool {
	if t.state == int32(pb.STATE_DEALING) || t.state == int32(pb.STATE_BET) || t.state == int32(pb.STATE_PAUSE) || t.state == int32(pb.STATE_CHARGE) {
		return true
	} else {
		return false
	}
}

// 进入房间响应消息
func (t *Desk) coinEnterMsg(userid string) *pb.AK47CoinEnterRoomRsp {
	msg := new(pb.AK47CoinEnterRoomRsp)
	//房间数据
	msg.Roominfo = handler.PackAK47CoinRoom(t.DeskData)
	msg.Roominfo.State = t.state
	msg.Gameid = t.Game.Id
	// 房间暂停状态原因
	if t.state == int32(pb.STATE_PAUSE) {
		msg.Roominfo.StatePauseReaon = int32(t.reason)
	}
	//坐下玩家信息
	msg.Userinfo = t.coinSeatBetsMsg(userid)
	//位置下注信息
	msg.Betsinfo = t.coinBetsMsg()
	msg.BetInfo = t.betInfo

	if t.DeskAct != nil {
		msg.Actseat = t.DeskAct.ActSeat
		msg.Actstate = t.DeskAct.ActState
		msg.Timer = int32(BetTime - t.timer)
		msg.Ante = uint32(t.DeskAct.ActAnte)
		msg.Pot = t.DeskGame.BetNum
		msg.Dealer = t.DeskGame.DealerSeat
		msg.Totaltimer = BetTime
		msg.Turn = t.DeskAct.ActTimes

		if t.timer > BetTime {
			msg.Timer = int32(ChargeInGameTime - t.timer)
			msg.Totaltimer = ChargeInGameTime - 120
		}
		if t.state == int32(pb.STATE_CHARGE) {
			msg.Timer = int32(ChargeInGameTime - t.timer)
			msg.Totaltimer = ChargeInGameTime
			msg.RechargeTime = int64(ChargeInGameTime-t.timer) + utils.LocalTime().Unix()
			amount, give := t.getRechargeAmount(t.DeskAct.ActSeat)
			msg.Amount, msg.GiveAmount = int32(amount), int32(give)
			msg.ChargeInfo = t.getChargeInfo(t.DeskAct.ActSeat)
		}
	}
	//假位置
	for k := range t.FakeSeats {
		msg.Roominfo.FakeSeats = append(msg.Roominfo.FakeSeats, k)
	}
	return msg
}

// 进入消息
func (t *Desk) coinCameinMsg(userid string) {
	msg := new(pb.AK47CameinNtf)
	msg.Userinfo = t.coinRoleMsg(userid)
	t.broadcast(msg)
}

// 召唤机器人
func (t *Desk) loadRobot() {
	t.robotTime++
	// if t.robotTime < 5 {
	// 	return
	// }
	// t.robotTime = 0
	t.callRobot()
}

// 随机人机数
func (t *Desk) randRobotNum() int {
	return utils.RandMN(t.Game.AK47.Single_Robot[0], t.Game.AK47.Single_Robot[1])
}

// 获取随机人机表情
func (t *Desk) getEmoji() []byte {
	emoji := table.GetTables().RobotEmojiBaseTable.Get()
	if emoji == nil {
		return []byte{}
	}

	var choices []utils.Choice
	for i, v := range emoji.RobotTypeWeight {
		choices = append(choices, utils.Choice{Weight: int(v), Item: i})
	}
	c, _ := utils.WeightedChoice(choices)
	i := c.Item.(int)
	typ := emoji.RobotType[i]

	emoji_config := table.GetTables().RobotEmojiConfigTable.Get(typ)
	// emoji_config := emoji.FindEmojiConfig(typ)
	data, err := json.Marshal(emoji_config)
	if err != nil {
		glog.Errorf("emoji marshal err %v", err)
		return []byte{}
	}
	return data
}

func (t *Desk) callRobot() {
	//桌子满了不召唤
	if len(t.roles) >= int(t.Game.Count) {
		return
	}

	r, n := t.roleCountNum()
	// 没有真人不召唤||有2个以上真人也不召唤
	if r == 0 || r >= 2 {
		return
	}
	n += t.robotCalling      // 正在召唤中机器人
	n -= len(t.robotLeaving) // 离开中机器人

	// 初始化机器人数
	if t.robotNum <= 0 {
		t.robotNum = t.randRobotNum()
		// t.robotTime = 0
	}

	var callRobot int
	if n == t.robotNum {
		return
	} else if n > t.robotNum {
		// 踢出多余的机器人
		kicks := n - t.robotNum
		var robots []string
		for k, v := range t.roles {
			if v.Robot {
				robots = append(robots, k)
			}
		}
		for i := 0; i < kicks && len(robots) > 0; i++ {
			ridx := utils.RandIntN(len(robots))
			robotid := robots[ridx]
			robots = append(robots[:ridx], robots[ridx+1:]...)
			// t.notifyGateUserLeft(robotid, pb.OK, 0)
			// t.userLeaveDesk(robotid)

			if t.robotLeaving == nil {
				t.robotLeaving = make(map[string]bool)
			}
			t.robotLeaving[robotid] = true
			ntf := &pb.RobotLeaveNtf{RobotId: robotid}
			t.send2userid(robotid, ntf)
		}
		return
	} else {
		// 召唤机器人
		callRobot = t.robotNum - n
	}

	callRobot = int(math.Min(float64(callRobot), float64(int(t.Game.Count)-len(t.roles))))
	if callRobot > 0 {
		glog.Infof("hua %s robot num: %d, call: %d", t.Rid, t.robotNum, callRobot)
		msg := new(pb.RobotMsg)
		msg.Gameid = t.Game.Id
		msg.Roomid = t.DeskData.Rid
		msg.Code = t.DeskData.Code
		msg.Rtype = t.DeskData.Rtype
		msg.Ltype = t.DeskData.Ltype
		msg.Gtype = t.DeskData.Gtype
		msg.EnvBet = int32(t.DeskData.Multiple)
		msg.Min = int32(t.Game.Min_Access)
		msg.Max = int32(t.Game.Max_Access)
		msg.Num = uint32(callRobot)
		msg.LimitTime = utils.BsonNow().UnixMilli() + ReadyTime*1000
		msg.Emoji = t.getEmoji() //随机人机表情
		t.dbmsPid.Tell(msg)

		t.robotCalling += callRobot
	}
}

func (t *Desk) roleCountNum() (r, n int) { //r真人 n 机器人
	for _, v := range t.roles {
		if v.User.GetRobot() {
			n++
		} else {
			r++
		}
	}
	return
}

func (t *Desk) roleCountNumNoWatch() (r, n int) { //r真人 n 机器人
	for _, v := range t.seats {
		if v.Watch {
			continue
		}
		if role, ok := t.roles[v.Userid]; ok {
			if role.GetRobot() {
				n++
			} else {
				r++
			}
		}
	}
	return
}

// 位置上玩家数据
func (t *Desk) coinRoleMsg(userid string) (msg *pb.AK47RoomUser) {
	if v, ok := t.roles[userid]; ok {
		if v.Seat == 0 {
			return //没有坐下不广播
		}
		msg = handler.PackAK47CoinUser(v.User)
		msg.Seat = v.Seat
		msg.Offline = v.Offline
		if val, ok := t.seats[v.Seat]; ok {
			msg.Dealer = val.BeDealer
			msg.Bet = val.Bet
			msg.Num = val.DealerN
			msg.Niu = val.Niu
			msg.Ready = val.Ready
			msg.Watch = val.Watch
			if t.DeskAct != nil && !val.Watch {
				if val, ok := t.DeskAct.ActSeats[v.Seat]; ok {
					msg.Pack = val.Pack
					msg.Lose = val.Lose
					msg.See = val.See
					msg.Bet2 = val.Bet
					msg.Totalbet = val.ActNum
				}
			}
		}

		if t.DeskPriv != nil {
			msg.Score = t.DeskPriv.PrivScore[userid]
		}
	}
	return
}

// 所有坐下玩家数据
func (t *Desk) coinSeatBetsMsg(userid string) (msg []*pb.AK47RoomUser) {
	for k, v := range t.seats {
		msg2 := t.coinRoleMsg(v.Userid)
		if msg2 == nil {
			continue
		}
		//自己手牌
		if v.Userid == userid && t.isSee(k) {
			msg2.Cards = t.getHandCards(k)
			msg2.Changecards = t.getHandChangeCards(k)
			msg2.Typ = algo.HuaType(msg2.Changecards)
		}
		msg = append(msg, msg2)
	}
	return
}

// 玩家下注数据
func (t *Desk) coinBetsMsg() (msg []*pb.AK47RoomBets) {
	for k, v := range t.seats {
		msg2 := &pb.AK47RoomBets{
			Seat: k,
			Bets: v.Bet,
		}
		msg = append(msg, msg2)
	}
	return
}

// 是否全部发牌动画播放完成
// func (t *Desk) allReady2() bool {
// 	var num1 int = t.ready2Num() //播放完发牌动画人数
// 	var num2 int = t.readyNum()  //游戏人数
// 	return num1 == num2
// }

func (t *Desk) allReady2() bool {
	for _, v := range t.seats {
		if !v.Ready { //跳过未参与玩家
			continue
		}
		if !v.Ready2 {
			return false
		}
	}
	return true
}

//.

// '是否全部准备状态
func (t *Desk) allReady() bool {
	var num int = t.readyNum()
	if num != len(t.roles) {
		return false
	}
	//准备人数大于2
	if num < 2 {
		return false
	}
	//全部准备立即开始
	return true
}

//.

//' 超时操作

// 准备超时,不等待全部准备
func (t *Desk) readyTimeout() {
	var num int = t.readyNum()
	if num >= 2 {
		t.gameStart() //开始牌局
		return
	}
	//房间人数为0时解散
	t.checkPubOver()
}

// 播放完发牌动画人数
func (t *Desk) ready2Num() (num int) {
	for _, v := range t.seats {
		if v.Ready2 {
			num++
		}
	}
	return
}

// 准备人数（游戏中）
func (t *Desk) readyNum() (num int) {
	for _, v := range t.seats {
		if v.Ready {
			num++
		}
	}
	return
}

// 房间内人数（不一定准备了）
func (t *Desk) roleNum() (num int) {
	for range t.seats {
		num++
	}
	return
}

// 充值超时
func (t *Desk) chargeTimeout() {
	if t.state != int32(pb.STATE_CHARGE) {
		return
	}
	t.state = int32(pb.STATE_BET)

	seatid := t.ActSeat
	userid := t.getUserid(seatid)
	t.coinFold(userid)
}

// 操作超时放弃
func (t *Desk) betTimeout() {
	if t.remainPlayerNum() > 1 {
		if t.DeskAct.ActState&int32(pb.ACT_REPLY_BI) == int32(pb.ACT_REPLY_BI) { //比牌操作超时
			seat := t.DeskAct.ActSeat
			userid := t.getUserid(seat)
			t.coinReplyBi(userid, false)
		} else { //下注操作超时
			seat := t.DeskAct.ActSeat
			userid := t.getUserid(seat)
			t.coinFold(userid)
			t.seatTimeout(userid)
		}
	}
}

func (t *Desk) seatTimeout(userid string) {
	seatid := t.getSeatid(userid)
	seat := t.getSeat(seatid)
	if seat == nil {
		return
	}
	seat.Timeout = true

	role := t.getRole(userid)
	if role == nil {
		return
	}
	role.TimeoutCount++
}

func (t *Desk) clearTimeout(userid string) {
	seatid := t.getSeatid(userid)
	seat := t.getSeat(seatid)
	if seat == nil {
		return
	}
	//如果该轮没有超时，则重置超时次数
	if !seat.Timeout {
		role := t.getRole(userid)
		if role == nil {
			return
		}
		role.TimeoutCount = 0
	}
}

// 检查泡沫状态牌型
func (t *Desk) CheckFoamStateCardType() (ok bool, cardType int) {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return false, 0
	}

	role := t.getRole(userid)
	if role == nil {
		return false, 0
	}

	if role.User.State == 4 {
		return true, t.Game.AK47.NewbieMode.FoamCardType
	} else {
		return false, 0
	}
}

// 检查特定局
func (t *Desk) CheckSpecialRound(role *data.DeskRole) (ok bool, cardType int) {
	//判断是否是新手
	if role.User.State != 1 {
		return false, 0
	}

	round := role.Round + 1

	var idx int
	for k, v := range t.Game.AK47.NewbieMode.SpecialRound {
		if int(round) <= v {
			idx = k
			ok = true
			break
		}
	}

	if ok {
		return ok, t.Game.AK47.NewbieMode.SpecialRoundCardType[idx]
	} else {
		return false, 0
	}
}

// 获取新手模式牌型
func (t *Desk) getNewbieModeCardType() int {
	for _, v := range t.seats {
		if !v.Ready {
			continue
		}
		if r, ok := t.roles[v.Userid]; ok {
			if !r.Robot {
				ok, cardType := t.CheckSpecialRound(r) //特定局
				if ok {
					return cardType
				}

				if r.State == 3 {
					diamond := r.OutDiamond
					index := 0
					for i, v := range t.Game.AK47.NewbieMode.CivilianCanWithdrawRange {
						if diamond < int64(v) {
							index = i
							break
						} else {
							index = i + 1
						}
					}
					if index > len(t.Game.AK47.NewbieMode.CivilianCanWithdrawRange)-1 {
						index = len(t.Game.AK47.NewbieMode.CivilianCanWithdrawRange) - 1
					}
					winRate := t.Game.AK47.NewbieMode.CivilianWinRate[index]
					win := utils.RandWan(int32(winRate))
					if win {
						weight := t.Game.AK47.NewbieMode.CivilianWinWeight
						var choices []utils.Choice
						for i, v := range weight {
							choices = append(choices, utils.Choice{Weight: v, Item: i})
						}
						c, _ := utils.WeightedChoice(choices)
						i := c.Item.(int)
						return t.Game.AK47.NewbieMode.CivilianWinCardType[i]
					} else {
						weight := t.Game.AK47.NewbieMode.CivilianLoseWeight
						var choices []utils.Choice
						for i, v := range weight {
							choices = append(choices, utils.Choice{Weight: v, Item: i})
						}
						c, _ := utils.WeightedChoice(choices)
						i := c.Item.(int)
						return t.Game.AK47.NewbieMode.CivilianLoseCardType[i]
					}
				} else {
					diamond := r.OutDiamond
					index := 0
					for i, v := range t.Game.AK47.NewbieMode.CanWithdrawRange {
						if diamond < int64(v) {
							index = i
							break
						} else {
							index = i + 1
						}
					}
					if index > len(t.Game.AK47.NewbieMode.CanWithdrawRange)-1 {
						index = len(t.Game.AK47.NewbieMode.CanWithdrawRange) - 1
					}
					winRate := t.Game.AK47.NewbieMode.WinRate[index]
					win := utils.RandWan(int32(winRate))
					if win {
						weight := t.Game.AK47.NewbieMode.WinWeight
						var choices []utils.Choice
						for i, v := range weight {
							choices = append(choices, utils.Choice{Weight: v, Item: i})
						}
						c, _ := utils.WeightedChoice(choices)
						i := c.Item.(int)
						return t.Game.AK47.NewbieMode.WinCardType[i]
					} else {
						weight := t.Game.AK47.NewbieMode.LoseWeight
						var choices []utils.Choice
						for i, v := range weight {
							choices = append(choices, utils.Choice{Weight: v, Item: i})
						}
						c, _ := utils.WeightedChoice(choices)
						i := c.Item.(int)
						return t.Game.AK47.NewbieMode.LoseCardType[i]
					}
				}

			}
		}
	}
	return 11
}

// 局中换牌判断
func (t *Desk) IsRoleTriggerBetChange(role *data.DeskRole, winscore1 int64, winscore2 int64) bool {
	money := role.GetMoney()
	winscore := winscore1 + winscore2

	for _, v := range t.Game.AK47.ControlType {
		min := v.ChargeRange[0]
		max := v.ChargeRange[1]
		if max == -1 {
			max = int(math.Inf(1))
		}

		if int(money) >= min && int(money) < max {
			if winscore > int64(v.WinScoreLimit) {
				t.winScore1 = int(winscore1)
				t.winScore2 = int(winscore2)
				t.chargeMoney = int(money)
				return true
			} else {
				return false
			}
		}
	}

	return false
}

// 开局换牌判断
func (t *Desk) IsRoleTriggerStartChange(role *data.DeskRole) bool {
	money := role.GetMoney()               //总充值
	winscore := role.GetWinScore()         //当前赢分
	factor := handler.GetFactor(role.User) //玩家系数

	controlType := 0

	for _, v := range t.Game.AK47.ControlType {
		min := v.ChargeRange[0]
		max := v.ChargeRange[1]
		if max == -1 {
			max = int(math.Inf(1))
		}

		if int(money) >= min && int(money) < max {
			controlType = v.ControlType
			if controlType == 1 {
				idx := utils.GetRangeIndex(v.WinScoreConfig, int(winscore))
				weight := v.WinScoreWeight[idx]
				if utils.RandWan(int32(weight)) {
					t.controlType = controlType
					t.winScore = int(winscore)
					t.chargeMoney = int(money)
					return true
				}
			} else if controlType == 2 {
				idx := utils.GetRangeIndex(v.PlayerFactorRange, int(factor*100))
				weight := v.PlayerFactorWeight[idx]
				if utils.RandWan(int32(weight)) {
					t.controlType = controlType
					t.playerFactor = int(factor * 100)
					t.chargeMoney = int(money)
					return true
				}
			} else {
				return false
			}
		}
	}

	return false
}

// 是否触发开局换牌
func (t *Desk) IsTriggerStartChange() bool {

	for _, v := range t.seats {
		if !v.Ready {
			continue
		}
		if r, ok := t.roles[v.Userid]; ok {
			if !r.Robot {
				if t.IsRoleTriggerStartChange(r) {
					return true
				}
			}
		}
	}
	return false
}

// 获取平均系数
func (t *Desk) getAverageFactor() float64 {
	var total float64
	var num int
	for _, v := range t.seats {
		if !v.Ready {
			continue
		}
		if r, ok := t.roles[v.Userid]; ok {
			if !r.Robot {
				total += handler.GetFactor(r.User)
				num++
			}
		}
	}
	ret := total / float64(num)
	if ret < 0 {
		ret = 0
	} else if ret > 2 {
		ret = 2
	}
	return ret
}

// 获取最终系数
func (t *Desk) getFinalFactor() float64 {

	var ret float64

	if t.Game.AK47.RoomFactorSwitch == 0 {
		//玩家平均系数
		averageFactor := t.getAverageFactor()
		ret = 2 - averageFactor
		glog.Debugf("平均系数: %#v", averageFactor)
	} else {
		//房间系数
		msg := &pb.GetRoomFactor{}
		msg.GameId = t.Game.Id

		res := t.reqRoom(msg)
		var response *pb.GetedRoomFactor
		var ok bool

		if response, ok = res.(*pb.GetedRoomFactor); !ok {
			glog.Errorf("get  room factor failed: %#v", res)
		}

		//玩家平均系数
		averageFactor := t.getAverageFactor()

		glog.Debugf("房间系数：%#v, 平均系数: %#v", response.Factor, averageFactor)

		ret = (2 - averageFactor + response.Factor) / 2
	}

	if ret < 0 {
		ret = 0
	} else if ret > 2 {
		ret = 2
	}
	return ret

}

// 获取点控系数
func (t *Desk) getPointControlFactor() int {
	for _, v := range t.roles {
		if !v.Robot && v.PCSwitch {
			return int(v.PCFactor)
		}
	}
	return int(t.getFinalFactor() * 100)
}

// 开始游戏
func (t *Desk) gameStart() {
	//未参与玩家离开位置
	// t.startSitup()
	//抽水
	// t.drawfee()
	//初始化
	t.gameStartInit()
	//洗牌
	t.shuffle()
	//发牌
	t.deal()
	//胜率检测
	t.winRateCheck()
	//开局换牌处理
	t.StartChangeHandler()
	//策略局处理
	t.StrategyHandler()
	//冤家牌
	t.hedgeCard()
	//详情初始化
	t.detailInit()
	//扣底注
	t.raiseAnte()
	//选庄
	t.dealerHandler()
	//机器人处理
	t.robotHandler()

}

// 参与游戏的机器人数目
func (t *Desk) getRobotNum() int {
	count := 0
	for _, v := range t.seats {
		if !v.Ready {
			continue
		}
		if role, ok := t.roles[v.Userid]; ok && role != nil {
			if role.Robot {
				count++
			}
		}
	}
	return count
}

// 获取最大手牌玩家
func (t *Desk) getMaxPlayer() uint32 {
	var max uint32 = 0
	for k, v := range t.seats {
		if k == max {
			continue
		}
		if !v.Ready {
			continue
		}
		if role, ok := t.roles[v.Userid]; ok && role != nil {
			if role.Robot {
				continue
			}
		}
		if max == 0 {
			max = k
			continue
		}
		cs1 := t.getHandCards(k)
		cs2 := t.getHandCards(max)

		cs3 := t.getHandChangeCards(k)
		cs4 := t.getHandChangeCards(max)

		if algo.AK47Compare(cs1, cs2, cs3, cs4) { //k赢
			max = k
		}
	}
	return max
}

// 获取最大手牌位置(可能是人机可能是玩家)
func (t *Desk) getMaxSeat() uint32 {
	var max uint32 = 0

	for k, v := range t.seats {
		if k == max {
			continue
		}
		if !v.Ready {
			continue
		}
		if !t.isAlive(k) {
			continue
		}
		if max == 0 {
			max = k
			continue
		}
		cs1 := t.getHandCards(k)
		cs2 := t.getHandCards(max)

		cs3 := t.getHandChangeCards(k)
		cs4 := t.getHandChangeCards(max)

		if algo.AK47Compare(cs1, cs2, cs3, cs4) { //k赢
			max = k
		}
	}
	return max
}

// 获取核心人机
func (t *Desk) GetCoreRobot() uint32 {
	var robots []uint32
	for k, v := range t.seats {
		if !v.Ready {
			continue
		}
		if !t.isAlive(k) {
			continue
		}
		if r, ok := t.roles[v.Userid]; ok && v != nil {
			if r.Robot {
				robots = append(robots, k)
			}
		}
	}
	if len(robots) == 0 {
		return 0
	} else {
		return robots[utils.RandIntN(len(robots))]
	}
}

// 获取所有人机位置
func (t *Desk) GetAllRobot() []uint32 {
	var robots []uint32
	for k, v := range t.seats {
		if !v.Ready {
			continue
		}
		if !t.isAlive(k) {
			continue
		}
		if r, ok := t.roles[v.Userid]; ok && v != nil {
			if r.Robot {
				robots = append(robots, k)
			}
		}
	}
	return robots
}

// 机器人处理
func (t *Desk) robotHandler() {
	if t.getRobotNum() == 0 {
		return
	}
	max := t.getMaxPlayer()
	_ = max

	id := t.cardTypeId
	ct := t.Game.AK47.FindCardType(id)
	rsg := t.Game.AK47.FindRobotStrategyGroup(ct.RobotStrategyGroup)

	for k, v := range t.seats {
		if k == max {
			continue
		}
		if !v.Ready {
			continue
		}
		if role, ok := t.roles[v.Userid]; ok && role != nil {
			if role.Robot {
				if v.Core || v.SeeAndPack || v.Identity != "" {
					continue
				} else {
					cs1 := t.getHandCards(k)
					cs2 := t.getHandCards(max)
					cs3 := t.getHandChangeCards(k)
					cs4 := t.getHandChangeCards(max)

					typ := algo.HuaType(cs3)

					var rs data.AK47GameRobotStrategy

					if algo.AK47Compare(cs1, cs2, cs3, cs4) {
						switch typ {
						case algo.BaoZi:
							rs = t.Game.AK47.FindRobotStrategy(rsg.BaoZiBigger)
						case algo.TongHuaShun:
							rs = t.Game.AK47.FindRobotStrategy(rsg.TongHuaShunBigger)
						case algo.ShunZi:
							rs = t.Game.AK47.FindRobotStrategy(rsg.ShunZiBigger)
						case algo.TongHua:
							rs = t.Game.AK47.FindRobotStrategy(rsg.TongHuaBigger)
						case algo.DuiZi:
							rs = t.Game.AK47.FindRobotStrategy(rsg.DuiZiBigger)
						case algo.GaoPai:
							rs = t.Game.AK47.FindRobotStrategy(rsg.GaoPaiBigger)
						}
					} else {
						switch typ {
						case algo.BaoZi:
							rs = t.Game.AK47.FindRobotStrategy(rsg.BaoZiSmaller)
						case algo.TongHuaShun:
							rs = t.Game.AK47.FindRobotStrategy(rsg.TongHuaShunSmaller)
						case algo.ShunZi:
							rs = t.Game.AK47.FindRobotStrategy(rsg.ShunZiSmaller)
						case algo.TongHua:
							rs = t.Game.AK47.FindRobotStrategy(rsg.TongHuaSmaller)
						case algo.DuiZi:
							rs = t.Game.AK47.FindRobotStrategy(rsg.DuiZiSmaller)
						case algo.GaoPai:
							rs = t.Game.AK47.FindRobotStrategy(rsg.GaoPaiSmaller)
						}
					}

					msg := &pb.AK47CoinRobotStrategyNtf{}
					msg.Id = rs.Id
					for _, v := range rs.ActionWeight {
						msg.ActionWeight = append(msg.ActionWeight, &pb.AK47CoinActionWeight{Values: v})
					}
					msg.ActionTime = rs.ActionTime
					msg.SeeWeight = rs.SeeWeight
					msg.AgreeBi = rs.AgreeBi
					for _, v := range rs.SeeTime {
						msg.SeeTime = append(msg.SeeTime, &pb.AK47CoinSeeTime{Values: v})
					}
					t.send2userid(v.Userid, msg)
				}
			}
		}
	}
}

// 设置所有人机看牌即弃
func (t *Desk) SetAllRobotSeeAndPack(bet_change bool) {
	for _, v := range t.GetAllRobot() {
		var cards, ccards []uint32
		cards, ccards, t.DeskGame.Cards = algo.AK47GetCard([]uint32{}, algo.GaoPai, t.DeskGame.Cards)
		t.SetSeeAndPackRobot(v, cards, ccards, bet_change)
	}
}

// 设置核心人机
func (t *Desk) SetCoreRobot(seatid uint32, cards []uint32, ccards []uint32, bet_change bool) {
	seat := t.getSeat(seatid)
	if seat == nil {
		return
	}

	seat.Core = true
	if len(cards) > 0 {
		seat.Cards = cards
		seat.ChangeCards = ccards
	}

	msg := &pb.AK47CoinRobotStrategyNtf{}
	msg.Core = true
	if bet_change {
		msg.BetChange = true
		msg.ChangeCards = cards
		msg.ChangeCcards = ccards

	}
	t.send2userid(seat.Userid, msg)
}

// 设置看牌即弃人机
func (t *Desk) SetSeeAndPackRobot(seatid uint32, cards []uint32, ccards []uint32, bet_change bool) {
	seat := t.getSeat(seatid)
	if seat == nil {
		return
	}

	seat.SeeAndPack = true
	if len(cards) > 0 {
		seat.Cards = cards
		seat.ChangeCards = ccards
	}

	msg := &pb.AK47CoinRobotStrategyNtf{}
	msg.SeeAndPack = true
	if bet_change {
		msg.BetChange = true
		msg.ChangeCards = cards
		msg.ChangeCcards = ccards
	}
	t.send2userid(seat.Userid, msg)

}

// 通知所有主机僚机
func (t *Desk) NotifyAllMasterAndSlave() {
	for _, v := range t.GetAllRobot() {
		seat := t.getSeat(v)
		if seat == nil {
			continue
		}
		if seat.Identity == "" { //不是主机也不是僚机
			continue
		}
		msg := &pb.AK47CoinRobotStrategyNtf{}
		msg.Identity = seat.Identity
		t.send2userid(seat.Userid, msg)
	}
}

// 局中检测赢分
func (t *Desk) BetCheckWinScore(seatid uint32) {
	// if t.DeskType == int32(pb.DESK_TYPE_NEWBIEW) { //跳过新手
	// 	return
	// }

	if t.isStrategy { //策略局，跳过局中换牌
		return
	}

	if t.changeCardType == 1 || t.changeCardType == 2 { //已经开局换牌或已经局中换牌，则跳过
		return
	}

	if t.isRobot(seatid) { //跳过检测机器人
		return
	}

	seat := t.getSeat(seatid)
	status := t.getStatus(seatid)
	if seat == nil || status == nil {
		return
	}

	total := t.DeskGame.BetNum //总投注
	bet := status.ActNum       //下注额
	winscore1 := total - bet   //当局赢分

	role := t.getRole(seat.Userid)
	if role == nil {
		return
	}

	winscore2 := int64(role.GetWinScore()) //当前赢分

	if !t.IsRoleTriggerBetChange(role, winscore1, winscore2) {
		return
	}

	max := t.getMaxSeat()
	if t.isRobot(max) {
		t.SetCoreRobot(max, []uint32{}, []uint32{}, false) //不需要换牌，但是需要标记人机为核心人机
		return
	}

	core := t.GetCoreRobot()
	if core == 0 {
		return
	}

	maxCards := t.getHandCards(max)
	cmaxCards := t.getHandChangeCards(max)

	var bigger, cbigger []uint32

	bigger, cbigger, t.DeskGame.LaiCards, t.DeskGame.Cards = algo.GetAK47MaxCard(maxCards, cmaxCards, t.DeskGame.LaiCards, t.DeskGame.Cards)

	if len(bigger) != 0 {
		t.SetCoreRobot(core, bigger, cbigger, true)
	} else {
		t.SetAllRobotSeeAndPack(true)
	}
	t.changeCardType = 2
}

// 开局换牌处理
func (t *Desk) StartChangeHandler() {
	// if t.DeskType == int32(pb.DESK_TYPE_NEWBIEW) { //跳过新手
	// 	return
	// }

	if !t.IsTriggerStartChange() {
		return
	}

	max := t.getMaxSeat()
	if t.isRobot(max) { //如果最大牌是人机，则跳过
		t.SetCoreRobot(max, []uint32{}, []uint32{}, false) //不需要换牌，但是需要标记人机为核心人机
		return
	}

	core := t.GetCoreRobot()
	if core == 0 { //没有人机参与游戏，则跳过
		return
	}

	maxCards := t.getHandCards(max)
	cmaxCards := t.getHandChangeCards(max)

	var bigger, cbigger []uint32

	bigger, cbigger, t.DeskGame.LaiCards, t.DeskGame.Cards = algo.GetAK47MaxCard(maxCards, cmaxCards, t.DeskGame.LaiCards, t.DeskGame.Cards)

	if len(bigger) != 0 {
		t.SetCoreRobot(core, bigger, cbigger, false)
	} else {
		t.SetAllRobotSeeAndPack(false)
	}

	t.changeCardType = 1
}

// 获取玩家状态
func (t *Desk) GetPlayerState() int {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return 0
	}

	role := t.getRole(userid)
	if role == nil {
		return 0
	}

	return role.User.State
}

// 100策略AAA诱导首充
func (t *Desk) Strategy100Handler(ignoreState bool) {
	state := t.GetPlayerState()
	if state != 3 && !ignoreState { //平民状态走100
		return
	}

	//触发概率
	if !utils.RandWan(int32(t.Game.AK47.NewbieMode.CivilianTriggerStrategyProb)) {
		return
	}

	//配置
	// roundConfig := 15                 //累计局数
	scoreConfig := []int{5000, 15000} //携带分范围
	// TriggerProb := 5000               //触发概率
	// TiggerTimes := []int{1, 1, 1}     //触发次数

	//只有0.5底分和1.0底分触发
	if t.Ante != 50 && t.Ante != 100 {
		return
	}

	userid := t.GetOnlyOnePlayer()
	if userid == "" { //没有真人或真人不唯一，跳过
		return
	}

	seatid := t.getSeatid(userid)
	seat := t.getSeat(seatid)
	role := t.getRole(userid)
	if role == nil || seat == nil {
		return
	}

	//c类不触发
	if role.RegistArea == 2 {
		return
	}

	//判断是否触发过
	if role.Strategy100Flag {
		return
	}

	//判断底注
	if role.Diamond >= int64(t.Ante)*300 {
		return
	}

	//判断累计局数
	// if role.TPTotalRound < int32(t.Game.AK47.Strategy100Round) {
	// 	return
	// }

	//判断携带分范围
	if role.Diamond < int64(scoreConfig[0]) || role.Diamond > int64(scoreConfig[1]) {
		return
	}

	//设置局内充值
	if t.Ante == 50 {
		if role.Diamond >= 50*100 && role.Diamond < 80*100 {
			t.chargeAmount = 200 * 100
		} else if role.Diamond >= 80*100 && role.Diamond < 120*100 {
			t.chargeAmount = 300 * 100
		} else if role.Diamond >= 120*100 && role.Diamond <= 150*100 {
			t.chargeAmount = 500 * 100
		}
	} else if t.Ante == 100 {
		if role.Diamond >= 50*100 && role.Diamond < 80*100 {
			t.chargeAmount = 500 * 100
		} else if role.Diamond >= 80*100 && role.Diamond < 120*100 {
			t.chargeAmount = 800 * 100
		} else if role.Diamond >= 120*100 && role.Diamond <= 150*100 {
			t.chargeAmount = 1000 * 100
		}
	}

	//判断注册天数
	// days := role.GetRegistDays()
	// if days > len(TiggerTimes)-1 {
	// 	return false
	// }

	//判断触发次数
	// times := TiggerTimes[days]
	// if int(role.TPTiggerTimes) >= times {
	// 	return false
	// }

	//判断触发概率
	// if !utils.RandWan(int32(TriggerProb)) {
	// 	return false
	// }

	//判断人机数目
	_, n := t.roleCountNum()
	if n < 2 { //人机数目小于2人所有人机看牌即弃
		t.SetAllRobotSeeAndPack(false)
		return
	}

	//选取主机ab
	a, b := t.SelectAB()
	if a == 0 || b == 0 {
		t.SetAllRobotSeeAndPack(false)
		return
	}
	//选取僚机
	c, d := t.SelectCD()

	//重新洗牌
	t.shuffle()

	var wildCardsNum int
	var cardtype uint32

	var choices []utils.Choice
	//给玩家发牌 AAA
	choices = append(choices, utils.Choice{Weight: 20, Item: 1})
	choices = append(choices, utils.Choice{Weight: 50, Item: 2})
	choices = append(choices, utils.Choice{Weight: 30, Item: 3})
	ch, _ := utils.WeightedChoice(choices)
	wildCardsNum = ch.Item.(int)

	cardtype = algo.BaoZi1

	seat.Cards, seat.ChangeCards, t.DeskGame.Cards = t.getCard2(t.DeskGame.Cards, wildCardsNum, cardtype)

	//给主机a发牌 KKK-222
	seata := t.getSeat(a)

	choices = []utils.Choice{}
	choices = append(choices, utils.Choice{Weight: 50, Item: 1})
	choices = append(choices, utils.Choice{Weight: 50, Item: 2})
	ch, _ = utils.WeightedChoice(choices)
	wildCardsNum = ch.Item.(int)

	choices = []utils.Choice{}
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.BaoZi2})
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.BaoZi3})
	ch, _ = utils.WeightedChoice(choices)
	cardtype = ch.Item.(uint32)

	seata.Cards, seata.ChangeCards, t.DeskGame.Cards = t.getCard2(t.DeskGame.Cards, wildCardsNum, cardtype)

	//给主机b发牌 同花顺或顺子
	seatb := t.getSeat(b)

	choices = []utils.Choice{}
	choices = append(choices, utils.Choice{Weight: 30, Item: 0})
	choices = append(choices, utils.Choice{Weight: 70, Item: 1})
	ch, _ = utils.WeightedChoice(choices)
	wildCardsNum = ch.Item.(int)

	choices = []utils.Choice{}
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.TongHuaShun})
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.ShunZi})
	ch, _ = utils.WeightedChoice(choices)
	cardtype = ch.Item.(uint32)

	seatb.Cards, seatb.ChangeCards, t.DeskGame.Cards = t.getCard2(t.DeskGame.Cards, wildCardsNum, cardtype)

	//给僚机c发牌
	if c != 0 {
		seatc := t.getSeat(c)

		choices = []utils.Choice{}
		choices = append(choices, utils.Choice{Weight: 30, Item: 0})
		choices = append(choices, utils.Choice{Weight: 70, Item: 1})
		ch, _ = utils.WeightedChoice(choices)
		wildCardsNum = ch.Item.(int)

		if wildCardsNum == 0 { //对子 高牌
			choices = []utils.Choice{}
			choices = append(choices, utils.Choice{Weight: 70, Item: algo.DuiZi})
			choices = append(choices, utils.Choice{Weight: 30, Item: algo.GaoPai})
			ch, _ = utils.WeightedChoice(choices)
			cardtype = ch.Item.(uint32)

		} else if wildCardsNum == 1 { //同花 对子
			choices = []utils.Choice{}
			choices = append(choices, utils.Choice{Weight: 50, Item: algo.TongHua})
			choices = append(choices, utils.Choice{Weight: 50, Item: algo.DuiZi})
			ch, _ = utils.WeightedChoice(choices)
			cardtype = ch.Item.(uint32)
		}

		seatc.Cards, seatc.ChangeCards, t.DeskGame.Cards = t.getCard2(t.DeskGame.Cards, wildCardsNum, cardtype)

	}

	//给僚机d发牌 对子/高牌
	if d != 0 {
		seatd := t.getSeat(d)

		choices = []utils.Choice{}
		choices = append(choices, utils.Choice{Weight: 30, Item: 0})
		choices = append(choices, utils.Choice{Weight: 70, Item: 1})
		ch, _ = utils.WeightedChoice(choices)
		wildCardsNum = ch.Item.(int)

		if wildCardsNum == 0 { //对子 高牌
			choices = []utils.Choice{}
			choices = append(choices, utils.Choice{Weight: 70, Item: algo.DuiZi})
			choices = append(choices, utils.Choice{Weight: 30, Item: algo.GaoPai})
			ch, _ = utils.WeightedChoice(choices)
			cardtype = ch.Item.(uint32)

		} else if wildCardsNum == 1 { //同花 对子
			choices = []utils.Choice{}
			choices = append(choices, utils.Choice{Weight: 50, Item: algo.TongHua})
			choices = append(choices, utils.Choice{Weight: 50, Item: algo.DuiZi})
			ch, _ = utils.WeightedChoice(choices)
			cardtype = ch.Item.(uint32)
		}

		seatd.Cards, seatd.ChangeCards, t.DeskGame.Cards = t.getCard2(t.DeskGame.Cards, wildCardsNum, cardtype)
	}

	//通知所有主机僚机
	t.NotifyAllMasterAndSlave()

	//设置100标记
	role.SetStrategy100()
	//同步数据
	msg := &pb.TriggerStategy100{}
	t.send2userid(userid, msg)
	//增加触发次数
	// role.IncreTPTiggerTimes()
	//同步数据
	// msg := &pb.TPTriggeTimes{}
	// t.send2userid(userid, msg)

	//重置累计局数
	// role.ResetTPTotalRound()
	//同步数据
	// msg1 := &pb.TPTotalRound{}
	// msg1.Isreset = true
	// t.send2userid(userid, msg1)

	//设置本局为策略局
	t.isStrategy = true
	t.strategyType = 100

}

// 200策略大牌诱导充值
func (t *Desk) Strategy200Handler() {
	userid := t.GetOnlyOnePlayer()
	if userid == "" { //没有真人或真人不唯一，跳过
		return
	}

	seatid := t.getSeatid(userid)
	seat := t.getSeat(seatid)
	role := t.getRole(userid)
	if role == nil || seat == nil {
		return
	}

	// A类的新手桌不触发200
	if role.IsA() && t.Dtype == int32(pb.DESK_TYPE_NEWBIEW) {
		return
	}

	//c类不触发
	if role.RegistArea == 2 {
		return
	}

	if role.RegistArea == 1 {
		// 非平民 携带分超过30
		if role.State != 3 || role.Diamond > 3000 {
			return
		}

		if role.LoginTimes > 2 {
			return
		}

		if role.Novice200StrategyCount >= 2 && role.LoginTimes == 1 {
			return
		} else if role.Novice200StrategyCount >= 1 && role.LoginTimes == 2 {
			return
		}

		if !utils.RandWan(3000) {
			return
		}
	} else {
		if !role.Strategy100Flag {
			t.Strategy100Handler(true)
			return
		}

		state := t.GetPlayerState()
		if state != 4 { //泡沫状态走200
			return
		}

		//触发概率
		if !utils.RandWan(int32(t.Game.AK47.NewbieMode.FoamTriggerStrategyProb)) {
			return
		}

	}

	// 配置
	// roundConfig := 20               //累计局数
	scoreConfig := []int{5000, 15000} //携带分范围
	// TriggerProb := 5000             //触发概率

	//只有0.5底分和1.0底分触发
	if t.Ante != 50 && t.Ante != 100 {
		return
	}

	//判断是否触发过
	if role.Strategy200Flag {
		return
	}

	//判断底注
	if role.Diamond >= int64(t.Ante)*300 {
		return
	}

	// 判断累计局数
	// if role.TPTotalRound < int32(t.Game.AK47.Strategy200Round) {
	// 	return
	// }

	// 判断携带分范围
	if role.Diamond < int64(scoreConfig[0]) || role.Diamond > int64(scoreConfig[1]) {
		return
	}

	// 设置局内充值
	if t.Ante == 50 {
		if role.Diamond >= 50*100 && role.Diamond < 80*100 {
			t.chargeAmount = 200 * 100
		} else if role.Diamond >= 80*100 && role.Diamond < 120*100 {
			t.chargeAmount = 300 * 100
		} else if role.Diamond >= 120*100 && role.Diamond <= 150*100 {
			t.chargeAmount = 500 * 100
		}
	} else if t.Ante == 100 {
		if role.Diamond >= 50*100 && role.Diamond < 80*100 {
			t.chargeAmount = 500 * 100
		} else if role.Diamond >= 80*100 && role.Diamond < 120*100 {
			t.chargeAmount = 800 * 100
		} else if role.Diamond >= 120*100 && role.Diamond <= 150*100 {
			t.chargeAmount = 1000 * 100
		}
	}
	// 判断触发概率
	// if !utils.RandWan(int32(TriggerProb)) {
	// 	return
	// }

	// 判断人机数目
	_, n := t.roleCountNum()
	if n < 2 { //人机数目小于2人所有人机看牌即弃
		t.SetAllRobotSeeAndPack(false)
		return
	}

	// 选取主机ab
	a, b := t.SelectAB()
	if a == 0 || b == 0 {
		t.SetAllRobotSeeAndPack(false)
		return
	}
	// 选取僚机
	c, d := t.SelectCD()

	// 重新洗牌
	t.shuffle()

	var wildCardsNum int
	var cardtype uint32

	var choices []utils.Choice
	// 给玩家发牌 AAA-888
	choices = append(choices, utils.Choice{Weight: 50, Item: 1})
	choices = append(choices, utils.Choice{Weight: 50, Item: 2})
	ch, _ := utils.WeightedChoice(choices)
	wildCardsNum = ch.Item.(int)

	choices = []utils.Choice{}
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.BaoZi1})
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.BaoZi2})
	ch, _ = utils.WeightedChoice(choices)
	cardtype = ch.Item.(uint32)

	seat.Cards, seat.ChangeCards, t.DeskGame.Cards = t.getCard2(t.DeskGame.Cards, wildCardsNum, cardtype)

	// 给主机a发牌 AAA-888
	seata := t.getSeat(a)
	choices = []utils.Choice{}
	choices = append(choices, utils.Choice{Weight: 50, Item: 1})
	choices = append(choices, utils.Choice{Weight: 50, Item: 2})
	ch, _ = utils.WeightedChoice(choices)
	wildCardsNum = ch.Item.(int)

	choices = []utils.Choice{}
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.BaoZi1})
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.BaoZi2})
	ch, _ = utils.WeightedChoice(choices)
	cardtype = ch.Item.(uint32)

	seata.Cards, seata.ChangeCards, t.DeskGame.Cards = t.getCard2(t.DeskGame.Cards, wildCardsNum, cardtype)

	//主机a一定比玩家小
	if algo.AK47Compare(seata.Cards, seat.Cards, seata.ChangeCards, seat.ChangeCards) {
		seata.Cards, seat.Cards = seat.Cards, seata.Cards
		seata.ChangeCards, seat.ChangeCards = seat.ChangeCards, seata.ChangeCards
	}

	// 给主机b发牌 顺子/同花
	seatb := t.getSeat(b)
	choices = []utils.Choice{}
	choices = append(choices, utils.Choice{Weight: 30, Item: 0})
	choices = append(choices, utils.Choice{Weight: 70, Item: 1})
	ch, _ = utils.WeightedChoice(choices)
	wildCardsNum = ch.Item.(int)

	choices = []utils.Choice{}
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.TongHuaShun})
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.ShunZi})
	ch, _ = utils.WeightedChoice(choices)
	cardtype = ch.Item.(uint32)

	seatb.Cards, seatb.ChangeCards, t.DeskGame.Cards = t.getCard2(t.DeskGame.Cards, wildCardsNum, cardtype)

	// 给僚机c发牌 对子/高牌
	if c != 0 {
		seatc := t.getSeat(c)
		choices = []utils.Choice{}
		choices = append(choices, utils.Choice{Weight: 30, Item: 0})
		choices = append(choices, utils.Choice{Weight: 70, Item: 1})
		ch, _ = utils.WeightedChoice(choices)
		wildCardsNum = ch.Item.(int)

		if wildCardsNum == 0 { //对子 高牌
			choices = []utils.Choice{}
			choices = append(choices, utils.Choice{Weight: 70, Item: algo.DuiZi})
			choices = append(choices, utils.Choice{Weight: 30, Item: algo.GaoPai})
			ch, _ = utils.WeightedChoice(choices)
			cardtype = ch.Item.(uint32)

		} else if wildCardsNum == 1 { //同花 对子
			choices = []utils.Choice{}
			choices = append(choices, utils.Choice{Weight: 50, Item: algo.TongHua})
			choices = append(choices, utils.Choice{Weight: 50, Item: algo.DuiZi})
			ch, _ = utils.WeightedChoice(choices)
			cardtype = ch.Item.(uint32)
		}

		seatc.Cards, seatc.ChangeCards, t.DeskGame.Cards = t.getCard2(t.DeskGame.Cards, wildCardsNum, cardtype)
	}

	// 给僚机d发牌 对子/高牌
	if d != 0 {
		seatd := t.getSeat(d)
		choices = []utils.Choice{}
		choices = append(choices, utils.Choice{Weight: 30, Item: 0})
		choices = append(choices, utils.Choice{Weight: 70, Item: 1})
		ch, _ = utils.WeightedChoice(choices)
		wildCardsNum = ch.Item.(int)

		if wildCardsNum == 0 { //对子 高牌
			choices = []utils.Choice{}
			choices = append(choices, utils.Choice{Weight: 70, Item: algo.DuiZi})
			choices = append(choices, utils.Choice{Weight: 30, Item: algo.GaoPai})
			ch, _ = utils.WeightedChoice(choices)
			cardtype = ch.Item.(uint32)

		} else if wildCardsNum == 1 { //同花 对子
			choices = []utils.Choice{}
			choices = append(choices, utils.Choice{Weight: 50, Item: algo.TongHua})
			choices = append(choices, utils.Choice{Weight: 50, Item: algo.DuiZi})
			ch, _ = utils.WeightedChoice(choices)
			cardtype = ch.Item.(uint32)
		}

		seatd.Cards, seatd.ChangeCards, t.DeskGame.Cards = t.getCard2(t.DeskGame.Cards, wildCardsNum, cardtype)
	}
	// 通知所有主机僚机
	t.NotifyAllMasterAndSlave()

	//设置200标记
	role.SetStrategy200()
	//同步数据
	msg := &pb.TriggerStategy200{}
	t.send2userid(userid, msg)
	// // 重置累计局数
	// role.ResetTPTotalRound()
	// // 同步数据
	// msg1 := &pb.TPTotalRound{}
	// msg1.Isreset = true
	// t.send2userid(userid, msg1)

	// 设置本局为策略局
	t.isStrategy = true
	t.strategyType = 200
}

// 300策略大牌诱导充值
func (t *Desk) Strategy300Handler() {
	if t.Game.AK47.Strategy300Switch == 0 {
		return
	}

	state := t.GetPlayerState()
	if state != 2 { //正常状态走300
		return
	}

	// 配置
	// roundConfig := 20               //累计局数
	scoreConfig := []int{0, 100000} //携带分范围
	TriggerProb := 5000             //触发概率

	userid := t.GetOnlyOnePlayer()
	if userid == "" { //没有真人或真人不唯一，跳过
		return
	}

	seatid := t.getSeatid(userid)
	seat := t.getSeat(seatid)
	role := t.getRole(userid)
	if role == nil || seat == nil {
		return
	}

	//c类不触发
	if role.RegistArea == 2 {
		return
	}

	//判断底注
	if role.Diamond >= int64(t.Ante)*300 {
		return
	}

	// 判断累计局数
	if role.TPTotalRound < int32(t.Game.AK47.Strategy200Round) {
		return
	}

	// 判断携带分范围
	if role.Diamond < int64(scoreConfig[0]) || role.Diamond > int64(scoreConfig[1]) {
		return
	}

	// 判断触发概率
	if !utils.RandWan(int32(TriggerProb)) {
		return
	}

	// 判断人机数目
	_, n := t.roleCountNum()
	if n < 2 { //人机数目小于2人所有人机看牌即弃
		t.SetAllRobotSeeAndPack(false)
		return
	}

	// 选取主机ab
	a, b := t.SelectAB()
	if a == 0 || b == 0 {
		t.SetAllRobotSeeAndPack(false)
		return
	}
	// 选取僚机
	c, d := t.SelectCD()

	// 重新洗牌
	t.shuffle()

	var wildCardsNum int
	var cardtype uint32

	var choices []utils.Choice
	// 给玩家发牌 AAA-888
	choices = append(choices, utils.Choice{Weight: 50, Item: 1})
	choices = append(choices, utils.Choice{Weight: 50, Item: 2})
	ch, _ := utils.WeightedChoice(choices)
	wildCardsNum = ch.Item.(int)

	choices = []utils.Choice{}
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.BaoZi1})
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.BaoZi2})
	ch, _ = utils.WeightedChoice(choices)
	cardtype = ch.Item.(uint32)

	seat.Cards, seat.ChangeCards, t.DeskGame.Cards = t.getCard2(t.DeskGame.Cards, wildCardsNum, cardtype)

	// 给主机a发牌 AAA-888
	seata := t.getSeat(a)
	choices = []utils.Choice{}
	choices = append(choices, utils.Choice{Weight: 50, Item: 1})
	choices = append(choices, utils.Choice{Weight: 50, Item: 2})
	ch, _ = utils.WeightedChoice(choices)
	wildCardsNum = ch.Item.(int)

	choices = []utils.Choice{}
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.BaoZi1})
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.BaoZi2})
	ch, _ = utils.WeightedChoice(choices)
	cardtype = ch.Item.(uint32)

	seata.Cards, seata.ChangeCards, t.DeskGame.Cards = t.getCard2(t.DeskGame.Cards, wildCardsNum, cardtype)

	if utils.RandWan(6000) { //主机a比玩家大
		if !algo.AK47Compare(seata.Cards, seat.Cards, seata.ChangeCards, seat.ChangeCards) {
			seata.Cards, seat.Cards = seat.Cards, seata.Cards
			seata.ChangeCards, seat.ChangeCards = seat.ChangeCards, seata.ChangeCards
		}
	} else { //主机a比玩家小
		if algo.AK47Compare(seata.Cards, seat.Cards, seata.ChangeCards, seat.ChangeCards) {
			seata.Cards, seat.Cards = seat.Cards, seata.Cards
			seata.ChangeCards, seat.ChangeCards = seat.ChangeCards, seata.ChangeCards
		}
	}

	if algo.AK47Compare(seata.Cards, seat.Cards, seata.ChangeCards, seat.ChangeCards) {
		t.strategy300Bigger = true
	}

	// 给主机b发牌 顺子/同花
	seatb := t.getSeat(b)
	choices = []utils.Choice{}
	choices = append(choices, utils.Choice{Weight: 30, Item: 0})
	choices = append(choices, utils.Choice{Weight: 70, Item: 1})
	ch, _ = utils.WeightedChoice(choices)
	wildCardsNum = ch.Item.(int)

	choices = []utils.Choice{}
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.TongHuaShun})
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.ShunZi})
	ch, _ = utils.WeightedChoice(choices)
	cardtype = ch.Item.(uint32)

	seatb.Cards, seatb.ChangeCards, t.DeskGame.Cards = t.getCard2(t.DeskGame.Cards, wildCardsNum, cardtype)

	// 给僚机c发牌 对子/高牌
	if c != 0 {
		seatc := t.getSeat(c)
		choices = []utils.Choice{}
		choices = append(choices, utils.Choice{Weight: 30, Item: 0})
		choices = append(choices, utils.Choice{Weight: 70, Item: 1})
		ch, _ = utils.WeightedChoice(choices)
		wildCardsNum = ch.Item.(int)

		if wildCardsNum == 0 { //对子 高牌
			choices = []utils.Choice{}
			choices = append(choices, utils.Choice{Weight: 70, Item: algo.DuiZi})
			choices = append(choices, utils.Choice{Weight: 30, Item: algo.GaoPai})
			ch, _ = utils.WeightedChoice(choices)
			cardtype = ch.Item.(uint32)

		} else if wildCardsNum == 1 { //同花 对子
			choices = []utils.Choice{}
			choices = append(choices, utils.Choice{Weight: 50, Item: algo.TongHua})
			choices = append(choices, utils.Choice{Weight: 50, Item: algo.DuiZi})
			ch, _ = utils.WeightedChoice(choices)
			cardtype = ch.Item.(uint32)
		}

		seatc.Cards, seatc.ChangeCards, t.DeskGame.Cards = t.getCard2(t.DeskGame.Cards, wildCardsNum, cardtype)
	}

	// 给僚机d发牌 对子/高牌
	if d != 0 {
		seatd := t.getSeat(d)
		choices = []utils.Choice{}
		choices = append(choices, utils.Choice{Weight: 30, Item: 0})
		choices = append(choices, utils.Choice{Weight: 70, Item: 1})
		ch, _ = utils.WeightedChoice(choices)
		wildCardsNum = ch.Item.(int)

		if wildCardsNum == 0 { //对子 高牌
			choices = []utils.Choice{}
			choices = append(choices, utils.Choice{Weight: 70, Item: algo.DuiZi})
			choices = append(choices, utils.Choice{Weight: 30, Item: algo.GaoPai})
			ch, _ = utils.WeightedChoice(choices)
			cardtype = ch.Item.(uint32)

		} else if wildCardsNum == 1 { //同花 对子
			choices = []utils.Choice{}
			choices = append(choices, utils.Choice{Weight: 50, Item: algo.TongHua})
			choices = append(choices, utils.Choice{Weight: 50, Item: algo.DuiZi})
			ch, _ = utils.WeightedChoice(choices)
			cardtype = ch.Item.(uint32)
		}

		seatd.Cards, seatd.ChangeCards, t.DeskGame.Cards = t.getCard2(t.DeskGame.Cards, wildCardsNum, cardtype)
	}
	// 通知所有主机僚机
	t.NotifyAllMasterAndSlave()

	// 重置累计局数
	role.ResetTPTotalRound()
	// 同步数据
	msg1 := &pb.TPTotalRound{}
	msg1.Isreset = true
	t.send2userid(userid, msg1)

	// 设置本局为策略局
	t.isStrategy = true
	t.strategyType = 300
}

// 策略局处理
func (t *Desk) StrategyHandler() {
	_, n := t.roleCountNum()
	if n == 0 { //没有机器人不走策略局
		return
	}
	if t.changeCardType == 1 { //开局换牌不走策略局
		return
	}

	// if t.DeskType == int32(pb.DESK_TYPE_NEWBIEW) { //新手100策略
	// 	t.Strategy100Handler()
	// } else if t.DeskType == int32(pb.DESK_TYPE_NORMAL) { //正常200策略
	// 	t.Strategy200Handler()
	// }
	// if t.DeskType == int32(pb.DESK_TYPE_NEWBIEW) {
	// 	t.Strategy100Handler()
	// } else if t.DeskType == int32(pb.DESK_TYPE_NORMAL) { //正常桌
	// 	if !t.Strategy100Handler() { //如果没走100，则走200
	// 		t.Strategy200Handler()
	// 	}
	// }

	if t.DeskType == int32(pb.DESK_TYPE_NEWBIEW) { //新手桌 (玩家处于新手状态、平民状态属于新手桌)
		t.Strategy100Handler(false)
		t.Strategy200Handler()
	} else {
		t.Strategy200Handler()
		t.Strategy300Handler()
	}
}

// 选主机AB位置
func (t *Desk) SelectAB() (a uint32, b uint32) {
	var robots []uint32
	for k, v := range t.seats {
		if !v.Ready {
			continue
		}
		if !t.isAlive(k) {
			continue
		}
		if r, ok := t.roles[v.Userid]; ok && v != nil {
			if r.Robot && v.Identity == "" {
				robots = append(robots, k)
			}
		}
	}

	if len(robots) < 2 { //主机必须有两个
		return 0, 0
	} else {
		algo.Shuffle(robots) //随机两个位置为主机
		a = robots[0]
		b = robots[1]
		seata := t.getSeat(a)
		seatb := t.getSeat(b)
		seata.Identity = "a"
		seatb.Identity = "b"
		return
	}
}

// 选僚机CD位置
func (t *Desk) SelectCD() (c uint32, d uint32) {
	var robots []uint32
	for k, v := range t.seats {
		if !v.Ready {
			continue
		}
		if !t.isAlive(k) {
			continue
		}
		if r, ok := t.roles[v.Userid]; ok && v != nil {
			if r.Robot && v.Identity == "" {
				robots = append(robots, k)
			}
		}
	}

	if len(robots) == 0 { //没有僚机
		return 0, 0
	} else if len(robots) == 1 { //只有僚机c
		c = robots[0]
		d = 0
		seatc := t.getSeat(c)
		seatc.Identity = "c"
		return
	} else { //有僚机c和d
		algo.Shuffle(robots) //随机一下位置
		c = robots[0]
		d = robots[1]
		seatc := t.getSeat(c)
		seatd := t.getSeat(d)
		seatc.Identity = "c"
		seatd.Identity = "d"
		return
	}
}

// 获取一个机器人
func (t *Desk) GetOneRobot() string {
	userids := []string{}
	for k, v := range t.roles {
		if v.User.GetRobot() {
			userids = append(userids, k)
		}
	}

	l := len(userids)
	if l == 0 {
		return ""
	}

	return userids[utils.RandIntN(l)]
}

// 获取唯一玩家
func (t *Desk) GetOnlyOnePlayer() string {
	userids := []string{}
	for k, v := range t.roles {
		if !v.User.GetRobot() {
			userids = append(userids, k)
		}
	}

	if len(userids) == 1 {
		return userids[0]
	} else {
		return ""
	}
}

// 开始初始化
func (t *Desk) gameStartInit() {
	//设置房间游戏状态
	// t.timer = 0
	// t.state = int32(pb.STATE_GAME)
	t.gameInit()
	nodePid.Tell(&pb.GameStart{Rid: t.Rid}) // 通知node节点游戏开始了
}

// 详情初始化
func (t *Desk) detailInit() {
	t.detail = &data.Detail{}
	t.detail.BeginTime = time.Now().Unix()
	t.detail.Gtype = t.Gtype
	t.detail.RoomId = t.Game.Id
	t.detail.DeskId = t.Rid
	t.detail.WaterId = t.GameId
	t.detail.CardTypeId = t.cardTypeId

	r, _ := t.roleCountNumNoWatch()
	t.detail.RealGame = r > 1 // 2个以上真人玩家

	var players []string
	for _, v := range t.seats {
		if v.Ready {
			players = append(players, v.Userid)
		}
	}
	t.detail.Players = strings.Join(players, ",")
	for k, v := range t.seats {
		if !v.Ready {
			continue
		}
		ak47_detail := &data.AK47Detail{}
		ak47_detail.SeatId = k
		ak47_detail.UserId = v.Userid
		// ak47_detail.Cards = t.getHandCards(k)
		// ak47_detail.ChangeCards = t.getHandChangeCards(k)

		if role, ok := t.roles[v.Userid]; ok && role != nil {
			ak47_detail.BeforeScore = role.GetScore()
			ak47_detail.BeforeCash = role.GetDiamond()
			ak47_detail.BeforeBonus = role.GetCoin()
		}
		t.detail.AK47Detail = append(t.detail.AK47Detail, ak47_detail)
	}
}

// 初始化
func (t *Desk) gameInit() {
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0), //自由
		int32(pb.ROOM_TYPE1): //私人
		t.DeskGame.GameId = handler.GenGameId(t.Gtype) //初始化流水号
		// 新手配置
		userid := t.GetOnlyOnePlayer()
		if userid != "" {
			role := t.roles[userid]
			if role.RegistArea == 0 {
				// 切换为A类新手配置
				t.Game.AK47.NewbieMode = t.Game.AK47.ANewbieMode
			}
		}
		t.Ante = uint32(t.Game.AK47.Bottom)
		t.score = make(map[uint32]int64)
		t.over = make(map[uint32]*pb.AK47CoinOver)
		t.betInfo = make(map[int64]int32)

		t.DeskAct = new(data.DeskAct)
		t.DeskAct.ActSeats = make(map[uint32]*data.ActStatus)
		t.DeskAct.ActRechargeTimes = make(map[uint32]int)
		t.DeskAct.ActCallNum = 0
		t.DeskAct.ActRaiseNum = 0
		t.DeskAct.ActTimes = 0
		for _, v := range t.seats {
			v.Ready = true //自动准备
			//v.Ready = false
			v.BeDealer = 0
			v.DealerN = 0
			v.Bet = 0
			v.Cards = make([]uint32, 0)
			v.WinRate = 0
			v.ChangeCards = make([]uint32, 0)
			v.ShowCards = make([]uint32, 0)
			v.ShowChangeCards = make([]uint32, 0)
			v.Power = 0
			v.Niu = false
			v.Watch = false
			if role, ok := t.roles[v.Userid]; ok && role != nil {
				role.Ratio = role.GetRatio()
			}
		}
		//初始化ActStatus
		for k, v := range t.seats {
			if !v.Ready {
				continue
			}
			d := new(data.ActStatus)
			d.Alive = true
			t.DeskAct.ActSeats[k] = d
			if !t.isRobot(k) {
				d.Dirty = true //真人默认就是污染的
			}
		}
		t.cardTypeId = t.findCardTypeId()

	case int32(pb.ROOM_TYPE2): //百人
	}
}

// 结束重置
func (t *Desk) gameOverInit() {
	for _, v := range t.seats {
		t.clearTimeout(v.Userid)
		v.Ready = false
		v.Ready2 = false
		// v.Watch = false
		v.Core = false
		v.SeeAndPack = false
		v.Identity = ""
		v.Timeout = false
	}

	for _, v := range t.DeskAct.ActSeats {
		v.See = false
		v.Alive = false
		v.Lose = false
		v.Pack = false
		v.Dirty = false
		v.Bet = 0
		v.ActNum = 0
	}
	//重新加载配置
	t.Game = config.GetGame(t.Game.Id)
	t.Ante = uint32(t.Game.AK47.Bottom)

	t.DeskGame.Cards = make([]uint32, 0)
	t.DeskGame.BetNum = 0
	t.robotNum = t.randRobotNum()
	t.robotTime = 0
	t.robotCalling = 0
	t.robotLeaving = make(map[string]bool)
	t.changeCardType = 0
	t.controlType = 0
	t.winScore = 0
	t.playerFactor = 0
	t.winScore1 = 0
	t.winScore2 = 0
	t.chargeMoney = 0
	t.isStrategy = false
	t.isCharge = false
	t.condition = 0
	t.strategyType = 0
	t.cmpRecord = [][2]uint32{}
	t.chargeAmount = 0
	t.strategy300Bigger = false

	t.FakeSeats = handler.CreateFakeSeats(t.robotNum, t.seats, t.DeskData.Count) //重新生成假位置
	t.broadFakeSeat()                                                            //广播
	// t.gameInit()
	//t.timer = 0
	//结算加长5秒
	// t.timer = -5
	//播放结算动画
	t.pauseGame(PAUSE_REASON_4, 5, nil)
	// timerChan := time.After(5 * time.Second)
	// <-timerChan
	// t.state = int32(pb.STATE_READY) //设置房间状态
	// t.pushState()
}

// 未参与玩家离开位置
func (t *Desk) startSitup() {
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0), //自由
		int32(pb.ROOM_TYPE1): //私人
		for _, v := range t.seats {
			if v.Ready {
				continue
			}
			//玩家站起
			t.roleSitUp(v.Userid)
		}
	}
}

//.

// ' 打庄处理
func (t *Desk) dealerHandler() {
	//选择庄家
	t.dealer4()
	glog.Debugf("dealer -> %s", t.DeskGame.Dealer)
	glog.Debugf("dealer seat -> %d", t.DeskGame.DealerSeat)
	// t.timer = 0
	t.state = int32(pb.STATE_DEALING) //切换为发牌状态

	t.pushDealer() //广播庄家
	t.pushState()  //广播状态

}

// 状态消息
func (t *Desk) pushState() {
	msg := &pb.AK47PushStateNtf{
		State: t.state,
	}
	t.broadcast(msg)
}

// 庄家消息
func (t *Desk) pushDealer() {
	msg := &pb.AK47PushDealerNtf{
		DealerSeat: t.DeskGame.DealerSeat,
	}

	for k, v := range t.seats {
		if v.Ready {
			msg.Seats = append(msg.Seats, k)
		}
	}

	t.broadcast(msg)
}

// 设置局内充值条件
func (t *Desk) SetChargeCondition() {

	if !t.isStrategy { //不是策略局，跳过
		return
	}
	userid := t.GetOnlyOnePlayer()
	if userid == "" { //没有真人或真人不唯一，跳过
		return
	}
	seatid := t.getSeatid(userid)
	status := t.getStatus(seatid)
	role := t.getRole(userid)
	if status == nil || role == nil {
		return
	}

	if t.isCharge { //玩家充值
		ante_factor := t.Game.AK47.AnteFactor
		// upper_factor := 1000

		total := t.DeskGame.BetNum
		bet := status.ActNum
		winscore := total - bet

		if winscore < int64(t.Ante)*int64(ante_factor) {
			t.condition = 1
		} else if winscore >= int64(t.Ante)*int64(ante_factor) {
			t.condition = 2
		}

	} else { //玩家未充值
		score := role.User.GetScore()
		if !status.Alive { //玩家已弃牌
			t.condition = 4
		} else if score >= t.ActAnte*2 { //加注后可以下注
			t.condition = 1
		} else if score >= t.ActAnte && score < t.ActAnte*2 { //跟注后可以下注，加注后不可以下注
			t.condition = 2
		} else if score < t.ActAnte {
			t.condition = 3
		}
	}

}

// 玩家是否是blind
func (t *Desk) isPlayerBlind() bool {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return false
	}

	seatid := t.getSeatid(userid)
	status := t.getStatus(seatid)
	if status == nil {
		return false
	}

	return !status.See
}

// 场上人数是不是仅剩两家
func (t *Desk) isOnlyTwo() bool {
	return t.remainPlayerNum() == 2
}

// 是否是对子以上
func (t *Desk) isAbovePair(seat uint32) bool {
	cards := t.getHandChangeCards(seat)
	typ := algo.HuaType(cards)
	return typ > algo.DuiZi
}

// 广播操作状态消息
func (t *Desk) pushActState() {
	prev := t.getPrevActSeat()
	curr := t.DeskAct.ActSeat
	next := t.getNextActSeat()
	msg := &pb.AK47PushActStateNtf{
		State:         t.DeskAct.ActState,
		Seat:          t.DeskAct.ActSeat,
		Timer:         BetTime,
		Ante:          t.DeskAct.ActAnte,
		Turn:          t.DeskAct.ActTimes,
		IsPrevPlayer:  !t.isRobot(prev),
		Bigger:        t.strategy300Bigger,
		IsChaalOper:   t.isChaal(curr),
		IsPlayerBlind: t.isPlayerBlind(),
		IsOnlyTwo:     t.isOnlyTwo(),
		IsAbovePair:   t.isAbovePair(curr),
		NextSeat:      next,
		// Pot:      t.DeskGame.BetNum,
		// CallNum:  t.DeskHua.ActCallNum,
		// RaiseNum: t.DeskHua.ActRaiseNum,
	}

	if t.isStrategy {
		msg.IsCharge = t.isCharge
		msg.Condition = int32(t.condition)
	}

	t.broadcast(msg)
}

//.

// ' 随机选择庄
func (t *Desk) dealer4() {
	// if t.DeskGame.DealerSeat != 0 {
	// 	if v, ok := t.seats[t.DeskGame.DealerSeat]; ok {
	// 		//庄家位置在游戏中
	// 		if v.Ready {
	// 			return
	// 		}
	// 	}
	// }
	a := make([]uint32, 0)
	for k, v := range t.seats {
		//准备的人才是参与者
		if v.Ready {
			a = append(a, k)
		}
	}
	if len(a) == 0 {
		glog.Errorf("dealer4 err: %#v", t.seats)
		return
	}
	seat := a[0]
	// seat := a[rand.Intn(len(a))]
	if val, ok := t.seats[seat]; ok {
		t.DeskGame.Dealer = val.Userid
		t.DeskGame.DealerSeat = seat
	}
}

// 查找牌型ID
func (t *Desk) findCardTypeId() int {
	var factor_int int

	for _, role := range t.roles {
		if role.IsA() {
			// A类玩家指定牌型
			return t.Game.AK47.ACardType
		}
	}

	if t.DeskType == int32(pb.DESK_TYPE_NORMAL) {
		ok, cardType := t.CheckFoamStateCardType()
		if ok {
			return cardType
		}

		factor := t.getFinalFactor()
		factor_int = int(factor * 100)
		glog.Debugf("getFinalFactor:%d", factor_int)
	} else if t.DeskType == int32(pb.DESK_TYPE_POINTCONTROL) {
		factor_int = t.getPointControlFactor()
		glog.Debugf("getPointControlFactor:%d", factor_int)
		if factor_int > 200 {
			factor_int = 200
		} else if factor_int < 0 {
			factor_int = 0
		}
	} else if t.DeskType == int32(pb.DESK_TYPE_NEWBIEW) {
		return t.getNewbieModeCardType()
	}
	glog.Debugf("最终系数: %d", factor_int)
	index := 0
	for i, v := range t.Game.AK47.FinalFactorRange {
		if factor_int < v {
			index = i
			break
		} else {
			index = i + 1
		}
	}
	return t.Game.AK47.CardTypeRange[index]
}

func (t *Desk) getCard2(current_cards []uint32, wildCardsNum int, cardtype uint32) (cards, change_cards, remain_cards []uint32) {
	wildCards := []uint32{}
	for i := 0; i < wildCardsNum; i++ {
		wildCard := t.DeskGame.LaiCards[0]
		t.DeskGame.LaiCards = t.DeskGame.LaiCards[1:]
		wildCards = append(wildCards, wildCard)
	}

	return algo.AK47GetCard(wildCards, cardtype, current_cards)
}

// 生成牌
func (t *Desk) getCard(robot bool, current_cards []uint32) (cards, change_cards, remain_cards []uint32) {
	id := t.cardTypeId
	// id := 11
	glog.Debugf("===============当前牌型id: %d, 房间ID: %s=================", id, t.Game.Id)

	ct := t.Game.AK47.FindCardType(id)

	//先随机多少张万能牌
	var choices []utils.Choice
	if !robot {
		choices = append(choices, utils.Choice{Weight: ct.PlayerWildCardWeight[0], Item: 0})
		choices = append(choices, utils.Choice{Weight: ct.PlayerWildCardWeight[1], Item: 1})
		choices = append(choices, utils.Choice{Weight: ct.PlayerWildCardWeight[2], Item: 2})
		choices = append(choices, utils.Choice{Weight: ct.PlayerWildCardWeight[3], Item: 3})
	} else {
		choices = append(choices, utils.Choice{Weight: ct.RobotWildCardWeight[0], Item: 0})
		choices = append(choices, utils.Choice{Weight: ct.RobotWildCardWeight[1], Item: 1})
		choices = append(choices, utils.Choice{Weight: ct.RobotWildCardWeight[2], Item: 2})
		choices = append(choices, utils.Choice{Weight: ct.RobotWildCardWeight[3], Item: 3})
	}

	c, _ := utils.WeightedChoice(choices)
	wildCardsNum := c.Item.(int)

	//发万能牌
	wildCards := []uint32{}
	for i := 0; i < int(wildCardsNum); i++ {
		wildCard := t.DeskGame.LaiCards[0]
		t.DeskGame.LaiCards = t.DeskGame.LaiCards[1:]
		wildCards = append(wildCards, wildCard)
	}

	var cardtype uint32
	//再确认牌型
	if len(wildCards) == 0 {
		var choices []utils.Choice
		if !robot {
			choices = append(choices, utils.Choice{Weight: ct.Player0WildCardWeight[0], Item: algo.BaoZi})
			choices = append(choices, utils.Choice{Weight: ct.Player0WildCardWeight[1], Item: algo.TongHuaShun})
			choices = append(choices, utils.Choice{Weight: ct.Player0WildCardWeight[2], Item: algo.ShunZi})
			choices = append(choices, utils.Choice{Weight: ct.Player0WildCardWeight[3], Item: algo.TongHua})
			choices = append(choices, utils.Choice{Weight: ct.Player0WildCardWeight[4], Item: algo.DuiZi})
			choices = append(choices, utils.Choice{Weight: ct.Player0WildCardWeight[5], Item: algo.GaoPai})
		} else {
			choices = append(choices, utils.Choice{Weight: ct.Robot0WildCardWeight[0], Item: algo.BaoZi})
			choices = append(choices, utils.Choice{Weight: ct.Robot0WildCardWeight[1], Item: algo.TongHuaShun})
			choices = append(choices, utils.Choice{Weight: ct.Robot0WildCardWeight[2], Item: algo.ShunZi})
			choices = append(choices, utils.Choice{Weight: ct.Robot0WildCardWeight[3], Item: algo.TongHua})
			choices = append(choices, utils.Choice{Weight: ct.Robot0WildCardWeight[4], Item: algo.DuiZi})
			choices = append(choices, utils.Choice{Weight: ct.Robot0WildCardWeight[5], Item: algo.GaoPai})
		}
		c, _ := utils.WeightedChoice(choices)
		cardtype = c.Item.(uint32)
	} else if len(wildCards) == 1 {
		var choices []utils.Choice
		if !robot {
			choices = append(choices, utils.Choice{Weight: ct.Player1WildCardWeight[0], Item: algo.BaoZi})       //A
			choices = append(choices, utils.Choice{Weight: ct.Player1WildCardWeight[1], Item: algo.TongHuaShun}) //B
			choices = append(choices, utils.Choice{Weight: ct.Player1WildCardWeight[2], Item: algo.ShunZi})      //C
			choices = append(choices, utils.Choice{Weight: ct.Player1WildCardWeight[3], Item: algo.TongHua})     //D
			choices = append(choices, utils.Choice{Weight: ct.Player1WildCardWeight[4], Item: algo.DuiZi})       //E
		} else {
			choices = append(choices, utils.Choice{Weight: ct.Robot1WildCardWeight[0], Item: algo.BaoZi})       //A
			choices = append(choices, utils.Choice{Weight: ct.Robot1WildCardWeight[1], Item: algo.TongHuaShun}) //B
			choices = append(choices, utils.Choice{Weight: ct.Robot1WildCardWeight[2], Item: algo.ShunZi})      //C
			choices = append(choices, utils.Choice{Weight: ct.Robot1WildCardWeight[3], Item: algo.TongHua})     //D
			choices = append(choices, utils.Choice{Weight: ct.Robot1WildCardWeight[4], Item: algo.DuiZi})       //E
		}
		c, _ := utils.WeightedChoice(choices)
		cardtype = c.Item.(uint32)
	}

	return algo.AK47GetCard(wildCards, cardtype, current_cards)

}

// 胜率检测
func (t *Desk) winRateCheck() {
	id := t.cardTypeId
	ct := t.Game.AK47.FindCardType(id)
	if ct.WinRateCheck == 1 {
		r, n := t.roleCountNum()
		if r != 1 { //真人数目不是一个，跳过
			return
		}
		if n == 0 { //没有人机，跳过
			return
		}
		if utils.RandWan(int32(ct.PlayerWinRate)) { //玩家赢
			max := t.getMaxSeat()
			if !t.isRobot(max) { //如果最大玩家是真人，玩家赢，跳过
				return
			}

			//找唯一玩家
			userid := t.GetOnlyOnePlayer()
			if userid == "" {
				return
			}

			//获取seat
			seatid := t.getSeatid(userid)
			seat := t.getSeat(seatid)
			if seat == nil {
				return
			}
			seat_robot := t.getSeat(max)
			if seat_robot == nil {
				return
			}

			//换牌
			seat.Cards, seat_robot.Cards = seat_robot.Cards, seat.Cards
			seat.ChangeCards, seat_robot.ChangeCards = seat_robot.ChangeCards, seat.ChangeCards
		} else { //玩家输
			max := t.getMaxSeat()
			if t.isRobot(max) { //如果最大玩家是机器人，玩家输，跳过
				return
			}

			// 找一个机器人
			userid := t.GetOneRobot()
			if userid == "" {
				return
			}

			// 获取seat
			seatid := t.getSeatid(userid)
			seat := t.getSeat(seatid)
			if seat == nil {
				return
			}
			seat_player := t.getSeat(max)
			if seat_player == nil {
				return
			}

			// 换牌
			seat.Cards, seat_player.Cards = seat_player.Cards, seat.Cards
			seat.ChangeCards, seat_player.ChangeCards = seat_player.ChangeCards, seat.ChangeCards
		}
	}
}

// '发牌
func (t *Desk) deal() {
	var hand = 3
	for _, v := range t.seats {
		if !v.Ready {
			continue
		}

		v.Cards = make([]uint32, hand, hand)
		v.WinRate = 0
		v.ChangeCards = make([]uint32, hand, hand)
		if r, ok := t.roles[v.Userid]; ok {
			cards, change_cards, remain := t.getCard(r.Robot, t.DeskGame.Cards)
			copy(v.Cards, cards)
			v.WinRate = algo.CaclWinRate(cards)
			copy(v.ChangeCards, change_cards)
			t.DeskGame.Cards = remain
		}

		//准备的人才是参与者
		// copy(v.Cards, t.DeskGame.Cards[:hand])
		// t.DeskGame.Cards = t.DeskGame.Cards[hand:]
		glog.Debugf("deal -> %v", v.Cards)
		//发牌消息
		//看不到牌值
		// cards2 := make([]uint32, hand, hand)
		// msg := resDraw(k, t.state, cards2)
		// t.broadcast(msg)
	}
	//配牌
	// cfg.Reload()
	// for i, v := range t.seats {
	// 	if !v.Ready {
	// 		continue
	// 	}

	// 	name := fmt.Sprintf("card%d", i)
	// 	conf := cfg.Section(nodeName).Key(name).Value()
	// 	card := strings.Split(conf, ",")
	// 	if len(card) != 3 {
	// 		continue
	// 	}

	// 	var uint32Slice []uint32
	// 	for _, c := range card {
	// 		uintValue, err := strconv.ParseUint(c, 16, 32)
	// 		if err != nil {
	// 			continue
	// 		}
	// 		uint32Value := uint32(uintValue)
	// 		uint32Slice = append(uint32Slice, uint32Value)
	// 	}
	// 	copy(v.Cards, uint32Slice)
	// }
}

//.

// '自由场结算消息
func (t *Desk) resCoinOver(score map[uint32]int64) (msg *pb.AK47CoinGameoverNtf) {
	t.state = int32(pb.STATE_OVER)
	msg = &pb.AK47CoinGameoverNtf{
		Dealer: t.DeskGame.Dealer,
		State:  t.state,
	}
	for _, v := range t.over {
		if t.Game.AK47.ShowSwitch == 0 {
			v.Cards = []uint32{}
			v.Changecards = []uint32{}
		} else if !v.Show {
			v.Cards = []uint32{}
			v.Changecards = []uint32{}
		} else {
			seat := t.getSeat(v.Seat)
			if seat != nil {
				showCards := seat.ShowCards
				showChangeCards := seat.ShowChangeCards
				if len(showCards) != 0 && len(showChangeCards) != 0 {
					v.Cards = showCards
					v.Changecards = showChangeCards
				}
			}
		}
		msg.Data = append(msg.Data, v)
	}
	return
}

//.

// '私人局结算消息
func (t *Desk) resOver(score map[uint32]int64) (msg *pb.AK47GameoverNtf) {
	msg = &pb.AK47GameoverNtf{
		Dealer:     t.DeskGame.Dealer,
		DealerSeat: t.DeskGame.DealerSeat,
		Round:      t.DeskGame.Round,
		//LeftRound:  (t.DeskData.Round - t.DeskGame.Round),
	}
	if t.DeskData.Round > t.DeskGame.Round {
		msg.LeftRound = (t.DeskData.Round - t.DeskGame.Round)
	}
	for k, v := range score {
		d := &pb.AK47RoomOver{
			Seat:  k,
			Score: v,
		}
		if val, ok := t.seats[k]; ok {
			d.Bets = val.Bet
			d.Value = val.Power
			d.Cards = val.Cards
			d.Total = t.DeskPriv.PrivScore[val.Userid]
			if p, ok2 := t.roles[val.Userid]; ok2 {
				d.Coin = p.User.GetCoin()
				d.Nickname = p.User.GetNickname()
				d.Photo = p.User.GetPhoto()
				d.VipLv = int32(p.User.Vip.Lv)
			}
		}
		msg.Data = append(msg.Data, d)
	}
	return
}

// 上个操作位置
func (t *Desk) getPrevActSeat() uint32 {
	seat := t.DeskAct.ActSeat
	if seat == 0 {
		seat = t.DeskGame.DealerSeat
	}
	var i uint32 = seat
	for {
		i = t.PrevSeat(i)
		if i == seat {
			break
		}
		if t.qualified(i) {
			return i
		}
	}
	return 0
}

// 切换上个操作位置
func (t *Desk) setPrevActSeat() {
	curr := t.getPrevActSeat()
	t.timer = 0
	t.DeskAct.ActSeat = curr
	t.setPrevActState()
	t.SetChargeCondition()
	t.pushActState()
}

// 下个操作位置
func (t *Desk) getNextActSeat() uint32 {
	seat := t.DeskAct.ActSeat
	if seat == 0 {
		seat = t.DeskGame.DealerSeat
	}
	var i uint32 = seat
	for {
		i = t.nextSeat(i)
		if i == seat {
			break
		}
		if t.qualified(i) {
			return i
		}
	}
	return 0
}

// 全部比牌
func (t *Desk) allBi() {
	curr := t.getNextActSeat()

	seats := []uint32{}
	//消息构造
	msg := new(pb.AK47CoinAllBiNtf)
	for k, v := range t.DeskAct.ActSeats {
		if v.Alive {
			seats = append(seats, k)
			cards := t.getHandCards(k)
			changecards := t.getHandChangeCards(k)
			typ := algo.HuaType(changecards)

			info := &pb.AK47CoinAllBiInfo{
				Seat:        k,
				Cards:       cards,
				Changecards: changecards,
				Type:        typ,
			}
			msg.Info = append(msg.Info, info)
		}
	}

	winner := curr
	for k, v := range t.DeskAct.ActSeats {
		if k == winner {
			continue
		}
		if !v.Alive {
			continue
		}
		cs1 := t.getHandCards(k)
		cs2 := t.getHandCards(winner)
		cs3 := t.getHandChangeCards(k)
		cs4 := t.getHandChangeCards(winner)

		if algo.AK47Compare(cs1, cs2, cs3, cs4) { //k赢
			if val, ok := t.DeskAct.ActSeats[winner]; ok {
				val.Alive = false
				t.Lose(winner)
			}
			winner = k
		} else { //winner赢
			v.Alive = false
			t.Lose(k)
		}
	}
	msg.Winner = winner

	others := t.GetOtherSeats(seats)
	msg2 := new(pb.AK47CoinAllBiNtf)
	utils.Clone(msg2, msg)
	for _, v := range msg2.Info {
		v.Cards = []uint32{}
		v.Changecards = []uint32{}
	}

	t.broadcast4(seats, msg)
	t.broadcast4(others, msg2)
	//等待播放比牌动画
	// timerChan1 := time.After(2 * time.Second)
	// <-timerChan1
	t.pauseGame(PAUSE_REASON_3, 2, winner)
}

func (t *Desk) leaveCheck(seatid uint32) bool {
	if t.remainPlayerNum() == 1 {
		winner := t.getWinner()
		if winner == seatid { //最后赢的玩家不能离开
			return false
		}
	}
	return true
}

// 强制看牌
func (t *Desk) forceSee() {
	if t.DeskAct.ActTimes != 4 {
		return
	}

	for k, v := range t.seats {
		if !v.Ready { //跳过未参与玩家
			continue
		}
		if act, ok := t.DeskAct.ActSeats[k]; ok && act != nil {
			if act.See { //跳过已经see的
				continue
			}
			act.See = true
			msg := new(pb.AK47CoinSeeRsp)
			msg.Cards = v.Cards
			msg.Changecards = v.ChangeCards
			msg.Typ = algo.HuaType(msg.Cards)
			t.send2userid(v.Userid, msg)

			// 广播看牌不用All功能
			msg2 := new(pb.AK47CoinSeeNtf)
			msg2.Seat = k
			msg2.Userid = v.Userid
			msg2.Actseat = t.DeskAct.ActSeat
			msg2.Actstate = uint32(t.DeskAct.ActState)
			t.broadcast(msg2)
		}
	}

	// msg2 := new(pb.AK47CoinSeeNtf)
	// msg2.All = true
	// t.broadcast(msg2)
}

// 切换下个操作位置
func (t *Desk) setNextActSeat() {
	//游戏仅剩1人结束
	if t.remainPlayerNum() == 1 {
		winner := t.getWinner()
		t.gameOver(winner)
		return
	}

	last := t.DeskAct.ActSeat
	curr := t.getNextActSeat()
	//总下注额达到底池上限结束
	if t.DeskGame.BetNum >= int64(t.Game.AK47.Pool_Limit) {
		//等待播放下注动画(最后一个下注的人)
		// timerChan := time.After(2 * time.Second)
		// <-timerChan
		//消息构造
		t.pauseGame(PAUSE_REASON_2, 2, nil)
		return
	}

	//结束,winner = last
	// if curr == 0 {
	// t.gameOver()
	// return
	// }

	//轮数计数
	if t.roundOver(last, curr) {
		t.DeskAct.ActTimes++
		//TODO 底池操作
	}
	//强制看牌判断
	t.forceSee()
	//最大20局
	if t.DeskAct.ActTimes == 20 {
		//TODO 结束限制
		t.pauseGame(PAUSE_REASON_2, 1, nil)
		return
	}
	t.timer = 0
	t.DeskAct.ActSeat = curr

	//检测局中换牌
	t.BetCheckWinScore(t.DeskAct.ActSeat)
	//设置下家操作值
	t.setNextActState()
	//设置局内充值条件
	t.SetChargeCondition()
	//广播下家操作状态消息
	t.pushActState()
}

// 是否结束一轮操作
func (t *Desk) roundOver(last, curr uint32) bool {
	//庄家操作完
	if last == t.DeskGame.DealerSeat {
		return true
	}
	//庄家已经输掉
	i := last
	for {
		i = t.nextSeat(i)
		if i == curr {
			break
		}
		if i == t.DeskGame.DealerSeat {
			return true
		}
	}
	return false
}

// 上家的位置
func (t *Desk) PrevSeat(seat uint32) uint32 {
	prevSeat := seat - 1
	if prevSeat == 0 {
		prevSeat = t.DeskData.Count
	}
	return prevSeat
}

// 下家的位置
func (t *Desk) nextSeat(seat uint32) uint32 {
	nextSeat := seat + 1
	if nextSeat > t.DeskData.Count {
		nextSeat = 1
	}
	return nextSeat
}

// 剩余玩家数
func (t *Desk) remainPlayerNum() int {
	num := 0
	for _, v := range t.DeskAct.ActSeats {
		if v.Alive {
			num++
		}
	}
	return num
}

// 获取胜利者
func (t *Desk) getWinner() (winner uint32) {
	for k, v := range t.DeskAct.ActSeats {
		if v.Alive {
			winner = k
		}
	}
	return
}

// 是否合格,是否已经放弃或者比牌输掉
func (t *Desk) qualified(seat uint32) bool {
	if v, ok := t.DeskAct.ActSeats[seat]; ok {
		if v.Alive {
			return true
		}
	}
	return false
}

// 设置上个位置操作值（仅回复比牌）
func (t *Desk) setPrevActState() {
	var val int32

	val |= int32(pb.ACT_REPLY_BI)

	t.DeskAct.ActState = val
}

// 判断是否存在sideshow
func (t *Desk) CanSideShow(val int32) bool {
	return val&int32(pb.ACT_SIDESHOW) != 0
}

// 判断是否存在show
func (t *Desk) CanShow(val int32) bool {
	return val&int32(pb.ACT_SHOW) != 0
}

// show限制
func (t *Desk) LimitShow(val int32) int32 {
	//自己和上家必须都看牌
	prevSeat := t.getPrevActSeat()
	seat := t.DeskAct.ActSeat
	if v, ok := t.DeskAct.ActSeats[prevSeat]; ok {
		if v1, ok := t.DeskAct.ActSeats[seat]; ok {
			if v.See && v1.See {
				val |= int32(pb.ACT_SIDESHOW)
			}
		}
	}

	//如果是两个人sideshow变为show
	if t.remainPlayerNum() == 2 {
		if t.CanSideShow(val) {
			val ^= int32(pb.ACT_SIDESHOW)
			val |= int32(pb.ACT_SHOW)
		} else {
			val |= int32(pb.ACT_SHOW)
		}
	}

	//前三轮不允许比牌
	if t.DeskAct.ActTimes < 3 && t.remainPlayerNum() > 2 {
		if t.CanSideShow(val) {
			val ^= int32(pb.ACT_SIDESHOW)
		}
		if t.CanShow(val) {
			val ^= int32(pb.ACT_SHOW)
		}

		return val
	} else {
		return val
	}
}

// 设置下个位置操作值
func (t *Desk) setNextActState() {
	var val int32

	//弃牌标记判断
	val |= int32(pb.ACT_PACK) //弃牌操作

	//blind chaal标记判断
	if v, ok := t.DeskAct.ActSeats[t.DeskAct.ActSeat]; ok {
		if !v.See {
			val |= int32(pb.ACT_BLIND)
		} else {
			val |= int32(pb.ACT_CHAAL)
		}
	}

	//sideshow标记判断
	t.DeskAct.ActState = t.LimitShow(val)
	//比牌标记判断
	//第一轮不能比
	// if t.DeskHua.ActTimes != 0 {
	// 	//TODO 比牌金币限制
	// 	val |= int32(pb.ACT_BI)
	// }

	//看牌标记判断
	// if v, ok := t.DeskHua.ActSeats[t.DeskHua.ActSeat]; ok {
	// 	if !v.See {
	// 		val |= int32(pb.ACT_SEE)
	// 	}
	// }

	//TODO 加注和跟注上限
	// switch t.DeskData.Rtype {
	// case int32(pb.ROOM_TYPE0): //自由
	// 	user := t.getUserBySeat(t.DeskHua.ActSeat)
	// 	if user != nil {
	// 		if user.GetCoin() >= t.DeskHua.ActRaiseNum {
	// 			//为下注的指定倍数时才可以加注
	// 			val |= int32(pb.ACT_RAISE)
	// 		}
	// 	}
	// case int32(pb.ROOM_TYPE1): //私人
	// 	val |= int32(pb.ACT_RAISE)
	// }
	//上家下注大于当前位置下注
	// if t.DeskHua.ActCallNum > 0 {
	// 	val |= int32(pb.ACT_CALL)
	// }

}

func (t *Desk) raiseAnte() {
	msg := new(pb.AK47CoinRaiseAnteNtf)
	for k, v := range t.seats {
		if !v.Ready {
			continue
		}
		// d := new(data.ActStatus)
		// d.Alive = true
		// t.DeskAct.ActSeats[k] = d
		//开始时下暗注
		info := t.setBet(k, v.Userid, int64(t.DeskData.Ante), fmt.Sprintf("ak47%s房间下底注", t.DeskData.Rid))
		msg.Info = append(msg.Info, info)
		t.recordBet(int64(t.DeskData.Ante))
	}
	msg.Pot = t.DeskGame.BetNum
	t.broadcast(msg)
}

// 开始游戏初始化操作
func (t *Desk) initAct() {

	//初始化玩家操作
	//TODO 加注和跟注下限
	// t.DeskHua.ActCallNum = int64(t.DeskData.Ante)      //跟住
	// t.DeskHua.ActRaiseNum = int64(t.DeskData.Ante) * 2 //加注

	t.timer = 0
	t.DeskAct.ActAnte = int64(t.Ante)
	t.DeskAct.ActSeat = t.getNextActSeat()
	//设置下家操作值
	t.setNextActState()
	//设置局内充值条件
	t.SetChargeCondition()
	//广播下家操作状态消息
	t.pushActState()
}

// 下注成功设置
func (t *Desk) setBet(seat uint32, userid string, num int64, desc string) *pb.AK47CoinRaiseInfo {
	if v, ok := t.seats[seat]; ok {
		v.Bet += num
	}
	if v, ok := t.DeskAct.ActSeats[seat]; ok {
		v.ActNum += num
		v.Bet = num
	}
	t.DeskGame.BetNum += num
	return t.setBetMsg(seat, userid, num, desc)
}

// 开始时下暗注
func (t *Desk) setBetMsg(seat uint32, userid string, num int64, desc string) *pb.AK47CoinRaiseInfo {
	msg := &pb.AK47CoinRaiseInfo{
		Seat:   seat,
		Userid: userid,
		Value:  num, //下底注
		// Total:  num, //总下注
		// Pot:    t.DeskGame.BetNum,
	}

	if v, ok := t.DeskAct.ActSeats[seat]; ok {
		msg.Total = v.ActNum
	}

	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0): //自由
		t.sendCurrency(userid, (-1 * num), int32(pb.LOG_TYPE114), desc)
	case int32(pb.ROOM_TYPE1): //私人
		t.sendCoin(userid, (-1 * num), int32(pb.LOG_TYPE5))
	}
	// t.broadcast(msg)
	return msg
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

func (t *Desk) broadFakeSeat() {
	ntf := &pb.UpdateFakeSeatNtf{
		Gtype:  t.Gtype,
		Roomid: t.Rid,
	}
	for k := range t.FakeSeats {
		ntf.Seat = append(ntf.Seat, k)
	}
	t.broadcast5(ntf)
}

// 冤家牌
func (t *Desk) hedgeCard() {
	r, _ := t.roleCountNumNoWatch()
	if r < 2 {
		return
	}
	var max, min float64 = 0, 2
	var maxSeat, minSeat uint32
	hedge := false
	for k, v := range t.seats {
		if !v.Ready {
			continue
		}
		if r, ok := t.roles[v.Userid]; ok {
			if !r.Robot {
				f := handler.GetFactor(r.User)
				if f > max {
					max = f
					maxSeat = k
				}
				if f < min {
					min = f
					minSeat = k
				}
				// 是否触发过冤家牌
				sign := 1
				for i := 0; i < 5; i++ {
					if r.Hedge&sign != 0 {
						// 触发过冤家牌
						hedge = true
					}
					sign = sign << 1
				}
			}
		}
	}
	changeCard := false
	if max-min > 0.2 {
		glog.Infof("check hedge card, waterid:%s", t.GameId)
		// 差距超过0.2，判断要不要发冤家牌
		if hedge {
			if utils.RandWan(3000) {
				// 触发过冤家牌,只有30%概率再触发
				changeCard = true
			}
		} else {
			// 没触发过冤家牌,必定触发
			changeCard = true
		}
	}
	if changeCard {
		// 换牌
		glog.Infof("hedge card rid:%s, waterid:%s", t.Rid, t.GameId)
		ct := t.Game.AK47.FindCardType(t.cardTypeId)

		wildNum := handler.GetWildCardNum(ct, false, 3)
		wildcards := t.getWildCards(wildNum)
		baozi, changeBaozi, cards := algo.AK47GetCard(wildcards, algo.BaoZi, t.DeskGame.Cards)
		copy(t.DeskGame.Cards, cards)
		// t.DeskGame.Cards = cards

		wildcards = t.getWildCards(handler.GetWildCardNum(ct, false, 1))
		ths, changeThs, cards := algo.AK47GetCard(wildcards, algo.TongHuaShun, t.DeskGame.Cards)
		copy(t.DeskGame.Cards, cards)

		s := t.getSeat(maxSeat)
		copy(s.Cards, ths)
		copy(s.ChangeCards, changeThs)
		// s.Cards = ths
		s = t.getSeat(minSeat)
		copy(s.Cards, baozi)
		copy(s.ChangeCards, changeBaozi)
		// s.Cards = baozi
		// 冤家牌记录
		t.addHedgeRecord(true)
	}
}

func (t *Desk) addHedgeRecord(trigger bool) {
	for _, s := range t.seats {
		if s.Watch {
			continue
		}
		if role, ok := t.roles[s.Userid]; ok {
			if trigger {
				role.Hedge = role.Hedge<<1 | 1
			} else {
				role.Hedge = role.Hedge << 1
			}
			msg := &pb.HedgeTrigger{Userid: role.Userid}
			role.Pid.Tell(msg)
		}
	}
}

func (t *Desk) getWildCards(num int) []uint32 {
	//发万能牌
	wildCards := []uint32{}
	for i := 0; i < int(num); i++ {
		wildCard := t.DeskGame.LaiCards[0]
		t.DeskGame.LaiCards = t.DeskGame.LaiCards[1:]
		wildCards = append(wildCards, wildCard)
	}
	return wildCards
}

// vim: set foldmethod=marker foldmarker=//',//.:
