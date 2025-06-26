package entity

// 玩家数据
type PlayerUser struct {
	Userid        string  `bson:"_id" json:"userid"`                    // 用户id
	UserType      string  `bson:"-"`                                    // 用户类型名称
	Nickname      string  `bson:"nickname" json:"nickname"`             // 用户昵称
	RealName      string  `bson:"real_name" json:"real_name"`           // 真实姓名
	Photo         string  `bson:"photo" json:"photo"`                   // 头像
	Sex           uint32  `bson:"sex" json:"sex"`                       // 用户性别,男1 女2 非男非女3
	Phone         string  `bson:"phone" json:"phone"`                   // 绑定的手机号码
	Phone2        string  `bson:"phone2" json:"phone2"`                 // 提现绑定的手机号码
	Tourist       string  `bson:"tourist" json:"tourist"`               // 游客
	RegistArea    int     `json:"registArea" bson:"regist_area"`        // ab测试 0:A 1:B
	Auth          string  `bson:"auth" json:"auth"`                     // 密码验证码
	Password      string  `bson:"password" json:"password"`             // MD5密码
	RegistIP      string  `bson:"regist_ip" json:"regist_ip"`           // 注册账户时的IP地址
	LoginIP       string  `bson:"login_ip" json:"login_ip"`             // 登录账户时的IP地址
	Diamond       int64   `bson:"diamond" json:"diamond"`               // 钻石(彩金cash)
	ShadowDiamond int64   `bson:"shadow_diamond" json:"shadow_diamond"` // 隐藏钻石(用户看不见的)
	Coin          int64   `bson:"coin" json:"coin"`                     // 金币(奖励金bonus)
	GiveDiamond   int64   `json:"giveDiamond" bson:"give_diamond"`      // 赠送彩金
	OutDiamond    int64   `json:"outDiamond" bson:"out_diamond"`        // 可提现彩金
	CashWater     uint64  `bson:"cash_water" json:"cash_water"`         // 彩金流水(游戏产生的)
	Status        int     `bson:"status" json:"status"`                 // 正常1  锁定2  黑名单3 白名单4
	Robot         bool    `bson:"robot" json:"robot"`                   // 是否是机器人
	State         int     `json:"state" bson:"state"`                   // 玩家状态 1.新手 2.正常 3.平民
	CustomTypes   string  `json:"customtypes" bson:"custom_types"`      // 自定义用户类型
	LoginTimes    int32   `bson:"login_times" json:"login_times"`       // 登录天数
	OnlineStatus  bool    `bson:"online_status" json:"online_status"`   // 是否在线
	Vip           UserVip `json:"vip" bson:"vip"`                       // vip
}

type UserVip struct {
	Lv            int     `json:"lv" bson:"lv"`                        //当前等级
	Daily         bool    `json:"daily" bson:"daily"`                  //每日奖励
	Weekly        bool    `json:"weekly" bson:"weekly"`                //每周奖励
	WeeklyTime    int64   `json:"weeklyTime" bson:"weekly_time"`       //下次领取周领奖时间
	Weekday       int     `json:"weekday" bson:"weekday"`              //周几
	LevelReward   []int32 `json:"levelReward" bson:"level_reward"`     //升级奖励
	Exp           int64   `json:"exp" bson:"exp"`                      //当前经验值
	WithdrawCount int     `json:"withdrawCount" bson:"withdraw_count"` //每日提现次数
}
