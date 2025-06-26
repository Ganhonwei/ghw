package data

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"

	"github.com/globalsign/mgo/bson"
)

const (
	//物品类型
	DIAMOND uint32 = 1 //钻石
	COIN    uint32 = 2 //金币
	CARD    uint32 = 3 //房卡
	CHIP    uint32 = 4 //筹码
	VIP     uint32 = 5 //VIP
	//支付方式
	RMB uint32 = 1 //人民币
	DIA uint32 = 2 //钻石
)

// 商城
type Shop struct {
	Id               string  `bson:"_id" json:"id"`                              //购买ID
	NewbieGive       int64   `bson:"newbie_give" json:"newbie_give"`             //新手赠送
	RechargeInterval []int64 `bson:"recharge_interval" json:"recharge_interval"` //充值区间
	WithdrawInterval []int64 `json:"withdraw_interval" bson:"withdraw_interval"` //提现区间
	DefaultRecharge  []int64 `json:"default_recharge" bson:"default_recharge"`   //最低充值区间
	DefaultWithdraw  []int64 `bson:"default_withdraw" json:"default_withdraw"`   //最低提现区间
	VB               []int64 `bson:"vb_give" json:"vb_give"`                     //vb释放
}

// 商城充值
type Buy struct {
	Id   string //购买ID
	Give bool   //是否赠送
}

// 自定义充值
type CustomShop struct {
	Amount uint32 //金额
}

// 商店罐子
type ShopPot struct {
	Id        string // 商品id
	Number    int64  // 数量
	FlowWater int64  // 需要打码量
	OverTime  int64  // 过期时间
	CTime     int64  // 开始时间
}

func GetShopList() []Shop {
	var list []Shop
	ListByQ(Shops, nil, &list)
	return list
}

// Save 写入数据库
func (t *Shop) Save() bool {
	//t.Ctime = bson.Now()
	return Upsert(Shops, bson.M{"_id": t.Id}, t)
}

// Save 写入数据库
func (t *Shop) SaveUpdate() bool {
	//t.Ctime = bson.Now()
	return Upsert(Shops, bson.M{"_id": t.Id}, t)
}

// Save 写入数据库
func DelShop(k any) bool {
	//t.Ctime = bson.Now()
	return Delete(Shops, bson.M{"_id": k})
}

// Save 写入数据库
func ClearShop() bool {
	//t.Ctime = bson.Now()
	return DeleteAll(Shops, bson.M{})
}

// 判断能不能买
func (s *Buy) CanOrder(user *User, id string, ctype int32) bool {
	shop := table.GetTables().ShopTable.Get(s.Id)
	return shop != nil
}

// 充值之后
func (s *Buy) RechargeAfter(user *User, id string, ctype int32, save bool) (*pb.PaySuccessNtf, error) {
	glog.Infof("user %s recharge success, shop:%s", user.Userid, id)
	rsp := new(pb.PaySuccessNtf)
	rsp.Id = id
	rsp.Rtype = pb.SHOP
	return rsp, nil
}

// 完善订单
func (s *Buy) BuildPayOrder(user *User, order *PayOrder, game bool) {
	order.ShopId = utils.String(s.Id)

	shop := table.GetTables().ShopTable.Get(s.Id)
	order.Amount = uint32(shop.Price)
	order.Score = uint32(shop.Price)
	order.ShopName = fmt.Sprint("商城充值:", order.Amount/100)
	order.ShopType = SHOP

	// 赠送
	giveRatio := shop.GiveRatio[user.RegistArea].Nums
	if s.Give {
		order.OtherPresent = fmt.Sprintf("%d", int64(shop.Price)*int64(giveRatio[0])/10000)
		order.GiveCash = int64(shop.Price) * int64(giveRatio[1]) / 10000
		order.GiveWithdrawal = int64(shop.Price) * int64(giveRatio[2]) / 10000
	}
}

func (s *Buy) RechageLType() int32 {
	return int32(pb.LOG_TYPE74)
}

// ============================================== 自定义充值 ==============================================
// 判断能不能买
func (s *CustomShop) CanOrder(user *User, id string, ctype int32) bool {
	// if s.Amount < 20000 || s.Amount > 2000000 {
	// 	glog.Errorf("user %s custom recharge fail, price %d", user.Userid, s.Amount)
	// 	return false
	// }
	// if s.Amount%1000 != 0 {
	// 	glog.Errorf("user %s custom recharge fail, price %d", user.Userid, s.Amount)
	// 	return false
	// }
	if s.Amount <= 0 {
		glog.Errorf("user %s custom recharge fail, price %d", user.Userid, s.Amount)
		return false
	}
	return true
}

// 充值之后
func (s *CustomShop) RechargeAfter(user *User, id string, ctype int32, save bool) (*pb.PaySuccessNtf, error) {
	glog.Infof("user %s recharge success, shop:%s", user.Userid, id)
	rsp := new(pb.PaySuccessNtf)
	rsp.Id = id
	rsp.Rtype = pb.CUSTOM
	return rsp, nil
}

// 完善订单
func (s *CustomShop) BuildPayOrder(user *User, order *PayOrder, game bool) {
	order.Amount = s.Amount

	order.Score = s.Amount
	order.ShopName = fmt.Sprintf("自定义充值:%d", s.Amount/100)
	order.ShopType = CUSTOM_RECHARGE

	//赠送
	custom := table.GetTables().CustomShopTable.Get()
	giveRatio := custom.GiveRate[user.RegistArea].Nums

	order.OtherPresent = fmt.Sprintf("%d", int64(s.Amount)*int64(giveRatio[0])/10000)
	order.GiveCash = int64(s.Amount) * int64(giveRatio[1]) / 10000
	order.GiveWithdrawal = int64(s.Amount) * int64(giveRatio[2]) / 10000
}

func (s *CustomShop) RechageLType() int32 {
	return int32(pb.LOG_TYPE119)
}
