package main

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/internal/global/pay/entity"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/globalsign/mgo/bson"
)

// 5月1日-5月9日 玩CRASH超过50局玩家的ID，把A B类标出来
func HandleExportCrash50() {
	begin := utils.Str2Time("2024-05-12 00:00:00", location)
	// begin := utils.Str2Time("2024-04-15 00:00:00")
	end := utils.Str2Time("2024-05-13 00:00:00", location)
	// 对局数据
	m2 := []bson.M{
		{
			"$match": bson.M{
				"begin_time": bson.M{"$gt": begin.Unix(), "$lt": end.Unix()},
				"gtype":      pb.CRASH,
				// "gtype": pb.HUA,
			},
		},
		{
			"$project": bson.M{
				"userids": bson.M{"$split": []string{"$players", ","}},
			},
		},
	}
	details := []bson.M{}
	err = data.Details.Pipe(m2).All(&details)
	if err != nil {
		glog.Error("error1: ", err)
		return
	}

	// playerGtypeTimes := make(map[string]map[int]int32) // 对局游戏类型次数
	playerGameTimes := make(map[string]int32) // 对局次数
	for _, detail := range details {
		for _, userid := range detail["userids"].([]interface{}) {
			uid := userid.(string)
			if len(uid) >= 18 { // 人机id18位长度+
				continue
			}
			playerGameTimes[uid]++
		}
	}
	var userids50 []string
	for userid, times := range playerGameTimes {
		if times >= 10 {
			userids50 = append(userids50, userid)
		}
	}
	if len(userids50) == 0 {
		glog.Warning("查询到用户列表为空")
		return
	}
	// 用户列表
	var players []*entity.PlayerUser
	err = data.PlayerUsers.
		Find(bson.M{"_id": bson.M{"$in": userids50}}).
		All(&players)
	if err != nil {
		glog.Error("error2: ", err)
		return
	}

	f := excelize.NewFile()
	sheet := "Sheet1"
	// 设置表头
	f.SetCellValue(sheet, "A1", "玩家id")
	f.SetCellValue(sheet, "B1", "玩家类型")
	f.SetCellValue(sheet, "C1", "Crash局数")
	line := 1
	for _, player := range players {
		line++
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", line), player.Userid)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", line), getPlayerRegistAreaName(player.RegistArea))
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", line), playerGameTimes[player.Userid])
	}

	// 保存文件
	if err := f.SaveAs(getExportFileName("crash50")); err != nil {
		glog.Error("error3: ", err)
	}
}

func getPlayerRegistAreaName(registArea int) string {
	switch registArea {
	case 0:
		return "A"
	case 1:
		return "B"
	case 2:
		return "C"
	}
	return strconv.Itoa(registArea)
}

func getExportFileName(name string) string {
	now := time.Now().Format("2006-01-02.15.04.05")
	return ExportDir + "/" + name + "_" + fmt.Sprint(now) + ".xlsx"
}
