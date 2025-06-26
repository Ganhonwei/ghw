package handler

import (
	"bytes"
	"fmt"
	"goserver/pkg/data"
	"goserver/pkg/game/config"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/dhushon/decxls"
)

// 1.房间基础配置表
type UP1 struct {
	Gtype         int32   `xls:"玩法类型"`
	Status        int     `xls:"桌子开关（0=关，1=开）"`
	Ai_Status     int     `xls:"桌子ai开关（0=关，1=开）"`
	Algo_Switch   int     `xls:"算法（1=old，2=new）"`
	Count         uint32  `xls:"人数上限"`
	ChipLimit     uint32s `xls:"筹码额度"`
	BetLimited    int32s  `xls:"下注上限（小，大，7）"`
	BetInterval   int32ss `xls:"总下注区间（-1=无限制）"`
	BetsRoom      int32ss `xls:"总下注对应房间"`
	NoviceRoom    int32s  `xls:"新手房间"`
	ExceptionRoom int32s  `xls:"异常房间"`
	ControlRoom   int32s  `xls:"点控房间"`
	BetTime       int32   `xls:"下注时间"`
	Msg_Score     int     `xls:"跑马灯显示分"`
	BreakingPrice int     `xls:"破产礼包价格"`
	BreakingGive  int     `xls:"破产礼包赠送"`
	BreakingTimes int     `xls:"破产礼包每日次数"`
}

// 2.库存配置表
type UP2 struct {
	Id              string `xls:"房间ID"`
	Name            string `xls:"房间名"`
	Dtype           int32  `xls:"房间类型"`
	Stock_Expect    int64  `xls:"库存期望"`
	Stock_Alarm     int    `xls:"库存报警"`
	MingTax         int32  `xls:"明税（万分比）"`
	AnTax           int32  `xls:"暗税（万分比）"`
	FactorInterval  int32s `xls:"最终系数区间"`
	WinningInterval int32s `xls:"系数对应平台胜率（万分比）"`
	MaxLoseScore    int32  `xls:"单局平台输分上限"`
	CheatTimes      int32  `xls:"防刷水局数"`
	CheatTie        int32  `xls:"刷水开和概率（万分比）"`
	NewibewId       int    `xls:"B类新手配置"`
	ANewibewId      int    `xls:"A类新手配置"`
}

// 3.人机策略
type UP3 struct {
	Num           int32  `xls:"参与人机"`
	InitScore     int32s `xls:"初始携带分数"`
	BetWeight     int32s `xls:"下注权重（小，大，7）"`
	BetTime       int32s `xls:"下注时间区间"`
	MinBet        int32  `xls:"最小下注基数"`
	BetInterval   int32s `xls:"下注区间"`
	BetPro        int32s `xls:"下注概率0:不下 1:下"`
	LeaveLimit    int32s `xls:"人机携带退出下上限"`
	LeaveBetTimes int32s `xls:"人机下注退出下上限"`
}

// 4.新手策略
type UP4 struct {
	Id              int    `xls:"ID"`
	OutCashInterval int32s `xls:"新手可提彩金区间"`
	Winning         int32s `xls:"对应胜率(万分比）"`
	OutCashLimited  int32  `xls:"新手可提现彩金上限"`
}

// 1.心想事成
type UPStrategy1 struct {
	Id         int    `xls:"策略id"`
	O          int    `xls:"策略开关（o）"`
	Weight     int    `xls:"优先级"`  // 优先级
	Mutex      ints   `xls:"互斥策略"` // 互斥策略
	UserType   ints   `xls:"适用玩家账号类型（a，b，c）"`
	ChargeType ints   `xls:"适用玩家类型（新手，平民，普充，小r，中r，大r，超大r）"`
	M          int64s `xls:"援助返奖点（m）"`
	// A          ints   `xls:"怀疑警戒（a）"`
	// D          int64s `xls:"allin特征（d）"`
	// B          ints   `xls:"扰动冷却（b）"`
	// S          int32s `xls:"赢倍上限（s）"`
	UZ ints   `xls:"总触发次数上限（u总）"`
	U  ints   `xls:"有效触发次数上限（u）"`
	C  int64s `xls:"冷却线（c）"`
}

// 2.求死不能
type UPStrategy2 struct {
	Id         int      `xls:"策略id"`
	O          int      `xls:"策略开关（o）"`
	Weight     int      `xls:"优先级"`  // 优先级
	Mutex      ints     `xls:"互斥策略"` // 互斥策略
	UserType   ints     `xls:"适用玩家账号类型（a，b，c）"`
	ChargeType ints     `xls:"适用玩家类型（新手，平民，普充，小r，中r，大r，超大r）"`
	D          int64s   `xls:"allin特征（d）"`
	M          float64s `xls:"援助返奖点（m）"`
	RR         float64s `xls:"风控返奖率（rr）"`
	PR         int64s   `xls:"风控赢钱上限（pr）"`
	T          ints     `xls:"固定触发次数上限（t）"`
	P          int32s   `xls:"随机拯救概率（p）"`
}

// 3.7管严
type UPStrategy3 struct {
	Id         int      `xls:"策略id"`
	O          int      `xls:"策略开关（o）"`
	Weight     int      `xls:"优先级"`  // 优先级
	Mutex      ints     `xls:"互斥策略"` // 互斥策略
	UserType   ints     `xls:"适用玩家账号类型（a，b，c）"`
	ChargeType ints     `xls:"适用玩家类型（新手，平民，普充，小r，中r，大r，超大r）"`
	X          int64    `xls:"未付费玩家默充金额（x）"`
	PU         float64s `xls:"天罚线（pu）"`
	PP         int64s   `xls:"天罚利（pp）"`
	P3         int32s   `xls:"压制概率（p3）"`
	P4         int32s   `xls:"倍杀概率（p4）"`
	N          ints     `xls:"倍杀轮回（n）"`
	C          int64s   `xls:"倍杀冷却（c）"`
	T          ints     `xls:"单日倍杀次数上限（t日）"`
	TZ         ints     `xls:"倍杀次数上限（t总）"`
}

// 4.高潮涌现
type UPStrategy4 struct {
	Id         int      `xls:"策略id"`
	O          int      `xls:"策略开关（o）"`
	Weight     int      `xls:"优先级"`  // 优先级
	Mutex      ints     `xls:"互斥策略"` // 互斥策略
	UserType   ints     `xls:"适用玩家账号类型（a，b，c）"`
	ChargeType ints     `xls:"适用玩家类型（新手，平民，普充，小r，中r，大r，超大r）"`
	X          int64    `xls:"未付费玩家默充金额（x）"`
	HP         ints     `xls:"基础高潮率（hp）"`
	HP1        ints     `xls:"递增高潮率（hp1）"`
	FP         float64s `xls:"禁止高潮返奖率（fp）"`
	H          ints     `xls:"不爆倍数上限（h）"`
	P0         float64s `xls:"潮后返奖率（p0）"`
	G          ints     `xls:"高潮盈利倍率上限（g）"`
	F          ints     `xls:"前戏局局数（f）"`
	S          ints     `xls:"单日高潮次数上限（s）"`
	C          ints     `xls:"贤者局数（c）"`
}

func parseUPStrategy1(f *excelize.File) (ret []UPStrategy1, err error) {
	sheet := "1.心想事成"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("7UP配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseUPStrategy2(f *excelize.File) (ret []UPStrategy2, err error) {
	sheet := "2.求死不能"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("7UP配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseUPStrategy3(f *excelize.File) (ret []UPStrategy3, err error) {
	sheet := "3.7管严"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("7UP配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseUPStrategy4(f *excelize.File) (ret []UPStrategy4, err error) {
	sheet := "4.高潮涌现"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("LHD配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseUP1(f *excelize.File) (ret []UP1, err error) {
	sheet := "1.房间基础配置表"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("7up配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseUP2(f *excelize.File) (ret []UP2, err error) {
	sheet := "2.库存配置表"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("7up配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseUP3(f *excelize.File) (ret []UP3, err error) {
	sheet := "3.人机策略"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("7up配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseUP4(f *excelize.File) (ret []LHD4, err error) {
	sheet := "4.新手配置"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("LH配置表解析错误: %s", sheet)
		return
	}
	return
}

// 更新7up配置
func updateUP(d []byte, save bool) error {
	byte_reader := bytes.NewReader(d)

	f, err := excelize.OpenReader(byte_reader)
	if err != nil {
		return err
	}

	up1, err := parseUP1(f)
	if err != nil {
		return err
	}
	up2, err := parseUP2(f)
	if err != nil {
		return err
	}
	up3, err := parseUP3(f)
	if err != nil {
		return err
	}
	up4, err := parseUP4(f)
	if err != nil {
		return err
	}

	newbiews := make(map[int]data.FreeNewbiewMode)
	for _, l := range up4 {
		newbiews[l.Id] = data.FreeNewbiewMode{
			OutCashInterval: l.OutCashInterval.Value,
			Winning:         l.Winning.Value,
			OutCashLimited:  l.OutCashLimited,
		}
	}

	var games []data.Game
	base := up1[0]
	robot := up3[0]
	for _, v1 := range up2 {
		games = append(games, data.Game{
			Id:            v1.Id,
			Name:          v1.Name,
			Gtype:         base.Gtype,
			Dtype:         v1.Dtype,
			Status:        base.Status,
			Ai_Status:     base.Ai_Status,
			Algo_Switch:   base.Algo_Switch,
			Count:         base.Count,
			Stock_Expect:  v1.Stock_Expect,
			Stock_Alarm:   v1.Stock_Alarm,
			BreakingPrice: base.BreakingPrice,
			BreakingGive:  base.BreakingGive,
			BreakingTimes: base.BreakingTimes,
			UP: data.LHDGame{
				ChipLimit:        base.ChipLimit.Value,
				BetLimit:         base.BetLimited.Value,
				BetInterval:      base.BetInterval.Value,
				BetRoom:          base.BetsRoom.Value,
				NoviceRoom:       base.NoviceRoom.Value,
				ExceptionRoom:    base.ExceptionRoom.Value,
				ControlRoom:      base.ControlRoom.Value,
				FinalCoefficient: v1.FactorInterval.Value,
				Winning:          v1.WinningInterval.Value,
				LoseLimited:      v1.MaxLoseScore,
				Cheat:            v1.CheatTimes,
				CheatTie:         v1.CheatTie,
				LHMTax:           v1.MingTax,
				LHATax:           v1.AnTax,
				LHBetTime:        base.BetTime,
				Msg_Score:        int64(base.Msg_Score),
				NewbiewMode:      newbiews[v1.NewibewId],
				ANewbiewMode:     newbiews[v1.ANewibewId],
				Robot: data.LHDGameRobotStrategy{
					Num:        robot.Num,
					InitScore:  robot.InitScore.Value,
					BetWeight:  robot.BetWeight.Value,
					TimeArea:   robot.BetTime.Value,
					Min:        robot.MinBet,
					BetArea:    robot.BetInterval.Value,
					BetPro:     robot.BetPro.Value,
					LeaveLimit: robot.LeaveLimit.Value,
					BetTimes:   robot.LeaveBetTimes.Value,
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

// 更新up配置
func updateUPStrategy(save bool, f *excelize.File) error {
	up1s, err := parseUPStrategy1(f)
	if err != nil {
		return err
	}

	up2s, err := parseUPStrategy2(f)
	if err != nil {
		return err
	}

	up3s, err := parseUPStrategy3(f)
	if err != nil {
		return err
	}

	up4s, err := parseUPStrategy4(f)
	if err != nil {
		return err
	}

	up1 := up1s[0]
	up2 := up2s[0]
	up3 := up3s[0]
	up4 := up4s[0]

	c := data.SevenStrategy{
		Id: 1,
		XXSC: data.LHDXXSCStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         up1.Id,
				Weight:     up1.Weight,
				Mutex:      up1.Mutex.Value,
				O:          up1.O,
				UserType:   up1.UserType.Value,
				ChargeType: up1.ChargeType.Value,
			},
			M: up1.M.Value,
			// A:  up1.A.Value,
			// D:  up1.D.Value,
			// B:  up1.B.Value,
			// S:  up1.S.Value,
			U:  up1.U.Value,
			UZ: up1.UZ.Value,
			C:  up1.C.Value,
		},
		QSBN: data.LHDQSBNStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         up2.Id,
				Weight:     up2.Weight,
				Mutex:      up2.Mutex.Value,
				O:          up2.O,
				UserType:   up2.UserType.Value,
				ChargeType: up2.ChargeType.Value,
			},
			M:  up2.M.Value,
			RR: up2.RR.Value,
			D:  up2.D.Value,
			PR: up2.PR.Value,
			T:  up2.T.Value,
			P:  up2.P.Value,
		},
		QGY: data.LHDLKYHStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         up3.Id,
				Weight:     up3.Weight,
				Mutex:      up3.Mutex.Value,
				O:          up3.O,
				UserType:   up3.UserType.Value,
				ChargeType: up3.ChargeType.Value,
			},
			X:  up3.X,
			PU: up3.PU.Value,
			PP: up3.PP.Value,
			P3: up3.P3.Value,
			P4: up3.P4.Value,
			N:  up3.N.Value,
			C:  up3.C.Value,
			T:  up3.T.Value,
			TZ: up3.TZ.Value,
		},
		GCYX: data.LHDGCYXStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         up4.Id,
				Weight:     up4.Weight,
				Mutex:      up4.Mutex.Value,
				O:          up4.O,
				UserType:   up4.UserType.Value,
				ChargeType: up4.ChargeType.Value,
			},
			X:   up4.X,
			HP:  up4.HP.Value,
			HP1: up4.HP1.Value,
			FP:  up4.FP.Value,
			H:   up4.H.Value,
			P0:  up4.P0.Value,
			G:   up4.G.Value,
			F:   up4.F.Value,
			S:   up4.S.Value,
			C:   up4.C.Value,
		},
	}

	config.SetSevenStrategy(c)
	if save {
		c.Save()
	}

	return nil
}
