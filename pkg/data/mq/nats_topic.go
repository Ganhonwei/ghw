package mq

// 初始化jetStream流, {streamName, topics...}
var initialStream = [][]string{
	{StreamGame, "game.>"},
	{StreamActivity, "activity.>"},
	{StreamSync, "sync.>"},
	{StreamReport, "report.>"},
	{StreamAlert, "alert.>"},
	{StreamUtr, "utr.>"},
	{StreamOnlineUser, "online.>"},
	{StreamWithdraw, "withdraw.>"},
}

const (
	// ========================== jetstraem定义 ==========================
	StreamGame       = "stream-game"       // 游戏行为jetstraem消息流
	StreamActivity   = "stream-activity"   // 活动jetstraem消息流
	StreamSync       = "stream-sync"       // 同步jetstraem消息流
	StreamReport     = "stream-report"     // 上报jetstraem消息流
	StreamAlert      = "stream-alert"      // 告警jetstraem消息流
	StreamUtr        = "stream-utr"        // utrjetstraem消息流
	StreamOnlineUser = "stream-onlineuser" // 用户在线jetstraem消息流
	StreamWithdraw   = "stream-withdraw"   // 提现jetstraem消息流

	// ========================== mq主题订阅 ==========================
	// ------------- 游戏行为 -------------
	TopicGameLogin               = "game.login"                 // 玩家登录上报
	TopicGameBets                = "game.bets"                  // 玩家打码上报
	TopicGameToLobby             = "game.to_lobby"              // 玩家回到大厅上报
	TopicShareSuperior           = "game.share_superior"        // 分享用户注册
	TopicExternalBet             = "game.external.bet"          // 外接游戏打码
	TopicExternalReword          = "game.external.reword"       // 外接游戏返奖
	TopicExternalCancel          = "game.external.cancel"       // 外接游戏撤销
	TopicPaychannelUpdate        = "game.paychannel.update"     // 支付渠道更新通知
	TopicPayOrderError           = "game.pay.order.error"       // 支付渠道下单失败通知
	TopicGameCustomerReply       = "game.customer.reply"        // 客服回复消息通知
	TopicGameCustomerRevoke      = "game.customer.revoke"       // 客服撤回消息通知
	TopicGameCustomerRead        = "game.customer.read"         // 客服消息玩家已读通知
	TopicGameCustomerScore       = "game.customer.score"        // 客服会话玩家评分
	TopicGameCustomerSessionOver = "game.customer.session_over" // 客服会话结束

	// ------------- 打码量排名活动 -------------
	TopicActivityBetRankUpdate = "activity.betrank.update" // 排行榜更新通知
	TopicActivityBetRankWindow = "activity.betrank.window" // 排行榜弹窗

	// ------------- 同步 -------------
	TopicSyncUser                              = "sync.user"                                  // 用户同步
	TopicSyncUserFinance                       = "sync.user.finance"                          // 用户财务同步
	TopicSyncTradeRecord                       = "sync.trade.record"                          // 交易记录同步
	TopicSyncWithdrawRecord                    = "sync.withdraw.record"                       // 提现记录同步
	TopicSyncWithdrawRecordTransferOrderDetail = "sync.withdraw.record.transfer.order.detail" // 提现记录转账订单详情同步
	TopicSyncLogWater                          = "sync.log.water"                             // 日志水流同步
	TopicSyncDetail                            = "sync.detail"                                // 详情同步
	TopicSyncLogLogin                          = "sync.log.login"                             // 登录日志同步
	TopicSyncExternalBet                       = "sync.external.bet"                          // 外接打码上报
	TopicSyncExternalReward                    = "sync.external.reward"                       // 外接打码奖励上报
	TopicSyncExternalCancel                    = "sync.external.cancel"                       // 外接打码取消上报
	TopicSyncActivityBetRankPrize              = "sync.activity.betrank.prize"                // 活动打码量排行榜奖励同步
	TopicSyncActivityTurnDrawLog               = "sync.activity.turn.draw.log"                // 活动转盘抽奖日志同步
	TopicSyncActivityTurnPrizeLog              = "sync.activity.turn.prize.log"               // 活动转盘抽奖结果同步
	TopicSyncLaunchInviteLog                   = "sync.launch.invite.log"                     // 启动邀请日志同步
	TopicSyncShareAgentIncomeRecord            = "sync.share.agent.income.record"             // 分享代理收益记录同步
	TopicSyncShareAgentIncomeRecordTackLog     = "sync.share.agent.income.record.tack.log"    // 分享代理收益记录领取日志同步
	TopicSyncLogVBDiamond                      = "sync.log.vb_diamond"                        // vbbank变化日志同步
	TopicSyncLogOutDiamond                     = "sync.log.out_diamond"                       // 可提现金变化日志同步
	TopicSyncVolatilitySubsidy                 = "sync.volatility_subsidy"                    // 波动返水领取记录同步
	TopicSyncOnlineUsersLog                    = "sync.log.online_users"                      // 在线人数日志同步
	TopicSyncUserCoupon                        = "sync.col_user_coupon"                       // 在线人数日志同步
	TopicSyncCustomerChatSession               = "sync.customer_chat_session"                 // 客服聊天会话信息
	TopicSyncCustomerChatMessage               = "sync.customer_chat_message"                 // 客服聊天记录
	TopicSyncCustomerSeatLog                   = "sync.customer_seat_log"                     // 客服坐席操作记录
	TopicSyncCumRechargeWheelUnlockLog         = "sync.cum_recharge_wheel_unlock_log"         // 累充转盘解锁记录同步
	TopicSyncCumRechargeWheelSpinLog           = "sync.cum_recharge_wheel_spin_log"           // 累充转盘抽奖记录同步

	// ------------- 上报 -------------
	TopicReportRegister   = "report.register"    // 注册上报
	TopicReportLogin      = "report.login"       // 登录上报
	TopicReportDeposit    = "report.deposit"     // 充值上报
	TopicReportAdParam    = "report.ad.param"    // ad参数上报
	TopicReportFBRegister = "report.fb.register" // fb注册上报
	TopicReportFBDeposit  = "report.fb.deposit"  // fb充值上报
	TopicReportLog        = "report.log"         // 日志上报

	// ------------- 告警 -------------
	TopicAlertEmail = "alert.email" // 邮件告警

	// ------------- utr -------------
	TopicUtrUpload       = "utr.upload"        // utr上传
	TopicUtrUploadResult = "utr.upload.result" // utr上传结果
	TopicFillOrder       = "utr.fill.order"    // utr补单

	// ------------- 用户在线 -------------
	TopicOnlineUser = "online.user" // 用户在线

	// ========================== 请求响应 ==========================
	// ------------- 打码量排名活动 -------------
	RequestActivityBetRankList       = "request.activity.betrank.list"        // 排行榜列表全量请求
	RequestActivityBetRankUpdate     = "request.activity.betrank.update"      // 排行榜更新请求
	RequestActivityBetRankSkipWindow = "request.activity.betrank.skip_window" // 排行榜弹窗skip for today
	RequestActivityBetRankReword     = "request.activity.betrank.reword"      // 排行榜领奖
	RequestActivityBetRankHistory    = "request.activity.betrank.history"     // 排行榜历史请求
	RequestActivityBetRankMyRecord   = "request.activity.betrank.myrecord"    // 排行榜个人获奖历史请求
	// ------------- 用户在线 -------------
	RequestOnlineUserCount = "request.online.user.count" // 用户在线人数请求
	RequestOnlineUserList  = "request.online.user.list"  // 用户在线列表请求
	// ------------- 客服聊天 -------------
	RequestCustomerSend = "request.customer.send" // 用户发送聊天消息

	// ------------- 活动 -------------
	TopicActivityCouponUpdate = "activity.coupon.update" // 优惠券更新通知

	// ------------- 提现 -------------
	TopicWithdrawAutoTransfer = "withdraw.auto.transfer" // 提现自动转单检查
	TopicWithdrawTransfer     = "withdraw.transfer"      // 提现转单
	TopicWithdrawFail         = "withdraw.fail"          // 提现失败
)

// nats 空参数
var NatsNilData = map[string]any{}

// RequestUserArgs 携带用户信息参数请求
type RequestUserArgs struct {
	Userid string `json:"userid"`
}

type RequestActivityBetRankListArgs struct {
	Userid     string `json:"userid"`
	RegistArea int    `json:"registArea"`
}

type RequestActivityBetRankUpdateArgs struct {
	RankType int32    `json:"rankType"` // 1日榜 2周榜 3月榜
	Userids  []string `json:"userids"`
	// Users map[string]*actor.PID `json:"users"`
}

type RequestActivityBetRankSkipWindowArgs struct {
	Userid string `json:"userid"`
	WType  int32  `json:"wtype"` // 弹窗类型: 1.排行榜上榜报喜弹窗 2.快上榜提醒弹窗 3.掉榜弹窗
}

type RequestActivityBetRankRewordArgs struct {
	Id     string `json:"id"`
	Userid string `json:"userid"`
}

// 订单下单失败通知
type PublishPayOrderError struct {
	ChannelId   uint32 `json:"channelId"`   // 通道id
	ChannelName string `json:"channelName"` // 通道名
	Userid      string `json:"userid"`      // 下单用户
	MerOrderId  string `json:"merOrderId"`  // 订单号
	OutTradeNo  string `json:"outTradeNo"`  // 三方订单号
	Amount      uint32 `json:"amount"`      // 订单金额
	Err         string `json:"err"`         // 错误信息
	Response    string `json:"response"`    // 通道响应数据
}

type RequestEmptyArgs struct {
}
