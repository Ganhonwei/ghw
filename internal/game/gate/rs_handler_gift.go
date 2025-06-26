package gate

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"sort"
	"time"
)

func (rs *RoleActor) giftPopReset() {
	if rs.GiftPopMap == nil {
		rs.GiftPopMap = make(map[string]data.GiftPop)
	}

	for k, p := range rs.GiftPopMap {
		p.TodayPop = 0
		rs.GiftPopMap[k] = p
	}

	rs.User.SkipGiftWelfare = false
}

func (rs *RoleActor) checkGiftPop() {
	// 破冰解提
	now := utils.BsonNow().Unix()
	if rs.Money <= 0 {
		if rs.GiftPopMap == nil {
			rs.GiftPopMap = make(map[string]data.GiftPop)
		}

		pop := false
		bean := table.GetTables().GiftPopTable.Get("tcpb1")
		if int64(bean.CarryAmount) <= rs.Diamond {
			if p, ok := rs.GiftPopMap["tcpb1"]; ok {
				if p.TodayPop < bean.DailyPop && now-p.LastPopTime > int64(bean.CoolDown*60) {
					pop = true
					p.LastPopTime = now
					p.TodayPop++
					rs.GiftPopMap["tcpb1"] = p
				}
			} else {
				rs.GiftPopMap["tcpb1"] = data.GiftPop{
					Id:          "tcpb1",
					LastPopTime: now,
					TodayPop:    1,
				}
				pop = true
			}
		}
		if pop {
			gift := table.GetTables().GiftRechargeTable.Get(bean.GiftId[0])
			if len(gift.Show) > 0 && gift.Show[rs.RegistArea] == 0 {
				return
			}
			ntf := &pb.GiftWithdrawPopNtf{
				Price: gift.Price,
				Id:    gift.Id,
			}
			rs.Send(ntf)
			glog.Infof("user %s trigger pbjt gift", rs.Userid)
		}
	}
}

// 破冰高送
func (rs *RoleActor) pbGift() {

	// 0 首冲
	// 1 二充
	// 2 三充
	if rs.ChargeTimes > 2 {
		rs.Send(&pb.HideSomethingNtf{Type: 1})
		return
	}

	if rs.GiftPopMap == nil {
		rs.GiftPopMap = make(map[string]data.GiftPop)
	}

	pop := false
	nowTime := utils.BsonNow().Unix()

	key := fmt.Sprintf("tcpb%d", rs.ChargeTimes+2)

	bean := table.GetTables().GiftPopTable.Get(key)

	//判断携带金额
	if bean.CarryAmount > 0 && int64(bean.CarryAmount) > rs.Diamond {
		return
	}

	if p, ok := rs.GiftPopMap[key]; ok {
		//判断是否过期
		if bean.LimitedTime > 0 && nowTime-p.EntryTime >= int64(bean.LimitedTime*60) {
			return
		}
		pop = true
		//弹出次数足够并且过了冷却时间再弹出
		// if p.TodayPop < bean.DailyPop && nowTime-p.LastPopTime > int64(bean.CoolDown*60) {
		// 	pop = true
		// 	p.LastPopTime = nowTime
		// 	p.TodayPop++
		// 	rs.GiftPopMap[key] = p
		// }
	} else {
		pop = true
		rs.GiftPopMap[key] = data.GiftPop{
			Id:          key,
			LastPopTime: nowTime,
			EntryTime:   nowTime,
			TodayPop:    1,
		}
	}

	var gtype int32
	if rs.ChargeTimes == 0 {
		gtype = 1
	} else if rs.ChargeTimes == 1 {
		gtype = 5
	} else if rs.ChargeTimes == 2 {
		gtype = 6
	}

	if pop {
		g := rs.GiftPopMap[key]
		ntf := &pb.GiftWelfareNtf{Gtype: gtype, OverTime: g.EntryTime + int64(bean.LimitedTime*60)}

		gifts := make([]*pb.GiftWelfare, 0)
		for _, g := range bean.GiftId {
			gift := table.GetTables().GiftRechargeTable.Get(g)
			if gift == nil || gift.GiftType != gtype || gift.Show[rs.RegistArea] == 0 {
				continue
			}
			var giveRatio int32
			for _, v := range gift.GiveRatio[rs.RegistArea].Nums {
				giveRatio += v
			}
			b := &pb.GiftWelfare{
				Id:     gift.Id,
				Price:  gift.Price,
				Give:   giveRatio,
				Weight: gift.ShowWeight[rs.RegistArea],
			}
			gifts = append(gifts, b)
		}
		sort.Slice(gifts, func(i, j int) bool {
			return gifts[i].Weight < gifts[j].Weight
		})

		// 过滤支付渠道支持商品
		gifts, payMin, payMax := handleMatchPayShops(gifts,
			func(s *pb.GiftWelfare) int64 { return int64(s.Price) },
			func(s *pb.GiftWelfare, options []int32) { s.PaymentOptions = options },
			func(s *pb.GiftWelfare, apps []int32) { s.PaymentApps = apps })
		ntf.Gifts = gifts
		ntf.MinRecharge = int32(payMin)
		ntf.MaxRecharge = int32(payMax)

		ntf.Skip = rs.User.SkipGiftWelfare
		rs.Send(ntf)
		glog.Infof("user %s trigger pbgs gift", rs.Userid)
	}
}

// 破产礼包
func (rs *RoleActor) breakingGift() {
	if rs.GiftPopMap == nil {
		rs.GiftPopMap = make(map[string]data.GiftPop)
	}

	pop := false
	nowTime := utils.BsonNow().Unix()

	bean := table.GetTables().GiftPopTable.Get("tcpc1")

	if bean.CarryAmount > 0 && int64(bean.CarryAmount) <= rs.Diamond {
		return
	}
	if p, ok := rs.GiftPopMap["tcpc1"]; ok {
		if bean.LimitedTime > 0 && nowTime-p.EntryTime >= int64(bean.LimitedTime*60) {
			// 过期了
			return
		}
		if p.TodayPop < bean.DailyPop && nowTime-p.LastPopTime > int64(bean.CoolDown*60) {
			pop = true
			p.LastPopTime = nowTime
			p.TodayPop++
			rs.GiftPopMap["tcpc1"] = p
		}
	} else {
		pop = true
		rs.GiftPopMap["tcpc1"] = data.GiftPop{
			Id:          "tcpc1",
			LastPopTime: nowTime,
			EntryTime:   nowTime,
			TodayPop:    1,
		}
	}

	if pop {
		// 破产礼包触发前检查是否触发波动返水
		r, err := rs.rolePid.RequestFuture(&pb.VolatilitySubsidyCheck{Userid: rs.Userid}, 3*time.Second).Result()
		if err != nil {
			glog.Error(err)
		} else {
			rsp, ok := r.(*pb.VolatilitySubsidyChecked)
			if !ok {
				glog.Errorf("type error: %#v", r)
			} else if rsp.Trigger {
				glog.Infof("user %s trigger pbgs gift, also trigger volatility subsidy", rs.Userid)
				return
			}
		}

		ntf := &pb.GiftWelfareNtf{Gtype: 2}
		gifts := make([]*pb.GiftWelfare, 0)
		for _, g := range bean.GiftId {
			gift := table.GetTables().GiftRechargeTable.Get(g)
			if gift == nil || gift.GiftType != 4 || gift.Show[rs.RegistArea] == 0 {
				continue
			}
			var giveRatio int32
			for _, v := range gift.GiveRatio[rs.RegistArea].Nums {
				giveRatio += v
			}
			b := &pb.GiftWelfare{
				Id:     gift.Id,
				Price:  gift.Price,
				Give:   giveRatio,
				Weight: gift.ShowWeight[rs.RegistArea],
			}
			gifts = append(gifts, b)
		}
		sort.Slice(gifts, func(i, j int) bool {
			return gifts[i].Weight > gifts[j].Weight
		})

		gifts, payMin, payMax := handleMatchPayShops(gifts,
			func(s *pb.GiftWelfare) int64 { return int64(s.Price) },
			func(s *pb.GiftWelfare, options []int32) { s.PaymentOptions = options },
			func(s *pb.GiftWelfare, apps []int32) { s.PaymentApps = apps })

		ntf.Gifts = gifts
		ntf.MinRecharge = int32(payMin)
		ntf.MaxRecharge = int32(payMax)
		rs.Send(ntf)
		glog.Infof("user %s trigger pbgs gift", rs.Userid)
	}
}
