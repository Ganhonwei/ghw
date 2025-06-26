package entity

// 客服配置配置
type CustomerSetting struct {
	Id           int32 `bson:"_id" json:"id"`                     // 客服id
	SeatDispatch int32 `bson:"seat_dispatch" json:"seatDispatch"` // 分单逻辑开关 0自动,1当前未完成对话数最少,2当日累积总对话数,3手动
}

// 客服坐席管理
type CustomerSeat struct {
	Customer       string   `bson:"_id" json:"id"`                     // 客服id
	CustomerName   string   `bson:"customer_name" json:"customerName"` // 客服名称
	Open           bool     `bson:"open" json:"open"`                  // 坐席状态开关
	Ctime          int64    `bson:"ctime" json:"ctime"`                // 创建时间
	Creater        string   `bson:"creater" json:"creater"`            // 创建时间
	QuickPhrases   []string `bson:"quick_phrases" json:"quickPhrases"` // 快捷语
	ActiveSessions int32    `bson:"-" json:"-"`                        // 当前未完成对话数最少
	TotalSessions  int32    `bson:"-" json:"-"`                        // 当日累积总对话数
}

// 客服坐席操作记录
type CustomerSeatLog struct {
	Id         int64  `bson:"_id" json:"id"`                 // id
	Customer   string `bson:"customer" json:"customer"`      // 客服id
	Complete   bool   `bson:"complete" json:"complete"`      // 是否包含打开/关闭完整记录
	OpenTime   int64  `bson:"open_time" json:"openTime"`     // 坐席打开时间毫秒
	CloseTime  int64  `bson:"close_time" json:"closeTime"`   // 坐席关闭时间毫秒
	ActiveTime int64  `bson:"active_time" json:"activeTime"` // 在线时间毫秒
	OpenUser   string `bson:"open_user" json:"openUser"`     // 打开用户
	CloseUser  string `bson:"close_user" json:"closeUser"`   // 关闭用户
}

// 客服聊天会话信息
type CustomerChatSession struct {
	SessionId    int64  `bson:"_id" json:"id"`                     // 对话id
	SessionId2   string `bson:"id2" json:"id2"`                    // 对话id string
	Userid       string `bson:"userid" json:"userid"`              // 玩家id
	Username     string `bson:"username" json:"username"`          // 玩家名称
	Customer     string `bson:"customer" json:"customer"`          // 分配客服id
	CustomerName string `bson:"customer_name" json:"customerName"` // 分配客服名称
	Score        int32  `bson:"score" json:"score"`                // 玩家评分1-5, 0未评分
	Ctime        int64  `bson:"ctime" json:"ctime"`                // 对话创建时间毫秒
	Etime        int64  `bson:"etime" json:"etime"`                // 对话结束时间毫秒
	SeatTime     int64  `bson:"seat_time" json:"seatTime"`         // 对话分配时间毫秒
	ReadTime     int64  `bson:"read_time" json:"readTime"`         // 首次已读时间毫秒
	ReplyTime    int64  `bson:"reply_time" json:"replyTime"`       // 首次回复时间毫秒
	ScoreTime    int64  `bson:"score_time" json:"scoreTime"`       // 评分时间毫秒
	QuestionType int32  `bson:"question_type" json:"questionType"` // 问题类型: 1.充值 2.提现 3.故障或疑问 4.其他
	Resolved     bool   `bson:"resolved" json:"resolved"`          // 解决状态
	Remark       string `bson:"remark" json:"remark"`              // 备注
	// Read         bool   `gorm:"column:read" json:"read"`           // 最新消息是否已读
	// Reply        bool   `gorm:"column:reply" json:"reply"`         // 最新消息是否已回复
	// Customers     []string `bson:"customers" json:"customers"`          // 分配客服记录
	// CustomersTime []int64  `bson:"customers_time" json:"customersTime"` // 分配客服时间记录
}

// 客服聊天记录
type CustomerChatMessage struct {
	Id         int64  `bson:"_id" json:"id"`                 // 消息id
	Id2        string `bson:"id2" json:"id2"`                // 消息id2
	SessionId  int64  `bson:"session_id" json:"sessionId"`   // 对话id
	SessionId2 string `bson:"session_id2" json:"sessionId2"` // 对话id2
	Userid     string `bson:"userid" json:"userid"`          // userid
	Reply      bool   `bson:"reply" json:"reply"`            // 是否是客服回复
	Sender     string `bson:"sender" json:"sender"`          // 发送者userid
	SenderName string `bson:"sender_name" json:"senderName"` // 发送者名称
	Ctype      int32  `bson:"ctype" json:"ctype"`            // 消息内容类型: 0.文本 1.图片 2.视频 3.pdf 4.其他文件
	Content    string `bson:"content" json:"content"`        // 消息内容
	Read       bool   `bson:"read" json:"read"`              // 是否已读
	Ctime      int64  `bson:"ctime" json:"ctime"`            // 创建时间毫秒
	Rtime      int64  `bson:"rtime" json:"rtime"`            // 读消息时间毫秒
	Reader     string `bson:"reader" json:"reader"`          // 读消息userid
	Revoke     bool   `bson:"revoke" json:"revoke"`          // 消息已撤回
	RevokeTime int64  `bson:"revoke_time" json:"revokeTime"` // 消息撤回时间
	Filename   string `bson:"filename" json:"filename"`      // ctype=1/2/3/4时文件名
	Filesize   int64  `bson:"filesize" json:"filesize"`      // ctype=1/2/3/4时文件名
}

// 客服快捷语
type CustomerChatPhrase struct {
	Id       int64  `bson:"_id" json:"id"`            // 快捷语id
	Id2      string `bson:"-" json:"id2"`             // 快捷语id2
	Customer string `bson:"customer" json:"customer"` // 客服id
	Phrase   string `bson:"phrase" json:"phrase"`     // 快捷语
	NextId   int64  `bson:"next_id" json:"nextId"`    // 上一个快捷语排序
	NextId2  string `bson:"-" json:"nextId2"`         // 上一个快捷语排序2
	Ctime    int64  `bson:"ctime" json:"ctime"`       // 创建时间
}

// 客服聊天会话
type CustomerChatSessionRecord struct {
	SessionId2   string                     `json:"sessionId2"`   // 对话id2
	Userid       string                     `json:"userid"`       // userid
	Username     string                     `json:"username"`     // username
	Customer     string                     `json:"customer"`     // customer
	CustomerName string                     `json:"customerName"` // customerName
	Ctime        string                     `json:"ctime"`        // 对话创建时间毫秒
	Etime        int64                      `json:"etime"`        // 对话结束时间毫秒
	Etime2       string                     `json:"etime2"`       // 对话结束时间格式化
	Unread       int32                      `json:"unread"`       // 未读消息数
	Online       bool                       `json:"online"`       // 是否在线
	VipLv        int32                      `json:"vipLv"`        // vip等级
	Photo        string                     `json:"photo"`        // 玩家头像
	LatestMsg    *CustomerChatMessageRecord `json:"latestMsg"`    // 最新消息

	Status       int32  `json:"status"`       // 沟通状态: 1.未查看 2.已查看 3.已结束
	QuestionType int32  `json:"questionType"` // 问题类型: 1.充值 2.提现 3.故障或疑问 4.其他
	Resolved     bool   `json:"resolved"`     // 解决状态
	Remark       string `json:"remark"`       // 备注
}

// 客服聊天记录
type CustomerChatMessageRecord struct {
	Id2        string `json:"id2"`        // 消息id2
	Userid     string `json:"userid"`     // userid
	Reply      bool   `json:"reply"`      // 是否是客服回复
	Sender     string `json:"sender"`     // 发送者userid
	SenderName string `json:"senderName"` // 发送者名称
	Ctype      int32  `json:"ctype"`      // 消息内容类型: 0.文本 1.图片 2.视频
	Content    string `json:"content"`    // 消息内容
	Read       bool   `json:"read"`       // 是否已读
	Ctime      string `json:"ctime"`      // 创建时间毫秒
	Filename   string `json:"filename"`   // ctype=1/2/3/4时文件名
	Filesize   int64  `json:"filesize"`   // ctype=1/2/3/4时文件大小
}

type SystemUserRecord struct {
	Id       string `json:"id"`
	UserName string `json:"userName"`
}

// 客服表现统计
type CustomerExpStat struct {
	Id              string `json:"id"`              // id
	SDate           string `json:"sdate"`           // 日期
	Customer        string `json:"customer"`        // 客服id
	CustomerName    string `json:"customerName"`    // 客服姓名
	SeatTime        string `json:"seatTime"`        // 在班时长(分钟)
	Sessions        int64  `json:"sessions"`        // 分配对话数
	EndSessions     int64  `json:"endSessions"`     // 已完成对话
	ReplyRate       string `json:"replyRate"`       // 应答率
	EndRate         string `json:"endRate"`         // 完成率
	SessionsTimeAvg string `json:"sessionsTimeAvg"` // 单次对话平均时长（s）
	ReplyTimeAvg    string `json:"replyTimeAvg"`    // 首次回复平均时效（s）
	ScoreAvg        string `json:"scoreAvg"`        // 好评度（1-100%）

	ReplySessions  int64 `json:"replySessions"`
	SessionsTimeMs int64 `json:"sessionsTimeMs"`
	ReplyTimeMs    int64 `json:"replyTimeMs"`
	Scores         int64 `json:"scores"`
	ScoresTimes    int64 `json:"scoresTimes"`
}
