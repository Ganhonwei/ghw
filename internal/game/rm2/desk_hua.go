package rm2

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/algo"
	"goserver/pkg/game/config"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"goserver/pkg/utils"
	"math"
	"strconv"
	"time"
)

// '进入房间响应消息
func (t *Desk) privEnterMsg(userid string) *pb.RMEnterRoomRsp {
	msg := new(pb.RMEnterRoomRsp)
	//房间数据
	msg.Roominfo = handler.PackRMCoinRoom(t.DeskData)
	msg.Gmode = t.Gmode
	msg.Roominfo.State = t.state
	msg.Roominfo.CurRound = t.DeskGame.Round
	//TODO 添加操作信息
	//坐下玩家信息
	msg.Userinfo = t.coinSeatBetsMsg(userid)
	//位置下注信息
	msg.Betsinfo = t.coinBetsMsg()
	//投票信息
	msg.Voteinfo = t.voteInfoMsg()

	// 房间暂停状态原因
	if t.state == int32(pb.STATE_PAUSE) {
		msg.Roominfo.StatePauseReaon = int32(t.reason)
	}
	if t.DeskAct != nil {
		msg.Actseat = t.DeskAct.ActSeat
		msg.Actstate = t.DeskAct.ActState
		msg.Ante = uint32(t.DeskAct.ActAnte)
		msg.Pot = t.DeskGame.BetNum
		msg.Roominfo.Dealer = t.getSeatid(t.DeskData.Cid)
		msg.Timer = int32(BetTime - t.timer)
		msg.Dealer = t.DeskGame.DealerSeat //庄家位置
		msg.Totaltimer = BetTime

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
			msg.ChargingSeats = t.chargingSeats
		}
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

// 投票信息
func (t *Desk) voteInfoMsg() (msg *pb.RMRoomVote) {
	msg = new(pb.RMRoomVote)
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
		msg.DismissCountdown = 10 - t.dismissDelaySeconds
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
func (t *Desk) launchVote(userid string, vote uint32) (msg *pb.RMLaunchVoteRsp) {
	msg = new(pb.RMLaunchVoteRsp)
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
	seat := t.getSeatid(userid)
	if v, ok := t.seats[seat]; ok {
		v.Vote = vote //投票
	}
	//发起投票者
	t.DeskPriv.VoteSeat = seat
	t.DeskPriv.Votes = []uint32{vote}
	t.DeskPriv.LaunchVoteTimes++
	//超时10s
	glog.Debugf("VoteTime: %d, %d, %s", seat, vote, userid)
	t.DeskPriv.VoteStartTime = utils.Timestamp()
	t.DeskPriv.VoteTime = utils.Timestamp() + 10
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
	// 未发起投票或投票结束
	if t.DeskPriv.VoteSeat == 0 || t.DeskPriv.VoteTime == 0 {
		return
	}
	var now = utils.Timestamp()
	if now >= t.DeskPriv.VoteTime {
		t.dismiss(true)
	}
}

// 投票
func (t *Desk) privVote(userid string, vote uint32) (msg *pb.RMVoteRsp) {
	msg = new(pb.RMVoteRsp)
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
		t.DeskPriv.Votes = append(t.DeskPriv.Votes, vote)
	}
	t.pushVote(seat, vote)
	t.dismiss(false)
	return
}

// 广播投票消息
func (t *Desk) pushVote(seat, vote uint32) {
	msg := &pb.RMVoteRsp{
		Seat: seat,
		Vote: vote,
	}
	msg.Agree, msg.Disagree, msg.Votes = t.voteStat()

	t.broadcast(msg)
}

// 广播投票消息
func (t *Desk) pushVoteResult(vote uint32) {
	msg := &pb.RMVoteResultNtf{
		Vote: vote,
	}
	msg.Agree, msg.Disagree, msg.Votes = t.voteStat()
	t.broadcast(msg)
}

// 统计投票
func (t *Desk) voteStat() (agree []uint32, disagree []uint32, votes []uint32) {
	votes = t.DeskPriv.Votes
	for k, v := range t.seats {
		if v.Vote == 1 {
			agree = append(agree, k)
		} else if v.Vote == 2 { // 没投的算否
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
		//停止服务
		// msg1 := new(pb.ServeStop)
		// t.selfPid.Tell(msg1)
		//待回合结束等10s停止服务
		t.dismissOnRoundOver = true
		t.dismissDelaySeconds = 0
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
		// winner := t.pauseArg.(uint32)
		// t.pauseArg = nil
		// t.gameOver(winner)
	case PAUSE_REASON_4:
		t.state = int32(pb.STATE_READY) //设置房间状态
		t.pushState()
		// 踢出玩家
		t.limitOver()
		switch t.Rtype {
		case int32(pb.ROOM_TYPE1): // 私人房
			// 回合结束踢出离线超时玩家
			t.kickPrivOfflineTimeout()

			// 回合结束解散牌桌
			if t.dismissOnRoundOver {
				t.pushPrivSettle()
				return
			}

			if t.DeskGame.Round < uint32(t.DeskData.Round) {
				if t.Gmode == 0 {
					var chargings []uint32
					for userId, role := range t.roles {
						if role.User.GetScore() < int64(t.Game.Min_Access) {
							seatId := t.getSeatid(userId)
							chargings = append(chargings, seatId)
						}
					}
					// 需要充值的玩家
					if len(chargings) > 0 {
						t.chargeInGameBegin(chargings...)
						return
					}
				}
				if len(t.seats) < 2 {
					// 人数不够下一轮,结算
					glog.Info("desk user just %d, not start next round %d/%d", len(t.seats), t.DeskGame.Round, t.DeskData.Round)
					t.pushPrivSettle()
				} else {
					// 对战自动开始下一轮
					t.gameStart()
				}
			} else {
				t.pushPrivSettle()
				glog.Infof("game round over: %v, %v", t.DeskGame.Round, t.DeskData.Round)
			}
		default:
			// 真实玩家换桌
			t.mutexChangeDesk()
		}
	}

}

func (t *Desk) checkState() {
	if t.state == int32(pb.STATE_FREE) && t.roleNum() >= 2 {
		t.state = int32(pb.STATE_READY)
		t.pushState() //广播状态变更
	}
}

// ' 超时处理
func (t *Desk) coinTimeout() {
	// t.checkPubOver2()
	// t.checkPubOver()
	t.checkState()
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
		var num int = t.readyNum()   //游戏人数
		var num2 int = t.ready2Num() //播放完发牌动画人数
		if num2 >= num {
			return
		}
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
		} else {
			t.timer++
		}
	case int32(pb.STATE_DECLARE):
		if t.timer == BetTime {
			t.timer = 0
			winner := t.getWinner()
			t.gameOver(winner)
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
	msg := new(pb.RMCoinWaitTooLongNtf)
	msg.Seat = seat
	t.broadcast(msg)
}

// ' 超时处理
func (t *Desk) privTimeout() {
	// t.checkPubOver2()
	t.kickPrivOfflineTimeout()
	t.voteTimeout()
	t.againTimeout()
	t.settleExitTimeout()

	switch t.state {
	case int32(pb.STATE_FREE):
		fallthrough
	case int32(pb.STATE_READY):
		if t.dismissOnRoundOver {
			if t.dismissDelaySeconds == 0 {
				// 通知10s后解散
				msg := &pb.DeskDismissNtf{Delay: 10}
				t.broadcast(msg)
			}
			// 回合结束10s后解散
			if t.dismissDelaySeconds >= 10 {
				t.dismissOnRoundOver = false
				t.dismissDelaySeconds = 0
				msg1 := new(pb.ServeStop)
				t.selfPid.Tell(msg1)
			} else {
				t.dismissDelaySeconds++
			}
		}

		//私人房x秒后未开局强制解散
		if t.DeskData.Round == 0 && t.checkExpire() {
			//关闭房间
			glog.Infof("Desk expired: %v, %v", t.state, t.seats)
			t.gameStop()
		}
		// var num int = t.readyNum()
		// if num < 2 { //大于2人时才计时
		// 	return
		// }
		// if t.timer == ReadyTime {
		// 	//准备超时,不等待全部准备
		// 	//t.readyTimeout()
		// 	t.gameStart() //开始牌局
		// 	t.timer = 0
		// } else {
		// 	t.timer++
		// }
		return
	case int32(pb.STATE_DEALING):
		var num int = t.readyNum()   //游戏人数
		var num2 int = t.ready2Num() //播放完发牌动画人数
		if num2 >= num {
			return
		}
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
		} else {
			t.timer++
		}
	case int32(pb.STATE_DECLARE):
		if t.timer == BetTime {
			t.timer = 0
			winner := t.getWinner()
			t.gameOver(winner)
		} else {
			t.timer++
		}
	case int32(pb.STATE_PAUSE):
		if t.timer == t.pauseTime {
			// 从暂停状态恢复游戏
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
		if t.timer == ChargeInGameTime {
			t.timer = 0
			t.chargeTimeout()
		} else {
			t.timer++
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
			//离开状态消息
			//玩家离开牌桌
			t.notifyNodeLeaveEarly(t.Rid, userid, pb.OK)
			t.notifyGateUserLeft(userid, pb.OK, int32(pb.OK))
			//清除数据
			t.userLeaveDesk(userid, pb.PrivOfflineTimeout)
		}
	}
}

// 获取库存ID
func (t *Desk) getStockGameId() (string, bool) {
	// if t.DeskType == int32(pb.DESK_TYPE_NEWBIEW) { //新手模式走新手库存
	// 	return "2106", false
	// } else {
	// 	return t.Game.Id, true
	// }
	return t.Game.Id, true
}

func (t *Desk) getGiveStockGameId() string {
	return "2107"
}

// 更新赠送金库存
func (t *Desk) _changeGiveStock(cash_give int64) {
	if cash_give == 0 {
		return
	}
	msg := &pb.ChangeStock{}
	msg.GameId = t.getGiveStockGameId()
	msg.Gtype = int32(pb.RUMMY2)
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
	msg.Gtype = int32(pb.RUMMY2)
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
			Gtype:     int32(pb.RUMMY2),
			GameTime:  time.Now().Unix() - t.detail.BeginTime,
			CashStock: cash_stock,
			CashMing:  cash_ming,
			CashAn:    cash_an,
			First:     role.RoundGames[int32(pb.RUMMY2)] <= 1,
		}
		t.roomPid.Tell(m)
	}
}

// 处理点控
func (t *Desk) handlePointControl(role *data.DeskRole, rmdetail *data.RMDetail) {
	if t.DeskType != int32(pb.DESK_TYPE_POINTCONTROL) {
		return
	}
	if role.Robot {
		return
	}
	if !role.PCSwitch || role.PCScore == 0 {
		return
	}

	diff := rmdetail.AfterCash - rmdetail.BeforeCash
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

func (t *Desk) packCoinGameOverInfo(seatid uint32, actScore, score_final int64) (over *pb.RMCoinGameOverInfo) {
	//结算数据
	over = &pb.RMCoinGameOverInfo{
		Seat: seatid,
	}
	seat := t.seats[seatid]
	role := t.getRole(seat.Userid)
	if len(seat.SortCards) == 0 {
		ret_cards := seat.Cards
		algo.Sort(ret_cards)
		over.Cards = append(over.Cards, &pb.RMFinishInfo{Cards: ret_cards})
	} else {
		for _, v2 := range seat.SortCards {
			over.Cards = append(over.Cards, &pb.RMFinishInfo{Cards: v2})
		}
	}
	over.Nickname = role.Nickname
	over.Photo = role.Photo
	over.VipLv = int32(role.Vip.Lv)
	over.Point = actScore
	over.WinOrLose = score_final
	t.over[seatid] = over
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

// 输
func (t *Desk) Lose(seatid uint32, zhahu bool, winner uint32, drop bool, drop_score int) {
	userid := t.getUserid(seatid)
	status := t.getStatus(seatid)
	role := t.getRole(userid)
	seat := t.getSeat(seatid)
	if status == nil || role == nil || seat == nil {
		return
	}

	var _score int64 //得分
	if zhahu {       //诈胡
		_score = 80
	} else if drop {
		_score = int64(drop_score)
	} else {
		_score = algo.CalcScore(seat.SortCards, t.WildCard)
		if winner == 0 {
			_score = 0
		}
	}

	status.ActScore = 0 - _score            //得分
	num := _score * int64(t.Game.RM.Bottom) //输分
	if num != 0 {
		t.setBet(seatid, userid, num, fmt.Sprintf("rm%s房间输分", t.DeskData.Rid)) //扣分
	}

	score := -status.ActNum //输赢

	// 对战房
	if t.Rtype == int32(pb.ROOM_TYPE1) {
		loseScore := -num
		t.PrivLoses[userid]++
		t.PrivScore[userid] += loseScore

		// 娱乐模式只扣娱乐分
		if t.Gmode == 1 {
			t.score[seatid] = loseScore
			t.over[seatid] = t.packCoinGameOverInfo(seatid, status.ActScore, loseScore)
			t.RecordDetail(seatid, status.ActNum, loseScore, nil)
			t.GameTime(seatid)
			t.rummyRecord(seatid, loseScore)
			return
		}
	}

	var score_final int64 //最终输
	//非人机计算库存
	if !role.Robot {
		score_deduction := t.CheckGiveDiamond(seatid, score)
		score_final, status.Stock = t.calcStockAndTax(score, role.Ratio, 0)
		status.Stock.CashGive = -score_deduction
		status.Stock.Robot = role.Robot
	} else {
		score_final = score
	}
	t.score[seatid] = score_final
	t.shareAmount(userid, score_final)
	t.RecordDetail(seatid, status.ActNum, score_final, status.Stock)
	t.GameTime(seatid)
	// t.EventPost(seat, false)

	//结算数据
	t.over[seatid] = t.packCoinGameOverInfo(seatid, status.ActScore, score_final)

	t.CheckOutDiamond(seatid, score_final)
	t.rummyRecord(seatid, score_final)
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
	// if !status.Alive {
	// 	return
	// }

	var _score int64 //得分
	for k, v := range t.DeskAct.ActSeats {
		if k != seatid {
			_score += 0 - v.ActScore
		}
	}
	status.ActScore = _score

	score := _score * int64(t.Game.RM.Bottom) //赢分
	var score_final int64                     //最终赢
	//非人机计算库存
	// if !role.Robot {
	score_final, status.Stock = t.calcStockAndTax(score, role.Ratio)
	status.Stock.Robot = role.Robot
	// } else {
	// score_final = score
	// }
	t.score[seatid] = score_final

	// 对战房
	if t.Rtype == int32(pb.ROOM_TYPE1) {
		t.PrivWins[userid]++
		t.PrivScore[userid] += score_final

		// 娱乐模式只加娱乐分
		if t.Gmode == 1 {
			t.sendFraction(userid, score_final, int32(pb.LOG_TYPE45))
			t.over[seatid] = t.packCoinGameOverInfo(seatid, status.ActScore, score_final)
			t.RecordDetail(seatid, status.ActNum, score_final, nil)
			t.GameTime(seatid)
			t.rummyRecord(seatid, score_final)
			return
		}
	}

	t.sendCurrency(userid, score_final, int32(pb.LOG_TYPE128), fmt.Sprintf("rummy%s房间赢分", t.DeskData.Rid)) //加钱

	t.shareAmount(userid, score_final)
	t.RecordDetail(seatid, status.ActNum, score_final, status.Stock)
	t.GameTime(seatid)
	// t.EventPost(seat, false)

	//结算数据
	t.over[seatid] = t.packCoinGameOverInfo(seatid, status.ActScore, score_final)

	// 任务事件
	e := &event.GameRecordEvent{
		Gtype: uint32(pb.RUMMY2),
		Win:   true,
	}
	t.eventPost(userid, event.PLAY_TASK, e)

	outDiamond := score_final
	if t.Rtype == int32(pb.ROOM_TYPE1) {
		//私人房只加5%的可提现金
		outDiamond = int64(math.Round(float64(outDiamond) * 0.05))
	}
	t.CheckOutDiamond(seatid, outDiamond)
	t.rummyRecord(seatid, score_final)
	t.flowWater(userid, score_final)
}

// 记录详情
func (t *Desk) RecordDetail(seatid uint32, act_num, score_final int64, stock *data.ActStock) {
	userid := t.getUserid(seatid)
	role := t.getRole(userid)
	seat := t.getSeat(seatid)
	if role == nil || seat == nil {
		return
	}
	rmdetail := t.detail.FindRMDetail(seatid)
	if rmdetail == nil {
		return
	}
	//总投注、底注
	// rmdetail.Bet = act_num                   //总投注
	// rmdetail.Bottom = int64(t.DeskData.Ante) //底注

	//库存
	if stock != nil {
		rmdetail.CashStock = stock.CashStock
		rmdetail.BonusStock = stock.BonusStock
		rmdetail.CashMingTax = stock.CashMing
		rmdetail.BonusMingTax = stock.BonusMing
		rmdetail.CashAnTax = stock.CashAn
		rmdetail.BonusAnTax = stock.BonusAn
	}

	//结算
	rmdetail.Score = score_final
	if score_final > 0 {
		rmdetail.Result = "赢"
	} else if score_final < 0 {
		rmdetail.Result = "输"
	} else {
		rmdetail.Result = "平"
	}

	//账变后分数、彩金、奖励金
	rmdetail.AfterScore = role.GetScore()
	rmdetail.AfterCash = role.GetDiamond()
	rmdetail.AfterBonus = role.GetCoin()

	//最终牌型
	rmdetail.FinalCards = seat.SortCards

	//处理点控
	t.handlePointControl(role, rmdetail)

}

//	func (t *Desk) EventPost(seat uint32, win bool) {
//		userid := t.getUserid(seat)
//		role := t.getRole(userid)
//		if userid == "" || role == nil {
//			return
//		}
//		if role.Robot {
//			return
//		}
//		e := &event.GameRecordEvent{
//			Gtype: uint32(pb.HUA),
//			Win:   win,
//		}
//		cards := t.getHandCards(seat)
//		typ := algo.HuaType(cards)
//		hands := &event.PokerHandsEvent{
//			GameType: int(pb.HUA),
//			Win:      win,
//			PX:       typ,
//		}
//		t.eventPost(userid, event.PLAY_TASK, e)
//		// t.eventPost(userid, event.WIN_TASK, e)
//		t.eventPost(userid, event.POKER_HANDS, hands)
//	}

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
		Gtype:  int32(pb.RUMMY2),
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

// 换牌操作，返回是否跳过
func (t *Desk) ChangeCard(seatid uint32) bool {
	seat := t.getSeat(seatid)

	var choices []utils.Choice
	choices = append(choices, utils.Choice{Weight: 70, Item: 1}) //边张
	if algo.GetRMLaiziNum(seat.Cards, t.WildCard) < 6 {
		choices = append(choices, utils.Choice{Weight: 30, Item: 2}) //joker
	}
	c, _ := utils.WeightedChoice(choices)
	op := c.Item.(int)

	var card uint32
	if op == 1 {
		card, t.DeskGame.Cards = algo.DealRandomBianZhangCard(t.DeskGame.Cards, seat.Cards)
	} else if op == 2 {
		card, t.DeskGame.Cards = algo.DealRandomJoker(t.DeskGame.Cards, t.WildCard)
	}

	seat.Cards = append(seat.Cards, card)
	_, discard := algo.GetOneCardToDiscard(seat.Cards, t.WildCard)
	seat.Cards = algo.RemoveCard(seat.Cards, discard)
	t.DeskGame.Cards = append(t.DeskGame.Cards, discard)
	sort := algo.GroupTheCards(seat.Cards, t.WildCard)
	// sort := algo.SortCards4(seat.Cards, t.WildCard)
	score := algo.CalcScore(sort, t.WildCard)
	if score < 20 {
		return true
	} else {
		return false
	}
}

// 获取总分
func (t *Desk) GetTotal(seats []uint32) int64 {
	var total int64
	for _, v := range seats {
		seat := t.getSeat(v)
		sort := algo.GroupTheCards(seat.Cards, t.WildCard)
		// sort := algo.SortCards4(seat.Cards, t.WildCard)
		score := algo.CalcScore(sort, t.WildCard)
		total += score
	}
	return total
}

// 处理赢分限制
func (t *Desk) HandleWinScoreLimit(winner uint32) {
	if t.winScoreLimit == 0 {
		return
	}

	player := t.GetOnlyOnePlayer()
	if player == "" {
		return
	}

	seatid := t.getSeatid(player)
	if seatid != winner { //赢的不是玩家
		return
	}

	// var total int64
	// for k, v := range t.DeskAct.ActSeats {
	// 	if k != winner {
	// 		total += 0 - v.ActScore
	// 	}
	// }

	// if total < int64(t.winScoreLimit) {
	// 	return
	// }

	changeMap := make(map[uint32]struct{})
	changeSlice := []uint32{}
	for k, v := range t.seats {
		if !v.Ready { //跳过没有玩的
			continue
		}
		if k == winner { //跳过赢家（玩家）
			continue
		}
		if !t.isRobot(k) { //跳过非机器人
			continue
		}
		status := t.getStatus(k)
		if !status.Alive { //跳过弃牌的
			continue
		}
		changeMap[k] = struct{}{}
		changeSlice = append(changeSlice, k)
	}

	total := t.GetTotal(changeSlice)

	if total < int64(t.winScoreLimit) {
		return
	}

	for {
		if len(changeMap) == 0 {
			break
		}
		if total < int64(t.winScoreLimit) {
			break
		}
		for _, v := range changeSlice {
			if _, ok := changeMap[v]; ok {
				if t.ChangeCard(v) {
					delete(changeMap, v)
				}
			}
		}
		total = t.GetTotal(changeSlice)
	}

	//修改sort数据
	for _, v := range changeSlice {
		seat := t.getSeat(v)
		if seat == nil {
			continue
		}
		seat.SortCards = algo.GroupTheCards(seat.Cards, t.WildCard)
		// seat.SortCards = algo.SortCards4(seat.Cards, t.WildCard)
	}

	t.isTriggerWinScoreLimit = true
}

// '结束游戏
func (t *Desk) gameOver(winner uint32) {
	t.timer = 0
	t.state = int32(pb.STATE_OVER)
	t.pushState()

	t.detail.EndTime = time.Now().Unix() //详情记录结束时间
	t.detail.IsMustLose = t.isMustLose
	t.detail.MustLoseScore = t.mustLoseScore
	t.detail.CoreRobotSeat = t.coreRobotSeat
	t.detail.IsUpCardPool = t.isUpCardPool
	t.detail.CardPoolId = t.cardPoolId
	t.detail.DealType = t.dealType
	t.detail.IsTriggerWinScoreLimit = t.isTriggerWinScoreLimit
	t.detail.IsTriggerMustLose = t.isTriggerMustLose
	t.detail.IsTriggerMustLoseDrawCard = t.isTriggerMustLoseDrawCard
	t.detail.IsTriggerMustLoseLastCard = t.isTriggerMustLoseLastCard
	t.detail.IsTriggerMustLoseCoreRobotGoodCard = t.isTriggerMustLoseCoreRobotGoodCard

	t.HandleWinScoreLimit(winner)

	//当前局积分
	score := make(map[uint32]int64)

	//计算declare玩家得分(drop直接扣)
	for k, v := range t.seats {
		if !v.Ready { //跳过观战玩家
			continue
		}
		if k == winner { //跳过赢家
			continue
		}
		status := t.getStatus(k)
		if !status.Alive { //跳过drop玩家
			continue
		}
		t.Lose(k, false, winner, false, 0)
	}

	if winner != 0 {
		t.Win(winner)
	}

	//结算消息
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0): //自由
		//结算消息
		msg := t.resCoinOver(score)

		// msg.Winner = winner
		// msg.Wincoin = uint64(score[winner])

		t.broadcast(msg)
		//修改库存
		t.changeStock()
		// 游戏结束原因
		t.gameOverRasonRecord(winner)
		//详情
		t.saveDetail()
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
		// 对局数加1
		t.DeskGame.Round++
		//结算消息
		msg := t.resOver(score)
		t.broadcast(msg)
		//记录
		// if !t.DeskData.Pub {
		// 	t.saveRecord(score)
		// }
		//修改库存
		t.changeStock()
		// 游戏结束原因
		t.gameOverRasonRecord(winner)
		//详情
		t.saveDetail()
		//结束连庄处理
		// t.dealerOver()
		//重置状态
		t.gameOverInit()

		//踢出不足坐下玩家或超额玩家
		// t.limitOver()
		//踢除离线玩家
		// t.kickOffline()
		//关闭房间
		//t.gameStop()
	case int32(pb.ROOM_TYPE2): //百人
	}
}

// gameOverRasonRecord 结束原因记录
// 机器人自摸胡牌
// 机器人吃牌胡牌
// 玩家自摸胡牌
// 玩家吃牌胡牌
// 机器人弃牌
// 玩家弃牌
func (t *Desk) gameOverRasonRecord(winner uint32) {
	seat := t.getSeat(winner)
	if seat == nil {
		return
	}
	player := t.getPlayer(seat.Userid)
	if player == nil {
		return
	}
	var action, action2 int
	if len(seat.Actions) > 0 {
		action = seat.Actions[len(seat.Actions)-1]
	}

	for seatid, seat2 := range t.seats {
		if seatid == winner {
			continue
		}
		player2 := t.getPlayer(seat2.Userid)
		if player2 == nil {
			return
		}
		if len(seat2.Actions) == 0 {
			continue
		}
		action2 = seat2.Actions[len(seat2.Actions)-1]
		if utils.SliceIn(action2, ActionQi, ActionZhaHu) {
			break
		}
	}

	// 胡牌赢
	if action == ActionHu {
		if len(seat.Touchs) == 0 {
			// 天胡
			t.detail.RMGameOverReason = 1
		} else {
			switch seat.Touchs[len(seat.Touchs)-1] {
			case TouchCards:
				t.detail.RMGameOverReason = 2
			case TouchQiCards:
				t.detail.RMGameOverReason = 3
			}
		}
	} else if utils.SliceIn(action2, ActionQi, ActionZhaHu) {
		// 对手弃牌
		t.detail.RMGameOverReason = 4
	}
}

// 输赢记录
func (t *Desk) rummyRecord(seat uint32, num int64) {
	userid := t.getUserid(seat)
	role := t.getRole(userid)
	if userid == "" || role == nil {
		return
	}
	if role.Robot {
		return
	}

	var bet int64 = int64(t.Game.RM.Bottom * 40)
	if val, ok := t.DeskAct.ActSeats[seat]; ok {
		if val.Drop {
			bet = int64(t.Game.RM.Bottom * 20)
		}
	}
	msg := &pb.FreeSetRecord{Gtype: int32(pb.RUMMY2), Score: num, Bet: bet}
	if num > 0 { //赢
		msg.Rtype = 1
	} else if num < 0 { //输
		msg.Rtype = -1
	} else { //平
		msg.Rtype = 0
	}
	msg.GameTime = utils.BsonNow().Unix() - t.detail.BeginTime
	t.send2userid(userid, msg)

	if role.RoundGames == nil {
		role.RoundGames = make(map[int32]int32)
	}
	role.RoundGames[int32(pb.RUMMY2)]++

	// 打码上报
	if err = mq.NatsPublish(mq.TopicGameBets, &pb.PublishGameBets{
		UserPid:    role.Pid,
		Userid:     role.Userid,
		Gtype:      int32(pb.RUMMY2),
		Ts:         time.Now().Unix(),
		Bets:       int64(t.Game.RM.Bottom) * 80, // 底注*80
		Robot:      role.Robot,
		Username:   role.Nickname,
		Photo:      role.Photo,
		RegistArea: int32(role.RegistArea),
		VipLv:      int32(role.Vip.Lv),
		SuperId:    role.ShareSuperior,
		WaterId:    t.GameId,
		Score:      num,
	}); err != nil {
		glog.Error("publish user rummy2 bets error", err)
	}

	// 打码量日志
	log := handler.GameFlowWaterLog(num, int(pb.RUMMY2), 0, len(role.RechargeTarge), userid)
	myactor.Logger().Tell(log)

	role.BetRecord = append(role.BetRecord, bet)
	if len(role.BetRecord) > 200 {
		role.BetRecord = role.BetRecord[len(role.BetRecord)-200:]
	}

	t.rummyPlayerRecord(seat)
}

// 同步玩家rm对局属性
func (t *Desk) rummyPlayerRecord(seat uint32) {
	userid := t.getUserid(seat)
	role := t.getRole(userid)
	if userid == "" || role == nil {
		return
	}
	if role.Robot {
		return
	}
	msg := &pb.RMRoiRecordsync{
		Userid:              userid,
		RmWithout1StDrop:    int32(role.RmWithout1stDrop),
		RmWithout1StNotDrop: int32(role.RmWithout1stNotDrop),
	}

	if t.RMControl.Ctype == 2 {
		msg.RoiId = t.RMControl.RoiId

		if role.RmRoiDayLimits == nil {
			role.RmRoiDayLimits = make(map[string]int32)
		}
		if role.RmRoiLimits == nil {
			role.RmRoiLimits = make(map[string]int32)
		}
		role.RmRoiDayLimits[t.RMControl.RoiId]++
		role.RmRoiLimits[t.RMControl.RoiId]++
	}
	t.send2userid(userid, msg)
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
	// if role.OutDiamond >= 100*100 && role.Money == 0 {
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
	if role.TotalGameTime >= 120*60 && role.Money == 0 {
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
		// 私人房会单独触发局内充值
		if !t.DeskData.Pub {
			return
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
			t.checkPubOver()
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
	//停止服务
	msg1 := new(pb.ServeStop)
	t.selfPid.Tell(msg1)
}

// 返还钻石
func (t *Desk) backCost() {
	//已经打过的房间不返还
	if t.DeskPriv == nil ||
		t.DeskGame.Round != 0 {
		return
	}
	//已经开始游戏不返还
	if t.state != int32(pb.STATE_FREE) &&
		t.state != int32(pb.STATE_READY) {
		glog.Errorf("game start priv room cost not back: %d", t.state)
		return
	}
	if t.DeskData.Cost <= 0 {
		return
	}
	// if game not start update ratio
	if r, ok := t.roles[t.DeskData.Cid]; ok {
		r.Ratio = r.GetRatio()
		if r.GetScore() == 0 { // 避免没钱时加到赠送金
			r.Ratio = 1
		}
	}
	//A
	if t.DeskData.Payment != 1 {
		t.sendCurrency(t.DeskData.Cid,
			int64(t.DeskData.Cost), int32(pb.LOG_TYPE3), "RM娱乐对战房解散返还")
		return
	}
	//AA
	for k := range t.roles {
		t.sendCurrency(k, int64(t.DeskData.Cost), int32(pb.LOG_TYPE3), "RM娱乐对战房解散返还")
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
	ming_tax := t.Game.RM.MingTax
	an_tax := t.Game.RM.AnTax

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
				t.sendCurrency(userid, v, ltype, fmt.Sprintf("rm%s房间赢分", t.DeskData.Rid))
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
	// 	msg := &pb.RMPushDrawCoinNtf{
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
	t.detail.RMControl = t.RMControl
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

// 推送私人房结算通知
func (t *Desk) pushPrivSettle() {
	if t.Rtype != int32(pb.ROOM_TYPE1) {
		return
	}

	t.settleExitTime = utils.Timestamp() + PrivSettleExitDelay
	// 推送结算页
	ntf := &pb.PrivSettleNtf{
		Code:           t.Code,
		Gtype:          t.Gtype,
		Time:           time.Unix(int64(t.Ctime), 0).Format(utils.FORMAT2),
		ExitTime:       PrivSettleExitDelay,
		RoundMinAccess: int64(t.DeskData.Game.Min_Access),
	}
	for userid, player := range t.DeskPriv.PrivPlayer {
		var again bool
		switch t.Gmode {
		case 0:
			if role, ok := t.roles[userid]; ok {
				again = role.GetScore() >= int64(t.DeskData.Game.Min_Access)
			}
		case 1:
			again = true
		}
		vipLv, _ := strconv.Atoi(player[2])
		ntf.Settles = append(ntf.Settles, &pb.PrivSettle{
			Userid:   userid,
			Nickname: player[0],
			Photo:    player[1],
			VipLv:    int32(vipLv),
			Fraction: t.PrivScore[userid],
			Wins:     t.PrivWins[userid],
			Loses:    t.PrivLoses[userid],
			Again:    again,
		})
	}
	t.broadcast(ntf)
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

	seat := t.getSeatid(userid)
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
	seat := t.getSeatid(userid)
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
		} else if v.Again == 2 {
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
				t.sendCurrency(t.DeskData.Cid, -1*int64(t.DeskData.Cost), int32(pb.LOG_TYPE2), fmt.Sprintf("TP娱乐对战房%s再来一局", t.DeskData.Rid))
			}
		}
		// 再次开始游戏
		t.InitDesk()
		t.DeskData.Ctime = uint32(utils.Timestamp())
		t.DeskData.Expire = utils.Timestamp() + int64(config.GetPvpRoom().GameStartWait)
		t.AgainRound++
		t.gameStart()
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

//.

// vim: set foldmethod=marker foldmarker=//',//.:
