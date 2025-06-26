package data

import (
	"errors"
	"time"

	"github.com/globalsign/mgo/bson"
)

type UtrStatus int

const (
	UtrStatusPending UtrStatus = iota // 待处理
	UtrStatusSuccess                  // 成功
	UtrStatusFailed                   // 失败
)

type Utr struct {
	Id        string    `bson:"_id"`       //唯一id
	Uid       string    `bson:"uid"`       //用户id
	Orderid   string    `bson:"orderid"`   //订单id
	FileName  string    `bson:"file_name"` //文件名
	Status    int       `bson:"status"`    //状态
	RefNo     string    `bson:"refNo"`     //UTR号码或UPI Ref No
	RefType   string    `bson:"refType"`   //标识是"UTR"还是"UPI"
	PayTime   string    `bson:"payTime"`   //支付时间
	Amount    string    `bson:"amount"`    //支付金额
	Recipient string    `bson:"recipient"` //收款方名称
	CreatedAt time.Time `bson:"createdAt"` //创建时间
	UpdatedAt time.Time `bson:"updatedAt"` //更新时间
	IsNotify  bool      `bson:"is_notify"` //是否通知
}

func (u *Utr) Save() bool {
	u.Id = bson.NewObjectId().String()
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return Insert(Utrs, u)
}

func (u *Utr) Get() {
	GetByQ(Utrs, bson.M{
		"uid":       u.Uid,
		"orderid":   u.Orderid,
		"file_name": u.FileName,
		"status":    u.Status,
	}, u)
}

func (u *Utr) Update() error {
	m := bson.M{"_id": u.Id}
	n := bson.M{"$set": bson.M{
		"status":    u.Status,
		"refNo":     u.RefNo,
		"refType":   u.RefType,
		"payTime":   u.PayTime,
		"amount":    u.Amount,
		"recipient": u.Recipient,
		"updatedAt": time.Now(),
		"is_notify": u.IsNotify,
	}}
	if Update(Utrs, m, n) {
		return nil
	}
	return errors.New("update utr failed")
}

func (u *Utr) UpdateNotify() error {
	m := bson.M{"uid": u.Uid, "status": 2}
	n := bson.M{"$set": bson.M{
		"is_notify": true,
	}}
	if Update(Utrs, m, n) {
		return nil
	}
	return errors.New("update utr failed")
}

func (u *Utr) ListByQ(q bson.M) []Utr {
	var list []Utr
	ListByQ(Utrs, q, &list)
	return list
}

func (u *Utr) Has() bool {
	return Has(Utrs, bson.M{
		"orderid": u.Orderid,
		"refNo":   u.RefNo,
	})
}
