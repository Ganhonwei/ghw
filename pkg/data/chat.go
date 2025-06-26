package data

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

type OfflineChatLog struct {
	Uid     string `bson:"uid" json:"uid"`         //uid
	Content string `bson:"content" json:"content"` //内容
	Ctime   int64  `bson:"ctime" json:"ctime"`     //创建时间
	Sender  string `bson:"sender" json:"sender"`   //发送者userid
	Name    string `bson:"name" json:"name"`       //发送者姓名
}

type ChatLog struct {
	Uid      string    `bson:"uid" json:"uid"`           //uid
	Receiver string    `bson:"receiver" json:"receiver"` //接收者userid
	Title    string    `bson:"title" json:"title"`       //标题
	Content  string    `bson:"content" json:"content"`   //内容
	Ctime    time.Time `bson:"ctime" json:"ctime"`       //创建时间
	Sender   string    `bson:"sender" json:"sender"`     //发送者userid
	Name     string    `bson:"name" json:"name"`         //发送者姓名
	Read     bool      `bson:"read" json:"read"`         //是否已读
}

// 后台聊天日志
type FeedBackLog struct {
	Userid      string    `bson:"_id" json:"id"` //id
	ReplyStatus int       `json:"-" bson:"reply_status"`
	Utime       time.Time `bson:"utime" json:"utime"` //最后发送时间
	Logs        []ChatLog `bson:"log" json:"log"`     //记录
}

func (c *FeedBackLog) Save() {
	Upsert(ChatLogs, bson.M{"_id": c.Userid}, c)
}

func (c *FeedBackLog) Get() {
	Get(ChatLogs, c.Userid, c)
}
