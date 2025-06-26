package fortune_gems2

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"time"
)

// 对局初始化
func (t *Desk) gameInit() {
	// t.DeskData.Rtype
	t.GameId = handler.GenGameId(t.Gtype) //初始化流水号
	t.betNum = 0
	clear(t.minesPits)
	clear(t.minesPitsUser)
	t.step = 0
	t.stepPits = t.stepPits[:0]
	t.autoMines = false
	// 未开启自动下注
	if t.autoRound == 0 {
		t.mines = 0
		t.autoPits = nil
	}
	t.minesStepBoomPit = -1
	t.forceCashOut = false
	t.settleScore = 0
	t.settleMultiple = ""

	// 对局详情
	t.detail = &data.Detail{
		BeginTime:   time.Now().Unix(),
		Gtype:       t.Gtype,
		Rtype:       t.Rtype,
		Gmode:       t.Gmode,
		RoomId:      t.Rid,
		DeskId:      t.Game.Id,
		WaterId:     t.GameId,
		RealGame:    t.role != nil,
		MinesDetail: &data.MinesDetail{},
	}
}

func (t *Desk) gameOver(force, _ bool) {
	// 结算
	t.gameSettlement(force)
	// 对局详情
	t.gameDetail(force)

	// 切换结算状态
	t.changeDeskState(pb.STATE_OVER)

	//踢除离线玩家
	t.kickOffline()
}

// 结算
func (t *Desk) gameSettlement(force bool) {
	stepMultiple := getStepMultiple(t.mines, t.step)
	multiple := str2float(stepMultiple)

	var score int64
	if t.betNum > 0 {
		// 玩家输: 踩到雷了/强制结束
		if t.minesStepBoomPit >= 0 || t.forceCashOut {
			score = -t.betNum
		} else {
			score = int64(float64(t.betNum) * multiple)
			// 奖金上限
			winLmit := table.GetTables().MinesTable.Get(1).WinLimit[t.role.RegistArea]
			if score > winLmit {
				score = winLmit
			}
		}
	}
	if score > 0 {
		t.sendCurrency(t.role.Userid, 0, score, int32(pb.LOG_TYPE150), fmt.Sprintf("Mines房间%s返奖", t.DeskData.Rid))
	}

	// 净输赢分
	winScore := utils.CaseElse(score > 0, score-t.betNum, score)

	// 事件(非机器人)
	bean := &event.GameRecordEvent{Gtype: uint32(pb.FORTUNE_GEMS2), Win: winScore > 0}
	// 可提现金变化
	t.checkGiveDiamond(winScore)
	t.eventPost(event.PLAY_TASK, bean) // 玩游戏
	// 分享打码量
	t.shareAmount(t.betNum)
	// 游戏时长
	t.GameTime()
	// vip bank 游戏任务
	t.eventPost(event.VB_GAME_TASK, &event.VBGameTaskEvent{
		GameType: int32(pb.FORTUNE_GEMS2),
		Win:      winScore > 0,
		ActTimes: t.step,
	})

	// 打码上报
	if err = mq.NatsPublish(mq.TopicGameBets, &pb.PublishGameBets{
		UserPid:    t.role.Pid,
		Userid:     t.role.Userid,
		Gtype:      int32(pb.FORTUNE_GEMS2),
		Ts:         time.Now().Unix(),
		Bets:       t.betNum,
		Robot:      t.role.Robot,
		Username:   t.role.Nickname,
		Photo:      t.role.Photo,
		RegistArea: int32(t.role.RegistArea),
		VipLv:      int32(t.role.Vip.Lv),
		SuperId:    t.role.ShareSuperior,
		WaterId:    t.GameId,
		Score:      winScore,
	}); err != nil {
		glog.Error("publish user crash bets error", err)
	}

	t.settleScore = score
	t.settleMultiple = stepMultiple

	// 检查能否继续自动下注
	if t.autoRound > 0 {
		if !t.canBet(t.role.User, t.autoBets) {
			t.autoRound = 0
		}
	}

	winType := 1 // 1.赢,2.输,3.强制结算退还
	if force && t.step == 0 {
		winType = 3
	} else if winScore < 0 {
		winType = 2
	}

	// 展示玩家所有格子
	for pit, pitV := range t.minesPitsUser {
		if pitV == 0 {
			v := t.minesPits[pit]
			if v == 0 {
				t.minesPitsUser[pit] = 3 // 未踩无雷
			} else {
				t.minesPitsUser[pit] = 4 // 未踩有雷
			}
		}
	}

	showMultiple := stepMultiple

	_ = showMultiple
	_ = winType
	// if t.autoMines {
	// 	// 自动对局显示选择的步数倍数
	// 	nextStep := len(t.autoPits) + 1
	// 	maxSafeStep := len(t.minesPits) - int(t.mines)
	// 	showMultiple = getStepMultiple(t.mines, int32(min(nextStep, maxSafeStep)))
	// }
	// msg := &pb.FortuneGems2GameoverNtf{
	// 	Pits:         t.minesPitsUser,
	// 	User:         t.fortuneGems2UserDataMsg(),
	// 	WinType:      int32(winType),
	// 	Multiple:     showMultiple,
	// 	Bets:         int32(t.betNum),
	// 	Score:        score,
	// 	AutoRound:    t.autoRound,
	// 	AutoMines:    t.autoMines,
	// 	BoomPit:      t.minesStepBoomPit,
	// 	ForceCashOut: t.forceCashOut,
	// }
	// t.send2user(msg)
}

// 更新货币
func (t *Desk) sendCurrency(userid string, coin, diamond int64, ltype int32, desc string) {
	if coin == 0 && diamond == 0 {
		return
	}
	if t.role == nil || t.role.Userid != userid {
		return
	}
	//玩家在线
	if v := t.role; v != nil {
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
				Control: false,
			}
			v.Pid.Tell(msg)
			return
		}
	}
	glog.Infof("sendCoin userid %s, coin %d, diamond %d ltype %d", userid, coin, diamond, ltype)
	//TODO 检测是否在其它房间内,如果在则通过房间同步,否则正常同步
	msg := &pb.OfflineCurrency{
		Userid:  userid,
		Coin:    coin,
		Diamond: diamond,
		Type:    ltype,
		Desc:    desc,
		WaterId: t.GameId,
		Control: false,
	}
	//通过大厅通知其它节点
	t.rolePid.Tell(msg)
}

// 进入房间响应消息
func (t *Desk) fortuneGems2EnterMsg(userid string) *pb.FortuneGems2EnterRoomRsp {
	msg := new(pb.FortuneGems2EnterRoomRsp)
	if t.role == nil || t.role.Userid != userid {
		msg.Error = pb.NotInRoom
		return msg
	}

	//房间数据
	msg.Roominfo = handler.PackFortuneGems2Room(t.DeskData)
	t.fortuneGems2RoomDataMsg(msg.Roominfo)

	msg.Gameid = t.Game.Id
	msg.Roomid = t.DeskData.Rid
	msg.GameType = t.Game.Gtype
	msg.RoomType = t.Game.Rtype

	msg.Userinfo = t.fortuneGems2UserDataMsg()
	return msg
}

func (t *Desk) fortuneGems2RoomDataMsg(room *pb.FortuneGems2RoomData) {
	// room.State = t.state
	// room.Bets = t.betNum
	// room.Mines = t.mines
	// room.Pits = t.minesPitsUser
	// room.Bets = t.betNum
	// room.AutoRound = t.autoRound
	// room.AutoPits = t.autoPits
	// room.SettleScore = t.settleScore
	// room.Multiple = getStepMultiple(t.mines, t.step+1)
	// if t.state == int32(pb.STATE_LEAD) {
	// 	safePits := t.minesSafePits()
	// 	room.SafePits = safePits
	// 	if t.step > 0 {
	// 		room.Score = int64(float64(t.betNum) * str2float(getStepMultiple(t.mines, t.step)))
	// 	}
	// 	// 1timer=100ms 剩余强制结算时间
	// 	_, restTimer := t.getStepForceSettlementTimer()
	// 	if restTimer <= 100 {
	// 		room.ForceSettleTime = time.Now().Unix() + int64(restTimer/10)
	// 	}
	// 	room.AutoMines = t.autoMines
	// }

	// registArea := 0
	// if t.role != nil {
	// 	registArea = t.role.RegistArea
	// }

	// mines := table.GetTables().MinesTable.Get(1)
	// 筹码
	// for i, chip := range mines.Chips[registArea].Value {
	// 	room.Chip = append(room.Chip, uint32(chip))
	// 	if chip == mines.ChipDefault[registArea] {
	// 		room.ChipSeat = int32(i)
	// 	}
	// }
	// room.BetDefault = int64(mines.BetDefault[registArea])
	// room.BetLimit = []int64{
	// 	mines.BetLimit[registArea].Value[0],
	// 	mines.BetLimit[registArea].Value[1],
	// }
	// room.WinLimit = mines.WinLimit[registArea]
	// room.StepTime = mines.StepTime[registArea]
	// room.AutoBetRoundDefault = mines.AutoBetRoundDefault[registArea]

	// 雷数与步数对应返奖倍数
	// steps := table.GetTables().MinesStepTable.GetDataList()
	// for _, step := range steps {
	// 	rule := &pb.FortuneGems2StepRule{Mines: step.Mines}
	// 	var i int32 = 0
	// 	for ; i < 25; i++ {
	// 		v := getStepMultiple(step.Mines, i)
	// 		if v == "" {
	// 			break
	// 		}
	// 		rule.Steps = append(rule.Steps, v)
	// 	}
	// 	room.MinesStepRules = append(room.MinesStepRules, rule)
	// }
}

func (t *Desk) fortuneGems2UserDataMsg() (msg *pb.FortuneGems2RoomUser) {
	msg = new(pb.FortuneGems2RoomUser)
	if v := t.role; v != nil {
		msg.Userid = v.Userid
		msg.Nickname = v.Nickname
		msg.Phone = v.Phone
		msg.Sex = v.Sex
		msg.Photo = v.Photo
		msg.Coin = v.Coin
		msg.Diamond = v.Diamond
		msg.VipLv = int32(v.Vip.Lv)
		msg.Offline = v.Offline
	}
	return
}

// 能否下注
func (t *Desk) canBet(user *data.User, num int64) bool {
	if user.Diamond < num {
		// 彩金不够下注
		glog.Errorf("no enough coin d:%d,c:%d, num:%d", user.Diamond, num)
		return false
	}
	return true
}

// 对局详情
func (t *Desk) gameDetail(force bool) {
	if t.role == nil {
		return
	}
	t.detail.EndTime = utils.BsonNow().Unix()
	t.detail.Players = t.role.Userid

	detail := t.detail.MinesDetail
	detail.UserId = t.role.Userid
	detail.Score = t.settleScore
	detail.AfterScore = t.role.GetScore()
	detail.AfterCash = t.role.Diamond
	if t.settleScore > 0 {
		detail.BeforeScore = t.role.GetScore() - t.settleScore + t.betNum
		detail.BeforeCash = t.role.Diamond - t.settleScore + t.betNum
	} else {
		detail.BeforeScore = t.role.GetScore() + t.betNum
		detail.BeforeCash = t.role.Diamond + t.betNum
	}
	detail.Mines = t.mines
	detail.MinesPits = t.minesPitsUser
	detail.Step = t.step
	detail.StepPits = t.stepPits
	multipleStep := t.step
	if t.settleScore <= 0 {
		multipleStep = t.step + 1 // 输局记录止步倍数
	}
	detail.Multiple = str2float(getStepMultiple(t.mines, multipleStep))
	detail.AutoMines = t.autoMines
	detail.Bets = t.betNum
	detail.WinType = 1
	if force && t.step == 0 {
		detail.WinType = 3
	} else if t.settleScore < 0 {
		detail.WinType = 2
	}
	detail.Force = force
	detail.ForceCashOut = t.forceCashOut

	body, err := json.Marshal(t.detail)
	if err != nil {
		glog.Errorf("save detail error %#v", t.detail)
		return
	}
	msg := &pb.Detail{Data: body}
	myactor.Logger().Tell(msg)
}
