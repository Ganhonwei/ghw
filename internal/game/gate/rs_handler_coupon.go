package gate

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"strings"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (rs *RoleActor) CouponReq() {
	rsp, err := rs.rolePid.RequestFuture(&pb.CouponList{Userid: rs.Userid}, 2*time.Second).Result()
	if err != nil {
		rs.Send(&pb.CouponRsp{Error: pb.Failed})
		return
	}
	rs.Send(rsp)
}

// 自动发放优惠券
func (rs *RoleActor) autoSendCoupon() {
	user := rs.User

	// 充值均单价
	var payAvg int64
	if user.ChargeTimes > 0 {
		payAvg = int64(user.Money) / 100 / int64(user.ChargeTimes)
	}
	// if payAvg == 0 && user.Money > 0 {
	// 	r := make(map[string]any)
	// 	err := ck.Select(&r, `
	// 		SELECT AVG(amount) amount_avg FROM game.col_trade_record FINAL WHERE userid = ? AND order_status = 4
	// 		`, user.Userid)
	// 	if err != nil {
	// 		glog.Error(err)
	// 	} else {
	// 		payAvg = utils.ToInt64(r["amount_avg"])
	// 	}
	// }
	// payAvg = payAvg / 100

	couponIds := make([]string, 0)
	tables := table.GetTables().AutoCouponTable.GetDataList()
	for _, t := range tables {
		// 开关
		if t.AutoSwitch[user.RegistArea] == 0 {
			continue
		}

		// 支付均值
		if (payAvg < int64(t.OrderAvgRange[user.RegistArea].Nums[0]) && int64(t.OrderAvgRange[user.RegistArea].Nums[0]) != -1) ||
			(payAvg > int64(t.OrderAvgRange[user.RegistArea].Nums[1]) && int64(t.OrderAvgRange[user.RegistArea].Nums[1]) != -1) {
			continue
		}

		// 盈利比
		var ratio int64 = 0
		min := t.ProfitabilityRatioRange[user.RegistArea].Nums[0]
		max := t.ProfitabilityRatioRange[user.RegistArea].Nums[1]
		if user.Money > 0 {
			ratio = (int64(user.CashOut) + user.Diamond) * 10000 / int64(user.Money)
		}
		if (ratio < int64(min) && min != -1) || (ratio > int64(max) && max != -1) {
			continue
		}

		// 冷却时间检查
		if user.CouponIssueTime == nil {
			user.CouponIssueTime = make(map[string]int64)
		}
		lastIssueTime := user.CouponIssueTime[t.Id] // 上次发放时间
		if (time.Now().Unix()-lastIssueTime)/60 < int64(t.Cooldown[user.RegistArea]) {
			// 冷却中
			continue
		}

		glog.Infof("auto send coupon, userid: %s, couponid: %s", user.Userid, t.Id)
		couponIds = append(couponIds, t.Id)
		user.CouponIssueTime[t.Id] = time.Now().Unix()
	}
	// 发放优惠券
	rs.rolePid.Tell(&pb.IssueCoupon{
		Userid:   user.Userid,
		CouponId: couponIds,
	})
}

func (rs *RoleActor) IssueCouponNtf(ctx actor.Context) {
	arg := ctx.Message().(*pb.IssueCouponNtf)
	if rs.CouponIssueTime == nil {
		rs.CouponIssueTime = make(map[string]int64)
	}
	for _, coupon := range arg.Coupon {
		if strings.HasPrefix(coupon.Id, "m") {
			// 后台配置的券手动加一下
			rs.CouponIssueTime[coupon.Id] = 0
			rs.status = true
		}
	}
	rs.Send(arg)
}

func (rs *RoleActor) SeeCouponReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.SeeCouponReq)
	glog.Debugf("SeeCouponReq %#v", arg)
	manualMap := make(map[string]struct{})
	for _, id := range arg.CouponIds {
		if strings.HasPrefix(id, "m") {
			manualMap[id] = struct{}{}
		}
	}
	msg := new(pb.SeeCouponLog)
	for k := range manualMap {
		msg.CouponIds = append(msg.CouponIds, k)
	}
	if len(msg.CouponIds) > 0 {
		rs.rolePid.Tell(msg)
	}
	rs.Send(new(pb.SeeCouponRsp))
}
