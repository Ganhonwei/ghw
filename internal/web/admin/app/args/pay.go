package args

type PayListArgs struct {
	Userid         string   `json:"userid"`      // UID
	PayChannels    []int    `json:"payChannels"` // 支付通道
	OrderId        string   `json:"orderId"`     // 订单号
	OutTradeNo     string   `json:"outTradeNo"`  // 第三方订单号
	CtimeS         string   `json:"ctimeS"`      // 订单生成时间
	CtimeE         string   `json:"ctimeE"`
	PayTimeS       string   `json:"payTimeS"` // 订单完成时间
	PayTimeE       string   `json:"payTimeE"`
	OrderStatus    int      `json:"orderStatus"`    // 订单状态
	IsUTR          int      `json:"isUTR"`          // 是否有UTR 0全部 1是 2否
	IsTag          int      `json:"isTag"`          // 是否有标记 0全部 1是 2否
	IsFirstPay     int      `json:"isFirstPay"`     // 是否首充 0全部 1是 2否
	ChannelIds     []string `json:"channelIds"`     // 渠道
	ChannelClasses []string `json:"channelClasses"` // 渠道类
}

type ChannelEditArgs struct {
	Id              string   `json:"id"`
	Name            string   `json:"name"`
	SortId          int      `json:"sortId"`
	Status          int      `json:"status"`
	Wstatus         int      `json:"wstatus"`
	Ptype           int      `json:"ptype"`
	PayRate         float64  `json:"payRate"`
	WithdrawRate    float64  `json:"withdrawRate"`
	WithdrawFee     int64    `json:"withdrawFee"`
	PayOptions      []int32  `json:"payOptions"`
	PayApps         []int32  `json:"payApps"`
	WithdrawBanks   []string `json:"withdrawBanks"`
	WithdrawWallets []string `json:"withdrawWallets"`
	UtrRequired     bool     `json:"utrRequired"`
	PayWeight       int32    `json:"payWeight"`
	PayMin          int64    `json:"payMin"`
	PayMax          int64    `json:"payMax"`
	WithdrawWeight  int32    `json:"withdrawWeight"`
	WithdrawMin     int64    `json:"withdrawMin"`
	WithdrawMax     int64    `json:"withdrawMax"`
	Balance         int64    `json:"balance"`
}

type ChannelTestArgs struct {
	Id     string `json:"id"`     // 通道id
	Amount int32  `json:"amount"` // 支付金额
}

type LiveDataTrendArg struct {
	Stime  string  `json:"stime"`
	Etime  string  `json:"etime"`
	Trends []int32 `json:"trends"`
}
