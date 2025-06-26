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
type AB1 struct {
	Gtype           int32    `xls:"玩法类型"`
	Status          int      `xls:"桌子开关（0=关，1=开）"`
	Ai_Status       int      `xls:"桌子ai开关(0=关，1=开)"`
	Algo_Switch     int      `xls:"算法（1=old，2=new）"`
	Count           uint32   `xls:"人数上限"`
	ChipLimit       uint32s  `xls:"筹码额度"`
	BetLimited      int32s   `xls:"下注上限"`
	BetInterval     int32ss  `xls:"总下注区间（-1=无限制）"`
	BetsRoom        int32ss  `xls:"总下注对应房间"`
	NoviceRoom      int32s   `xls:"新手房间"`
	ExceptionRoom   int32s   `xls:"异常房间"`
	ControlRoom     int32s   `xls:"点控房间"`
	BetTime         int32    `xls:"下注时间"`
	Msg_Score       int      `xls:"跑马灯显示分"`
	BreakingPrice   int      `xls:"破产礼包价格"`
	BreakingGive    int      `xls:"破产礼包赠送"`
	BreakingTimes   int      `xls:"破产礼包每日次数"`
	UnRandomRate    int32    `xls:"非随机开牌概率"`
	NewbieMaxWin    int64    `xls:"新手赢分上限"`
	NewbieMinWin    int64    `xls:"新手赢分下限"`
	NewbieGiveRate  int32    `xls:"新手赢分下限送分概率"`
	WinnabilityRate int32    `xls:"可赢系数"`
	RTPFixRate      float64s `xls:"调整系数"`
}

// 2.库存配置表
type AB2 struct {
	Id               string  `xls:"房间ID"`
	RoomType         int     `xls:"房间类型0匹配1对战房2对战房娱乐"`
	Name             string  `xls:"房间名"`
	DType            int32   `xls:"房间类型"`
	Stock_Expect     int64   `xls:"库存期望"`
	Stock_Alarm      int     `xls:"库存报警"`
	MingTax          int32   `xls:"明税（万分比）"`
	AnTax            int32   `xls:"暗税（万分比）"`
	FactorInterval   int32s  `xls:"最终系数区间"`
	WinningInterval  int32s  `xls:"系数对应牌型ID"`
	MaxLoseScore     int32   `xls:"单局平台输分上限"`
	NoPeopleBet      int     `xls:"无人下注牌型ID"`
	Status           int     `xls:"房间开关"`
	DealerAccess     int     `xls:"庄家携带需求"`
	MinFirstEntry    int     `xls:"初始携带要求"`
	Min_Access       int     `xls:"每局最低准入"`
	Count            int     `xls:"人数上限"`
	Bottom           uint32  `xls:"底注"`
	ChipLimit        uint32s `xls:"筹码档位"`
	OperateTime      int     `xls:"操作时间"`
	RoomRecharge     int32s  `xls:"局内充值"`
	RoomRechargeGive int32s  `xls:"局内充值赠送"`
	NewibewId        int     `xls:"B类新手配置"`
	ANewibewId       int     `xls:"A类新手配置"`
}

// 3.发牌牌型配置
type AB3 struct {
	Id         int32 `xls:"牌型ID"` // 牌型id
	PlatWinPro int   `xls:"平台胜率"` // 平台胜率
	// JackpotPro ints  `xls:"16个结果的开奖概率[平台赢最多，平台赢第二多……，平台输第二多，平台输最多]"` // 开奖概率
}

// 4.人机策略
type AB4 struct {
	Num           int32  `xls:"参与人机"`
	InitScore     int32s `xls:"初始携带分数"`
	BetWeight1    int32s `xls:"下注权重1（不下注，ANDAN，BAHAR）"`
	BetWeight2    int32s `xls:"下注权重2[不下注，1-5，6-10，11-15，16-25，26-30，31-35，36-40，40以上]"`
	BetTime       int32s `xls:"下注时间区间"`
	MinBet        int32  `xls:"最小下注基数"`
	BetPro        int32s `xls:"下注概率（不下注，下注）"`
	BetInterval   int32s `xls:"下注区间"`
	LeaveLimit    int32s `xls:"人机携带退出下上限"`
	LeaveBetTimes int32s `xls:"人机下注退出下上限"`
}

// 5.新手模式
type AB5 struct {
	Id               int  `xls:"ID"`         //
	NWithdrawable    ints `xls:"新手可提彩金区间"`   // 新手可提彩金区间
	NWinRate         ints `xls:"对应胜率"`       // 对应胜率
	NWithdrawLimited int  `xls:"新手可提现彩金上限"`  // 新手可提现彩金上限
	NWinWeight       ints `xls:"新手赢时权重"`     // 新手赢时权重
	NWinCardType     ints `xls:"新手赢时权重对应牌型"` // 新手赢时权重对应牌型
	NLoseWeight      ints `xls:"新手输时权重"`     // 新手输时权重
	NLoseCardType    ints `xls:"新手输时权重对应牌型"` // 新手输时权重对应牌型
	CWithdrawable    ints `xls:"平民可提彩金区间"`   // 平民可提彩金区间
	CWinRate         ints `xls:"平民对应胜率"`     // 平民对应胜率
	CWithdrawLimited int  `xls:"平民可提现彩金上限"`  // 平民可提现彩金上限
	CWinWeight       ints `xls:"平民赢时权重"`     // 平民赢时权重
	CWinCardType     ints `xls:"平民赢时权重对应牌型"` // 平民赢时权重对应牌型
	CLoseWeight      ints `xls:"平民输时权重"`     // 平民输时权重
	CLoseCardType    ints `xls:"平民输时权重对应牌型"` // 平民输时权重对应牌型
	FrothCardType    int  `xls:"泡沫对应牌型"`     // 泡沫对应牌型
}

// 6.结果权重配置
type AB6 struct {
	Group           string `xls:"组合"`       // 组合
	PlatWinWeights  ints   `xls:"平台赢权重"`    // 平台赢权重
	PlatLoseWeights ints   `xls:"玩家赢权重"`    // 玩家赢权重
	NatureWeights   ints   `xls:"自然概率加权权重"` // 自然概率加权权重
}

// 1.安能求死
type ABStrategy1 struct {
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

// 2.安然躺赢
type ABStrategy2 struct {
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

func parseAB1(f *excelize.File) (ret []AB1, err error) {
	sheet := "1.房间基础配置表"
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

func parseAB2(f *excelize.File) (ret []AB2, err error) {
	sheet := "2.库存配置表"
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

func parseAB3(f *excelize.File) (ret []AB3, err error) {
	sheet := "3.发牌牌型配置"
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

func parseAB4(f *excelize.File) (ret []AB4, err error) {
	sheet := "4.人机策略"
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

func parseAB5(f *excelize.File) (ret []AB5, err error) {
	sheet := "5.新手模式"
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

func parseAB6(f *excelize.File) (ret []AB6, err error) {
	sheet := "6.结果权重配置"
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

func parseABStrategy1(f *excelize.File) (ret []ABStrategy1, err error) {
	sheet := "1.安能求死"
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

func parseABStrategy2(f *excelize.File) (ret []ABStrategy2, err error) {
	sheet := "2.安然躺赢"
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

// 更新abd配置
func updateAB(d []byte, save bool) error {
	byte_reader := bytes.NewReader(d)

	f, err := excelize.OpenReader(byte_reader)
	if err != nil {
		return err
	}

	ab1, err := parseAB1(f)
	if err != nil {
		return err
	}
	ab2, err := parseAB2(f)
	if err != nil {
		return err
	}
	ab3, err := parseAB3(f)
	if err != nil {
		return err
	}
	ab4, err := parseAB4(f)
	if err != nil {
		return err
	}
	ab5, err := parseAB5(f)
	if err != nil {
		return err
	}
	ab6, err := parseAB6(f)
	if err != nil {
		return err
	}
	newbiew := make(map[int]data.ABNewbieMode)
	for _, a := range ab5 {
		newbiew[a.Id] = data.ABNewbieMode{
			CanWithdrawRange:         a.NWithdrawable.Value,
			WinRate:                  a.NWinRate.Value,
			CanWithdrawLimit:         a.NWithdrawLimited,
			WinWeight:                a.NWinWeight.Value,
			WinCardType:              a.NWinCardType.Value,
			LoseWeight:               a.NLoseWeight.Value,
			LoseCardType:             a.NLoseCardType.Value,
			CivilianCanWithdrawRange: a.CWithdrawable.Value,
			CivilianWinRate:          a.CWinRate.Value,
			CivilianCanWithdrawLimit: a.CWithdrawLimited,
			CivilianWinWeight:        a.CWinWeight.Value,
			CivilianWinCardType:      a.CWinCardType.Value,
			CivilianLoseWeight:       a.CLoseWeight.Value,
			CivilianLoseCardType:     a.CLoseCardType.Value,
			FrothCardType:            a.FrothCardType,
		}
	}

	var games []data.Game
	base := ab1[0]
	robot := ab4[0]
	for _, v1 := range ab2 {
		game := data.Game{
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
			AB: data.ABGame{
				ChipLimit:        base.ChipLimit.Value,
				BetLimit:         base.BetLimited.Value,
				BetInterval:      base.BetInterval.Value,
				BetRoom:          base.BetsRoom.Value,
				NoviceRoom:       base.NoviceRoom.Value,
				ExceptionRoom:    base.ExceptionRoom.Value,
				ControlRoom:      base.ControlRoom.Value,
				FinalCoefficient: v1.FactorInterval.Value,
				CardTypeID:       v1.WinningInterval.Value,
				LoseLimited:      v1.MaxLoseScore,
				MTax:             v1.MingTax,
				ATax:             v1.AnTax,
				BetTime:          base.BetTime,
				Msg_Score:        int64(base.Msg_Score),
				NoPeopleBet:      v1.NoPeopleBet,
				RTPFixRate:       base.RTPFixRate.Value,
				Robot: data.ABGameRobotStrategy{
					Num:        robot.Num,
					InitScore:  robot.InitScore.Value,
					BetWeight1: robot.BetWeight1.Value,
					BetWeight2: robot.BetWeight2.Value,
					TimeArea:   robot.BetTime.Value,
					Min:        robot.MinBet,
					BetPro:     robot.BetPro.Value,
					BetArea:    robot.BetInterval.Value,
					LeaveLimit: robot.LeaveLimit.Value,
					BetTimes:   robot.LeaveBetTimes.Value,
				},
				NewbiewMode:  newbiew[v1.NewibewId],
				ANewbiewMode: newbiew[v1.ANewibewId],
			},
		}
		for _, ab := range ab6 {
			l := data.ABLeadWeight{
				Group:           ab.Group,
				PlatWinWeights:  ab.PlatWinWeights.Value,
				PlatLoseWeights: ab.PlatLoseWeights.Value,
				NatureWeights:   ab.NatureWeights.Value,
			}
			game.AB.LeadWeights = append(game.AB.LeadWeights, l)
		}

		for _, ab := range ab3 {
			c := data.ABGameCardType{
				Id:     int(ab.Id),
				WinPro: ab.PlatWinPro,
				// DrawPro: ab.JackpotPro.Value,
			}
			game.AB.CardType = append(game.AB.CardType, c)
		}
		// 对战房配置
		switch v1.RoomType {
		case 1: // 真金模式
			fallthrough
		case 2: // 娱乐模式
			game.RoomType = v1.RoomType
			game.Status = v1.Status
			game.Count = uint32(v1.Count)
			game.Min_Access = v1.Min_Access
			game.MinFirstEntry = v1.MinFirstEntry
			game.AB.ChipLimit = v1.ChipLimit.Value
			game.AB.DealerAccess = v1.DealerAccess
			game.AB.MinFirstEntry = v1.MinFirstEntry
			game.AB.Min_Access = v1.Min_Access
			game.AB.Count = v1.Count
			game.AB.Bottom = v1.Bottom
			game.AB.OperateTime = v1.OperateTime
			game.AB.RoomRecharge = v1.RoomRecharge.Value
			game.AB.RoomRechargeGive = v1.RoomRechargeGive.Value
		}
		games = append(games, game)
	}

	for _, v := range games {
		if save {
			v.Save()
		}

		config.SetGame(v)
	}

	return nil
}

// 更新ab配置
func updateABStrategy(save bool, f *excelize.File) error {
	ab1s, err := parseABStrategy1(f)
	if err != nil {
		return err
	}

	ab2s, err := parseABStrategy2(f)
	if err != nil {
		return err
	}

	ab1 := ab1s[0]
	ab2 := ab2s[0]

	c := data.ABStrategy{
		Id: 1,
		ANQS: data.ABANQSStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         ab1.Id,
				Weight:     ab1.Weight,
				Mutex:      ab1.Mutex.Value,
				O:          ab1.O,
				UserType:   ab1.UserType.Value,
				ChargeType: ab1.ChargeType.Value,
			},
			X:  ab1.X,
			D:  ab1.D.Value,
			M:  ab1.M.Value,
			RR: ab1.RR.Value,
			PR: ab1.PR.Value,
			T:  ab1.T.Value,
			PS: ab1.PS.Value,
		},
		ARTY: data.ABARTYStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         ab2.Id,
				Weight:     ab2.Weight,
				Mutex:      ab2.Mutex.Value,
				O:          ab2.O,
				UserType:   ab2.UserType.Value,
				ChargeType: ab2.ChargeType.Value,
			},
			X:  ab2.X,
			M:  ab2.M.Value,
			RR: ab2.RR.Value,
			PR: ab2.PR.Value,
			C:  ab2.C.Value,
			U:  ab2.U.Value,
			UZ: ab2.UZ.Value,
		},
	}

	config.SetABStrategy(c)
	if save {
		c.Save()
	}

	return nil
}
