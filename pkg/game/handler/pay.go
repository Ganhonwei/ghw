package handler

import (
	"bytes"
	"encoding/json"
	"fmt"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"io"
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var seed int32 = 1
var mutex sync.Mutex

// 生成订单
func CreateOrder(user *data.User, entity config.IPay, game bool, gameId string, payWay, payOption, payApp int32) (*data.PayOrder, error) {
	// channelId := getChannelId(true)
	order := &data.PayOrder{
		Userid:       user.Userid,
		NickName:     user.Nickname,
		RealName:     user.Nickname,
		Email:        fmt.Sprintf("%s@gmail.com", user.Userid),
		Mobile:       "7112223333",
		RegistIp:     user.RegistIP,
		OrderStatus:  data.Tradeing,
		Ctime:        utils.BsonNow(),
		OrderAddress: user.LoginIP,
		PackageId:    user.AD_BundleId,
		Gtype:        user.LastGType,
		UserType:     user.RegistArea,
		PayWay:       payWay,
		PayOption:    payOption,
		PayApp:       payApp,
	}
	// 根据不同的对象生成订单
	entity.BuildPayOrder(user, order, game)
	// 局内充值
	bean := config.GetGame(gameId)
	if bean.Id != "" {
		order.Gtype = bean.Gtype
	}

	// if order.ShopId == "" {
	// 	glog.Errorf("user %s not found shop id:%s", user.Userid, order.ShopId)
	// 	return nil, fmt.Errorf("not found shop id:%s", order.ShopId)
	// }
	// 生成订单号
	order.OrderID = fmt.Sprint(generateOrderId(uint32(utils.BsonNow().Year())))
	glog.Infof("user create order: user=%s, order=%s, option=%d, app=%d", order.OrderID, user.Userid, order.PayOption, order.PayApp)
	return order, nil
}

// 生成提现订单
func CreateWithDrawOrder(user *data.User, arg *pb.PayWithDrawReq) (*data.WithdrawOrder, error) {
	// channelId := getChannelId(false)
	// bean := config.GetWithDraw(arg.Id)
	order := &data.WithdrawOrder{
		Userid:       user.Userid,
		NickName:     user.Nickname,
		RealName:     user.RealName,
		Email:        fmt.Sprintf("%s@gmail.com", user.Userid),
		Mobile:       user.Phone,
		OrderStatus:  data.Withdrawing,
		Ctime:        utils.BsonNow(),
		OrderAddress: user.LoginIP,
		PackageId:    user.AD_BundleId,
		ShopId:       arg.Id,
		ShopName:     fmt.Sprint("withdraw ", arg.Amount),
		Score:        uint32(arg.Amount),
		Amount:       uint32(arg.Amount),
		// Commission:   bean.Commission,
		PayWay:     arg.PayWay,
		Bank:       user.Bank,
		BankNumber: user.BankAccounts,
		IFSC:       user.IFSC,
		// OutChannel:   utils.String(channelId),
		ExamineWay: data.HandExamineWay,
		UserType:   user.RegistArea,
		Diamond:    user.Diamond + user.GiveDiamond,
	}
	// ustd支付
	if order.PayWay == 1 {
		order.BankNumber = user.USDT
		order.Bank = ""
		order.IFSC = ""
	}
	// 手续费
	vip := table.GetTables().VipTable.Get(int32(user.Vip.Lv))
	order.Commission = uint32(arg.Amount) * uint32(vip.FeeRate) / 10000
	// order.Amount = bean.CostDiamond - order.Commission

	// C类10%手续费
	if user.RegistArea == 2 {
		order.Commission = uint32(float64(arg.Amount) * 0.1)
		// order.Amount = uint32(float64(bean.CostDiamond) * 0.9)
	}
	// if order.ShopId == 0 {
	// 	glog.Errorf("user %s not found shop id:%d", user.Userid, order.ShopId)
	// 	return nil, fmt.Errorf("not found shop id:%d", order.ShopId)
	// }
	// 生成订单号
	order.OrderID = fmt.Sprint(generateOrderId(uint32(utils.BsonNow().Year())))
	glog.Infof("user %s create order %s", order.OrderID, user.Userid)
	return order, nil
}

// 生成巴基斯坦提现订单
func CreatePKWithDrawOrder(user *data.User, arg *pb.PakWithDrawReq) (*data.WithdrawOrder, error) {
	// channelId := getChannelId(false)
	// bean := config.GetWithDraw(arg.Id)
	order := &data.WithdrawOrder{
		Userid:       user.Userid,
		NickName:     user.Nickname,
		RealName:     arg.BankHolder,
		Email:        fmt.Sprintf("%s@gmail.com", user.Userid),
		Mobile:       user.Phone2,
		OrderStatus:  data.Withdrawing,
		Ctime:        utils.BsonNow(),
		OrderAddress: user.LoginIP,
		PackageId:    user.AD_BundleId,
		ShopId:       arg.Id,
		ShopName:     fmt.Sprint("withdraw ", arg.Amount),
		Score:        uint32(arg.Amount),
		Amount:       uint32(arg.Amount),
		// Commission:   bean.Commission,
		Bank:       arg.BankName,
		BankNumber: arg.BankNumber,
		// OutChannel:   utils.String(channelId),
		ExamineWay: data.HandExamineWay,
		UserType:   user.RegistArea,
		Diamond:    user.Diamond + user.GiveDiamond,
	}
	// ustd支付
	if order.PayWay == 1 {
		order.BankNumber = user.USDT
		order.Bank = ""
		order.IFSC = ""
	}
	// 手续费
	vip := table.GetTables().VipTable.Get(int32(user.Vip.Lv))
	order.Commission = uint32(arg.Amount) * uint32(vip.FeeRate) / 10000
	// order.Amount = bean.CostDiamond - order.Commission

	// C类10%手续费
	if user.RegistArea == 2 {
		order.Commission = uint32(float64(arg.Amount) * 0.1)
		// order.Amount = uint32(float64(bean.CostDiamond) * 0.9)
	}
	// if order.ShopId == 0 {
	// 	glog.Errorf("user %s not found shop id:%d", user.Userid, order.ShopId)
	// 	return nil, fmt.Errorf("not found shop id:%d", order.ShopId)
	// }
	// 生成订单号
	order.OrderID = fmt.Sprint(generateOrderId(uint32(utils.BsonNow().Year())))
	glog.Infof("user %s create order %s", order.OrderID, user.Userid)
	return order, nil
}

// 获取渠道
func getChannelId(pay bool) uint32 {
	// return 3014
	var choices []int
	if pay { // 代收
		choices = config.Channels.NowPayChannel
	} else { // 代付
		choices = config.Channels.NowWithdrawChannel
	}
	if len(choices) > 0 {
		c, _ := utils.ChoiceInt(choices)
		return uint32(c)
	}
	c, _ := utils.ChoiceInt([]int{3011, 3014, 3015})
	return uint32(c)
}

// 生成订单号 渠道16位|时间32位|序列号16
func generateOrderId(channelId uint32) int64 {
	defer mutex.Unlock()
	mutex.Lock()
	now := time.Now().UnixMilli()
	seed += 1
	if seed >= (1<<31 - 1) {
		seed = 1
	}
	return (int64(channelId&0xFFFF) << 43) | ((now << 17) | int64(seed&0xFFFF))
}

func GenerateOrderId(channelId uint32) int64 {
	return generateOrderId(channelId)
}

func GetPayMent(id string, rtype pb.RechargeType) config.IPay {
	switch rtype {
	case pb.SHOP, pb.SHOPNOGIVE:
		shop := table.GetTables().ShopTable.Get(id)
		if shop == nil {
			return nil
		}
		return &data.Buy{
			Id:   id,
			Give: rtype == pb.SHOP,
		}
	case pb.ACTIVITY:
		act := config.GetActivity(id)
		return &act
	case pb.WEEKCARD:
		wid := fmt.Sprintf("lbzk%s", id)
		act := table.GetTables().GiftRechargeTable.Get(wid)
		if act == nil {
			return nil
		}
		cardId, _ := strconv.Atoi(id)
		return &data.WeeklyCard{ID: int32(cardId), Price: act.Price, Reward: act.Price}
	case pb.GAMERECHARGE:
		return &data.DeskRecharge{}
	case pb.LIMITEDGIFT:
		// wid, _ := strconv.Atoi(id)
		// act := config.GetLimitedGift(id)
		bean := table.GetTables().GiftRechargeTable.Get(id)
		if bean == nil {
			return nil
		}
		return &data.CommonGiftRecharge{
			Id:    id,
			Price: bean.Price,
			Gtype: bean.GiftType,
		}
	case pb.BREAKING:
		return &data.BreakingGiftPay{}
	case pb.SIGN:
		wid, _ := strconv.Atoi(id)
		act := config.GetDailySign(int32(wid))
		return &act
	case pb.CUSTOM:
		return &data.CustomShop{}
	case pb.COUPON:
		return &data.CouponPay{}
	}
	return nil
}

// func GetPayPartenr(channelId uint32) pay.IPay {
// 	switch channelId {
// 	case 3011:
// 		return mlpay.Config
// 	case 3014:
// 		return kingpay.Config
// 	case 3015:
// 		return xfpay.Config
// 	}
// 	return nil
// }

// 能不能下提现订单
func CanWithDraw(user *data.User, amount int32) (bool, pb.ErrCode, []string) {
	// bean := config.GetWithDraw(id)
	// if bean.Id == 0 {
	// 	glog.Errorf("no found withdraw config id:%d", id)
	// 	return false, pb.WithdrawOrderError, nil
	// }
	// count := user.WithDrawCountMap[id]
	// if count >= bean.Limit {
	// 	// 提现次数不足
	// 	return false, pb.WithdrawCountUnenough, nil
	// }
	shop := table.GetTables().WithdrawConfigTable.Get()
	if shop == nil {
		glog.Errorf("no found withdraw config")
		return false, pb.WithdrawOrderError, nil
	}

	if amount < shop.WithdrawInterval[0] || amount > shop.WithdrawInterval[1] {
		glog.Errorf("withdraw amount err")
		return false, pb.WithdrawOrderError, nil
	}

	vip := table.GetTables().VipTable.Get(int32(user.Vip.Lv))
	if user.Vip.WithdrawCount >= int(vip.WithdrawTimes) && vip.WithdrawTimes != -1 {
		// 提现次数不足
		return false, pb.WithdrawCountUnenough, nil
	}

	if (int64(amount) > vip.WithdrawAmounts ||
		int64(user.Vip.WithdrawAmount) >= vip.WithdrawAmounts ||
		int64(amount)+int64(user.Vip.WithdrawAmount) > vip.WithdrawAmounts) &&
		vip.WithdrawAmounts != -1 {
		// 提现金额不足
		return false, pb.WithdrawAmountUnenough, nil
	}

	// 可提现金额够不够
	if user.OutDiamond < int64(amount) {
		glog.Errorf("OutDiamond is not enough, user:%s, amount:%d", user.Userid, amount)
		return false, pb.NOENGOUTHFLOWWARTER, nil
	}
	// 手续费
	// 钱够不够
	if user.Diamond < int64(amount) {
		glog.Errorf("diamond is not enough, user:%s, amount:%d", user.Userid, amount)
		return false, pb.NotEnoughCoin, nil
	}
	commission := uint32(amount) * uint32(vip.FeeRate) / 10000
	if user.Diamond-int64(amount) < int64(commission) {
		// 手续费不够
		glog.Errorf("commission is not enough, user:%s, amount:%d", user.Userid, amount)
		return false, pb.WithdrawFees, nil
	}

	// 提现解锁(改为vip判断)
	// if user.Money <= 0 || user.Money < uint32(user.WithdrawLock) {
	// 	p := user.WithdrawLock - int32(user.Money)
	// 	if user.Money <= 0 {
	// 		p = CalWithdrawLock(user) - int32(user.Money)
	// 	}
	// 	return false, pb.RECHARGEREQUIRE, []string{utils.String(p)}
	// }
	// 玩家状态,非正常状态玩家不能提现
	// if user.State == data.NoveiceState || user.State == data.ExceptionState {
	// 	glog.Errorf("user state exception, user:%s, state:%d, id:%d", user.Userid, user.State, id)
	// 	return false, pb.RECHARGEREQUIRE
	// }
	return true, pb.OK, nil
}

// 获取流水类型
func GetRechargeLtype(id string, shoptype int32) int32 {
	ment := GetPayMent(id, pb.ACTIVITY)
	if shoptype == data.SHOP {
		ment = GetPayMent(id, pb.SHOP)
	} else if shoptype == data.GAME_RECHARGE {
		return int32(pb.LOG_TYPE80)
	} else if shoptype == data.LIMITED_GIFT {
		ment = GetPayMent(id, pb.LIMITEDGIFT)
	} else if shoptype == data.METAL_CARD {
		ment = GetPayMent(id, pb.WEEKCARD)
	} else if shoptype == data.BREAKING_GITF {
		ment = GetPayMent(id, pb.BREAKING)
	} else if shoptype == data.SIGN_RECHARGE {
		ment = GetPayMent(id, pb.SIGN)
	} else if shoptype == data.CUSTOM_RECHARGE {
		ment = GetPayMent(id, pb.CUSTOM)
	}
	return ment.RechageLType()
}

func RechargeInterceptor(phone string) bool {
	// 开关
	if !config.SettingIsOpen(4, data.PAYBLACK) {
		return false
	}
	m := config.GetPhoneBlackMap()
	if _, ok := m[phone]; ok {
		glog.Infof("this phone in black list, %s", phone)
		return true
	}
	return false
}

func SubmitOrder(url string, body []byte, pay bool) (any, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		glog.Errorf("err:%s", err)
		return nil, err
		// return []byte(""), err
	}
	req.Header.Add("Content-type", "application/json")

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   0,
			KeepAlive: 0,
		}).Dial,
	}

	client := &http.Client{Transport: transport}

	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("err:%s", err)
		return nil, err
	}

	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("err:%s", err)
		return nil, err
	}
	if pay {
		// 充值
		rsp := new(data.PayResponse)
		err := json.Unmarshal(respData, &rsp)
		if err != nil {
			glog.Errorf("err:%v", err)
			return nil, err
		}
		return rsp, nil
	} else {
		// 充值
		rsp := new(data.WithdrawResponse)
		err := json.Unmarshal(respData, &rsp)
		if err != nil {
			glog.Errorf("err:%s", err)
			return nil, err
		}
		return rsp, nil
	}
}

func SubmitPayNode(url string, body []byte, rsp any) error {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		glog.Errorf("err:%s", err)
		return err
	}
	req.Header.Add("Content-type", "application/json")

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   0,
			KeepAlive: 0,
		}).Dial,
	}

	client := &http.Client{Transport: transport}

	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("err: %s", err)
		return err
	}

	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("err: %s", err)
		return err
	}

	if err := json.Unmarshal(respData, rsp); err != nil {
		glog.Errorf("err: %v", err)
		return err
	}

	return nil
}

func GetBreakingPayMent(user *data.User) config.IPay {
	game := config.GetGame(user.BreakGameId)
	if game.Id == "" {
		return nil
	}
	return &data.BreakingGiftPay{
		Name:  fmt.Sprintf("破产礼包(%s)", pb.GameType_name[user.BreakGtype]),
		Price: uint32(game.BreakingPrice),
		Cash:  uint32(game.BreakingPrice),
		Give:  uint32(game.BreakingGive),
	}
}

// 充值获得的可提现金(C类用户)
func GetChargeWithdrawable(user *data.User, diamond int64, money uint32) int64 {
	if user.RegistArea != 2 {
		return 0
	}
	if diamond <= 0 || money <= 0 {
		return 0
	}

	return int64(math.Min(float64(money), float64(diamond)))
}

func GetChargeType(user *data.User) int {
	// classify := config.GetChargeClassify()
	classify := table.GetTables().ChargeClassifyTable.Get()
	if classify == nil {
		return 0
	}

	money := int64(user.Money)

	if int64(classify.NoCharge[0]) <= money && int64(classify.NoCharge[1]) > money {
		if user.State == data.NoveiceState {
			return data.NewBie
		}
		return data.Civilian
	} else if int64(classify.NormalCharge[0]) <= money && int64(classify.NormalCharge[1]) > money {
		return data.NormalCharge
	} else if int64(classify.Xr[0]) <= money && int64(classify.Xr[1]) > money {
		return data.XR
	} else if int64(classify.Zr[0]) <= money && int64(classify.Zr[1]) > money {
		return data.ZR
	} else if int64(classify.Dr[0]) <= money && int64(classify.Dr[1]) > money {
		return data.DR
	} else {
		return data.CDR
	}
}
