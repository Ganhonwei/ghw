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
type RM1 struct {
	Id               string `xls:"房间ID"`
	Name             string `xls:"房间名称"`
	Gtype            int32  `xls:"玩法类型"`
	Status           int    `xls:"房间开关"`
	Ai_Status        int    `xls:"房间ai开关"`
	Min_Access       int    `xls:"最低准入"`
	Max_Access       int    `xls:"最高准入"`
	Stock_Expect     int64  `xls:"库存期望"`
	Stock_Alarm      int    `xls:"库存报警"`
	SortId           int    `xls:"房间分组内排序标识"`
	Count            uint32 `xls:"人数上限"`
	Bottom           int    `xls:"底注"`
	Otime            int    `xls:"操作时间"`
	Scountdown       int    `xls:"结算倒数时间"`
	Match_Time       ints   `xls:"匹配时间区间"`
	Single_Robot     ints   `xls:"单局人机数区间"`
	Robot_Join       ints   `xls:"人机加入频率"`
	Robot_Leave      ints   `xls:"人机离开概率"`
	Prevent_Time     int    `xls:"防作弊匹配时间检测"`
	Prevent_Num      int    `xls:"防作弊匹配次数检测"`
	Prevent_Thaw     int    `xls:"防作弊匹配解冻时间"`
	Msg_Score        int    `xls:"跑马灯显示分"`
	RoomFactorSwitch int    `xls:"房间系数开关"`
	BreakingPrice    int    `xls:"破产礼包价格"`
	BreakingGive     int    `xls:"破产礼包赠送"`
	BreakingTimes    int    `xls:"破产礼包每日次数"`
	RoomType         int    `xls:"房间类型（0匹配1对战房2对战房娱乐）"`
	MinFirstEntry    int    `xls:"初始携带要求"`
	RoomRecharge     int32s `xls:"局内充值"`
	RoomRechargeGive int32s `xls:"局内充值赠送"`
	NoviceId         int    `xls:"B类新手配置"`
	ANoviceId        int    `xls:"A类新手配置"`
	ACardType        int    `xls:"A类指定牌型"`
}

// 2.房间发牌配置表
type RM2 struct {
	Id               string `xls:"房间ID"`
	NewbieType       int    `xls:"新手状态使用摸牌ID"`
	AbnormalType     int    `xls:"异常状态使用摸牌ID"`
	FinalFactorRange ints   `xls:"正常状态和点控最终系数区间"`
	CardTypeRange    ints   `xls:"正常状态和点控系数对应使用牌型"`
	MingTax          int32  `xls:"明税"`
	AnTax            int32  `xls:"暗税"`
	Mode             int    `xls:"模式"`
}

type RM2S []RM2

func (rm2 RM2S) Find(id string) RM2 {
	for _, v := range rm2 {
		if v.Id == id {
			return v
		}
	}
	return RM2{}
}

// 3.发牌牌型配置表
type RM3 struct {
	Id                            int  `xls:"摸牌ID"`
	PlayerWeight                  ints `xls:"玩家初始牌型权重"`
	PlayerWeight2CardTypeId       ints `xls:"玩家初始牌型权重对应初始牌型ID"`
	PlayerDrawCardWeight          ints `xls:"玩家摸牌权重"`
	RobotWeigtht                  ints `xls:"人机初始牌型权重"`
	RobotWeight2CardTypeId        ints `xls:"人机初始牌型权重对应初始牌型ID"`
	RobotDrawCardWeight           ints `xls:"人机摸牌权重"`
	RobotActionTime               ints `xls:"人机操作思考时间毫秒"`
	CardPoolIdWeight              ints `xls:"牌库编号权重概率"`
	CardPoolId                    ints `xls:"牌库编号"`
	UpCardPool                    int  `xls:"升档牌库"`
	MustLoseProb                  int  `xls:"必输概率"`
	MustLoseScore                 ints `xls:"必输分数线"`
	MustLoseCoreRobotGoodCardProb int  `xls:"必输局核心人机摸好牌概率"`
	WinScoreLimit                 int  `xls:"赢分限制"`
}

// 4.牌型配置
type RM4 struct {
	Id             int         `xls:"初始牌型ID"`
	CardTypeConfig stringslice `xls:"牌型配置"`
}

// 5.新手模式
type RM5 struct {
	Id                       int  `xls:"ID"`
	CanWithdrawRange         ints `xls:"新手可提彩金区间"`
	WinRate                  ints `xls:"对应胜率"`
	CanWithdrawLimit         int  `xls:"新手可提现彩金上限"`
	WinWeight                ints `xls:"新手赢时权重"`
	WinCardType              ints `xls:"新手赢时权重对应牌型"`
	LoseWeight               ints `xls:"新手输时权重"`
	LoseCardType             ints `xls:"新手输时权重对应牌型"`
	CivilianCanWithdrawRange ints `xls:"平民可提彩金区间"`
	CivilianWinRate          ints `xls:"平民对应胜率"`
	CivilianCanWithdrawLimit int  `xls:"平民可提现彩金上限"`
	CivilianWinWeight        ints `xls:"平民赢时权重"`
	CivilianWinCardType      ints `xls:"平民赢时权重对应牌型"`
	CivilianLoseWeight       ints `xls:"平民输时权重"`
	CivilianLoseCardType     ints `xls:"平民输时权重对应牌型"`
	SpecialRound             ints `xls:"新手特定局"`
	SpecialRoundCardType     ints `xls:"新手特定局牌型ID"`
	FoamCardType             int  `xls:"泡沫状态牌型"`
}

func parseRM1(f *excelize.File) (ret []RM1, err error) {
	sheet := "1.房间基础配置表"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("rm配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseRM2(f *excelize.File) (ret RM2S, err error) {
	sheet := "2.房间发牌配置表"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("rm配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseRM3(f *excelize.File) (ret []RM3, err error) {
	sheet := "3.发牌牌型配置表"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("rm配置表解析错误: %s", sheet)
		return
	}
	return
}
func parseRM4(f *excelize.File) (ret []RM4, err error) {
	sheet := "4.牌型配置"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("rm配置表解析错误: %s", sheet)
		return
	}
	return
}
func parseRM5(f *excelize.File) (ret []RM5, err error) {
	sheet := "5.新手模式"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("rm配置表解析错误: %s", sheet)
		return
	}
	return
}

// 更新tp配置
func updateRM(d []byte, save bool) error {
	byte_reader := bytes.NewReader(d)

	f, err := excelize.OpenReader(byte_reader)
	if err != nil {
		return err
	}

	rm1, err := parseRM1(f)
	if err != nil {
		return err
	}
	rm2, err := parseRM2(f)
	if err != nil {
		return err
	}
	rm3, err := parseRM3(f)
	if err != nil {
		return err
	}
	rm4, err := parseRM4(f)
	if err != nil {
		return err
	}
	rm5, err := parseRM5(f)
	if err != nil {
		return err
	}

	var v3s []data.RMGameCardType
	for _, v3 := range rm3 {
		v3s = append(v3s, data.RMGameCardType{
			Id:                            v3.Id,
			PlayerWeight:                  v3.PlayerWeight.Value,
			PlayerWeight2CardTypeId:       v3.PlayerWeight2CardTypeId.Value,
			PlayerDrawCardWeight:          v3.PlayerDrawCardWeight.Value,
			RobotWeigtht:                  v3.RobotWeigtht.Value,
			RobotWeight2CardTypeId:        v3.RobotWeight2CardTypeId.Value,
			RobotDrawCardWeight:           v3.RobotDrawCardWeight.Value,
			RobotActionTime:               v3.RobotActionTime.Value,
			CardPoolIdWeight:              v3.CardPoolIdWeight.Value,
			CardPoolId:                    v3.CardPoolId.Value,
			UpCardPool:                    v3.UpCardPool,
			MustLoseProb:                  v3.MustLoseProb,
			MustLoseScore:                 v3.MustLoseScore.Value,
			MustLoseCoreRobotGoodCardProb: v3.MustLoseCoreRobotGoodCardProb,
			WinScoreLimit:                 v3.WinScoreLimit,
		})
	}

	var v4s []data.RMGameCardType2
	for _, v4 := range rm4 {
		v4s = append(v4s, data.RMGameCardType2{
			Id:             v4.Id,
			CardTypeConfig: v4.CardTypeConfig.Value,
		})
	}

	v5s := make(map[int]data.RMNewbieMode)
	for _, v5 := range rm5 {
		v5s[v5.Id] = data.RMNewbieMode{
			CanWithdrawRange:         v5.CanWithdrawRange.Value,
			WinRate:                  v5.WinRate.Value,
			CanWithdrawLimit:         v5.CanWithdrawLimit,
			WinWeight:                v5.WinWeight.Value,
			WinCardType:              v5.WinCardType.Value,
			LoseWeight:               v5.LoseWeight.Value,
			LoseCardType:             v5.LoseCardType.Value,
			CivilianCanWithdrawRange: v5.CivilianCanWithdrawRange.Value,
			CivilianWinRate:          v5.CivilianWinRate.Value,
			CivilianCanWithdrawLimit: v5.CivilianCanWithdrawLimit,
			CivilianWinWeight:        v5.CivilianWinWeight.Value,
			CivilianWinCardType:      v5.CivilianWinCardType.Value,
			CivilianLoseWeight:       v5.CivilianLoseWeight.Value,
			CivilianLoseCardType:     v5.CivilianLoseCardType.Value,
			SpecialRound:             v5.SpecialRound.Value,
			SpecialRoundCardType:     v5.SpecialRoundCardType.Value,
			FoamCardType:             v5.FoamCardType,
		}
	}

	var games []data.Game
	for _, v1 := range rm1 {
		v2 := rm2.Find(v1.Id)
		if v2.Id != v1.Id {
			return fmt.Errorf("rm2 find error")
		}
		games = append(games, data.Game{
			Id:            v1.Id,
			Name:          v1.Name,
			Gtype:         v1.Gtype,
			Status:        v1.Status,
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
			RM: data.RMGame{
				Bottom:           v1.Bottom,
				Otime:            v1.Otime,
				Scountdown:       v1.Scountdown,
				Match_Time:       v1.Match_Time.Value,
				Single_Robot:     v1.Single_Robot.Value,
				Robot_Join:       v1.Robot_Join.Value,
				Robot_Leave:      v1.Robot_Leave.Value,
				Prevent_Time:     v1.Prevent_Time,
				Prevent_Num:      v1.Prevent_Num,
				Prevent_Thaw:     v1.Prevent_Thaw,
				Msg_Score:        v1.Msg_Score,
				NewbieType:       v2.NewbieType,
				AbnormalType:     v2.AbnormalType,
				FinalFactorRange: v2.FinalFactorRange.Value,
				CardTypeRange:    v2.CardTypeRange.Value,
				MingTax:          v2.MingTax,
				AnTax:            v2.AnTax,
				Mode:             v2.Mode,
				RoomFactorSwitch: v1.RoomFactorSwitch,
				CardType:         v3s,
				CardType2:        v4s,
				NewbieMode:       v5s[v1.NoviceId],
				ANewbieMode:      v5s[v1.ANoviceId],
				ACardType:        v1.ACardType,
				RoomRecharge:     v1.RoomRecharge.Value,
				RoomRechargeGive: v1.RoomRechargeGive.Value,
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
