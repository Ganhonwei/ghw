package data

import (
	"errors"
	"time"

	"github.com/globalsign/mgo/bson"
)

const (
	NOTICE_TYPE0 = 0 //购买消息
	NOTICE_TYPE1 = 1 //公告消息
	NOTICE_TYPE2 = 2 //广播消息
	NOTICE_TYPE3 = 3 //系统消息
	NOTICE_TYPE4 = 4 //活动消息
	NOTICE_TYPE5 = 5 //赠送消息
)

const (
	NOTICE_ACT_TYPE0 = 0 //无操作消息
	NOTICE_ACT_TYPE1 = 1 //支付消息
	NOTICE_ACT_TYPE2 = 2 //活动消息
)

const (
	TIPS = 1 // 系统提示
	POP  = 2 // 弹窗
)

// Notice 公告
type Notice struct {
	Id       string    `bson:"_id"`
	Rtype    int       `bson:"rtype"`    //0:跑马灯 1:系统公告
	Language int       `bson:"language"` //1:英语
	Num      int       `bson:"num"`      //发送频率(分)
	Del      int       `bson:"del"`      //是否移除
	Content  string    `bson:"content"`  //广播内容
	Stime    time.Time `bson:"stime"`    //开始时间
	Etime    time.Time `bson:"etime"`    //结束时间
	Ctime    time.Time `bson:"ctime"`    //创建时间
	LastSend int64     `json:"lastSend"` //上次发送时间
}

// Save 保存消息记录
func (t *Notice) Save() bool {
	//t.Id = bson.NewObjectId().String()
	t.Id = ObjectIdString(bson.NewObjectId())
	t.Ctime = bson.Now()
	t.Etime = bson.Now().AddDate(0, 0, 7)
	return Insert(Notices, t)
}

// GetNoticeList 获取公共消息(系统消息)
func GetNoticeList(rtype int) []Notice {
	var list []Notice
	//q := bson.M{"del": 0, "rtype": rtype,
	//	"userid": "",
	//	"etime":  bson.M{"$gt": bson.Now()}}
	q := bson.M{"del": 0}
	ListByQ(Notices, q, &list)
	return list
}

// GetLogNotices 获取玩家消息记录
func GetLogNotices(userid string, page int) ([]*Notice, error) {
	pageSize := 30 //TODO 优化数据量过大情况
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	var list = make([]*Notice, 0)
	err := Notices.
		Find(bson.M{"userid": userid}).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, errors.New("none record")
	}
	return list, nil
}
