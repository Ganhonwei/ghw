package data

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data/mq"
	"goserver/pkg/glog"

	"github.com/globalsign/mgo/bson"
)

// 用户优惠券
type UserCoupon struct {
	Id          string `json:"id" bson:"_id"`
	UserId      string `json:"user_id" bson:"user_id"`
	CouponId    string `json:"coupon_id" bson:"coupon_id"`       // 优惠券id
	IsAuto      bool   `json:"is_auto" bson:"is_auto"`           // 是否自动
	Amount      int64  `json:"amount" bson:"amount"`             // 优惠金额
	MinRecharge int64  `json:"min_recharge" bson:"min_recharge"` // 最低充值
	Status      bool   `json:"status" bson:"status"`             // 使用状态
	OrderId     string `json:"order_id" bson:"order_id"`         // 订单id
	OverTime    int64  `json:"over_time" bson:"over_time"`       // 过期时间
	Ctime       int64  `json:"ctime" bson:"ctime"`               // 创建时间
	Utime       int64  `json:"utime" bson:"utime"`               // 使用时间
}

func (u *UserCoupon) Save() bool {
	defer DataProducers.PublishDatas(mq.TopicSyncUserCoupon, []any{u})
	return Insert(UserCoupons, u)
}

func (u *UserCoupon) Sync() {
	DataProducers.PublishDatas(mq.TopicSyncUserCoupon, []any{u})
}

func GetCoupons(query bson.M) []UserCoupon {
	list := make([]UserCoupon, 0)
	ListByQ(UserCoupons, query, &list)
	return list
}

func GetOneCoupon(query bson.M) *UserCoupon {
	list := make([]*UserCoupon, 0)
	ListByQLimit(UserCoupons, query, &list, 1)
	if len(list) > 0 {
		return list[0]
	}
	return nil
}

func GetCouponCount(query bson.M) int {
	return Count(UserCoupons, query)
}

// 手动发放优惠券
type ManualCoupon struct {
	Id                 string   `json:"id" bson:"_id"`
	SendSwitch         bool     `json:"send_switch" bson:"send_switch"`                 // 开关
	MinRecharge        int64    `json:"min_recharge" bson:"min_recharge"`               // 最低充值
	Discount           int64    `json:"discount" bson:"discount"`                       // 减免金额
	ValidityTime       int32    `json:"validity_time" bson:"validity_time"`             // 有效期
	Num                int32    `json:"num" bson:"num"`                                 // 人均张数
	OrderAvgRange      []int64  `json:"order_avg_range" bson:"order_avg_range"`         // 订单平均金额
	ProfitabilityRatio []int64  `json:"profitability_ratio" bson:"profitability_ratio"` // 盈利比例
	RecharClassify     []int    `json:"rechar_classify" bson:"rechar_classify"`         // 充值分类
	AcctType           []int    `json:"acct_type" bson:"acct_type"`                     // 账户类型
	UserTag            []int64  `json:"user_tag" bson:"user_tag"`                       // 用户标签
	UserIds            []string `json:"user_ids" bson:"user_ids"`                       // 用户ID
	SendTime           int64    `json:"send_time" bson:"send_time"`                     // 发送时间
	UsedTime           int64    `json:"used_time" bson:"used_time"`                     // 用完时间
	Ctime              int64    `json:"ctime" bson:"ctime"`                             // 创建时间
	SendCount          int64    `json:"send_count" bson:"send_count"`                   // 发送数量
	SendUserCount      int64    `json:"send_user_count" bson:"send_user_count"`         // 发送用户数量
	SeeUserCount       int64    `json:"see_user_count" bson:"see_user_count"`           // 查看用户数量
	UsedCount          int64    `json:"used_count" bson:"used_count"`                   // 已使用数量
	UsedUserCount      int64    `json:"used_user_count" bson:"used_user_count"`         // 已使用用户数量
}

func GetManualCouponById(id string) *ManualCoupon {
	c := new(ManualCoupon)
	Get(ManualCoupons, id, &c)
	return c
}

func GetManualCoupons(q bson.M) []*ManualCoupon {
	list := make([]*ManualCoupon, 0)
	ListByQ(ManualCoupons, q, &list)
	return list
}

func (m *ManualCoupon) IncSendCount() {
	Update(ManualCoupons, bson.M{"_id": m.Id}, bson.M{"$inc": bson.M{"send_count": m.Num, "send_user_count": 1}})
}

func SeeCouponUserInc(ids []string) {
	UpdateAll(ManualCoupons, bson.M{"_id": bson.M{"$in": ids}}, bson.M{"$inc": bson.M{"see_user_count": 1}})
}

type CouponPay struct {
	Id           string `json:"id"`            // 优惠券id
	Amount       int64  `json:"amount"`        // 订单金额
	MinRecharge  int64  `json:"min_recharge"`  // 最低充值
	CouponAmount int64  `json:"coupon_amount"` // 优惠券金额
}

// 优惠券支付
func (c *CouponPay) CanOrder(user *User, id string, ctype int32) bool {
	if c.MinRecharge > 0 && c.Amount >= c.MinRecharge {
		return true
	}
	return false
}

// 充值之后
func (s *CouponPay) RechargeAfter(user *User, id string, ctype int32, save bool) (*pb.PaySuccessNtf, error) {
	glog.Infof("user %s recharge success, shop:%s", user.Userid, id)
	rsp := new(pb.PaySuccessNtf)
	rsp.Id = id
	rsp.Rtype = pb.COUPON
	return rsp, nil
}

// 完善订单
func (s *CouponPay) BuildPayOrder(user *User, order *PayOrder, game bool) {
	order.ShopId = s.Id

	order.Amount = uint32(s.Amount - s.CouponAmount)
	order.Score = uint32(s.Amount - s.CouponAmount)
	order.GiveCash = s.CouponAmount
	order.ShopName = fmt.Sprint("优惠券充值:", order.Amount/100)
	order.ShopType = COUPON
}

func (s *CouponPay) RechageLType() int32 {
	return int32(pb.LOG_TYPE152)
}
