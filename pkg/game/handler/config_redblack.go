package handler

import (
	"fmt"
	"goserver/pkg/data"
	"goserver/pkg/game/config"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/dhushon/decxls"
)

// 1.红运当头
type RBStrategy1 struct {
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

// 2.绝处逢生
type RBStrategy2 struct {
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

func parseRBStrategy1(f *excelize.File) (ret []RBStrategy1, err error) {
	sheet := "1.红运当头"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("RB配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseRBStrategy2(f *excelize.File) (ret []RBStrategy2, err error) {
	sheet := "2.绝处逢生"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("RB配置表解析错误: %s", sheet)
		return
	}
	return
}

func updateRBStrategy(save bool, f *excelize.File) error {
	c1, err := parseRBStrategy1(f)
	if err != nil {
		return err
	}

	c2, err := parseRBStrategy2(f)
	if err != nil {
		return err
	}

	c := data.RBStrategy{
		Id: 1,
		HYDT: data.RBHYDTStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         c1[0].Id,
				Weight:     c1[0].Weight,
				Mutex:      c1[0].Mutex.Value,
				O:          c1[0].O,
				UserType:   c1[0].UserType.Value,
				ChargeType: c1[0].ChargeType.Value,
			},
			X:  c1[0].X,
			M:  c1[0].M.Value,
			RR: c1[0].RR.Value,
			PR: c1[0].PR.Value,
			C:  c1[0].C.Value,
			U:  c1[0].U.Value,
			UZ: c1[0].UZ.Value,
		},
		JCFS: data.RBJCFSStrategy{
			BaseStrategy: &data.BaseStrategy{
				Id:         c2[0].Id,
				Weight:     c2[0].Weight,
				Mutex:      c2[0].Mutex.Value,
				O:          c2[0].O,
				UserType:   c2[0].UserType.Value,
				ChargeType: c2[0].ChargeType.Value,
			},
			X:  c2[0].X,
			D:  c2[0].D.Value,
			M:  c2[0].M.Value,
			RR: c2[0].RR.Value,
			PR: c2[0].PR.Value,
			T:  c2[0].T.Value,
			PS: c2[0].PS.Value,
		},
	}

	config.SetRBStrategy(c)
	if save {
		c.Save()
	}

	return nil
}
