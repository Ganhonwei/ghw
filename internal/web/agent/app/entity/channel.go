package entity

import "time"

// 渠道列表(包名)
type ChannelInfo struct {
	Id        string    `bson:"_id"`
	Name      string    `json:"name" bson:"name"`            // 渠道名称
	Name1     string    `json:"name1" bson:"name1"`          // 别名 比如：fb、google
	SecretKey string    `json:"secretKey" bson:"secret_key"` // 秘钥
	Status    int       `json:"status" bson:"status"`        // 开关状态 1：开；0：关
	Operator  string    `json:"operator" bson:"operator"`    // 操作人
	Ctime     time.Time `bson:"ctime" json:"ctime"`          // 操作时间
}

// 下拉列表渠道数据结构体
type PackageInfo struct {
	Key  string `bson:"-"`
	Name string `bson:"-"`
}

type BloggerAccount struct {
	Userid   string    `bson:"_id" json:"userid"`        // 用户id
	Channel  string    `json:"channel" bson:"channel"`   // 渠道名称
	Operator string    `json:"operator" bson:"operator"` // 操作人
	Ctime    time.Time `bson:"ctime" json:"ctime"`       // 操作时间
}

// 博主渠道、及博主账号ID
type BloggerChannel struct {
	Id    string   `bson:"-"`
	Name  string   `bson:"-"` // 渠道名称
	Name1 string   `bson:"-"` // 别名 比如：fb、google
	Users []string `bson:"-"`
}
