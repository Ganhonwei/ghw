package joker

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/algo"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"math"
	"time"
)

// '进入房间响应消息
func (t *Desk) privEnterMsg(userid string) *pb.JOKEREnterRoomRsp {
	msg := new(pb.JOKEREnterRoomRsp)
	//房间数据
	msg.Roominfo = handler.PackJOKERCoinRoom(t.DeskData)
	msg.Roominfo.State = t.state
	// 房间暂停状态原因
	if t.state == int32(pb.STATE_PAUSE) {
		msg.Roominfo.StatePauseReaon = int32(t.reason)
	}
	//TODO 添加操作信息
	//坐下玩家信息
	msg.Userinfo = t.coinSeatBetsMsg(userid)
	//位置下注信息
	msg.Betsinfo = t.coinBetsMsg()
	//投票信息
	msg.Voteinfo = t.voteInfoMsg()
	return msg
}

// 投票信息
func (t *Desk) voteInfoMsg() (msg *pb.JOKERRoomVote) {
	msg = new(pb.JOKERRoomVote)
	if t.DeskPriv != nil {
		msg.Seat = t.DeskPriv.VoteSeat
	}
	if msg.Seat == 0 {
		return
	}
	for k, v := range t.seats {
		if v.Vote == 1 {
			msg.Agree = append(msg.Agree, k)
		} else if v.Vote == 2 {
			msg.Disagree = append(msg.Disagree, k)
		}
	}
	return
}

//.

//'投票

func (t *Desk) checkVote() pb.ErrCode {
	if t.isFree() {
		return pb.OperateError
	}
	//if t.state > int32(pb.STATE_READY) {
	//	return pb.RunningNotVote
	//}
	if t.DeskPriv == nil {
		return pb.OperateError
	}
	return pb.OK
}

// 发起投票
func (t *Desk) launchVote(userid string, vote uint32) (msg *pb.JOKERLaunchVoteRsp) {
	msg = new(pb.JOKERLaunchVoteRsp)
	errcode := t.checkVote()
	if errcode != pb.OK {
		msg.Error = errcode
		return
	}
	if t.DeskPriv.VoteSeat != 0 {
		msg.Error = pb.VotingCantLaunchVote
		return
	}
	seat := t.getSeatid(userid)
	if v, ok := t.seats[seat]; ok {
		v.Vote = vote //投票
	}
	//发起投票者
	t.DeskPriv.VoteSeat = seat
	//超时设置(1分钟)
	glog.Debugf("VoteTime: %d, %d, %s", seat, vote, userid)
	t.DeskPriv.VoteTime = utils.Timestamp() + 60
	msg.Seat = seat
	t.broadcast(msg)
	t.pushVote(seat, vote)
	t.dismiss(false)
	return
}

func (t *Desk) voteTimeout() {
	errcode := t.checkVote()
	if errcode != pb.OK {
		return
	}
	if t.DeskPriv.VoteSeat == 0 {
		return
	}
	var now = utils.Timestamp()
	if now >= t.DeskPriv.VoteTime {
		t.dismiss(true)
	}
}

// 投票
func (t *Desk) privVote(userid string, vote uint32) (msg *pb.JOKERVoteRsp) {
	msg = new(pb.JOKERVoteRsp)
	errcode := t.checkVote()
	if errcode != pb.OK {
		msg.Error = errcode
		return
	}
	if t.DeskPriv.VoteSeat == 0 {
		msg.Error = pb.NotVoteTime
		return
	}
	seat := t.getSeatid(userid)
	if v, ok := t.seats[seat]; ok {
		v.Vote = vote //投票
	}
	t.pushVote(seat, vote)
	t.dismiss(false)
	return
}

// 广播投票消息
func (t *Desk) pushVote(seat, vote uint32) {
	msg := &pb.JOKERVoteRsp{
		Seat: seat,
		Vote: vote,
	}
	t.broadcast(msg)
}

// 广播投票消息
func (t *Desk) pushVoteResult(vote uint32) {
	msg := &pb.JOKERVoteResultNtf{
		Vote: vote,
	}
	t.broadcast(msg)
}

// 投票解散,agree >= unagree
func (t *Desk) dismiss(force bool) {
	var agree = 0
	var unagree = 0
	var voted = 0
	for _, v := range t.seats {
		if v.Vote == 1 {
			agree++
		} else {
			unagree++
		}
		if v.Vote != 0 {
			voted++
		}
	}
	//一半以上通过即可
	if agree >= unagree {
		//0解散,1不解散
		t.pushVoteResult(0)
		//停止服务
		msg1 := new(pb.ServeStop)
		t.selfPid.Tell(msg1)
	} else if force || voted == len(t.seats) {
		//结束投票
		t.pushVoteResult(1)
		//重置
		for _, v := range t.seats {
			v.Vote = 0
		}
		//发起投票者
		t.DeskPriv.VoteSeat = 0
		t.DeskPriv.VoteTime = 0
	}
}

const (
	PAUSE_REASON_0 = iota
	PAUSE_REASON_1 //普通比牌动画
	PAUSE_REASON_2 //最后一人下注动画（达到底池上限）、达到20局
	PAUSE_REASON_3 //全部比牌动画
	PAUSE_REASON_4 //结算动画
)

// 暂停游戏
func (t *Desk) pauseGame(reason int, time int, arg interface{}) {
	t.state = int32(pb.STATE_PAUSE)
	t.timer = 0
	t.pauseTime = time
	t.reason = reason
	t.pauseArg = arg
}

// 恢复游戏
func (t *Desk) resumeGame(reason int) {
	switch reason {
	case PAUSE_REASON_1:
		t.setNextActSeat()
	case PAUSE_REASON_2:
		t.allBi()
	case PAUSE_REASON_3:
		winner := t.pauseArg.(uint32)
		t.pauseArg = nil
		t.gameOver(winner)
	case PAUSE_REASON_4:
		t.state = int32(pb.STATE_READY) //设置房间状态
		t.pushState()
		//踢出玩家
		t.limitOver()
		// 真实玩家换桌
		t.mutexChangeDesk()
	}

}

// ' 超时处理
func (t *Desk) coinTimeout() {
	// t.checkPubOver2()
	// t.checkPubOver()
	switch t.state {
	case int32(pb.STATE_READY):
		t.kickFakeSeat()
		var num int = t.roleNum()
		if num < 2 { //大于等于2人时才计时
			return
		}
		t.checkRealUser() // 检测真人玩家够不够，够了人机就退出去, 让真人在里面玩
		if t.timer == ReadyTime {
			//准备超时,不等待全部准备
			//t.readyTimeout()
			t.timer = 0
			t.gameStart() //开始牌局
		} else {
			t.timer++
		}
		return
	case int32(pb.STATE_DEALING):
		// var num int = t.readyNum()   //游戏人数
		// var num2 int = t.ready2Num() //播放完发牌动画人数
		// if num2 >= num {
		// 	return
		// }
		if t.timer == DealingTime {
			t.timer = 0
			t.state = int32(pb.STATE_BET) //切换状态为下注状态
			t.pushState()
			t.initAct()
		} else {
			t.timer++
		}
	case int32(pb.STATE_BET):
		if t.timer == BetTime {
			t.timer = 0
			t.betTimeout()
		} else if t.timer == WaitTooLongTime {
			t.WaitTooLong(t.DeskAct.ActSeat)
			t.timer++
		} else if t.timer >= ChargeInGameTime {
			t.timer = 0
			t.betTimeout()
		} else {
			t.timer++
		}
	case int32(pb.STATE_PAUSE):
		if t.timer == t.pauseTime {
			reason := t.reason
			t.timer = 0
			t.reason = 0
			t.pauseTime = 0
			t.state = int32(pb.STATE_BET)
			t.resumeGame(reason)
		} else {
			t.timer++
		}
	case int32(pb.STATE_CHARGE):
		if t.chargingSeat > 0 && t.isRobot(t.chargingSeat) {
			// 机器人假装充值
			if t.timer >= t.robotChargingTime {
				t.timer = 0
				userid := t.getUserid(t.chargingSeat)
				glog.Infof("robot chargeInGame finish: %s", userid)
				t.chargeInGameFinish(userid)
			} else {
				t.timer++
			}

		} else {
			// 玩家正常局内充值
			if t.timer == ChargeInGameTime {
				t.timer = 0
				t.chargeTimeout()
			} else {
				t.timer++
			}
		}
	}

	// if t.timer == BetTime {
	// 	switch t.state {
	// 	case int32(pb.STATE_DEALER):
	// 		//抢庄超时,打庄
	// 		t.dealerHandler()
	// 	case int32(pb.STATE_BET):
	// 		//下注超时
	// 		t.betTimeout()
	// 	default:
	// 		t.timer = 0
	// 	}
	// } else {
	// 	t.timer++
	// }
}

// .等待太久广播
func (t *Desk) WaitTooLong(seat uint32) {
	msg := new(pb.JOKERCoinWaitTooLongNtf)
	msg.Seat = seat
	t.broadcast(msg)
}

// ' 超时处理
func (t *Desk) privTimeout() {
	t.checkPubOver2()
	t.voteTimeout()
	switch t.state {
	case int32(pb.STATE_READY):
		//过期关闭
		if t.checkExpire() {
			//关闭房间
			t.gameStop()
		}
		var num int = t.readyNum()
		if num < 2 { //大于2人时才计时
			return
		}
		if t.timer == ReadyTime {
			//准备超时,不等待全部准备
			//t.readyTimeout()
			t.gameStart() //开始牌局
			t.timer = 0
		} else {
			t.timer++
		}
		return
	}
	// if t.timer == BetTime {
	// 	switch t.state {
	// 	case int32(pb.STATE_DEALER):
	// 		//抢庄超时,打庄
	// 		t.dealerHandler()
	// 	case int32(pb.STATE_BET):
	// 		//下注超时
	// 		t.betTimeout()
	// 	default:
	// 		t.timer = 0
	// 	}
	// } else {
	// 	t.timer++
	// }
}

// 获取库存ID
func (t *Desk) getStockGameId() (string, bool) {
	if t.DeskType == int32(pb.DESK_TYPE_NEWBIEW) { //新手模式走新手库存
		return "6008", false
	} else {
		return t.Game.Id, true
	}
}

func (t *Desk) getGiveStockGameId() string {
	return "6009"
}

// 更新赠送金库存
func (t *Desk) _changeGiveStock(cash_give int64) {
	if cash_give == 0 {
		return
	}
	msg := &pb.ChangeStock{}
	msg.GameId = t.getGiveStockGameId()
	msg.Gtype = int32(pb.JOKER)
	msg.CashStock = cash_give

	res := t.reqRoom(msg)
	if _, ok := res.(*pb.ChangedStock); !ok {
		glog.Errorf("change stock failed: %#v", res)
	}
}

// 更新库存
func (t *Desk) _changeStock(cash_stock, bonus_stock, cash_ming, bonus_ming, cash_an, bonus_an int64) {
	//更新房间库存、税收
	msg := &pb.ChangeStock{}
	msg.GameId, msg.Real = t.getStockGameId()
	msg.Gtype = int32(pb.JOKER)
	msg.CashStock = cash_stock
	msg.BonusStock = bonus_stock
	msg.CashMingTax = cash_ming
	msg.BonusMingTax = bonus_ming
	msg.CashAnTax = cash_an
	msg.BonusAnTax = bonus_an

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

		role := t.roles[userid]
		if role.RoundGames == nil {
			role.RoundGames = make(map[int32]int32)
		}

		seat := t.getSeat(t.getSeatid(userid))
		if seat.Watch {
			return
		}

		m := &pb.NewbieStock{
			Gtype:     int32(pb.JOKER),
			GameTime:  time.Now().Unix() - t.detail.BeginTime,
			CashStock: cash_stock,
			CashMing:  cash_ming,
			CashAn:    cash_an,
			First:     role.RoundGames[int32(pb.JOKER)] <= 1,
		}
		t.roomPid.Tell(m)
	}
}

// 处理点控
func (t *Desk) handlePointControl(role *data.DeskRole, jokerdetail *data.JOKERDetail) {
	if t.DeskType != int32(pb.DESK_TYPE_POINTCONTROL) {
		return
	}
	if role.Robot {
		return
	}
	if !role.PCSwitch || role.PCScore == 0 {
		return
	}

	diff := jokerdetail.AfterCash - jokerdetail.BeforeCash
	if diff == 0 {
		return
	}
	role.PCScoreComplete += diff

	if role.PCScore > 0 { //控赢
		if role.PCScoreComplete >= role.PCScore {
			role.PCSwitch = false
			role.PCFactor = 0
			role.PCScoreComplete = 0
			role.PCScore = 0
			t.DeskType = int32(pb.DESK_TYPE_NORMAL)
		}
	} else if role.PCScore < 0 { //控输
		if role.PCScoreComplete <= role.PCScore {
			role.PCSwitch = false
			role.PCFactor = 0
			role.PCScoreComplete = 0
			role.PCScore = 0
			t.DeskType = int32(pb.DESK_TYPE_NORMAL)
		}
	}
	t.sendPointControl(role.Userid)
}

// 检查赠送彩金
func (t *Desk) CheckGiveDiamond(seateid uint32, score int64) (score_deduction int64) {
	if t.DeskType == int32(pb.DESK_TYPE_NEWBIEW) { //新手模式不检查赠送彩金
		return
	}
	userid := t.getUserid(seateid)
	role := t.getRole(userid)
	if role == nil {
		return
	}
	if role.Robot {
		return
	}
	if score >= 0 {
		return
	}

	//同步变化
	msg := &pb.GiveAndOutCash{
		Userid: userid,
		GameId: t.GameId,
		Desc:   fmt.Sprintf("房间%s", t.Rid),
	}
	if role.Offline {
		t.rolePid.Tell(msg)
	} else {
		t.send2userid(userid, msg)
	}

	return
}

// 检查可提现彩金
func (t *Desk) CheckOutDiamond(seatid uint32, score_final int64) {
	userid := t.getUserid(seatid)
	role := t.getRole(userid)
	if role == nil {
		return
	}
	if role.Robot {
		return
	}

	var incre int64
	if score_final > 0 { //赢
		old := role.OutDiamond               //旧可提彩金
		new := role.OutDiamond + score_final //新可提彩金
		if new > role.Diamond {              //可提彩金大于携带彩金
			new = role.Diamond //可提彩金不能超过携带彩金
		}
		incre = new - old //可提彩金增量
	} else if score_final < 0 { //输
		old := role.OutDiamond
		new := role.OutDiamond
		if role.OutDiamond > role.Diamond { //可提彩金不能超过携带彩金
			new = role.Diamond
		}
		incre = new - old //可提彩金增量
	}

	role.OutDiamond += incre //修改本地可提彩金数据

	//同步变化
	msg := &pb.GiveAndOutCash{
		Userid: userid,
		GameId: t.GameId,
		Desc:   fmt.Sprintf("房间%s", t.Rid),
	}
	msg.Out = incre
	if role.Offline {
		t.rolePid.Tell(msg)
	} else {
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

// 输
func (t *Desk) Lose(seatid uint32) {
	userid := t.getUserid(seatid)
	status := t.getStatus(seatid)
	role := t.getRole(userid)
	seat := t.getSeat(seatid)
	if status == nil || role == nil || seat == nil {
		return
	}
	if status.Alive {
		return
	}
	score := -status.ActNum //输赢
	var score_final int64   //最终输赢
	//非人机计算库存
	if !role.Robot {
		score_deduction := t.CheckGiveDiamond(seatid, score)
		score_final, status.Stock = t.calcStockAndTax(score, role.Ratio, score_deduction)
		status.Stock.CashGive = -score_deduction
		status.Stock.Robot = role.Robot
	} else {
		score_final = score
	}
	t.score[seatid] = score_final

	t.shareAmount(userid, score_final)
	t.RecordDetail(seatid, status.ActNum, score_final, status.Stock)
	t.EventPost(seatid, false)
	t.GameTime(seatid)

	//结算数据
	over := &pb.JOKERCoinOver{
		Seat:  seatid,
		Score: score_final,
	}
	over.Bets = seat.Bet
	over.Value = seat.Power
	over.Cards = seat.Cards
	over.Coin = role.GetCoin()
	over.Nickname = role.GetNickname()
	over.Photo = role.GetPhoto()
	if status.Pack || status.Lose { //弃牌不亮、sideshow输的不亮
		over.Show = false
	} else {
		over.Show = true
	}
	t.over[seatid] = over

	t.CheckOutDiamond(seatid, score_final)
	t.jokerRecord(seatid, score_final)
	t.flowWater(userid, score_final)
}

// 赢
func (t *Desk) Win(seatid uint32) {
	userid := t.getUserid(seatid)
	status := t.getStatus(seatid)
	role := t.getRole(userid)
	seat := t.getSeat(seatid)
	if status == nil || role == nil || seat == nil {
		return
	}
	if !status.Alive {
		return
	}
	score := t.DeskGame.BetNum //赢总投注
	var score_final int64      //最终输赢
	//非人机计算库存
	// if !role.Robot {
	score -= status.ActNum //先扣除下注额
	score_final, status.Stock = t.calcStockAndTax(score, role.Ratio)
	score_final += status.ActNum //再返还下注额
	status.Stock.Robot = role.Robot
	// } else {
	// score_final = score
	// }
	t.score[seatid] = score_final
	t.sendCurrency(userid, score_final, int32(pb.LOG_TYPE113), fmt.Sprintf("joker%s房间赢分", t.DeskData.Rid)) //加钱

	t.shareAmount(userid, score_final)
	t.RecordDetail(seatid, status.ActNum, score_final, status.Stock)
	t.EventPost(seatid, true)
	t.GameTime(seatid)

	//结算数据
	over := &pb.JOKERCoinOver{
		Seat:  seatid,
		Score: score_final,
	}
	over.Bets = seat.Bet
	over.Value = seat.Power
	over.Cards = seat.Cards
	over.Coin = role.GetCoin()
	over.Nickname = role.GetNickname()
	over.Photo = role.GetPhoto()
	if status.Pack || status.Lose { //弃牌不亮、sideshow输的不亮
		over.Show = false
	} else {
		over.Show = true
	}
	t.over[seatid] = over

	// 新手不给可提现金
	if role.State != data.NoveiceState {
		t.CheckOutDiamond(seatid, score_final-status.ActNum)
	}
	t.jokerRecord(seatid, score_final)
	t.flowWater(userid, score_final)
}

// 记录详情
func (t *Desk) RecordDetail(seat uint32, act_num, score_final int64, stock *data.ActStock) {
	userid := t.getUserid(seat)
	role := t.getRole(userid)
	if role == nil {
		return
	}
	jokerdetail := t.detail.FindJOKERDetail(seat)
	if jokerdetail == nil {
		return
	}

	//牌型
	jokerdetail.Cards = t.getHandCards(seat)

	//总投注、底注
	jokerdetail.Bet = act_num                   //总投注
	jokerdetail.Bottom = int64(t.DeskData.Ante) //底注

	//库存
	if stock != nil {
		jokerdetail.CashStock = stock.CashStock
		jokerdetail.BonusStock = stock.BonusStock
		jokerdetail.CashMingTax = stock.CashMing
		jokerdetail.BonusMingTax = stock.BonusMing
		jokerdetail.CashAnTax = stock.CashAn
		jokerdetail.BonusAnTax = stock.BonusAn
	}

	//结算
	jokerdetail.Score = score_final

	//账变后分数、彩金、奖励金
	jokerdetail.AfterScore = role.GetScore()
	jokerdetail.AfterCash = role.GetDiamond()
	jokerdetail.AfterBonus = role.GetCoin()

	//处理点控
	t.handlePointControl(role, jokerdetail)
}

func (t *Desk) EventPost(seat uint32, win bool) {
	userid := t.getUserid(seat)
	role := t.getRole(userid)
	if userid == "" || role == nil {
		return
	}
	if role.Robot {
		return
	}
	e := &event.GameRecordEvent{
		Gtype: uint32(pb.JOKER),
		Win:   win,
	}
	t.eventPost(userid, event.PLAY_TASK, e)
	cards := t.getHandCards(seat)
	typ := algo.HuaType(cards)
	// hands := &event.PokerHandsEvent{
	// 	GameType: int(pb.JOKER),
	// 	Win:      win,
	// 	PX:       typ,
	// }
	// t.eventPost(userid, event.WIN_TASK, e)
	// t.eventPost(userid, event.POKER_HANDS, hands)

	// vip bank 游戏任务
	vb := &event.VBGameTaskEvent{
		GameType: int32(pb.JOKER),
		Win:      win,
		HuaType:  typ,
	}
	t.eventPost(userid, event.VB_GAME_TASK, vb)
}

func (t *Desk) GameTime(seat uint32) {
	userid := t.getUserid(seat)
	role := t.getRole(userid)
	if userid == "" || role == nil {
		return
	}
	if role.Robot {
		return
	}

	msg := &pb.LogGameTime{
		Userid: userid,
		Gtype:  int32(pb.JOKER),
		Time:   time.Now().Unix() - t.detail.BeginTime,
	}
	role.TotalGameTime += uint64(msg.Time)
	myactor.Logger().Tell(msg)
	t.send2userid(userid, msg)
}

func (t *Desk) changeStock() {
	var cash_stock, bonus_stock, cash_ming, bonus_ming, cash_an, bonus_an int64
	var cash_give int64
	for _, v := range t.ActSeats {
		if v.Stock == nil {
			continue
		}
		if v.Stock.Robot {
			continue
		}
		cash_stock += v.Stock.CashStock
		bonus_stock += v.Stock.BonusStock
		cash_ming += v.Stock.CashMing
		bonus_ming += v.Stock.BonusMing
		cash_an += v.Stock.CashAn
		bonus_an += v.Stock.BonusAn
		cash_give += v.Stock.CashGive
	}
	t._changeStock(cash_stock, bonus_stock, cash_ming, bonus_ming, cash_an, bonus_an)
	t._changeGiveStock(cash_give)
}

// '结束游戏
func (t *Desk) gameOver(winner uint32) {
	t.timer = 0
	t.state = int32(pb.STATE_OVER)
	t.pushState() //广播状态

	//详情记录
	t.detail.EndTime = time.Now().Unix()       //详情记录结束时间
	t.detail.ChangeCardType = t.changeCardType //换牌类型(开局换，局中换)
	t.detail.ControlType = t.controlType       //控制方式(当前赢分，玩家系数)
	t.detail.WinScore = t.winScore             //当前赢分(开局换)
	t.detail.PlayerFactor = t.playerFactor     //玩家系数(开局换)
	t.detail.WinScore1 = t.winScore1           //当局赢分(局中换)
	t.detail.WinScore2 = t.winScore2           //当前赢分(局中换)
	t.detail.ChargeMoney = t.chargeMoney       //充值金额(开局、局中)
	t.detail.IsStrategy = t.isStrategy         //是否是策略局
	t.detail.StrategyType = t.strategyType     //策略局类型
	t.detail.IsCharge = t.isCharge             //是否充值

	t.Win(winner)

	t.StrategyShowCard()
	//结算消息
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0): //自由
		//结算消息
		msg := t.resCoinOver(t.score)

		msg.Winner = winner
		msg.Wincoin = uint64(t.score[winner])

		t.broadcast(msg)
		//修改库存
		t.changeStock()
		//详情
		t.saveDetail()
		//冤家牌记录
		t.addHedgeRecord(false)
		//记录
		// t.saveRecord(score)
		//结束连庄处理
		// t.dealerOver()
		//重置状态
		t.gameOverInit()
		//踢出机器人
		t.KickRobot()
		//踢出不足坐下玩家或超额玩家
		// t.limitOver()
		//踢除离线玩家
		t.kickOffline()
		//清除离开数据
		t.clearLeave()
	case int32(pb.ROOM_TYPE1): //私人
		//牌局数累加一次
		// if !t.DeskData.Pub {
		// 	t.DeskGame.Round++
		// }
		//结算消息
		// msg := t.resOver(score)
		// t.broadcast(msg)
		//记录
		// if !t.DeskData.Pub {
		// 	t.saveRecord(score)
		// }
		//结束连庄处理
		// t.dealerOver()
		//重置状态
		// t.gameOverInit()
		//踢出不足坐下玩家或超额玩家
		// t.limitOver()
		//踢除离线玩家
		// t.kickOffline()
		//关闭房间
		//t.gameStop()
	case int32(pb.ROOM_TYPE2): //百人
	}
}

// 输赢记录
func (t *Desk) jokerRecord(seat uint32, num int64) {
	userid := t.getUserid(seat)
	role := t.getRole(userid)
	if userid == "" || role == nil {
		return
	}
	if role.Robot {
		return
	}
	role.Round++

	var bet int64 = 0
	if v, ok := t.seats[seat]; ok {
		bet = v.Bet
	}
	msg := &pb.FreeSetRecord{Gtype: int32(pb.JOKER), Score: num, Bet: bet}
	if num > 0 { //赢
		msg.Rtype = 1
	} else if num < 0 { //输
		msg.Rtype = -1
	} else { //平
		msg.Rtype = 0
	}
	msg.GameTime = utils.BsonNow().Unix() - t.detail.BeginTime
	t.send2userid(userid, msg)

	// 新手转平民
	bean := table.GetTables().NewbieTable.Get()
	if role.State == data.NoveiceState && role.Round >= uint32(bean.Transfer[1]) {
		role.State = data.ExceptionState
	}

	if role.RoundGames == nil {
		role.RoundGames = make(map[int32]int32)
	}
	role.RoundGames[int32(pb.JOKER)]++

	// 打码上报
	if err = mq.NatsPublish(mq.TopicGameBets, &pb.PublishGameBets{
		UserPid:    role.Pid,
		Userid:     userid,
		Gtype:      int32(pb.JOKER),
		Ts:         time.Now().Unix(),
		Bets:       bet,
		Robot:      role.Robot,
		Username:   role.Nickname,
		Photo:      role.Photo,
		RegistArea: int32(role.RegistArea),
		VipLv:      int32(role.Vip.Lv),
		SuperId:    role.ShareSuperior,
		WaterId:    t.GameId,
		Score:      utils.CaseElse(num <= 0, num, num-bet),
	}); err != nil {
		glog.Error("publish user joker bets error", err)
	}
	// 打码量日志
	log := handler.GameFlowWaterLog(bet, int(pb.JOKER), 0, len(role.RechargeTarge), userid)
	myactor.Logger().Tell(log)

	//判断底注
	if role.Diamond >= int64(t.Ante)*300 {
		return
	}

	if t.Game.JOKER.Strategy300Switch == 0 {
		return
	}

	if t.DeskType != int32(pb.DESK_TYPE_NORMAL) { //正常桌才累计局数
		return
	}

	// 增加累计局数
	role.IncreTPTotalRound()
	//同步数据
	msg1 := &pb.TPTotalRound{}
	t.send2userid(userid, msg1)

	role.BetRecord = append(role.BetRecord, bet)
	if len(role.BetRecord) > 200 {
		role.BetRecord = role.BetRecord[len(role.BetRecord)-200:]
	}
}

// .
func (t *Desk) checkMin(role *data.DeskRole) bool {
	return role.GetScore() >= int64(t.Game.Min_Access)
}

func (t *Desk) checkMax(role *data.DeskRole) bool {
	if t.Game.Max_Access == -1 {
		return true
	}
	return role.GetScore() <= int64(t.Game.Max_Access)
}

func (t *Desk) checkBlackList(role *data.DeskRole) bool {
	return role.Status != 3
}

func (t *Desk) checkPointControl(role *data.DeskRole) bool {
	if t.DeskType == int32(pb.DESK_TYPE_POINTCONTROL) { //跳过点控桌子
		return true
	}
	if role.Robot { //跳过机器人
		return true
	}
	if role.PCSwitch {
		return false
	} else {
		return true
	}
}

func (t *Desk) checkNewbie(role *data.DeskRole) bool {
	if t.DeskType != int32(pb.DESK_TYPE_NEWBIEW) { //跳过非新手桌子
		return true
	}
	if role.Robot { //跳过人机
		return true
	}
	if role.State == 2 { //新手桌子上的玩家状态变为非新手，需要踢出
		return false
	} else {
		return true
	}
}

func (t *Desk) check105(role *data.DeskRole) bool {
	// if role.Kick105Flag {
	// 	return true
	// }
	// if role.OutDiamond >= 105*100 && role.Money == 0 {
	// 	return false
	// } else {
	// 	return true
	// }
	return true
}

func (t *Desk) check200(role *data.DeskRole) bool {
	// if role.Kick200Flag {
	// 	return true
	// }
	// if role.OutDiamond >= 200*100 && role.Money == 0 {
	// 	return false
	// } else {
	// 	return true
	// }
	return true
}

func (t *Desk) checkGameTime(role *data.DeskRole) bool {
	if role.State == 3 {
		return true
	}
	bean := table.GetTables().NewbieTable.Get()
	if bean == nil {
		return true
	}

	if role.TotalGameTime >= uint64(bean.Transfer[0])*60 && role.Money == 0 {
		return false
	} else {
		return true
	}
}

func (t *Desk) checkTimeout(role *data.DeskRole) bool {
	if role.TimeoutCount >= 2 {
		return false
	} else {
		return true
	}
}

func (t *Desk) checkCloseServer() bool {
	if t.closeServer {
		return false
	} else {
		return true
	}

}

// 清除离开数据
func (t *Desk) clearLeave() {
	msg := new(pb.ClearLeave)
	msg.Roomid = t.Rid
	nodePid.Tell(msg)
}

// 如果桌子没有玩家，踢出机器人
func (t *Desk) KickRobot() {
	r, _ := t.roleCountNum()
	if r != 0 { //存在真人跳过
		return
	}

	for k, v := range t.roles {
		if !v.Robot { //跳过非机器人
			continue
		}
		errcode := pb.OK
		t.notifyGateUserLeft(v.Userid, errcode, 0)
		t.userLeaveDesk(k)
	}
}

// '踢出不足坐下玩家或超额玩家
func (t *Desk) limitOver() {
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0): //自由
	case int32(pb.ROOM_TYPE1): //私人
		if !t.DeskData.Pub {
			//return
		}
	case int32(pb.ROOM_TYPE2): //百人
		return
	}
	for k, v := range t.roles {
		// score := v.User.GetScore()
		//if t.DeskData.Maximum == 0 {
		//	if coin >= t.DeskData.Minimum {
		//		continue
		//	}
		//} else {
		//	if coin >= t.DeskData.Minimum &&
		//		coin < t.DeskData.Maximum {
		//		continue
		//	}
		//}
		if v.Robot {
			continue
		}

		if t.checkMin(v) &&
			t.checkMax(v) &&
			// t.checkBlackList(v) &&
			t.checkPointControl(v) &&
			t.checkNewbie(v) &&
			// t.check105(v) &&
			// t.check200(v) &&
			t.checkGameTime(v) &&
			t.checkTimeout(v) &&
			t.checkCloseServer() {
			continue
		}
		errcode := t.leave(k)
		if errcode != pb.OK {
			continue
		}

		var err pb.ErrCode
		if !t.checkMin(v) {
			err = pb.NotEnoughCoin
		} else if !t.checkMax(v) {
			err = pb.TooManyCoin
		} else if !t.checkPointControl(v) {
			err = pb.PointControlKick
		} else if !t.checkNewbie(v) {
			err = pb.NewbieKick
		} else if !t.checkGameTime(v) {
			err = pb.KickGameTime
		} else if !t.checkTimeout(v) {
			err = pb.KickTimeout
		} else if !t.checkCloseServer() {
			err = pb.KickCloseServer
		}

		t.notifyGateUserLeft(v.Userid, errcode, int32(err))
		//离开状态消息
		t.userLeaveDesk(k, err)
	}
}

// 踢除离线玩家
func (t *Desk) kickOffline() {
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0), //自由
		int32(pb.ROOM_TYPE1), //私人
		int32(pb.ROOM_TYPE2): //百人
		for k, v := range t.roles {
			if !v.Offline {
				continue
			}
			errcode := t.leave(k)
			if errcode != pb.OK {
				continue
			}
			//离开状态消息
			t.userLeaveDesk(k)
		}
	}
}

// pub房间人数为0时解散
func (t *Desk) checkPubOver() {
	// switch t.DeskData.Rtype {
	// case int32(pb.ROOM_TYPE1): //私人
	// 	if !t.DeskData.Pub {
	// 		//return
	// 	}
	// default:
	// 	//return
	// }
	if len(t.roles) != 0 {
		return
	}
	// g := config.GetGame(t.DeskData.Unique)
	// if g.Id == t.DeskData.Unique {
	// 	return //配置房间不关闭
	// }
	//停止服务
	msg1 := new(pb.ServeStop)
	t.selfPid.Tell(msg1)
}

func (t *Desk) checkPubOver2() {
	switch t.state {
	case int32(pb.STATE_READY):
		t.closeTime++
		if t.closeTime == 60 {
			t.closeTime = 0
			// t.checkPubOver()
			t.KickRobot()
		}
	default:
		t.closeTime = 0
	}
}

//.

// '结束连庄处理,赢家当庄
func (t *Desk) dealerOver() {
	t.DeskGame.DealerSeat = t.DeskAct.ActSeat
	if val, ok := t.seats[t.DeskGame.DealerSeat]; ok {
		t.DeskGame.Dealer = val.Userid
	}
}

//.

// '结束游戏
// 是否过期
func (t *Desk) checkExpire() bool {
	var now = utils.Timestamp()
	if now > int64(t.DeskData.Expire) {
		glog.Debugf("game stop expire -> %d, %d",
			t.DeskData.Expire, now)
		return true
	}
	return false
}

// 是否结束游戏
func (t *Desk) checkOver() bool {
	if t.DeskData.Round == t.DeskGame.Round {
		glog.Debugf("game stop round -> %d, %d",
			t.DeskGame.Round, t.DeskData.Round)
		return true
	}
	return t.checkExpire()
}

// 结束牌局
func (t *Desk) gameStop() {
	if !t.checkOver() {
		return
	}
	if t.DeskData.Pub { //大厅房间不解散
		//return
	}
	//返回未开局钻石
	t.backCost()
	//停止服务
	msg1 := new(pb.ServeStop)
	t.selfPid.Tell(msg1)
}

// 返还钻石
func (t *Desk) backCost() {
	//已经打过的房间不返还
	if t.DeskGame.Round != 0 {
		return
	}
	//已经开始游戏不返还
	if t.state != int32(pb.STATE_READY) {
		return
	}
	//A
	if t.DeskData.Payment != 1 {
		t.sendDiamond(t.DeskData.Cid,
			int64(t.DeskData.Cost), int32(pb.LOG_TYPE3))
		return
	}
	//AA
	for k := range t.roles {
		t.sendDiamond(k, int64(t.DeskData.Cost), int32(pb.LOG_TYPE3))
	}
}

//.

// '个人记录
func (t *Desk) setRecord(score map[uint32]int64) {
	for k, v := range score {
		user := t.getUserBySeat(k)
		if user == nil {
			continue
		}
		pid := t.getPid(user.GetUserid())
		if pid == nil {
			continue
		}
		msg := new(pb.SetRecord)
		if v > 0 {
			msg.Rtype = 1
		} else if v < 0 {
			msg.Rtype = -1
		} else {
			msg.Rtype = 0
		}
		//更新游戏内数据
		user.SetRecord(msg.Rtype)
		//更新节点数据
		pid.Tell(msg)
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
	ming_tax := t.Game.JOKER.MingTax
	an_tax := t.Game.JOKER.AnTax

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

//.

// '结算
func (t *Desk) jiesuan2(ltype int32, score map[uint32]int64) {
	for k, v := range score {
		userid := t.getUserid(k)
		//抽水
		// v = t.drawcoin(userid, v)
		switch t.DeskData.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			if v > 0 {
				t.sendCurrency(userid, v, ltype, fmt.Sprintf("joker%s房间赢分", t.DeskData.Rid))
			}
		case int32(pb.ROOM_TYPE1): //私人
			if v > 0 {
				t.sendCoin(userid, v, ltype)
			}
			t.DeskPriv.PrivScore[userid] += v
		}
	}
}

// 抽水
func (t *Desk) drawcoin(userid string, val int64) int64 {
	// if val <= 0 {
	// 	return val
	// }
	// var num int64 = handler.DrawCoin(t.DeskData.Rtype, t.DeskData.Mode, val)
	// switch t.DeskData.Rtype {
	// case int32(pb.ROOM_TYPE0), //自由
	// 	int32(pb.ROOM_TYPE1): //私人
	// 	//反佣和收益消息,抽成日志记录 val - num
	// 	msg2 := handler.AgentProfitNumMsg(userid, t.DeskData.Gtype, num)
	// 	t.send3userid(userid, msg2)
	// case int32(pb.ROOM_TYPE2): //百人
	// 	//反佣和收益消息,抽成日志记录 val - num
	// 	msg2 := handler.AgentProfitNumMsg(userid, t.DeskData.Gtype, num)
	// 	t.send3userid(userid, msg2)
	// }
	// return val - num
	return 0
}

// 开始前扣除抽水
func (t *Desk) drawfee() {
	// switch t.DeskData.Rtype {
	// case int32(pb.ROOM_TYPE0), //自由
	// 	int32(pb.ROOM_TYPE1): //私人
	// case int32(pb.ROOM_TYPE2): //百人
	// 	return
	// }
	// if t.state != int32(pb.STATE_READY) {
	// 	return
	// }
	// //计算反佣和收益
	// var num int64 = handler.DrawFee(t.DeskData.Mode, t.DeskData.Ante)
	// for k, v := range t.seats {
	// 	if !v.Ready {
	// 		continue
	// 	}
	// 	if num <= 0 {
	// 		continue
	// 	}
	// 	t.sendCoin(v.Userid, (-1 * num), int32(pb.LOG_TYPE48))
	// 	//抽水消息广播
	// 	msg := &pb.JOKERPushDrawCoinNtf{
	// 		Rtype:  uint32(pb.LOG_TYPE48),
	// 		Userid: v.Userid,
	// 		Seat:   k,
	// 		Coin:   (-1 * num),
	// 	}
	// 	t.broadcast(msg)
	// 	//反佣和收益消息
	// 	msg2 := handler.AgentProfitNumMsg(v.Userid, t.DeskData.Gtype, num)
	// 	t.send3userid(v.Userid, msg2)
	// }
}

// 保存详情
func (t *Desk) saveDetail() {
	detail, err := json.Marshal(t.detail)
	if err != nil {
		glog.Errorf("save detail error %#v", t.detail)
		return
	}
	msg := &pb.Detail{}
	msg.Data = detail
	myactor.Logger().Tell(msg)
}

// 日志记录
func (t *Desk) saveRecord(score map[uint32]int64) {
	msg := new(pb.RoundRecord)
	msg.Roomid = t.DeskData.Rid
	msg.Round = t.DeskData.Round
	msg.Dealer = t.DeskGame.Dealer
	for k, v := range score {
		if val, ok := t.seats[k]; ok {
			msg2 := &pb.RoundRoleRecord{
				Userid: val.Userid,
				Cards:  val.Cards,
				Value:  val.Power,
				Bets:   val.Bet,
				Score:  v,
			}
			if val2, ok2 := t.roles[val.Userid]; ok2 {
				msg2.Rest = val2.User.GetCoin()
			}
			msg.Roles = append(msg.Roles, msg2)
		}
	}
	myactor.Logger().Tell(msg)
	for k := range score {
		user := t.getUserBySeat(k)
		if user == nil {
			continue
		}
		msg1 := new(pb.RoleRecord)
		msg1.Roomid = t.DeskData.Rid
		msg1.Gtype = t.DeskData.Gtype
		msg1.Userid = user.GetUserid()
		msg1.Nickname = user.GetNickname()
		msg1.Photo = user.GetPhoto()
		msg1.Rest = user.GetCoin()
		if t.DeskPriv != nil {
			msg1.Score = t.DeskPriv.PrivScore[user.GetUserid()]
			msg1.Joins = t.DeskPriv.Joins[user.GetUserid()]
		}
		myactor.Logger().Tell(msg1)
	}
}

// 移除多余的假位置
func (t *Desk) kickFakeSeat() {
	robotNum := t.robotNum
	size := len(t.FakeSeats)
	if robotNum >= size {
		return
	}
	removeAll := false
	if t.timer == ReadyTime {
		// 没坐人的去不移除
		removeAll = true
	}
	if !removeAll && !utils.RandWan(5000) {
		// 单个退出为概率性
		return
	}
	remove := make([]uint32, 0)
	for k := range t.FakeSeats {
		if _, ok := t.seats[k]; !ok {
			remove = append(remove, k)
			if removeAll {
				continue
			}
			break
		}
	}
	if len(remove) <= 0 {
		return
	}
	for _, v := range remove {
		delete(t.FakeSeats, v)
		// 通知
		ntf := &pb.CloseIdleSeatNtf{
			Seat:   v,
			Gtype:  t.Gtype,
			Roomid: t.Rid,
		}
		t.broadcast5(ntf)
	}
}

func (t *Desk) checkRealUser() {
	r, _ := t.roleCountNum()
	if r >= 2 {
		removeAll := false
		if ReadyTime == t.timer {
			removeAll = true // 时间到了直接把所有人机移出去
		}
		// 有超过2个真实玩家,踢出人机
		for k, v := range t.roles {
			if !v.Robot {
				continue
			}
			errcode := pb.OK
			t.notifyGateUserLeft(k, errcode, 0)
			t.userLeaveDesk(k)
			if removeAll {
				continue
			}
			break
		}
	}
}

// 强执换桌
func (t *Desk) mutexChangeDesk() {
	r, _ := t.roleCountNumNoWatch()
	if r < 2 {
		return
	}
	for _, s := range t.seats {
		if s.Watch {
			continue
		}
		if role, ok := t.roles[s.Userid]; ok {
			if role.Robot {
				continue
			}
			if r <= 1 {
				break
			}
			r--
			ntf := &pb.MutexChangeDeskNtf{
				Userid: role.Userid,
				Gtype:  t.Gtype,
				Roomid: t.Rid,
			}
			t.send2userid(role.Userid, ntf)
		}
	}
}

//.

// vim: set foldmethod=marker foldmarker=//',//.:
