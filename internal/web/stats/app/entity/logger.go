package entity

import "time"

// 用户ip重复记录
type UserIpRecords struct {
	Id     string    `bson:"_id"`
	Ip     string    `bson:"ip"`     // ip
	Userid []string  `bson:"userid"` //用户id
	Ctime  time.Time `bson:"ctime"`  //create Time
}
