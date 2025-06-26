package service

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"goserver/internal/global/pay/entity"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/globalsign/mgo/bson"
)

var seed int32 = 1
var mutex sync.Mutex

// 提现配置列表
var WithDrawMap *sync.Map
var PayMap *sync.Map

var (
	firstNames = []string{"John", "Alex", "Emily", "Emma", "Sophia", "Oliver", "James", "David"}
	lastNames  = []string{"Smith", "Johnson", "Brown", "Lee", "Wilson", "Taylor", "Clark", "Robinson"}
)

func InitPayChannel() {
	PayMap = new(sync.Map)
	var list []entity.PayChannel
	ListByQ(PayChannels, nil, &list)
	for _, l := range list {
		PayMap.Store(l.Id, l)
	}
	InitPayHistory(list)
}

func GetPayChannel(id string) entity.PayChannel {
	if v, ok := PayMap.Load(id); ok {
		return v.(entity.PayChannel)
	}
	return entity.PayChannel{}
}

func SavePayLog(p entity.PayChannelLog) {
	Insert(PayChannelLogs, &p)
}

func SetPayChannel(p []entity.PayChannel) {
	// PayMap = new(sync.Map)
	for _, channel := range p {
		if c, ok := PayMap.Load(channel.Id); !ok {
			id, _ := strconv.Atoi(channel.Id)
			PayHistory = append(PayHistory, QueryPayHistory(uint32(id)))
			WithdrawHistory = append(WithdrawHistory, QueryWithdrawHistory(uint32(id)))
		} else {
			if c.(entity.PayChannel).Status == 0 && channel.Status == 1 {
				id, _ := strconv.Atoi(channel.Id)
				PayHistory = append(PayHistory, QueryPayHistory(uint32(id)))
				WithdrawHistory = append(WithdrawHistory, QueryWithdrawHistory(uint32(id)))
			}
		}
		PayMap.Store(channel.Id, channel)
	}
	sort.Slice(PayHistory, func(i, j int) bool {
		return PayHistory[i].SuccessRate > PayHistory[j].SuccessRate
	})
	sort.Slice(WithdrawHistory, func(i, j int) bool {
		return WithdrawHistory[i].SuccessRate > WithdrawHistory[j].SuccessRate
	})
}

func InitPayHistory(list []entity.PayChannel) {
	PayHistory = make([]*PayChannelHistory, 0)
	for _, l := range list {
		id, err := strconv.Atoi(l.Id)
		if err != nil || l.Status == 0 {
			continue
		}
		PayHistory = append(PayHistory, QueryPayHistory(uint32(id)))
	}
	sort.Slice(PayHistory, func(i, j int) bool {
		return PayHistory[i].SuccessRate > PayHistory[j].SuccessRate
	})
	WithdrawHistory = make([]*PayChannelHistory, 0)
	for _, l := range list {
		id, err := strconv.Atoi(l.Id)
		if err != nil || l.Status == 0 {
			continue
		}
		WithdrawHistory = append(WithdrawHistory, QueryWithdrawHistory(uint32(id)))
	}
	sort.Slice(WithdrawHistory, func(i, j int) bool {
		return WithdrawHistory[i].SuccessRate > WithdrawHistory[j].SuccessRate
	})
}

func QueryPayHistory(channelId uint32) *PayChannelHistory {
	list := make([]entity.PayOrder, 0)
	PayOrders.Find(bson.M{"channel_id": channelId}).Sort("-ctime").Limit(100).All(&list)
	// ListByQLimit(PayOrders, bson.M{"channel_id": channelId}, &list, 100)

	history := &PayChannelHistory{
		ChannelId:   channelId,
		SuccessRate: 10000,
	}

	if len(list) <= 0 {
		return history
	}

	success := 0
	for _, l := range list {
		history.Info = append(history.Info, &PayInfo{
			OrderID: l.OrderID,
			Success: l.OrderStatus == 2,
		})
		if l.OrderStatus == 2 {
			success++
		}
	}

	if len(list) < 100 {
		history.SuccessRate = 10000
	} else {
		history.SuccessRate = success * 10000 / len(list)
	}
	return history
}

func QueryWithdrawHistory(channelId uint32) *PayChannelHistory {
	list := make([]entity.PayOrder, 0)
	WithdrawOrders.Find(bson.M{"channel_id": channelId}).Sort("-ctime").Limit(20).All(&list)
	// ListByQLimit(PayOrders, bson.M{"channel_id": channelId}, &list, 100)

	history := &PayChannelHistory{
		ChannelId:   channelId,
		SuccessRate: 10000,
	}

	if len(list) <= 0 {
		return history
	}

	success := 0
	for _, l := range list {
		history.Info = append(history.Info, &PayInfo{
			OrderID: l.OrderID,
			Success: l.OrderStatus == 2,
		})
		if l.OrderStatus == 2 {
			success++
		}
	}

	if len(list) < 100 {
		history.SuccessRate = 10000
	} else {
		history.SuccessRate = success * 10000 / len(list)
	}
	return history
}

// 生成订单
func CreateOrder(user data.PayRequest) (*entity.PayOrder, error) {
	// channelId := GetChannelId(true, nil, user.PayWay)
	channelId := GetPayChannelId(nil, user.PayWay, user.PayOption, user.PayApp, int64(user.Amount))
	phone := generateEmail()
	number := generateNumber()
	timestamp := utils.BsonNow().UnixNano() / 1e6
	order := &entity.PayOrder{
		BusinessID:  user.BusinessID,
		MerOrderID:  user.OrderID,
		Userid:      user.Userid,
		OrderStatus: data.Tradeing,
		Ctime:       timestamp,
		Amount:      user.Amount,
		ChannelId:   channelId,
		NickName:    number,
		RealName:    number,
		Email:       fmt.Sprintf("%s@gmail.com", phone),
		Mobile:      phone,
		RegistIp:    user.RegistIp,
		PayWay:      user.PayWay,
		PayOption:   user.PayOption,
		PayApp:      user.PayApp,
	}

	// 根据不同的对象生成订单
	// pay.BuildPayOrder(order)
	// if order.ShopId == "" {
	// 	glog.Errorf("user %s not found shop id:%s", user.Userid, order.ShopId)
	// 	return nil, fmt.Errorf("not found shop id:%s", order.ShopId)
	// }
	// 生成订单号
	order.OrderID = fmt.Sprint(generateOrderId(channelId))
	glog.Infof("user %s create order %s", order.OrderID, user.Userid)

	err := AddPayOrder(order)
	if err != nil {
		glog.Errorf("user %s create order %s", order.Userid, order.OrderID)
	}
	return order, err
}

// 生成后台测试订单
func CreateWebOrder(user data.WebPayRequest) (*entity.PayOrder, error) {
	phone := generateEmail()
	number := generateNumber()
	timestamp := utils.BsonNow().UnixNano() / 1e6
	order := &entity.PayOrder{
		BusinessID:  user.BusinessID,
		MerOrderID:  user.OrderID,
		Userid:      user.Userid,
		OrderStatus: data.Tradeing,
		Ctime:       timestamp,
		Amount:      user.Amount,
		ChannelId:   user.ChannelId,
		NickName:    number,
		RealName:    number,
		Email:       fmt.Sprintf("%s@gmail.com", phone),
		Mobile:      phone,
		RegistIp:    user.RegistIp,
		PayWay:      0,
	}

	// 根据不同的对象生成订单
	// pay.BuildPayOrder(order)
	// if order.ShopId == "" {
	// 	glog.Errorf("user %s not found shop id:%s", user.Userid, order.ShopId)
	// 	return nil, fmt.Errorf("not found shop id:%s", order.ShopId)
	// }
	// 生成订单号
	order.OrderID = fmt.Sprint(generateOrderId(user.ChannelId))
	glog.Infof("web %s create order %s", order.OrderID, user.Userid)

	err := AddPayOrder(order)
	if err != nil {
		glog.Errorf("user %s create order %s", order.Userid, order.OrderID)
	}
	return order, err
}

// 生成提现订单
func CreateWithDrawOrder(user data.WithdrawRequest) (*entity.WithdrawOrder, error) {
	// channelId := GetChannelId(false, nil, user.PayWay)
	channelId := GetWithdrawChannelId(nil, user.PayWay, int64(user.Amount), user.Country, user.Bank)
	phone := generateEmail()
	timestamp := utils.BsonNow().UnixNano() / 1e6
	name := ""
	if user.RealName == "" {
		name = generateNumber()
	} else {
		name = user.RealName
	}
	order := &entity.WithdrawOrder{
		MerOrderID:  user.OrderID,
		Userid:      user.Userid,
		RealName:    name,
		OrderStatus: data.Withdrawing,
		Ctime:       timestamp,
		Amount:      user.Amount,
		Bank:        user.Bank,
		BankNumber:  user.BankNumber,
		IFSC:        user.IFSC,
		ChannelId:   channelId,
		NickName:    name,
		Email:       fmt.Sprintf("%s@gmail.com", phone),
		Mobile:      phone,
		PayWay:      user.PayWay,
		Country:     user.Country,
	}
	// 生成订单号
	order.OrderID = fmt.Sprint(generateOrderId(channelId))
	glog.Infof("user %s create order %s", order.OrderID, user.Userid)
	err := AddWithDrawOrder(order)
	if err != nil {
		glog.Errorf("user %s create order %s", order.Userid, order.OrderID)
	}
	return order, nil
}

func generateEmail() string {
	rand.Seed(time.Now().UnixNano()) // 设置随机种子

	// 生成以 6-9 开头的十位数
	firstDigit := 6 + rand.Intn(4) // 生成 6-9 之间的随机数作为第一个数字
	number := firstDigit * 10      // 将个位数变为 0，得到十位数
	for i := 1; i < 9; i++ {
		number = number*10 + rand.Intn(10) // 生成 0-9 之间的随机数依次添加到后面
	}

	return fmt.Sprint(number)
}

func GenerateMJLPhone() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 国家代码和孟加拉国的区号
	areaCodes := []string{"0171", "0181", "0191", "0151", "0161", "0175", "0195", "0166", "0185"}

	// 随机选择一个区号
	areaCode := areaCodes[r.Intn(len(areaCodes))]

	// 生成 7 位本地号码
	localNumber := ""
	for i := 0; i < 7; i++ {
		localNumber += strconv.Itoa(r.Intn(10)) // 生成随机数字
	}

	// 组合成完整的电话号码
	return areaCode + localNumber
}

func generateNumber() string {
	// rand.Seed(time.Now().UnixNano())
	// min := 100000
	// max := 999999
	// randomNum := rand.Intn(max-min+1) + min
	rand.Seed(time.Now().UnixNano())

	firstName := firstNames[rand.Intn(len(firstNames))]
	lastName := lastNames[rand.Intn(len(lastNames))]

	fullName := firstName + " " + lastName

	return fmt.Sprint(fullName)
}

// 获取支付渠道
func getPayChannel() ([]entity.PayChannel, error) {
	var list []entity.PayChannel
	m := bson.M{"status": 1}
	err := PayChannels.Find(m).All(&list)
	if err != nil {
		return nil, errors.New("查询支付渠道数据失败:" + err.Error())
	}
	return list, nil
}

// 获取渠道
// Deprecated: use GetPayChannelId and GetWithdrawChannelId
func GetChannelId(pay bool, exclude []string, payWay int32) uint32 {
	if payWay == 1 {
		// 已经提交过
		if len(exclude) > 0 && exclude[0] == strconv.Itoa(int(USDT)) {
			return 0
		}
		return USDT
	}

	// return 5009
	var choices []int
	// var list []entity.PayChannel

	if pay {
		for _, c := range PayHistory {
			if o, ok := PayMap.Load(utils.String(c.ChannelId)); ok {
				payChannel := o.(entity.PayChannel)
				if payChannel.Status == 0 {
					continue
				}
			} else {
				continue
			}
			used := false
			for _, e := range exclude {
				if e == utils.String(c.ChannelId) {
					used = true
					break
				}
			}
			// if used || c.SuccessRate == 0 {
			if used {
				continue
			}
			choices = append(choices, int(c.ChannelId))
			// return c.ChannelId
		}
		// 选择一个渠道
		// if len(choices) != 0 {
		// 	weights := make([]utils.Choice, 0)
		// 	for i, c := range choices {
		// 		pow := math.Pow(2, float64(i+1))
		// 		weights = append(weights, utils.Choice{Weight: int(10000 / pow), Item: c})
		// 	}
		// 	choice, err := utils.WeightedChoice(weights)
		// 	if err == nil {
		// 		return uint32(choice.Item.(int))
		// 	}
		// }
	} else {
		for _, c := range WithdrawHistory {
			if o, ok := PayMap.Load(utils.String(c.ChannelId)); ok {
				payChannel := o.(entity.PayChannel)
				if payChannel.Wstatus == 0 {
					continue
				}
			} else {
				continue
			}
			used := false
			for _, e := range exclude {
				if e == utils.String(c.ChannelId) {
					used = true
					break
				}
			}
			// if used || c.SuccessRate == 0 {
			if used {
				continue
			}
			choices = append(choices, int(c.ChannelId))
			// return c.ChannelId
		}
	}
	// 选择一个渠道
	if len(choices) != 0 {
		weights := make([]utils.Choice, 0)
		for i, c := range choices {
			pow := math.Pow(2, float64(i+1))
			weights = append(weights, utils.Choice{Weight: int(10000 / pow), Item: c})
		}
		choice, err := utils.WeightedChoice(weights)
		if err == nil {
			return uint32(choice.Item.(int))
		}
	}

	// 随机一个渠道
	choices = make([]int, 0)
	PayMap.Range(func(key, value interface{}) bool {
		payChannel := value.(entity.PayChannel)
		if payChannel.Status == 0 && pay {
			return true
		}
		if payChannel.Wstatus == 0 && !pay {
			return true
		}
		for _, e := range exclude {
			if e == payChannel.Id {
				return true
			}
		}
		id, err := strconv.Atoi(payChannel.Id)
		if err == nil {
			choices = append(choices, id)
		}
		return true
	})

	// if pay { // 代收
	// 	if data.Has(PayChannels, bson.M{"status": 1, "_id": utils.String(channel)}) {
	// 		// 指定渠道开着的直接
	// 		return channel
	// 	}
	// 	m := bson.M{"status": 1, "_id": bson.M{"$nin": exclude}}
	// 	PayChannels.Find(m).All(&list)
	// 	if len(list) > 0 {
	// 		for _, item := range list {
	// 			id, err := strconv.Atoi(item.Id)
	// 			if err == nil {
	// 				choices = append(choices, id)
	// 			}
	// 		}
	// 	}
	// 	//choices = config.Channels.NowPayChannel
	// } else { // 代付
	// 	if data.Has(PayChannels, bson.M{"wstatus": 1, "_id": utils.String(channel)}) {
	// 		// 指定渠道开着的直接
	// 		return channel
	// 	}
	// 	m := bson.M{"wstatus": 1, "_id": bson.M{"$nin": exclude}}
	// 	PayChannels.Find(m).All(&list)
	// 	if len(list) > 0 {
	// 		for _, item := range list {
	// 			id, err := strconv.Atoi(item.Id)
	// 			if err == nil {
	// 				choices = append(choices, id)
	// 			}
	// 		}
	// 	}
	// 	//choices = config.Channels.NowWithdrawChannel
	// }
	if len(choices) > 0 {
		c, _ := utils.ChoiceInt(choices)
		return uint32(c)
	}
	if exclude != nil {
		return 0
	}
	return 0
	// c, _ := utils.ChoiceInt([]int{3010, 3011, 3014, 3015, 3016})
	// return uint32(c)
}

// 获取支付渠道
func GetPayChannelId(exclude []string, payWay, payOption, payApp int32, amount int64) uint32 {
	if payWay == 1 {
		// 已经提交过
		if len(exclude) > 0 && exclude[0] == strconv.Itoa(int(USDT)) {
			return 0
		}
		return USDT
	}

	// var choices []int
	var choices []utils.Choice

	PayMap.Range(func(_, c any) (next bool) {
		next = true
		payChannel := c.(entity.PayChannel)
		if payChannel.Status == 0 || payChannel.PayWeight == 0 {
			return
		}
		// 选择支付方式
		if payOption != 0 {
			if payOption == 3 {
				// 通道是否支持utr
				if !payChannel.UtrRequired {
					return
				}
			} else if payChannel.UtrRequired { // utr为true只能通过utr调起
				return
			} else if !utils.SliceIn(payOption, payChannel.PayOptions...) {
				return
			}
		}
		// 选择支付app
		if payApp != 0 {
			if !utils.SliceIn(payApp, payChannel.PayApps...) {
				return
			}
		}
		// 支付金额限制
		if amount < payChannel.PayMin || amount > payChannel.PayMax {
			return
		}
		// 排除使用过的
		if utils.SliceIn(payChannel.Id, exclude...) {
			return
		}

		channelId, err := strconv.Atoi(payChannel.Id)
		if err != nil {
			glog.Errorf("pay channelId to int error: %s, %v", payChannel.Id, err)
			return
		}
		choices = append(choices, utils.Choice{Weight: int(payChannel.PayWeight), Item: channelId})
		return
	})

	// 选择一个渠道
	if len(choices) > 0 {
		choice, err := utils.WeightedChoice(choices)
		if err == nil {
			return uint32(choice.Item.(int))
		}
	}

	return 0
}

// 获取提现渠道
func GetWithdrawChannelId(exclude []string, payWay int32, amount int64, country, bank string) uint32 {
	if payWay == 1 {
		// 已经提交过
		if len(exclude) > 0 && exclude[0] == strconv.Itoa(int(USDT)) {
			return 0
		}
		return USDT
	}

	var choices []utils.Choice
	for _, c := range WithdrawHistory {
		o, ok := PayMap.Load(utils.String(c.ChannelId))
		if !ok {
			continue
		}
		payChannel := o.(entity.PayChannel)
		if payChannel.Wstatus == 0 || payChannel.WithdrawWeight == 0 {
			continue
		}
		// 提现金额限制
		if amount < payChannel.WithdrawMin || amount > payChannel.WithdrawMax {
			continue
		}
		// used
		if utils.SliceIn(utils.String(c.ChannelId), exclude...) {
			continue
		}
		// 支持的银行 巴基斯坦才有
		has := true
		if utils.SliceIn(bank, payChannel.WithdrawBanks...) {
			has = true
		}
		if utils.SliceIn(bank, payChannel.WithdrawWallets...) {
			has = true
		}
		if !has && country == "PK" {
			continue
		}

		choices = append(choices, utils.Choice{Weight: int(payChannel.WithdrawWeight), Item: int(c.ChannelId)})
	}

	// 选择一个渠道
	if len(choices) != 0 {
		choice, err := utils.WeightedChoice(choices)
		if err == nil {
			return uint32(choice.Item.(int))
		}
	}

	// 随机一个渠道
	choices = make([]utils.Choice, 0)
	PayMap.Range(func(key, value interface{}) bool {
		payChannel := value.(entity.PayChannel)
		if payChannel.Wstatus == 0 || payChannel.WithdrawWeight == 0 {
			return true
		}

		// 提现金额限制
		if amount < payChannel.WithdrawMin || amount > payChannel.WithdrawMax {
			return true
		}
		// used
		if utils.SliceIn(payChannel.Id, exclude...) {
			return true
		}

		// 支持的银行 巴基斯坦才有
		has := false
		if utils.SliceIn(bank, payChannel.WithdrawBanks...) {
			has = true
		}
		if utils.SliceIn(bank, payChannel.WithdrawWallets...) {
			has = true
		}
		if !has && country == "PK" {
			return true
		}

		id, err := strconv.Atoi(payChannel.Id)
		if err != nil {
			glog.Errorf("withdraw channelId to int error: %s, %v", payChannel.Id, err)
			return true
		}
		choices = append(choices, utils.Choice{Weight: int(payChannel.WithdrawWeight), Item: id})
		return true
	})

	if len(choices) > 0 {
		choice, err := utils.WeightedChoice(choices)
		if err == nil {
			return uint32(choice.Item.(int))
		}
	}
	return 0
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

// 生成私钥对象
func ParsePrivateKey(key string) (*rsa.PrivateKey, error) {
	rsaBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, errors.New("key is error")
	}
	block := pem.Block{
		Bytes: rsaBytes,
		Type:  "RSA PRIVATE KEY",
	}
	// 解析DER编码的私钥，生成私钥对象
	prikey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return prikey.(*rsa.PrivateKey), nil
}

// 生成公钥对象
func ParsePublickKey(key string) (*rsa.PublicKey, error) {
	rsaBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, errors.New("key is error")
	}
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: rsaBytes,
	}
	// 解析DER编码的公钥，生成公钥对象
	pubkey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pubkey.(*rsa.PublicKey), nil
}

// func GetPayMent(id string, rtype pb.RechargeType) config.IPay {
// 	switch rtype {
// 	case pb.SHOP:
// 		shop := config.GetShop(id)
// 		return &shop
// 	case pb.ACTIVITY:
// 		act := config.GetActivity(id)
// 		return &act
// 	case pb.WEEKCARD:
// 		wid, _ := strconv.Atoi(id)
// 		act := config.GetWeeklyCardById(int32(wid))
// 		return &act
// 	case pb.GAMERECHARGE:
// 		return &data.DeskRecharge{}
// 	}
// 	return nil
// }

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

/*
	DB写入
*/
// 创建支付订单
func AddPayOrder(order *entity.PayOrder) error {
	if Insert(PayOrders, order) {
		return nil
	}
	return errors.New("支付订单写入失败:" + order.OrderID)
}

// 创建提现订单
func AddWithDrawOrder(order *entity.WithdrawOrder) error {
	if Insert(WithdrawOrders, order) {
		return nil
	}
	return errors.New("提现订单写入失败:" + order.OrderID)
}

// 更新支付订单
func UpdateOrder(info *entity.PayOrder) error {
	m := bson.M{"_id": info.OrderID}
	n := bson.M{"$set": bson.M{
		"out_trade_no":     info.OutTradeNo,
		"pay_time":         info.PayTime,
		"order_status":     info.OrderStatus,
		"out_trade_status": info.OutTradeStatus,
		"request_param":    info.RequestParam,
		"request_msg":      info.RequestMsg,
		"call_back_msg":    info.CallBackMsg,
		"channel_id":       info.ChannelId,
	}}
	if Update(PayOrders, m, n) {
		return nil
	}
	return errors.New("更新失败")
}

// 更新提现订单
func UpdateWithDrawOrder(info *entity.WithdrawOrder) error {
	m := bson.M{"_id": info.OrderID}
	n := bson.M{"$set": bson.M{
		"out_trade_no":     info.OutTradeNo,
		"pay_time":         info.PayTime,
		"order_status":     info.OrderStatus,
		"out_trade_status": info.OutTradeStatus,
		"request_param":    info.RequestParam,
		"request_msg":      info.RequestMsg,
		"call_back_msg":    info.CallBackMsg,
		"refuse_reason":    info.RefuseReason,
		"channel_id":       info.ChannelId,
		"inspect_times":    info.InspectTimes,
		"used_channel_id":  info.UsedChannelId,
	}}
	//n := bson.M{"$set": bson.M{"out_trade_no": info.OutTradeNo, "pay_time": info.PayTime, "order_status": info.OrderStatus, "out_trade_status": info.OutTradeStatus}}
	if Update(WithdrawOrders, m, n) {
		return nil
	}
	return errors.New("更新失败")
}

// 根据订单ID查询代收订单信息
func GetPayOrder(id string) (*entity.PayOrder, error) {
	orderInfo := new(entity.PayOrder)
	Get(PayOrders, id, orderInfo)
	if orderInfo == nil || orderInfo.OrderID == "" {
		return orderInfo, errors.New("代收订单不存在")
	}
	return orderInfo, nil
}

// 根据订单ID查询代付订单信息
func GetWithDrawOrder(id string) (*entity.WithdrawOrder, error) {
	orderInfo := new(entity.WithdrawOrder)
	Get(WithdrawOrders, id, orderInfo)
	if orderInfo == nil || orderInfo.OrderID == "" {
		orderInfo = new(entity.WithdrawOrder)
		return orderInfo, errors.New("代付订单不存在")
	}
	return orderInfo, nil
}

func GetWithDrawRecord(merOrderId string) (*data.WithdrawRecord, error) {
	record := &data.WithdrawRecord{}
	Get(WithdrawRecords, merOrderId, record)
	if record.Userid == "" {
		return record, errors.New("代付订单不存在")
	}
	return record, nil
}

// 根据用户id获取用户信息
func GetPlayerUser(userid string) (*entity.PlayerUser, error) {
	player := new(entity.PlayerUser)
	Get(PlayerUsers, userid, player)
	if player == nil || player.Userid == "" {
		return player, errors.New("player不存在")
	}
	return player, nil
}

// 支付埋点记录
func AddPayPointRecord(rec *entity.PayPointRecord) error {
	rec.ID = bson.NewObjectId().Hex()
	timestamp := utils.BsonNow().UnixNano() / 1e6
	rec.Ctime = timestamp
	if Insert(PayPointRecords, rec) {
		return nil
	}
	return errors.New("支付埋点记录写入失败:" + rec.OrderID)
}

// GetByQ(DataSummarys, bson.M{"date": summary.Date}, info)
func GetOrderById(orderid string, rtype int) string {
	businessid := ""
	switch rtype {
	case 1:
		payInfo := new(entity.PayOrder)
		GetByQ(PayOrders, bson.M{"_id": orderid}, payInfo)
		if payInfo != nil && payInfo.OrderID != "" {
			businessid = payInfo.BusinessID
		}
	case 2:
		withInfo := new(entity.WithdrawOrder)
		GetByQ(WithdrawOrders, bson.M{"_id": orderid}, withInfo)
		if withInfo != nil && withInfo.OrderID != "" {
			businessid = withInfo.BusinessID
		}
	}
	return businessid
}

func GetWithdrawByMerOrder(merId string) (*entity.WithdrawOrder, error) {
	order := new(entity.WithdrawOrder)
	GetByQ(WithdrawOrders, bson.M{"mer_order_id": merId}, order)
	if order.OrderID == "" {
		return nil, errors.New("no found order")
	}
	return order, nil
}
