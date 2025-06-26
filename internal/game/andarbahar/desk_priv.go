package andarbahar

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/algo"
	"goserver/pkg/game/config"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"goserver/pkg/utils"
	"math"
	"strconv"
	"strings"
)

// 开始游戏 todo 判断庄家
func (t *Desk) privGameStart() pb.ErrCode {
	// 没有庄家不能开始
	if _, ok := t.seats[t.DealerSeat]; !ok || t.Dealer == "" {
		glog.Errorf("an desk no delaer can't start %d", t.DealerSeat)
		return pb.DealerSitFailed
	}
	if len(t.seats) < 2 {
		return pb.DeskUndernumber
	}
	//初始化
	t.privGameInit()
	// 洗牌、开joker牌
	t.privChoiceJoker()
	// 开始下注
	t.privStartBet()
	// //详情初始化
	// t.detailInit()
	// //扣底注
	// t.raiseAnte()
	// //选庄
	// t.dealerHandler()
	return pb.OK
}

// 私人房游戏开始
func (t *Desk) privGameInit() {
	// 记录私人房间参与游戏玩家信息,中途退出在结算时一起展示
	for userid, role := range t.roles {
		if _, ok := t.DeskPriv.PrivPlayer[userid]; !ok {
			vipLv := strconv.Itoa(role.Vip.Lv)
			t.DeskPriv.PrivPlayer[userid] = [3]string{role.Nickname, role.Photo, vipLv}
		}
		role.Ratio = role.GetRatio()
	}

	t.DeskGame.GameId = handler.GenGameId(t.Gtype) // 生成对局号
	t.DeskGame.BeginTime = utils.BsonNow().Unix()  // 开始时间
	t.changingSeat = make(map[string]*ChangingSeat)
	t.ABDeskPriv = new(data.ABDeskPriv)
	t.ABDeskPriv.Joker = 0
	t.ABDeskPriv.Bets = make(map[string]int64)
	t.ABDeskPriv.AndarCards = make([]uint32, 0)
	t.ABDeskPriv.BaharCards = make([]uint32, 0)
	t.ABDeskPriv.AndarBets = make(map[string]int64)
	t.ABDeskPriv.BaharBets = make(map[string]int64)
	t.ABDeskPriv.ABScore = make(map[string]int64)
	t.ABDeskPriv.ABRoundBetTimes = make(map[string]int32)
	t.ABDeskPriv.CardRoundTime = 0
	t.ABDeskPriv.CardRound = 0
	t.ABDeskPriv.Winner = 0
}

// 私人房开始下注
func (t *Desk) privStartBet() {
	t.state = int32(pb.STATE_BET)
	t.CardRoundTime = utils.Timestamp()
	msg := &pb.ABPushStateNtf{
		State:  pb.DeskState(t.state),
		During: t.Game.AB.BetTime,
	}
	t.broadcast(msg)
}

// 状态消息
func (t *Desk) pushState() {
	msg := &pb.ABPushStateNtf{
		State: pb.DeskState(t.state),
	}
	t.broadcast(msg)
}

// 选择joker牌
func (t *Desk) privChoiceJoker() {
	t.shuffle()
	// t.testTailJoker();

	t.ABDeskPriv.Joker = t.DeskGame.Cards[0] // joker
	t.DeskGame.Cards = t.DeskGame.Cards[1:]
	t.state = int32(pb.STATE_DEALING)
	t.pushState()

	msg := &pb.ABJokerNtf{
		Value: t.ABDeskPriv.Joker,
	}
	t.broadcast(msg)
}

// testTailJoker 把joker牌排到最后，配合前端测试用
func (t *Desk) testTailJoker() {
	times, length := 0, len(t.DeskGame.Cards)
	tail := t.DeskGame.Cards[length-1]
	for i, card := range t.DeskGame.Cards {
		if algo.Rank(card) == algo.Rank(tail) {
			t.DeskGame.Cards[i], t.DeskGame.Cards[length-times-2] = t.DeskGame.Cards[length-times-2], t.DeskGame.Cards[i]
			times++
			if times == 3 {
				break
			}
		}
	}
	t.DeskGame.Cards[0], t.DeskGame.Cards[length-4] = t.DeskGame.Cards[length-4], t.DeskGame.Cards[0]
}

// 私人房超时处理
func (t *Desk) privTimeout() {
	t.kickPrivOfflineTimeout() // 踢出离线超时玩家
	t.changingSeatTimeout()
	t.voteTimeout() // 解散投票超时
	// 再来一句投票超时
	t.againTimeout()
	// 结算页倒计时
	t.settleExitTimeout()

	switch t.state {
	case int32(pb.STATE_FREE):
		fallthrough
	case int32(pb.STATE_READY):
		if t.dismissOnRoundOver {
			now := utils.Timestamp()
			if t.dismissDelayTime == 0 {
				t.dismissDelayTime = now + 10
				// 通知10s后解散
				msg := &pb.DeskDismissNtf{Delay: 10}
				t.broadcast(msg)
				return
			}
			// 回合结束10s后解散
			if now >= t.dismissDelayTime {
				glog.Infof("dismiss desk: %v", t.DeskData.Rid)
				t.dismissOnRoundOver = false
				t.dismissDelayTime = 0
				msg1 := new(pb.ServeStop)
				t.selfPid.Tell(msg1)
			}
		}

		//私人房x秒后未开局强制解散
		if t.DeskGame.Round == 0 && t.checkExpire() {
			//关闭房间
			glog.Infof("Desk expired: %v, %v", t.state, t.seats)
			t.gameStop()
		}
	case int32(pb.STATE_BET):
		// 下注时间到, 开ab牌
		now := utils.Timestamp()
		if now-t.CardRoundTime > int64(t.Game.AB.BetTime) {
			t.privFlopCard()
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
		return
	case int32(pb.STATE_CHARGE):
		if t.timer == ChargeInGameTime*2 { // 500ms的tick
			t.timer = 0
			t.chargeTimeout()
		} else {
			t.timer++
		}
	}
}

// 离线重连超时玩家踢出
func (t *Desk) kickPrivOfflineTimeout() {
	now := utils.Timestamp()
	for userid, role := range t.roles {
		if role.Offline && now > role.PrivOfflineTimeout {
			errcode := t.privLeaveCheck(userid)
			if errcode != pb.OK && errcode != pb.LeaveEarly {
				// 牌局未开始,房主离线超时了解散房间
				if !t.IsGaming() &&
					userid == t.DeskData.Cid &&
					t.DeskGame.Round == 0 &&
					t.DeskPriv.AgainRound == 0 {
					//停止服务
					msg1 := new(pb.ServeStop)
					t.selfPid.Tell(msg1)
					return
				}
				continue
			}
			glog.Infof("kickout offline timeout user: %s", userid)
			//通知网关离开
			t.notifyGateUserLeft(userid, pb.OK, int32(pb.OK))
			//清除数据
			t.userLeaveDesk(userid, pb.PrivOfflineTimeout)
		}
	}
}

// 第一/二轮发牌
func (t *Desk) privFlopCard() {
	// 开牌状态
	t.state = int32(pb.STATE_LEAD)
	t.pushState()
	t.ABDeskPriv.CardRound++

	if t.ABDeskPriv.CardRound == 1 {
		t.privFlopCard1()
	} else {
		t.privFlopCard2()
	}
}

// 第一轮翻牌
func (t *Desk) privFlopCard1() {
	card1 := t.DeskGame.Cards[0]
	card2 := t.DeskGame.Cards[1]
	t.ABDeskPriv.AndarCards = append(t.ABDeskPriv.AndarCards, card1)
	t.ABDeskPriv.BaharCards = append(t.ABDeskPriv.BaharCards, card2)

	var winSeat, offset uint32 = 0, 2
	switch algo.Rank(t.DeskPriv.Joker) {
	case algo.Rank(card1):
		winSeat = 1
		t.ABDeskPriv.BaharCards = make([]uint32, 0) // 不发bahar牌
		offset = 1
	case algo.Rank(card2):
		winSeat = 2
	}
	t.DeskGame.Cards = t.DeskGame.Cards[offset:]

	// 广播开牌
	msg := &pb.ABFlopNtf{
		CardRound: t.ABDeskPriv.CardRound,
		Andar:     t.ABDeskPriv.AndarCards,
		Bahar:     t.ABDeskPriv.BaharCards,
	}
	t.broadcast(msg)

	if winSeat == 0 {
		// 发第一轮牌动画
		t.pauseGame(PAUSE_REASON_1, 4, nil)

	} else {
		// 发到 joker 牌 game over
		glog.Info("ab priv room game over %d win", winSeat)
		t.ABDeskPriv.Winner = winSeat
		t.privGameOver()

	}
}

// 第二轮翻牌
func (t *Desk) privFlopCard2() {
	for i := 0; i < len(t.DeskGame.Cards); i++ {
		card := t.DeskGame.Cards[i]
		seat := uint32(i%2 + 1)
		switch seat {
		case 1:
			t.ABDeskPriv.AndarCards = append(t.ABDeskPriv.AndarCards, card)
		case 2:
			t.ABDeskPriv.BaharCards = append(t.ABDeskPriv.BaharCards, card)
		}

		// 发到 joker 牌
		if algo.Rank(card) == algo.Rank(t.DeskPriv.Joker) {
			t.ABDeskPriv.Winner = seat
			// 广播开牌
			msg := &pb.ABFlopNtf{
				CardRound: t.ABDeskPriv.CardRound,
				Andar:     t.ABDeskPriv.AndarCards,
				Bahar:     t.ABDeskPriv.BaharCards,
			}
			t.broadcast(msg)
			// 发第二轮牌动画
			t.pauseGame(PAUSE_REASON_2, i+2, nil)

			t.DeskGame.Cards = t.DeskGame.Cards[i+1:]
			return
		}
	}
	glog.Error("not found flop joker card")
}

// 私人房游戏结束结算
func (t *Desk) privGameOver() {
	t.state = int32(pb.STATE_OVER)
	// 对局数加1
	t.DeskGame.Round++

	// 对局结算
	details := t.privGameSettlement()
	// 记录详情
	t.privGameDetail(details)
	// 广播结束消息
	msg := t.resOverPriv()
	t.broadcast(msg)

	// 等待结算动画
	t.pauseGame(PAUSE_REASON_4, 6, nil)
}

// 结账
func (t *Desk) privGameSettlement() (details []*data.ABUserDetail) {
	var winBets, loseBets map[string]int64
	var dealerWin, dealerLose int64
	andar, bahar := t.getPrivAndarBaharBets()
	switch t.ABDeskPriv.Winner {
	case 1: // andar win
		winBets = t.ABDeskPriv.AndarBets
		loseBets = t.ABDeskPriv.BaharBets
		// bahar下注额为庄家盈利
		dealerWin, dealerLose = bahar, andar
	case 2: // bahar win
		winBets = t.ABDeskPriv.BaharBets
		loseBets = t.ABDeskPriv.AndarBets
		// andar下注额为庄家盈利
		dealerWin, dealerLose = andar, bahar
	}

	for userid, role := range t.roles {
		if userid == t.DeskGame.Dealer {
			continue
		}
		win, lose := winBets[userid], loseBets[userid]
		score := win - lose // 玩家赢分
		score_final := score
		var stock *data.ActStock
		var winScore, diemond, coin int64
		if win > lose { // 赢
			// 真金模式计算税
			if t.Gmode == 0 {
				score_final, stock = t.calcStockAndTax(score_final, role.Ratio)
				stock.Robot = false
			}
			// 返还下注额
			winScore = score_final + win

		} else { // 输
			winScore = win * 2 // 返还下注额
		}

		if winScore != 0 {
			switch t.Gmode {
			case 0: // 真金
				diemond, coin = t.sendCurrencyPriv(userid, winScore, int32(pb.LOG_TYPE88),
					fmt.Sprintf("ab房间%s位置%d赢分", t.Game.Id, t.ABDeskPriv.Winner))
			case 1: // 娱乐
				t.sendFraction(userid, winScore, int32(pb.LOG_TYPE88))
			}
		}

		// 结算记录
		t.ABDeskPriv.ABScore[userid] = score_final
		// 私人房最终结算记录
		t.DeskPriv.PrivScore[userid] += score_final
		if win > lose {
			t.DeskPriv.PrivWins[userid]++
		} else {
			t.DeskPriv.PrivLoses[userid]++
		}

		// 对局记录
		details = append(details, t.privCreateDetail(userid, data.Currency{Diamond: diemond, Coin: coin}, stock))
		// 娱乐模式结算over
		if t.Gmode == 1 {
			continue
		}
		// todo stock 详情记录
		// t.RecordDetail(seatid, status.ActNum, score_final, status.Stock)
		// 玩游戏事件
		t.eventPost(userid, event.PLAY_TASK, &event.GameRecordEvent{
			Gtype: uint32(pb.ABAR),
			Win:   win > lose,
		})
		// vip bank 游戏任务
		t.eventPost(userid, event.VB_GAME_TASK, &event.VBGameTaskEvent{
			GameType: int32(pb.ABAR),
			Win:      win > lose,
		})
		if stock != nil {
			// 库存税收记录
			t.changeStock(0, 0, stock.CashMing, stock.BonusMing, 0, 0, t.Game.Id)
		}

		// 检查可提现金, 赢分时+5%
		outDiamond := score_final
		if outDiamond > 0 {
			outDiamond = int64(math.Round(float64(outDiamond) * 0.05))
		}
		t.checkOutDiamond(userid, outDiamond)
	}

	// 庄家结算
	dealerId := t.DeskGame.Dealer
	dealer := t.getRole(dealerId)
	score := dealerWin - dealerLose
	score_final := score
	var stock *data.ActStock
	var diemond, coin int64
	if score > 0 { // 庄家赢
		// 计算税
		score_final, stock = t.calcStockAndTax(score, dealer.Ratio)
		stock.Robot = false
	}
	if score_final != 0 {
		switch t.Gmode {
		case 0:
			diemond, coin = t.sendCurrencyPriv(dealerId, score_final, int32(pb.LOG_TYPE88),
				fmt.Sprintf("ab房间%s位置%d赢分", t.Game.Id, t.ABDeskPriv.Winner))
		case 1:
			t.sendFraction(dealerId, score_final, int32(pb.LOG_TYPE88))
		}
	}
	// 结算记录
	t.ABDeskPriv.ABScore[dealerId] = score_final
	// 私人房结算输赢记录
	t.DeskPriv.PrivScore[dealerId] += score_final
	if score_final > 0 {
		t.DeskPriv.PrivWins[dealerId]++
	} else {
		t.DeskPriv.PrivLoses[dealerId]++
	}

	// 对局记录
	details = append(details, t.privCreateDetail(dealerId, data.Currency{Diamond: diemond, Coin: coin}, stock))
	// 娱乐模式结算over
	if t.Gmode == 1 {
		return
	}
	// 玩游戏事件
	t.eventPost(dealerId, event.PLAY_TASK, &event.GameRecordEvent{
		Gtype: uint32(pb.ABAR),
		Win:   dealerWin > dealerLose,
	})
	// vip bank 游戏任务
	t.eventPost(dealerId, event.VB_GAME_TASK, &event.VBGameTaskEvent{
		GameType: int32(pb.ABAR),
		Win:      dealerWin > dealerLose,
	})
	if stock != nil {
		// 库存税收记录
		t.changeStock(0, 0, stock.CashMing, stock.BonusMing, 0, 0, t.Game.Id)
	}

	// 检查可提现金, 赢分时+5%
	outDiamond := score_final
	if outDiamond > 0 {
		outDiamond = int64(math.Round(float64(outDiamond) * 0.05))
	}
	t.checkOutDiamond(dealerId, outDiamond)
	return
}

// 对局详情记录
func (t *Desk) privCreateDetail(userid string, score data.Currency, stock *data.ActStock) (detail *data.ABUserDetail) {
	detail = &data.ABUserDetail{
		Userid: userid,
		Win:    score.GetSum(),
		Result: t.getWinReslut(score.GetSum()),
	}
	// 获取玩家位置对应下注
	detail.SeatBets = make(map[string]int64)
	if bets, ok := t.DeskPriv.ABDeskPriv.AndarBets[userid]; ok {
		detail.SeatBets[utils.String(uint32(Andar))] += bets
	}
	if bets, ok := t.DeskPriv.ABDeskPriv.BaharBets[userid]; ok {
		detail.SeatBets[utils.String(uint32(Bahar))] += bets
	}

	role := t.roles[userid]
	if role != nil {
		detail.AfterScore = role.GetScore()
		detail.AfterBonus = role.Coin
		detail.AfterCash = role.Diamond
		detail.BeforeScore = role.GetScore() - score.GetSum()
		detail.BeforeBonus = role.Coin - score.Coin
		detail.BeforeCash = role.Diamond - score.Diamond
	}
	if stock != nil {
		detail.CashMingTax = stock.CashMing
		detail.BonusMingTax = stock.BonusMing
		detail.CashAnTax = stock.CashAn
		detail.BonusAnTax = stock.BonusAn
	}
	return
}

// 对局详情
func (t *Desk) privGameDetail(details []*data.ABUserDetail) {
	if t.Rtype != int32(pb.ROOM_TYPE1) {
		return
	}
	andar, bahar := t.getPrivAndarBaharBets()
	de := data.ABDetail{
		Winner:       t.ABDeskPriv.Winner,
		SideWinner:   0,
		Joker:        t.ABDeskPriv.Joker,
		ACards:       t.ABDeskPriv.AndarCards,
		BCards:       t.ABDeskPriv.BaharCards,
		JackpotValue: t.getPrivJackpotValue(),
		Bets:         andar + bahar,
		UserDetail:   details,
	}
	var playerWin int64 = 0
	for _, ld := range details {
		playerWin += ld.Win
	}
	de.PlayerWin = playerWin
	// 保存详情
	t.savePrivDetail(de)
}

// 中奖的牌值
func (t *Desk) getPrivJackpotValue() uint32 {
	if t.ABDeskPriv.Winner == Andar {
		return t.ABDeskPriv.AndarCards[len(t.ABDeskPriv.AndarCards)-1]
	}
	return t.ABDeskPriv.BaharCards[len(t.ABDeskPriv.BaharCards)-1]
}

// 同意换座申请超时
func (t *Desk) changingSeatTimeout() {
	now := utils.Timestamp()
	for k, c := range t.changingSeat {
		if c.Accept == 0 {
			if now-c.Time > 10 {
				c.Accept = 2
				// 10秒超时
				msg := &pb.ABChangeSeatAcceptRsp{Accept: 2}
				t.send2userid(c.Userid, msg)
				t.send2userid(c.ToUserid, msg)
			}
			continue
		}
		// 60s内不能再发起换座, 超过不用记录了
		if now-c.Time > 60 {
			delete(t.changingSeat, k)
		}
	}
}

// 发起投票
func (t *Desk) launchVote(userid string, vote uint32) (msg *pb.ABLaunchVoteRsp) {
	msg = new(pb.ABLaunchVoteRsp)
	errcode := t.checkVote()
	if errcode != pb.OK {
		msg.Error = errcode
		return
	}
	if t.DeskPriv.VoteSeat != 0 {
		msg.Error = pb.VotingCantLaunchVote
		return
	}
	// 投票间隔60s
	if t.DeskPriv.VoteStartTime > utils.Timestamp()-60 {
		msg.Error = pb.VoteIntervalWait
		return
	}
	// 已投票通过回合结束解散
	if t.dismissOnRoundOver {
		msg.Error = pb.LeaveDismiss
		return
	}
	// 只能发起3次投票
	if t.DeskPriv.LaunchVoteTimes >= 3 {
		msg.Error = pb.NotTimes
		return
	}
	seat := t.getSeat(userid)
	if v, ok := t.seats[seat]; ok {
		v.Vote = vote //投票
	}
	//发起投票者
	t.DeskPriv.VoteSeat = seat
	t.DeskPriv.Votes = []uint32{vote}
	t.DeskPriv.LaunchVoteTimes++
	//超时设置(10秒)
	glog.Debugf("VoteTime: %d, %d, %s", seat, vote, userid)
	t.DeskPriv.VoteStartTime = utils.Timestamp()
	t.DeskPriv.VoteTime = utils.Timestamp() + 10

	msg.Seat = seat
	t.broadcast(msg)
	t.pushVote(seat, vote)
	t.dismiss(false)
	return
}

// 投票
func (t *Desk) privVote(userid string, vote uint32) (msg *pb.ABVoteRsp) {
	msg = new(pb.ABVoteRsp)
	errcode := t.checkVote()
	if errcode != pb.OK {
		msg.Error = errcode
		return
	}
	if t.DeskPriv.VoteSeat == 0 {
		msg.Error = pb.NotVoteTime
		return
	}
	seat := t.getSeat(userid)
	if v, ok := t.seats[seat]; ok {
		v.Vote = vote //投票
		t.DeskPriv.Votes = append(t.DeskPriv.Votes, vote)
	}
	t.pushVote(seat, vote)
	t.dismiss(false)
	return
}

// 投票超时
func (t *Desk) voteTimeout() {
	errcode := t.checkVote()
	if errcode != pb.OK {
		return
	}
	// 未发起投票或投票结束
	if t.DeskPriv.VoteSeat == 0 || t.DeskPriv.VoteTime == 0 {
		return
	}
	var now = utils.Timestamp()
	if now >= t.DeskPriv.VoteTime {
		t.dismiss(true)
	}
}

func (t *Desk) checkVote() pb.ErrCode {
	if t.Rtype != int32(pb.ROOM_TYPE1) {
		return pb.OperateError
	}
	if t.DeskPriv == nil {
		return pb.OperateError
	}
	return pb.OK
}

// 广播投票消息
func (t *Desk) pushVote(seat, vote uint32) {
	msg := &pb.ABVoteRsp{
		Seat: seat,
		Vote: vote,
	}
	msg.Agree, msg.Disagree, msg.Votes = t.voteStat()
	t.broadcast(msg)
}

// 广播投票消息
func (t *Desk) pushVoteResult(vote uint32) {
	msg := &pb.ABVoteResultNtf{Vote: vote}
	msg.Agree, msg.Disagree, msg.Votes = t.voteStat()
	t.broadcast(msg)
}

// 统计投票
func (t *Desk) voteStat() (agree []uint32, disagree []uint32, votes []uint32) {
	votes = t.DeskPriv.Votes
	for k, v := range t.seats {
		if v.Vote == 1 {
			agree = append(agree, k)
		} else if v.Vote == 2 {
			disagree = append(disagree, k)
		}
	}
	return
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
	if agree > unagree {
		//0解散,1不解散
		t.pushVoteResult(0)
		//待回合结束等10s停止服务
		// msg1 := new(pb.ServeStop)
		// t.selfPid.Tell(msg1)
		// 标记回合结束解散
		t.dismissOnRoundOver = true
		t.dismissDelayTime = 0
		t.DeskPriv.VoteTime = 0
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

// 投票信息
func (t *Desk) voteInfoMsg() (msg *pb.ABRoomVote) {
	msg = new(pb.ABRoomVote)
	if t.DeskPriv != nil {
		msg.Seat = t.DeskPriv.VoteSeat
	}
	if msg.Seat == 0 {
		return
	}
	msg.Votes = t.DeskPriv.Votes
	for k, v := range t.seats {
		if v.Vote == 1 {
			msg.Agree = append(msg.Agree, k)
		} else if v.Vote == 2 {
			msg.Disagree = append(msg.Disagree, k)
		}
	}
	msg.Dismiss = t.dismissOnRoundOver
	if t.dismissOnRoundOver {
		msg.DismissCountdown = t.dismissDelayTime - utils.Timestamp()
		if msg.DismissCountdown < 0 {
			msg.DismissCountdown = 0
		}
	}
	return
}

// 获取当前下注额
func (t *Desk) getPrivAndarBaharBets() (andar int64, bahar int64) {
	for _, bet := range t.ABDeskPriv.AndarBets {
		andar += bet
	}
	for _, bet := range t.ABDeskPriv.BaharBets {
		bahar += bet
	}
	return
}

// ab 私人房下注
func (t *Desk) privBet(userid string, seat uint32, chip int64) (rsp *pb.ABBetRsp) {
	rsp = new(pb.ABBetRsp)
	if t.state != int32(pb.STATE_BET) {
		rsp.Error = pb.BetOver
		return
	}
	if chip <= 0 {
		return
	}
	if t.Game.Status == 0 {
		rsp.Error = pb.RoomMaintenance
		return
	}
	user := t.getPlayer(userid)
	if user == nil {
		glog.Errorf("userid %s not exist", userid)
		rsp.Error = pb.NotInRoom
		return
	}
	// 每个下注回合每个位置只能下注一次
	betTimeKey := fmt.Sprintf("%d,%s,%d", t.ABDeskPriv.CardRound, userid, seat)
	if t.ABDeskPriv.ABRoundBetTimes[betTimeKey] > 0 {
		rsp.Error = pb.AlreadyBeted
		return
	}
	//庄家不用下注
	if userid == t.DeskGame.Dealer {
		rsp.Error = pb.BetDealerFailed
		return
	}
	role := t.roles[userid]
	switch t.Gmode {
	case 0: // 真金
		// if !t.canBet(role, chip) {
		if role.GetScore() < chip {
			// 下注不够局内充值
			seatid := t.getSeat(userid)
			rsp.Amount, rsp.GiveAmount = t.getRechargeAmount(seatid)
			rsp.Error = pb.NotEnoughCoin
			t.send2userid(userid, rsp)
			t.chargeInGameBegin(seatid)
			return
		}
	case 1: // 娱乐
		// 娱乐分可以为负
		// if chip > role.GetFraction() {
		// 	return pb.NotEnoughCoin
		// }
	}

	// 是否超出庄家赔付额度
	var seatCoin, userCoin int64 // 下注位置总数, 个人位置总数
	var recover func()
	switch seat {
	case 1:
		t.ABDeskPriv.AndarBets[userid] += chip
		userCoin = t.ABDeskPriv.AndarBets[userid]
		recover = func() { t.ABDeskPriv.AndarBets[userid] -= chip }
	case 2:
		t.ABDeskPriv.BaharBets[userid] += chip
		userCoin = t.ABDeskPriv.BaharBets[userid]
		recover = func() { t.ABDeskPriv.BaharBets[userid] -= chip }
	default:
		glog.Errorf("unknown ab bet seat %v", seat)
		rsp.Error = pb.Failed
		return
	}
	andar, bahar := t.getPrivAndarBaharBets()
	dealerlimit := andar - bahar
	if dealerlimit < 0 {
		dealerlimit = -dealerlimit
	}
	if dealer, ok := t.roles[t.DeskGame.Dealer]; !ok || dealerlimit > dealer.GetScore() {
		recover() // 还原下注金额
		if !ok {
			rsp.Error = pb.NotDealerRoom
			return
		}
		rsp.Error = pb.DealerPayLimit
		return
	}
	if seat == 1 {
		seatCoin = andar
	} else {
		seatCoin = bahar
	}

	// 扣下注额
	switch t.Gmode {
	case 0: // 真金
		t.sendCurrencyPriv(userid, (-1 * chip), int32(pb.LOG_TYPE87),
			fmt.Sprintf("andarbahar%s私人房间下注", t.DeskData.Rid))
	case 1: // 娱乐
		t.sendFraction(userid, (-1 * chip), int32(pb.LOG_TYPE87))
	}
	// 该下注回合位置下注次数+1
	t.ABDeskPriv.ABRoundBetTimes[betTimeKey]++

	t.ABDeskPriv.Bets[userid] += chip
	seatid := t.getSeat(userid)
	rsp = &pb.ABBetRsp{
		Beseat: seat,
		Value:  uint32(chip),
		Userid: userid,
		Coin:   seatCoin,
		Bets:   userCoin,
		Seat:   seatid,
	}
	msg := &pb.ABBetNtf{
		Beseat: seat,
		Value:  uint32(chip),
		Userid: userid,
		Coin:   seatCoin,
		Bets:   userCoin,
		Seat:   seatid,
	}
	t.send2userid(userid, rsp)
	t.broadcast(msg)

	t.chargeBetTimeRecover(seatid)
	return
}

// 局内充值，玩家充值成功已经操作后，不需要在等待100多s
func (t *Desk) chargeBetTimeRecover(seatid uint32) {
	if t.state != int32(pb.STATE_BET) {
		return
	}
	now := utils.Timestamp()
	// 只有充值成功时间下注时间会延长到180s
	if now-t.CardRoundTime > -1*int64(t.Game.AB.BetTime) {
		return
	}
	// 下注的是充值玩家
	if t.betingChargeSeatid != seatid {
		return
	}
	t.betingChargeSeatid = 0
	// 下注时间恢复为10s
	t.CardRoundTime = now + int64(t.Game.AB.BetTime)

	// 广播下注状态时间
	t.privStartBet()
}

// 私人房更新货币
func (t *Desk) sendCurrencyPriv(userid string, score int64, ltype int32, desc string) (diamond, coin int64) {
	if score == 0 {
		return
	}
	//玩家在线
	if v, ok := t.roles[userid]; ok && v != nil {
		// radio := v.GetRatio() // 彩金比值
		diamond = int64(math.Round((float64(score) * v.Ratio)))
		coin = score - diamond

		v.User.AddCoin(coin)
		v.User.AddDiamond(diamond)

		// 记录私人房真金输赢
		var privDiamond int64
		if t.isPrivCashRoom() &&
			(ltype == int32(pb.LOG_TYPE87) || ltype == int32(pb.LOG_TYPE88)) {
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
	return
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

// 游戏结束消息
func (t *Desk) resOverPriv() *pb.ABPrivGameoverNtf {
	msg := &pb.ABPrivGameoverNtf{
		State:      t.state,
		Winner:     int32(t.ABDeskPriv.Winner),
		Userinfo:   t.privSeatBetsMsg(),
		Dealer:     t.DeskGame.Dealer,
		DealerSeat: t.DeskGame.DealerSeat,
		Round:      t.DeskData.Round,
		CurRound:   t.DeskGame.Round,
		LeftRound:  t.DeskData.Round - t.DeskGame.Round,
	}

	// todo scores
	for k, v := range t.ABScore {
		score := &pb.ABRoomScore{
			Userid: k,
			Score:  v,
			Seat:   t.getSeat(k),
		}
		msg.Scores = append(msg.Scores, score)

		// todo 分享活动
		t.shareAmount(k, v)
		// 时长记录
		t.GameTime(k)
	}
	return msg
}

// 发起再来一局投票
func (t *Desk) launchAgain(userid string) (msg *pb.PrivLaunchAgainRsp) {
	var again uint32 = 1
	msg = new(pb.PrivLaunchAgainRsp)
	// 已有结果
	if t.DeskPriv.Again != 0 {
		msg.Error = pb.OperateError
		return
	}
	if t.DeskPriv.AgainSeat != 0 {
		// 已发起，直接投票
		t.privAgain(userid, again)
		return
	}

	seat := t.getSeat(userid)
	// 发起者
	t.DeskPriv.AgainSeat = seat
	//超时设置(10秒)
	glog.Debugf("againTime: %d, %d, %s", seat, again, userid)
	t.DeskPriv.AgainTime = utils.Timestamp() + 10
	msg.Seat = seat
	msg.Userid = userid
	t.broadcast(msg)
	t.DeskPriv.AgainVotes = make([]uint32, 0)
	t.privAgain(userid, again)
	return
}

// 投票
func (t *Desk) privAgain(userid string, again uint32) (msg *pb.PrivAgainRsp) {
	msg = new(pb.PrivAgainRsp)
	errcode := t.checkAgain()
	if errcode != pb.OK {
		msg.Error = errcode
		return
	}
	// 已有结果
	if t.DeskPriv.Again != 0 {
		return
	}
	seat := t.getSeat(userid)
	if v, ok := t.seats[seat]; ok {
		v.Again = again //是否再来一局
		t.DeskPriv.AgainVotes = append(t.DeskPriv.AgainVotes, again)
	}
	t.pushAgain(seat, again)
	t.againResult(false)
	return
}

func (t *Desk) checkAgain() pb.ErrCode {
	if t.DeskPriv == nil {
		return pb.OperateError
	}
	if t.DeskPriv.Again != 0 {
		// 已有结果
		return pb.OperateError
	}
	if t.DeskPriv.AgainSeat == 0 {
		return pb.NotVoteTime
	}
	return pb.OK
}

// 广播再来一句消息
func (t *Desk) pushAgain(seat, again uint32) {
	msg := &pb.PrivAgainRsp{
		Seat:   seat,
		Userid: t.getUserid(seat),
		Again:  again,
	}
	msg.Agree, msg.Disagree, msg.Votes = t.againStat()
	t.broadcast(msg)
}

// 统计再来一局投票
func (t *Desk) againStat() (agree []uint32, disagree []uint32, votes []uint32) {
	votes = t.DeskPriv.AgainVotes
	for k, v := range t.seats {
		if v.Again == 1 {
			agree = append(agree, k)
		} else if v.Again == 2 { // || t.DeskPriv.Again == 2 已有结果没投的算否
			disagree = append(disagree, k)
		}
	}
	return
}

// 再来一局结果统计
func (t *Desk) againResult(force bool) {
	var agree, unagree, unvote int
	for _, v := range t.seats {
		if v.Again == 1 {
			agree++
		} else if v.Again == 2 {
			unagree++
		} else {
			unvote++
		}
	}
	if unagree > 0 || force { // 投票超时
		t.DeskPriv.Again = 2
		//结束投票
		t.pushAgainResult(1)

		// 解散房间
		// msg1 := new(pb.ServeStop)
		// t.selfPid.Tell(msg1)

	} else if unvote == 0 {
		t.DeskPriv.Again = 1
		// 全票通过开启下一局
		t.pushAgainResult(0)

		// 人数不够,不开下一句,等待解散
		if len(t.seats) < 2 {
			return
		}
		//重置
		for _, v := range t.seats {
			v.Again = 0
		}
		t.DeskPriv.AgainSeat = 0
		t.DeskPriv.AgainTime = 0
		t.DeskPriv.Again = 0
		t.settleExitTime = 0 // 不在结算页解散了

		// 私人房娱乐模式，扣房主房费
		if t.isPrivFunRoom() && t.DeskData.Cost > 0 {
			if _, ok := t.roles[t.DeskData.Cid]; ok {
				t.sendCurrencyPriv(t.DeskData.Cid, -1*int64(t.DeskData.Cost), int32(pb.LOG_TYPE2), fmt.Sprintf("TP娱乐对战房%s再来一局", t.DeskData.Rid))
			}
		}
		// 再次开始游戏
		t.InitDesk()
		t.DeskData.Ctime = uint32(utils.Timestamp())
		t.DeskData.Expire = utils.Timestamp() + int64(config.GetPvpRoom().GameStartWait)
		t.AgainRound++
		if err := t.privGameStart(); err != pb.OK {
			// 下一轮失败,结算
			glog.Infof("start failed, game round over: %v, %v, %v", t.DeskGame.Round, t.DeskData.Round, err)
			t.pushPrivSettle()
		}
	}
}

// 广播再来一句投票消息
func (t *Desk) pushAgainResult(again uint32) {
	msg := &pb.PrivAgainResultNtf{Again: again}
	msg.Agree, msg.Disagree, msg.Votes = t.againStat()
	t.broadcast(msg)
}

// 再来一局超时检测
func (t *Desk) againTimeout() {
	errcode := t.checkAgain()
	if errcode != pb.OK {
		return
	}
	glog.Info("desk again vote timeout.")
	var now = utils.Timestamp()
	if now >= t.DeskPriv.AgainTime {
		t.againResult(true)
	}
}

// 结算页再来一句不通过超时退出
func (t *Desk) settleExitTimeout() {
	if t.settleExitTime <= 0 {
		return
	}
	var now = utils.Timestamp()
	if now > t.settleExitTime {
		t.settleExitTime = 0

		// 解散房间
		msg1 := new(pb.ServeStop)
		t.selfPid.Tell(msg1)
	}
}

// 保存详情
func (t *Desk) savePrivDetail(de data.ABDetail) {
	detail := data.Detail{
		WaterId:   t.GameId,
		BeginTime: t.BeginTime,
		EndTime:   utils.BsonNow().Unix(),
		Gtype:     int32(pb.ABAR),
		Rtype:     t.Rtype,
		Gmode:     t.Gmode,
		RoomId:    t.Rid,
		DeskId:    t.Game.Id,
		ABDetail:  &de,
	}
	players := make([]string, 0)
	for _, ld := range de.UserDetail {
		players = append(players, ld.Userid)
	}
	detail.Players = strings.Join(players, ",")
	body, err := json.Marshal(detail)
	if err != nil {
		glog.Errorf("save detail error %#v", detail)
		return
	}
	msg := &pb.Detail{}
	msg.Data = body
	myactor.Logger().Tell(msg)
}

// 局内充值开始
func (t *Desk) chargeInGameBegin(seatids ...uint32) {
	if len(seatids) == 0 {
		return
	}
	if t.state == int32(pb.STATE_CHARGE) {
		return
	}
	t.chargePrevState = t.state
	t.chargingSeats = seatids
	t.state = int32(pb.STATE_CHARGE)
	t.timer = 0

	for _, seatid := range seatids {
		if t.Rtype == int32(pb.ROOM_TYPE1) &&
			t.chargePrevState != int32(pb.STATE_BET) {
			// 私人房回合结束时充值金额推送
			msg1 := new(pb.PrivChargeAmountNtf)
			msg1.Amount, msg1.GiveAmount = t.getRechargeAmount(seatid)
			t.send2seat(seatid, msg1)
		}

		msg2 := new(pb.ChargeInGameNtf)
		msg2.Seat = seatid
		msg2.Totaltimer = ChargeInGameTime + utils.LocalTime().Unix()
		t.broadcast(msg2)
	}
}

// 局内充值取消
func (t *Desk) chargeInGameCancel(userid string) {
	if t.state != int32(pb.STATE_CHARGE) {
		return
	}
	switch t.Rtype {
	case int32(pb.ROOM_TYPE1): // 私人房
		t.state = t.chargePrevState
		if t.state == int32(pb.STATE_READY) {
			t.privRoundChargingResume(true, userid)
		}
	default:
		t.state = int32(pb.STATE_BET)
	}
}

// 私人房回合结束充值恢复
func (t *Desk) privRoundChargingResume(cancel bool, userid string) {
	seatid := t.getSeat(userid)
	var chargingSeats []uint32
	if cancel {
		// 踢出充值中玩家
		for _, seat := range t.chargingSeats {
			if seat != seatid {
				chargingSeats = append(chargingSeats, seat)
				continue
			}
			userid := t.getUserid(seat)
			//玩家离开牌桌
			t.notifyGateUserLeft(userid, pb.OK, int32(pb.OK))
			//清除数据
			t.userLeaveDesk(userid)
		}
	} else {
		// 充值成功
		for _, seat := range t.chargingSeats {
			if seatid != seat {
				chargingSeats = append(chargingSeats, seat)
			}
		}
	}
	t.chargingSeats = chargingSeats
	// 还有充值中玩家
	if len(t.chargingSeats) > 0 {
		return
	}
	t.state = t.chargePrevState
	switch t.state {
	case int32(pb.STATE_BET):
		// 记录下注中充值成功玩家
		t.betingChargeSeatid = seatid
		// 下注时间180s
		t.CardRoundTime = utils.Timestamp() + int64(180-t.Game.AB.BetTime)
		msg := &pb.ABPushStateNtf{
			State:  pb.DeskState(t.state),
			During: 180,
		}
		t.broadcast(msg)
	default: // 回合结束时触发的局内充值
		// 对战自动开始下一轮
		if err := t.privGameStart(); err != pb.OK {
			// 下一轮失败,结算
			glog.Infof("start failed, game round over: %v, %v, %v", t.DeskGame.Round, t.DeskData.Round, err)
			t.pushPrivSettle()
		}
	}
}

// 获取局内充值金额
func (t *Desk) getRechargeAmount(seat uint32) (uint32, uint32) {
	rechargeMap := t.ActRechargeTimes
	configs := t.Game.AB.RoomRecharge
	gives := t.Game.AB.RoomRechargeGive
	if len(configs) <= 0 {
		return 0, 0
	}

	if num, ok := rechargeMap[seat]; ok {
		// rechargeMap[seat] = num + 1
		if len(configs) > num {
			return uint32(configs[num]), uint32(gives[num])
		} else {
			return uint32(configs[len(configs)-1]), uint32(gives[len(gives)-1])
		}
	} else {
		// rechargeMap[seat] = 1
		return uint32(configs[0]), uint32(gives[0])
	}
}

// 充值超时
func (t *Desk) chargeTimeout() {
	if t.state != int32(pb.STATE_CHARGE) {
		return
	}
	switch t.Rtype {
	case int32(pb.ROOM_TYPE1): // 私人房
		t.state = t.chargePrevState
		t.privRoundChargingResume(true, "")
	default:
		glog.Errorf("unkonwn chargeTimeout roomType: %v", t.Rtype)
	}
}

// isPrivRoom 是否是私人房
func (t *Desk) isPrivRoom() bool {
	return t.Rtype == int32(pb.ROOM_TYPE1)
}

// isPrivFunRoom 是否私人房真金模式
func (t *Desk) isPrivCashRoom() bool {
	return t.Rtype == int32(pb.ROOM_TYPE1) && t.Gmode == 0
}

// isPrivFunRoom 是否私人房娱乐模式
func (t *Desk) isPrivFunRoom() bool {
	return t.isPrivRoom() && t.Gmode == 1
}

// privLeftSeats 获取房间剩余座位号
func (t *Desk) privLeftSeats() (restSeats []uint32) {
	for i := 1; i <= int(t.DeskData.Count); i++ {
		seat := uint32(i)
		if _, ok := t.seats[seat]; !ok {
			restSeats = append(restSeats, seat)
		}
	}
	return
}
