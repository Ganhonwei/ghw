package rm2

import (
	"context"
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
			(ltype == int32(pb.LOG_TYPE127) || ltype == int32(pb.LOG_TYPE128)) {
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
		(ltype == int32(pb.LOG_TYPE127) || ltype == int32(pb.LOG_TYPE128)) {
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
		(ltype == int32(pb.LOG_TYPE127) || ltype == int32(pb.LOG_TYPE128)) {
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
	if t.state == int32(pb.STATE_DEALING) || t.state == int32(pb.STATE_BET) || t.state == int32(pb.STATE_PAUSE) || t.state == int32(pb.STATE_DECLARE) {
		return true
	} else {
		return false
	}
}

// 进入房间响应消息
func (t *Desk) coinEnterMsg(userid string) *pb.RMCoinEnterRoomRsp {
	msg := new(pb.RMCoinEnterRoomRsp)
	//房间数据
	msg.Roominfo = handler.PackRMCoinRoom(t.DeskData)
	msg.Roominfo.State = t.state //状态
	msg.Gameid = t.Game.Id
	//坐下玩家信息
	msg.Userinfo = t.coinSeatBetsMsg(userid)
	//位置下注信息
	msg.Betsinfo = t.coinBetsMsg()
	// 房间暂停状态原因
	if t.state == int32(pb.STATE_PAUSE) {
		msg.Roominfo.StatePauseReaon = int32(t.reason)
	}
	if t.DeskAct != nil {
		msg.Actseat = t.DeskAct.ActSeat //当前操作玩家的座位号
		msg.Actstate = t.DeskAct.ActState
		msg.Timer = int32(BetTime - t.timer) //计时器
		msg.Ante = uint32(t.DeskAct.ActAnte)
		msg.Pot = t.DeskGame.BetNum
		msg.Dealer = t.DeskGame.DealerSeat //庄家位置
		msg.Totaltimer = BetTime
	}
	if t.DeskGame != nil {
		msg.WildCard = t.DeskGame.WildCard //万能牌
		if len(t.DeskGame.QiCards) != 0 {
			msg.QiCard = t.DeskGame.QiCards[len(t.DeskGame.QiCards)-1] //弃牌堆
		}
	}
	if t.state == int32(pb.STATE_DECLARE) {
		msg.HuSeat = t.getWinner()
	}
	//假位置
	for k := range t.FakeSeats {
		msg.Roominfo.FakeSeats = append(msg.Roominfo.FakeSeats, k)
	}
	return msg
}

// 进入消息
func (t *Desk) coinCameinMsg(userid string) {
	msg := new(pb.RMCameinNtf)
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
	return utils.RandMN(t.Game.RM.Single_Robot[0], t.Game.RM.Single_Robot[1])
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
	n += t.robotCalling // 正在召唤中机器人

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
			t.notifyGateUserLeft(robotid, pb.OK, 0)
			t.userLeaveDesk(robotid)
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

// 位置上玩家数据
func (t *Desk) coinRoleMsg(userid string) (msg *pb.RMRoomUser) {
	if v, ok := t.roles[userid]; ok {
		if v.Seat == 0 {
			return //没有坐下不广播
		}
		msg = handler.PackRMCoinUser(v.User)
		msg.Seat = v.Seat
		msg.Offline = v.Offline
		if val, ok := t.seats[v.Seat]; ok {
			msg.Dealer = val.BeDealer
			msg.Bet = val.Bet
			msg.Num = val.DealerN
			msg.Niu = val.Niu
			msg.Ready = val.Ready
			msg.Watch = val.Watch
			msg.FinishCard = val.Finish
			msg.Declare = val.Declare

			if t.DeskAct != nil && !val.Watch {
				if val, ok := t.DeskAct.ActSeats[v.Seat]; ok {
					msg.Pack = val.Pack
					msg.Lose = val.Lose
					msg.See = val.See
					msg.Bet2 = val.Bet
					msg.Totalbet = val.ActNum
					msg.Drop = val.Drop
					msg.Zhahu = val.ZhaHu
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
func (t *Desk) coinSeatBetsMsg(userid string) (msg []*pb.RMRoomUser) {
	for k, v := range t.seats {
		msg2 := t.coinRoleMsg(v.Userid)
		if msg2 == nil {
			continue
		}
		//自己手牌
		if v.Userid == userid {
			msg2.Cards = t.getHandCards(k)
			sort := t.getSortCards(k)
			for _, v := range sort {
				msg2.Sort = append(msg2.Sort, &pb.RMSortCard{Cards: v})
			}
		}
		msg = append(msg, msg2)
	}
	return
}

// 玩家下注数据
func (t *Desk) coinBetsMsg() (msg []*pb.RMRoomBets) {
	for k, v := range t.seats {
		msg2 := &pb.RMRoomBets{
			Seat: k,
			Bets: v.Bet,
		}
		msg = append(msg, msg2)
	}
	return
}

// 是否全部发牌动画播放完成
func (t *Desk) allReady2() bool {
	var num1 int = t.ready2Num() //播放完发牌动画人数
	var num2 int = t.readyNum()  //游戏人数
	return num1 == num2
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

// 操作超时放弃
func (t *Desk) betTimeout() {
	seatid := t.DeskAct.ActSeat
	userid := t.getUserid(seatid)
	seat := t.getSeat(seatid)

	if len(seat.Cards) == 13 { //没摸没打
		t.drawCard(userid, 1, true)
		card := seat.Cards[len(seat.Cards)-1] //打摸得牌
		t.discard(userid, card, true)
	} else if len(seat.Cards) == 14 { //摸了没打
		if seat.Sort {
			cards := seat.SortCards[len(seat.SortCards)-1]
			card := cards[len(cards)-1]
			// 剔除排序牌组中的牌
			seat.SortCards[len(seat.SortCards)-1] = cards[:len(cards)-1]
			if len(cards) == 1 {
				seat.SortCards = seat.SortCards[:len(seat.SortCards)-1]
			}
			t.discard(userid, card, true)
		} else {
			card := seat.Cards[len(seat.Cards)-1]
			t.discard(userid, card, true)
		}
	}

	// if t.remainPlayerNum() > 1 {
	// 	if t.DeskAct.ActState&int32(pb.ACT_REPLY_BI) == int32(pb.ACT_REPLY_BI) { //比牌操作超时
	// 		seat := t.DeskAct.ActSeat
	// 		userid := t.getUserid(seat)
	// 		t.coinReplyBi(userid, false)
	// 	} else { //下注操作超时
	// 		seat := t.DeskAct.ActSeat
	// 		userid := t.getUserid(seat)
	// 		t.coinFold(userid)
	// 	}
	// }
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
					for i, v := range t.Game.RM.NewbieMode.CivilianCanWithdrawRange {
						if diamond < int64(v) {
							index = i
							break
						} else {
							index = i + 1
						}
					}

					if index > len(t.Game.RM.NewbieMode.CivilianCanWithdrawRange)-1 {
						index = len(t.Game.RM.NewbieMode.CivilianCanWithdrawRange) - 1
					}

					winRate := t.Game.RM.NewbieMode.CivilianWinRate[index]
					win := utils.RandWan(int32(winRate))
					if win {
						weight := t.Game.RM.NewbieMode.CivilianWinWeight
						var choices []utils.Choice
						for i, v := range weight {
							choices = append(choices, utils.Choice{Weight: v, Item: i})
						}
						c, _ := utils.WeightedChoice(choices)
						i := c.Item.(int)
						return t.Game.RM.NewbieMode.CivilianWinCardType[i]
					} else {
						weight := t.Game.RM.NewbieMode.CivilianLoseWeight
						var choices []utils.Choice
						for i, v := range weight {
							choices = append(choices, utils.Choice{Weight: v, Item: i})
						}
						c, _ := utils.WeightedChoice(choices)
						i := c.Item.(int)
						return t.Game.RM.NewbieMode.CivilianLoseCardType[i]
					}
				} else {
					diamond := r.OutDiamond
					index := 0
					for i, v := range t.Game.RM.NewbieMode.CanWithdrawRange {
						if diamond < int64(v) {
							index = i
							break
						} else {
							index = i + 1
						}
					}

					if index > len(t.Game.RM.NewbieMode.CanWithdrawRange)-1 {
						index = len(t.Game.RM.NewbieMode.CanWithdrawRange) - 1
					}

					winRate := t.Game.RM.NewbieMode.WinRate[index]
					win := utils.RandWan(int32(winRate))
					if win {
						weight := t.Game.RM.NewbieMode.WinWeight
						var choices []utils.Choice
						for i, v := range weight {
							choices = append(choices, utils.Choice{Weight: v, Item: i})
						}
						c, _ := utils.WeightedChoice(choices)
						i := c.Item.(int)
						return t.Game.RM.NewbieMode.WinCardType[i]
					} else {
						weight := t.Game.RM.NewbieMode.LoseWeight
						var choices []utils.Choice
						for i, v := range weight {
							choices = append(choices, utils.Choice{Weight: v, Item: i})
						}
						c, _ := utils.WeightedChoice(choices)
						i := c.Item.(int)
						return t.Game.RM.NewbieMode.LoseCardType[i]
					}
				}

			}
		}
	}
	return 11
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

	if t.Game.RM.RoomFactorSwitch == 0 {
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
	t.alive()
	//选庄
	t.dealer()
	// if t.Game.RM.Mode == 2 {
	// 	t.deal2()
	// } else {
	// 	t.deal1()
	// }

	t.deal3()

	//消息推送
	t.pushMsg()
	//详情初始化
	t.detailInit()
	//扣底注
	// t.raiseAnte()

	//机器人处理
	// t.robotHandler()

}

func (t *Desk) getRoiControlRecord(user *data.User) (r RMRoiStrategy, ok bool) {
	if user == nil {
		return
	}
	// 玩家收益率
	var winRate float64
	if user.Money > 0 {
		winRate = (float64(user.CashOut) + float64(user.OutDiamond)) / float64(user.Money)
	}
	// 玩家筹码是房间低分多少倍
	dieRate := float64(user.GetScore()) / float64(t.Game.RM.Bottom)
	zlog.Infof("roi生效判定: %s, rc=%v, diamond=%v, winRate=%v, dieRate=%v", user.Userid, len(user.RechargeTarge), float64(user.GetScore()), winRate, dieRate)
	for _, roi := range rmRoiConfig.Roi {
		// 充值次数范围
		if roi.RcCountRange != "" && !handler.AnalysisFreeStrategyParam(roi.RcCountRange, float64(len(user.RechargeTarge))) {
			zlog.Infof("roi1 continue: %v, %v", roi.RcCountRange, len(user.RechargeTarge))
			continue
		}
		// 钱包携带金币数量
		if roi.AmountRange != "" && !handler.AnalysisFreeStrategyParam(roi.AmountRange, float64(user.GetScore())) {
			zlog.Infof("roi2 continue: %v, %v", roi.AmountRange, float64(user.GetScore()))
			continue
		}
		// 策略id总生效次数
		var limit, dayLimit int32
		if user.RmRoiLimits != nil {
			limit = user.RmRoiLimits[roi.Id]
		}
		if user.RmRoiDayLimits != nil {
			dayLimit = user.RmRoiDayLimits[roi.Id]
		}
		zlog.Infof("roi生效limit判定: %s, %s, limit=%d, dayLimit=%d, %d,%d", user.Userid, roi.Id, limit, dayLimit, roi.Limit, roi.DayLimit)
		if roi.Limit > 0 && roi.Limit < limit {
			continue
		}
		if roi.DayLimit > 0 && roi.DayLimit < dayLimit {
			continue
		}

		// 玩家收益率
		if roi.WinRateRange != "" && !handler.AnalysisFreeStrategyParam(roi.WinRateRange, winRate) {
			zlog.Infof("roi3 continue: %v, %v", roi.WinRateRange, winRate)
			continue
		}
		// 玩家筹码是房间低分多少倍
		if roi.IsDie != "" && !handler.AnalysisFreeStrategyParam(roi.IsDie, dieRate) {
			zlog.Infof("roi4 continue: %v, %v", roi.IsDie, dieRate)
			continue
		}

		// 匹配成功
		ok = true
		r = roi
		zlog.Infof("roi生效: %s, %s, %#v", user.Userid, roi.Id, roi)
		return
	}
	return
}

// 随机发牌
func (t *Desk) deal3() {
	ctrl := &RmControl{}
	t.control = ctrl
	playerid := t.GetOnlyOnePlayer()
	player := t.getPlayer(playerid)
	roi, roiOK := t.getRoiControlRecord(player)
	// ROI控制
	if roiOK {
		zlog.Infof("roi匹配ok: %v, %#v", roiOK, roi)
		ctrl.Ctype = 2
		ctrl.WinRate = roi.WinRate
		ctrl.NotDrawPlayer1st = roi.NotDrawPlayer1st
		ctrl.NotDrawRobot1st = roi.NotDrawRobot1st
		ctrl.CtrlMin = roi.CtrlMin
		ctrl.CtrlMax = roi.CtrlMax
		ctrl.EffctRate = roi.EffctRate

		if err := json.Unmarshal([]byte(roi.DrawNumPlayer), &ctrl.DrawNumPlayer); err != nil {
			zlog.Errorf("DrawNumPlayer error: %s", roi.DrawNumPlayer)
			roiOK = false
		}
		if err := json.Unmarshal([]byte(roi.DrawNumRobot), &ctrl.DrawNumRobot); err != nil {
			zlog.Errorf("DrawNumRobot error: %s", roi.DrawNumRobot)
			roiOK = false
		}
		if err := json.Unmarshal([]byte(roi.Hand), &ctrl.Hand); err != nil {
			zlog.Errorf("Hand error: %s", roi.Hand)
			roiOK = false
		}
		if roiOK {
			roiOK = utils.RandWan(int32(math.Abs(float64(ctrl.WinRate))))
			zlog.Infof("roi winRate: %v, %v, %#v", roiOK, int32(math.Abs(float64(ctrl.WinRate))), ctrl)
		}
	}

	// 走库存控制
	if !roiOK {
		// 获取房间当前库存
		req := &pb.ChangeStock{GameId: t.Game.Id, Gtype: -1}
		res := t.reqRoom(req)
		var stock *pb.ChangedStock
		if res != nil {
			stock = res.(*pb.ChangedStock)
		}

		// 匹配库存控制
		var control *tb.RmStockControlRecord
		if stock != nil {
			for _, c := range table.GetTables().RmStockControlTable.GetDataList() {
				// 判断库存档位进行控制匹配
				if c.Gid == t.Game.Id && stock.CashStock >= c.StockMin && stock.CashStock < c.StockMax {
					control = c
					break
				}
			}
		}

		if stock != nil {
			zlog.Infof("rm房间%s, 当前库存: %d, control: %#v", t.Game.Id, stock.CashStock, control)
		}

		if control != nil {
			ctrl.Ctype = 1
			ctrl.WinRate = control.WinRate
			ctrl.NotDrawPlayer1st = false
			ctrl.NotDrawRobot1st = control.NotDrawRobot1st
			ctrl.DrawNumPlayer = control.DrawNumPlayer
			ctrl.DrawNumRobot = control.DrawNumRobot
			ctrl.Hand = []int32{control.Hand05, control.Hand59, control.Hand1020}
			ctrl.CtrlMin = control.CtrlMin
			ctrl.CtrlMax = control.CtrlMax
			ctrl.EffctRate = control.EffctRate
			ctrl.Hierarchy = control.Hierarchy
		}
	}

	var hasRobot bool
	var players string
	for _, role := range t.roles {
		if role.Robot {
			hasRobot = true
		}
		players += role.Userid + ","
	}

	// roi模式生效或走库存控制
	if hasRobot && ((ctrl.Ctype == 2 && roiOK) || (ctrl.Ctype == 1 && utils.RandWan(int32(math.Abs(float64(ctrl.WinRate)))))) {
		zlog.Infof("控制发牌: %s", players)
		ctrl.controlEffect = true
		t.deal3Control(ctrl)
	} else {
		zlog.Infof("随机发牌: %s", players)
		ctrl.Ctype = 0
		ctrl.controlEffect = false
		t.deal3Random()
	}

	// 记录
	t.RMControl.Ctype = ctrl.Ctype
	t.RMControl.Hierarchy = ctrl.Hierarchy
	t.RMControl.ControlEffect = ctrl.controlEffect
	t.RMControl.NotDrawRobot1st = ctrl.NotDrawRobot1st
	t.RMControl.NotDrawPlayer1st = ctrl.NotDrawPlayer1st
	if ctrl.Ctype == 2 {
		t.RMControl.RoiId = roi.Id
	}
}

// 随机发牌
func (t *Desk) deal3Random() {
	// 洗牌
	t.shuffle()
	// 插入joker
	t.joker()

	// 随机各发13张牌
	var hand = 13
	for _, v := range t.seats {
		if !v.Ready {
			continue
		}
		v.Cards = make([]uint32, hand)
		if r, ok := t.roles[v.Userid]; ok {
			_ = r
			cards := t.DeskGame.Cards[:hand]
			copy(v.Cards, cards)
			t.DeskGame.Cards = t.DeskGame.Cards[hand:]
		}
	}
}

func (t *Desk) deal3Control(ctrl *RmControl) {
	drawNumPlayer := utils.RandMN(int(ctrl.DrawNumPlayer[0]), int(ctrl.DrawNumPlayer[1]))
	drawNumRobot := utils.RandMN(int(ctrl.DrawNumRobot[0]), int(ctrl.DrawNumRobot[1]))
	// 记录
	t.RMControl.DrawNumPlayer = drawNumPlayer
	t.RMControl.DrawNumRobot = drawNumRobot
	t.RMControl.NotDrawRobot1st = ctrl.NotDrawRobot1st
	t.RMControl.NotDrawPlayer1st = ctrl.NotDrawPlayer1st
	t.RMControl.Hierarchy = ctrl.Hierarchy

	// 洗牌
	t.shuffle()
	// 插入joker
	t.joker()

	// 在104张牌（不含大小王）内，随机的生成两幅胡牌的牌。
	// 准天胡牌型
	huCt := []string{"N01", "N01", "N01", "N01", "E01", "F01", "E02", "F0E", "E03", "F03", "N02", "N02", "G04"}
	var players, robots []uint32
	for seat, v := range t.seats {
		if !v.Ready {
			continue
		}

		v.Cards = make([]uint32, len(huCt))
		if r, ok := t.roles[v.Userid]; ok {
			_ = r
			cards, remain := algo.RMGetCard(huCt, t.DeskGame.Cards, t.WildCard)
			copy(v.Cards, cards)
			t.DeskGame.Cards = remain

			if r.Robot {
				robots = append(robots, seat)
			} else {
				players = append(players, seat)
			}
		}
	}

	// 抽牌换牌，先机器人
	seats := make([]uint32, 0, len(t.seats))
	seats = append(seats, robots...)
	seats = append(seats, players...)
	for _, seat := range seats {
		v := t.seats[seat]
		if !v.Ready {
			continue
		}

		if r, ok := t.roles[v.Userid]; ok {
			var drawNum int
			if r.Robot {
				drawNum = drawNumRobot
			} else {
				drawNum = drawNumPlayer
			}
			var notDraw1st = (r.Robot && ctrl.NotDrawRobot1st) || (!r.Robot && ctrl.NotDrawPlayer1st)
			actTimeCards := [3]int32{ctrl.Hand[0], ctrl.Hand[1], ctrl.Hand[2]}
			zlog.Infof("换牌前: %d, %d, %s", len(t.DeskGame.Cards), len(v.Cards), algo.CardsString(v.Cards))
			v.Cards, t.DeskGame.Cards = algo.DrawNeedCards(r.Robot, v.Cards, t.DeskGame.Cards, t.WildCard, drawNum, notDraw1st, actTimeCards)
			zlog.Infof("换牌后:  %d, %d, %s", len(t.DeskGame.Cards), len(v.Cards), algo.CardsString(v.Cards))
		}
	}
}

func (t *Desk) deal1() {
	// 洗牌
	t.shuffle()
	// 插入joker
	t.joker()
	// 发牌
	t.deal()
}

func (t *Desk) deal2() {
	convertIndex := func(index string) string {
		chars := strings.Split(index, "")

		nums := []int{}
		for _, char := range chars {
			num, _ := strconv.Atoi(char)
			nums = append(nums, num)
		}
		sort.Ints(nums)

		var ret string
		for _, num := range nums {
			ret += strconv.Itoa(num)
		}
		return ret
	}

	type Group struct {
		Task   string     `json:"index" bson:"index"`
		Joker  uint32     `json:"joker" bson:"joker"`
		Seats  [][]uint32 `json:"seats" bson:"seats"`
		Remain []uint32   `json:"remain" bson:"remain"`
	}

	poolSize := 1000 //牌库大小

	ct := t.Game.RM.FindCardType(t.cardTypeId)           //发牌牌型配置
	idIndex := utils.RandSliceIndex(ct.CardPoolIdWeight) //id索引
	id := ct.CardPoolId[idIndex]                         //使用id索引获取id

	//升档牌库判断
	player := t.GetOnlyOnePlayer()
	playerRole := t.getRole(player)
	if playerRole != nil && playerRole.DropCount >= 3 {
		id = ct.UpCardPool
		t.isUpCardPool = true //记录是否升档牌库
	}
	t.cardPoolId = id //记录牌库编号

	idString := strconv.Itoa(id)     //id转为string
	sortId := convertIndex(idString) //排序后的id(从小到大)
	id2 := utils.RandIntN(poolSize)  //随机牌库内的某套牌id
	key := fmt.Sprintf("%s:%d", sortId, id2)

	glog.Debugf("===============当前牌型: %s, 房间ID: %d=================", key, t.Game.Id)

	val, err := redisClient.Get(context.Background(), key).Result()
	if err != nil {
		glog.Errorf("redis get error:%#v", err)
		return
	}
	group := &Group{}
	err = json.Unmarshal([]byte(val), group)
	if err != nil {
		glog.Errorf("json unmarshal error:%#v", err)
		return
	}

	restoreSeats := func(char string, sortChars []string, seats [][]uint32) (ok bool, seat []uint32, afterSortChars []string, afterSeats [][]uint32) {
		var index int
		for k, v := range sortChars {
			if v == char {
				index = k
				ok = true
				break
			}
		}

		if !ok {
			return
		}

		seat = seats[index]
		afterSortChars = append(sortChars[:index], sortChars[index+1:]...)
		afterSeats = append(seats[:index], seats[index+1:]...)
		return
	}

	shuffleCards := func(cards []uint32) {
		for i := len(cards) - 1; i > 0; i-- {
			j := utils.RandIntN(i + 1)
			cards[i], cards[j] = cards[j], cards[i]
		}
	}

	shuffleSeats := func(seats [][]uint32) {
		for i := len(seats) - 1; i > 0; i-- {
			j := utils.RandIntN(i + 1)
			seats[i], seats[j] = seats[j], seats[i]
		}
	}

	_ = shuffleCards

	//还原牌顺序
	seats := [][]uint32{}
	chars := strings.Split(idString, "")   //原id
	sortChars := strings.Split(sortId, "") //排序后id
	for _, char := range chars {
		var seat []uint32
		var ok bool
		ok, seat, sortChars, group.Seats = restoreSeats(char, sortChars, group.Seats)
		if !ok {
			glog.Error("restore seats error")
			return
		}
		seats = append(seats, seat)
	}

	randomDeal := func() { //随机发牌
		glog.Debugf("random deal")
		shuffleSeats(seats)                       //洗一下牌库
		t.DeskGame.WildCard = group.Joker         //设置万能牌
		t.DeskGame.Cards = group.Remain           //设置牌堆
		shuffleCards(t.DeskGame.Cards)            //洗一下牌堆
		t.DeskGame.QiCards = t.DeskGame.Cards[:1] //弃牌堆第一张牌
		t.DeskGame.Cards = t.DeskGame.Cards[1:]

		//发牌
		for k, v := range t.seats {
			if !v.Ready {
				continue
			}
			v.Cards = seats[k-1]
		}
		t.dealType = 1
	}

	configDeal := func() bool { //按配置发牌
		glog.Debugf("config deal")

		player := t.GetOnlyOnePlayer()
		if player == "" {
			glog.Error("get only on player error")
			return false
		}

		playerSeat := t.getSeatid(player)
		if playerSeat == 0 {
			glog.Error("get player seat error")
			return false
		}

		coreRobot := t.GetCoreRobot()
		if coreRobot == 0 {
			glog.Error("get core robot error")
			return false
		}

		ok := t.SetCoreRobot(coreRobot)
		if !ok {
			glog.Error("set core robot error")
			return false
		}

		t.coreRobotSeat = coreRobot

		t.DeskGame.WildCard = group.Joker         //设置万能牌
		t.DeskGame.Cards = group.Remain           //设置牌堆
		shuffleCards(t.DeskGame.Cards)            //洗一下牌堆
		t.DeskGame.QiCards = t.DeskGame.Cards[:1] //弃牌堆第一张牌
		t.DeskGame.Cards = t.DeskGame.Cards[1:]

		for k, v := range t.seats {
			if !v.Ready {
				continue
			}

			if k == playerSeat { //玩家牌
				v.Cards = seats[0]
			} else if k == coreRobot { //核心人机牌
				v.Cards = seats[1]
			} else { //其他人机牌
				v.Cards = seats[k-1]
			}
		}
		t.dealType = 2
		return true
	}

	r, n := t.roleCountNumReady()

	if t.DeskType == int32(pb.DESK_TYPE_NORMAL) && n == 0 { //正常模式并且没有人机随机分配
		randomDeal()
	} else { //其他情况按配置顺序分配（玩家、核心人机、其他人机）
		ok := configDeal()
		if !ok {
			randomDeal()
		}
	}

	//必输局、必输分判断
	if r == 1 && n > 0 && utils.RandWan(int32(ct.MustLoseProb)) {
		t.isMustLose = true
		t.mustLoseScore = utils.RandMN(ct.MustLoseScore[0], ct.MustLoseScore[1]) //随机必输分
		t.mustLoseCoreRobotGoodCardProb = ct.MustLoseCoreRobotGoodCardProb
	}

	t.winScoreLimit = ct.WinScoreLimit
}

// 判断是否触发必输控制
func (t *Desk) isTriggerMustLoseControl() (ret bool) {
	defer func() {
		if ret && !t.isTriggerMustLose {
			t.isTriggerMustLose = true
		}
	}()

	if !t.isMustLose { //不是必输局
		return
	}

	return true
	// userid := t.GetOnlyOnePlayer()
	// if userid == "" {
	// 	return
	// }

	// seatid := t.getSeatid(userid)
	// seat := t.getSeat(seatid)
	// if seat == nil {
	// 	return
	// }

	// sort := algo.SortCards4(seat.Cards, t.WildCard)
	// score := algo.CalcScore(sort, t.WildCard)

	// if score <= int64(t.mustLoseScore) {
	// 	ret = true
	// 	return
	// } else {
	// 	return
	// }
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

// 设置核心人机
func (t *Desk) SetCoreRobot(seatid uint32) bool {
	seat := t.getSeat(seatid)
	if seat == nil {
		return false
	}
	seat.Core = true
	return true
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

// 机器人处理
func (t *Desk) robotHandler() {
	// if t.getRobotNum() == 0 {
	// 	return
	// }
	// max := t.getMaxPlayer()
	// _ = max

	// id := t.cardTypeId
	// ct := t.Game.RM.FindCardType(id)
	// rsg := t.Game.RM.FindRobotStrategyGroup(ct.RobotStrategyGroup)

	// for k, v := range t.seats {
	// 	if k == max {
	// 		continue
	// 	}
	// 	if !v.Ready {
	// 		continue
	// 	}
	// 	if role, ok := t.roles[v.Userid]; ok && role != nil {
	// 		if role.Robot {
	// 			cs1 := t.getHandCards(k)
	// 			cs2 := t.getHandCards(max)

	// 			var rs data.TPGameRobotStrategy
	// 			if algo.HuaCompare(cs1, cs2) { //人机手牌比最大玩家大
	// 				rs = t.Game.RM.FindRobotStrategy(rsg.Bigger)
	// 			} else { //人机手牌比最大玩家小
	// 				rs = t.Game.RM.FindRobotStrategy(rsg.Smaller)
	// 			}

	// 			msg := &pb.RMCoinRobotStrategyNtf{}
	// 			msg.Id = rs.Id
	// 			for _, v := range rs.ActionWeight {
	// 				msg.ActionWeight = append(msg.ActionWeight, &pb.RMCoinActionWeight{Values: v})
	// 			}
	// 			msg.ActionTime = rs.ActionTime
	// 			msg.SeeWeight = rs.SeeWeight
	// 			for _, v := range rs.SeeTime {
	// 				msg.SeeTime = append(msg.SeeTime, &pb.RMCoinSeeTime{Values: v})
	// 			}
	// 			t.send2userid(v.Userid, msg)
	// 		}
	// 	}
	// }
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
	t.detail.RoomId = t.Game.Id
	t.detail.DeskId = t.Rid
	t.detail.WaterId = t.GameId
	t.detail.CardTypeId = t.cardTypeId

	// todo 控制模式
	// t.controlEffect
	// t.control

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
		rm_detail := &data.RMDetail{}
		rm_detail.SeatId = k
		rm_detail.UserId = v.Userid
		rm_detail.InitCards = make([]uint32, len(v.Cards))
		copy(rm_detail.InitCards, v.Cards)
		rm_detail.WildCard = t.WildCard
		if role, ok := t.roles[v.Userid]; ok && role != nil {
			rm_detail.BeforeScore = role.GetScore()
			rm_detail.BeforeCash = role.GetDiamond()
			rm_detail.BeforeBonus = role.GetCoin()
		}
		t.detail.RMDetail = append(t.detail.RMDetail, rm_detail)
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
				t.Game.RM.NewbieMode = t.Game.RM.ANewbieMode
			}
		}

		t.score = make(map[uint32]int64)
		t.over = make(map[uint32]*pb.RMCoinGameOverInfo)
		t.isMustLose = false
		t.mustLoseScore = 0
		t.mustLoseCoreRobotGoodCardProb = 0
		t.winScoreLimit = 0
		t.coreRobotSeat = 0
		t.isUpCardPool = false
		t.cardPoolId = 0
		t.dealType = 0
		t.isTriggerWinScoreLimit = false
		t.isTriggerMustLose = false
		t.isTriggerMustLoseDrawCard = false
		t.isTriggerMustLoseLastCard = false
		t.isTriggerMustLoseCoreRobotGoodCard = false
		t.RMControl = new(data.RMControl)

		t.DeskAct = new(data.DeskAct)
		t.DeskAct.ActSeats = make(map[uint32]*data.ActStatus)
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
			v.Power = 0
			v.Niu = false
			v.Watch = false
			if role, ok := t.roles[v.Userid]; ok && role != nil {
				role.Ratio = role.GetRatio()

				// 记录私人房间参与游戏玩家信息,中途退出的在结算时一起展示
				if t.DeskData.Rtype == int32(pb.ROOM_TYPE1) {
					if _, ok := t.DeskPriv.PrivPlayer[role.Userid]; !ok {
						vipLv := strconv.Itoa(role.Vip.Lv)
						t.DeskPriv.PrivPlayer[role.Userid] = [3]string{role.Nickname, role.Photo, vipLv}
					}
				}
			}
		}

		if t.Rtype == int32(pb.ROOM_TYPE1) {
			t.cardTypeId = 15 // 私人房使用牌型15
		} else {
			t.cardTypeId = t.findCardTypeId()
		}

	case int32(pb.ROOM_TYPE2): //百人
	}
}

// 结束重置
func (t *Desk) gameOverInit() {
	t.ActSeat = 0

	for _, v := range t.seats {
		t.clearDrop(v.Userid)
		v.Ready = false
		v.Ready2 = false
		// v.Watch = false
		v.Card = 0
		v.Cards = []uint32{}
		v.SortCards = [][]uint32{}
		v.Finish = 0
		v.Sort = false
		v.Winner = false
		v.Declare = false
		v.Core = false
		v.Drop = false
		v.ActTimes = 0
		v.DrawedNeed = 0
		v.Actions = []int{}
		v.Touchs = []int{}
	}

	for _, v := range t.DeskAct.ActSeats {
		v.See = false
		v.Alive = false
		v.Lose = false
		v.Pack = false
		v.Drop = false
		v.ZhaHu = false
		v.Bet = 0
		v.ActNum = 0
	}
	//重新加载配置
	t.Game = config.GetGame(t.Game.Id)
	t.Ante = uint32(t.Game.RM.Bottom)

	t.DeskGame.Cards = make([]uint32, 0)

	t.DeskGame.WildCard = 0
	t.DeskGame.QiCards = []uint32{}

	t.DeskGame.BetNum = 0
	t.robotNum = t.randRobotNum()
	t.robotTime = 0
	t.robotCalling = 0
	t.addCard = false
	t.FakeSeats = handler.CreateFakeSeats(t.robotNum, t.seats, t.DeskData.Count) //重新生成假位置
	t.broadFakeSeat()
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

// ' 消息推送
func (t *Desk) pushMsg() {
	//选择庄家
	// t.dealer4()
	glog.Debugf("dealer -> %s", t.DeskGame.Dealer)
	glog.Debugf("dealer seat -> %d", t.DeskGame.DealerSeat)
	// t.timer = 0

	t.pushDealer() //广播消息

	t.state = int32(pb.STATE_DEALING) //切换为发牌状态
	t.pushState()                     //广播状态

}

// 状态消息
func (t *Desk) pushState() {
	msg := &pb.RMPushStateNtf{
		State: t.state,
	}
	t.broadcast(msg)
}

// 庄家消息
func (t *Desk) pushDealer() {
	msg := &pb.RMPushDealerNtf{
		DealerSeat: t.DeskGame.DealerSeat, //庄家
		WildCard:   t.DeskGame.WildCard,   //万能牌
		QiCard:     t.DeskGame.QiCards[0], //弃牌堆第一张牌
		CoreRobot:  t.coreRobotSeat,       //核心人机位置
	}

	//每个玩家定庄牌
	for k, v := range t.seats {
		if v.Ready {
			msg.Seats = append(msg.Seats, &pb.RMPushDealerInfo{
				Seat: k,      //座位号
				Card: v.Card, //定庄牌
			})
		}
	}
	t.broadcast(msg)

	//推送玩家手牌
	for k, v := range t.seats {
		if v.Ready {
			msg := &pb.RMPushCardsNtf{
				Seat:  k,
				Cards: v.Cards,
			}
			t.send2seat(k, msg)
		}
	}
}

// 广播操作状态消息
func (t *Desk) pushActState() {
	msg := &pb.RMPushActStateNtf{
		State: t.DeskAct.ActState,
		Seat:  t.DeskAct.ActSeat,
		Timer: BetTime,
		Ante:  t.DeskAct.ActAnte,
		Turn:  t.DeskAct.ActTimes,
		// Pot:      t.DeskGame.BetNum,
		// CallNum:  t.DeskHua.ActCallNum,
		// RaiseNum: t.DeskHua.ActRaiseNum,
	}

	// 摸牌前判断人机是否弃牌
	if t.isRobotDrop(msg.Seat) {
		msg.Rdrop = true
	}

	t.broadcast(msg)
}

//.

// ' 选择庄家
func (t *Desk) dealer() {
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

	//随机定庄牌
	var d []uint32
	d = append(d, algo.RMCARDS...)
	for i := len(d) - 1; i > 0; i-- {
		j := utils.RandIntN(i + 1)
		d[i], d[j] = d[j], d[i]
	}

	//发定庄牌
	for _, v := range t.seats {
		if v.Ready {
			v.Card = d[0]
			d = d[1:]
		}
	}

	//配牌开始
	// cfg.Reload()

	// conf := cfg.Section(nodeName).Key("dealercard").Value()
	// card1 := strings.Split(conf, ",")
	// var uint32Slice []uint32
	// for _, c := range card1 {
	// 	uintValue, err := strconv.ParseUint(c, 16, 32)
	// 	if err != nil {
	// 		continue
	// 	}
	// 	uint32Value := uint32(uintValue)
	// 	uint32Slice = append(uint32Slice, uint32Value)
	// }

	// //发定庄牌
	// for i, v := range t.seats {
	// 	idx := int(i - 1)
	// 	if idx < len(uint32Slice) {
	// 		v.Card = uint32Slice[idx]
	// 	}
	// }
	//配牌结束

	//比较定庄牌确定庄家
	var dealer uint32 = 0
	var card uint32 = 0
	//比较定庄牌
	for k, v := range t.seats {
		if !v.Ready {
			continue
		}
		if dealer == 0 {
			dealer = k
			card = v.Card
			continue
		}
		if algo.RMCompare(v.Card, card) {
			dealer = k
			card = v.Card
		}
	}

	// seat := a[0]
	// seat := a[rand.Intn(len(a))]

	if val, ok := t.seats[dealer]; ok {
		t.DeskGame.Dealer = val.Userid
		t.DeskGame.DealerSeat = dealer
	}
}

func (t *Desk) seatDrop(userid string) {
	seatid := t.getSeatid(userid)
	seat := t.getSeat(seatid)
	if seat == nil {
		return
	}
	seat.Drop = true

	role := t.getRole(userid)
	if role == nil {
		return
	}
	role.DropCount++
}

func (t *Desk) clearDrop(userid string) {
	seatid := t.getSeatid(userid)
	seat := t.getSeat(seatid)
	if seat == nil {
		return
	}

	if !seat.Drop {
		role := t.getRole(userid)
		if role == nil {
			return
		}
		role.DropCount = 0
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
		return true, t.Game.RM.NewbieMode.FoamCardType
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
	for k, v := range t.Game.RM.NewbieMode.SpecialRound {
		if int(round) <= v {
			idx = k
			ok = true
			break
		}
	}

	if ok {
		return ok, t.Game.RM.NewbieMode.SpecialRoundCardType[idx]
	} else {
		return false, 0
	}
}

// 查找牌型ID
func (t *Desk) findCardTypeId() int {
	var factor_int int

	for _, role := range t.roles {
		if role.IsA() {
			// A类玩家指定牌型
			return t.Game.RM.ACardType
		}
	}

	if t.DeskType == int32(pb.DESK_TYPE_NORMAL) {
		ok, cardType := t.CheckFoamStateCardType()
		if ok {
			return cardType
		}

		factor := t.getFinalFactor()
		factor_int = int(factor * 100)
	} else if t.DeskType == int32(pb.DESK_TYPE_POINTCONTROL) {
		factor_int = t.getPointControlFactor()
		if factor_int > 200 {
			factor_int = 200
		} else if factor_int < 0 {
			factor_int = 0
		}
	} else if t.DeskType == int32(pb.DESK_TYPE_NEWBIEW) {
		return t.getNewbieModeCardType()
	}

	index := 0
	for i, v := range t.Game.RM.FinalFactorRange {
		if factor_int < v {
			index = i
			break
		} else {
			index = i + 1
		}
	}
	return t.Game.RM.CardTypeRange[index]
}

// 生成牌
func (t *Desk) getCard(robot bool, current_cards []uint32) (cards []uint32, remain_cards []uint32) {
	id := t.cardTypeId
	// id := 11
	glog.Debugf("===============当前牌型id: %d, 房间ID: %d=================", id, t.Game.Id)

	ct := t.Game.RM.FindCardType(id)

	var choices []utils.Choice
	if !robot {
		for k, v := range ct.PlayerWeight {
			choices = append(choices, utils.Choice{Weight: v, Item: k})
		}
	} else {
		for k, v := range ct.RobotWeigtht {
			choices = append(choices, utils.Choice{Weight: v, Item: k})
		}
	}
	c, _ := utils.WeightedChoice(choices)
	index := c.Item.(int)

	var id2 int
	if !robot {
		id2 = ct.PlayerWeight2CardTypeId[index]
	} else {
		id2 = ct.RobotWeight2CardTypeId[index]
	}

	ct2 := t.Game.RM.FindCardType2(id2)

	return algo.RMGetCard(ct2.CardTypeConfig, current_cards, t.WildCard)
}

// 插入joker
func (t *Desk) joker() {
	t.DeskGame.Cards = append(t.DeskGame.Cards, algo.RMJOKERS...)
	t.DeskGame.Cards = append(t.DeskGame.Cards, algo.RMJOKERS...)
	for i := len(t.DeskGame.Cards) - 1; i > 0; i-- {
		j := utils.RandIntN(i + 1)
		t.DeskGame.Cards[i], t.DeskGame.Cards[j] = t.DeskGame.Cards[j], t.DeskGame.Cards[i]
	}
}

// '发牌
func (t *Desk) deal() {
	var hand = 13
	for _, v := range t.seats {
		if !v.Ready {
			continue
		}

		v.Cards = make([]uint32, hand, hand)
		if r, ok := t.roles[v.Userid]; ok {
			// _ = r
			// v.Cards = append(v.Cards, t.DeskGame.Cards[:hand]...)
			// t.DeskGame.Cards = t.DeskGame.Cards[hand:]

			cards, remain := t.getCard(r.Robot, t.DeskGame.Cards)

			// if len(cards) != 13 {
			// 	fmt.Println("1")
			// }

			// for _, v := range cards {
			// 	if v == 0 {
			// 		fmt.Println("1")
			// 	}
			// }

			copy(v.Cards, cards)
			t.DeskGame.Cards = remain
		}

		//准备的人才是参与者
		// copy(v.Cards, t.DeskGame.Cards[:hand])
		// t.DeskGame.Cards = t.DeskGame.Cards[hand:]
		// glog.Debugf("deal -> %v", v.Cards)
		//发牌消息
		//看不到牌值
		// cards2 := make([]uint32, hand, hand)
		// msg := resDraw(k, t.state, cards2)
		// t.broadcast(msg)
	}
	//配牌开始
	// cfg.Reload()

	// //设置万能牌
	// wild := cfg.Section(nodeName).Key("wildcard").Value()
	// if wild != "" {
	// 	uintValue, _ := strconv.ParseUint(wild, 16, 32)
	// 	t.WildCard = uint32(uintValue)
	// }

	// //发牌
	// for i, v := range t.seats {
	// 	if !v.Ready {
	// 		continue
	// 	}

	// 	name := fmt.Sprintf("card%d", i)
	// 	conf := cfg.Section(nodeName).Key(name).Value()
	// 	card := strings.Split(conf, ",")
	// 	if len(card) != 13 {
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
	//配牌结束
}

//.

// '自由场结算消息
func (t *Desk) resCoinOver(score map[uint32]int64) (msg *pb.RMCoinGameoverNtf) {
	t.state = int32(pb.STATE_OVER)
	msg = &pb.RMCoinGameoverNtf{
		// Dealer: t.DeskGame.Dealer,
		// State:  t.state,
	}
	for _, v := range t.over {
		msg.Infos = append(msg.Infos, v)
	}

	// for k, v := range score {
	// info := &pb.RMCoinGameOverInfo{
	// 	Seat: k,
	// }
	// if v1, ok := t.seats[k]; ok {
	// 	if len(v1.SortCards) == 0 {
	// 		ret_cards := v1.Cards
	// 		algo.Sort(ret_cards)
	// 		info.Cards = append(info.Cards, &pb.RMFinishInfo{Cards: ret_cards})
	// 	} else {
	// 		for _, v2 := range v1.SortCards {
	// 			info.Cards = append(info.Cards, &pb.RMFinishInfo{Cards: v2})
	// 		}
	// 	}

	// 	role := t.getRole(v1.Userid)
	// 	if role != nil {
	// 		info.Nickname = role.Nickname
	// 		info.Photo = role.Photo
	// 	}
	// }

	// status := t.getStatus(k)
	// info.Point = status.ActScore
	// info.WinOrLose = v
	// msg.Infos = append(msg.Infos, info)
	// d := &pb.RMCoinOver{
	// 	Seat:  k,
	// 	Score: v,
	// }
	// if val, ok := t.seats[k]; ok {
	// 	d.Bets = val.Bet
	// 	d.Value = val.Power
	// 	d.Cards = val.Cards
	// 	if p, ok2 := t.roles[val.Userid]; ok2 {
	// 		d.Coin = p.User.GetCoin()
	// 		d.Nickname = p.User.GetNickname()
	// 		d.Photo = p.User.GetPhoto()
	// 	}
	// }
	// msg.Data = append(msg.Data, d)
	// }
	return
}

//.

// '私人局结算消息
func (t *Desk) resOver(score map[uint32]int64) (msg *pb.RMGameoverNtf) {
	msg = &pb.RMGameoverNtf{
		Dealer:     t.DeskGame.Dealer,
		DealerSeat: t.DeskGame.DealerSeat,
		Round:      t.DeskData.Round,
		CurRound:   t.DeskGame.Round,
	}
	if t.DeskData.Round > t.DeskGame.Round {
		msg.LeftRound = (t.DeskData.Round - t.DeskGame.Round)
	}
	for k, v := range t.over {
		d := &pb.RMRoomOver{
			Seat:  k,
			Score: v.WinOrLose,
			Total: v.Point,
		}
		for _, hu := range v.Cards {
			d.HuCards = append(d.HuCards, &pb.RMFinishCard{Cards: hu.Cards})
		}
		if val, ok := t.seats[k]; ok {
			d.Bets = val.Bet
			d.Value = val.Power
			d.Cards = val.Cards
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
	t.pushActState()
}

// 下个操作位置
func (t *Desk) getNextActSeat() uint32 {
	seat := t.DeskAct.ActSeat
	if seat == 0 {
		seat = t.DeskGame.DealerSeat
		return seat
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
	//消息构造
	msg := new(pb.RMCoinAllBiNtf)
	for k, v := range t.DeskAct.ActSeats {
		if v.Alive {
			cards := t.getHandCards(k)
			typ := algo.HuaType(cards)

			info := &pb.RMCoinAllBiInfo{
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
		if algo.HuaCompare(cs1, cs2) { //k赢
			if val, ok := t.DeskAct.ActSeats[winner]; ok {
				val.Alive = false
			}
			winner = k
		} else { //winner赢
			v.Alive = false
		}
	}
	msg.Winner = winner
	t.broadcast(msg)
	//等待播放比牌动画
	// timerChan1 := time.After(2 * time.Second)
	// <-timerChan1
	t.pauseGame(PAUSE_REASON_3, 2, winner)
}

func (t *Desk) AddCard() []uint32 {
	var d []uint32

	d = append(d, algo.RMCARDS...)
	d = append(d, algo.RMCARDS...)
	d = append(d, algo.RMJOKERS...)
	d = append(d, algo.RMJOKERS...)

	for i := len(d) - 1; i > 0; i-- {
		j := utils.RandIntN(i + 1)
		d[i], d[j] = d[j], d[i]
	}

	return d
}

// 必输局上家出牌限制
func (t *Desk) mustLoseRobotDiscardLimit(card uint32) (newCard uint32) {
	newCard = card //默认不换
	if !t.isTriggerMustLoseControl() {
		return
	}

	curr := t.DeskAct.ActSeat
	next := t.getNextActSeat()

	//当前操作的人为机器人，下个操作的人为玩家
	if t.isRobot(curr) && !t.isRobot(next) {
		seat := t.getSeat(next)
		if seat == nil {
			return
		}

		need := algo.IsNeedCard(seat.Cards, card, t.WildCard)
		if need {
			newCard, t.DeskGame.Cards = algo.DealRandomExceptBianZhangCardAndJoker(t.DeskGame.Cards, seat.Cards, t.WildCard) //换个新的牌放到弃牌堆
			t.DeskGame.Cards = append(t.DeskGame.Cards, card)                                                                //原本出的牌放到牌堆里
			if !t.isTriggerMustLoseLastCard {
				t.isTriggerMustLoseLastCard = true
			}
		}
	}
	return
}

// 切换下个操作位置
func (t *Desk) setNextActSeat() {
	//游戏仅剩1人结束
	if t.remainPlayerNum() == 1 {
		winner := t.getRemainWinner()
		t.gameOver(winner)
		return
	}

	//牌不够了加两幅牌
	if len(t.DeskGame.Cards) == 0 {
		if !t.addCard {
			t.addCard = true
			t.DeskGame.Cards = t.AddCard()
		} else {
			var winner uint32 = 0
			t.gameOver(winner)
			return
		}
	}

	last := t.DeskAct.ActSeat
	curr := t.getNextActSeat()
	//总下注额达到底池上限结束
	// if t.DeskGame.BetNum >= int64(t.Game.RM.Pool_Limit) {
	//等待播放下注动画(最后一个下注的人)
	// timerChan := time.After(2 * time.Second)
	// <-timerChan
	//消息构造
	// t.pauseGame(PAUSE_REASON_2, 2, nil)
	// return
	// }

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
	//最大20局
	// if t.DeskAct.ActTimes == 20 {
	//TODO 结束限制
	// t.pauseGame(PAUSE_REASON_2, 1, nil)
	// return
	// }
	t.timer = 0
	t.DeskAct.ActSeat = curr
	//设置下家操作值
	t.setNextActState()
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

// 是否第一回合
func (t *Desk) isFirstRound(seat uint32) bool {
	if t.ActTimes > 1 {
		return false
	}
	if t.ActTimes == 0 {
		return true
	}
	// 非庄家第一回合
	if t.ActTimes == 1 && seat != t.DeskGame.DealerSeat {
		return true
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

// 最后一名玩家胜利者
func (t *Desk) getRemainWinner() uint32 {
	for k, v := range t.DeskAct.ActSeats {
		if v.Alive {
			return k
		}
	}
	return 0
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
	// for k, v := range t.DeskAct.ActSeats {
	// 	if v.Alive {
	// 		winner = k
	// 	}
	// }
	// return
	for k, v := range t.seats {
		if v.Winner {
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

// 设置下个位置操作值
func (t *Desk) setNextActState() {
	var val int32

	//弃牌标记判断
	// val |= int32(pb.ACT_PACK) //弃牌操作

	//blind chaal标记判断
	// if v, ok := t.DeskAct.ActSeats[t.DeskAct.ActSeat]; ok {
	// 	if !v.See {
	// 		val |= int32(pb.ACT_BLIND)
	// 	} else {
	// 		val |= int32(pb.ACT_CHAAL)
	// 	}
	// }

	//sideshow标记判断
	// prevSeat := t.getPrevActSeat()
	// seat := t.DeskAct.ActSeat
	// if v, ok := t.DeskAct.ActSeats[prevSeat]; ok {
	// 	if v1, ok := t.DeskAct.ActSeats[seat]; ok {
	// 		if v.See && v1.See {
	// 			val |= int32(pb.ACT_SIDESHOW)
	// 		}
	// 	}
	// }

	// if t.remainPlayerNum() == 2 {
	// 	val ^= int32(pb.ACT_SIDESHOW)
	// 	val |= int32(pb.ACT_SHOW)
	// }

	t.DeskAct.ActState = val
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

func (t *Desk) alive() {
	// msg := new(pb.RMCoinRaiseAnteNtf)
	for k, v := range t.seats {
		if !v.Ready {
			continue
		}
		d := new(data.ActStatus)
		d.Alive = true
		t.DeskAct.ActSeats[k] = d
		//开始时下暗注
		// info := t.setBet(k, v.Userid, int64(t.DeskData.Ante), fmt.Sprintf("tp%s房间下底注", t.DeskData.Rid))
		// msg.Info = append(msg.Info, info)
	}
	// msg.Pot = t.DeskGame.BetNum
	// t.broadcast(msg)
}

// 开始游戏初始化操作
func (t *Desk) initAct() {

	//初始化玩家操作
	//TODO 加注和跟注下限
	// t.DeskHua.ActCallNum = int64(t.DeskData.Ante)      //跟住
	// t.DeskHua.ActRaiseNum = int64(t.DeskData.Ante) * 2 //加注

	t.timer = 0
	t.DeskAct.ActAnte = int64(t.Game.RM.Bottom)
	t.DeskAct.ActSeat = t.getNextActSeat()
	//设置下家操作值
	t.setNextActState()
	//广播下家操作状态消息
	t.pushActState()
}

// 下注成功设置
func (t *Desk) setBet(seat uint32, userid string, num int64, desc string) *pb.RMCoinRaiseInfo {
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
func (t *Desk) setBetMsg(seat uint32, userid string, num int64, desc string) *pb.RMCoinRaiseInfo {
	msg := &pb.RMCoinRaiseInfo{
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
		t.sendCurrency(userid, (-1 * num), int32(pb.LOG_TYPE127), desc)
	case int32(pb.ROOM_TYPE1): //私人
		switch t.Gmode {
		case 0: // 真金
			t.sendCurrency(userid, (-1 * num), int32(pb.LOG_TYPE127), desc)
		case 1: // 娱乐模式
			t.sendFraction(userid, (-1 * num), int32(pb.LOG_TYPE127))
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

// vim: set foldmethod=marker foldmarker=//',//.:
