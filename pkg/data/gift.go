package data

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/glog"
	"goserver/pkg/table"
)

type CommonGiftRecharge struct {
	Id    string
	Gtype int32
	Price int32
}

// 判断能不能买
func (a *CommonGiftRecharge) CanOrder(user *User, id string, ctype int32) bool {
	return true
}

// 充值之后
func (act *CommonGiftRecharge) RechargeAfter(user *User, id string, ctype int32, save bool) (*pb.PaySuccessNtf, error) {
	rsp := new(pb.PaySuccessNtf)
	rsp.Id = id
	rsp.Rtype = pb.LIMITEDGIFT

	// user.CommonGiftRechargeId = 0
	// user.CommonGiftRechargeOverTime = 0
	// user.OverCommonGiftRecharge = append(user.OverCommonGiftRecharge, act.Id)

	// if save {
	// 	user.UpdateCommonGiftRecharge()
	// }
	rsp.Atype = pb.ACT_TYPE3
	glog.Infof("user %s recharge success, limitId:%s", user.Userid, id)
	return rsp, nil
}

// 完善订单
func (act *CommonGiftRecharge) BuildPayOrder(user *User, order *PayOrder, game bool) {
	order.ShopId = act.Id
	order.Amount = uint32(act.Price)
	order.Score = uint32(act.Price)

	// 赠送
	bean := table.GetTables().GiftRechargeTable.Get(act.Id)
	giveRate := bean.GiveRatio[user.RegistArea].Nums
	order.OtherPresent = fmt.Sprintf("%d", int64(act.Price)*int64(giveRate[0])/10000)
	order.GiveCash = int64(act.Price) * int64(giveRate[1]) / 10000
	order.GiveWithdrawal = int64(act.Price) * int64(giveRate[2]) / 10000
	// if giveRate > 0 {
	// 	order.OtherPresent = fmt.Sprintf("%d", int64(act.Price)*int64(giveRate)/10000)
	// }

	switch act.Gtype {
	case 1:
		order.ShopType = LIMITED_GIFT
		order.ShopName = "破冰高送礼包" + act.Id
	case 2:
		order.ShopType = Withdraw_GIFT
		order.ShopName = "破冰解提礼包" + act.Id
	case 4:
		order.ShopType = BREAKING_GITF
		order.ShopName = "破产通用礼包" + act.Id
	}

}

func (act *CommonGiftRecharge) RechageLType() int32 {
	switch act.Gtype {
	case 1:
		return int32(pb.LOG_TYPE84)
	case 2:
		return int32(pb.LOG_TYPE133)
	default:
		return int32(pb.LOG_TYPE86)
	}
}
