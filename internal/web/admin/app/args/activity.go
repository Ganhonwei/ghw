package args

type CoupnStatsArgs struct {
	CtimeS         string   `json:"ctimeS"` // 日期
	CtimeE         string   `json:"ctimeE"`
	AliasIds       []string `json:"aliasIds"`       // 渠道别名
	ChannelClasses []string `json:"channelClasses"` // 渠道类
	UserTypes      []int32  `json:"userTypes"`
	RegistArea     int32    `json:"registArea"` // 是否首充 0全部 1是 2否
}

type CoupnHandArgs struct {
	CtimeS     string `json:"ctimeS"` // 日期
	CtimeE     string `json:"ctimeE"`
	StimeS     string `json:"stimeS"` // 日期
	StimeE     string `json:"stimeE"`
	UtimeS     string `json:"utimeS"` // 日期
	UtimeE     string `json:"utimeE"`
	UserTypes  []int  `json:"userTypes"`
	RegistArea []int  `json:"registArea"`
}
