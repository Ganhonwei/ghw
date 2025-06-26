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
type LHD1 struct {
	Gtype         int32   `xls:"玩法类型"`
	Status        int     `xls:"桌子开关（0=关，1=开）"`
	Ai_Status     int     `xls:"桌子ai开关（0=关，1=开）"`
	Algo_Switch   int     `xls:"算法（1=old，2=new）"`
	Count         uint32  `xls:"人数上限"`
	ChipLimit     uint32s `xls:"筹码额度"`
	BetLimited    int32s  `xls:"下注上限（龙，虎，和）"`
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
type LHD2 struct {
	Id              string `xls:"房间ID"`
	Name            string `xls:"房间名"`
	DType           int32  `xls:"房间类型"`
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
type LHD3 struct {
	Num           int32  `xls:"参与人机"`
	InitScore     int32s `xls:"初始携带分数"`
	BetWeight     int32s `xls:"下注权重（龙，虎，和）"`
	BetTime       int32s `xls:"下注时间区间"`
	MinBet        int32  `xls:"最小下注基数"`
	BetInterval   int32s `xls:"下注区间"`
	BetPro        int32s `xls:"下注概率0:不下 1:下"`
	LeaveLimit    int32s `xls:"人机携带退出下上限"`
	LeaveBetTimes int32s `xls:"人机下注退出下上限"`
}

// 4.新手策略
type LHD4 struct {
	Id              int    `xls:"ID"`
	OutCashInterval int32s `xls:"新手可提彩金区间"`
	Winning         int32s `xls:"对应胜率(万分比）"`
	OutCashLimited  int32  `xls:"新手可提现彩金上限"`
}

// 1.心想事成
type LHDStrategy1 struct {
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
type LHDStrategy2 struct {
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

// 3.龙狂有祸
type LHDStrategy3 struct {
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
type LHDStrategy4 struct {
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

func parseLH1(f *excelize.File) (ret []LHD1, err error) {
	sheet := "1.房间基础配置表"
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

func parseLH2(f *excelize.File) (ret []LHD2, err error) {
	sheet := "2.库存配置表"
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

func parseLH3(f *excelize.File) (ret []LHD3, err error) {
	sheet := "3.人机策略"
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

func parseLH4(f *excelize.File) (ret []LHD4, err error) {
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

func parseLHDStrategy1(f *excelize.File) (ret []LHDStrategy1, err error) {
	sheet := "1.心想事成"
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

func parseLHDStrategy2(f *excelize.File) (ret []LHDStrategy2, err error) {
	sheet := "2.求死不能"
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

func parseLHDStrategy3(f *excelize.File) (ret []LHDStrategy3, err error) {
	sheet := "3.龙狂有祸"
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

func parseLHDStrategy4(f *excelize.File) (ret []LHDStrategy4, err error) {
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

// 更新lhd配置
func updateLHD(d []byte, save bool) error {
	byte_reader := bytes.NewReader(d)

	f, err := excelize.OpenReader(byte_reader)
	if err != nil {
		return err
	}

	lh1, err := parseLH1(f)
	if err != nil {
		return err
	}
	lh2, err := parseLH2(f)
	if err != nil {
		return err
	}
	lh3, err := parseLH3(f)
	if err != nil {
		return err
	}

	lh4, err := parseLH4(f)
	if err != nil {
		return err
	}

	newbiews := make(map[int]data.FreeNewbiewMode)
	for _, l := range lh4 {
		newbiews[l.Id] = data.FreeNewbiewMode{
			OutCashInterval: l.OutCashInterval.Value,
			Winning:         l.Winning.Value,
			OutCashLimited:  l.OutCashLimited,
		}
	}

	var games []data.Game
	base := lh1[0]
	robot := lh3[0]
	for _, v1 := range lh2 {
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
			LHD: data.LHDGame{
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

// 更新lhd配置
func updateLHDStrategy(save bool, f *excelize.File) error {
	lh1s, err := parseLHDStrategy1(f)
	if err != nil {
		return err
	}

	lh2s, err := parseLHDStrategy2(f)
	if err != nil {
		return err
	}

	lh3s, err := parseLHDStrategy3(f)
	if err != nil {
		return err
	}

	lh4s, err := parseLHDStrategy4(f)
	if err != nil {
		return err
	}

	lh1 := lh1s[0]
	lh2 := lh2s[0]
	lh3 := lh3s[0]
	lh4 := lh4s[0]

	c := data.LhdStrategy{
		Id: 1,
		XXSC: data.LHDXXSCStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         lh1.Id,
				Weight:     lh1.Weight,
				Mutex:      lh1.Mutex.Value,
				O:          lh1.O,
				UserType:   lh1.UserType.Value,
				ChargeType: lh1.ChargeType.Value,
			},
			M: lh1.M.Value,
			// A:  lh1.A.Value,
			// D:  lh1.D.Value,
			// B:  lh1.B.Value,
			// S:  lh1.S.Value,
			U:  lh1.U.Value,
			UZ: lh1.UZ.Value,
			C:  lh1.C.Value,
		},
		QSBN: data.LHDQSBNStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         lh2.Id,
				Weight:     lh2.Weight,
				Mutex:      lh2.Mutex.Value,
				O:          lh2.O,
				UserType:   lh2.UserType.Value,
				ChargeType: lh2.ChargeType.Value,
			},
			M:  lh2.M.Value,
			RR: lh2.RR.Value,
			D:  lh2.D.Value,
			PR: lh2.PR.Value,
			T:  lh2.T.Value,
			P:  lh2.P.Value,
		},
		LKYH: data.LHDLKYHStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         lh3.Id,
				Weight:     lh3.Weight,
				Mutex:      lh3.Mutex.Value,
				O:          lh3.O,
				UserType:   lh3.UserType.Value,
				ChargeType: lh3.ChargeType.Value,
			},
			X:  lh3.X,
			PU: lh3.PU.Value,
			PP: lh3.PP.Value,
			P3: lh3.P3.Value,
			P4: lh3.P4.Value,
			N:  lh3.N.Value,
			C:  lh3.C.Value,
			T:  lh3.T.Value,
			TZ: lh3.TZ.Value,
		},
		GCYX: data.LHDGCYXStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         lh4.Id,
				Weight:     lh4.Weight,
				Mutex:      lh4.Mutex.Value,
				O:          lh4.O,
				UserType:   lh4.UserType.Value,
				ChargeType: lh4.ChargeType.Value,
			},
			X:   lh4.X,
			HP:  lh4.HP.Value,
			HP1: lh4.HP1.Value,
			FP:  lh4.FP.Value,
			H:   lh4.H.Value,
			P0:  lh4.P0.Value,
			G:   lh4.G.Value,
			F:   lh4.F.Value,
			S:   lh4.S.Value,
			C:   lh4.C.Value,
		},
	}

	config.SetLHDStrategy(c)
	if save {
		c.Save()
	}

	return nil
}
