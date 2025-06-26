package service

import (
	"goserver/internal/web/stats/app/entity"

	"gopkg.in/mgo.v2/bson"
)

type gameService struct{}

// 获取支付渠道设置列表
func (this *gameService) GetPayChannelList(page, pageSize int, m bson.M) ([]entity.PayChannel, error) {
	var list []entity.PayChannel
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "sort_id", true)
	err := PayChannels.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)

	return list, err
}
