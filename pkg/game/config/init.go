package config

// 配置从数据库初始化
func ConfigInit() {
	InitNotice()           //公告服务
	InitShop()             //商城服务
	InitEnv()              //变量服务
	InitGame()             //游戏服务
	InitTask()             //任务服务
	InitVip()              //vip
	InitLogin()            //登录奖励服务
	InitLucky()            //lucky服务
	InitActivity()         //Activity服务
	InitDailySign()        //签到
	InitOnlineReward()     //在线奖励
	InitLimitedGift()      //限时礼包
	InitWithDraw()         //提现配置
	InitSetting()          //功能开关
	InitPayChannel()       //支付渠道
	InitShare()            //分享活动
	InitBeginner()         //新手配置
	InitEmoji()            //人机表情
	InitPvpRoom()          // 对战房基础配置
	InitRankWithdraw()     // 提现排行榜
	InitVipRobots()        // 人机vip配置
	InitLuckyDraw()        // 小米手机活动
	InitMarqueeWithdraws() // 新跑马灯配置
	InitCrash()            // crash策略
	InitLHD()              // 龙虎斗策略
	InitSeven()            // 7up策略
	InitAB()               // andarbahar策略
	InitCP()               // 彩票策略
	InitRB()               // redblack策略
	InitAviator()          // 飞机策略
	InitCoupon()           // 优惠券
}

// 节点变量初始化, 节点连接时同步数据
func Init2Gate(id, secret, key, machid, pattern, notifyUrl string) {
	InitNotice2()       //公告服务
	InitShop2()         //商城服务
	InitEnv2()          //变量服务
	InitGame2()         //游戏服务
	InitTask2()         //任务服务
	InitVip2()          //vip
	InitLogin2()        //登录奖励服务
	InitLucky2()        //lucky服务
	InitActivity2()     //Activity服务
	InitDailySign2()    //签到
	InitOnlineReward2() //在线奖励
	InitLimitedGift2()  //限时礼包
	InitWithDraw2()     //提现配置
	InitSetting2()      //功能开关
	InitPayChannel2()   //支付渠道
	InitShare2()        //分享活动
	InitBeginner2()     //新手配置
	InitEmoji2()        //人机表情

	// WxLoginInit(id, secret) //微信登录

	// WxPayInit(id, key, machid, pattern, notifyUrl) //微信支付
}

// 逻辑服初始化
func Init2Game() {
	InitNotice2()       //公告服务
	InitShop2()         //商城服务
	InitEnv2()          //变量服务
	InitGame2()         //游戏服务
	InitTask2()         //任务服务
	InitLogin2()        //登录奖励服务
	InitLucky2()        //lucky服务
	InitActivity2()     //Activity服务
	InitDailySign2()    //签到
	InitOnlineReward2() //在线奖励
	InitLimitedGift2()  //限时礼包
	InitWithDraw2()     //提现配置
	InitSetting2()      //功能开关
	InitPayChannel2()   //支付渠道
	InitShare2()        //分享活动
	InitBeginner2()     //新手配置
	InitEmoji2()        //人机表情
	InitVip2()          //vip
}
