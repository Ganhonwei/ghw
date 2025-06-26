package data

import (
	"fmt"

	"goserver/pkg/utils"

	"github.com/globalsign/mgo/bson"
)

// 在同一个collection中
const (
	ROOMID_KEY = "last_room_id" //房间唯一id
	USERID_KEY = "last_user_id" //玩家唯一id
	FACEID_KEY = "last_face_id" //头像id
)

type IDGen struct {
	Id   string `bson:"_id"`  //key
	Init string `bson:"init"` //初始值
	Curr string `bson:"curr"` //当前id
	Step string `bson:"step"` //step
	Max  string `bson:"max"`  //max
}

type UidData struct {
	Uid    int64 `bson:"_id"`     //Uid
	IsUsed bool  `bson:"is_used"` //是否使用过
}

func (this *IDGen) Save() bool {
	//glog.Debugf("GenID %#v", this)
	return Upsert(IDGens, bson.M{"_id": this.Id}, this)
}

func (this *IDGen) Get() {
	Get(IDGens, this.Id, this)
}

func (this *IDGen) GenID() string {
	//glog.Debugf("GenID %s, %s, %s, %s", this.Id, this.Curr, this.Step, this.Max)

	if this.Id == ROOMID_KEY || this.Id == USERID_KEY {
		this.Curr = utils.StringAdd(this.Curr)
		if this.Curr >= this.Max {
			this.Max = utils.StringAdd2(this.Curr, this.Step)
			// this.Save()
		}
	} else if this.Id == FACEID_KEY {
		this.Curr = utils.StringAdd2(this.Curr, this.Step)
		if this.Curr == this.Max {
			this.Curr = this.Init
		}
	}
	//暂时每次存储
	this.Save()
	return this.Curr
}

func (this *IDGen) Index() string {
	return this.Curr
}

// 初始化
func InitIDGen(key string) (r *IDGen) {
	r = new(IDGen)
	r.Id = key
	r.Get()

	if key == ROOMID_KEY || key == USERID_KEY {
		if r.Curr == "" {
			switch key {
			case ROOMID_KEY:
				r.Curr = "1"
				r.Step = "100"
				r.Max = "101"
			case USERID_KEY:
				r.Curr = "100000"
				r.Step = "100"
				r.Max = "100100"
			}
		} else {
			r.Curr = r.Max
		}
	} else if key == FACEID_KEY {
		if r.Curr == "" {
			switch key {
			case FACEID_KEY:
				r.Init = "1"
				r.Curr = "0"
				r.Step = "1"
				r.Max = "151"
			}
		}
	}

	return
}

// String returns a hex string representation of the id.
// Example: ObjectId("4d88e15b60f486e428412dc9").
func ObjectIdString(id bson.ObjectId) string {
	return fmt.Sprintf(`ObjectId("%x")`, string(id))
}

func GetUnusedUid() *UidData {
	da := new(UidData)
	GetByQ(Uids, bson.M{"is_used": false}, da)
	return da
}

func (u *UidData) UpdateUid() {
	Update(Uids, bson.M{"_id": u.Uid}, bson.M{"is_used": true})
}
