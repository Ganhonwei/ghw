package config

import (
	"goserver/pkg/data"
	"sync"

	"github.com/globalsign/mgo/bson"
)

// 提现配置列表
var ManualCouponMap *sync.Map

func InitCoupon() {
	ManualCouponMap = new(sync.Map)
	coupons := data.GetManualCoupons(bson.M{"send_switch": true})
	for _, coupon := range coupons {
		ManualCouponMap.Store(coupon.Id, coupon)
	}
}

func GetManualCouponMap() *sync.Map {
	return ManualCouponMap
}

func SetManualCoupon(coupon *data.ManualCoupon) {
	if coupon.SendSwitch {
		ManualCouponMap.Store(coupon.Id, coupon)
	} else {
		ManualCouponMap.Delete(coupon.Id)
	}
}

func SetManualCouponMap(id string) {
	coupon := data.GetManualCouponById(id)
	if coupon != nil {
		ManualCouponMap.Store(coupon.Id, coupon)
	}
}
