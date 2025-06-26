package handler

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"sort"
	"strconv"
)

// 增加vip经验
func AddVipExp(user *data.User, exp int64) {
	bean := table.GetTables().VipTable.Get(int32(user.Vip.Lv))
	if bean.NextLv == -1 {
		user.Vip.Exp = int64(bean.Recharge)
		return
	}
	user.Vip.Exp += exp
	CheckVipLv(user)
}

func CheckVipLv(user *data.User) {
	beans := table.GetTables().VipTable.GetDataList()
	sort.Slice(beans, func(i, j int) bool {
		return beans[i].Id < beans[j].Id
	})

	lv := user.Vip.Lv //记录之前等级

	user.Vip.Lv = int(beans[len(beans)-1].Id)
	for _, v := range beans {
		if user.Vip.Exp >= int64(v.Recharge) {
			user.Vip.Lv = int(v.Id)
		}
	}

	if user.Vip.Lv != lv { //等级发生改变
		user.Vip.BeforeLv = lv
		if user.Vip.Lv > user.Vip.MaxLv {
			user.Vip.MaxLv = user.Vip.Lv
		}
	}

	// bean := config.GetVip(user.Vip.Lv)
	// if bean.NextLevel <= 0 {
	// 	return
	// }
	// bean = config.GetVip(bean.NextLevel)
	// if bean.Id == 0 {
	// 	return
	// }

	// if bean.Recharge <= user.Vip.Exp {
	// 	user.Vip.Lv = bean.Id
	// 	CheckVipLv(user)
	// 	return
	// }
}

func BuildVipData(user *data.User) []*pb.VipData {
	beans := make([]*pb.VipData, 0)
	// vips := config.GetVips()
	vips := table.GetTables().VipTable.GetDataList()

	sort.Slice(vips, func(i, j int) bool {
		return vips[i].Id < vips[j].Id
	})

	conf := table.GetTables().VBGameTaskTable.Get(2)

	giveRate := 0
	// shop := config.GetShop()
	// if user.VBBank > shop.VB[0] {
	// 	giveRate = int(shop.VB[0] * 10000 / shop.DefaultRecharge[0])
	// }

	for _, v := range vips {
		covert := table.GetTables().VipBonusTable.Get(v.Id)
		b := &pb.VipData{
			Id:       int32(v.Id),
			Recharge: int64(v.Recharge),
			// DailyReward:  v.DailySign,
			WeeklyReward:  int64(v.WeeklySign),
			UpgradeReward: int64(v.Upgrade),
			WithdrawCount: v.WithdrawTimes,
			Commission:    v.FeeRate,
			// EmojiPrice:    v.EmojiCost,
			// VoiceSwitch:   v.VoiceSwitch == 1,
			// Photo:        v.UnlockPhoto > 0,
			TaskReward: int64(conf.Rewark),
			// Compensation: table.GetTables().LoseCompensationTable.Get().Compensate,
			ShopGive:        int32(giveRate),
			Interest:        int32(table.GetTables().VipInterestTable.Get(v.Id).InterestRate[user.RegistArea]),
			WithdrawAmount:  int32(v.WithdrawAmounts),
			Interest0:       table.GetTables().VipInterestTable.Get(v.Id).InterestRate[user.RegistArea],
			WithdrawAmount0: v.WithdrawAmounts,
		}
		if covert != nil {
			b.UnlockCash = covert.OnceAmount[user.RegistArea]
			b.CovertRate = covert.ConverRate[user.RegistArea]
		}

		beans = append(beans, b)
	}
	return beans
}

func RobotVip(user *data.User) {
	r := config.GetVipRobot()
	if len(r) <= 0 {
		return
	}
	vip := config.GetVipRobot()[0]
	index, err := utils.ChoiceIntIndex(vip.VipWeights)
	if err == nil {
		user.Vip.Lv = vip.VipChoices[index]
	}
	// 头像
	if user.Vip.Lv >= 4 {
		user.Photo = strconv.Itoa(utils.RandIntN(42) + 1)
	}
	if user.Vip.Lv >= 8 {
		user.Photo = strconv.Itoa(utils.RandIntN(54) + 1)
	}
}
