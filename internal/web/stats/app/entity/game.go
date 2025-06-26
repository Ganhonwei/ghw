package entity

import "time"

// 支付渠道
type PayChannel struct {
	Id      string `bson:"_id" json:"id"`          // 渠道id
	Name    string `bson:"name" json:"name"`       // 渠道名称
	Balance int    `json:"balance" bson:"balance"` // 账户余额
	Ptype   int    `bson:"ptype" json:"ptype"`     // 类型 1：原生；2：唤醒；0：未知
	// Pattern int       `json:"pattern" bson:"pattern"` // 模式:0：随机模式；1：权重模式
	Status       int       `bson:"status" json:"status"`              // 充值开关 0: 关; 1: 开;
	Wstatus      int       `json:"wstatus" bson:"wstatus"`            // 提现开关 0: 关; 1: 开;
	PayRate      float64   `json:"payRate" bson:"pay_rate"`           // 代收税率
	WithdrawRate float64   `json:"withdrawRate" bson:"withdraw_rate"` // 代付税率
	WithdrawFee  int64     `json:"withdrawFee" bson:"withdraw_fee"`   //代付固定手续费
	SortId       int       `bson:"sort_id" json:"sort_id"`            // 权重
	Ctime        time.Time `bson:"ctime"`                             // 修改时间
	FBalance     float64   `bson:"-"`
	FWithdrawFee float64   `bson:"-"`
}

// 游戏房间
type Game struct {
	Unique   string `json:"unique"`
	DeskType int32  `json:"desk_type"`
	// Game 游戏金币房间配置
	Id           string `bson:"_id" json:"_id"`                   //房间ID
	Name         string `bson:"name" json:"name"`                 //房间名称
	Gtype        int32  `bson:"gtype" json:"gtype"`               //游戏类型 1：TP; 2：LHD; 3：SEVEN; 4:RUMMY; 5:AK47; 6:JOKER; 7:CRASH
	Status       int    `bson:"status" json:"status"`             //房间开关 0: 关; 1: 开;
	Ai_Status    int    `bson:"ai_status" json:"ai_status"`       //房间AI开关 0: 关; 1: 开;
	Algo_Switch  int    `bson:"algo_switch" json:"algo_switch"`   //算法（1=old,2=new）
	Kick_Score   int    `json:"kick_score" bson:"kick_score"`     //踢人分数
	Min_Access   int    `json:"min_access" bson:"min_access"`     //最低准入
	Max_Access   int    `json:"max_access" bson:"max_access"`     //最高准入
	Stock_Expect int64  `json:"stock_expect" bson:"stock_expect"` //库存期望
	Stock_Alarm  int    `json:"stock_alarm" bson:"stock_alarm"`   //库存报警
	SortId       int    `json:"sortId" bson:"sort_id"`            //房间分组内排序标识
	Count        uint32 `bson:"count" json:"count"`               //房间人数上限
	//ID bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	// Id     string    `bson:"_id" json:"id"`        //unique ID
	// Gtype  int32     `bson:"gtype" json:"gtype"`   //游戏类型1 niu,2 san,3 jiu
	Rtype int32 `bson:"rtype" json:"rtype"` //房间类型0免佣,1抽佣
	Dtype int32 `bson:"dtype" json:"dtype"` //桌子类型
	Ltype int32 `bson:"ltype" json:"ltype"` //彩票类型1bjpk10,1mlaft
	// Name   string    `bson:"name" json:"name"`     //房间名称
	// Status uint32    `bson:"status" json:"status"` //房间状态1打开,2关闭,3隐藏
	// Count  uint32    `bson:"count" json:"count"`   //房间限制人数
	Ante  uint32    `bson:"ante" json:"ante"`   //房间底分
	Cost  uint32    `bson:"cost" json:"cost"`   //房间抽佣百分比
	Vip   uint32    `bson:"vip" json:"vip"`     //房间vip限制
	Chip  uint32    `bson:"chip" json:"chip"`   //房间进入筹码限制
	Deal  bool      `bson:"deal" json:"deal"`   //房间是否可以上庄
	Carry uint32    `bson:"carry" json:"carry"` //房间上庄最小携带筹码限制
	Down  uint32    `bson:"down" json:"down"`   //房间下庄最小携带筹码限制
	Top   uint32    `bson:"top" json:"Top"`     //房间下庄最大携带筹码限制
	Sit   uint32    `bson:"sit" json:"sit"`     //房间内坐下限制
	Del   int       `bson:"del" json:"del"`     //是否移除
	Node  string    `bson:"node" json:"node"`   //所在节点(game.huiyin1|game.huiyin2)
	Ctime time.Time `bson:"ctime" json:"ctime"` //创建时间
	//Num   uint32    `bson:"num" json:"num"`      //启动房间数量
	Minimum       int64  `bson:"minimum" json:"minimum"`                 //房间最低限制
	Maximum       int64  `bson:"maximum" json:"maximum"`                 //房间最高限制
	Pub           bool   `bson:"pub" json:"pub"`                         //公开展示
	Mode          uint32 `bson:"mode" json:"mode"`                       //模式，0普通，1疯狂
	Multiple      uint32 `bson:"multiple" json:"multiple"`               //倍数，0低，1中，2高
	BreakingPrice int    `bson:"break_price" json:"break_price"`         //破产礼包价格
	BreakingGive  int    `bson:"break_give" json:"break_give"`           //破产礼包赠送
	BreakingTimes int    `bson:"break_times" json:"break_times"`         //破产礼包每日次数
	RoomType      int    `bson:"room_type" json:"room_type"`             // 房间类型(0匹配,1对战房,2对战房娱乐模式)
	MinFirstEntry int    `bson:"min_first_entry" json:"min_first_entry"` // 初始携带要求
	GoldMingTax   int    `bson:"gold_ming_tax" json:"gold_ming_tax"`     // 对战真金模式明税（万分比）
	TP            TPGame //tp
}

type TPGame struct {
	Bottom            int       `json:"bottom" bson:"bottom"`                         //底注
	Rounds            int       `json:"rounds" bson:"rounds"`                         //轮次上限
	Than_Rounds       int       `json:"than_rounds" bson:"than_rounds"`               //比牌轮次
	Pool_Limit        int       `json:"pool_limit" bson:"pool_limit"`                 //筹码池上限
	Otime             int       `json:"otime" bson:"otime"`                           //操作时间
	Tcountdown        int       `json:"tcountdown" bson:"tcountdown"`                 //比牌倒计时
	Scountdown        int       `json:"scountdown" bson:"scountdown"`                 //结算倒数时间
	Match_Time        []int     `json:"match_time" bson:"match_time"`                 //匹配时间区间
	Single_Robot      []int     `json:"single_robot" bson:"single_robot"`             //单局人机数区间
	Robot_Join        []int     `json:"robot_join" bson:"robot_join"`                 //人机加入概率
	Robot_Leave       int       `json:"robot_leave" bson:"robot_leave"`               //人机离开概率(万分比)
	Prevent_Time      int       `json:"prevent_time" bson:"prevent_time"`             //防作弊匹配时间检测
	Prevent_Num       int       `json:"prevent_num" bson:"prevent_num"`               //防作弊匹配次数检测
	Prevent_Thaw      int       `json:"prevent_thaw" bson:"prevent_thaw"`             //防作弊匹配解冻时间
	Msg_Score         int       `json:"msg_score" bson:"msg_score"`                   //跑马灯显示分
	Ctime             time.Time `bson:"ctime"`                                        //修改时间
	NewbieType        int       `json:"newbie_type" bson:"newbie_type"`               //新手状态使用牌型
	AbnormalType      int       `json:"abnormal_type" bson:"abnormal_type"`           //异常状态使用牌型
	FinalFactorRange  []int     `json:"final_factor_range" bson:"final_factor_range"` //正常状态和点控最终系数区间
	CardTypeRange     []int     `json:"card_type_range" bson:"card_type_range"`       //正常状态和点控系数对应使用牌型
	MingTax           int32     `json:"ming_tax" bson:"ming_tax"`                     //明税
	AnTax             int32     `json:"an_tax" bson:"an_tax"`                         //暗税
	RoomRecharge      []int32   `json:"room_recharge" bson:"room_recharge"`           //局内充值
	RoomRechargeGive  []int32   `json:"room_recharge_give" bson:"room_recharge_give"` //局内充值赠送
	AnteFactor        int       `json:"ante_factor" bson:"ante_factor"`               //底注系数
	ShowSwitch        int       `json:"show_switch" bson:"show_switch"`               //结算摊牌开关
	Strategy100Switch int       `json:"strategy100_switch" bson:"strategy100_switch"` //策略100开关
	Strategy200Switch int       `json:"strategy200_switch" bson:"strategy200_switch"` //策略200开关
	Strategy300Switch int       `json:"strategy300_switch" bson:"strategy300_switch"` //策略300开关
	Strategy100Round  int       `json:"strategy100_round" bson:"strategy100_round"`   //策略100累计局数
	Strategy200Round  int       `json:"strategy200_round" bson:"strategy200_round"`   //策略200累计局数
	Strategy300Round  int       `json:"strategy300_round" bson:"strategy300_round"`   //策略300累计局数
	RoomFactorSwitch  int       `json:"room_factor_switch" bson:"room_factor_switch"` //房间系数开关
	ACardType         int       `json:"a_card_type" bson:"a_card_type"`               //A类指定牌型
	// CardType           []TPGameCardType           //发牌牌型配置
	// RobotStrategyGroup []TPGameRobotStrategyGroup //人机策略组配置
	// RobotStrategy      []TPGameRobotStrategy      //人机策略配置
	// NewbieMode         TPNewbieMode               //B类新手模式
	// ANewbieMode        TPNewbieMode               //A类新手模式
	// ControlType        []TPControlType            //控制方式
}
