package service

import (
	"errors"
	"goserver/internal/web/admin/app/entity"

	"github.com/globalsign/mgo/bson"
)

type uploadService struct{}

// 上传记录列表
func (this *uploadService) GetUploadRecords(page, pageSize int, m bson.M) ([]entity.UploadRecords, error) {
	var list []entity.UploadRecords
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	err := Uploads.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList(list)
	return list, err
}

func (this *uploadService) GetUploadRecordsTotal(m bson.M) (int64, error) {
	return int64(Count(Uploads, m)), nil
}

func (this *uploadService) chipList(list []entity.UploadRecords) []entity.UploadRecords {
	for k, v := range list {
		switch v.FType {
		case "0":
			v.FTypeName = "GAME"
		case "1":
			v.FTypeName = "TP"
		case "2":
			v.FTypeName = "DRAGON TIGER"
		case "3":
			v.FTypeName = "7UPDOWN"
		case "4":
			v.FTypeName = "RUMMY"
		case "5":
			v.FTypeName = "TP-AK47"
		case "6":
			v.FTypeName = "TP-JOKER"
		case "7":
			v.FTypeName = "CRASH"
		case "8":
			v.FTypeName = "ANDARBAHAR"
		case "9":
			v.FTypeName = "彩票"
		case "10":
			v.FTypeName = "飞机"
		case "11":
			v.FTypeName = "红黑大战"
		case "12":
			v.FTypeName = "RUMMY双人"
		case "13":
			v.FTypeName = "TP2"
		case "14":
			v.FTypeName = "MINES"
		}

		switch v.Status {
		case 0:
			v.StatusName = "失败"
		case 1:
			v.StatusName = "成功"
		}
		list[k] = v
	}
	return list
}

func (this *uploadService) AddUploadRecords(info *entity.UploadRecords) error {
	info.Id = bson.NewObjectId().Hex()
	info.Ctime = bson.Now()
	if !Insert(Uploads, info) {
		return errors.New("写入失败:" + info.Id)
	}
	return nil
}
