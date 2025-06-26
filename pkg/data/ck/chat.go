package ck

import (
	"goserver/pkg/data/mq"
	"time"

	"github.com/bytedance/sonic"
)

// 客服聊天会话信息
type CustomerChatSession struct {
	Ver          int64  `gorm:"column:ver" json:"ver"`                    // 插入时间戳
	SessionId    int64  `gorm:"column:id" json:"id"`                      // 对话id
	SessionId2   string `gorm:"column:id2" json:"id2"`                    // 对话id string
	Userid       string `gorm:"column:userid" json:"userid"`              // 玩家id
	Username     string `gorm:"column:username" json:"username"`          // 玩家名称
	Customer     string `gorm:"column:customer" json:"customer"`          // 分配客服id
	CustomerName string `gorm:"column:customer_name" json:"customerName"` // 分配客服名称
	Score        int32  `gorm:"column:score" json:"score"`                // 玩家评分1-5, 0未评分
	Ctime        int64  `gorm:"column:ctime" json:"ctime"`                // 对话创建时间毫秒
	Etime        int64  `gorm:"column:etime" json:"etime"`                // 对话结束时间毫秒
	SeatTime     int64  `gorm:"column:seat_time" json:"seatTime"`         // 对话分配时间毫秒
	ReadTime     int64  `gorm:"column:read_time" json:"readTime"`         // 首次已读时间毫秒
	ReplyTime    int64  `gorm:"column:reply_time" json:"replyTime"`       // 首次回复时间毫秒
	ScoreTime    int64  `gorm:"column:score_time" json:"scoreTime"`       // 评分时间毫秒
	QuestionType int32  `gorm:"column:question_type" json:"questionType"` // 问题类型: 1.充值 2.提现 3.故障或疑问 4.其他
	Resolved     bool   `gorm:"column:resolved" json:"resolved"`          // 解决状态
	Remark       string `gorm:"column:remark" json:"remark"`              // 备注
	// Read         bool   `gorm:"column:read" json:"read"`                  // 最新消息是否已读
	// Reply        bool   `gorm:"column:reply" json:"reply"`                // 最新消息是否已回复
	// Customers     []string `gorm:"column:customers" json:"customers"`          // 分配客服记录
	// CustomersTime []int64  `gorm:"column:customers_time" json:"customersTime"` // 分配客服时间记录
}

func (*CustomerChatSession) TableName() string {
	return "col_customer_chat_sessions"
}

func (c *CustomerChatSession) SyncMarshal(session any) (err error) {
	bytes, err := sonic.Marshal(session)
	if err != nil {
		return
	}
	err = sonic.Unmarshal(bytes, c)
	if err != nil {
		return
	}
	c.Ver = time.Now().UnixMilli()
	err = mq.NatsPublish(mq.TopicSyncCustomerChatSession, c)
	return
}

// 客服聊天记录
type CustomerChatMessage struct {
	Ver        int64  `gorm:"column:ver" json:"ver"`                // 插入时间戳
	Id         int64  `gorm:"column:id" json:"id"`                  // 消息id
	Id2        string `gorm:"column:id2" json:"id2"`                // 消息id2
	SessionId  int64  `gorm:"column:session_id" json:"sessionId"`   // 对话id
	SessionId2 string `gorm:"column:session_id2" json:"sessionId2"` // 对话id2
	Userid     string `gorm:"column:userid" json:"userid"`          // userid
	Reply      bool   `gorm:"column:reply" json:"reply"`            // 是否是客服回复
	Sender     string `gorm:"column:sender" json:"sender"`          // 发送者userid
	SenderName string `gorm:"column:sender_name" json:"senderName"` // 发送者名称
	Ctype      int32  `gorm:"column:ctype" json:"ctype"`            // 消息内容类型: 0.文本 1.图片 2.视频
	Content    string `gorm:"column:content" json:"content"`        // 消息内容
	Read       bool   `gorm:"column:read" json:"read"`              // 是否已读
	Ctime      int64  `gorm:"column:ctime" json:"ctime"`            // 创建时间毫秒
	Rtime      int64  `gorm:"column:rtime" json:"rtime"`            // 读消息时间毫秒
	Reader     string `gorm:"column:reader" json:"reader"`          // 接收者userid
	Revoke     bool   `gorm:"column:revoke" json:"revoke"`          // 消息已撤回
	RevokeTime int64  `gorm:"column:revoke_time" json:"revokeTime"` // 消息撤回时间
	Filename   string `gorm:"column:filename" json:"filename"`      // ctype=1/2/3/4时文件名
	Filesize   int64  `gorm:"column:filesize" json:"filesize"`      // ctype=1/2/3/4时文件名
}

func (*CustomerChatMessage) TableName() string {
	return "col_customer_chat_messages"
}

func (c *CustomerChatMessage) SyncMarshal(message any) (err error) {
	bytes, err := sonic.Marshal(message)
	if err != nil {
		return
	}
	err = sonic.Unmarshal(bytes, c)
	if err != nil {
		return
	}
	c.Ver = time.Now().UnixMilli()
	err = mq.NatsPublish(mq.TopicSyncCustomerChatMessage, c)
	return
}

// 客服坐席操作记录
type CustomerSeatLog struct {
	Ver        int64  `gorm:"column:ver" json:"ver"`                // 插入时间戳
	Id         int64  `gorm:"column:id" json:"id"`                  // id
	Customer   string `gorm:"column:customer" json:"customer"`      // 客服id
	Complete   bool   `gorm:"column:complete" json:"complete"`      // 是否包含打开/关闭完整记录
	OpenTime   int64  `gorm:"column:open_time" json:"openTime"`     // 坐席打开时间毫秒
	CloseTime  int64  `gorm:"column:close_time" json:"closeTime"`   // 坐席关闭时间毫秒
	ActiveTime int64  `gorm:"column:active_time" json:"activeTime"` // 在线时间毫秒
	OpenUser   string `gorm:"column:open_user" json:"openUser"`     // 打开用户
	CloseUser  string `gorm:"column:close_user" json:"closeUser"`   // 关闭用户
}

func (*CustomerSeatLog) TableName() string {
	return "col_customer_seat_log"
}

func (c *CustomerSeatLog) SyncMarshal(log any) (err error) {
	bytes, err := sonic.Marshal(log)
	if err != nil {
		return
	}
	err = sonic.Unmarshal(bytes, c)
	if err != nil {
		return
	}
	c.Ver = time.Now().UnixMilli()
	err = mq.NatsPublish(mq.TopicSyncCustomerSeatLog, c)
	return
}
