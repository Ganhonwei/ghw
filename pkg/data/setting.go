package data

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

const (
	MAILREPLY    = 1
	IPLIMIT      = 2
	PAYBLACK     = 3
	SERVERSTATUS = 4
	AUTOTRANSFER = 6
	AUTODELIVERY = 7
)

// 功能开关
type SystemSwitch struct {
	Id        string    `bson:"_id" json:"id"`               //ID
	Name      string    `bson:"name" json:"name"`            //名称
	Gtype     int       `json:"gtype" bson:"gtype"`          //游戏ID
	Stype     int       `bson:"stype" json:"stype"`          //系统设置类型 0:游戏图标;1:活动图标;4:控制设置
	Rtype     int       `json:"rtype" bson:"rtype"`          //类型 1:邮箱自动回复；2:IP限制；3:支付黑名单
	PayStatus int       `json:"payStatus" bson:"pay_status"` //是否付费开关 0：全部用户;1：付费用户;2：未付费用户
	Status    int       `bson:"status" json:"status"`        //开关 0: 关; 1: 开;
	SortId    int       `bson:"sort_id" json:"sort_id"`      //权重
	Value     string    `json:"value" bson:"value"`
	Tab       []int     `json:"tab" bson:"tab"` // 页签
	Ctime     time.Time `bson:"ctime"`          //修改时间
}

func (g *SystemSwitch) Save() {
	Insert(SystemSwitchs, g)
}

func GetSettings() []SystemSwitch {
	var list []SystemSwitch
	ListByQ(SystemSwitchs, nil, &list)
	return list
}

func GetSetting(q bson.M) *SystemSwitch {
	var s *SystemSwitch
	GetByQ(SystemSwitchs, q, &s)
	return s
}
