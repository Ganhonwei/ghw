package handler

import (
	"bytes"
	"fmt"
	"goserver/pkg/data"
	"goserver/pkg/game/config"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/dhushon/decxls"
)

type PLANE1 struct {
	Gtype         int32   `xls:"玩法类型"`
	Status        int     `xls:"桌子开关（0=关，1=开）"`
	Ai_Status     int     `xls:"桌子ai开关（0=关，1=开）"`
	Algo_Switch   int     `xls:"算法（1=old，2=new）"`
	Count         uint32  `xls:"人数上限"`
	ChipLimit     uint32s `xls:"筹码额度"`
	RobotChipPro  int32s  `xls:"人机选筹码概率"`
	BetLimited    int32   `xls:"下注上限"`
	BetInterval   int32ss `xls:"总下注区间（-1=无限制）"`
	BetsRoom      int32ss `xls:"总下注对应房间"`
	NoviceRoom    int32s  `xls:"新手房间"`
	ExceptionRoom int32s  `xls:"异常房间"`
	ControlRoom   int32s  `xls:"点控房间"`
	BetTime       int32   `xls:"下注时间"`
	Msg_Score     int     `xls:"跑马灯显示分"`
	Jackpot       int32s  `xls:"Jackpot奖励倍数"`
	LoseFlow      int32s  `xls:"可赔付分数浮动区间"`
	FlyMultiple   int32ss `xls:"无人下注飞行倍数区间"`
	BreakingPrice int     `xls:"破产礼包价格"`
	BreakingGive  int     `xls:"破产礼包赠送"`
	BreakingTimes int     `xls:"破产礼包每日次数"`
	BackRate      float64 `xls:"返奖率"`
}

// 2.库存配置表
type PLANE2 struct {
	Id                     string `xls:"房间ID"`
	Name                   string `xls:"房间名"`
	DType                  int32  `xls:"房间类型"`
	Stock_Expect           int64  `xls:"库存期望"`
	Stock_Alarm            int    `xls:"库存报警"`
	MingTax                int32  `xls:"明税（万分比）"`
	AnTax                  int32  `xls:"暗税（万分比）"`
	FactorInterval         int32s `xls:"最终系数区间"`
	CoefficientFix         int32s `xls:"系数修正（9档，万分比）"`
	MaxLoseScore           int32  `xls:"单局平台输分上限"`
	BoomPro1               int32  `xls:"区间1爆炸概率（万分比）"`
	BoomPro2               int32  `xls:"区间2爆炸概率（万分比）"`
	BoomPro3               int32  `xls:"区间3爆炸概率（万分比）"`
	BoomPro4               int32  `xls:"区间4爆炸概率（万分比）"`
	NewbiewMaxWithdrawable int32s `xls:"新手可提现彩金上限AB"`
	InstantBangPro         int32  `xls:"秒爆概率"` //秒爆
}

// 3.人机策略
type PLANE3 struct {
	Num           int32  `xls:"参与人机"`
	InitScore     int32s `xls:"初始携带分数"`
	BetTime       int32s `xls:"下注时间区间"`
	MinBet        int32  `xls:"最小下注基数"`
	BetInterval   int32s `xls:"下注区间"`
	BetPro        int32s `xls:"下注概率0:不下 1:下"`
	LeaveLimit    int32s `xls:"人机携带退出下上限"`
	LeaveBetTimes int32s `xls:"人机下注退出下上限"`
	RobotPlanePro int32s `xls:"人机逃离概率（万分比）"`
}

// 1.扶摇直上策略
type PLANEStrategy1 struct {
	Id         int      `xls:"策略id"`
	O          int      `xls:"策略开关（o）"`
	Weight     int      `xls:"优先级"`  // 优先级
	Mutex      ints     `xls:"互斥策略"` // 互斥策略
	UserType   ints     `xls:"适用玩家账号类型（a，b，c）"`
	ChargeType ints     `xls:"适用玩家类型（新手，平民，普充，小r，中r，大r，超大r）"`
	X          int64    `xls:"未付费玩家默充金额（x）"`
	Q          float64  `xls:"单点爆率修正值（q）"`
	C          int64s   `xls:"冷却线（c）"`
	U          ints     `xls:"有效触发次数上限（u）"`
	UZ         ints     `xls:"总触发次数上限（u总）"`
	H          float64s `xls:"不爆倍数上限（h）"`
	M          int64s   `xls:"携带金额上限（m）"`
	RR         float64s `xls:"风控返奖率（rr）"`
	PR         int64s   `xls:"风控赢钱上限（pr）"`
}

// 2.欲薅无门策略
type PLANEStrategy2 struct {
	Id         int      `xls:"策略id"`
	O          int      `xls:"策略开关（o）"`
	Weight     int      `xls:"优先级"`  // 优先级
	Mutex      ints     `xls:"互斥策略"` // 互斥策略
	UserType   ints     `xls:"适用玩家账号类型（a，b，c）"`
	ChargeType ints     `xls:"适用玩家类型（新手，平民，普充，小r，中r，大r，超大r）"`
	T          ints     `xls:"生效次数上限（t）"`
	R          ints     `xls:"监控局数（r）"`
	A          float64s `xls:"逃跑平均倍数警戒线（a）"`
	P          int32s   `xls:"瞬爆概率（p）"`
	X          float64s `xls:"收割倍率（x）"`
	M          float64s `xls:"逃跑中位数警戒线（m）"`
}

// 3.虚假情报策略
type PLANEStrategy3 struct {
	Id         int      `xls:"策略id"`
	O          int      `xls:"策略开关（o）"`
	Weight     int      `xls:"优先级"`  // 优先级
	Mutex      ints     `xls:"互斥策略"` // 互斥策略
	UserType   ints     `xls:"适用玩家账号类型（a，b，c）"`
	ChargeType ints     `xls:"适用玩家类型（新手，平民，普充，小r，中r，大r，超大r）"`
	P0         float64  `xls:"录单返奖率（p0）"`
	P1         float64s `xls:"观察返奖率（p1）"`
	P2         float64s `xls:"逃后返奖率（p2）"`
}

// 4.起死回生策略
type PLANEStrategy4 struct {
	Id         int      `xls:"策略id"`
	O          int      `xls:"策略开关（o）"`
	Weight     int      `xls:"优先级"`  // 优先级
	Mutex      ints     `xls:"互斥策略"` // 互斥策略
	UserType   ints     `xls:"适用玩家账号类型（a，b，c）"`
	ChargeType ints     `xls:"适用玩家类型（新手，平民，普充，小r，中r，大r，超大r）"`
	M          float64s `xls:"援助返奖点m"`     // 援助返奖点m
	D          int64s   `xls:"allin特征d"`   // allin特征d
	N          float64s `xls:"期望返奖点n"`     // 期望返奖点n
	A          float64s `xls:"期高返奖点o"`     // 期高返奖点o
	Q          float64  `xls:"单点爆率修正值（q）"` // 单点爆率修正值
	S          float64s `xls:"不爆倍数上限（s）"`  // 不爆倍数上限（s）
	R          ints     `xls:"监控局数（r）"`
	RR         float64s `xls:"风控返奖率（rr）"`   // 风控返奖率（rr）
	PF         int64s   `xls:"风控赢钱上限（pf）"`  // 风控赢钱上限（pf）
	T          ints     `xls:"固定触发次数上限（t）"` // 固定触发次数上限（t）
	P          int32s   `xls:"随机拯救概率（p）"`   // 随机拯救率
	X          int64    `xls:"未付费玩家默充金额（x）"`
}

// 5.奖池风控策略
type PLANEStrategy5 struct {
	Id         int      `xls:"策略id"`
	O          int      `xls:"策略开关（o）"`
	Weight     int      `xls:"优先级"`  // 优先级
	Mutex      ints     `xls:"互斥策略"` // 互斥策略
	UserType   ints     `xls:"适用玩家账号类型（a，b，c）"`
	ChargeType ints     `xls:"适用玩家类型（新手，平民，普充，小r，中r，大r，超大r）"`
	M          float64s `xls:"赢家上限（m）"`      // 赢家上限（m）
	N          int64    `xls:"未付费玩家默充金额（x）"` // 未付费玩家默充金额（x）
	B          float64s `xls:"警戒倍数（b）"`      // 警戒倍数（b）
	Q          float64  `xls:"单点爆率修正值（q）"`   // 单点爆率修正值（q）
	T          ints     `xls:"自由空间（t）"`      // 自由空间（t）
}

// 6.冒险奖励策略
type PLANEStrategy6 struct {
	Id         int      `xls:"策略id"`
	O          int      `xls:"策略开关（o）"`
	Weight     int      `xls:"优先级"`  // 优先级
	Mutex      ints     `xls:"互斥策略"` // 互斥策略
	UserType   ints     `xls:"适用玩家账号类型（a，b，c）"`
	ChargeType ints     `xls:"适用玩家类型（新手，平民，普充，小r，中r，大r，超大r）"`
	F          float64s `xls:"奖励点（f）"`       // 奖励点（f)
	R          ints     `xls:"码局监控（r）"`      // 码局监控（r）
	C          int64s   `xls:"码量监控（c）"`      // 码量监控（c）
	W          ints     `xls:"博局监控（w）"`      // 博局监控（w）
	M          float64s `xls:"博倍监控（m）"`      // 博倍监控（m）
	N          int64s   `xls:"当局码量（n）"`      // 当局码量（n）
	X          int64    `xls:"未付费玩家默充金额（x）"` // 未付费玩家默充金额（x）
	S          float64s `xls:"不爆倍数上限（s）"`    // 不爆倍数上限（s）
	Q          float64  `xls:"单点爆率修正值（q）"`   // 单点爆率修正值（q）
	RR         float64s `xls:"风控返奖率（rr）"`    // 风控返奖率（rr）
	PF         int64s   `xls:"风控赢钱上限（pf）"`   // 风控赢钱上限（pf）
	T          ints     `xls:"固定触发次数上限（t）"`  // 固定触发次数上限（t）
	P          int32s   `xls:"随机拯救概率（p）"`    // 随机拯救概率（p）
}

// 7.人狂有祸策略
type PLANEStrategy7 struct {
	Id         int      `xls:"策略id"`
	O          int      `xls:"策略开关（o）"`
	Weight     int      `xls:"优先级"`  // 优先级
	Mutex      ints     `xls:"互斥策略"` // 互斥策略
	UserType   ints     `xls:"适用玩家账号类型（a，b，c）"`
	ChargeType ints     `xls:"适用玩家类型（新手，平民，普充，小r，中r，大r，超大r）"`
	PU         float64s `xls:"天罚线（pu）"`
	P7         float64s `xls:"天罚返奖率（p7）"`
	P1         int32s   `xls:"瞬爆概率（p1）"`
	W          int64s   `xls:"赢局监控（w）"`
	C          int64s   `xls:"肥码上限（c）"`
	P2         int32s   `xls:"天罚概率（p2）"`
	N          float64s `xls:"罚线调控（n）"`
	X          int64    `xls:"未付费玩家默充金额（x）"`
}

// 8.高潮涌现策略
type PLANEStrategy8 struct {
	Id         int      `xls:"策略id"`
	O          int      `xls:"策略开关（o）"`
	Weight     int      `xls:"优先级"`  // 优先级
	Mutex      ints     `xls:"互斥策略"` // 互斥策略
	UserType   ints     `xls:"适用玩家账号类型（a，b，c）"`
	ChargeType ints     `xls:"适用玩家类型（新手，平民，普充，小r，中r，大r，超大r）"`
	X          int64    `xls:"未付费玩家默充金额（x）"`
	HP         int32s   `xls:"基础高潮率（hp）"`
	HP1        int32s   `xls:"递增高潮率（hp1）"`
	FP         float64s `xls:"禁止高潮返奖率（fp）"`
	H          int32s   `xls:"不爆倍数上限（h）"`
	P0         float64s `xls:"潮后返奖率（p0）"`
	G          int32s   `xls:"高潮盈利倍率上限（g）"`
	F          int32s   `xls:"前戏局局数（f）"`
	S          int32s   `xls:"单日高潮次数上限（s）"`
	C          int32s   `xls:"贤者局数（c）"`
}

func parsePlane1(f *excelize.File) (ret []PLANE1, err error) {
	sheet := "1.房间基础配置表"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("PLANE配置表解析错误: %s", sheet)
		return
	}
	return
}

func parsePlane2(f *excelize.File) (ret []PLANE2, err error) {
	sheet := "2.库存配置表"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("PLANE配置表解析错误: %s", sheet)
		return
	}
	return
}

func parsePlane3(f *excelize.File) (ret []PLANE3, err error) {
	sheet := "3.人机策略"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("PLANE配置表解析错误: %s", sheet)
		return
	}
	return
}

// 更新crash配置
func updatePLANE(d []byte, save bool) error {
	byte_reader := bytes.NewReader(d)

	f, err := excelize.OpenReader(byte_reader)
	if err != nil {
		return err
	}

	c1, err := parsePlane1(f)
	if err != nil {
		return err
	}
	c2, err := parsePlane2(f)
	if err != nil {
		return err
	}
	c3, err := parsePlane3(f)
	if err != nil {
		return err
	}

	var games []data.Game
	base := c1[0]
	robot := c3[0]
	for _, v1 := range c2 {
		games = append(games, data.Game{
			Id:            v1.Id,
			Name:          v1.Name,
			Gtype:         base.Gtype,
			Status:        base.Status,
			Ai_Status:     base.Ai_Status,
			Algo_Switch:   base.Algo_Switch,
			Count:         base.Count,
			Dtype:         v1.DType,
			Stock_Expect:  v1.Stock_Expect,
			Stock_Alarm:   v1.Stock_Alarm,
			BreakingPrice: base.BreakingPrice,
			BreakingGive:  base.BreakingGive,
			BreakingTimes: base.BreakingTimes,
			PLANE: data.PLANEGame{
				ChipLimit:        base.ChipLimit.Value,
				RobotChipPro:     base.RobotChipPro.Value,
				BetLimit:         base.BetLimited,
				BetInterval:      base.BetInterval.Value,
				BetRoom:          base.BetsRoom.Value,
				NoviceRoom:       base.NoviceRoom.Value,
				ExceptionRoom:    base.ExceptionRoom.Value,
				ControlRoom:      base.ControlRoom.Value,
				FinalCoefficient: v1.FactorInterval.Value,
				CoefficientFix:   v1.CoefficientFix.Value,
				LoseLimited:      v1.MaxLoseScore,
				BetTime:          base.BetTime,
				MTax:             v1.MingTax,
				ATax:             v1.AnTax,
				BoomPro1:         v1.BoomPro1,
				BoomPro2:         v1.BoomPro2,
				BoomPro3:         v1.BoomPro3,
				BoomPro4:         v1.BoomPro4,
				BackRate:         base.BackRate,
				InstantBangPro:   v1.InstantBangPro,
				Jackpot:          base.Jackpot.Value,
				Msg_Score:        int64(base.Msg_Score),
				LoseFlow:         base.LoseFlow.Value,
				FlyMultiple:      base.FlyMultiple.Value,
				MaxWithdrawable:  v1.NewbiewMaxWithdrawable.Value,
				Robot: data.PLANEGameRobotStrategy{
					Num:        robot.Num,
					InitScore:  robot.InitScore.Value,
					TimeArea:   robot.BetTime.Value,
					Min:        robot.MinBet,
					BetArea:    robot.BetInterval.Value,
					BetPro:     robot.BetPro.Value,
					LeaveLimit: robot.LeaveLimit.Value,
					BetTimes:   robot.LeaveBetTimes.Value,
					RobotLeave: robot.RobotPlanePro.Value,
				},
			},
		})
	}

	for _, v := range games {
		if save {
			v.Save()
		}

		config.SetGame(v)
	}

	return nil
}

func parseAvStrategy1(f *excelize.File) (ret []PLANEStrategy1, err error) {
	sheet := "1.扶摇直上"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("AVIATOR配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseAvStrategy2(f *excelize.File) (ret []PLANEStrategy2, err error) {
	sheet := "2.欲薅无门"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("AVIATOR配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseAvStrategy3(f *excelize.File) (ret []PLANEStrategy3, err error) {
	sheet := "3.虚假情报"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("AVIATOR配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseAvStrategy4(f *excelize.File) (ret []PLANEStrategy4, err error) {
	sheet := "4.起死回生"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("AVIATOR配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseAvStrategy5(f *excelize.File) (ret []PLANEStrategy5, err error) {
	sheet := "5.奖池风控"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("AVIATOR配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseAvStrategy6(f *excelize.File) (ret []PLANEStrategy6, err error) {
	sheet := "6.冒险奖励"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("AVIATOR配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseAvStrategy7(f *excelize.File) (ret []PLANEStrategy7, err error) {
	sheet := "7.人狂有祸"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("AVIATOR配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseAvStrategy8(f *excelize.File) (ret []PLANEStrategy8, err error) {
	sheet := "8.高潮涌现"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("AVIATOR配置表解析错误: %s", sheet)
		return
	}
	return
}

func updateAvStrategy(save bool, f *excelize.File) error {
	c1, err := parseAvStrategy1(f)
	if err != nil {
		return err
	}

	c2, err := parseAvStrategy2(f)
	if err != nil {
		return err
	}

	c3, err := parseAvStrategy3(f)
	if err != nil {
		return err
	}

	c4, err := parseAvStrategy4(f)
	if err != nil {
		return err
	}

	c5, err := parseAvStrategy5(f)
	if err != nil {
		return err
	}

	c6, err := parseAvStrategy6(f)
	if err != nil {
		return err
	}

	c7, err := parseAvStrategy7(f)
	if err != nil {
		return err
	}

	c8, err := parseAvStrategy8(f)
	if err != nil {
		return err
	}

	c := data.AviatorStrategy{
		Id: 1,
		FYZS: data.PLANEFYZSStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         c1[0].Id,
				Weight:     c1[0].Weight,
				Mutex:      c1[0].Mutex.Value,
				O:          c1[0].O,
				UserType:   c1[0].UserType.Value,
				ChargeType: c1[0].ChargeType.Value,
			},
			X:  c1[0].X,
			Q:  c1[0].Q,
			U:  c1[0].U.Value,
			UZ: c1[0].UZ.Value,
			H:  c1[0].H.Value,
			M:  c1[0].M.Value,
			C:  c1[0].C.Value,
			RR: c1[0].RR.Value,
			PR: c1[0].PR.Value,
		},
		YHWM: data.PLANEYHWMStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         c2[0].Id,
				Weight:     c2[0].Weight,
				Mutex:      c2[0].Mutex.Value,
				O:          c2[0].O,
				UserType:   c2[0].UserType.Value,
				ChargeType: c2[0].ChargeType.Value,
			},
			T: c2[0].T.Value,
			R: c2[0].R.Value,
			M: c2[0].M.Value,
			A: c2[0].A.Value,
			P: c2[0].P.Value,
			X: c2[0].X.Value,
		},
		XJQB: data.PLANEXJQBStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         c3[0].Id,
				Weight:     c3[0].Weight,
				Mutex:      c3[0].Mutex.Value,
				O:          c3[0].O,
				UserType:   c3[0].UserType.Value,
				ChargeType: c3[0].ChargeType.Value,
			},
			P0: c3[0].P0,
			P1: c3[0].P1.Value,
			P2: c3[0].P2.Value,
		},
		QSHS: data.PLANEQSHSStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         c4[0].Id,
				Weight:     c4[0].Weight,
				Mutex:      c4[0].Mutex.Value,
				O:          c4[0].O,
				UserType:   c4[0].UserType.Value,
				ChargeType: c4[0].ChargeType.Value,
			},
			M:  c4[0].M.Value,
			D:  c4[0].D.Value,
			N:  c4[0].N.Value,
			A:  c4[0].A.Value,
			Q:  c4[0].Q,
			S:  c4[0].S.Value,
			R:  c4[0].R.Value,
			RR: c4[0].RR.Value,
			PF: c4[0].PF.Value,
			T:  c4[0].T.Value,
			P:  c4[0].P.Value,
		},
		JCFK: data.PLANEJCFKStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         c5[0].Id,
				Weight:     c5[0].Weight,
				Mutex:      c5[0].Mutex.Value,
				O:          c5[0].O,
				UserType:   c5[0].UserType.Value,
				ChargeType: c5[0].ChargeType.Value,
			},
			M: c5[0].M.Value,
			N: c5[0].N,
			B: c5[0].B.Value,
			Q: c5[0].Q,
			T: c5[0].T.Value,
		},
		MXJL: data.PLANEMXJLStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         c6[0].Id,
				Weight:     c6[0].Weight,
				Mutex:      c6[0].Mutex.Value,
				O:          c6[0].O,
				UserType:   c6[0].UserType.Value,
				ChargeType: c6[0].ChargeType.Value,
			},
			F:  c6[0].F.Value,
			R:  c6[0].R.Value,
			C:  c6[0].C.Value,
			W:  c6[0].W.Value,
			M:  c6[0].M.Value,
			N:  c6[0].N.Value,
			X:  c6[0].X,
			S:  c6[0].S.Value,
			Q:  c6[0].Q,
			RR: c6[0].RR.Value,
			PF: c6[0].PF.Value,
			T:  c6[0].T.Value,
			P:  c6[0].P.Value,
		},
		RKYH: data.PLANERKYHStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         c7[0].Id,
				Weight:     c7[0].Weight,
				Mutex:      c7[0].Mutex.Value,
				O:          c7[0].O,
				UserType:   c7[0].UserType.Value,
				ChargeType: c7[0].ChargeType.Value,
			},
			PU: c7[0].PU.Value,
			P7: c7[0].P7.Value,
			P1: c7[0].P1.Value,
			W:  c7[0].W.Value,
			C:  c7[0].C.Value,
			P2: c7[0].P2.Value,
			N:  c7[0].N.Value,
			X:  c7[0].X,
		},
		GCYX: data.PLANEGCYXStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         c8[0].Id,
				Weight:     c8[0].Weight,
				Mutex:      c8[0].Mutex.Value,
				O:          c8[0].O,
				UserType:   c8[0].UserType.Value,
				ChargeType: c8[0].ChargeType.Value,
			},
			X:   c8[0].X,
			HP:  c8[0].HP.Value,
			HP1: c8[0].HP1.Value,
			FP:  c8[0].FP.Value,
			H:   c8[0].H.Value,
			P0:  c8[0].P0.Value,
			G:   c8[0].G.Value,
			F:   c8[0].F.Value,
			S:   c8[0].S.Value,
			C:   c8[0].C.Value,
		},
	}

	config.SetAviatorStrategy(c)
	if save {
		c.Save()
	}

	return nil
}
