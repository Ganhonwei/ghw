package service

import (
	"sync"

	"github.com/nsqio/go-nsq"
)

const (
	XDPAY        = 3010
	MLPLAY       = 3011
	PAY1916      = 3012
	LETSPAY      = 3013
	KINGPAY      = 3014
	XFPAY        = 3015
	SAILSPAY     = 3016
	FLYPAY       = 3017
	UWINPAY      = 3018
	TANSAFEPAY   = 3019
	RAMAPAY      = 3020
	ICEPAY       = 3021
	WEPAY        = 3022
	OEPAY        = 3023
	PAY9S        = 3024
	BLIZZARDPY   = 3025
	METAGOPAY    = 3026
	UNIVERSALPAY = 3027
	KKPLUSPAY    = 3028
	DRAGONPAY    = 3029
	AIPAY        = 3030
	GAMEPAY      = 3031
	GENTLEPAY    = 3032
	PAY66        = 3033
	AI2PAY       = 3034
	COWPAY       = 3035
	LETSPAYFAST  = 3036
	PAYWOOK      = 3037
	MEPAY        = 3038
	DDPAY        = 3039
	ATPAY        = 3040
	NETPAY       = 3041
	WYPPAY       = 3042
	YUNPAY       = 3043
	CXPAY        = 3044
	CSMPAY       = 3045
	WEPAY2       = 3046
	LEOPAY       = 3047
	UPAY         = 3048
	EOCPAY       = 3049

	// 孟加拉
	MJL_UNI = 5001
	// ustd pay
	USDT           = 5002
	MJL_JY         = 5003
	MJL_Transafe   = 5004
	MJL_MM         = 5005
	MJL_SAFEPAY    = 5006
	MJL_99         = 5007
	MJL_GOPAY      = 5008
	MJL_SHPAY      = 5009
	MJLBLIZZARDPAY = 5010
	MJL_TTTPAY     = 5011
	MJL_AI2PAY     = 5012
	MJL_66PAY      = 5013
	MJL_EOCPAY     = 5014

	//巴基斯坦
	BJST_AI2PAY = 6001
	BJST_PAKPAY = 6002
	BJST_EOCPAY = 6003
)

var ChannelNameMap = map[int]string{
	XDPAY:       "xdpay",
	MLPLAY:      "mlpay",
	PAY1916:     "1916pay",
	LETSPAY:     "letspay",
	KINGPAY:     "kingpay",
	XFPAY:       "xfpay",
	SAILSPAY:    "sailspay",
	FLYPAY:      "flypay",
	UWINPAY:     "uwinpay",
	TANSAFEPAY:  "tansafepay",
	RAMAPAY:     "ramapay",
	ICEPAY:      "icepay",
	WEPAY:       "wepay",
	OEPAY:       "oepay",
	PAY9S:       "9spay",
	BLIZZARDPY:  "blizzardpay",
	KKPLUSPAY:   "kkpluspay",
	DRAGONPAY:   "dragonpay",
	AIPAY:       "aipay",
	GAMEPAY:     "gamepay",
	GENTLEPAY:   "gentlepay",
	PAY66:       "pay66",
	AI2PAY:      "ai2pay",
	COWPAY:      "cowpay",
	LETSPAYFAST: "letspayfast",
	PAYWOOK:     "paywook",
	MEPAY:       "mepay",
	DDPAY:       "ddpay",
	ATPAY:       "atpay",
	NETPAY:      "netpay",
	WYPPAY:      "wyppay",
	YUNPAY:      "yunpay",
}
var (
	PayHistory []*PayChannelHistory
	Mutex      sync.RWMutex
)

var (
	WithdrawHistory []*PayChannelHistory
	WithdrawMutex   sync.RWMutex
)

var (
	Producer *nsq.Producer
	Env      string
)

type PayChannelHistory struct {
	ChannelId   uint32
	Info        []*PayInfo
	SuccessRate int // 成功率
}

type PayInfo struct {
	OrderID string
	Success bool
}
