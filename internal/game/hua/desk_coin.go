package hua

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/gen/tb"
	"goserver/pkg/data"
	"goserver/pkg/game/algo"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"goserver/pkg/zlog"
	"math"
	"math/rand"
	"sort"
	"strconv"
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

		// 记录私人房真金输赢
		var privDiamond int64
		if t.isPrivCashRoom() &&
			(ltype == int32(pb.LOG_TYPE110) || ltype == int32(pb.LOG_TYPE111)) {
			privDiamond = score
			v.User.AddPrivDiamond(privDiamond)
		}

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
				Priv:    privDiamond,
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
			Priv:    privDiamond,
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
	// 记录私人房真金输赢
	var privDiamond int64
	if t.isPrivCashRoom() &&
		(ltype == int32(pb.LOG_TYPE110) || ltype == int32(pb.LOG_TYPE111)) {
		privDiamond = num
	}

	//玩家在线
	if v, ok := t.roles[userid]; ok && v != nil {
		v.User.AddCoin(num)
		v.User.AddPrivDiamond(privDiamond)
		//在线
		if !v.Offline {
			//货币变更及时同步
			msg := &pb.ChangeCurrency{
				Userid: userid,
				Coin:   num,
				Type:   ltype,
				Priv:   privDiamond,
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
		Priv:   privDiamond,
	}
	//通过大厅通知其它节点
	t.rolePid.Tell(msg)
}

func (t *Desk) sendDiamond(userid string, num int64, ltype int32) {
	if num == 0 {
		return
	}
	// 记录私人房真金输赢
	var privDiamond int64
	if t.isPrivCashRoom() &&
		(ltype == int32(pb.LOG_TYPE110) || ltype == int32(pb.LOG_TYPE111)) {
		privDiamond = num
	}

	//玩家在线
	if v, ok := t.roles[userid]; ok && v != nil {
		v.User.AddDiamond(num)
		v.User.AddPrivDiamond(privDiamond)

		//在线
		if !v.Offline {
			//货币变更及时同步
			msg := &pb.ChangeCurrency{
				Userid:  userid,
				Diamond: num,
				Type:    ltype,
				Priv:    privDiamond,
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
		Priv:    privDiamond,
	}
	//通过大厅通知其它节点
	t.rolePid.Tell(msg)
}

// 更新娱乐分
func (t *Desk) sendFraction(userid string, num int64, ltype int32) {
	if v, ok := t.roles[userid]; ok && v != nil {
		v.User.AddFraction(num)
		//在线
		if !v.Offline {
			//娱乐分变更通知
			msg := &pb.ChangeFractionNtf{
				Userid:     userid,
				Type:       ltype,
				Fraction:   v.User.GetFraction(),
				ChFraction: num,
			}
			v.Pid.Tell(msg)
		}
	}
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
func (t *Desk) coinEnterMsg(userid string) *pb.JHCoinEnterRoomRsp {
	msg := new(pb.JHCoinEnterRoomRsp)
	//房间数据
	msg.Roominfo = handler.PackJHCoinRoom(t.DeskData)
	msg.Roominfo.State = t.state
	// 测试透视
	msg.Roominfo.UCards = t.playersHand()
	msg.Gameid = t.Game.Id
	//坐下玩家信息
	msg.Userinfo = t.coinSeatBetsMsg(userid)
	//位置下注信息
	msg.Betsinfo = t.coinBetsMsg()
	msg.BetInfo = t.betInfo

	// 房间暂停状态原因
	if t.state == int32(pb.STATE_PAUSE) {
		msg.Roominfo.StatePauseReaon = int32(t.reason)
	}

	if t.DeskAct != nil {
		msg.Actseat = t.DeskAct.ActSeat
		msg.Actstate = t.DeskAct.ActState
		msg.Ante = uint32(t.DeskAct.ActAnte)
		msg.Pot = t.DeskGame.BetNum
		msg.Dealer = t.DeskGame.DealerSeat
		msg.Timer = int32(BetTime - t.timer)
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
	msg := new(pb.JHCameinNtf)
	msg.Userinfo = t.coinRoleMsg(userid)
	t.broadcast(msg)
}

// 召唤机器人
func (t *Desk) loadRobot() {
	t.robotTime++
	t.callRobot()
}

// 随机人机数
func (t *Desk) randRobotNum() int {
	return utils.RandMN(t.Game.TP.Single_Robot[0], t.Game.TP.Single_Robot[1])
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
	// 游戏进行中不召唤
	if t.IsGaming() {
		return
	}
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
	if t.robotNum == 0 {
		return
	}
	// if t.robotNum <= 0 {
	// 	t.robotNum = t.randRobotNum()
	// 	t.robotTime = 0
	// }

	var callRobot int
	if n == t.robotNum {
		return
	} else if n < t.robotNum {
		// 召唤机器人
		callRobot = t.robotNum - n
	}
	// else {
	// 	// 踢出多余的机器人
	// 	kicks := n - t.robotNum
	// 	var robots []string
	// 	for k, v := range t.roles {
	// 		if v.Robot {
	// 			robots = append(robots, k)
	// 		}
	// 	}
	// 	for i := 0; i < kicks && len(robots) > 0; i++ {
	// 		ridx := utils.RandIntN(len(robots))
	// 		robotid := robots[ridx]
	// 		robots = append(robots[:ridx], robots[ridx+1:]...)
	// 		// t.notifyGateUserLeft(robotid, pb.OK, 0)
	// 		// t.userLeaveDesk(robotid)
	// 		// msg 通知人机主动退出, 避免立即踢出但结算动画未播放完成
	// 		if t.robotLeaving == nil {
	// 			t.robotLeaving = make(map[string]bool)
	// 		}
	// 		t.robotLeaving[robotid] = true
	// 		ntf := &pb.RobotLeaveNtf{RobotId: robotid}
	// 		t.send2userid(robotid, ntf)
	// 	}
	// 	return
	// }

	callRobot = int(math.Min(float64(callRobot), float64(int(t.Game.Count)-len(t.roles))))
	if callRobot > 0 {
		msg := new(pb.RobotMsg)
		msg.Gameid = t.Game.Id
		msg.Roomid = t.DeskData.Rid
		msg.Code = t.DeskData.Code
		msg.Rtype = t.DeskData.Rtype
		msg.Ltype = t.DeskData.Ltype
		msg.Gtype = t.DeskData.Gtype
		msg.EnvBet = int32(t.DeskData.Multiple)
		// msg.Min = int32(t.Game.Min_Access)
		// msg.Max = int32(t.Game.Max_Access)
		maxAccess := float64(t.Game.Max_Access)
		if maxAccess <= 0 {
			// 无上限房间，筹码池 1/10
			maxAccess = float64(t.Game.TP.Pool_Limit) / 10
		}
		msg.Min = int32(float64(maxAccess) * t.getRobotExpressRecord().RobotCarryRate[0])
		msg.Max = int32(float64(maxAccess) * t.getRobotExpressRecord().RobotCarryRate[1])
		msg.Num = uint32(callRobot)
		msg.LimitTime = utils.BsonNow().UnixMilli() + ReadyTime*1000
		msg.Emoji = t.getEmoji() //随机人机表情
		t.dbmsPid.Tell(msg)

		t.robotCalling += callRobot

		zlog.Infof("%s call robot num: %d, call: %d, carry: %d~%d", t.GameId, t.robotNum, callRobot, msg.Min, msg.Max)
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
func (t *Desk) coinRoleMsg(userid string) (msg *pb.JHRoomUser) {
	if v, ok := t.roles[userid]; ok {
		if v.Seat == 0 {
			return //没有坐下不广播
		}
		msg = handler.PackJHCoinUser(v.User)
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
func (t *Desk) coinSeatBetsMsg(userid string) (msg []*pb.JHRoomUser) {
	for k, v := range t.seats {
		msg2 := t.coinRoleMsg(v.Userid)
		if msg2 == nil {
			continue
		}
		//自己手牌
		if v.Userid == userid && t.isSee(k) {
			msg2.Cards = t.getHandCards(k)
			msg2.Typ = algo.HuaType(msg2.Cards)
		}
		msg = append(msg, msg2)
	}
	return
}

// 玩家下注数据
func (t *Desk) coinBetsMsg() (msg []*pb.JHRoomBets) {
	for k, v := range t.seats {
		msg2 := &pb.JHRoomBets{
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

// 参与牌局的人是否全部播放完发牌动画
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
	switch t.Rtype {
	case int32(pb.ROOM_TYPE1): // 私人房
		t.state = t.chargePrevState
		switch t.state {
		case int32(pb.STATE_BET): // 下注中
			seatid := t.ActSeat
			userid := t.getUserid(seatid)
			t.coinFold(userid)
		default:
			glog.Errorf("unknown charge before state: %v", t.state)
			fallthrough
		case int32(pb.STATE_READY): // 回合结束时的最低金额检测充值
			t.privRoundChargingResume(true, "")
		}
	default:
		t.state = int32(pb.STATE_BET)

		seatid := t.ActSeat
		userid := t.getUserid(seatid)
		t.coinFold(userid)
	}
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
		return true, t.Game.TP.NewbieMode.FoamCardType
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
	for k, v := range t.Game.TP.NewbieMode.SpecialRound {
		if int(round) <= v {
			idx = k
			ok = true
			break
		}
	}

	if ok {
		return ok, t.Game.TP.NewbieMode.SpecialRoundCardType[idx]
	} else {
		return false, 0
	}
}

// 获取新手模式牌型
func (t *Desk) getNewbieModeCardType() int {
	for _, v := range t.seats {
		// if !v.Ready {
		// continue
		// }
		if r, ok := t.roles[v.Userid]; ok {
			if !r.Robot {
				ok, cardType := t.CheckSpecialRound(r) //特定局
				if ok {
					return cardType
				}

				if r.State == 3 { //平民配置
					diamond := r.OutDiamond
					index := 0
					for i, v := range t.Game.TP.NewbieMode.CivilianCanWithdrawRange {
						if diamond < int64(v) {
							index = i
							break
						} else {
							index = i + 1
						}
					}
					if index > len(t.Game.TP.NewbieMode.CivilianCanWithdrawRange)-1 {
						index = len(t.Game.TP.NewbieMode.CivilianCanWithdrawRange) - 1
					}
					winRate := t.Game.TP.NewbieMode.CivilianWinRate[index]
					win := utils.RandWan(int32(winRate))
					if win {
						weight := t.Game.TP.NewbieMode.CivilianWinWeight
						var choices []utils.Choice
						for i, v := range weight {
							choices = append(choices, utils.Choice{Weight: v, Item: i})
						}
						c, _ := utils.WeightedChoice(choices)
						i := c.Item.(int)
						return t.Game.TP.NewbieMode.CivilianWinCardType[i]
					} else {
						weight := t.Game.TP.NewbieMode.CivilianLoseWeight
						var choices []utils.Choice
						for i, v := range weight {
							choices = append(choices, utils.Choice{Weight: v, Item: i})
						}
						c, _ := utils.WeightedChoice(choices)
						i := c.Item.(int)
						return t.Game.TP.NewbieMode.CivilianLoseCardType[i]
					}
				} else { //新手配置
					diamond := r.OutDiamond
					index := 0
					for i, v := range t.Game.TP.NewbieMode.CanWithdrawRange {
						if diamond < int64(v) {
							index = i
							break
						} else {
							index = i + 1
						}
					}
					if index > len(t.Game.TP.NewbieMode.CanWithdrawRange)-1 {
						index = len(t.Game.TP.NewbieMode.CanWithdrawRange) - 1
					}
					winRate := t.Game.TP.NewbieMode.WinRate[index]
					win := utils.RandWan(int32(winRate))
					if win {
						weight := t.Game.TP.NewbieMode.WinWeight
						var choices []utils.Choice
						for i, v := range weight {
							choices = append(choices, utils.Choice{Weight: v, Item: i})
						}
						c, _ := utils.WeightedChoice(choices)
						i := c.Item.(int)
						return t.Game.TP.NewbieMode.WinCardType[i]
					} else {
						weight := t.Game.TP.NewbieMode.LoseWeight
						var choices []utils.Choice
						for i, v := range weight {
							choices = append(choices, utils.Choice{Weight: v, Item: i})
						}
						c, _ := utils.WeightedChoice(choices)
						i := c.Item.(int)
						return t.Game.TP.NewbieMode.LoseCardType[i]
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

	for _, v := range t.Game.TP.ControlType {
		if v.Switch != 1 {
			continue
		}
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

	for _, v := range t.Game.TP.ControlType {
		if v.Switch != 1 {
			continue
		}
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

	if t.Game.TP.RoomFactorSwitch == 0 {
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
	if t.Rtype == int32(pb.ROOM_TYPE1) {
		t.privGameStart()
		return
	}

	//初始化
	t.gameStartInit()
	//洗牌
	t.shuffle()
	//发牌
	t.dealRandom()
	//详情初始化
	t.detailInit()
	//玩家胜率检测换牌
	t.winRateChangePlayerCards()

	// 按优先级顺序执行策略, 检查互斥
	strategys := []string{Strategy1001, Strategy1002, Strategy1003}
	strategyChecks := []func(){
		// 1001怦然心动 玩家是最大高牌换其他牌型
		t.checkStrategy1001,
		// 1002乐极生悲
		t.checkStrategy1002,
		// 1003高潮涌现
		t.checkStrategy1003,
	}
	sort.Slice(strategyChecks, func(i, j int) bool {
		a := table.GetTables().TpStrategyTable.Get(strategys[i])
		b := table.GetTables().TpStrategyTable.Get(strategys[j])
		return a.Priority > b.Priority
	})
	// sort.Slice(strategys, func(i, j int) bool {
	// 	a := table.GetTables().TpStrategyTable.Get(strategys[i])
	// 	b := table.GetTables().TpStrategyTable.Get(strategys[j])
	// 	return a.Priority > b.Priority
	// })
	// zlog.Infof("%s round strategys %v", t.detail.WaterId, strategys)
	for _, checker := range strategyChecks {
		checker()
	}

	// // 1001怦然心动 玩家是最大高牌换其他牌型
	// t.checkStrategy1001()
	// // 1002乐极生悲
	// t.checkStrategy1002()
	// // 1003高潮涌现
	// t.checkStrategy1003()
	//人机策略初始化
	t.robotStartInit()

	//胜率检测
	// t.winRateCheck()
	//开局换牌处理
	// t.StartChangeHandler()
	//策略局处理
	// t.StrategyHandler()
	//冤家牌
	// t.hedgeCard()
	//扣底注
	t.raiseAnte()
	//选庄
	t.dealerHandler()
	//机器人处理
	// t.robotHandler()
}

func (t *Desk) isStoryPro() {
	playerid := t.GetOnlyOnePlayer()
	if playerid == "" {
		return
	}
	player := t.getPlayer(playerid)
	if player == nil {
		return
	}
	if !utils.SliceIn(data.GAME_RECHARGE, player.RechargeTarge...) {
		// 有局内充值
		return
	}

}

func (t *Desk) findStoryControl(role *data.DeskRole) *tb.TpStoryModelRecord {

	checkScore := func(score int64, record *tb.TpStoryModelRecord) bool {
		var weight int32
		l := len(record.ScoreRange)
		l1 := len(record.TriggerProb)

		if score < int64(record.ScoreRange[0]) {
			weight = record.TriggerProb[0]
		} else if score >= int64(record.ScoreRange[l-1]) {
			weight = record.TriggerProb[l1-1]
		} else {
			for i := 0; i < l-1; i++ {
				if score >= int64(record.ScoreRange[i]) && score < int64(record.ScoreRange[i+1]) {
					weight = record.TriggerProb[i+1]
				}
			}
		}
		if utils.RandWan(weight) {
			return true
		}

		return false
	}

	conf := table.GetTables().TpStoryModelTable.GetDataList()
	roomId, _ := strconv.ParseInt(t.Game.Id, 10, 32)
	for _, v := range conf {
		if v.Room == int32(roomId) && !t.IsStoryCD(role.Seat, v.Id) && checkScore(role.GetScore(), v) {
			return v
		}
	}
	return nil
}

func (t *Desk) findStoryCardTypeId(role *data.DeskRole, record *tb.TpStoryModelRecord) int32 {

	index2CardType := func(index int) uint32 {
		switch index {
		case 0:
			return algo.BaoZi
		case 1:
			return algo.TongHuaShun
		case 2:
			return algo.ShunZi
		case 3:
			return algo.TongHua
		case 4:
			return algo.DuiZi
		case 5:
			return algo.GaoPai
		default:
			return algo.Null
		}
	}

	checkFollowRate := func(index int) bool {
		cardType := index2CardType(index)
		if _, ok := role.TpUserFollowRateTrigger[int32(cardType)]; !ok {
			return true
		}

		var success int32
		if _, ok := role.TpUserFollowRateSuccess[int32(cardType)]; !ok {
			success = 0
		} else {
			success = role.TpUserFollowRateSuccess[int32(cardType)]
		}

		rate := float64(success) / float64(role.TpUserFollowRateTrigger[int32(cardType)])
		rate *= 10000
		rate32 := int32(rate)
		if rate32 < record.FollowRate {
			return false
		}
		return true
	}

	var choices []utils.Choice
	for k, v := range record.CardTypeWeight {
		if checkFollowRate(k) {
			choices = append(choices, utils.Choice{Weight: int(v), Item: k})
		}
	}
	c, _ := utils.WeightedChoice(choices)
	idx := c.Item.(int)

	switch idx {
	case 0:
		return record.Baozi
	case 1:
		return record.Shunjin
	case 2:
		return record.Shunzi
	case 3:
		return record.Tonghua
	case 4:
		return record.Duizi
	case 5:
		return record.Gaopai
	default:
		return record.Baozi
	}
}

func (t *Desk) findControlStrategy(role *data.DeskRole) *tb.TpControlStrategyRecord {

	checkRoomId := func(roomId int32, record *tb.TpControlStrategyRecord) bool {
		m := make(map[int32]struct{})
		for _, v := range record.Room {
			m[v] = struct{}{}
		}
		_, ok := m[roomId]
		return ok
	}

	checkStockRange := func(stock int32, record *tb.TpControlStrategyRecord) bool {
		if record.StockRange[0] == -1 {
			return true
		}
		return stock >= record.StockRange[0] && stock <= record.StockRange[1]
	}

	checkPrefixCondition := func(role *data.DeskRole, record *tb.TpControlStrategyRecord) bool {
		if record.PrefixCondition[0] == -1 {
			return true
		}
		m := make(map[int32]struct{})
		for _, v := range role.TpUserControlStrategyHistory {
			m[v] = struct{}{}
		}

		for _, v := range record.PrefixCondition {
			if _, ok := m[v]; !ok {
				return false
			}
		}
		return true
	}

	checkScoreRange := func(role *data.DeskRole, record *tb.TpControlStrategyRecord) bool {
		if record.ScoreRange[0] == -1 {
			return true
		}

		score := int32(role.GetScore())
		return score >= record.ScoreRange[0] && score <= record.ScoreRange[1]
	}

	checkWithdrawRange := func(role *data.DeskRole, record *tb.TpControlStrategyRecord) bool {
		if record.WithdrawRange[0] == -1 {
			return true
		}

		return int32(role.OutDiamond) >= record.WithdrawRange[0] && int32(role.OutDiamond) <= record.WithdrawRange[1]
	}

	//检查今日充值次数区间
	checkChargeNum := func(num int32, v *tb.TpControlStrategyRecord) bool {
		if v.TodayChargeTimeRange[0] == -1 {
			return true
		}
		if num >= v.TodayChargeTimeRange[0] && num <= v.TodayChargeTimeRange[1] {
			return true
		}
		return false
	}

	//检查今日充值金额区间
	checkChargeAmount := func(amount int64, v *tb.TpControlStrategyRecord) bool {
		if v.TodayChargeAmountRange[0] == -1 {
			return true
		}
		if int32(amount) >= v.TodayChargeAmountRange[0] && int32(amount) <= v.TodayChargeAmountRange[1] {
			return true
		}
		return false
	}

	//检查今日提现次数区间
	checkWithdrawNum := func(num int32, v *tb.TpControlStrategyRecord) bool {
		if v.TodayWithdrawTimeRange[0] == -1 {
			return true
		}
		if num >= v.TodayWithdrawTimeRange[0] && num <= v.TodayWithdrawTimeRange[1] {
			return true
		}
		return false
	}

	//检查今日提现金额区间
	checkWithdrawAmount := func(amount int64, v *tb.TpControlStrategyRecord) bool {
		if v.TodayWithdrawAmountRange[0] == -1 {
			return true
		}
		if int32(amount) >= v.TodayWithdrawAmountRange[0] && int32(amount) <= v.TodayWithdrawAmountRange[1] {
			return true
		}
		return false
	}

	//检查盈利区间
	checkWinOrLose := func(amount int64, v *tb.TpControlStrategyRecord) bool {
		if v.WinOrLoseRange[0] == -1 {
			return true
		}
		if int32(amount) >= v.WinOrLoseRange[0] && int32(amount) <= v.WinOrLoseRange[1] {
			return true
		}
		return false
	}

	//检查玩家系数区间
	checkPlayerFactor := func(factor float64, v *tb.TpControlStrategyRecord) bool {
		if v.PlayerFactorRange[0] == -1 {
			return true
		}
		if factor >= v.PlayerFactorRange[0] && factor <= v.PlayerFactorRange[1] {
			return true
		}
		return false
	}

	checkFromRegisterRange := func(role *data.DeskRole, record *tb.TpControlStrategyRecord) bool {
		if record.FromRegisterRange[0] == -1 {
			return true
		}
		now := time.Now()
		d := now.Sub(role.Ctime)
		m := d.Minutes()
		m32 := int32(m)
		return m32 >= record.FromRegisterRange[0] && m32 <= record.FromRegisterRange[1]
	}

	checkFromFirstChargeRange := func(role *data.DeskRole, record *tb.TpControlStrategyRecord) bool {
		if record.FromFirstChargeRange[0] == -1 {
			return true
		}
		now := time.Now().UnixMilli()
		d := float64(now-role.FirstChargeTime) / 1000
		m := d / 60
		m32 := int32(m)
		return m32 >= record.FromFirstChargeRange[0] && m32 <= record.FromFirstChargeRange[1]
	}

	checkLastChargeRange := func(role *data.DeskRole, record *tb.TpControlStrategyRecord) bool {
		if record.LastChargeRange[0] == -1 {
			return true
		}
		now := time.Now().UnixMilli()
		d := float64(now-role.LastChargeTime) / 1000
		m := d / 60
		m32 := int32(m)
		return m32 >= record.LastChargeRange[0] && m32 <= record.LastChargeRange[1]
	}

	checkLastChargeGameTimeRange := func(role *data.DeskRole, record *tb.TpControlStrategyRecord) bool {
		if record.LastChargeGameTimeRange[0] == -1 {
			return true
		}
		now := time.Now().UnixMilli()
		d := float64(now-role.TpUserGameTime) / 1000
		m := d / 60
		m32 := int32(m)
		return m32 >= record.LastChargeGameTimeRange[0] && m32 <= record.LastChargeGameTimeRange[1]
	}

	checkTodayGameRoundRange := func(role *data.DeskRole, record *tb.TpControlStrategyRecord) bool {
		if record.TodayGameRoundRange[0] == -1 {
			return true
		}
		return role.TpUserTodayGameRound >= record.TodayGameRoundRange[0] && role.TpUserTodayGameRound <= record.TodayGameRoundRange[1]
	}

	checkTodayTriggerNum := func(role *data.DeskRole, record *tb.TpControlStrategyRecord) bool {
		if record.DailyTrigger == -1 {
			return true
		}
		if role.TpUserTodayControlStrategyNum == nil {
			return true
		}

		if num, ok := role.TpUserTodayControlStrategyNum[record.Id]; ok {
			return num < record.DailyTrigger
		}
		return true
	}

	conf := table.GetTables().TpControlStrategyTable.GetDataList()
	roomId, _ := strconv.ParseInt(t.Game.Id, 10, 32)
	stock := t.getRoomStock()
	factor := handler.GetFactor(role.User)
	for _, v := range conf {
		if checkRoomId(int32(roomId), v) &&
			checkStockRange(stock, v) &&
			checkPrefixCondition(role, v) &&
			checkScoreRange(role, v) &&
			checkWithdrawRange(role, v) &&
			checkChargeNum(role.TpUserTodayChargeNum, v) &&
			checkChargeAmount(role.TpUserTodayChargeAmount, v) &&
			checkWithdrawNum(role.TpUserTodayWithdrawNum, v) &&
			checkWithdrawAmount(role.TpUserTodayWithdrawAmount, v) &&
			checkWinOrLose(role.TpUserTotalWinOrLoseAmount, v) &&
			checkPlayerFactor(factor, v) &&
			checkFromRegisterRange(role, v) &&
			checkFromFirstChargeRange(role, v) &&
			checkLastChargeRange(role, v) &&
			checkLastChargeGameTimeRange(role, v) &&
			checkTodayGameRoundRange(role, v) &&
			checkTodayTriggerNum(role, v) {
			return v
		}
	}

	return nil
}

// 获取房间库存
func (t *Desk) getRoomStock() int32 {
	msg := &pb.GetRoomFactor{}
	msg.GameId = t.Game.Id

	res := t.reqRoom(msg)
	var response *pb.GetedRoomFactor
	var ok bool

	if response, ok = res.(*pb.GetedRoomFactor); !ok {
		glog.Errorf("get room factor failed: %#v", res)
		return 0
	}

	return int32(response.CashStock)
}

// newbieProbe 新手试探期
func (t *Desk) newbieProbe() (newbie bool) {
	playerid := t.GetOnlyOnePlayer()
	if playerid == "" {
		return
	}
	player := t.getPlayer(playerid)
	if player == nil {
		glog.Errorf("not found player %s", playerid)
		return
	}
	// 已充值跳出新手局
	if player.GetMoney() > 0 {
		return
	}

	// 玩家能否进试探局, 玩到999进入稳定局
	if player.TpNewbieProbeId > 0 && player.TpNewbieProbeId != 999 {
		probe := table.GetTables().TpNewbieProbeTable.Get(player.TpNewbieProbeId)
		if probe != nil &&
			utils.InSlice(t.Game.Id, probe.Rooms) &&
			player.OutDiamond >= int64(probe.WithdrawRange[0]) &&
			player.OutDiamond <= int64(probe.WithdrawRange[1]) {
			t.newbieProbeRecord = probe
		}
	}
	if t.newbieProbeRecord != nil {
		t.cardTypeId = int(t.newbieProbeRecord.CardId)
		return true
	}

	// 进入新手稳定局, 清除试探局id
	var newbieStableRecord *tb.TpNewbieStableRecord
	if player.TpNewbieStableId > 0 {
		newbieStableRecord = table.GetTables().TpNewbieStableTable.Get(player.TpNewbieStableId)
	}
	if newbieStableRecord != nil {
		player.TpNewbieProbeId = 0
		player.TpNewbieStableId = newbieStableRecord.Id

		// 匹配牌型
		out := int32(player.OutDiamond)
		var cardId int32
		if out < newbieStableRecord.Withdraws[0] {
			cardId = newbieStableRecord.CardIds[0]
		} else if out > newbieStableRecord.Withdraws[len(newbieStableRecord.Withdraws)-1] {
			cardId = newbieStableRecord.CardIds[len(newbieStableRecord.CardIds)-1]
		} else {
			for i := 0; i < len(newbieStableRecord.Withdraws)-1; i++ {
				if out >= newbieStableRecord.Withdraws[i] && out < newbieStableRecord.Withdraws[i+1] {
					cardId = newbieStableRecord.CardIds[i+1]
					break
				}
			}
		}
		record := new(tb.TpNewbieStableRecord)
		err := utils.Clone(record, newbieStableRecord)
		if err != nil {
			glog.Errorf("clone error: %v, %v", err, newbieStableRecord)
			return
		}
		record.CardId = cardId
		t.newbieStableRecord = record
		t.cardTypeId = int(record.CardId)
		return true
	}
	// 匹配稳定局
	for _, v := range table.GetTables().TpNewbieStableTable.GetDataList() {
		out := int32(player.OutDiamond)
		var cardId int32
		if out < v.Withdraws[0] {
			cardId = v.CardIds[0]
		} else if out > v.Withdraws[len(v.Withdraws)-1] {
			cardId = v.CardIds[len(v.CardIds)-1]
		} else {
			for i := 0; i < len(v.Withdraws)-1; i++ {
				if out >= v.Withdraws[i] && out < v.Withdraws[i+1] {
					cardId = v.CardIds[i+1]
					break
				}
			}
		}
		if cardId != 0 {
			record := new(tb.TpNewbieStableRecord)
			err := utils.Clone(record, v)
			if err != nil {
				glog.Errorf("clone error: %v, %v", err, v)
				return
			}
			player.TpNewbieProbeId = 0
			player.TpNewbieStableId = record.Id
			record.CardId = cardId
			t.newbieStableRecord = record
			t.cardTypeId = int(record.CardId)
			return true
		}
	}
	return
}

// newbieProbe 游戏结束新手试探期逻辑
func (t *Desk) newbieProbeGameOver() {
	playerid := t.GetOnlyOnePlayer()
	if playerid == "" {
		return
	}
	player := t.getPlayer(playerid)
	if player == nil {
		return
	}
	// 生效策略局数+1

	var newbieProbeId int32
	if t.newbieProbeRecord != nil {
		newbieProbeId = t.newbieProbeRecord.Cd
	} else if t.newbieStableRecord != nil {
		newbieProbeId = t.newbieStableRecord.Id
	}
	if newbieProbeId == 0 {
		return
	}
	if player.TpNewbieStableRounds == nil {
		player.TpNewbieStableRounds = make(map[int32]int32)
	}
	player.TpNewbieStableRounds[newbieProbeId]++

	// 记录生效策略
	seatid := t.getSeatid(playerid)
	tpDetail := t.detail.FindTPDetail(seatid)
	tpDetail.NewBieProbeId = newbieProbeId
	// 记录策略生效局数
	tpDetail.NewBieProbeRound = player.TpNewbieStableRounds[tpDetail.NewBieProbeId]

	t._newbieProbeGameOver(playerid, player)
	// 同步数据到网关
	msg := &pb.JHSetTpNewbieProbe{
		TpNewbieProbeId:      player.TpNewbieProbeId,
		TpNewbieStableId:     player.TpNewbieStableId,
		TpNewbieStableRounds: player.TpNewbieStableRounds,
	}
	t.send2userid(playerid, msg)
}

// newbieProbe 游戏结束新手试探期逻辑
func (t *Desk) _newbieProbeGameOver(playerid string, player *data.User) {
	// 新手试探局
	if t.newbieProbeRecord != nil {
		// 结束后进入稳定局
		if t.newbieProbeRecord.OverStableId > 0 {
			player.TpNewbieProbeId = 0
			player.TpNewbieStableId = t.newbieProbeRecord.OverStableId
			return
		}
		// 使用策略
		seatid := t.getSeatid(playerid)
		status := t.getStatus(seatid)
		if status == nil {
			glog.Error("player not exists: ", playerid)
			return
		}
		var acted bool
		for i := len(status.ActHistory) - 1; i >= 0; i-- {
			act := status.ActHistory[i]
			blind := act&int32(pb.ACT_BLIND) == int32(pb.ACT_BLIND)
			if act&int32(pb.ACT_PACK) == int32(pb.ACT_PACK) { // 弃牌
				index := len(t.newbieProbeRecord.DropIds) - 1
				if i < index {
					index = i
				}
				player.TpNewbieProbeId = t.newbieProbeRecord.DropIds[index]
				acted = true
				break
			}
			// 闷牌(加跟比)操作
			if blind {
				if act&int32(pb.ACT_RAISE) == int32(pb.ACT_RAISE) ||
					act&int32(pb.ACT_CHAAL) == int32(pb.ACT_CHAAL) ||
					act&int32(pb.ACT_SHOW) == int32(pb.ACT_SHOW) ||
					act&int32(pb.ACT_SIDESHOW) == int32(pb.ACT_SIDESHOW) {
					index := len(t.newbieProbeRecord.BlindIds) - 1
					if i < index {
						index = i
					}
					player.TpNewbieProbeId = t.newbieProbeRecord.BlindIds[index]
					acted = true
					break
				}
			}
			// 加注
			if act&int32(pb.ACT_RAISE) == int32(pb.ACT_RAISE) {
				index := len(t.newbieProbeRecord.RaiseIds) - 1
				if i < index {
					index = i
				}
				player.TpNewbieProbeId = t.newbieProbeRecord.RaiseIds[index]
				acted = true
				break
			}
			// 跟注比牌
			if act&int32(pb.ACT_CHAAL) == int32(pb.ACT_CHAAL) ||
				act&int32(pb.ACT_SHOW) == int32(pb.ACT_SHOW) ||
				act&int32(pb.ACT_SIDESHOW) == int32(pb.ACT_SIDESHOW) {
				index := len(t.newbieProbeRecord.CallIds) - 1
				if i < index {
					index = i
				}
				player.TpNewbieProbeId = t.newbieProbeRecord.CallIds[index]
				acted = true
				break
			}
		}
		// 其他情形
		if !acted {
			player.TpNewbieProbeId = t.newbieProbeRecord.OtherId
		}
	}

	// 新手稳定局
	if t.newbieStableRecord != nil {
		if t.newbieStableRecord.Rounds <= player.TpNewbieStableRounds[t.newbieStableRecord.Id] {
			// 策略局数满了, 进入下一策略
			player.TpNewbieStableId = t.newbieStableRecord.OverId
			player.TpNewbieStableRounds[player.TpNewbieStableId] = 0
		}
	}
}

func (t *Desk) findNewbieProbeRecord(playerOutDiamond int64) (probe *tb.TpNewbieProbeRecord) {
	for _, v := range table.GetTables().TpNewbieProbeTable.GetDataList() {
		_ = v
		if !utils.InSlice(t.Game.Id, v.Rooms) {
			continue
		}
		if playerOutDiamond >= int64(v.WithdrawRange[0]) &&
			playerOutDiamond <= int64(v.WithdrawRange[1]) {
			probe = v
			break
		}
	}
	return
}

// privGameStart
func (t *Desk) privGameStart() {
	//初始化
	t.gameStartInit()
	//洗牌
	t.shuffle()
	//发牌
	t.deal()
	//详情初始化
	t.detailInit()
	//扣底注
	t.raiseAnte()
	//选庄
	t.dealerHandler()
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
		if algo.HuaCompare(cs1, cs2) { //k赢
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
		if algo.HuaCompare(cs1, cs2) { //k赢
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

	id := t.cardTypeId
	// ct := t.Game.TP.FindCardType(id)
	ct := table.GetTables().TpCardTypeConfigTable.Get(int32(id))
	if ct == nil {
		glog.Errorf("找不到TpCardTypeConfigTable配置:cardType: %v", id)
		return
	}

	count := t.GetCount()
	var groupId int32 = 0
	switch count {
	case 2:
		groupId = ct.RobotStrategy2
	case 3:
		groupId = ct.RobotStrategy3
	case 4:
		groupId = ct.RobotStrategy4
	case 5:
		groupId = ct.RobotStrategy5
	}

	// rsg := t.Game.TP.FindRobotStrategyGroup(int(groupId))
	rsg := table.GetTables().TpRobotStrategyGroupTable.Get(int32(groupId))
	if rsg == nil {
		glog.Errorf("找不到TpRobotStrategyGroupTable配置:groupId: %v", groupId)
		return
	}

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

					var rs data.TPGameRobotStrategy

					//人机手牌类型
					typ := algo.HuaTypeUpOrDown(cs1)

					if algo.HuaCompare(cs1, cs2) { //人机手牌比最大玩家大
						switch typ {
						case algo.BaoziUp:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct1Bigger))
						case algo.BaoziDown:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct2Bigger))
						case algo.TonghuashunUp:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct3Bigger))
						case algo.TonghuashunDown:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct4Bigger))
						case algo.ShunziUp:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct5Bigger))
						case algo.ShunziDown:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct6Bigger))
						case algo.TonghuaUp:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct7Bigger))
						case algo.TonghuaDown:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct8Bigger))
						case algo.DuiziUp:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct9Bigger))
						case algo.DuiziDown:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct10Bigger))
						case algo.GaopaiUp:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct11Bigger))
						case algo.GaopaiDown:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct12Bigger))
						}
					} else { //人机手牌比最大玩家小
						switch typ {
						case algo.BaoziUp:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct1Smaller))
						case algo.BaoziDown:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct2Smaller))
						case algo.TonghuashunUp:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct3Smaller))
						case algo.TonghuashunDown:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct4Smaller))
						case algo.ShunziUp:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct5Smaller))
						case algo.ShunziDown:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct6Smaller))
						case algo.TonghuaUp:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct7Smaller))
						case algo.TonghuaDown:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct8Smaller))
						case algo.DuiziUp:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct9Smaller))
						case algo.DuiziDown:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct10Smaller))
						case algo.GaopaiUp:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct11Smaller))
						case algo.GaopaiDown:
							rs = t.Game.TP.FindRobotStrategy(int(rsg.Ct12Smaller))
						}
					}

					msg := &pb.JHCoinRobotStrategyNtf{}
					msg.Id = rs.Id
					for _, v := range rs.ActionWeight {
						msg.ActionWeight = append(msg.ActionWeight, &pb.JHCoinActionWeight{Values: v})
					}
					msg.ActionTime = rs.ActionTime
					msg.SeeWeight = rs.SeeWeight
					msg.AgreeBi = rs.AgreeBi
					for _, v := range rs.SeeTime {
						msg.SeeTime = append(msg.SeeTime, &pb.JHCoinSeeTime{Values: v})
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
		gaopai, remain := algo.GetGaoPaiCard(t.DeskGame.Cards)
		t.DeskGame.Cards = remain
		t.SetSeeAndPackRobot(v, gaopai, bet_change)
	}
}

// 设置核心人机
func (t *Desk) SetCoreRobot(seatid uint32, cards []uint32, bet_change bool) {
	seat := t.getSeat(seatid)
	if seat == nil {
		return
	}

	seat.Core = true
	if len(cards) > 0 {
		seat.Cards = cards
	}

	msg := &pb.JHCoinRobotStrategyNtf{}
	msg.Core = true
	if bet_change {
		msg.BetChange = true
		msg.ChangeCards = cards
	}
	t.send2userid(seat.Userid, msg)
}

// 设置看牌即弃人机
func (t *Desk) SetSeeAndPackRobot(seatid uint32, cards []uint32, bet_change bool) {
	seat := t.getSeat(seatid)
	if seat == nil {
		return
	}

	seat.SeeAndPack = true
	if len(cards) > 0 {
		seat.Cards = cards
	}

	msg := &pb.JHCoinRobotStrategyNtf{}
	msg.SeeAndPack = true
	if bet_change {
		msg.BetChange = true
		msg.ChangeCards = cards
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
		msg := &pb.JHCoinRobotStrategyNtf{}
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
		t.SetCoreRobot(max, []uint32{}, false) //不需要换牌，但是需要标记人机为核心人机
		return
	}

	core := t.GetCoreRobot()
	if core == 0 {
		return
	}

	maxCards := t.getHandCards(max)
	bigger, remain := algo.GetMaxCard(maxCards, t.DeskGame.Cards)
	t.DeskGame.Cards = remain

	if len(bigger) != 0 {
		t.SetCoreRobot(core, bigger, true)
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
		t.SetCoreRobot(max, []uint32{}, false) //不需要换牌，但是需要标记人机为核心人机
		return
	}

	core := t.GetCoreRobot()
	if core == 0 { //没有人机参与游戏，则跳过
		return
	}

	maxCards := t.getHandCards(max)
	bigger, remain := algo.GetMaxCard(maxCards, t.DeskGame.Cards)
	t.DeskGame.Cards = remain

	if len(bigger) != 0 {
		t.SetCoreRobot(core, bigger, false)
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
	// if t.Game.TP.Strategy100Switch == 0 {
	// 	return
	// }

	state := t.GetPlayerState()
	if state != 3 && !ignoreState { //平民状态走100
		return
	}

	//触发概率
	if !utils.RandWan(int32(t.Game.TP.NewbieMode.CivilianTriggerStrategyProb)) {
		return
	}

	//配置
	// roundConfig := 20                 //累计局数
	scoreConfig := []int{5000, 15000} //携带分范围
	// TriggerProb := 5000               //触发概率
	// TiggerTimes := []int{1, 1, 1} //触发次数

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
	// if role.TPTotalRound < int32(t.Game.TP.Strategy100Round) {
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
	// 	return
	// }

	//判断触发次数
	// times := TiggerTimes[days]
	// if int(role.TPTiggerTimes) >= times {
	// 	return
	// }

	// //判断触发概率
	// if !utils.RandWan(int32(TriggerProb)) {
	// 	return
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

	//给玩家发牌 AAA
	seat.Cards, t.DeskGame.Cards = algo.GetCard(algo.BaoZi1, t.DeskGame.Cards)

	//给主机a发牌 KKK-222/同花顺
	seata := t.getSeat(a)
	var choices []utils.Choice
	choices = append(choices, utils.Choice{Weight: 30, Item: algo.BaoZi2})
	choices = append(choices, utils.Choice{Weight: 30, Item: algo.BaoZi3})
	choices = append(choices, utils.Choice{Weight: 40, Item: algo.TongHuaShun})
	ch, _ := utils.WeightedChoice(choices)
	cardtype := ch.Item.(uint32)
	seata.Cards, t.DeskGame.Cards = algo.GetCard(cardtype, t.DeskGame.Cards)

	//给主机b发牌 顺子/同花
	seatb := t.getSeat(b)
	choices = []utils.Choice{}
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.ShunZi})
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.TongHua})
	ch, _ = utils.WeightedChoice(choices)
	cardtype = ch.Item.(uint32)
	seatb.Cards, t.DeskGame.Cards = algo.GetCard(cardtype, t.DeskGame.Cards)

	//给僚机c发牌 对子/高牌
	if c != 0 {
		seatc := t.getSeat(c)
		choices = []utils.Choice{}
		choices = append(choices, utils.Choice{Weight: 50, Item: algo.DuiZi})
		choices = append(choices, utils.Choice{Weight: 50, Item: algo.GaoPai})
		ch, _ = utils.WeightedChoice(choices)
		cardtype = ch.Item.(uint32)
		seatc.Cards, t.DeskGame.Cards = algo.GetCard(cardtype, t.DeskGame.Cards)
	}

	//给僚机d发牌 对子/高牌
	if d != 0 {
		seatd := t.getSeat(d)
		choices = []utils.Choice{}
		choices = append(choices, utils.Choice{Weight: 50, Item: algo.DuiZi})
		choices = append(choices, utils.Choice{Weight: 50, Item: algo.GaoPai})
		ch, _ = utils.WeightedChoice(choices)
		cardtype = ch.Item.(uint32)
		seatd.Cards, t.DeskGame.Cards = algo.GetCard(cardtype, t.DeskGame.Cards)
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

// 200策略(玩家必赢)大牌诱导充值
func (t *Desk) Strategy200Handler() {
	// if t.Game.TP.Strategy200Switch == 0 {
	// 	return
	// }
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
		// 只有A类有100策略
		if !role.Strategy100Flag {
			t.Strategy100Handler(true)
			return
		}

		state := t.GetPlayerState()
		if state != 4 { //泡沫状态走200
			return
		}

		//触发概率
		if !utils.RandWan(int32(t.Game.TP.NewbieMode.FoamTriggerStrategyProb)) {
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
	// if role.TPTotalRound < int32(t.Game.TP.Strategy200Round) {
	// 	return
	// }

	// 判断携带分范围
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

	var choices []utils.Choice
	// 给玩家发牌 AAA-888/同花顺
	// choices = []utils.Choice{}
	// choices = append(choices, utils.Choice{Weight: 25, Item: algo.BaoZi1})
	// choices = append(choices, utils.Choice{Weight: 50, Item: algo.BaoZi2})
	// choices = append(choices, utils.Choice{Weight: 50, Item: algo.BaoZi3})
	// choices = append(choices, utils.Choice{Weight: 50, Item: algo.TongHuaShun})
	// ch, _ := utils.WeightedChoice(choices)
	// cardtype := ch.Item.(uint32)
	seat.Cards, t.DeskGame.Cards = algo.GetCard(algo.BaoZi, t.DeskGame.Cards)

	// 给主机a发牌 AAA-888/同花顺
	seata := t.getSeat(a)
	// choices = []utils.Choice{}
	// choices = append(choices, utils.Choice{Weight: 25, Item: algo.BaoZi1})
	// choices = append(choices, utils.Choice{Weight: 50, Item: algo.BaoZi2})
	// choices = append(choices, utils.Choice{Weight: 25, Item: algo.BaoZi3})
	// choices = append(choices, utils.Choice{Weight: 50, Item: algo.TongHuaShun})
	// ch, _ = utils.WeightedChoice(choices)
	// cardtype = ch.Item.(uint32)
	seata.Cards, t.DeskGame.Cards = algo.GetCard(algo.BaoZi, t.DeskGame.Cards)

	//主机a一定比玩家小
	if algo.HuaCompare(seata.Cards, seat.Cards) {
		seat.Cards, seata.Cards = seata.Cards, seat.Cards
	}

	// 给主机b发牌 顺子/同花
	seatb := t.getSeat(b)
	choices = []utils.Choice{}
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.ShunZi})
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.TongHua})
	ch, _ := utils.WeightedChoice(choices)
	cardtype := ch.Item.(uint32)
	seatb.Cards, t.DeskGame.Cards = algo.GetCard(cardtype, t.DeskGame.Cards)

	// 给僚机c发牌 对子/高牌
	if c != 0 {
		seatc := t.getSeat(c)
		choices = []utils.Choice{}
		choices = append(choices, utils.Choice{Weight: 50, Item: algo.DuiZi})
		choices = append(choices, utils.Choice{Weight: 50, Item: algo.GaoPai})
		ch, _ = utils.WeightedChoice(choices)
		cardtype = ch.Item.(uint32)
		seatc.Cards, t.DeskGame.Cards = algo.GetCard(cardtype, t.DeskGame.Cards)
	}

	// 给僚机d发牌 对子/高牌
	if d != 0 {
		seatd := t.getSeat(d)
		choices = []utils.Choice{}
		choices = append(choices, utils.Choice{Weight: 50, Item: algo.DuiZi})
		choices = append(choices, utils.Choice{Weight: 50, Item: algo.GaoPai})
		ch, _ = utils.WeightedChoice(choices)
		cardtype = ch.Item.(uint32)
		seatd.Cards, t.DeskGame.Cards = algo.GetCard(cardtype, t.DeskGame.Cards)
	}
	// 通知所有主机僚机
	t.NotifyAllMasterAndSlave()

	//设置200标记
	role.SetStrategy200()
	//同步数据
	msg := &pb.TriggerStategy200{}
	t.send2userid(userid, msg)
	// 重置累计局数
	// role.ResetTPTotalRound()
	// 同步数据
	// msg1 := &pb.TPTotalRound{}
	// msg1.Isreset = true
	// t.send2userid(userid, msg1)

	// 设置本局为策略局
	t.isStrategy = true
	t.strategyType = 200
}

// 300策略(玩家非必赢)大牌诱导充值
func (t *Desk) Strategy300Handler() {
	if t.Game.TP.Strategy300Switch == 0 {
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

	// 判断底注
	if role.Diamond >= int64(t.Ante)*300 {
		return
	}

	// 判断累计局数
	if role.TPTotalRound < int32(t.Game.TP.Strategy300Round) {
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

	var choices []utils.Choice
	// 给玩家发牌 AAA-888/同花顺
	// choices = []utils.Choice{}
	// choices = append(choices, utils.Choice{Weight: 25, Item: algo.BaoZi1})
	// choices = append(choices, utils.Choice{Weight: 50, Item: algo.BaoZi2})
	// choices = append(choices, utils.Choice{Weight: 50, Item: algo.BaoZi3})
	// choices = append(choices, utils.Choice{Weight: 50, Item: algo.TongHuaShun})
	// ch, _ := utils.WeightedChoice(choices)
	// cardtype := ch.Item.(uint32)
	seat.Cards, t.DeskGame.Cards = algo.GetCard(algo.BaoZi, t.DeskGame.Cards)

	// 给主机a发牌 AAA-888/同花顺
	seata := t.getSeat(a)
	// choices = []utils.Choice{}
	// choices = append(choices, utils.Choice{Weight: 25, Item: algo.BaoZi1})
	// choices = append(choices, utils.Choice{Weight: 50, Item: algo.BaoZi2})
	// choices = append(choices, utils.Choice{Weight: 25, Item: algo.BaoZi3})
	// choices = append(choices, utils.Choice{Weight: 50, Item: algo.TongHuaShun})
	// ch, _ = utils.WeightedChoice(choices)
	// cardtype = ch.Item.(uint32)
	seata.Cards, t.DeskGame.Cards = algo.GetCard(algo.BaoZi, t.DeskGame.Cards)

	if utils.RandWan(6000) { //主机a比玩家大
		if algo.HuaCompare(seat.Cards, seata.Cards) {
			seat.Cards, seata.Cards = seata.Cards, seat.Cards
		}
	} else { //主机a比玩家小
		if algo.HuaCompare(seata.Cards, seat.Cards) {
			seat.Cards, seata.Cards = seata.Cards, seat.Cards
		}
	}

	if algo.HuaCompare(seata.Cards, seat.Cards) {
		t.strategy300Bigger = true
	}

	// 给主机b发牌 顺子/同花
	seatb := t.getSeat(b)
	choices = []utils.Choice{}
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.ShunZi})
	choices = append(choices, utils.Choice{Weight: 50, Item: algo.TongHua})
	ch, _ := utils.WeightedChoice(choices)
	cardtype := ch.Item.(uint32)
	seatb.Cards, t.DeskGame.Cards = algo.GetCard(cardtype, t.DeskGame.Cards)

	// 给僚机c发牌 对子/高牌
	if c != 0 {
		seatc := t.getSeat(c)
		choices = []utils.Choice{}
		choices = append(choices, utils.Choice{Weight: 50, Item: algo.DuiZi})
		choices = append(choices, utils.Choice{Weight: 50, Item: algo.GaoPai})
		ch, _ = utils.WeightedChoice(choices)
		cardtype = ch.Item.(uint32)
		seatc.Cards, t.DeskGame.Cards = algo.GetCard(cardtype, t.DeskGame.Cards)
	}

	// 给僚机d发牌 对子/高牌
	if d != 0 {
		seatd := t.getSeat(d)
		choices = []utils.Choice{}
		choices = append(choices, utils.Choice{Weight: 50, Item: algo.DuiZi})
		choices = append(choices, utils.Choice{Weight: 50, Item: algo.GaoPai})
		ch, _ = utils.WeightedChoice(choices)
		cardtype = ch.Item.(uint32)
		seatd.Cards, t.DeskGame.Cards = algo.GetCard(cardtype, t.DeskGame.Cards)
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

	// if t.DeskType == int32(pb.DESK_TYPE_NEWBIEW) { //新手桌
	// 	t.Strategy100Handler()
	// } else if t.DeskType == int32(pb.DESK_TYPE_NORMAL) { //正常桌
	// 	if !t.Strategy100Handler() { //如果没走100，则走200
	// 		t.Strategy200Handler()
	// 	}
	// }

	if t.DeskType == int32(pb.DESK_TYPE_NEWBIEW) { //新手桌 (玩家处于新手状态、平民状态属于新手桌)
		// t.Strategy100Handler(false)
		// t.Strategy200Handler()
	} else {
		// t.Strategy200Handler()
		// t.Strategy300Handler()
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

// 获取唯一玩家seatid
func (t *Desk) GetOnlyOneSeatId() uint32 {
	userid := t.GetOnlyOnePlayer()
	return t.getSeatid(userid)
}

// 获取唯一玩家role
func (t *Desk) GetOnlyOneRole() *data.DeskRole {
	return t.getRole(t.GetOnlyOnePlayer())
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
	t.detail.Rtype = t.Rtype
	t.detail.Gmode = t.Gmode
	t.detail.RoomId = t.Game.Id
	t.detail.DeskId = t.Rid
	t.detail.WaterId = t.GameId
	t.detail.CardTypeId = t.cardTypeId
	t.detail.IsChangeTable = t.isChangeTable
	t.detail.TpStrategy = &data.TpStrategy{}

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
		tp_detail := &data.TPDetail{}
		tp_detail.SeatId = k
		tp_detail.UserId = v.Userid
		// tp_detail.Cards = t.getHandCards(k)
		if role, ok := t.roles[v.Userid]; ok && role != nil {
			tp_detail.BeforeScore = role.GetScore()
			tp_detail.BeforeCash = role.GetDiamond()
			tp_detail.BeforeBonus = role.GetCoin()
		}
		t.detail.TPDetail = append(t.detail.TPDetail, tp_detail)
	}
}

// 初始化
func (t *Desk) gameInit() {
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0), //自由
		int32(pb.ROOM_TYPE1): //私人
		t.DeskGame.GameId = handler.GenGameId(t.Gtype) //初始化流水号
		// 新手配置
		// userid := t.GetOnlyOnePlayer()
		// if userid != "" {
		// 	role := t.roles[userid]
		// 	if role.RegistArea == 0 {
		// 		// 切换为A类新手配置
		// 		t.Game.TP.NewbieMode = t.Game.TP.ANewbieMode
		// 	}
		// }

		t.Ante = uint32(t.Game.TP.Bottom)
		t.score = make(map[uint32]int64)
		t.over = make(map[uint32]*pb.JHCoinOver)
		t.betInfo = make(map[int64]int32)
		t.model = 0
		t.chargingAmounts = make(map[uint32]*data.TpChargeAmount)
		t.robotLeaving = make(map[string]bool)

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
			v.ShowCards = make([]uint32, 0)
			v.Power = 0
			v.Niu = false
			v.Watch = false
			v.TryLeave = false
			v.LeaveWin = false
			v.SideShowBeRejected = false
			v.SeeRate = 0
			if role, ok := t.roles[v.Userid]; ok && role != nil {
				role.Ratio = role.GetRatio()

				// 记录私人房间参与游戏玩家信息,中途退出的在结算时一起展示
				if t.DeskData.Rtype == int32(pb.ROOM_TYPE1) {
					if _, ok := t.DeskPriv.PrivPlayer[role.Userid]; !ok {
						lv := strconv.Itoa(role.Vip.Lv)
						t.DeskPriv.PrivPlayer[role.Userid] = [3]string{role.Nickname, role.Photo, lv}
					}
				}
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

		if t.Rtype == int32(pb.ROOM_TYPE1) {
			t.cardTypeId = 15 // 私人房使用牌型15
		} else {
			// t.cardTypeId = t.findCardTypeId()
		}

	case int32(pb.ROOM_TYPE2): //百人
	}
}

// 结束重置
func (t *Desk) gameOverInit() {
	for _, v := range t.seats {
		t.clearTimeout(v.Userid)
		v.Ready = true
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
		v.Chaal = false
		v.Bet = 0
		v.ActNum = 0
	}
	//重新加载配置
	t.Game = config.GetGame(t.Game.Id)
	t.Ante = uint32(t.Game.TP.Bottom)

	t.DeskGame.Cards = make([]uint32, 0)
	t.DeskGame.BetNum = 0
	t.cardTypeId = 0 //清空牌型ID
	// t.robotNum = 0
	// t.robotNum = t.randRobotNum()
	t.robotNum = t.calcRobotNumber()
	t.robotTime = 0
	t.robotCalling = 0
	// t.robotLeaving = make(map[string]bool)
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
	t.newbieProbeRecord = nil
	t.newbieStableRecord = nil
	t.isChangeTable = false

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
	t.dealer4Random()
	glog.Debugf("dealer -> %s", t.DeskGame.Dealer)
	glog.Debugf("dealer seat -> %d", t.DeskGame.DealerSeat)
	// t.timer = 0
	t.state = int32(pb.STATE_DEALING) //切换为发牌状态

	t.pushDealer() //广播庄家
	t.pushState()  //广播状态

}

// 状态消息
func (t *Desk) pushState() {
	msg := &pb.JHPushStateNtf{
		State: t.state,
	}
	// 测试透视
	msg.UCards = t.playersHand()
	t.broadcast(msg)
}

// 透视
func (t *Desk) playersHand() (r map[string]*pb.JHHandCards) {
	r = make(map[string]*pb.JHHandCards, 0)
	return
	// for _, seat := range t.seats {
	// 	r[seat.Userid] = &pb.JHHandCards{Cards: seat.Cards}
	// }
	// return r
}

// 庄家消息
func (t *Desk) pushDealer() {
	msg := &pb.JHPushDealerNtf{
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
		ante_factor := t.Game.TP.AnteFactor
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
	cards := t.getHandCards(seat)
	typ := algo.HuaType(cards)
	return typ >= algo.DuiZi
}

// 广播操作状态消息
func (t *Desk) pushActState() {
	prev := t.getPrevActSeat()
	curr := t.DeskAct.ActSeat
	next := t.getNextActSeat()
	msg := &pb.JHPushActStateNtf{
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

	// 判断人机行为
	t.robotActState(msg)

	// if t.isStrategy {
	// 	msg.IsCharge = t.isCharge
	// 	msg.Condition = int32(t.condition)
	// }

	t.broadcast(msg)
}

// ' 随机选择庄
func (t *Desk) dealer4Random() {

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

	seat := a[rand.Intn(len(a))]
	if val, ok := t.seats[seat]; ok {
		t.DeskGame.Dealer = val.Userid
		t.DeskGame.DealerSeat = seat
	}
}

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

	randDealer := func() {
		// seat := a[0]
		seat := a[rand.Intn(len(a))]
		if val, ok := t.seats[seat]; ok {
			t.DeskGame.Dealer = val.Userid
			t.DeskGame.DealerSeat = seat
		}
	}

	setDealer := func(seatid uint32) {
		seat := t.seats[seatid]
		t.DeskGame.Dealer = seat.Userid
		t.DeskGame.DealerSeat = seatid
	}

	seatid := t.GetOnlyOneSeatId()
	if seatid == 0 {
		randDealer()
		return
	}

	conf := table.GetTables().TpCardTypeConfigTable.Get(int32(t.cardTypeId))
	if conf == nil {
		glog.Errorf("TpCardTypeConfigTable: cardTypeId not found: ", t.cardTypeId)
		randDealer()
		return
	}

	if conf.PlayerPosition == -1 { //-1为不限制，随机
		randDealer()
		return
	}

	if conf.PlayerPosition == 0 { //0为庄家
		setDealer(seatid)
		return
	}

	count := conf.PlayerPosition //不为0

	//寻找庄家位
	var i uint32 = seatid
	var dealer uint32 = 0
	for {
		if count == 0 {
			break
		}
		i = t.PrevSeat(i)
		if i == seatid { //人数不够则随机
			dealer = 0
			break
		}
		if t.qualified(i) {
			dealer = i
			count--
		}
	}

	if dealer == 0 {
		randDealer()
		return
	}

	setDealer(dealer)

}

// 查找牌型ID
func (t *Desk) findCardTypeId() int {
	var factor_int int

	// for _, role := range t.roles {
	// 	if role.IsA() {
	// 		// A类玩家指定牌型
	// 		return t.Game.TP.ACardType
	// 	}
	// }

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
	for i, v := range t.Game.TP.FinalFactorRange {
		if factor_int < v {
			index = i
			break
		} else {
			index = i + 1
		}
	}
	return t.Game.TP.CardTypeRange[index]
}

// 生成牌
func (t *Desk) getNextCard(count int, current_cards []uint32) (cards []uint32, remain_cards []uint32) {
	cards = make([]uint32, count, count)
	copy(cards, current_cards[0:count])
	remain_cards = current_cards[count:]
	return
}

// 生成牌
func (t *Desk) getCard(robot bool, current_cards []uint32) (cards []uint32, remain_cards []uint32) {
	id := t.cardTypeId
	// id := 19
	glog.Debugf("===============当前牌型id: %d, 房间ID: %s=================", id, t.Game.Id)

	// ct := t.Game.TP.FindCardType(id)
	ct := table.GetTables().TpCardTypeConfigTable.Get(int32(id))
	if ct == nil {
		glog.Error("TpCardTypeConfigTable: cardTypeId not found: ", id)
		return
	}

	var choices []utils.Choice
	if !robot {
		choices = append(choices, utils.Choice{Weight: int(ct.PlayerWeight[0]), Item: algo.BaoziUp})
		choices = append(choices, utils.Choice{Weight: int(ct.PlayerWeight[1]), Item: algo.BaoziDown})
		choices = append(choices, utils.Choice{Weight: int(ct.PlayerWeight[2]), Item: algo.TonghuashunUp})
		choices = append(choices, utils.Choice{Weight: int(ct.PlayerWeight[3]), Item: algo.TonghuashunDown})
		choices = append(choices, utils.Choice{Weight: int(ct.PlayerWeight[4]), Item: algo.ShunziUp})
		choices = append(choices, utils.Choice{Weight: int(ct.PlayerWeight[5]), Item: algo.ShunziDown})
		choices = append(choices, utils.Choice{Weight: int(ct.PlayerWeight[6]), Item: algo.TonghuaUp})
		choices = append(choices, utils.Choice{Weight: int(ct.PlayerWeight[7]), Item: algo.TonghuaDown})
		choices = append(choices, utils.Choice{Weight: int(ct.PlayerWeight[8]), Item: algo.DuiziUp})
		choices = append(choices, utils.Choice{Weight: int(ct.PlayerWeight[9]), Item: algo.DuiziDown})
		choices = append(choices, utils.Choice{Weight: int(ct.PlayerWeight[10]), Item: algo.GaopaiUp})
		choices = append(choices, utils.Choice{Weight: int(ct.PlayerWeight[11]), Item: algo.GaopaiDown})
	} else {
		choices = append(choices, utils.Choice{Weight: int(ct.RobotWeight[0]), Item: algo.BaoziUp})
		choices = append(choices, utils.Choice{Weight: int(ct.RobotWeight[1]), Item: algo.BaoziDown})
		choices = append(choices, utils.Choice{Weight: int(ct.RobotWeight[2]), Item: algo.TonghuashunUp})
		choices = append(choices, utils.Choice{Weight: int(ct.RobotWeight[3]), Item: algo.TonghuashunDown})
		choices = append(choices, utils.Choice{Weight: int(ct.RobotWeight[4]), Item: algo.ShunziUp})
		choices = append(choices, utils.Choice{Weight: int(ct.RobotWeight[5]), Item: algo.ShunziDown})
		choices = append(choices, utils.Choice{Weight: int(ct.RobotWeight[6]), Item: algo.TonghuaUp})
		choices = append(choices, utils.Choice{Weight: int(ct.RobotWeight[7]), Item: algo.TonghuaDown})
		choices = append(choices, utils.Choice{Weight: int(ct.RobotWeight[8]), Item: algo.DuiziUp})
		choices = append(choices, utils.Choice{Weight: int(ct.RobotWeight[9]), Item: algo.DuiziDown})
		choices = append(choices, utils.Choice{Weight: int(ct.RobotWeight[10]), Item: algo.GaopaiUp})
		choices = append(choices, utils.Choice{Weight: int(ct.RobotWeight[11]), Item: algo.GaopaiDown})
	}
	c, _ := utils.WeightedChoice(choices)
	cardtype := c.Item.(int)

	return algo.GetCard4(uint32(cardtype), current_cards)

}

func (t *Desk) roleCountNumReady() (r, n int) { //r真人 n 机器人
	for _, v := range t.seats {
		if !v.Ready {
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

// 获取参与人数
func (t *Desk) GetCount() int {
	r, n := t.roleCountNumReady()
	return r + n
}

// 胜率检测
func (t *Desk) winRateCheck() {
	playerSeatId := t.GetOnlyOneSeatId() //只有一个真人的时候生效
	if playerSeatId == 0 {
		return
	}
	conf := table.GetTables().TpCardTypeConfigTable.Get(int32(t.cardTypeId))
	if conf == nil {
		glog.Error("TpCardTypeConfigTable: cardTypeId not found: ", t.cardTypeId)
		return
	}

	var rate int32
	count := t.GetCount()
	switch count {
	case 2:
		rate = conf.WinRate2
	case 3:
		rate = conf.WinRate3
	case 4:
		rate = conf.WinRate4
	}

	type seatCards struct {
		seat  uint32
		cards []uint32
	}

	var cards []seatCards
	for k, v := range t.seats {
		if !v.Ready {
			continue
		}
		cards = append(cards, seatCards{k, v.Cards})
	}

	getPlayerCard := func() []uint32 {
		for _, v := range cards {
			if v.seat == playerSeatId {
				return v.cards
			}
		}
		return nil
	}

	sort.Slice(cards, func(i, j int) bool {
		return algo.HuaCompare(cards[i].cards, cards[j].cards)
	})

	// 剧情模式玩家必拿最大牌
	// if t.model == 3 {
	// 	rate = 10000
	// }
	if rate > 0 && utils.RandWan(rate) { //拿最大牌
		// t.WinRateValid = true
		if cards[0].seat == playerSeatId {
			return
		}
		playerCards := getPlayerCard()        //获取玩家牌
		playerSeat := t.getSeat(playerSeatId) //获取玩家
		playerSeat.Cards = cards[0].cards     //将最大牌给玩家
		robotSeat := t.getSeat(cards[0].seat) //获取人机
		robotSeat.Cards = playerCards         //将玩家牌给人机

	} else if rate < 0 && utils.RandWan(int32(math.Abs(float64(rate)))) { //拿第三大牌（2人局拿第二大牌）
		// t.WinRateValid = true
		idx := 2
		if count == 2 {
			idx = 1
		}
		if cards[idx].seat == playerSeatId {
			return
		}
		playerSeat := t.getSeat(playerSeatId)   //获取玩家
		playerCards := getPlayerCard()          //获取玩家牌
		playerSeat.Cards = cards[idx].cards     //将第三大牌给玩家
		robotSeat := t.getSeat(cards[idx].seat) //获取人机
		robotSeat.Cards = playerCards           //将玩家牌给人机
	} else {
		return
	}

}

// 胜率检测
// func (t *Desk) winRateCheck() {
// 	id := t.cardTypeId
// 	ct := t.Game.TP.FindCardType(id)
// 	if ct.WinRateCheck == 1 {
// 		r, n := t.roleCountNum()
// 		if r != 1 { //真人数目不是一个，跳过
// 			return
// 		}
// 		if n == 0 { //没有人机，跳过
// 			return
// 		}
// 		if utils.RandWan(int32(ct.PlayerWinRate)) { //玩家赢
// 			max := t.getMaxSeat()
// 			if !t.isRobot(max) { //如果最大玩家是真人，玩家赢，跳过
// 				return
// 			}

// 			//找唯一玩家
// 			userid := t.GetOnlyOnePlayer()
// 			if userid == "" {
// 				return
// 			}

// 			//获取seat
// 			seatid := t.getSeatid(userid)
// 			seat := t.getSeat(seatid)
// 			if seat == nil {
// 				return
// 			}
// 			seat_robot := t.getSeat(max)
// 			if seat_robot == nil {
// 				return
// 			}

// 			//换牌
// 			seat.Cards, seat_robot.Cards = seat_robot.Cards, seat.Cards
// 		} else { //玩家输
// 			max := t.getMaxSeat()
// 			if t.isRobot(max) { //如果最大玩家是机器人，玩家输，跳过
// 				return
// 			}

// 			// 找一个机器人
// 			userid := t.GetOneRobot()
// 			if userid == "" {
// 				return
// 			}

// 			// 获取seat
// 			seatid := t.getSeatid(userid)
// 			seat := t.getSeat(seatid)
// 			if seat == nil {
// 				return
// 			}
// 			seat_player := t.getSeat(max)
// 			if seat_player == nil {
// 				return
// 			}

// 			// 换牌
// 			seat.Cards, seat_player.Cards = seat_player.Cards, seat.Cards
// 		}
// 	}
// }

// '发牌
func (t *Desk) dealRandom() {
	var hand = 3
	for _, v := range t.seats {
		if !v.Ready {
			continue
		}
		cards, remain := t.getNextCard(hand, t.DeskGame.Cards)
		v.Cards = cards
		t.DeskGame.Cards = remain
		v.WinRate = algo.CaclWinRate(cards)
		glog.Debugf("deal -> %s, %v", v.Userid, v.Cards)
	}
}

// '发牌
func (t *Desk) deal() {
	var hand = 3
	for k, v := range t.seats {
		if !v.Ready {
			continue
		}

		v.Cards = make([]uint32, hand, hand)
		v.WinRate = 0
		if r, ok := t.roles[v.Userid]; ok {
			cards, remain := t.getCard(r.Robot, t.DeskGame.Cards)
			copy(v.Cards, cards)
			v.WinRate = algo.CaclWinRate(cards)
			t.SetTpUserFollowRateTrigger(k, int32(algo.HuaType(cards)))
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
func (t *Desk) resCoinOver(score map[uint32]int64) (msg *pb.JHCoinGameoverNtf) {
	t.state = int32(pb.STATE_OVER)
	msg = &pb.JHCoinGameoverNtf{
		Dealer: t.DeskGame.Dealer,
		State:  t.state,
	}
	for _, v := range t.over {
		if t.Game.TP.ShowSwitch == 0 {
			v.Cards = []uint32{}
		} else if !v.Show {
			v.Cards = []uint32{}
		} else {
			seat := t.getSeat(v.Seat)
			if seat != nil {
				showCards := seat.ShowCards
				if len(showCards) != 0 {
					v.Cards = showCards
				}
			}
		}
		msg.Data = append(msg.Data, v)
	}
	return
}

// '私人局结算消息
func (t *Desk) resOver(score map[uint32]int64) (msg *pb.JHGameoverNtf) {
	msg = &pb.JHGameoverNtf{
		Dealer:     t.DeskGame.Dealer,
		DealerSeat: t.DeskGame.DealerSeat,
		Round:      t.DeskData.Round,
		CurRound:   t.DeskGame.Round,
	}
	if t.DeskData.Round > t.DeskGame.Round {
		msg.LeftRound = (t.DeskData.Round - t.DeskGame.Round)
	}
	for k, v := range score {
		d := &pb.JHRoomOver{
			Seat:  k,
			Score: v,
		}
		if seat, ok := t.seats[k]; ok {
			d.Bets = seat.Bet
			d.Value = seat.Power
			//d.Cards = val.Cards
			d.Total = t.DeskPriv.PrivScore[seat.Userid]
			if p, ok2 := t.roles[seat.Userid]; ok2 {
				d.Coin = p.User.GetCoin()
				d.Fraction = p.User.GetFraction()
				d.Nickname = p.User.GetNickname()
				d.Photo = p.User.GetPhoto()
			}

			// 手牌
			over, ok := t.over[k]
			if !ok || over == nil {
				continue
			}
			if t.Game.TP.ShowSwitch == 0 {
				d.Cards = []uint32{}
			} else if !over.Show {
				d.Cards = []uint32{}
			} else {
				if seat != nil {
					showCards := seat.ShowCards
					if len(showCards) != 0 {
						d.Cards = showCards
					} else {
						d.Cards = seat.Cards
					}
				}
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
	// t.SetChargeCondition()
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
	msg := new(pb.JHCoinAllBiNtf)
	for k, v := range t.DeskAct.ActSeats {
		if v.Alive {
			seats = append(seats, k)
			cards := t.getHandCards(k)
			typ := algo.HuaType(cards)

			info := &pb.JHCoinAllBiInfo{
				Seat:  k,
				Cards: cards,
				Type:  typ,
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
		if algo.HuaCompare(cs1, cs2) { //k赢 winner输
			if val, ok := t.DeskAct.ActSeats[winner]; ok {
				val.Alive = false
				t.Lose(winner)
			}
			winner = k
		} else { //winner赢 k输
			v.Alive = false
			t.Lose(k)
		}
	}
	msg.Winner = winner

	others := t.GetOtherSeats(seats)
	msg2 := new(pb.JHCoinAllBiNtf)
	utils.Clone(msg2, msg)
	for _, v := range msg2.Info {
		v.Cards = []uint32{}
	}

	t.broadcast4(seats, msg)
	t.broadcast4(others, msg2)

	//等待播放比牌动画
	// timerChan1 := time.After(2 * time.Second)
	// <-timerChan1
	t.pauseGame(PAUSE_REASON_3, 2, winner)
}

// 离开检查
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
			msg := new(pb.JHCoinSeeRsp)
			msg.Cards = v.Cards
			msg.Typ = algo.HuaType(msg.Cards)
			t.send2userid(v.Userid, msg)

			// 广播看牌不用All功能
			msg2 := new(pb.JHCoinSeeNtf)
			msg2.Seat = k
			msg2.Userid = v.Userid
			msg2.Actseat = t.DeskAct.ActSeat
			msg2.Actstate = uint32(t.DeskAct.ActState)
			t.broadcast(msg2)
		}
	}

	// msg2 := new(pb.JHCoinSeeNtf)
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
	if t.DeskGame.BetNum >= int64(t.Game.TP.Pool_Limit) {
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
		if t.DeskAct.ActTimes == 1 {
			for k, v := range t.seats {
				if !v.Ready {
					continue
				}
				status := t.getStatus(k)
				if status == nil {
					continue
				}
				if !status.Alive {
					continue
				}
				t.SetTpUserFollowRateSuccess(k, int32(algo.HuaType(v.Cards)))
			}
		}
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
	// t.BetCheckWinScore(t.DeskAct.ActSeat)
	//设置下家操作值
	t.setNextActState()
	//设置局内充值条件
	// t.SetChargeCondition()
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
	if t.DeskAct.ActTimes < int32(t.Game.TP.Than_Rounds) && t.remainPlayerNum() > 2 {
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

	//sideshow show标记判断
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

	msg := new(pb.JHCoinRaiseAnteNtf)
	for k, v := range t.seats {
		if !v.Ready {
			continue
		}
		// d := new(data.ActStatus)
		// d.Alive = true
		// t.DeskAct.ActSeats[k] = d

		//开始时下暗注
		info := t.setBet(k, v.Userid, int64(t.DeskData.Ante), fmt.Sprintf("tp%s房间下底注", t.DeskData.Rid))
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
	// t.SetChargeCondition()
	//广播下家操作状态消息
	t.pushActState()
}

// 下注成功设置
func (t *Desk) setBet(seat uint32, userid string, num int64, desc string) *pb.JHCoinRaiseInfo {
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
func (t *Desk) setBetMsg(seat uint32, userid string, num int64, desc string) *pb.JHCoinRaiseInfo {
	msg := &pb.JHCoinRaiseInfo{
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
		t.sendCurrency(userid, (-1 * num), int32(pb.LOG_TYPE110), desc)
	case int32(pb.ROOM_TYPE1): //私人
		switch t.Gmode {
		case 0: // 真金
			t.sendCurrency(userid, (-1 * num), int32(pb.LOG_TYPE110), desc)
		case 1: // 娱乐模式
			t.sendFraction(userid, (-1 * num), int32(pb.LOG_TYPE110))
		}
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
					sign = sign << 1
					if r.Hedge&sign != 0 {
						// 触发过冤家牌
						hedge = true
					}
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
		baozi, cards := algo.GetCard(algo.BaoZi, t.DeskGame.Cards)
		t.DeskGame.Cards = cards
		ths, cards := algo.GetCard(algo.TongHuaShun, t.DeskGame.Cards)
		t.DeskGame.Cards = cards
		s := t.getSeat(maxSeat)
		s.Cards = ths
		s = t.getSeat(minSeat)
		s.Cards = baozi
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

// vim: set foldmethod=marker foldmarker=//',//.:
