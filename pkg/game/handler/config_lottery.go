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
type Lottery1 struct {
	Gtype           int32    `xls:"玩法类型"`
	Status          int      `xls:"桌子开关（0=关，1=开）"`
	Ai_Status       int      `xls:"桌子ai开关（0=关，1=开）"`
	Algo_Switch     int      `xls:"算法（1=old，2=new）"`
	Count           uint32   `xls:"人数上限"`
	ChipLimit       uint32s  `xls:"筹码额度"`
	BetLimited      int32s   `xls:"下注上限"`
	BetInterval     int32ss  `xls:"总下注区间（-1=无限制）"`
	BetsRoom        int32ss  `xls:"总下注对应房间"`
	NoviceRoom      int32s   `xls:"新手房间"`
	ExceptionRoom   int32s   `xls:"异常房间"`
	ControlRoom     int32s   `xls:"点控房间"`
	BetTime         int32    `xls:"下注时间（秒）"`
	Msg_Score       int      `xls:"跑马灯显示分"`
	UnRandomRate    int32    `xls:"非随机开牌概率"`
	NewbieMaxWin    int64    `xls:"新手赢分上限"`
	NewbieMinWin    int64    `xls:"新手赢分下限"`
	NewbieGiveRate  int32    `xls:"新手赢分下限送分概率"`
	WinnabilityRate int32    `xls:"可赢系数"`
	JackpotExceed   int64    `xls:"JACKPOT超分必出开"`
	RTPFixRate      float64s `xls:"调整系数"`
}

// 2.库存配置表
type Lottery2 struct {
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
	DrawWeight      ints   `xls:"开奖权重（高牌，对子，同花，顺子，同花顺，豹子）"`
	NewibewId       int    `xls:"B类新手配置"`
	ANewibewId      int    `xls:"A类新手配置"`
}

// 3.人机策略
type Lottery3 struct {
	Num           int32  `xls:"参与人机"`
	InitScore     int32s `xls:"初始携带分数"`
	BetWeight     int32s `xls:"下注权重（高牌，对子，同花，顺子，同花顺，豹子）"`
	BetTime       int32s `xls:"下注时间区间"`
	MinBet        int32  `xls:"最小下注基数"`
	BetInterval   int32s `xls:"下注区间"`
	BetPro        int32s `xls:"下注概率"`
	LeaveLimit    int32s `xls:"人机携带退出下上限"`
	LeaveBetTimes int32s `xls:"人机下注退出下上限"`
}

// 4.新手策略
type Lottery4 struct {
	Id              int    `xls:"ID"`
	OutCashInterval int32s `xls:"新手可提彩金区间"`
	Winning         int32s `xls:"对应胜率(万分比）"`
	OutCashLimited  int32  `xls:"新手可提现彩金上限"`
}

// 5.杀富济贫
type Lottery5 struct {
	SFChargeFloor int64 `xls:"杀富充值下限"`
	SFFactorFloor int32 `xls:"杀富个人系数下限"`
	SFStockUpper  int64 `xls:"杀富库存上限"`
	SFTriggerPro  int32 `xls:"杀富触发概率"`
	JPChargeUpper int64 `xls:"济贫充值上限"`
	JPFactorUpper int32 `xls:"济贫个人系数上限"`
	JPStockFloor  int64 `xls:"济贫库存下限"`
	JPTriggerPro  int32 `xls:"济贫触发概率"`
}

// 2.来易去难
type CPStrategy2 struct {
	Id         int      `xls:"策略id"`
	O          int      `xls:"策略开关（o）"`
	Weight     int      `xls:"优先级"`  // 优先级
	Mutex      ints     `xls:"互斥策略"` // 互斥策略
	UserType   ints     `xls:"适用玩家账号类型（a，b，c）"`
	ChargeType ints     `xls:"适用玩家类型（新手，平民，普充，小r，中r，大r，超大r）"`
	X          int64    `xls:"未付费玩家默充金额（x）"`
	D          int64s   `xls:"allin特征（d）"`
	M          float64s `xls:"援助返奖点（m）"`
	RR         float64s `xls:"风控返奖率（rr）"`
	PR         int64s   `xls:"风控赢钱上限（pr）"`
	T          ints     `xls:"固定触发次数上限（t）"`
	PS         int32s   `xls:"随机拯救概率（ps）"`
}

// 1.来玩就赢
type CPStrategy1 struct {
	Id         int      `xls:"策略id"`
	O          int      `xls:"策略开关（o）"`
	Weight     int      `xls:"优先级"`  // 优先级
	Mutex      ints     `xls:"互斥策略"` // 互斥策略
	UserType   ints     `xls:"适用玩家账号类型（a，b，c）"`
	ChargeType ints     `xls:"适用玩家类型（新手，平民，普充，小r，中r，大r，超大r）"`
	X          int64    `xls:"未付费玩家默充金额（x）"`
	M          int64s   `xls:"援助返奖点（m）"`
	RR         float64s `xls:"风控返奖率（rr）"`
	PR         int64s   `xls:"风控赢钱上限（pr）"`
	C          int64s   `xls:"冷却线（c）"`
	U          ints     `xls:"有效触发次数上限（u）"`
	UZ         ints     `xls:"总次数上限（u总）"`
}

func parseLottery1(f *excelize.File) (ret []Lottery1, err error) {
	sheet := "1.房间基础配置表"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("Lottery配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseLottery2(f *excelize.File) (ret []Lottery2, err error) {
	sheet := "2.库存配置表"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("Lottery配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseLottery3(f *excelize.File) (ret []Lottery3, err error) {
	sheet := "3.人机策略"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("Lottery配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseLottery4(f *excelize.File) (ret []Lottery4, err error) {
	sheet := "4.新手配置"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("Lottery配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseLottery5(f *excelize.File) (ret []Lottery5, err error) {
	sheet := "5.杀富济贫"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("Lottery配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseCPStrategy1(f *excelize.File) (ret []CPStrategy1, err error) {
	sheet := "1.来玩就赢"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("AB配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseCPStrategy2(f *excelize.File) (ret []CPStrategy2, err error) {
	sheet := "2.来易去难"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("AB配置表解析错误: %s", sheet)
		return
	}
	return
}

// 更新彩票配置
func updateLottery(d []byte, save bool) error {
	byte_reader := bytes.NewReader(d)

	f, err := excelize.OpenReader(byte_reader)
	if err != nil {
		return err
	}

	l1, err := parseLottery1(f)
	if err != nil {
		return err
	}
	l2, err := parseLottery2(f)
	if err != nil {
		return err
	}
	l3, err := parseLottery3(f)
	if err != nil {
		return err
	}
	l4, err := parseLottery4(f)
	if err != nil {
		return err
	}
	l5, err := parseLottery5(f)
	if err != nil {
		return err
	}

	newbiews := make(map[int]data.FreeNewbiewMode)
	for _, l := range l4 {
		newbiews[l.Id] = data.FreeNewbiewMode{
			OutCashInterval: l.OutCashInterval.Value,
			Winning:         l.Winning.Value,
			OutCashLimited:  l.OutCashLimited,
		}
	}

	var games []data.Game
	base := l1[0]
	robot := l3[0]
	for _, v1 := range l2 {
		games = append(games, data.Game{
			Id:           v1.Id,
			Name:         v1.Name,
			Gtype:        base.Gtype,
			Status:       base.Status,
			Ai_Status:    base.Ai_Status,
			Algo_Switch:  base.Algo_Switch,
			Count:        base.Count,
			Dtype:        v1.DType,
			Stock_Expect: v1.Stock_Expect,
			Stock_Alarm:  v1.Stock_Alarm,
			LOTTERY: data.CPGame{
				ChipLimit:        base.ChipLimit.Value,
				BetLimit:         base.BetLimited.Value,
				BetInterval:      base.BetInterval.Value,
				BetRoom:          base.BetsRoom.Value,
				NoviceRoom:       base.NoviceRoom.Value,
				ExceptionRoom:    base.ExceptionRoom.Value,
				ControlRoom:      base.ControlRoom.Value,
				FinalCoefficient: v1.FactorInterval.Value,
				WinningPro:       v1.WinningInterval.Value,
				LoseLimited:      v1.MaxLoseScore,
				DrawPro:          v1.DrawWeight.Value,
				MTax:             v1.MingTax,
				ATax:             v1.AnTax,
				BetTime:          base.BetTime,
				Msg_Score:        int64(base.Msg_Score),
				UnRandomRate:     base.UnRandomRate,
				NewbieMaxWin:     base.NewbieMaxWin,
				NewbieMinWin:     base.NewbieMinWin,
				NewbieGiveRate:   base.NewbieGiveRate,
				WinnabilityRate:  base.WinnabilityRate,
				JackpotExceed:    base.JackpotExceed,
				NewbiewMode:      newbiews[v1.NewibewId],
				ANewbiewMode:     newbiews[v1.ANewibewId],
				RTPFixRate:       base.RTPFixRate.Value,
				Robot: data.CPGameRobotStrategy{
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
				SFJP: data.CPSFJPStrategy{
					SFChargeFloor: l5[0].SFChargeFloor,
					SFFactorFloor: l5[0].SFFactorFloor,
					SFStockUpper:  l5[0].SFStockUpper,
					SFTriggerPro:  l5[0].SFTriggerPro,
					JPChargeUpper: l5[0].JPChargeUpper,
					JPFactorUpper: l5[0].JPFactorUpper,
					JPStockFloor:  l5[0].JPStockFloor,
					JPTriggerPro:  l5[0].JPTriggerPro,
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

// 更新彩票配置
func updateCPStrategy(save bool, f *excelize.File) error {
	cp1s, err := parseCPStrategy1(f)
	if err != nil {
		return err
	}

	cp2s, err := parseCPStrategy2(f)
	if err != nil {
		return err
	}

	cp1 := cp1s[0]
	cp2 := cp2s[0]

	c := data.CPStrategy{
		Id: 1,
		LYQN: data.CPLYQNStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         cp2.Id,
				Weight:     cp2.Weight,
				Mutex:      cp2.Mutex.Value,
				O:          cp2.O,
				UserType:   cp2.UserType.Value,
				ChargeType: cp2.ChargeType.Value,
			},
			X:  cp2.X,
			D:  cp2.D.Value,
			M:  cp2.M.Value,
			RR: cp2.RR.Value,
			PR: cp2.PR.Value,
			T:  cp2.T.Value,
			PS: cp2.PS.Value,
		},
		LWJY: data.CPLWJYStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         cp1.Id,
				Weight:     cp1.Weight,
				Mutex:      cp1.Mutex.Value,
				O:          cp1.O,
				UserType:   cp1.UserType.Value,
				ChargeType: cp1.ChargeType.Value,
			},
			X:  cp1.X,
			M:  cp1.M.Value,
			RR: cp1.RR.Value,
			PR: cp1.PR.Value,
			C:  cp1.C.Value,
			U:  cp1.U.Value,
			UZ: cp1.UZ.Value,
		},
	}

	config.SetCPStrategy(c)
	if save {
		c.Save()
	}

	return nil
}
