package data

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type CdKeyConfig struct {
	Key            string    `json:"key" bson:"_id"`                        // CDKEY
	Remark         string    `json:"remark" bson:"remark"`                  // 功能备注
	Diamond        int64     `bson:"diamond" json:"diamond"`                // 彩金
	GiveDiamond    int64     `json:"giveDiamond" bson:"give_diamond"`       // 赠送彩金
	OutDiamond     int64     `json:"outDiamond" bson:"out_diamond"`         // 可提现彩金
	Bouns          int64     `json:"bouns" bson:"bouns"`                    // BOUNS
	UserId         string    `json:"userId" bson:"user_id"`                 // 指定用户ID
	OperatorId     string    `json:"operatorId" bson:"operator_id"`         // 操作人ID
	OperatorName   string    `json:"operatorName" bson:"operator_name"`     // 操作人名称
	ReceivePeoples int32     `json:"receivePeoples" bson:"receive_peoples"` // 领取人数
	Ctime          time.Time `bson:"ctime"`                                 // 创建时间
	Etime          time.Time `bson:"etime"`                                 // 结束时间
	Utime          time.Time `bson:"utime"`                                 // 修改时间
}

func GetCdkConfig(code string) CdKeyConfig {
	config := new(CdKeyConfig)
	GetByQ(Cdks, bson.M{"_id": code, "etime": bson.M{"$gt": bson.Now()}}, &config)
	if config == nil {
		return CdKeyConfig{}
	}
	return *config
}

func (c CdKeyConfig) UpdateReceivePeople() {
	Update(Cdks, bson.M{"_id": c.Key}, bson.M{"$inc": bson.M{"receive_peoples": 1}})
}
