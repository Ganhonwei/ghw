package data

import "goserver/pkg/utils"

const (
	KMISMS = 0
	MTSMS  = 1
	CLSMS  = 2
	XXYSMS = 3

	// 孟加拉
	KMIMJLSMS = 4
)

// 发送验证码记录
type LogSmsRecord struct {
	Id       string `json:"id" bson:"_id"`
	SmsId    string `json:"smsId" bson:"sms_id"`       // 验证码id
	SendTime int64  `json:"sendTime" bson:"send_time"` // 发送时间
	Reslut   bool   `json:"reslut" bson:"reslut"`      // 结果
	SmsCount int    `json:"smsCount" bson:"sms_count"` // 发送数量
	Phone    string `json:"phone" bson:"phone"`        // 电话
	State    string `json:"state" bson:"state"`        // 状态
	Channel  int    `json:"channel" bson:"channel"`    // 渠道
	Ctime    int64  `json:"ctime" bson:"ctime"`        // 创建时间
}

func (log *LogSmsRecord) Save() {
	log.Ctime = utils.BsonNow().Unix()
	Insert(LogSmsRecords, log)
}

// 用户通过输入验证码进入游戏记录
type LogUserSms struct {
	Id    string `json:"id" bson:"_id"`
	Phone string `json:"phone" bson:"phone"`
	Code  string `json:"code" bson:"code"`
	Ctime int64  `json:"ctime" bson:"ctime"`
}

func (log *LogUserSms) Save() {
	Insert(LogUserSmsRecords, log)
}
