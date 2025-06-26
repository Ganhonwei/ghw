package mines

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"strconv"
	"time"
)

func (t *Desk) tick() {
	switch t.state {
	case int32(pb.STATE_FREE):
		t.timer++
		if t.timer >= 5 { // 0.5s
			t.changeDeskState(pb.STATE_READY)
		}
	case int32(pb.STATE_READY): // 空闲状态
		t.timer++
		if t.timer >= t.timerDuring {
			t.gameInit()
			t.changeDeskState(pb.STATE_BET)
		}
	case int32(pb.STATE_BET): // 下注状态
		t.timer++
		if t.autoRound > 0 {
			// 自动对局自动下注
			t.minesAutoBets(t.autoBets)
		} else {
			// 检查玩家离线关闭房间
			if t.timer >= 50 { // 5s
				t.timer = 0
				t.kickOffline()
			}
		}
	case int32(pb.STATE_LEAD): // 下注完成已埋雷
		if t.autoMines {
			// 自动对局自动踩雷
			t.minesAutoSteps()
		} else {
			// 强制结算检测
			t.timer++
			t.checkStepSettlementForce()
		}
	case int32(pb.STATE_OVER): // 结算状态
		t.timer++
		if t.timer >= t.timerDuring {
			t.changeDeskState(pb.STATE_READY) // 切换回准备状态
		}
	}
}

// 改变桌子状态
func (t *Desk) changeDeskState(state pb.DeskState) {
	t.state = int32(state)
	t.timer = 0
	switch state {
	default:
		t.timerDuring = 0
	case pb.STATE_READY: // 准备
		t.timerDuring = 5 // 0.5s
	case pb.STATE_OVER: // 结算
		mines := table.GetTables().MinesTable.Get(1)
		if t.autoMines { // 自动下注结算展示时间
			t.timerDuring = int(10 * mines.AutoBetSettleTime[t.role.RegistArea])
		} else { // 手动下注结算展示时间
			t.timerDuring = int(10 * mines.BetSettleTime[t.role.RegistArea])
		}
	}

	ntf := &pb.MinesPushStateNtf{State: state}
	ntf.During = int32(100 * t.timerDuring)
	t.send2user(ntf)
}

// 强制结算检测
func (t *Desk) checkStepSettlementForce() {
	// 1timer=100ms
	stepTimer, restTimer := t.getStepForceSettlementTimer()
	// 倒计时10s通知
	if restTimer == 100 {
		t.send2user(&pb.MinesCountDownNtf{ForceSettleTime: time.Now().Unix() + int64(restTimer/10)})
		return
	}
	// 强制结算
	if stepTimer <= t.timer {
		t.gameOver(true, false)
	}
}

// 1timer=100ms
func (t *Desk) getStepForceSettlementTimer() (stepTimer, restTimer int) {
	stepTimer = int(10 * table.GetTables().MinesTable.Get(1).StepTime[t.role.RegistArea])
	restTimer = stepTimer - t.timer
	return
}

// 是否游戏进行中
func (t *Desk) IsGaming() bool {
	return t.state == int32(pb.STATE_LEAD) || t.state == int32(pb.STATE_OVER)
}

// 埋雷
func (t *Desk) minesPitsInit() {
	pits := make([]*int32, len(t.minesPits))
	for i := range pits {
		pits[i] = &t.minesPits[i]
	}

	for range t.mines {
		pit := utils.RandIntN(len(pits))
		*pits[pit] = 1
		copy(pits[pit:], pits[pit+1:])
		pits = pits[:len(pits)-1]
	}

	glog.Infof("mines pit init: userid=%s, %#v", t.role.Userid, t.minesPits)
}

// mines 下注
func (t *Desk) minesBet(userid string, arg *pb.MinesBetReq) (rsp *pb.MinesBetRsp) {
	rsp = new(pb.MinesBetRsp)

	if t.state != int32(pb.STATE_BET) {
		rsp.Error = pb.BetOver
		return
	}

	if arg.Mines <= 0 || arg.Bets <= 0 {
		rsp.Error = pb.Failed
		return
	}

	user := t.getPlayer()
	if user == nil || user.Userid != userid {
		rsp.Error = pb.NotInRoom
		return
	}

	bets := int64(arg.Bets)
	if !t.canBet(user, bets) {
		rsp.Error = pb.NotEnoughCoin
		return
	}

	// 下注限制
	betLimit := table.GetTables().MinesTable.Get(1).BetLimit[user.RegistArea]
	if bets < int64(betLimit.Value[0]) || bets > int64(betLimit.Value[1]) {
		rsp.Error = pb.BetTopLimit
		return
	}

	// 自动下注
	if arg.AutoRound > 0 {
		// 检测自动下注坑位
		var autoPits []int32
		var autoPitsM = make(map[int32]bool)
		for _, pit := range arg.AutoPits {
			if pit < 0 || pit > 24 {
				glog.Errorf("autoRound pits error: %s, %v", user.Userid, arg)
				rsp.Error = pb.Failed
				return
			}
			if autoPitsM[pit] {
				continue
			}
			autoPitsM[pit] = true
			autoPits = append(autoPits, pit)
		}
		if len(autoPits) == 0 {
			glog.Errorf("autoRound but no pits: %s, %v", user.Userid, arg)
			rsp.Error = pb.Failed
			return
		}

		t.mines = arg.Mines
		t.autoBets = bets
		t.autoRound = arg.AutoRound
		t.autoPits = autoPits
		// wait tick autoBet

		rsp.Multiple = getStepMultiple(arg.Mines, int32(len(autoPits)))
		rsp.AutoRound = arg.AutoRound
		rsp.AutoPits = autoPits
		rsp.SafePits = int32(len(t.minesPits)) - t.mines - t.step
		return
	}

	t.sendCurrency(userid, 0, (-1 * bets), int32(pb.LOG_TYPE148), fmt.Sprintf("Mines房间%s下注", t.DeskData.Rid))
	t.betNum = bets

	t.mines = arg.Mines
	// 埋雷
	t.minesPitsInit()
	t.changeDeskState(pb.STATE_LEAD)

	rsp.Multiple = getStepMultiple(arg.Mines, 1)
	rsp.Bets = arg.Bets
	rsp.SafePits = int32(len(t.minesPits)) - t.mines - t.step
	return
}

// 自动下注
func (t *Desk) minesAutoBets(bets int64) {
	if t.role == nil {
		return
	}
	if !t.canBet(t.role.User, bets) {
		t.gameOver(false, true)
		return
	}

	t.autoRound--
	t.autoMines = true

	t.sendCurrency(t.role.Userid, 0, (-1 * bets), int32(pb.LOG_TYPE149), fmt.Sprintf("Mines房间%s自动下注", t.DeskData.Rid))
	t.betNum = bets

	var showMultiple string
	if t.autoMines {
		// 自动对局显示选择的步数倍数
		nextStep := len(t.autoPits) + 1
		maxSafeStep := len(t.minesPits) - int(t.mines)
		showMultiple = getStepMultiple(t.mines, int32(min(nextStep, maxSafeStep)))
	}
	msg := &pb.MinesAutoBetNtf{
		Bets:      int32(bets),
		AutoRound: t.autoRound,
		Multiple:  showMultiple,
		AutoPits:  t.autoPits,
	}
	t.send2user(msg)

	// 埋雷
	t.minesPitsInit()
	t.changeDeskState(pb.STATE_LEAD)
}

// 自动踩雷
func (t *Desk) minesAutoSteps() {
	for _, pit := range t.autoPits {
		if boom := t.minesPitStep(pit, true); boom {
			// 踩到雷了
			break
		}
	}
	t.gameOver(false, true)
}

// 随机选择一个没踩过的位置
func (t *Desk) minesRandomPitStep() (pit int32) {
	// 未选择的位置
	var pits []int32
	for pit, pitV := range t.minesPitsUser {
		if pitV == 0 {
			pits = append(pits, int32(pit))
		}
	}

	if len(pits) == 1 {
		return pits[0]
	}
	i := utils.RandIntN(len(pits))
	return pits[i]
}

// 踩雷
func (t *Desk) minesPitStep(pit int32, auto bool) (boom bool) {
	if t.state != int32(pb.STATE_LEAD) {
		return true
	}
	if !auto {
		t.timer = 0
	}

	t.stepPits = append(t.stepPits, pit)

	if v := t.minesPits[pit]; v != 0 {
		t.minesPitsUser[pit] = 2 // 2已踩有雷
		t.minesStepBoomPit = pit // 最后踩雷位置
		return true
	}

	// 没雷
	t.minesPitsUser[pit] = 1 // 1已踩无雷
	t.step++                 // 步数+1
	return false
}

// 计算剩余安全步数
func (t *Desk) minesSafePits() int32 {
	return int32(len(t.minesPits)) - t.mines - t.step
}

func str2float(str string) float64 {
	if str == "" {
		return 0
	}
	r2, _ := strconv.ParseFloat(str, 64)
	return r2
}

// getStepMultiple 获取指定步数返奖率
func getStepMultiple(mines, step int32) (m string) {
	defer func() {
		// 校验值
		if m == "" {
			return
		}
		v := str2float(m)
		if v == 0 {
			m = ""
			return
		}
		m = fmt.Sprintf("%.2f", v) // 保留两位小数,向下取整
	}()

	steps := table.GetTables().MinesStepTable.Get(mines)
	if steps == nil {
		return
	}
	switch step {
	default:
		return
	case 0:
		return steps.Step0
	case 1:
		return steps.Step1
	case 2:
		return steps.Step2
	case 3:
		return steps.Step3
	case 4:
		return steps.Step4
	case 5:
		return steps.Step5
	case 6:
		return steps.Step6
	case 7:
		return steps.Step7
	case 8:
		return steps.Step8
	case 9:
		return steps.Step9
	case 10:
		return steps.Step10
	case 11:
		return steps.Step11
	case 12:
		return steps.Step12
	case 13:
		return steps.Step13
	case 14:
		return steps.Step14
	case 15:
		return steps.Step15
	case 16:
		return steps.Step16
	case 17:
		return steps.Step17
	case 18:
		return steps.Step18
	case 19:
		return steps.Step19
	case 20:
		return steps.Step20
	case 21:
		return steps.Step21
	case 22:
		return steps.Step22
	case 23:
		return steps.Step23
	case 24:
		return steps.Step24
	}
}
