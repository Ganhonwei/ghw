package entity

import "time"

// 游戏房间
type Game struct {
	Id              string    `bson:"_id" json:"id"`                    //房间ID
	Name            string    `bson:"name" json:"name"`                 //房间名称
	Gtype           int       `bson:"gtype" json:"gtype"`               //游戏类型 1：TP; 2：TPAK47; 3：TPJOKER;
	Status          int       `bson:"status" json:"status"`             //房间开关 0: 关; 1: 开;
	Ai_Status       int       `bson:"ai_status" json:"ai_status"`       //房间AI开关 0: 关; 1: 开;
	Min_Access      int       `json:"min_access" bson:"min_access"`     //最低准入
	Max_Access      int       `json:"max_access" bson:"max_access"`     //最高准入
	Stock_Expect    int       `json:"stock_expect" bson:"stock_expect"` //库存期望
	Stock_Alarm     int       `json:"stock_alarm" bson:"stock_alarm"`   //库存报警
	FMin_Access     float64   `bson:"-"`                                //最低准入
	FMax_Access     float64   `bson:"-"`                                //最高准入
	FStock_Expect   float64   `bson:"-"`                                //库存期望
	FStock_Alarm    float64   `bson:"-"`                                //库存报警
	SortId          int       `json:"sortId" bson:"sort_id"`            //房间分组内排序标识
	Count           uint32    `bson:"count" json:"count"`               //房间人数上限
	Bottom          int       `json:"bottom" bson:"bottom"`             //底注
	Rounds          int       `json:"rounds" bson:"rounds"`             //轮次上限
	Than_Rounds     int       `json:"than_rounds" bson:"than_rounds"`   //比牌轮次
	Pool_Limit      int       `json:"pool_limit" bson:"pool_limit"`     //筹码池上限
	Otime           int       `json:"otime" bson:"otime"`               //操作时间
	Tcountdown      int       `json:"tcountdown" bson:"tcountdown"`     //比牌倒计时
	Scountdown      int       `json:"scountdown" bson:"scountdown"`     //结算倒数时间
	Match_Time      []int     `json:"match_time" bson:"match_time"`     //匹配时间区间
	Match_TimeStr   string    `bson:"-"`
	Single_Robot    []int     `json:"single_robot" bson:"single_robot"` //单局人机数区间
	Single_RobotStr string    `bson:"-"`
	Robot_Join      []int     `json:"robot_join" bson:"robot_join"` //人机加入概率
	Robot_JoinStr   string    `bson:"-"`
	Robot_Leave     []int     `json:"robot_leave" bson:"robot_leave"` //人机离开概率(万分比)
	Robot_LeaveStr  string    `bson:"-"`
	Prevent_Time    int       `json:"prevent_time" bson:"prevent_time"` //防作弊匹配时间检测
	Prevent_Num     int       `json:"prevent_num" bson:"prevent_num"`   //防作弊匹配次数检测
	Prevent_Thaw    int       `json:"prevent_thaw" bson:"prevent_thaw"` //防作弊匹配解冻时间
	Msg_Score       int       `json:"msg_score" bson:"msg_score"`       //跑马灯显示分
	Ctime           time.Time `bson:"ctime"`                            //修改时间
}

// 玩法类型
var GameTypes = map[int]string{
	1: "TP",
	2: "TPAK47",
	3: "TPJOKER",
}

// 房间状态
var RoomStatus = map[int]string{
	0: "关",
	1: "开",
}

// 游戏系统设置
type GameSystem struct {
	Id        string    `bson:"_id" json:"id"`               //ID
	Name      string    `bson:"name" json:"name"`            //游戏名称
	Gtype     int       `json:"gtype" bson:"gtype"`          // 游戏ID
	Rtype     int       `json:"rtype" bson:"rtype"`          //类型 1:邮箱自动回复；2:IP限制；3:支付黑名单；4：服务器维护；5：邮箱提醒
	Stype     int       `bson:"stype" json:"stype"`          //系统设置类型 0:游戏图标；1:活动图标；2:新手设置；
	Status    int       `bson:"status" json:"status"`        //开关 0: 关; 1: 开;
	PayStatus int       `json:"payStatus" bson:"pay_status"` // 是否付费开关 0：全部用户;1：付费用户;2：未付费用户
	IsButton  bool      `bson:"-"`                           // 是否是按钮
	SortId    int       `bson:"sort_id" json:"sort_id"`      //权重
	Ctime     time.Time `bson:"ctime"`                       //修改时间
	Value     string    `bson:"value"`                       // 值
	Tab       []int     `json:"tab" bson:"tab"`              // 页签
	Tab1      int       `bson:"-"`
	Tab2      int       `bson:"-"`
	Tab3      int       `bson:"-"`
}

// 支付渠道
type PayChannel struct {
	Id              string    `bson:"_id" json:"id"`                           // 渠道id
	Name            string    `bson:"name" json:"name"`                        // 渠道名称
	Ptype           int       `bson:"ptype" json:"ptype"`                      // 通道类型: 0.未知 1.纯原生 2.纯唤醒 3.原唤混
	Status          int       `bson:"status" json:"status"`                    // 充值开关 0: 关; 1: 开;
	Wstatus         int       `bson:"wstatus" json:"wstatus"`                  // 提现开关 0: 关; 1: 开;
	PayRate         float64   `bson:"pay_rate" json:"payRate"`                 // 代收税率
	WithdrawRate    float64   `bson:"withdraw_rate" json:"withdrawRate"`       // 代付税率
	WithdrawFee     int64     `bson:"withdraw_fee" json:"withdrawFee"`         // 代付固定手续费(代付单笔额外费用)
	Balance         int       `bson:"balance" json:"balance"`                  // 账户余额
	SortId          int       `bson:"sort_id" json:"sort_id"`                  // SortId
	Ctime           time.Time `bson:"ctime"`                                   // 修改时间
	PayOptions      []int32   `bson:"pay_options" json:"payOptions"`           // 支持的支付方式: 1.DirectLaunchTheApp 2.QRcode
	PayApps         []int32   `bson:"pay_apps" json:"payApps"`                 // 支持的支付app: 1.Paytm 2.PhonePe 3.Mobikwik 4.BHIM 5.GooglePay 6.Other UPI
	WithdrawBanks   []string  `bson:"withdraw_banks" json:"withdrawBanks"`     // 支持的提现银行
	WithdrawWallets []string  `bson:"withdraw_wallets" json:"withdrawWallets"` // 支持的提现钱包
	UtrRequired     bool      `bson:"utr_required" json:"utrRequired"`         // 是否需要UTR FillInUTR
	PayWeight       int32     `bson:"pay_weight" json:"payWeight"`             // 代收权重
	PayMin          int64     `bson:"pay_min" json:"payMin"`                   // 充值金额最小
	PayMax          int64     `bson:"pay_max" json:"payMax"`                   // 充值金额最大
	WithdrawWeight  int32     `bson:"withdraw_weight" json:"withdrawWeight"`   // 代付权重
	WithdrawMin     int64     `bson:"withdraw_min" json:"withdrawMin"`         // 提现金额最小
	WithdrawMax     int64     `bson:"withdraw_max" json:"withdrawMax"`         // 提现金额最大

	FPtype              string  `bson:"-" json:"fPtype"`              // 通道类型: 0.未知 1.纯原生 2.纯唤醒 3.原唤混
	FPayRate            string  `bson:"-" json:"fPayRate"`            // 代收税率
	FWithdrawRate       string  `bson:"-" json:"fWithdrawRate"`       // 代付税率
	FWithdrawFee        string  `bson:"-" json:"fWithdrawFee"`        // 代付固定手续费(代付单笔额外费用)
	FPayOptions         string  `bson:"-" json:"fPayOptions"`         // 支持的支付方式
	FPayApps            string  `bson:"-" json:"fPayApps"`            // 支持的支付app
	FWithdrawBanks      string  `bson:"-" json:"fWithdrawBanks"`      // 支持的提现银行
	FWithdrawWallets    string  `bson:"-" json:"fWithdrawWallets"`    // 支持的提现钱包
	FUtrRequired        string  `bson:"-" json:"fUtrRequired"`        // 是否需要UTR
	FPayRange           string  `bson:"-" json:"fPayRange"`           // 充值金额区间
	FWithdrawRange      string  `bson:"-" json:"fWithdrawRange"`      // 充值金额区间
	FBalance            string  `bson:"-" json:"fBalance"`            // 通道余额（刷新）
	FPaySuccessHourRate string  `bson:"-" json:"fPaySuccessHourRate"` // 近1小时代收成功率（刷新）
	PaySuccessHourRate  float64 `bson:"-" json:"paySuccessHourRate"`  // 近1小时代收成功率（刷新）
	Version             string  `bson:"-" json:"version"`             // 版本

	FPayOptionIds string `bson:"-" json:"fPayOptionIds"` // 支持的支付方式 ID
	FPayAppIds    string `bson:"-" json:"fPayAppIds"`    // 支持的支付app ID
	FPayMin       string `bson:"-" json:"fPayMin"`       // 充值金额最小
	FPayMax       string `bson:"-" json:"fPayMax"`       // 充值金额最大
	FWithdrawMin  string `bson:"-" json:"fWithdrawMin"`  // 提现金额最小
	FWithdrawMax  string `bson:"-" json:"fWithdrawMax"`  // 提现金额最大
}

// 库存曲线
type StockHistory struct {
	Timestamp    int64     `bson:"timestamp" json:"timestamp"`           //时间戳
	SDate        time.Time `bson:"-"`                                    // 统计日期
	CashStock    int64     `bson:"cash_stock" json:"cash_stock"`         //彩金库存
	FCashStock   float64   `bson:"-"`                                    //彩金库存
	BonusStock   int64     `bson:"bonus_stock" json:"bonus_stock"`       //奖励金库存
	FBonusStock  float64   `bson:"-"`                                    //奖励金库存
	CashMingTax  int64     `bson:"cash_ming_tax" json:"cash_ming_tax"`   //彩金明税
	BonusMingTax int64     `bson:"bonus_ming_tax" json:"bonus_ming_tax"` //奖励金明税
	CashAnTax    int64     `bson:"cash_an_tax" json:"cash_an_tax"`       //彩金暗税
	BonusAnTax   int64     `bson:"bonus_an_tax" json:"bonus_an_tax"`     //奖励金暗税
}

// 库存
type Stock struct {
	Id            string         `bson:"_id" json:"id"`                        //房间ID
	Type          int32          `bson:"type" json:"type"`                     //游戏类型 1：TP类; 2：Rummy; 3：百人;
	CashStock     int64          `bson:"cash_stock" json:"cash_stock"`         //彩金库存
	BonusStock    int64          `bson:"bonus_stock" json:"bonus_stock"`       //奖励金库存
	CashMingTax   int64          `bson:"cash_ming_tax" json:"cash_ming_tax"`   //彩金明税
	BonusMingTax  int64          `bson:"bonus_ming_tax" json:"bonus_ming_tax"` //奖励金明税
	CashAnTax     int64          `bson:"cash_an_tax" json:"cash_an_tax"`       //彩金暗税
	BonusAnTax    int64          `bson:"bonus_an_tax" json:"bonus_an_tax"`     //奖励金暗税
	Factor        float64        `bson:"factor" json:"factor"`                 //当前房间系数
	FCashStock    float64        `bson:"-"`                                    //彩金库存
	FBonusStock   float64        `bson:"-"`                                    //奖励金库存
	FCashMingTax  float64        `bson:"-"`                                    //彩金明税
	FBonusMingTax float64        `bson:"-"`                                    //奖励金明税
	FCashAnTax    float64        `bson:"-"`                                    //彩金暗税
	FBonusAnTax   float64        `bson:"-"`                                    //奖励金暗税
	GameInfo      Game           `bson:"-"`                                    // 用户信息
	IsCashStock   bool           `bson:"-"`                                    // 彩金库存是否负值
	History       []StockHistory `bson:"history" json:"history"`               //库存曲线
}

// gamestock
type GameStock struct {
	Id            string  `bson:"_id" json:"id"`                        // 游戏ID
	CashStock     int64   `bson:"cash_stock" json:"cash_stock"`         //彩金库存
	BonusStock    int64   `bson:"bonus_stock" json:"bonus_stock"`       //奖励金库存
	CashMingTax   int64   `bson:"cash_ming_tax" json:"cash_ming_tax"`   //彩金明税
	BonusMingTax  int64   `bson:"bonus_ming_tax" json:"bonus_ming_tax"` //奖励金明税
	CashAnTax     int64   `bson:"cash_an_tax" json:"cash_an_tax"`       //彩金暗税
	BonusAnTax    int64   `bson:"bonus_an_tax" json:"bonus_an_tax"`     //奖励金暗税
	Factor        float64 `bson:"factor" json:"factor"`                 //当前房间系数
	FCashStock    float64 `bson:"-"`                                    //彩金库存
	FBonusStock   float64 `bson:"-"`                                    //奖励金库存
	FCashMingTax  float64 `bson:"-"`                                    //彩金明税
	FBonusMingTax float64 `bson:"-"`                                    //奖励金明税
	FCashAnTax    float64 `bson:"-"`                                    //彩金暗税
	FBonusAnTax   float64 `bson:"-"`                                    //奖励金暗税
	IsCashStock   bool    `bson:"-"`                                    // 彩金库存是否负值
}

// 龙虎斗库存修改
type ModifyGameStock struct {
	Gtype int32 `json:"gtype" bson:"gtype"` // 游戏类型
	Stock int64 `json:"stock" bson:"stock"` // 游戏库存
}

type WebResponse struct {
	Code   int    `json:"code"` // 错误码    200:success
	ErrMsg string `json:"msg"`  // 错误信息
	Body   []byte `json:"body"` // 响应体
}

type WebQueryBalence struct {
	ChannelId uint32 `json:"channel_id"` // 渠道id
	Balence   int64  `json:"balence"`    // 可用余额
}

/* 之前的 */

// 所属节点
var GameNodesName = map[string]string{
	"game.huiyin1": "节点1区赛车彩种",
	"game.huiyin2": "节点2区飞艇彩种",
}

// 所属节点
var GameNodes = map[int]string{
	1: "节点1区赛车彩种",
	2: "节点2区飞艇彩种",
}
var Game2Nodes = map[int]string{
	1: "game.huiyin1",
	2: "game.huiyin2",
}

// 彩种类型
var LotteryTypes = map[int]string{
	1: "赛车彩种",
	2: "飞艇彩种",
}

// 房间类型
var RoomTypes = map[int]string{
	0: "免佣",
	1: "抽佣",
}

// 是否上庄
var IsDeal = map[int]string{
	0: "否",
	1: "是",
}
var Is2Deal = map[int]bool{
	0: false,
	1: true,
}

// 彩种类型
var LotteryCodes = map[int]string{
	1: "bjpk10",
	2: "mlaft",
}

// 是否机器人
var IsRobot = map[bool]string{
	false: "玩家",
	true:  "机器人",
}
