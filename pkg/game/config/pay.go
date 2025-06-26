package config

import (
	"goserver/gen/pb"
	"goserver/pkg/data"

	"strconv"
	"sync"
)

type IPay interface {
	BuildPayOrder(user *data.User, order *data.PayOrder, game bool)
	CanOrder(user *data.User, id string, ctype int32) bool
	RechargeAfter(user *data.User, id string, ctype int32, save bool) (*pb.PaySuccessNtf, error)
	RechageLType() int32
}

type PayChannels struct {
	// 当前指定支付渠道
	NowPayChannel []int
	// 当前指定提现渠道
	NowWithdrawChannel []int

	channelMutex sync.Mutex
}

// 提现配置列表
var WithDrawMap *sync.Map
var PayMap *sync.Map

// 充值号码黑名单
var PayBlackMap *sync.Map

var Channels PayChannels

// 启动初始化
func InitWithDraw() {
	WithDrawMap = new(sync.Map)
	l := data.GetWithdrawList()
	for _, v := range l {
		SetWithDraw(v)
	}
}

// 启动初始化
func InitPayChannel() {
	PayMap = new(sync.Map)
	PayBlackMap = new(sync.Map)
	l := data.GetPayChannelList()
	SetPayChannel(l)
	p := data.GetPhoneBlackList()
	for _, rl := range p {
		SetPhoneBlack(rl)
	}
}

// 启动初始化
func InitWithDraw2() {
	WithDrawMap = new(sync.Map)
}

// 启动初始化
func InitPayChannel2() {
	PayMap = new(sync.Map)
	PayBlackMap = new(sync.Map)
}

// 同步数据获取
func GetWithDraw2() map[int32]data.Withdraw {
	m := make(map[int32]data.Withdraw)
	WithDrawMap.Range(func(k, v interface{}) bool {
		m[k.(int32)] = v.(data.Withdraw)
		return true
	})
	return m
}

// GetWithDraws 客户端获取消息列表
func GetWithDraws() []data.Withdraw {
	list := make([]data.Withdraw, 0)
	WithDrawMap.Range(func(k, v interface{}) bool {
		if val, ok := v.(data.Withdraw); ok {
			list = append(list, val)
		}
		return true
	})
	return list
}

func DelWithDraw(k int32) {
	WithDrawMap.Delete(k)
}

func SetWithDraw(v data.Withdraw) {
	WithDrawMap.Store(v.Id, v)
}

// 获取提现配置
func GetWithDraw(id int32) data.Withdraw {
	if v, ok := WithDrawMap.Load(id); ok {
		return v.(data.Withdraw)
	}
	return data.Withdraw{}
}

// 保存支付渠道设置
func SetPayChannel(datas []data.PayChannel) {
	defer Channels.channelMutex.Unlock()

	Channels.channelMutex.Lock()
	Channels.NowPayChannel = make([]int, 0)
	Channels.NowWithdrawChannel = make([]int, 0)
	for _, data := range datas {
		PayMap.Store(data.Id, data)

		//充值渠道
		if data.Status == 1 {
			id, err := strconv.Atoi(data.Id)
			if err != nil {
				return
			}
			Channels.NowPayChannel = append(Channels.NowPayChannel, id)
		}
		//提现渠道
		if data.Wstatus == 1 {
			id, err := strconv.Atoi(data.Id)
			if err != nil {
				return
			}
			Channels.NowWithdrawChannel = append(Channels.NowWithdrawChannel, id)
		}
	}
}

func GetPayChannelMap() map[string]data.PayChannel {
	res := make(map[string]data.PayChannel)
	PayMap.Range(func(key, value any) bool {
		res[key.(string)] = value.(data.PayChannel)
		return true
	})
	return res
}

// 保存支付黑名单
func SetPhoneBlack(d data.RechargeLimit) {
	PayBlackMap.Store(d.Phone, d)
}

// 移除支付黑名单
func DelPhoneBlack(d data.RechargeLimit) {
	PayBlackMap.Delete(d.Phone)
}

func GetPhoneBlackMap() map[string]data.RechargeLimit {
	m := make(map[string]data.RechargeLimit)
	PayBlackMap.Range(func(key, value any) bool {
		if k, ok := key.(string); ok {
			m[k] = value.(data.RechargeLimit)
		}
		return true
	})
	return m
}

// 充值配置
// func PayConfigInit(cfg *ini.File) {
// MlpayInit(cfg)
// XFPayInit(cfg)
// KingPayInit(cfg)
// SailsPayInit(cfg)
// }

// func MlpayInit(cfg *ini.File) {
// 	url := cfg.Section("mlpay").Key("payorderurl").Value()
// 	appid := cfg.Section("mlpay").Key("appid").Value()
// 	applicationId := cfg.Section("mlpay").Key("applicationId").Value()
// 	md5 := cfg.Section("mlpay").Key("md5").Value()
// 	withdrawMd5 := cfg.Section("mlpay").Key("withDrawMd5").Value()
// 	noticUrl := cfg.Section("mlpay").Key("paynotifyurl").Value()
// 	withDrawnotifyUrl := cfg.Section("mlpay").Key("withDrawnotifyurl").Value()
// 	mlpay.InitConfig(url, appid, applicationId, md5, withdrawMd5, noticUrl, withDrawnotifyUrl)
// }

// func XFPayInit(cfg *ini.File) {
// 	url := cfg.Section("xfpay").Key("payorderurl").Value()
// 	appid := cfg.Section("xfpay").Key("machid").Value()
// 	md5 := cfg.Section("xfpay").Key("md5").Value()
// 	withdrawMd5 := cfg.Section("xfpay").Key("withDrawMd5").Value()
// 	noticUrl := cfg.Section("xfpay").Key("paynotifyurl").Value()
// 	withDrawnotifyUrl := cfg.Section("xfpay").Key("withDrawnotifyurl").Value()
// 	xfpay.InitConfig(url, appid, md5, withdrawMd5, noticUrl, withDrawnotifyUrl)
// }

// func KingPayInit(cfg *ini.File) {
// 	url := cfg.Section("kingpay").Key("payorderurl").Value()
// 	appid := cfg.Section("kingpay").Key("appid").Value()
// 	applicationId := cfg.Section("kingpay").Key("applicationId").Value()
// 	serverPubRSA := cfg.Section("kingpay").Key("serverPubRSA").Value()
// 	clientPriRSA := cfg.Section("kingpay").Key("clientPriRSA").Value()
// 	noticUrl := cfg.Section("kingpay").Key("paynotifyurl").Value()
// 	withDrawnotifyUrl := cfg.Section("kingpay").Key("withDrawnotifyurl").Value()
// 	kingpay.InitConfig(url, appid, applicationId, serverPubRSA, clientPriRSA, noticUrl, withDrawnotifyUrl)
// }

// func SailsPayInit(cfg *ini.File) {
// 	url := cfg.Section("sailspay").Key("payorderurl").Value()
// 	md5 := cfg.Section("sailspay").Key("md5").Value()
// 	appid := cfg.Section("sailspay").Key("machid").Value()
// 	noticUrl := cfg.Section("sailspay").Key("paynotifyurl").Value()
// 	withDrawnotifyUrl := cfg.Section("sailspay").Key("withDrawnotifyurl").Value()
// 	sailspay.InitConfig(url, appid, md5, noticUrl, withDrawnotifyUrl)
// }
