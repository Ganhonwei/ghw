package dbms

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"strings"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/globalsign/mgo/bson"
)

func (a *RoleActor) CouponList(ctx actor.Context) {
	arg := ctx.Message().(*pb.CouponList)
	glog.Debugf("CouponList %#v", arg)

	rsp := new(pb.CouponRsp)
	defer ctx.Respond(rsp)

	list := data.GetCoupons(bson.M{
		"user_id": arg.Userid,
		"status":  false,
		"over_time": bson.M{
			"$gt": time.Now().Unix(),
		},
	})

	if len(list) == 0 {
		return
	}

	role := a.roles[arg.Userid]
	if role == nil {
		return
	}

	couponMap := config.GetManualCouponMap()
	for _, coupon := range list {
		if strings.HasPrefix(coupon.CouponId, "m") && couponMap != nil {
			if _, ok := couponMap.Load(coupon.CouponId); !ok {
				// 没开
				continue
			}
		}
		rsp.Coupons = append(rsp.Coupons, &pb.Coupon{
			Id:          coupon.CouponId,
			Discount:    coupon.Amount,
			MinRecharge: coupon.MinRecharge,
			OverTime:    coupon.OverTime,
		})
	}

}

func (a *RoleActor) IssueCoupon(ctx actor.Context) {
	arg := ctx.Message().(*pb.IssueCoupon)
	glog.Debugf("IssueCoupon %#v", arg)

	// 配置表指定的
	ntf := new(pb.IssueCouponNtf)
	for _, couponId := range arg.CouponId {
		table := table.GetTables().AutoCouponTable.Get(couponId)
		if table != nil {
			if coupon := a.issueAutoCoupon(arg.Userid, couponId); coupon != nil {
				ntf.Coupon = append(ntf.Coupon, &pb.Coupon{
					Id:          couponId,
					Discount:    coupon.Amount,
					MinRecharge: coupon.MinRecharge,
					OverTime:    coupon.OverTime,
				})
			}
		}
	}

	// 后台配置的券
	coupons := a.issueManualCoupon(arg.Userid)
	if len(coupons) > 0 {
		for _, coupon := range coupons {
			ntf.Coupon = append(ntf.Coupon, &pb.Coupon{
				Id:          coupon.CouponId,
				Discount:    coupon.Amount,
				MinRecharge: coupon.MinRecharge,
				OverTime:    coupon.OverTime,
			})
		}
	}

	// 有券就通知
	if r, ok := a.roles[arg.Userid]; ok && len(ntf.Coupon) > 0 {
		r.Pid.Tell(ntf)
	}
}

func (a *RoleActor) issueAutoCoupon(userid, couponId string) *data.UserCoupon {
	user := a.getUserById(userid)
	if user == nil {
		glog.Errorf("IssueCoupon user not found, userid:%s, couponId:%s", userid, couponId)
		return nil
	}

	table := table.GetTables().AutoCouponTable.Get(couponId)
	count := data.GetCouponCount(bson.M{
		"user_id":   userid,
		"coupon_id": couponId,
		"status":    false,
	})
	if count >= int(table.CurrentNum[user.RegistArea]) {
		glog.Warningf("IssueCoupon coupon count is max, userid:%s, couponId:%s", userid, couponId)
		return nil
	}

	// 添加优惠券
	coupon := &data.UserCoupon{
		Id:          bson.NewObjectId().String(),
		UserId:      userid,
		CouponId:    couponId,
		Amount:      int64(table.Coupon[user.RegistArea] * 100),
		MinRecharge: int64(table.MinRecharge[user.RegistArea] * 100),
		Status:      false,
		OverTime:    time.Now().Unix() + int64(table.ValidityTime[user.RegistArea]*60*60),
		Ctime:       time.Now().Unix(),
		IsAuto:      true,
	}
	if coupon.Save() {
		return coupon
	}
	return nil
}

func (a *RoleActor) issueManualCoupon(userid string) []*data.UserCoupon {
	user := a.getUserById(userid)
	if user == nil {
		glog.Errorf("IssueCoupon user not found, userid:%s", userid)
		return nil
	}

	coupons := config.GetManualCouponMap()

	var payAvg int64
	if user.ChargeTimes > 0 {
		payAvg = int64(user.Money) / 100 / int64(user.ChargeTimes)
	}
	// r := make(map[string]any)
	// err := ck.Select(&r, `
	// SELECT AVG(amount) amount_avg FROM game.col_trade_record FINAL WHERE userid = ? AND order_status = 4
	// `, user.Userid)
	// if err != nil {
	// 	glog.Error(err)
	// } else {
	// 	payAvg = utils.ToInt64(r["amount_avg"])
	// }

	if user.CouponIssueTime == nil {
		user.CouponIssueTime = make(map[string]int64)
	}

	list := make([]*data.UserCoupon, 0)
	coupons.Range(func(key, value any) bool {
		coupon := value.(*data.ManualCoupon)
		if coupon == nil {
			return true
		}
		if _, ok := user.CouponIssueTime[coupon.Id]; ok {
			// 领过了
			return true
		}
		if time.Now().Unix() < coupon.SendTime {
			// 没到发放时间
			return true
		}

		if len(coupon.UserIds) > 0 {
			has := false
			for _, id := range coupon.UserIds {
				if id == userid {
					has = true
					break
				}
			}
			if !has {
				return true
			}
		}

		has := len(coupon.AcctType) == 0 // 没有就是全部
		for _, acctType := range coupon.AcctType {
			if acctType == user.RegistArea {
				has = true
				break
			}
		}
		if !has {
			return true
		}

		has = len(coupon.RecharClassify) == 0 // 没有就是全部
		chargeType := handler.GetChargeType(user)
		for _, recharClassify := range coupon.RecharClassify {
			if recharClassify == chargeType {
				has = true
				break
			}
		}
		if !has {
			return true
		}

		// 支付均值
		if (payAvg < coupon.OrderAvgRange[0] && coupon.OrderAvgRange[0] != -1) ||
			(payAvg >= coupon.OrderAvgRange[1] && coupon.OrderAvgRange[1] != -1) {
			return true
		}

		// 盈利比
		var ratio int64
		if user.Money > 0 {
			ratio = (int64(user.CashOut) + user.Diamond) * 100 / int64(user.Money)
		}
		if (ratio < coupon.ProfitabilityRatio[0] && coupon.ProfitabilityRatio[0] != -1) ||
			(ratio >= coupon.ProfitabilityRatio[1] && coupon.ProfitabilityRatio[1] != -1) {
			return true
		}

		for i := 0; i < int(coupon.Num); i++ {
			bean := &data.UserCoupon{
				Id:          bson.NewObjectId().String(),
				UserId:      userid,
				CouponId:    coupon.Id,
				Amount:      coupon.Discount,
				MinRecharge: coupon.MinRecharge,
				Status:      false,
				OverTime:    time.Now().Unix() + int64(coupon.ValidityTime*60*60),
				Ctime:       time.Now().Unix(),
			}
			list = append(list, bean)
			bean.Save()
		}

		// 统计次数
		coupon.IncSendCount()
		user.CouponIssueTime[coupon.Id] = time.Now().Unix()
		return true
	})

	return list
}

func (a *RoleActor) SubCoupon(userid, couponId string) bool {
	user := a.getUserById(userid)
	if user == nil {
		glog.Errorf("SubCoupon user not found, userid:%s, couponId:%s", userid, couponId)
		return false
	}
	list := make([]*data.UserCoupon, 0)

	data.ListByQLimit(data.UserCoupons, bson.M{
		"user_id":   userid,
		"coupon_id": couponId,
		"status":    false,
	}, &list, 1)
	if len(list) == 0 {
		return false
	}

	coupon := list[0]
	coupon.Status = true
	coupon.Utime = time.Now().Unix()
	data.Update(data.UserCoupons, bson.M{
		"_id": coupon.Id,
	}, bson.M{
		"$set": bson.M{
			"status": true,
			"utime":  coupon.Utime,
		},
	})
	coupon.Sync() // 同步ck

	// 手动发的券要记录
	if strings.HasPrefix(coupon.CouponId, "m") {
		data.Update(data.ManualCoupons, bson.M{
			"_id": coupon.CouponId,
		}, bson.M{
			"$inc": bson.M{"used_count": 1},
		})
	}
	return true
}

func (a *RoleActor) GetCoupon(ctx actor.Context) {
	arg := ctx.Message().(*pb.GetCoupon)
	glog.Debugf("GetCoupon %#v", arg)

	rsp := new(pb.GetCouponed)
	defer ctx.Respond(rsp)

	coupon := data.GetOneCoupon(bson.M{
		"coupon_id": arg.CouponId,
		"status":    false,
		"over_time": bson.M{
			"$gt": time.Now().Unix(),
		},
	})
	if coupon == nil {
		return
	}
	couponMap := config.GetManualCouponMap()
	if strings.HasPrefix(coupon.CouponId, "m") && couponMap != nil {
		if _, ok := couponMap.Load(coupon.CouponId); !ok {
			// 没开
			return
		}
	}

	rsp.CouponId = coupon.CouponId
	rsp.CouponAmount = coupon.Amount
	rsp.MinAmount = coupon.MinRecharge
}

func (a *RoleActor) SeeCouponLog(ctx actor.Context) {
	arg := ctx.Message().(*pb.SeeCouponLog)
	glog.Debugf("SeeCouponLog %#v", arg)
	data.SeeCouponUserInc(arg.CouponIds)
}
