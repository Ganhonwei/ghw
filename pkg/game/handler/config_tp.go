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
type TP1 struct {
	Id                string `xls:"房间ID"`
	Name              string `xls:"房间名称"`
	Gtype             int32  `xls:"玩法类型"`
	Status            int    `xls:"房间开关"`
	Ai_Status         int    `xls:"房间ai开关"`
	Kick_Score        int    `xls:"踢出分数"`
	Min_Access        int    `xls:"最低准入"`
	Max_Access        int    `xls:"最高准入"`
	Stock_Expect      int64  `xls:"库存期望"`
	Stock_Alarm       int    `xls:"库存报警"`
	SortId            int    `xls:"房间分组内排序标识"`
	Count             uint32 `xls:"人数上限"`
	Bottom            int    `xls:"底注"`
	Rounds            int    `xls:"轮次上限"`
	Than_Rounds       int    `xls:"可比牌轮次"`
	Pool_Limit        int    `xls:"筹码池上限"`
	Otime             int    `xls:"操作时间"`
	Tcountdown        int    `xls:"比牌倒计时"`
	Scountdown        int    `xls:"结算倒数时间"`
	Match_Time        ints   `xls:"匹配时间区间"`
	Single_Robot      ints   `xls:"单局人机数区间"`
	Robot_Join        ints   `xls:"人机加入频率"`
	Robot_Leave       int    `xls:"人机离开概率（万分比）"`
	Prevent_Time      int    `xls:"防作弊匹配时间检测"`
	Prevent_Num       int    `xls:"防作弊匹配次数检测"`
	Prevent_Thaw      int    `xls:"防作弊匹配解冻时间"`
	Msg_Score         int    `xls:"跑马灯显示分"`
	RoomRecharge      int32s `xls:"局内充值"`
	RoomRechargeGive  int32s `xls:"局内充值赠送"`
	AnteFactor        int    `xls:"底注系数"`
	ShowSwitch        int    `xls:"结算摊牌开关"`
	Strategy100Switch int    `xls:"策略100开关"`
	Strategy200Switch int    `xls:"策略200开关"`
	Strategy300Switch int    `xls:"策略300开关"`
	Strategy100Round  int    `xls:"策略100累计局数"`
	Strategy200Round  int    `xls:"策略200累计局数"`
	Strategy300Round  int    `xls:"策略300累计局数"`
	RoomFactorSwitch  int    `xls:"房间系数开关"`
	BreakingPrice     int    `xls:"破产礼包价格"`
	BreakingGive      int    `xls:"破产礼包赠送"`
	BreakingTimes     int    `xls:"破产礼包每日次数"`
	RoomType          int    `xls:"房间类型（0匹配1对战房2对战房娱乐）"`
	MinFirstEntry     int    `xls:"初始携带要求"`
	NoviceId          int    `xls:"B类新手配置"`
	ANoviceId         int    `xls:"A类新手配置"`
	ACardType         int    `xls:"A类指定牌型"`
}

// 2.房间发牌配置表
type TP2 struct {
	Id               string `xls:"房间ID"`
	NewbieType       int    `xls:"新手状态使用牌型"`
	AbnormalType     int    `xls:"异常状态使用牌型"`
	FinalFactorRange ints   `xls:"正常状态和点控最终系数区间"`
	CardTypeRange    ints   `xls:"正常状态和点控系数对应使用牌型"`
	MingTax          int32  `xls:"明税"`
	AnTax            int32  `xls:"暗税"`
}

type TP2S []TP2

func (tp2 TP2S) Find(id string) TP2 {
	for _, v := range tp2 {
		if v.Id == id {
			return v
		}
	}
	return TP2{}
}

// 3.发牌牌型配置
type TP3 struct {
	Id                 int  `xls:"牌型ID"`
	PlayerWeight       ints `xls:"玩家权重"`
	RobotWeight        ints `xls:"人机权重"`
	WinRateCheck       int  `xls:"是否启用胜率检测（0=不启用，1=启用）"`
	PlayerWinRate      int  `xls:"玩家获胜概率（万分比）"`
	RobotStrategyGroup int  `xls:"人机策略组"`
}

// 4.人机策略组配置
type TP4 struct {
	Id                 int `xls:"策略组ID"`
	BaoZiBigger        int `xls:"人机豹子且比玩家大"`
	TongHuaShunBigger  int `xls:"人机同花顺且比玩家大"`
	ShunZiBigger       int `xls:"人机顺子比玩家大"`
	TongHuaBigger      int `xls:"人机同花比玩家大"`
	DuiZiBigger        int `xls:"人机对子比玩家大"`
	GaoPaiBigger       int `xls:"人机高牌比玩家大"`
	BaoZiSmaller       int `xls:"人机豹子且比玩家小"`
	TongHuaShunSmaller int `xls:"人机同花顺且比玩家小"`
	ShunZiSmaller      int `xls:"人机顺子比玩家小"`
	TongHuaSmaller     int `xls:"人机同花比玩家小"`
	DuiZiSmaller       int `xls:"人机对子比玩家小"`
	GaoPaiSmaller      int `xls:"人机高牌比玩家小"`
}

// 5.人机策略配置
type TP5 struct {
	Id           int32   `xls:"人机策略ID"`
	ActionWeight int32ss `xls:"操作权重（弃牌，比牌，普通下注，加倍下注）中括号表示轮数"`
	ActionTime   int32s  `xls:"操作时间[最低思考时间，最长思考时间]"`
	SeeWeight    int32s  `xls:"seen看牌权重[自己回合看牌概率万分比；其他玩家回合看牌概率万分比]"`
	SeeTime      int32ss `xls:"看牌时机[自己回合最低思考时间，自己回合最高思考时间；其他玩家回合最低思考时间，其他玩家回合最高思考时间]"`
	AgreeBi      int32   `xls:"被比牌同意概率（万分比）"`
}

// 6.新手模式
type TP6 struct {
	Id                            int  `xls:"ID"`
	CanWithdrawRange              ints `xls:"新手可提彩金区间"`
	WinRate                       ints `xls:"对应胜率"`
	CanWithdrawLimit              int  `xls:"新手可提现彩金上限"`
	WinWeight                     ints `xls:"新手赢时权重"`
	WinCardType                   ints `xls:"新手赢时权重对应牌型"`
	LoseWeight                    ints `xls:"新手输时权重"`
	LoseCardType                  ints `xls:"新手输时权重对应牌型"`
	CivilianCanWithdrawRange      ints `xls:"平民可提彩金区间"`
	CivilianWinRate               ints `xls:"平民对应胜率"`
	CivilianCanWithdrawLimit      int  `xls:"平民可提现彩金上限"`
	CivilianWinWeight             ints `xls:"平民赢时权重"`
	CivilianWinCardType           ints `xls:"平民赢时权重对应牌型"`
	CivilianLoseWeight            ints `xls:"平民输时权重"`
	CivilianLoseCardType          ints `xls:"平民输时权重对应牌型"`
	CivilianTriggerStrategyProb   int  `xls:"平民触发策略概率"`
	SpecialRound                  ints `xls:"新手特定局"`
	SpecialRoundCardType          ints `xls:"新手特定局牌型ID"`
	NewbieToCivilianWithdrawLimit int  `xls:"新手转平民可提上限"`
	FoamCheckTotalDepositLimit    int  `xls:"泡沫检查总充值上限"`
	FoamGiftRate                  int  `xls:"泡沫赠送率"`
	FoamCardType                  int  `xls:"泡沫状态牌型"`
	FoamTriggerStrategyProb       int  `xls:"泡沫触发策略概率"`
}

// 7.控制方式
type TP7 struct {
	ChargeRange        ints `xls:"充值区间"`
	ControlType        int  `xls:"控制方式"`
	WinScoreConfig     ints `xls:"当前赢分额度配置"`
	WinScoreWeight     ints `xls:"当前赢分额度配置对应概率"`
	PlayerFactorRange  ints `xls:"玩家系数范围配置"`
	PlayerFactorWeight ints `xls:"玩家系数范围配置对应换牌概率"`
	WinScoreLimit      int  `xls:"用户赢分上限"`
	Switch             int  `xls:"开关"`
}

func parsetp1(f *excelize.File) (ret []TP1, err error) {
	sheet := "1.房间基础配置表"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("tp配置表解析错误: %s", sheet)
		return
	}
	return
}

func parsetp2(f *excelize.File) (ret TP2S, err error) {
	sheet := "2.房间发牌配置表"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("tp配置表解析错误: %s", sheet)
		return
	}
	return
}

func parsetp3(f *excelize.File) (ret []TP3, err error) {
	sheet := "3.发牌牌型配置"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("tp配置表解析错误: %s", sheet)
		return
	}
	return
}
func parsetp4(f *excelize.File) (ret []TP4, err error) {
	sheet := "4.人机策略组配置"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("tp配置表解析错误: %s", sheet)
		return
	}
	return
}
func parsetp5(f *excelize.File) (ret []TP5, err error) {
	sheet := "5.人机策略配置"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("tp配置表解析错误: %s", sheet)
		return
	}
	return
}
func parsetp6(f *excelize.File) (ret []TP6, err error) {
	sheet := "6.新手模式"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("tp配置表解析错误: %s", sheet)
		return
	}
	return
}
func parsetp7(f *excelize.File) (ret []TP7, err error) {
	sheet := "7.控制方式"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("tp配置表解析错误: %s", sheet)
		return
	}
	return
}

// 更新tp配置
func updateTP(d []byte, save bool) error {
	byte_reader := bytes.NewReader(d)

	f, err := excelize.OpenReader(byte_reader)
	if err != nil {
		return err
	}

	tp1, err := parsetp1(f)
	if err != nil {
		return err
	}
	tp2, err := parsetp2(f)
	if err != nil {
		return err
	}
	tp3, err := parsetp3(f)
	if err != nil {
		return err
	}
	tp4, err := parsetp4(f)
	if err != nil {
		return err
	}
	tp5, err := parsetp5(f)
	if err != nil {
		return err
	}
	tp6, err := parsetp6(f)
	if err != nil {
		return err
	}
	tp7, err := parsetp7(f)
	if err != nil {
		return err
	}

	var v3s []data.TPGameCardType
	for _, v3 := range tp3 {
		v3s = append(v3s, data.TPGameCardType{
			Id:                 v3.Id,
			PlayerWeight:       v3.PlayerWeight.Value,
			RobotWeight:        v3.RobotWeight.Value,
			WinRateCheck:       v3.WinRateCheck,
			PlayerWinRate:      v3.PlayerWinRate,
			RobotStrategyGroup: v3.RobotStrategyGroup,
		})
	}

	var v4s []data.TPGameRobotStrategyGroup
	for _, v4 := range tp4 {
		v4s = append(v4s, data.TPGameRobotStrategyGroup{
			Id:                 v4.Id,
			BaoZiBigger:        v4.BaoZiBigger,
			TongHuaShunBigger:  v4.TongHuaShunBigger,
			ShunZiBigger:       v4.ShunZiBigger,
			TongHuaBigger:      v4.TongHuaBigger,
			DuiZiBigger:        v4.DuiZiBigger,
			GaoPaiBigger:       v4.GaoPaiBigger,
			BaoZiSmaller:       v4.BaoZiSmaller,
			TongHuaShunSmaller: v4.TongHuaShunSmaller,
			ShunZiSmaller:      v4.ShunZiSmaller,
			TongHuaSmaller:     v4.TongHuaSmaller,
			DuiZiSmaller:       v4.DuiZiSmaller,
			GaoPaiSmaller:      v4.GaoPaiSmaller,
		})
	}

	var v5s []data.TPGameRobotStrategy
	for _, v5 := range tp5 {
		v5s = append(v5s, data.TPGameRobotStrategy{
			Id:           v5.Id,
			ActionWeight: v5.ActionWeight.Value,
			ActionTime:   v5.ActionTime.Value,
			SeeWeight:    v5.SeeWeight.Value,
			SeeTime:      v5.SeeTime.Value,
			AgreeBi:      v5.AgreeBi,
		})
	}

	v6s := make(map[int]data.TPNewbieMode)
	for _, v6 := range tp6 {
		v6s[v6.Id] = data.TPNewbieMode{
			CanWithdrawRange:              v6.CanWithdrawRange.Value,
			WinRate:                       v6.WinRate.Value,
			CanWithdrawLimit:              v6.CanWithdrawLimit,
			WinWeight:                     v6.WinWeight.Value,
			WinCardType:                   v6.WinCardType.Value,
			LoseWeight:                    v6.LoseWeight.Value,
			LoseCardType:                  v6.LoseCardType.Value,
			CivilianCanWithdrawRange:      v6.CivilianCanWithdrawRange.Value,
			CivilianWinRate:               v6.CivilianWinRate.Value,
			CivilianCanWithdrawLimit:      v6.CivilianCanWithdrawLimit,
			CivilianWinWeight:             v6.CivilianWinWeight.Value,
			CivilianWinCardType:           v6.CivilianWinCardType.Value,
			CivilianLoseWeight:            v6.CivilianLoseWeight.Value,
			CivilianLoseCardType:          v6.CivilianLoseCardType.Value,
			CivilianTriggerStrategyProb:   v6.CivilianTriggerStrategyProb,
			SpecialRound:                  v6.SpecialRound.Value,
			SpecialRoundCardType:          v6.SpecialRoundCardType.Value,
			NewbieToCivilianWithdrawLimit: v6.NewbieToCivilianWithdrawLimit,
			FoamCheckTotalDepositLimit:    v6.FoamCheckTotalDepositLimit,
			FoamGiftRate:                  v6.FoamGiftRate,
			FoamCardType:                  v6.FoamCardType,
			FoamTriggerStrategyProb:       v6.FoamTriggerStrategyProb,
		}
	}

	var v7s []data.TPControlType
	for _, v7 := range tp7 {
		v7s = append(v7s, data.TPControlType{
			ChargeRange:        v7.ChargeRange.Value,
			ControlType:        v7.ControlType,
			WinScoreConfig:     v7.WinScoreConfig.Value,
			WinScoreWeight:     v7.WinScoreWeight.Value,
			PlayerFactorRange:  v7.PlayerFactorRange.Value,
			PlayerFactorWeight: v7.PlayerFactorWeight.Value,
			WinScoreLimit:      v7.WinScoreLimit,
			Switch:             v7.Switch,
		})
	}

	var games []data.Game
	for _, v1 := range tp1 {
		v2 := tp2.Find(v1.Id)
		if v2.Id != v1.Id {
			return fmt.Errorf("tp2 find error")
		}
		games = append(games, data.Game{
			Id:            v1.Id,
			Name:          v1.Name,
			Gtype:         v1.Gtype,
			Status:        v1.Status,
			Kick_Score:    v1.Kick_Score,
			Ai_Status:     v1.Ai_Status,
			Min_Access:    v1.Min_Access,
			Max_Access:    v1.Max_Access,
			Stock_Expect:  v1.Stock_Expect,
			Stock_Alarm:   v1.Stock_Alarm,
			SortId:        v1.SortId,
			Count:         v1.Count,
			BreakingPrice: v1.BreakingPrice,
			BreakingGive:  v1.BreakingGive,
			BreakingTimes: v1.BreakingTimes,
			RoomType:      v1.RoomType,
			MinFirstEntry: v1.MinFirstEntry,
			TP: data.TPGame{
				Bottom:       v1.Bottom,
				Rounds:       v1.Rounds,
				Than_Rounds:  v1.Than_Rounds,
				Pool_Limit:   v1.Pool_Limit,
				Otime:        v1.Otime,
				Tcountdown:   v1.Tcountdown,
				Scountdown:   v1.Scountdown,
				Match_Time:   v1.Match_Time.Value,
				Single_Robot: v1.Single_Robot.Value,
				Robot_Join:   v1.Robot_Join.Value,
				Robot_Leave:  v1.Robot_Leave,
				Prevent_Time: v1.Prevent_Time,
				Prevent_Num:  v1.Prevent_Num,
				Prevent_Thaw: v1.Prevent_Thaw,
				Msg_Score:    v1.Msg_Score,
				// NewbieType:         v2.NewbieType,
				// AbnormalType:       v2.AbnormalType,
				FinalFactorRange:   v2.FinalFactorRange.Value,
				CardTypeRange:      v2.CardTypeRange.Value,
				MingTax:            v2.MingTax,
				AnTax:              v2.AnTax,
				CardType:           v3s,
				RobotStrategyGroup: v4s,
				RobotStrategy:      v5s,
				NewbieMode:         v6s[v1.NoviceId],
				ANewbieMode:        v6s[v1.ANoviceId],
				ACardType:          v1.ACardType,
				ControlType:        v7s,
				RoomRecharge:       v1.RoomRecharge.Value,
				RoomRechargeGive:   v1.RoomRechargeGive.Value,
				AnteFactor:         v1.AnteFactor,
				ShowSwitch:         v1.ShowSwitch,
				Strategy100Switch:  v1.Strategy100Switch,
				Strategy200Switch:  v1.Strategy200Switch,
				Strategy300Switch:  v1.Strategy300Switch,
				Strategy100Round:   v1.Strategy100Round,
				Strategy200Round:   v1.Strategy200Round,
				Strategy300Round:   v1.Strategy300Round,
				RoomFactorSwitch:   v1.RoomFactorSwitch,
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
