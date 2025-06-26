package controllers

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"goserver/gen/pb"
	"goserver/internal/web/admin/app/entity"
	"goserver/internal/web/admin/app/libs"
	"goserver/internal/web/admin/app/service"
	"goserver/pkg/utils"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
	"github.com/globalsign/mgo/bson"
)

type GameController struct {
	BaseController
}

// 库存列表
func (this *GameController) Inventory() {
	page, _ := this.GetInt("page")
	gtype, _ := this.GetInt("gtype")

	if page < 1 {
		page = 1
	}
	if gtype == 0 {
		gtype = 1
	}
	if gtype == 9 || gtype == 8 || gtype == 11 {
		// 彩票 or AB
		gid := strconv.Itoa(gtype)
		m := bson.M{"_id": gid}
		list, _ := service.GameService.GetGameStock(m)
		count := len(list)
		this.Data["list"] = list
		this.Data["count"] = count
	} else {
		m := bson.M{"gtype": gtype}
		gamelist, _ := service.GameService.GetGameList(0, -1, m)
		count, _ := service.GameService.GetGameListTotal(m)
		var list []entity.Stock
		if len(gamelist) > 0 {
			ids := make([]string, 0)
			for _, v := range gamelist {
				ids = append(ids, v.Id)
			}
			m1 := bson.M{"_id": bson.M{"$in": ids}}
			list, _ = service.GameService.GetStockList(0, -1, m1)
			if len(list) > 0 {
				for i := range list {
					for j := range gamelist {
						var user entity.Game
						if list[i].Id == gamelist[j].Id {
							user = gamelist[j]
							list[i].GameInfo = user
							break
						}
						list[i].GameInfo = user
					}
				}
			}
		}
		this.Data["count"] = count
		this.Data["list"] = list
	}

	this.Data["pageTitle"] = "库存管理"
	this.Data["gtype"] = gtype

	// this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("GameController.Inventory", "gtype", gtype), true).ToString()
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "inventoryopt")
	this.Data["isShow"] = this.auth.HasAccessPerm(this.controllerName, "stockinfo")
	this.display()
}

// 库存曲线
func (this *GameController) StockInfo() {
	id := this.GetString("roomId")
	gtype, _ := this.GetInt("gtype")
	isRoom := true
	if id == "" {
		this.checkError(fmt.Errorf("房间ID不能为空"))
	}
	roomlist := make([]entity.Game, 0)
	if gtype != 0 {
		if gtype != 9 || gtype != 8 || gtype != 11 {
			// // 彩票 or AB
			m := bson.M{}
			m["gtype"] = gtype
			roomlist, _ = service.GameService.GetRoomDropList(m)
		}
	}
	var info entity.Stock
	if gtype == 9 || gtype == 8 || gtype == 11 {
		isRoom = false
		// 彩票库存曲线
		info, _ = service.GameService.GetGameStockInfo(strconv.Itoa(gtype))
	} else {
		info, _ = service.GameService.GetStockInfo(id)
	}
	dates := make([]interface{}, 0)
	data := make([]interface{}, 0)  // 要展示的彩金库存
	data1 := make([]interface{}, 0) // 要展示的奖励金库存
	if len(info.History) > 0 {
		// 使用 sort.Slice() 函数对 history 数据进行倒序排序
		sort.Slice(info.History, func(i, j int) bool {
			return info.History[i].Timestamp > info.History[j].Timestamp
		})
		// // 获取当前时间
		now := time.Now()
		// // 计算近三天前的时间
		// threeDaysAgo :=
		c, _ := service.ConvertToIndiaTime(now.AddDate(0, 0, -1).Unix())

		// 输出排序后的 history 数据
		for _, item := range info.History {
			times, _ := service.ConvertToIndiaTime(item.Timestamp) //time.Unix(item.Timestamp, 0)
			// 判断时间是否在近三天范围内
			if times.After(c) || times.Equal(c) {
				// 将 Unix 时间戳转换为 time.Time 类型的时间值
				t, _ := service.ConvertToIndiaTime(item.Timestamp) //time.Unix(item.Timestamp, 0)
				// 格式化时间并输出字符串
				strTime := t.Format("2006-01-02 15:04:05")
				dates = append(dates, strTime)
				// 保留两位小数，并转换为字符串
				data = append(data, fmt.Sprintf("%.2f", item.FCashStock))
				data1 = append(data1, fmt.Sprintf("%.2f", item.FBonusStock))
			}
		}

		for i := 0; i < len(dates)/2; i++ {
			j := len(dates) - i - 1
			// 交换切片中的元素
			dates[i], dates[j] = dates[j], dates[i]
		}
		for i := 0; i < len(data)/2; i++ {
			j := len(data) - i - 1
			// 交换切片中的元素
			data[i], data[j] = data[j], data[i]
		}
		for i := 0; i < len(data1)/2; i++ {
			j := len(data1) - i - 1
			// 交换切片中的元素
			data1[i], data1[j] = data1[j], data1[i]
		}
	}

	// this.checkError(err)
	this.Data["pageTitle"] = "库存曲线"
	this.Data["TimedLabel"] = dates
	this.Data["CashStocks"] = data
	this.Data["BonusStocks"] = data1
	this.Data["roomList"] = roomlist
	this.Data["roomId"] = id
	this.Data["gtype"] = gtype
	this.Data["isRoom"] = isRoom
	// this.Data["gtype"] = gtype
	// this.Data["count"] = count
	// this.Data["list"] = list
	this.display()
}

// 修改库存
func (this *GameController) StockEdit() {
	id := this.GetString("id")
	typeId, _ := this.GetInt("type")
	inputStock := this.GetString("new_stock")
	remark := this.GetString("remark")
	title := ""
	if id == "" {
		this.checkError(fmt.Errorf("房间ID不能为空"))
	}
	stock, err := service.GameService.GetStockInfo(id)
	this.checkError(err)
	if this.isPost() {
		if inputStock == "" {
			this.checkError(fmt.Errorf("库存值不能为空"))
		}
		newStock, err := strconv.Atoi(inputStock)
		if err != nil {
			this.checkError(fmt.Errorf("库存值只能是数字"))

		}
		if remark == "" {
			this.checkError(errors.New("请输入备注!"))
		}
		info := new(pb.ModifyStock)
		info.GameId = stock.Id
		if typeId == 1 {
			info.CashStock = int64(newStock)
			info.BonusStock = stock.BonusStock
		} else {
			info.BonusStock = int64(newStock)
			info.CashStock = stock.CashStock
		}
		info.CashMingTax = stock.CashMingTax
		info.BonusMingTax = stock.BonusMingTax
		info.CashAnTax = stock.CashAnTax
		info.BonusAnTax = stock.BonusAnTax
		result, err := service.GmRequest(pb.WebModifyStock, pb.CONFIG_UPSERT, info)
		beego.Trace("result: ", result)
		if err != nil {
			this.checkError(err)
		}
		if typeId == 1 {
			service.ActionService.Add("stock_cash_edit", this.auth.GetUser().UserName, "", id, id, remark)
		} else {
			service.ActionService.Add("stock_bonus_edit", this.auth.GetUser().UserName, "", id, id, remark)
		}

		this.redirect(beego.URLFor("GameController.Inventory"))
	}
	if typeId == 1 {
		title = "修改彩金库存"
	} else {
		title = "修改奖励金库存"
	}

	this.Data["pageTitle"] = title
	this.Data["type"] = typeId
	this.Data["info"] = stock
	this.display()
}

// 彩票库存修改
func (this *GameController) CPStockEdit() {
	if this.isPost() {
		gtype, _ := this.GetInt("gtype")
		stock, _ := this.GetInt("stock")
		if gtype == 0 {
			this.checkError(fmt.Errorf("游戏ID不正确"))
		}
		info := new(entity.ModifyGameStock)
		info.Gtype = int32(gtype)
		info.Stock = int64(stock)

		result, err := service.GmRequest(pb.WebGameStock, pb.CONFIG_UPSERT, info)
		beego.Trace("result: ", result)
		if err != nil {
			this.checkError(err)
		}
		service.ActionService.Add("stock_cash_edit", this.auth.GetUser().UserName, "", utils.String(gtype), utils.String(gtype), "")
		this.redirect(beego.URLFor("GameController.Inventory", "gtype", stock))
	}
	this.display()
}

// 点控列表
func (this *GameController) Control() {
	page, _ := this.GetInt("page")
	userid := this.GetString("userid")
	typeId, _ := this.GetInt("typeId")
	if page < 1 {
		page = 1
	}
	list := make([]entity.PlayerUser, 0)
	count := int64(len(list))
	if userid != "" {
		info, err := service.PlayerService.GetUser(userid)
		if err != nil {
			flash := beego.NewFlash()
			flash.Error("没有查询到用户ID为" + userid + "的玩家信息！")
			flash.Store(&this.Controller)
		}
		if info.Userid != "" {
			if info.PCScoreComplete > 0 {
				info.PCSchedule = utils.Float64((info.PCScoreComplete / info.PCScore) * 100)
			}
			list = append(list, *info)
			count = int64(len(list))
		}
	} else {
		// 没有输入条件时展示所有点控玩家
		m := bson.M{}
		m["point_control_switch"] = true
		count, _ = service.PlayerService.GetTotal(m)
		list, _ = service.PlayerService.GetList(page, this.pageSize, m)
	}
	this.Data["pageTitle"] = "点控玩家"
	this.Data["userid"] = userid
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["typeId"] = typeId
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("GameController.Control", "typeId", typeId, "userid", userid), true).ToString()
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "controlopt")
	this.display()
}

// 点控设置
func (this *GameController) ControlEdit() {
	userid := this.GetString("userid")
	info := new(entity.PlayerUser)
	if this.isPost() {
		pcswitch, _ := this.GetBool("switch")
		factor := this.GetString("factor")
		score, _ := this.GetInt("score")
		if factor == "" {
			this.checkError(fmt.Errorf("系统设置不能为空"))
		}
		num, err := strconv.Atoi(factor)
		if err != nil {
			this.checkError(fmt.Errorf("系统设置输入内容格式不正确"))
		}
		msg := new(entity.PointControl)
		msg.UserId = userid
		msg.Switch = pcswitch
		msg.Factor = int32(num)
		msg.Score = int64(score)
		err = this.validControl(msg)
		this.checkError(err)
		result, err := service.GmRequest(pb.WebPointControl, pb.CONFIG_UPSERT, msg)
		beego.Trace("result: ", result)
		if err != nil {
			this.checkError(err)
		}
		service.ActionService.Add("update_point_control", this.auth.GetUser().UserName,
			"", userid, userid, "")
		this.redirect(beego.URLFor("GameController.Control", "userid", userid))
	} else {
		if userid != "" {
			info, _ = service.PlayerService.GetUser(userid)
		}
	}
	this.Data["pageTitle"] = "设置点控"
	this.Data["info"] = info
	this.display()
}

// 验证
func (this *GameController) validControl(pc *entity.PointControl) error {
	valid := validation.Validation{}
	// valid.Required(pc.Factor, "factor").Message("系统设置不能为空")
	valid.Range(pc.Factor, 0, 200, "factor").Message("系统设置数值不正确")
	valid.Required(pc.Score, "score").Message("分数设置格式不正确")

	if valid.HasErrors() {
		for _, err := range valid.Errors {
			return errors.New(err.Message)
		}
	}

	return nil
}

// 游戏列表
func (this *GameController) GameList() {
	// status, _ := this.GetInt("status")
	page, _ := this.GetInt("page")
	startDate := this.GetString("start_date")
	endDate := this.GetString("end_date")
	if page < 1 {
		page = 1
	}

	m := service.FindByDate(startDate, endDate, "ctime", "ctime")
	// m["del"] = status
	list, _ := service.GameService.GetGameList(page, this.pageSize, m)
	count, _ := service.GameService.GetGameListTotal(m)

	le := len(list)
	for i := 0; i < le; i++ {
		if len(list[i].Match_Time) > 0 {
			list[i].Match_TimeStr = strings.Join(this.IntToString(list[i].Match_Time), ",")
		}
		if len(list[i].Robot_Join) > 0 {
			list[i].Robot_JoinStr = strings.Join(this.IntToString(list[i].Robot_Join), ",")
		}
		if len(list[i].Robot_Leave) > 0 {
			list[i].Robot_LeaveStr = strings.Join(this.IntToString(list[i].Robot_Leave), ",")
		}
		if len(list[i].Single_Robot) > 0 {
			list[i].Single_RobotStr = strings.Join(this.IntToString(list[i].Single_Robot), ",")
		}
	}

	/*
		p.Match_TimeStr = strings.Join(this.IntToString(p.Match_Time), ",")
			p.Robot_JoinStr = strings.Join(this.IntToString(p.Robot_Join), ",")
			p.Robot_LeaveStr = strings.Join(this.IntToString(p.Robot_Leave), ",")
			p.Single_RobotStr = strings.Join(this.IntToString(p.Single_Robot), ",")
	*/

	this.Data["pageTitle"] = "房间列表"
	// this.Data["status"] = status
	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("GameController.GameList", "start_date", startDate, "end_date", endDate), true).ToString()
	this.Data["startDate"] = startDate
	this.Data["endDate"] = endDate
	this.display()
}

// 添加公告
func (this *GameController) GameAdd() {
	if this.isPost() {
		name := this.GetString("name")
		gtype, _ := this.GetInt("gtype")
		status, _ := this.GetInt("status")
		ai_status, _ := this.GetInt("ai_status")
		min_access, _ := this.GetInt("min_access")
		max_access, _ := this.GetInt("max_access")
		stock_expect, _ := this.GetInt("stock_expect")
		stock_alarm, _ := this.GetInt("stock_alarm")
		sortId, _ := this.GetInt("sortId")
		count, _ := this.GetInt("count")
		bottom, _ := this.GetInt("bottom")
		rounds, _ := this.GetInt("rounds")
		than_rounds, _ := this.GetInt("than_rounds")
		pool_limit, _ := this.GetInt("pool_limit")
		otime, _ := this.GetInt("otime")
		tcountdown, _ := this.GetInt("tcountdown")
		scountdown, _ := this.GetInt("scountdown")
		match_time := this.GetString("match_time")
		single_robot := this.GetString("single_robot")
		robot_join := this.GetString("robot_join")
		robot_leave := this.GetString("robot_leave")
		prevent_time, _ := this.GetInt("prevent_time")
		prevent_num, _ := this.GetInt("prevent_num")
		prevent_thaw, _ := this.GetInt("prevent_thaw")
		msg_score, _ := this.GetInt("msg_score")

		game := new(entity.Game)
		game.Name = name
		game.Gtype = gtype
		game.Status = int(status)
		game.Ai_Status = int(ai_status)
		game.Min_Access = int(min_access)
		game.Max_Access = int(max_access)
		game.Stock_Expect = int(stock_expect)
		game.Stock_Alarm = int(stock_alarm)
		game.SortId = int(sortId)
		game.Count = uint32(count)
		game.Bottom = int(bottom)
		game.Rounds = int(rounds)
		game.Than_Rounds = int(than_rounds)
		game.Pool_Limit = int(pool_limit)
		game.Otime = int(otime)
		game.Tcountdown = int(tcountdown)
		game.Scountdown = int(scountdown)
		game.Match_Time = this.SliceVlaue(match_time)
		game.Single_Robot = this.SliceVlaue(single_robot)
		game.Robot_Join = this.SliceVlaue(robot_join)
		game.Robot_Leave = this.SliceVlaue(robot_leave)
		game.Prevent_Time = int(prevent_time)
		game.Prevent_Num = int(prevent_num)
		game.Prevent_Thaw = int(prevent_thaw)
		game.Msg_Score = int(msg_score)

		fmt.Printf("game %#v\n", game)
		err := this.validGame(game)
		this.checkError(err)
		if err == nil {
			err = service.GameService.AddGame(game)
			this.checkError(err)

			// 通知服务器
			b := make(map[string]entity.Game)
			b[game.Id] = *game
			atype := pb.CONFIG_UPSERT
			_, err = service.GmRequest(pb.WebGame, atype, b)
			this.checkError(err)

			service.ActionService.AddGame(this.auth.GetUser().UserName, game.Id)
		}
		this.redirect(beego.URLFor("GameController.GameList"))
	}

	this.Data["pageTitle"] = "添加房间"
	this.Data["types1"] = entity.GameTypes
	this.Data["types2"] = entity.RoomStatus
	this.Data["types3"] = entity.RoomStatus
	this.display()
}

func (this *GameController) SliceVlaue(val string) []int {
	if val == "" {
		return make([]int, 0)
	}
	strSlice := strings.Split(val, ",")
	intSlice := make([]int, len(strSlice))
	for i, str := range strSlice {
		num, _ := strconv.Atoi(str)
		intSlice[i] = num
	}
	return intSlice
}

func (this *GameController) IntToString(intSlice []int) []string {
	if len(intSlice) <= 0 {
		return make([]string, 0)
	}
	strSlice := make([]string, len(intSlice))
	for i, num := range intSlice {
		strSlice[i] = strconv.Itoa(num)
	}
	return strSlice
}

func (this *GameController) validGame(shop *entity.Game) error {
	valid := validation.Validation{}
	valid.Required(shop.Name, "name").Message("名字不能为空")
	if valid.HasErrors() {
		for _, err := range valid.Errors {
			return errors.New(err.Message)
		}
	}

	return nil
}

// 发布
// func (this *GameController) Game() {
// 	id := this.GetString("id")

// 	shop, err := service.GameService.GetGame(id)
// 	if err != nil {
// 		this.checkError(err)
// 	} else {

// 		b := make(map[string]entity.Game)
// 		b[shop.Id] = shop
// 		atype := pb.CONFIG_UPSERT
// 		_, err = service.GmRequest(pb.WebGame, atype, b)
// 		if err != nil {
// 			flash := beego.NewFlash()
// 			flash.Error(fmt.Sprintf("%v", err))
// 			flash.Store(&this.Controller)
// 		}
// 	}

// 	service.ActionService.Game(this.auth.GetUserName(), id)

// 	this.redirect(beego.URLFor("GameController.GameList"))
// }

// 移除房间
func (this *GameController) GameDel() {
	id := this.GetString("id")

	shop, err := service.GameService.GetGame(id)
	if err != nil {
		this.checkError(err)
	} else {
		// shop.Del = 1
		b := make(map[string]entity.Game)
		b[shop.Id] = shop
		_, err = service.GmRequest(pb.WebGame, pb.CONFIG_DELETE, b)
		if err != nil {
			flash := beego.NewFlash()
			flash.Error(fmt.Sprintf("%v", err))
			flash.Store(&this.Controller)
		} else {
			err3 := service.GameService.DelGame(id)
			this.checkError(err3)
		}
	}

	service.ActionService.DelGame(this.auth.GetUserName(), id)

	this.redirect(beego.URLFor("GameController.GameList"))
}

// 编辑游戏房间
func (this *GameController) GameEdit() {
	id := this.GetString("id")

	if this.isPost() && id != "" {
		// cost, _ := this.GetInt("cost")
		shop, err := service.GameService.GetGame(id)
		this.checkError(err)

		name := this.GetString("name")
		gtype, _ := this.GetInt("gtype")
		status, _ := this.GetInt("status")
		ai_status, _ := this.GetInt("ai_status")
		min_access, _ := this.GetInt("min_access")
		max_access, _ := this.GetInt("max_access")
		stock_expect, _ := this.GetInt("stock_expect")
		stock_alarm, _ := this.GetInt("stock_alarm")
		sortId, _ := this.GetInt("sortId")
		count, _ := this.GetInt("count")
		bottom, _ := this.GetInt("bottom")
		rounds, _ := this.GetInt("rounds")
		than_rounds, _ := this.GetInt("than_rounds")
		pool_limit, _ := this.GetInt("pool_limit")
		otime, _ := this.GetInt("otime")
		tcountdown, _ := this.GetInt("tcountdown")
		scountdown, _ := this.GetInt("scountdown")
		match_time := this.GetString("match_time")
		single_robot := this.GetString("single_robot")
		robot_join := this.GetString("robot_join")
		robot_leave := this.GetString("robot_leave")
		prevent_time, _ := this.GetInt("prevent_time")
		prevent_num, _ := this.GetInt("prevent_num")
		prevent_thaw, _ := this.GetInt("prevent_thaw")
		msg_score, _ := this.GetInt("msg_score")

		shop.Name = name
		shop.Gtype = gtype
		shop.Status = int(status)
		shop.Ai_Status = int(ai_status)
		shop.Min_Access = int(min_access)
		shop.Max_Access = int(max_access)
		shop.Stock_Expect = int(stock_expect)
		shop.Stock_Alarm = int(stock_alarm)
		shop.SortId = int(sortId)
		shop.Count = uint32(count)
		shop.Bottom = int(bottom)
		shop.Rounds = int(rounds)
		shop.Than_Rounds = int(than_rounds)
		shop.Pool_Limit = int(pool_limit)
		shop.Otime = int(otime)
		shop.Tcountdown = int(tcountdown)
		shop.Scountdown = int(scountdown)
		shop.Match_Time = this.SliceVlaue(match_time)
		shop.Single_Robot = this.SliceVlaue(single_robot)
		shop.Robot_Join = this.SliceVlaue(robot_join)
		shop.Robot_Leave = this.SliceVlaue(robot_leave)
		shop.Prevent_Time = int(prevent_time)
		shop.Prevent_Num = int(prevent_num)
		shop.Prevent_Thaw = int(prevent_thaw)
		shop.Msg_Score = int(msg_score)

		err3 := service.GameService.UpdateGame(&shop)
		this.checkError(err3)
		// 通知服务器
		b := make(map[string]entity.Game)
		b[shop.Id] = shop
		atype := pb.CONFIG_UPSERT
		_, err = service.GmRequest(pb.WebGame, atype, b)
		this.checkError(err)

		service.ActionService.EditGame(this.auth.GetUser().UserName, id)
		this.redirect(beego.URLFor("GameController.GameList"))
	} else {
		p, err := service.GameService.GetGame(id)
		this.checkError(err)
		p.Match_TimeStr = strings.Join(this.IntToString(p.Match_Time), ",")
		p.Robot_JoinStr = strings.Join(this.IntToString(p.Robot_Join), ",")
		p.Robot_LeaveStr = strings.Join(this.IntToString(p.Robot_Leave), ",")
		p.Single_RobotStr = strings.Join(this.IntToString(p.Single_Robot), ",")
		this.Data["game"] = p
		this.Data["types1"] = entity.GameTypes
		this.Data["types2"] = entity.RoomStatus
		this.Data["types3"] = entity.RoomStatus
		this.Data["pageTitle"] = "编辑房间"
		this.display()
	}
}

// 游戏系统设置
func (this *GameController) SystemList() {
	stype, _ := this.GetInt("stype")
	page, _ := this.GetInt("page")
	// if page < 1 {
	// 	page = -1
	// }
	// systemservice
	// isPersonnel := this.auth.HasAccessPerm(this.controllerName, "systemservice")
	// if isPersonnel {
	// 	stype = 3
	// }
	page = -1
	count := int64(0)
	if stype == 3 {
		m := bson.M{}
		list, _ := service.GameService.GetPayChannelList(page, -1, m)
		count, _ = service.GameService.GetPayChannelTotal(m)
		this.Data["list"] = list
	} else if stype == 5 {
		//  邮箱设置
		m := bson.M{"stype": 7, "rtype": 5}
		list, _ := service.GameService.GetSystemList(page, -1, m)
		count, _ = service.GameService.GetSystemListTotal(m)
		info := new(entity.GameSystem)
		if len(list) > 0 {
			info = &list[0]
		}
		this.Data["sysInfo"] = info
	} else {
		m := bson.M{"stype": stype}
		list, _ := service.GameService.GetSystemList(page, -1, m)
		count, _ = service.GameService.GetSystemListTotal(m)
		if stype == 4 {
			// 控制设置
			newinfo := new(entity.GameSystem)
			newinfo.IsButton = true
			newinfo.Name = "关服设置"
			newinfo.Stype = 4
			newinfo.Ctime = bson.Now()
			list = append(list, *newinfo)
			count += 1

			var f_list []entity.GameSystem
			for _, item := range list {
				if !(item.Stype == 4 && (item.Rtype == 6 || item.Rtype == 7)) {
					f_list = append(f_list, item)
				} else {
					count--
				}
			}
			list = f_list
		}
		this.Data["list"] = list
	}

	this.Data["pageTitle"] = "系统设置"
	this.Data["stype"] = stype
	this.Data["count"] = count
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("GameController.SystemList"), true).ToString()
	this.Data["isOperation"] = this.auth.HasAccessPerm(this.controllerName, "systemlistopt")
	this.Data["isPersonnel"] = false // isPersonnel
	this.display()
}

// 修改游戏系统设置
func (this *GameController) SystemEdit() {
	id := this.GetString("stype")
	count, _ := this.GetInt("count")
	channel := make([]entity.PayChannel, 0)
	game := make([]entity.GameSystem, 0)
	if this.isPost() {
		id := this.GetString("stype")
		stype, _ := strconv.Atoi(id)

		isPersonnel := this.auth.HasAccessPerm(this.controllerName, "systemservice")
		if isPersonnel && stype != 3 && this.auth.GetUserName() != "admin" {
			this.checkError(fmt.Errorf("你没有该权限！"))
		}
		// cost, _ := this.GetInt("cost")

		// list, _ := service.GameService.GetGameSystem(stype)
		if stype == 0 || stype == 1 {
			// 游戏图标|活动icon
			for i := 0; i < count; i++ {
				s := "row_" + strconv.Itoa(i) + "_id"
				pid := this.GetString(s)
				s = "row_" + strconv.Itoa(i) + "_name"
				pname := this.GetString(s)
				s = "row_" + strconv.Itoa(i) + "_status"
				pstatus, _ := this.GetInt(s)
				s = "row_" + strconv.Itoa(i) + "_sort"
				psort, _ := this.GetInt(s)
				sysInfo := new(entity.GameSystem)
				sysInfo.Id = pid
				sysInfo.Name = pname
				sysInfo.Status = pstatus
				sysInfo.SortId = psort
				sysInfo.Stype = stype
				if stype == 1 {
					// 活动图标
					s = "row_" + strconv.Itoa(i) + "_pay_status"
					paystatus, _ := this.GetInt(s)
					sysInfo.PayStatus = paystatus
				} else {
					// 游戏图标
					s = "row_" + strconv.Itoa(i) + "_gtype"
					gtype, _ := this.GetInt(s)
					s = "row_" + strconv.Itoa(i) + "_tab_1"
					tab1, _ := this.GetInt(s)
					s = "row_" + strconv.Itoa(i) + "_tab_2"
					tab2, _ := this.GetInt(s)
					s = "row_" + strconv.Itoa(i) + "_tab_3"
					tab3, _ := this.GetInt(s)
					sysInfo.Gtype = gtype
					if tab1 != 0 {
						sysInfo.Tab = append(sysInfo.Tab, 1)
					}
					if tab2 != 0 {
						sysInfo.Tab = append(sysInfo.Tab, 2)
					}
					if tab3 != 0 {
						sysInfo.Tab = append(sysInfo.Tab, 3)
					}
				}
				game = append(game, *sysInfo)
			}

			// 通知服务器
			b := make(map[string]entity.GameSystem)
			strId := ""
			for _, item := range game {
				// 循环体
				err := service.GameService.AddOrUpdateGameSystem(&item)
				this.checkError(err)
				// 循环体
				b[item.Id] = item
				strId += item.Id + ","
			}
			_, err2 := service.GmRequest(pb.WebSwitch, pb.CONFIG_UPSERT, b)
			this.checkError(err2)

			service.ActionService.Add("set_system_icon_config", this.auth.GetUser().UserName,
				"", utils.String(strId), utils.String(strId), "")

			this.redirect(beego.URLFor("GameController.SystemList", "stype", stype))
		} else if stype == 2 {
			// 新手设置
			for i := 0; i < count; i++ {
				s := "row_" + strconv.Itoa(i) + "_id"
				pid := this.GetString(s)
				s = "row_" + strconv.Itoa(i) + "_name"
				pname := this.GetString(s)
				s = "row_" + strconv.Itoa(i) + "_limit"
				pstatus, _ := this.GetInt(s)
				sysInfo := new(entity.GameSystem)
				sysInfo.Id = pid
				sysInfo.Name = pname
				sysInfo.Status = pstatus
				sysInfo.Stype = stype
				game = append(game, *sysInfo)
			}
			// 通知服务器
			b := make(map[string]entity.GameSystem)
			strId := ""
			for i := range game {
				// 循环体
				err := service.GameService.AddOrUpdateGameSystem(&game[i])
				this.checkError(err)
				// 循环体
				itemid := game[i].Id
				b[itemid] = game[i]
				strId += itemid + ","
			}
			_, err2 := service.GmRequest(pb.WebSwitch, pb.CONFIG_UPSERT, b)
			this.checkError(err2)

			service.ActionService.Add("set_game_icon_config", this.auth.GetUser().UserName,
				"", utils.String(strId), utils.String(strId), "")

			this.redirect(beego.URLFor("GameController.SystemList", "stype", 2))
		} else if stype == 3 {
			// 支付渠道
			for i := 0; i < count; i++ {
				s := "row_" + strconv.Itoa(i) + "_id"
				pid := this.GetString(s)
				s = "row_" + strconv.Itoa(i) + "_name"
				pname := this.GetString(s)
				s = "row_" + strconv.Itoa(i) + "_status"
				pstatus, _ := this.GetInt(s)
				s = "row_" + strconv.Itoa(i) + "_sort"
				wstatus, _ := this.GetInt(s)
				s = "row_" + strconv.Itoa(i) + "_type"
				ptype, _ := this.GetInt(s)
				s = "row_" + strconv.Itoa(i) + "_payRate"
				ppayrate, _ := this.GetFloat(s)
				s = "row_" + strconv.Itoa(i) + "_withdrawRate"
				pwithdrawrate, _ := this.GetFloat(s)
				s = "row_" + strconv.Itoa(i) + "_withdrawFee"
				pwithdrawfee, _ := this.GetInt(s)
				s = "row_" + strconv.Itoa(i) + "_balance"
				pbalance, _ := this.GetFloat(s)
				balance := 0
				if pbalance != 0 {
					result := pbalance * 100
					balance = int(math.Round(result))
				}
				withdrawfee := 0
				if pwithdrawfee != 0 {
					result := pwithdrawfee * 100
					withdrawfee = int(result)
				}
				payInfo := new(entity.PayChannel)
				payInfo.Wstatus = wstatus
				payInfo.Id = pid
				payInfo.Name = pname
				payInfo.Status = pstatus
				payInfo.Ptype = ptype
				payInfo.SortId = 1
				payInfo.Balance = int(balance)
				payInfo.PayRate = ppayrate
				payInfo.WithdrawRate = pwithdrawrate
				payInfo.WithdrawFee = int64(withdrawfee)
				channel = append(channel, *payInfo)
			}
			// statusNum := 0
			// for _, v := range channel {
			// 	if v.Status == 1 {
			// 		statusNum += 1
			// 	}
			// }
			// if statusNum > 1 {
			// 	this.checkError(fmt.Errorf("支付渠道不能打开多个"))
			// }

			// 通知服务器
			b := make([]entity.PayChannel, 0)
			strId := ""
			for _, i := range channel {
				// 循环体
				err := service.GameService.AddOrUpdatePayChannel(&i)
				this.checkError(err)
				// 循环体
				// itemid := channel[i].Id
				// b[itemid] = channel[i]
				b = append(b, i)
				strId += i.Id + ","
			}
			_, err2 := service.PayRequest(pb.WebPayChannel, "/syncpaychannel", b)
			this.checkError(err2)

			service.ActionService.Add("set_pay_channel", this.auth.GetUser().UserName,
				"", utils.String(strId), utils.String(strId), "")

			this.redirect(beego.URLFor("GameController.SystemList", "stype", 3))
		} else if stype == 4 {
			for i := 0; i < count; i++ {
				s := "row_" + strconv.Itoa(i) + "_id"
				pid := this.GetString(s)
				s = "row_" + strconv.Itoa(i) + "_rtype"
				rtype, _ := this.GetInt(s)
				s = "row_" + strconv.Itoa(i) + "_name"
				pname := this.GetString(s)
				s = "row_" + strconv.Itoa(i) + "_status"
				pstatus, _ := this.GetInt(s)

				sysInfo := new(entity.GameSystem)
				sysInfo.Id = pid
				sysInfo.Name = pname
				sysInfo.Status = pstatus
				sysInfo.Stype = stype
				if rtype != 0 {
					sysInfo.Rtype = rtype
				}
				game = append(game, *sysInfo)
			}
			b := make(map[string]entity.GameSystem)
			strId := ""
			for _, item := range game {
				// 循环体
				err := service.GameService.AddOrUpdateGameSystem(&item)
				this.checkError(err)
				itemid := item.Id
				if item.Rtype != 0 {
					b[itemid] = item
				}
				strId += itemid + ","
			}
			_, err2 := service.GmRequest(pb.WebSwitch, pb.CONFIG_UPSERT, b)
			this.checkError(err2)
			service.ActionService.Add("set_system_controls_config", this.auth.GetUser().UserName,
				"", utils.String(strId), utils.String(strId), "")
			this.redirect(beego.URLFor("GameController.SystemList", "stype", stype))
		}
	} else {
		if id != "" {
			stype, _ := strconv.Atoi(id)
			if stype == 3 {
				//支付渠道
				// tlist, _ := service.GameService.GetPayChannel()
				// // 3021 IcePay更改名称为CloudPay
				// channel = []entity.PayChannel{
				// 	{Id: "3010", Name: "XDPAY", Wstatus: 0, Status: 0, SortId: 1, Ptype: 0},
				// 	{Id: "3011", Name: "MLPAY", Wstatus: 0, Status: 0, SortId: 2, Ptype: 0},
				// 	{Id: "3012", Name: "1916PAY", Wstatus: 0, Status: 0, SortId: 3, Ptype: 0},
				// 	{Id: "3014", Name: "KingPay", Wstatus: 0, Status: 0, SortId: 4, Ptype: 0},
				// 	{Id: "3015", Name: "XFPAY", Wstatus: 0, Status: 0, SortId: 5, Ptype: 0},
				// 	{Id: "3016", Name: "SailsPay", Wstatus: 0, Status: 0, SortId: 6, Ptype: 0},
				// 	{Id: "3017", Name: "FLYPAY", Wstatus: 0, Status: 0, SortId: 7, Ptype: 0},
				// 	{Id: "3018", Name: "UwinPay", Wstatus: 0, Status: 0, SortId: 8, Ptype: 0},
				// 	{Id: "3013", Name: "LetsPay", Wstatus: 0, Status: 0, SortId: 9, Ptype: 0},
				// 	{Id: "3020", Name: "RamaPay", Wstatus: 0, Status: 0, SortId: 10, Ptype: 0},
				// 	{Id: "3021", Name: "CloudPay", Wstatus: 0, Status: 0, SortId: 11, Ptype: 0},
				// 	{Id: "3022", Name: "WePay", Wstatus: 0, Status: 0, SortId: 12, Ptype: 0},
				// 	{Id: "3023", Name: "OePay", Wstatus: 0, Status: 0, SortId: 13, Ptype: 0},
				// 	{Id: "3024", Name: "9sPay", Wstatus: 0, Status: 0, SortId: 14, Ptype: 0},
				// 	{Id: "3025", Name: "BLIZZARDPAY", Wstatus: 0, Status: 0, SortId: 14, Ptype: 0},
				// 	{Id: "3026", Name: "METAGOPAY", Wstatus: 0, Status: 0, SortId: 15, Ptype: 0},
				// 	{Id: "3027", Name: "UNIVERSALPAY", Wstatus: 0, Status: 0, SortId: 15, Ptype: 0},
				// 	{Id: "5001", Name: "MJL-Universe", Wstatus: 0, Status: 0, SortId: 15, Ptype: 0},
				// 	{Id: "5002", Name: "USDT", Wstatus: 0, Status: 0, SortId: 16, Ptype: 0},
				// 	{Id: "5003", Name: "MJL-JY", Wstatus: 0, Status: 0, SortId: 17, Ptype: 0},
				// 	{Id: "5004", Name: "MJL-Transafe", Wstatus: 0, Status: 0, SortId: 18, Ptype: 0},
				// 	{Id: "5005", Name: "MJL-MMPay", Wstatus: 0, Status: 0, SortId: 19, Ptype: 0},
				// 	{Id: "5006", Name: "MJL-SafePay", Wstatus: 0, Status: 0, SortId: 20, Ptype: 0},
				// 	{Id: "5007", Name: "MJL-99Pay", Wstatus: 0, Status: 0, SortId: 21, Ptype: 0},
				// 	{Id: "5008", Name: "MJL-GoPay", Wstatus: 0, Status: 0, SortId: 22, Ptype: 0},
				// 	{Id: "5009", Name: "MJL-SHPay", Wstatus: 0, Status: 0, SortId: 23, Ptype: 0},
				// 	{Id: "5010", Name: "MJL-BLIZZARDPAY", Wstatus: 0, Status: 0, SortId: 24, Ptype: 0},
				// }
				// for i, v := range channel {
				// 	for _, j := range tlist {
				// 		if v.Id == j.Id {
				// 			fbalance := service.Chip2Float(int64(j.Balance))
				// 			// fWithdrawFee := service.Chip2Float(int64(j.WithdrawFee))
				// 			v.Wstatus = j.Wstatus
				// 			v.Status = j.Status
				// 			v.Ptype = j.Ptype
				// 			v.SortId = j.SortId
				// 			v.FBalance = fbalance
				// 			v.PayRate = j.PayRate
				// 			v.WithdrawRate = j.WithdrawRate
				// 			// v.FWithdrawFee = fWithdrawFee
				// 			channel[i] = v
				// 		}
				// 	}
				// }
				// this.Data["channel"] = channel
				// this.Data["count"] = len(channel)
			} else {
				tlist, _ := service.GameService.GetGameSystem(stype)
				game = make([]entity.GameSystem, 0)
				if stype == 0 {
					game = []entity.GameSystem{
						{Gtype: 1, Name: "TeenPatti经典", Stype: 0, Status: 0, SortId: 1, Tab: []int{1}, Tab1: 1},
						{Gtype: 2, Name: "DRAGON TIGER", Stype: 0, Status: 0, SortId: 2, Tab: []int{1}, Tab1: 1},
						{Gtype: 3, Name: "7UPDOWN", Stype: 0, Status: 0, SortId: 3, Tab: []int{1}, Tab1: 1},
						{Gtype: 4, Name: "Rummy", Stype: 0, Status: 0, SortId: 4, Tab: []int{1}, Tab1: 1},
						{Gtype: 5, Name: "TeenPattiAK47", Stype: 0, Status: 0, SortId: 5, Tab: []int{1}, Tab1: 1},
						{Gtype: 6, Name: "TeenPattiJoker", Stype: 0, Status: 0, SortId: 6, Tab: []int{1}, Tab1: 1},
						{Gtype: 7, Name: "CRASH", Stype: 0, Status: 0, SortId: 7, Tab: []int{1}, Tab1: 1},
						{Gtype: 8, Name: "ANDARBAHAR", Stype: 0, Status: 0, SortId: 8, Tab: []int{1}, Tab1: 1},
						{Gtype: 9, Name: "彩票", Stype: 0, Status: 0, SortId: 9, Tab: []int{1}, Tab1: 1},
						{Gtype: 10, Name: "飞机", Stype: 0, Status: 0, SortId: 10, Tab: []int{1}, Tab1: 1},
						{Gtype: 11, Name: "红黑大战", Stype: 0, Status: 0, SortId: 11, Tab: []int{1}, Tab1: 1},
						{Gtype: 12, Name: "Rummy双人", Stype: 0, Status: 0, SortId: 12, Tab: []int{1}, Tab1: 1},
						{Gtype: 13, Name: "TP2", Stype: 0, Status: 0, SortId: 13, Tab: []int{1}, Tab1: 1},
						{Gtype: 14, Name: "MINES", Stype: 0, Status: 0, SortId: 14, Tab: []int{1}, Tab1: 1},

						{Gtype: 600101, Name: "Ganesha Fortune", Stype: 0, Status: 0, SortId: 20, Tab: []int{2}, Tab2: 1},
						{Gtype: 600002, Name: "Lucky Neko", Stype: 0, Status: 0, SortId: 21, Tab: []int{2}, Tab2: 1},
						{Gtype: 600073, Name: "Ganesha Gold", Stype: 0, Status: 0, SortId: 22, Tab: []int{2}, Tab2: 1},
						{Gtype: 600022, Name: "Fortune Ox", Stype: 0, Status: 0, SortId: 23, Tab: []int{2}, Tab2: 1},
						{Gtype: 600039, Name: "Speed Winner", Stype: 0, Status: 0, SortId: 24, Tab: []int{2}, Tab2: 1},
						{Gtype: 600120, Name: "Treasures of Aztec", Stype: 0, Status: 0, SortId: 25, Tab: []int{2}, Tab2: 1},
						{Gtype: 600054, Name: "Wild Bounty Showdown", Stype: 0, Status: 0, SortId: 26, Tab: []int{2}, Tab2: 1},
						{Gtype: 600104, Name: "Rise of Apollo", Stype: 0, Status: 0, SortId: 27, Tab: []int{2}, Tab2: 1},
						{Gtype: 600037, Name: "Fortune Tiger", Stype: 0, Status: 0, SortId: 28, Tab: []int{2}, Tab2: 1},
						{Gtype: 600028, Name: "Destiny of Sun & Moon", Stype: 0, Status: 0, SortId: 29, Tab: []int{2}, Tab2: 1},
						{Gtype: 600041, Name: "Legend of Perseus", Stype: 0, Status: 0, SortId: 30, Tab: []int{2}, Tab2: 1},
						{Gtype: 600098, Name: "CaiShen Wins", Stype: 0, Status: 0, SortId: 31, Tab: []int{2}, Tab2: 1},
						{Gtype: 600025, Name: "Songkran Splash", Stype: 0, Status: 0, SortId: 32, Tab: []int{2}, Tab2: 1},
						{Gtype: 600093, Name: "Asgardian Rising", Stype: 0, Status: 0, SortId: 33, Tab: []int{2}, Tab2: 1},
						{Gtype: 600004, Name: "Wild Bandito", Stype: 0, Status: 0, SortId: 34, Tab: []int{2}, Tab2: 1},
						{Gtype: 600009, Name: "Supermarket Spree", Stype: 0, Status: 0, SortId: 35, Tab: []int{2}, Tab2: 1},
						{Gtype: 600012, Name: "Cocktail Nights", Stype: 0, Status: 0, SortId: 36, Tab: []int{2}, Tab2: 1},
						{Gtype: 600119, Name: "Galactic Gems", Stype: 0, Status: 0, SortId: 37, Tab: []int{2}, Tab2: 1},
						{Gtype: 600110, Name: "Ways of the Qilin", Stype: 0, Status: 0, SortId: 38, Tab: []int{2}, Tab2: 1},
						{Gtype: 600029, Name: "Honey Trap of Diao Chan", Stype: 0, Status: 0, SortId: 39, Tab: []int{2}, Tab2: 1},
						{Gtype: 600081, Name: "Dragon Hatch", Stype: 0, Status: 0, SortId: 40, Tab: []int{2}, Tab2: 1},
						{Gtype: 600086, Name: "Leprechaun Riches", Stype: 0, Status: 0, SortId: 41, Tab: []int{2}, Tab2: 1},
						{Gtype: 600099, Name: "Egypt's Book of Mystery", Stype: 0, Status: 0, SortId: 42, Tab: []int{2}, Tab2: 1},
						{Gtype: 600102, Name: "Dreams of Macau", Stype: 0, Status: 0, SortId: 43, Tab: []int{2}, Tab2: 1},
						{Gtype: 600117, Name: "Queen of Bounty", Stype: 0, Status: 0, SortId: 44, Tab: []int{2}, Tab2: 1},
						{Gtype: 600103, Name: "Candy Bonanza", Stype: 0, Status: 0, SortId: 45, Tab: []int{2}, Tab2: 1},

						{Gtype: 600607, Name: "Auto-Roulette", Stype: 0, Status: 0, SortId: 46, Tab: []int{3}, Tab3: 1},
						{Gtype: 600572, Name: "Lightning Blackjack", Stype: 0, Status: 0, SortId: 47, Tab: []int{3}, Tab3: 1},
						{Gtype: 600526, Name: "Lightning Roulette", Stype: 0, Status: 0, SortId: 48, Tab: []int{3}, Tab3: 1},
						{Gtype: 600513, Name: "Super Sic Bo", Stype: 0, Status: 0, SortId: 49, Tab: []int{3}, Tab3: 1},
						{Gtype: 600594, Name: "Dragon Tiger", Stype: 0, Status: 0, SortId: 50, Tab: []int{3}, Tab3: 1},
						{Gtype: 600583, Name: "Dream Catcher", Stype: 0, Status: 0, SortId: 51, Tab: []int{3}, Tab3: 1},
						{Gtype: 600642, Name: "Fan Tan", Stype: 0, Status: 0, SortId: 52, Tab: []int{3}, Tab3: 1},
						{Gtype: 600536, Name: "Golden Wealth Baccarat", Stype: 0, Status: 0, SortId: 53, Tab: []int{3}, Tab3: 1},
						{Gtype: 600631, Name: "Bac Bo", Stype: 0, Status: 0, SortId: 54, Tab: []int{3}, Tab3: 1},

						{Gtype: 600276, Name: "Fortune Gems", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						{Gtype: 600369, Name: "Fortune Gems 2", Stype: 0, Status: 0, SortId: 56, Tab: []int{3}, Tab3: 1},
						{Gtype: 600402, Name: "Fortune Gems 3", Stype: 0, Status: 0, SortId: 57, Tab: []int{3}, Tab3: 1},
						{Gtype: 600332, Name: "Money Coming", Stype: 0, Status: 0, SortId: 58, Tab: []int{3}, Tab3: 1},
						{Gtype: 602287, Name: "Crazy Time", Stype: 0, Status: 0, SortId: 59, Tab: []int{3}, Tab3: 1},
						{Gtype: 600312, Name: "Jackpot Fishing", Stype: 0, Status: 0, SortId: 60, Tab: []int{3}, Tab3: 1},
						{Gtype: 600247, Name: "Crazy777", Stype: 0, Status: 0, SortId: 61, Tab: []int{3}, Tab3: 1},
						{Gtype: 602687, Name: "Money Coming Expand Bets", Stype: 0, Status: 0, SortId: 62, Tab: []int{3}, Tab3: 1},
						{Gtype: 600326, Name: "Super Ace", Stype: 0, Status: 0, SortId: 63, Tab: []int{3}, Tab3: 1},
						{Gtype: 604137, Name: "Lightning Roulette", Stype: 0, Status: 0, SortId: 64, Tab: []int{3}, Tab3: 1},
						{Gtype: 602344, Name: "Funky Time", Stype: 0, Status: 0, SortId: 65, Tab: []int{3}, Tab3: 1},
						{Gtype: 600339, Name: "Ocean King", Stype: 0, Status: 0, SortId: 66, Tab: []int{3}, Tab3: 1},
						{Gtype: 604278, Name: "Fortune Roulette", Stype: 0, Status: 0, SortId: 67, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600101, Name: "Ganesha Fortune", Stype: 0, Status: 0, SortId: 68, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600009, Name: "Supermarket Spree", Stype: 0, Status: 0, SortId: 69, Tab: []int{3}, Tab3: 1},
						{Gtype: 600325, Name: "Charge Buffalo", Stype: 0, Status: 0, SortId: 70, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600120, Name: "Treasures of Aztec", Stype: 0, Status: 0, SortId: 71, Tab: []int{3}, Tab3: 1},
						{Gtype: 600374, Name: "Fortune Dragon", Stype: 0, Status: 0, SortId: 72, Tab: []int{3}, Tab3: 1},
						{Gtype: 600019, Name: "Fortune Rabbit", Stype: 0, Status: 0, SortId: 73, Tab: []int{3}, Tab3: 1},
						{Gtype: 604266, Name: "Lucky Jaguar", Stype: 0, Status: 0, SortId: 74, Tab: []int{3}, Tab3: 1},
						{Gtype: 600076, Name: "Double Fortune", Stype: 0, Status: 0, SortId: 75, Tab: []int{3}, Tab3: 1},
						{Gtype: 600761, Name: "3 Coin Treasures", Stype: 0, Status: 0, SortId: 76, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600583, Name: "Dream Catcher", Stype: 0, Status: 0, SortId: 77, Tab: []int{3}, Tab3: 1},
						{Gtype: 600561, Name: "Super Andar Bahar", Stype: 0, Status: 0, SortId: 78, Tab: []int{3}, Tab3: 1},
						{Gtype: 600766, Name: "Crazy Hunter", Stype: 0, Status: 0, SortId: 79, Tab: []int{3}, Tab3: 1},
						{Gtype: 600362, Name: "Boxing King", Stype: 0, Status: 0, SortId: 80, Tab: []int{3}, Tab3: 1},

						// {Gtype: 606132, Name: "Coin Tree", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600082, Name: "Vampire's Charm", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600116, Name: "Wild Fireworks", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600322, Name: "Fortune Monkey", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600296, Name: "Jungle King", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 606138, Name: "Fruity Wheel", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600399, Name: "Jackpot Joker", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600307, Name: "Mega Ace", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600302, Name: "Fa Fa Fa", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 201001, Name: "Golden Bank", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600318, Name: "World Cup", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600100, Name: "Mahjong Ways 2", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// // {Gtype: 600004, Name: "Jurassic Kingdom", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 606141, Name: "Go For Champion", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600271, Name: "Golden Empire", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600218, Name: "Ali Baba", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600092, Name: "Mahjong Ways", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600311, Name: "Bubble Beauty", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600095, Name: "Fortune Mouse", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 606139, Name: "Treasure Quest", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 606140, Name: "Fortune Gems Scratch", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600304, Name: "SevenSevenSeven", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600036, Name: "Butterfly Blossom", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600366, Name: "Golden Joker", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600071, Name: "Jungle Delight", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600080, Name: "Captain's Bounty", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600396, Name: "Devil Fire 2", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 606123, Name: "Golden Bank 2", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 606122, Name: "Safari Mystery", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600255, Name: "Lucky Goldbricks", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600013, Name: "Guardians of Ice & Fire", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600052, Name: "Forge of Wealth", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600278, Name: "Book of Gold", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600764, Name: "Potion Wizard", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600053, Name: "Wild Coaster", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600269, Name: "Medusa", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600308, Name: "Samba", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 604297, Name: "Lucky Doggy", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600397, Name: "Zeus", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
						// {Gtype: 600301, Name: "war of dragons", Stype: 0, Status: 0, SortId: 55, Tab: []int{3}, Tab3: 1},
					}
				} else if stype == 1 {
					game = []entity.GameSystem{
						{Name: "商城罐子", Stype: 1, Status: 0, PayStatus: 0, SortId: 1},
						{Name: "VIP", Stype: 1, Status: 0, PayStatus: 0, SortId: 2},
						{Name: "入门礼包", Stype: 1, Status: 0, PayStatus: 0, SortId: 3},
						{Name: "在线奖励", Stype: 1, Status: 0, PayStatus: 0, SortId: 4},
						{Name: "每日签到", Stype: 1, Status: 0, PayStatus: 0, SortId: 5},
						{Name: "任务活动", Stype: 1, Status: 0, PayStatus: 0, SortId: 6},
						{Name: "牌型活动", Stype: 1, Status: 0, PayStatus: 0, SortId: 6},
						{Name: "金银铜卡", Stype: 1, Status: 0, PayStatus: 0, SortId: 7},
						{Name: "首充", Stype: 1, Status: 0, PayStatus: 0, SortId: 8},
						{Name: "分享活动", Stype: 1, Status: 0, PayStatus: 0, SortId: 9},
						{Name: "包赔活动", Stype: 1, Status: 0, PayStatus: 0, SortId: 10},
						{Name: "BUG有奖", Stype: 1, Status: 0, PayStatus: 0, SortId: 11},
						{Name: "限时活动", Stype: 1, Status: 0, PayStatus: 0, SortId: 12},
						{Name: "Slot开关", Stype: 1, Status: 0, PayStatus: 0, SortId: 13},
						{Name: "对战房开关", Stype: 1, Status: 0, PayStatus: 0, SortId: 14},
						{Name: "排行榜开关", Stype: 1, Status: 0, PayStatus: 0, SortId: 15},
						{Name: "小米手机活动开关", Stype: 1, Status: 0, PayStatus: 0, SortId: 16},
						{Name: "真人视讯", Stype: 1, Status: 0, PayStatus: 0, SortId: 17},
						{Name: "大富翁", Stype: 1, Status: 0, PayStatus: 0, SortId: 18},
						{Name: "拼多多", Stype: 1, Status: 0, PayStatus: 0, SortId: 19},
						{Name: "社媒关注奖金活动", Stype: 1, Status: 0, PayStatus: 0, SortId: 20},
						{Name: "VB活动任务", Stype: 1, Status: 0, PayStatus: 0, SortId: 21},
						{Name: "利息获取", Stype: 1, Status: 0, PayStatus: 0, SortId: 22},
						{Name: "输分补偿", Stype: 1, Status: 0, PayStatus: 0, SortId: 23},
					}
				} else if stype == 2 {
					game = []entity.GameSystem{
						{Name: "充值金额上限", Stype: 2, Status: 1000, SortId: 1},
						{Name: "提现金额上限", Stype: 2, Status: 1000, SortId: 2},
						{Name: "总资产上限", Stype: 2, Status: 8000, SortId: 3},
						{Name: "游戏局数上限", Stype: 2, Status: 100, SortId: 4},
					}
				} else if stype == 4 {
					game = []entity.GameSystem{
						{Name: "提现自动审核", Stype: 4, Status: 0, SortId: 1},
						{Name: "提现失败自动退回", Stype: 4, Status: 0, SortId: 1},
						{Name: "邮箱自动回复充提信息", Stype: 4, Rtype: 1, Status: 0, SortId: 2},
						{Name: "国内IP限制", Stype: 4, Rtype: 2, Status: 0, SortId: 3},
						{Name: "支付黑名单", Stype: 4, Rtype: 3, Status: 0, SortId: 4},
						// {Name: "服务器维护", Stype: 4, Rtype: 4, Status: 0, SortId: 5},
						{Name: "开服限制登录", Stype: 4, Rtype: 4, Status: 0, SortId: 5},
						// {Name: "提现自动转单", Stype: 4, Rtype: 6, Status: 1, SortId: 6},
						// {Name: "提现自动派单", Stype: 4, Rtype: 7, Status: 1, SortId: 7},
					}
				}
				for i, v := range game {
					for _, j := range tlist {
						if v.Name == j.Name {
							v.Stype = j.Stype
							v.Status = j.Status
							v.SortId = j.SortId
							v.PayStatus = j.PayStatus
							v.Id = j.Id
							v.Tab = j.Tab
							if j.Gtype != 0 {
								v.Gtype = j.Gtype
							}
							if len(j.Tab) > 0 {
								v.Tab = j.Tab
							}
							game[i] = v
						}
					}
				}
				tabList := []int{1, 2, 3}
				this.Data["tablist"] = tabList
				this.Data["channel"] = game
				this.Data["count"] = len(game)
			}
		}
		this.Data["stype"] = id
		this.Data["pageTitle"] = "编辑"
		this.display()
	}
}

/*
 */
func (this *GameController) CloseGameService() {
	if this.isPost() {
		msg := new(entity.IPwhite)
		result, err := service.GmRequest(pb.WebCloseServer, pb.CONFIG_UPSERT, msg)
		beego.Trace("result: ", result)
		if err != nil {
			this.checkError(err)
		} else {
			this.showMsg("关服成功！", MSG_OK, "")
		}
	}
}

// 邮件提醒设置
func (this *GameController) EmailSettings() {
	email := this.GetString("email")
	id := this.GetString("id")
	if this.isPost() {
		if email == "" {
			this.checkError(fmt.Errorf("邮箱地址不能为空！"))
		}
		parts := strings.Split(email, ";")
		for _, v := range parts {
			if v != "" {
				if !isValidEmail(v) {
					this.checkError(fmt.Errorf("%s邮箱地址格式不正确", v))
				}
			}
		}
		info := new(entity.GameSystem)
		info.Id = id
		info.Rtype = 5
		info.Name = "预警提醒邮箱"
		info.SortId = 0
		info.Stype = 7
		info.Value = email
		err := service.GameService.AddOrUpdateGameSystem(info)
		if err != nil {
			this.checkError(err)
		} else {
			this.showMsg("设置成功！", MSG_OK, beego.URLFor("GameController.SystemList", "stype", 5))
		}
	}
}

func isValidEmail(email string) bool {
	// 定义邮箱地址的正则表达式模式
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	// 编译正则表达式
	reg := regexp.MustCompile(pattern)

	// 使用正则表达式匹配邮箱地址
	return reg.MatchString(email)
}
