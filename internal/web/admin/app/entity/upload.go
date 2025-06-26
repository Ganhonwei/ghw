package entity

import "time"

type UploadRecords struct {
	Id         string    `json:"id" bson:"_id"`                 // id
	Name       string    `json:"name" bson:"name"`              // 上传文件名称
	FType      string    `json:"fType" bson:"f_type"`           // 上传文件类型
	FTypeName  string    `bson:"-"`                             // 上传文件类型
	Status     int       `json:"status" bson:"status"`          // 上传状态 0：失败；1：成功
	StatusName string    `bson:"-"`                             // 上传状态 0：失败；1：成功
	ResMessage string    `json:"resMessage" bson:"res_message"` // 响应消息
	Operator   string    `json:"operator" bson:"operator"`      // 操作人
	Ctime      time.Time `bson:"ctime" json:"ctime"`            // 上传时间
	Remark     string    `json:"remark" bson:"remark"`          // 备注
}
