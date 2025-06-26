package main

import (
	"errors"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	"github.com/globalsign/mgo/bson"
)

// 用户各类游戏对局统计第一次查全部
func UserGameDataFull() {
	startTime := utils.Str2Time("2023-10-01 00:00:00", location)
	endTime := utils.Str2Time("2024-05-17 00:00:00", location)

	gtypeUserRounds := make(map[int]map[string]int, 16)

	// 先清空表
	// data.UserGameDatas.Remove(bson.M{})

	// 查外接游戏局数
	glog.Infof("查询外接游戏局数 start: ....")
	m2 := []bson.M{
		{"$match": bson.M{"amount": bson.M{"$ne": 0}}},
		{"$group": bson.M{"_id": bson.M{"round_id": "$round_id", "game_id": "$game_id", "user_id": "$user_id"}}},
		{"$group": bson.M{"_id": bson.M{"user_id": "$_id.user_id", "game_id": "$_id.game_id"}, "rounds": bson.M{"$sum": 1}}},
	}
	r2 := []bson.M{}
	err = data.NsqLogExternalBets.Pipe(m2).All(&r2)
	if err != nil {
		glog.Error("查询外接游戏局数 fail err: ", err)
		return
	}
	glog.Infof("查询外接游戏局数 finish: %d", len(r2))
	for _, item := range r2 {
		userid := item["_id"].(bson.M)["user_id"].(string)
		game_id := item["_id"].(bson.M)["game_id"].(int)
		rounds := item["rounds"].(int)

		if _, ok := gtypeUserRounds[game_id]; !ok {
			gtypeUserRounds[game_id] = make(map[string]int, 1024)
		}
		gtypeUserRounds[game_id][userid] += rounds
	}

	// 10天10天的查对局详情
	start, end := startTime, startTime.AddDate(0, 0, 10)
	for ; start.Before(endTime); start, end = end, end.AddDate(0, 0, 10) {
		m1 := []bson.M{
			{"$match": bson.M{
				"begin_time": bson.M{"$gte": start.Unix(), "$lt": end.Unix()},
				"players":    bson.M{"$ne": ""},
			}},
			{"$project": bson.M{"gtype": "$gtype", "userids": bson.M{"$split": []string{"$players", ","}}}},
		}

		var details []bson.M
		err := data.Details.Pipe(m1).All(&details)
		if err != nil {
			glog.Errorf("", err)
			return
		}

		glog.Infof("用户对局统计: %v ~ %v, details=%d", start, end, len(details))

		for _, detail := range details {
			gtype := detail["gtype"].(int)
			if _, ok := gtypeUserRounds[gtype]; !ok {
				gtypeUserRounds[gtype] = make(map[string]int, 1024)
			}
			for _, uid := range detail["userids"].([]interface{}) {
				userid := uid.(string)
				if len(userid) >= 16 { // 人机id18位长度+
					continue
				}
				gtypeUserRounds[gtype][userid]++
			}
		}
	}

	glog.Infof("写入数据: %d", len(gtypeUserRounds))
	for gtype, ur := range gtypeUserRounds {
		for userid, round := range ur {
			userGame := &data.UserGameData{
				Id:     bson.NewObjectId().Hex(),
				UserId: userid,
				Gtype:  int64(gtype),
				Number: int64(round),
			}

			addErr := AddOrUpdateUserGame(userGame)
			if addErr != nil {
				glog.Error("UserGameData fail err: ", addErr)
			}

			// if !data.Insert(data.UserGameDatas, userGame) {
			// 	glog.Errorf("写入失败: %v", userGame)
			// }
		}
	}
	glog.Infof("写入数据完成: %d", len(gtypeUserRounds))
}

// 新增或更新用户游戏局数
func AddOrUpdateUserGame(addinfo *data.UserGameData) error {
	info := new(data.UserGameData)
	data.GetByQ(data.UserGameDatas, bson.M{"userid": addinfo.UserId, "gtype": addinfo.Gtype}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"number": addinfo.Number,
		}
		if data.Update(data.UserGameDatas, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + info.Id)
	} else {
		// 新增
		addinfo.Id = bson.NewObjectId().Hex()
		if !data.Insert(data.UserGameDatas, addinfo) {
			return errors.New("写入失败:" + addinfo.Id)
		}
		return nil
	}
}
