package service

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"goserver/gen/pb"
	"goserver/internal/web/admin/app/entity"
	"goserver/pkg/data/ck"
	"goserver/pkg/utils"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	"github.com/globalsign/mgo/bson"
)

type playerService struct{}

func (this *playerService) MineChild(id string, child []string) error {
	agent, err := this.GetPlayerAgent(id)
	if err != nil {
		return err
	}
	for _, val := range child {
		if val == agent {
			return nil
		}
	}
	return errors.New("玩家不存在")
}

// 获取上级代理
func (this *playerService) GetPlayerAgent(id string) (string, error) {
	q := bson.M{"_id": id}
	f := []string{"agent"}
	var feeNum bson.M
	GetByQWithFields(PlayerUsers, q, f, &feeNum)
	if v, ok := feeNum["agent"]; ok {
		return v.(string), nil
	}
	return "", errors.New("用户不存在")
}

// 获取
func (this *playerService) GetPlayerCK(userid string) (*entity.PlayerUser, error) {
	player := new(entity.PlayerUser)
	sql1 := `select * from game.col_user cu final WHERE robot = 0 and simulation_robot = 0 and userid = ?`
	var args1 []any
	args1 = append(args1, userid)
	ck.Select(&player, sql1, args1...)
	if player.Userid == "" {
		return nil, errors.New("用户不存在")
	}
	if player.Photo != "" {
		_, err := strconv.Atoi(player.Photo)
		if err == nil {
			// 系统头像
			player.Photo = "0"
			player.PhotoUrl = "" // "assets/avatars/img_" + player.Photo + ".png"
		} else {
			// 自定义上传头像
			player.PhotoUrl = player.Photo
		}
	} else {
		player.Photo = "0"
	}

	return player, nil
}

// 获取
func (this *playerService) GetPlayer(userid string) (*entity.PlayerUser, error) {
	player := new(entity.PlayerUser)
	Get(PlayerUsers, userid, player)
	if player.Userid == "" {
		return nil, errors.New("用户不存在")
	}
	return player, nil
}

// 获取所有
func (this *playerService) GetAllPlayer() ([]entity.PlayerUser, error) {
	return this.GetList(1, -1, bson.M{})
}

// 获取所有下属玩家
func (this *playerService) GetAllBuilds(agent string) []bson.M {
	var list []bson.M
	if agent == "" {
		return list
	}
	q := bson.M{"agent": agent}
	f := []string{"_id"}
	ListByQWithFields(PlayerUsers, q, f, &list)
	return list
}

// 获取所有下属玩家
func (this *playerService) GetAllBuilds2(agent string) (list []string) {
	list = make([]string, 0)
	if agent == "" {
		return list
	}
	list2 := this.GetAllBuilds(agent)
	for _, v := range list2 {
		if v2, ok := v["_id"]; ok {
			list = append(list, v2.(string))
		}
	}
	return list
}

// 获取代理商总数
func (this *playerService) GetBuilds(agent string) int {
	m := bson.M{"agent": agent}
	return Count(PlayerUsers, m)
}

// 获取列表
func (this *playerService) GetList(page, pageSize int, m bson.M) ([]entity.PlayerUser, error) {
	var list []entity.PlayerUser
	if pageSize == -1 {
		pageSize = 100000
	}
	m["robot"] = false
	m["simulation_robot"] = false
	// fmt.Printf("==========user list filter: %v\n", m)
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	err := PlayerUsers.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	//转换为分展示
	list = this.chipList2(list)
	// 查询游戏局数

	return list, err
}

// 获取用户列表
func (this *playerService) GetListCk(page, pageSize int, params map[string]any) (list []entity.PlayerUser, total int, err error) {
	if pageSize == -1 {
		pageSize = 100000
	}
	sql1 := `select %s from game.col_user cu final WHERE robot = 0 and simulation_robot = 0 %s %s`
	// 拼接条件
	where1 := ""
	var args1 []any
	if user_id, ok := params["_id"]; ok {
		where1 += " and userid in ?"
		args1 = append(args1, user_id)
	}
	if starttime, ok := params["starttime"]; ok {
		where1 += " and ctime >= toDateTime(?, ?) "
		args1 = append(args1, starttime)
		args1 = append(args1, locationName)
	}
	if endtime, ok := params["endtime"]; ok {
		where1 += " and ctime <= toDateTime(?, ?) "
		args1 = append(args1, endtime)
		args1 = append(args1, locationName)
	}
	if regist_area, ok := params["regist_area"]; ok {
		where1 += " and regist_area in ?"
		args1 = append(args1, regist_area)
	}
	if nickname, ok := params["nickname"]; ok {
		where1 += " and nickname = ?"
		args1 = append(args1, nickname)
	}
	if phone, ok := params["phone"]; ok {
		where1 += " and phone = ?"
		args1 = append(args1, phone)
	}
	if ip, ok := params["ip"]; ok {
		where1 += " and (regist_ip = ? or login_ip = ?)"
		args1 = append(args1, ip)
		args1 = append(args1, ip)
	}
	if adid, ok := params["ad__adid"]; ok {
		where1 += " and ad__adid = ?"
		args1 = append(args1, adid)
	}
	if banks, ok := params["bank_accounts"]; ok {
		where1 += " and bank_accounts = ?"
		args1 = append(args1, banks)
	}
	if state, ok := params["state"]; ok {
		where1 += " and state = ?"
		args1 = append(args1, state)
	}
	if moneystart, ok := params["money_start"]; ok {
		where1 += " and money >= ?"
		args1 = append(args1, moneystart)
	}
	if moneyend, ok := params["money_end"]; ok {
		where1 += " and money <= ?"
		args1 = append(args1, moneyend)
	}
	if login_start_time, ok := params["login_start_time"]; ok {
		where1 += " and login_time >= toDateTime(?, ?) "
		args1 = append(args1, login_start_time)
		args1 = append(args1, locationName)
	}
	if login_end_time, ok := params["login_end_time"]; ok {
		where1 += " and login_time <= toDateTime(?, ?) "
		args1 = append(args1, login_end_time)
		args1 = append(args1, locationName)
	}
	if ad__bundle_id, ok := params["ad__bundle_id"]; ok {
		where1 += " and ad__bundle_id in ?"
		args1 = append(args1, ad__bundle_id)
	}
	// 查总数
	sql_count := fmt.Sprintf(sql1, "count(*) c", where1, "")
	err = ck.Select(&total, sql_count, args1...)
	if err != nil || total == 0 {
		return
	}

	selects := "*"
	order := " ORDER BY ctime DESC LIMIT ?, ?"
	sql_list := fmt.Sprintf(sql1, selects, where1, order)
	args_list := args1
	offset, limit := PageCalc(page, pageSize)
	args_list = append(args_list, offset, limit)
	err = ck.Select(&list, sql_list, args_list...)
	if err != nil || len(list) == 0 {
		return
	}

	//转换为分展示
	list = this.chipList22(list)
	// 查询游戏局数

	return list, total, err
}

func PlayersGameRounds(userid []string) (rst map[string]int) {

	return nil
}

// 获取黑名单用户列表按照加入黑名单时间排序
func (this *playerService) GetBlackList(page, pageSize int, m bson.M) ([]entity.PlayerUser, error) {
	var list []entity.PlayerUser
	if pageSize == -1 {
		pageSize = 100000
	}
	m["robot"] = false
	m["simulation_robot"] = false
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "status_time", false)
	err := PlayerUsers.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)

	//转换为分展示
	list = this.chipList2(list)
	return list, err
}

// 根据ID获取用户信息
func (this *playerService) GetUser(id string) (*entity.PlayerUser, error) {
	userInfo := new(entity.PlayerUser)
	Get(PlayerUsers, id, userInfo)
	if userInfo.Userid == "" {
		return nil, errors.New("用户不存在")
	}
	userInfo = this.chipList6(userInfo)
	return userInfo, nil
}

// 根据ID获取用户信息
func (this *playerService) GetByUser(ids []string) ([]entity.PlayerUser, error) {
	var list []entity.PlayerUser
	q := bson.M{"_id": bson.M{"$in": ids}}
	err := PlayerUsers.
		Find(q).All(&list)
	return list, err
}

// 转换为分展示
func (this *playerService) chipList6(v *entity.PlayerUser) *entity.PlayerUser {
	if v != nil {
		t := (v.Diamond + v.ShadowDiamond)
		v.Assets = Chip2Float(int64(t))
		v.ShowAmount = Chip2Float(int64(v.Coin + v.Diamond))
		v.FDiamond = Chip2Float(int64(v.Diamond))
		v.FCoin = Chip2Float(int64(v.Coin))
		v.FShadowDiamond = Chip2Float(int64(v.ShadowDiamond))
		v.FMoney = Chip2Float(int64(v.Money))
		v.FCashOut = Chip2Float(int64(v.CashOut))
		w := (v.Diamond + v.ShadowDiamond + int64(v.CashOut)) - int64(v.Money)
		v.FWin = Chip2Float(int64(w))
	}
	return v
}

// 根据条件获取用户信息
func (this *playerService) GetByUserList(m bson.M) ([]entity.PlayerUser, error) {
	var list []entity.PlayerUser
	m["robot"] = false
	m["simulation_robot"] = false
	err := PlayerUsers.
		Find(m).
		All(&list)
	return list, err
}
func (this *playerService) GetByUserIdArray(m bson.M) ([]string, error) {
	var user_arr []string
	m["robot"] = false
	m["simulation_robot"] = false
	err := PlayerUsers.Find(m).Distinct("_id", &user_arr)
	return user_arr, err
}

// 转换为分展示
func (this *playerService) chipList2(list []entity.PlayerUser) []entity.PlayerUser {
	for k, v := range list {
		t := (v.Diamond + v.ShadowDiamond)
		v.Assets = Chip2Float(int64(t))
		v.ShowAmount = Chip2Float(int64(v.Coin + v.Diamond))
		v.FDiamond = Chip2Float(int64(v.Diamond))
		v.FCoin = Chip2Float(int64(v.Coin))
		v.FShadowDiamond = Chip2Float(int64(v.ShadowDiamond))
		v.FMoney = Chip2Float(int64(v.Money))
		v.FGiveDiamond = Chip2Float(int64(v.GiveDiamond))
		v.FOutDiamond = Chip2Float(int64(v.OutDiamond))
		v.FCashOut = Chip2Float(int64(v.CashOut))
		w := (v.Diamond + v.ShadowDiamond + int64(v.CashOut)) - int64(v.Money)
		v.FWin = Chip2Float(int64(w))
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		c1, _ := ConvertToIndiaTime(v.LoginTime.Unix())
		v.LoginTime = c1
		c2, _ := ConvertToIndiaTime(v.StatusTime.Unix())
		v.StatusTime = c2
		if v.FluctuateLine > 0 {
			v.FluctuateLine = v.FluctuateLine // / 100
		}
		// // 检测是否多账号
		// m := bson.M{}
		// m["userid"] = bson.M{
		// 	"$regex":   fmt.Sprintf("\\b%s\\b", v.Userid),
		// 	"$options": "i",
		// }
		// iplist, _ := LoggerService.GetUserIpRecod(m)
		// if len(iplist) > 0 {
		// 	for _, item := range iplist {
		// 		if len(item.Userid) > 1 {
		// 			v.IsIpRepeat = true
		// 		}
		// 	}
		// }

		// if ipCount > 0 {
		// 	v.IsIpRepeat = true
		// }
		// isIp, _ := LoggerService.CheckAccountByIp(v.Userid)
		// v.IsIpRepeat = isIp

		isAppId := false
		// if v.AD_ADID != "" {
		// 	m := bson.M{
		// 		"_id":      bson.M{"$ne": v.Userid},
		// 		"ad__adid": v.AD_ADID,
		// 	}
		// 	appidCount, _ := this.GetTotal(m)
		// 	if appidCount > 0 {
		// 		isAppId = true
		// 	}
		// }

		v.IsAppIdRepeat = isAppId
		if v.CustomTypes != "" {
			v.UserType = v.CustomTypes
		} else {
			switch v.State {
			case 1:
				v.UserType = "新手"
			case 2:
				// 用户类型判断
				if v.Money >= 0 && v.Money <= 9999 {
					v.UserType = "零充"
				} else if v.Money >= 10000 && v.Money <= 99999 {
					v.UserType = "普R"
				} else if v.Money >= 100000 && v.Money <= 499999 {
					v.UserType = "小R"
				} else if v.Money >= 500000 && v.Money <= 999999 {
					v.UserType = "中R"
				} else if v.Money >= 1000000 && v.Money <= 9999999 {
					v.UserType = "大R"
				} else if v.Money >= 10000000 {
					v.UserType = "超大R"
				}
			case 3:
				v.UserType = "平民"
			case 4:
				v.UserType = "泡沫"
			}
		}
		pipeline := []bson.M{
			{
				"$match": bson.M{
					"userid": v.Userid,
				},
			},
			{
				"$group": bson.M{
					"_id": "$userid",
					"total": bson.M{
						"$sum": "$number",
					},
				},
			},
		}
		result := []bson.M{}
		UserGameDatas.Pipe(pipeline).All(&result)
		if len(result) > 0 {
			v.GameNumber = result[0]["total"].(int64)
		}

		list[k] = v
	}
	return list
}

func (this *playerService) chipList22(list []entity.PlayerUser) []entity.PlayerUser {
	for k, v := range list {
		t := (v.Diamond + v.ShadowDiamond)
		v.Assets = Chip2Float(int64(t))
		v.ShowAmount = Chip2Float(int64(v.Coin + v.Diamond))
		v.FDiamond = Chip2Float(int64(v.Diamond))
		v.FCoin = Chip2Float(int64(v.Coin))
		v.FShadowDiamond = Chip2Float(int64(v.ShadowDiamond))
		v.FMoney = Chip2Float(int64(v.Money))
		v.FGiveDiamond = Chip2Float(int64(v.GiveDiamond))
		v.FOutDiamond = Chip2Float(int64(v.OutDiamond))
		v.FCashOut = Chip2Float(int64(v.CashOut))
		w := (v.Diamond + v.ShadowDiamond + int64(v.CashOut)) - int64(v.Money)
		v.FWin = Chip2Float(int64(w))
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		c1, _ := ConvertToIndiaTime(v.LoginTime.Unix())
		v.LoginTime = c1
		c2, _ := ConvertToIndiaTime(v.StatusTime.Unix())
		v.StatusTime = c2
		if v.FluctuateLine > 0 {
			v.FluctuateLine = v.FluctuateLine // / 100
		}
		// // 检测是否多账号
		// m := bson.M{}
		// m["userid"] = bson.M{
		// 	"$regex":   fmt.Sprintf("\\b%s\\b", v.Userid),
		// 	"$options": "i",
		// }
		// iplist, _ := LoggerService.GetUserIpRecod(m)
		// if len(iplist) > 0 {
		// 	for _, item := range iplist {
		// 		if len(item.Userid) > 1 {
		// 			v.IsIpRepeat = true
		// 		}
		// 	}
		// }

		// if ipCount > 0 {
		// 	v.IsIpRepeat = true
		// }
		// isIp, _ := LoggerService.CheckAccountByIp(v.Userid)
		// v.IsIpRepeat = isIp

		isAppId := false
		// if v.AD_ADID != "" {
		// 	m := bson.M{
		// 		"_id":      bson.M{"$ne": v.Userid},
		// 		"ad__adid": v.AD_ADID,
		// 	}
		// 	appidCount, _ := this.GetTotal(m)
		// 	if appidCount > 0 {
		// 		isAppId = true
		// 	}
		// }

		v.IsAppIdRepeat = isAppId
		if v.CustomTypes != "" {
			v.UserType = v.CustomTypes
		} else {
			switch v.State {
			case 1:
				v.UserType = "新手"
			case 2:
				// 用户类型判断
				if v.Money >= 0 && v.Money <= 9999 {
					v.UserType = "零充"
				} else if v.Money >= 10000 && v.Money <= 99999 {
					v.UserType = "普R"
				} else if v.Money >= 100000 && v.Money <= 499999 {
					v.UserType = "小R"
				} else if v.Money >= 500000 && v.Money <= 999999 {
					v.UserType = "中R"
				} else if v.Money >= 1000000 && v.Money <= 9999999 {
					v.UserType = "大R"
				} else if v.Money >= 10000000 {
					v.UserType = "超大R"
				}
			case 3:
				v.UserType = "平民"
			case 4:
				v.UserType = "泡沫"
			}
		}
		pipeline := []bson.M{
			{
				"$match": bson.M{
					"userid": v.Userid,
				},
			},
			{
				"$group": bson.M{
					"_id": "$userid",
					"total": bson.M{
						"$sum": "$number",
					},
				},
			},
		}
		result := []bson.M{}
		UserGameDatas.Pipe(pipeline).All(&result)
		if len(result) > 0 {
			v.GameNumber = result[0]["total"].(int64)
		}

		list[k] = v
	}
	return list
}

// 获取总数
func (this *playerService) GetTotal(m bson.M) (int64, error) {
	m["robot"] = false
	m["simulation_robot"] = false
	return int64(Count(PlayerUsers, m)), nil
}

// 根据游戏局数范围返回userid
func (this *playerService) GetUserGameNumber(startNum, endNum int) ([]string, error) {
	userid_arr := make([]string, 0)
	n := bson.M{}
	if startNum != 0 && endNum != 0 {
		n = bson.M{"$gte": startNum, "$lte": endNum}
	} else if startNum != 0 && endNum == 0 {
		n = bson.M{"$gte": startNum}
	} else if startNum == 0 && endNum != 0 {
		n = bson.M{"$lte": endNum}
	}
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": "$userid",
				"total": bson.M{
					"$sum": "$number",
				},
			},
		},
		{
			"$match": bson.M{
				"total": n,
			},
		},
	}
	result := []bson.M{}
	err := UserGameDatas.Pipe(pipeline).All(&result)
	for _, item := range result {
		id := item["_id"].(string)
		userid_arr = append(userid_arr, id)
	}
	return userid_arr, err
}

// 修改用户信息
func (this *playerService) UpdateUser(user *entity.PlayerUser) error {
	if user.Userid == "" {
		return errors.New("用户ID不能为空")
	}
	m := bson.M{"_id": user.Userid}
	n := bson.M{
		"nickname":       user.Nickname,
		"phone2":         user.Phone2,
		"bank":           user.Bank,
		"ifsc":           user.IFSC,
		"real_name":      user.RealName,
		"bank_accounts":  user.BankAccounts,
		"ad__adid":       user.AD_ADID,
		"ad__key":        user.AD_Key,
		"ad__os_version": user.AD_OsVersion,
		"ad__bundle_id":  user.AD_BundleId,
		// "media_source":     user.MediaSource,
		"ad__app_id":       user.AD_AppId,
		"ad__ref_game_id":  user.AD_RefGameId,
		"ad__ref_pkg_name": user.AD_RefPkgName,
	}
	if Update(PlayerUsers, m, bson.M{"$set": n}) {
		return nil
	}
	return errors.New("更新失败")
}

// 更新用户状态
func (this *playerService) UpdateUserStatus(user *entity.PlayerUser) error {
	if user.Userid == "" {
		return errors.New("用户ID不能为空")
	}
	m := bson.M{"_id": user.Userid}
	n := bson.M{"status": user.Status, "status_time": bson.Now()}
	if Update(PlayerUsers, m, bson.M{"$set": n}) {
		return nil
	}
	return errors.New("更新失败")
}

// 更新用户类型
func (this *playerService) UpdateUserTypes(userid, usertype string) error {
	m := bson.M{"_id": userid}
	n := bson.M{"custom_types": usertype}
	if Update(PlayerUsers, m, bson.M{"$set": n}) {
		return nil
	}
	return errors.New("更新失败")
}

// 获取总数
func (this *playerService) GetIPwhiteTotal(m bson.M) (int64, error) {
	return int64(Count(IPwhites, m)), nil
}

func (this *playerService) GetIPwhite(page, pageSize int, m bson.M) ([]entity.IPwhite, error) {
	var list []entity.IPwhite
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	err := IPwhites.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList9(list)
	return list, err
}

// 添加IP限制
func (this *playerService) AddIPwhite(info *entity.IPwhite) error {
	if info.IP == "" {
		return errors.New("IP不能为空")
	}
	info.Ctime = bson.Now()
	if !Insert(IPwhites, info) {
		return errors.New("写入失败:" + info.IP)
	}
	return nil
}

// 删除IP限制
func (this *playerService) DelIPwhite(info *entity.IPwhite) error {
	if info.IP == "" {
		return errors.New("IP不能为空")
	}
	m := bson.M{"_id": info.IP}
	if Delete(IPwhites, m) {
		return nil
	}
	return errors.New("更新失败")
}

func (this *playerService) GetIPwhiteByIP(id string) (*entity.IPwhite, error) {
	userInfo := new(entity.IPwhite)
	Get(IPwhites, id, userInfo)
	if userInfo.IP == "" {
		return userInfo, errors.New("用户不存在")
	}
	return userInfo, nil
}

/*
	开服IP限制相关接口
*/

// 获取总数
func (this *playerService) GetServerWhiteTotal(m bson.M) (int64, error) {
	return int64(Count(ServerWhites, m)), nil
}

func (this *playerService) GetServerWhite(page, pageSize int, m bson.M) ([]entity.IPwhite, error) {
	var list []entity.IPwhite
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	err := ServerWhites.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList9(list)
	return list, err
}

// 转换为分展示
func (this *playerService) chipList9(list []entity.IPwhite) []entity.IPwhite {
	for k, v := range list {
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		list[k] = v
	}
	return list
}

// 添加IP限制
func (this *playerService) AddServerWhite(info *entity.IPwhite) error {
	if info.IP == "" {
		return errors.New("IP不能为空")
	}
	info.Ctime = bson.Now()
	if !Insert(ServerWhites, info) {
		return errors.New("写入失败:" + info.IP)
	}
	return nil
}

// 删除IP限制
func (this *playerService) DelServerWhite(info *entity.IPwhite) error {
	if info.IP == "" {
		return errors.New("IP不能为空")
	}
	m := bson.M{"_id": info.IP}
	if Delete(ServerWhites, m) {
		return nil
	}
	return errors.New("更新失败")
}

func (this *playerService) GetServerWhiteByIP(id string) (*entity.IPwhite, error) {
	userInfo := new(entity.IPwhite)
	Get(ServerWhites, id, userInfo)
	if userInfo.IP == "" {
		return userInfo, errors.New("用户不存在")
	}
	return userInfo, nil
}

// 后台金币赠送记录
func (this *playerService) AddGoldGiftLog(gift *entity.GoldGiftLog) error {
	gift.Id = bson.NewObjectId().Hex()
	gift.Ctime = bson.Now()
	if !Insert(GoldGiftLogs, gift) {
		return errors.New("写入失败:" + gift.Id)
	}
	return nil
}

/*金币流水*/
// 获取列表
func (this *playerService) GetGoldList(page, pageSize int, m bson.M) ([]entity.LogWater, error) {
	var list []entity.LogWater
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	err := LogWaters.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	//转换为分展示
	list = this.chipList3(list)
	return list, err
}

// 获取总数
func (this *playerService) GetGoldListTotal(m bson.M) (int64, error) {
	return int64(Count(LogWaters, m)), nil
}

// 转换为分展示
func (this *playerService) chipList3(list []entity.LogWater) []entity.LogWater {
	for k, v := range list {
		v.FAddDiamond = Chip2Float(int64(v.AddDiamond))
		v.FAddCoin = Chip2Float(int64(v.AddCoin))
		v.FAddOtherAsset = Chip2Float(int64(v.AddOtherAsset))
		v.FChangeAsset = Chip2Float(int64(v.ChangeAsset))
		v.FOldDiamond = Chip2Float(int64(v.OldDiamond))
		v.FNowDiamond = Chip2Float(int64(v.NowDiamond))
		v.FOldCoin = Chip2Float(int64(v.OldCoin))
		v.FNowCoin = Chip2Float(int64(v.NowCoin))
		v.FOldOtherAsset = Chip2Float(int64(v.OldOtherAsset))
		v.FNowOtherAsset = Chip2Float(int64(v.NowOtherAsset))
		v.OldShowAmount = Chip2Float(int64(v.OldDiamond + v.OldCoin))
		v.NowShowAmount = Chip2Float(int64(v.NowDiamond + v.NowCoin))
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		if v.ChangeAsset < 0 {
			v.IsChangeAsset = true
		} else {
			v.IsChangeAsset = false
		}
		list[k] = v
	}
	return list
}

// 查询符合条件的所有流水记录
func (this *playerService) GetGoldAll(m bson.M) ([]entity.LogWater, error) {
	var list []entity.LogWater
	err := LogWaters.
		Find(m).
		All(&list)
	return list, err
}

/*对局详情*/
// func (this *playerService) GetDetailList(page, pageSize int, m bson.M, n bson.M) ([]entity.Detail, error) {
// 	var list []entity.Detail
// 	var d_list []entity.Detail
// 	var nsq_list []entity.NsqBet
// 	if pageSize == -1 {
// 		pageSize = 100000
// 	}

// 	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "begin_time", false)
// 	NsqLogExternalBets.Find(n).Sort("-ctime").Limit(pageSize).All(&nsq_list)
// 	err := Details.
// 		Find(m).
// 		Sort(sortFieldR).
// 		Skip(skipNum).
// 		Limit(pageSize).
// 		All(&d_list)
// 	for _, item := range nsq_list {
// 		info := new(entity.Detail)
// 		info.WaterId = item.NsqId
// 		info.BeginTime = item.Ctime
// 		info.EndTime = item.Ctime
// 		info.Gtype = item.GameId
// 		info.RoomId = item.MerchantOrderNo
// 		info.Players = item.UserId
// 		list = append(list, *info)
// 	}
// 	list = append(list, d_list...)
// 	//转换为分展示
// 	// list = this.chipList4(list)
// 	return list, err
// }

func (this *playerService) GetDetailList(page, pageSize int, m bson.M, n bson.M) ([]entity.Detail, *entity.TotalDetail, int64, error) {
	var list []entity.Detail
	var d_list []entity.Detail
	var nsq_list []entity.NsqBet
	totalinfo := new(entity.TotalDetail)
	count := int64(0)
	if pageSize == -1 {
		pageSize = 100000
	}
	// skipNum, sortFieldR := parsePageAndSort(page, pageSize, "begin_time", false)
	// 外接数据
	pipeline := []bson.M{
		{
			"$match": n,
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"round_id": "$round_id",
					"ctime":    "$ctime",
				},
			},
		},
		{
			"$project": bson.M{
				"_id":   "$_id.round_id",
				"ctime": "$_id.ctime",
			},
		},
	}
	result := make([]entity.MergeDetail, 0)
	pipe := NsqLogExternalBets.Pipe(pipeline)
	err := pipe.All(&result)
	// 对局数据

	pipeline1 := []bson.M{
		{
			"$match": m,
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"WaterId": "$_id",
					"ctime":   "$begin_time",
				},
			},
		},
		{
			"$project": bson.M{
				"_id":   "$_id.WaterId",
				"ctime": "$_id.ctime",
			},
		},
	}
	result1 := make([]entity.MergeDetail, 0)
	pipe1 := Details.Pipe(pipeline1)
	err = pipe1.All(&result1)
	det_map := make([]entity.MergeDetail, 0)
	det_map = append(det_map, result...)
	det_map = append(det_map, result1...)

	// 对det_map按照ctime字段进行倒序排序
	sort.Slice(det_map, func(i, j int) bool {
		return det_map[i].Ctime > det_map[j].Ctime
	})
	count = int64(len(det_map))
	// 计算分页时的起始索引和结束索引
	startIndex := (page - 1) * pageSize
	endIndex := page * pageSize

	// 对列表进行分页
	var paginatedList []entity.MergeDetail
	if startIndex < len(det_map) {
		if endIndex > len(det_map) {
			endIndex = len(det_map)
		}
		paginatedList = det_map[startIndex:endIndex]
	}
	// // 汇总
	// var totalIds []string
	// for _, v := range det_map {
	// 	totalIds = append(totalIds, v.Id)
	// }
	var ids []string
	for _, v := range paginatedList {
		ids = append(ids, v.Id)
	}
	// 汇总数据
	// var total_nsq_list []entity.NsqBet
	// total_query := bson.M{}
	// total_query["round_id"] = bson.M{"$in": totalIds}
	// total_query["amount"] = bson.M{"$ne": 0}
	// NsqLogExternalBets.Find(total_query).Sort("-ctime").All(&total_nsq_list)
	// // 外接返奖
	// var total_rewardlist []entity.NsqReward
	// NsqLogExternalRewards.Find(total_query).All(&total_rewardlist)
	// totalNSQRewardTable := make(map[string]map[string]entity.NsqReward)
	// for _, reword := range total_rewardlist {
	// 	uidReward, ok := totalNSQRewardTable[reword.RoundId]
	// 	if !ok {
	// 		uidReward = make(map[string]entity.NsqReward)
	// 		totalNSQRewardTable[reword.RoundId] = uidReward
	// 	}
	// 	uidReward[reword.UserId] = reword
	// }
	// var total_list []entity.Detail
	// total_query = bson.M{}
	// total_query["_id"] = bson.M{"$in": totalIds}
	// total_query["players"] = bson.M{"$ne": ""}
	// Details.Find(total_query).Sort("-begin_time").All(&total_list)

	// totalWin := make([]int64, 0)
	// totalLose := make([]int64, 0)
	// for _, item := range total_nsq_list {
	// 	// totalinfo.Number++
	// 	isLose := false
	// 	// 外接返奖信息
	// 	if uidReward, ok := totalNSQRewardTable[item.RoundId]; ok {
	// 		if reward, ok := uidReward[item.UserId]; ok {
	// 			// 存在返奖
	// 			isLose = true
	// 			// totalinfo.WinNumber++
	// 			// totalinfo.Win += reward.Amount
	// 			totalWin = append(totalWin, reward.Amount)
	// 			// info.EndTime = reward.Ctime
	// 			// info.FWin = fmt.Sprintf("%.2f", Chip2Float(reward.Amount))
	// 			// info.FScore = fmt.Sprintf("%.2f", Chip2Float(reward.Amount))
	// 		}
	// 	}
	// 	if isLose {
	// 		// totalinfo.LoseNumber++
	// 		// totalinfo.Lose += item.Amount
	// 		totalLose = append(totalLose, item.Amount)
	// 	}
	// }
	// for _, item := range total_list {
	// 	totalinfo.Number++
	// 	switch item.Gtype {
	// 	case 1:
	// 		for _, v := range item.TPDetail {
	// 			if len(v.UserId) > 16 {
	// 				continue
	// 			}
	// 			score := v.Score + v.CashMingTax + v.CashAnTax
	// 			if v.Score > 0 {
	// 				// 赢
	// 				totalinfo.WinNumber++
	// 				totalinfo.Win += score
	// 				totalWin = append(totalWin, score)
	// 			} else {
	// 				// 输
	// 				totalinfo.LoseNumber++
	// 				totalinfo.Lose += score
	// 				totalLose = append(totalLose, score)
	// 			}
	// 		}
	// 	case 2:
	// 		for _, v := range item.LHDetail.UserDetail {
	// 			if len(v.Userid) > 16 {
	// 				continue
	// 			}
	// 			score := v.Win + v.CashMingTax + v.CashAnTax
	// 			if v.Win > 0 {
	// 				// 赢
	// 				totalinfo.WinNumber++
	// 				totalinfo.Win += score
	// 				totalWin = append(totalWin, score)
	// 			} else {
	// 				// 输
	// 				totalinfo.LoseNumber++
	// 				totalinfo.Lose += score
	// 				totalLose = append(totalLose, score)
	// 			}
	// 		}
	// 	case 3:
	// 		for _, v := range item.UPDetail.UserDetail {
	// 			if len(v.Userid) > 16 {
	// 				continue
	// 			}
	// 			score := v.Win + v.CashMingTax + v.CashAnTax
	// 			if v.Win > 0 {
	// 				// 赢
	// 				totalinfo.WinNumber++
	// 				totalinfo.Win += score
	// 				totalWin = append(totalWin, score)
	// 			} else {
	// 				// 输
	// 				totalinfo.LoseNumber++
	// 				totalinfo.Lose += score
	// 				totalLose = append(totalLose, score)
	// 			}
	// 		}
	// 	case 4:
	// 		for _, v := range item.RMDetail {
	// 			if len(v.UserId) > 16 {
	// 				continue
	// 			}
	// 			score := v.Score + v.CashMingTax + v.CashAnTax
	// 			if v.Score > 0 {
	// 				// 赢
	// 				totalinfo.WinNumber++
	// 				totalinfo.Win += score
	// 				totalWin = append(totalWin, score)
	// 			} else {
	// 				// 输
	// 				totalinfo.LoseNumber++
	// 				totalinfo.Lose += score
	// 				totalLose = append(totalLose, score)
	// 			}
	// 		}
	// 	case 5:
	// 		for _, v := range item.AK47Detail {
	// 			if len(v.UserId) > 16 {
	// 				continue
	// 			}
	// 			score := v.Score + v.CashMingTax + v.CashAnTax
	// 			if v.Score > 0 {
	// 				// 赢
	// 				totalinfo.WinNumber++
	// 				totalinfo.Win += score
	// 				totalWin = append(totalWin, score)
	// 			} else {
	// 				// 输
	// 				totalinfo.LoseNumber++
	// 				totalinfo.Lose += score
	// 				totalLose = append(totalLose, score)
	// 			}
	// 		}
	// 	case 6:
	// 		for _, v := range item.JOKERDetail {
	// 			if len(v.UserId) > 16 {
	// 				continue
	// 			}
	// 			score := v.Score + v.CashMingTax + v.CashAnTax
	// 			if v.Score > 0 {
	// 				// 赢
	// 				totalinfo.WinNumber++
	// 				totalinfo.Win += score
	// 				totalWin = append(totalWin, score)
	// 			} else {
	// 				// 输
	// 				totalinfo.LoseNumber++
	// 				totalinfo.Lose += score
	// 				totalLose = append(totalLose, score)
	// 			}
	// 		}
	// 	case 7:
	// 		for _, v := range item.CRASHDetail.UserDetail {
	// 			if len(v.Userid) > 16 {
	// 				continue
	// 			}
	// 			score := v.Win + v.CashMingTax + v.CashAnTax
	// 			if v.Win > 0 {
	// 				// 赢
	// 				totalinfo.WinNumber++
	// 				totalinfo.Win += score
	// 				totalWin = append(totalWin, score)
	// 			} else {
	// 				// 输
	// 				totalinfo.LoseNumber++
	// 				totalinfo.Lose += score
	// 				totalLose = append(totalLose, score)
	// 			}
	// 		}
	// 	case 8:
	// 		for _, v := range item.ABDetail.UserDetail {
	// 			if len(v.Userid) > 16 {
	// 				continue
	// 			}
	// 			score := v.Win + v.CashMingTax + v.CashAnTax
	// 			if v.Win > 0 {
	// 				// 赢
	// 				totalinfo.WinNumber++
	// 				totalinfo.Win += score
	// 				totalWin = append(totalWin, score)
	// 			} else {
	// 				// 输
	// 				totalinfo.LoseNumber++
	// 				totalinfo.Lose += score
	// 				totalLose = append(totalLose, score)
	// 			}
	// 		}
	// 	case 9:
	// 		for _, v := range item.CPDetail.UserDetail {
	// 			if len(v.Userid) > 16 {
	// 				continue
	// 			}
	// 			score := v.Win + v.CashMingTax + v.CashAnTax
	// 			if v.Win > 0 {
	// 				// 赢
	// 				totalinfo.WinNumber++
	// 				totalinfo.Win += score
	// 				totalWin = append(totalWin, score)
	// 			} else {
	// 				// 输
	// 				totalinfo.LoseNumber++
	// 				totalinfo.Lose += score
	// 				totalLose = append(totalLose, score)
	// 			}
	// 		}
	// 	case 10:
	// 		for _, v := range item.CRASHDetail.UserDetail {
	// 			if len(v.Userid) > 16 {
	// 				continue
	// 			}
	// 			score := v.Win + v.CashMingTax + v.CashAnTax
	// 			if v.Win > 0 {
	// 				// 赢
	// 				totalinfo.WinNumber++
	// 				totalinfo.Win += score
	// 				totalWin = append(totalWin, score)
	// 			} else {
	// 				// 输
	// 				totalinfo.LoseNumber++
	// 				totalinfo.Lose += score
	// 				totalLose = append(totalLose, score)
	// 			}
	// 		}
	// 	}
	// }
	// if totalinfo.Lose != 0 {
	// 	totalinfo.LoseAvg = ComputeFloat(totalinfo.Lose, 100) / float64(totalinfo.LoseNumber)
	// 	revenue := totalinfo.Win + totalinfo.Lose
	// 	totalinfo.Revenue = ComputeFloat(revenue, 100)
	// 	medianLose := median(totalLose)
	// 	modeLose := mode(totalLose)
	// 	rs := int64(medianLose)
	// 	totalinfo.LoseMedian = ComputeFloat(rs, 100)
	// 	if len(modeLose) > 0 {
	// 		for k, l := range modeLose {
	// 			if k == 5 {
	// 				break
	// 			}
	// 			score := ComputeFloat(l, 100)
	// 			if totalinfo.LoseMode != "" {
	// 				totalinfo.LoseMode += ","
	// 			}
	// 			totalinfo.LoseMode += fmt.Sprintf("%s%.2f", "", score)
	// 		}
	// 	}
	// }
	// if totalinfo.Win != 0 {
	// 	totalinfo.WinAvg = ComputeFloat(totalinfo.Win, 100) / float64(totalinfo.WinNumber)
	// 	medianWin := median(totalWin)
	// 	modeWin := mode(totalWin)
	// 	totalinfo.WinMedian = ComputeFloat(int64(medianWin), 100)
	// 	if len(modeWin) > 0 {
	// 		for k, l := range modeWin {
	// 			if k == 5 {
	// 				break
	// 			}
	// 			score := ComputeFloat(l, 100)
	// 			if totalinfo.LoseMode != "" {
	// 				totalinfo.WinMode += ","
	// 			}
	// 			totalinfo.WinMode += fmt.Sprintf("%s%.2f", "", score)
	// 		}
	// 	}
	// }
	// 分页查询
	query := bson.M{}
	query["round_id"] = bson.M{"$in": ids}
	query["amount"] = bson.M{"$ne": 0}
	NsqLogExternalBets.Find(query).Sort("-ctime").All(&nsq_list)
	// 外接返奖
	var rewardlist []entity.NsqReward
	NsqLogExternalRewards.Find(query).All(&rewardlist)
	nsqRewardTable := make(map[string]map[string]entity.NsqReward)
	for _, reword := range rewardlist {
		uidReward, ok := nsqRewardTable[reword.RoundId]
		if !ok {
			uidReward = make(map[string]entity.NsqReward)
			nsqRewardTable[reword.RoundId] = uidReward
		}
		uidReward[reword.UserId] = reword
	}

	query = bson.M{}
	query["_id"] = bson.M{"$in": ids}
	query["players"] = bson.M{"$ne": ""}
	Details.Find(query).Sort("-begin_time").All(&d_list)
	list = append(list, d_list...)

	for _, item := range nsq_list {
		info := new(entity.Detail)
		info.WaterId = item.RoundId
		info.BeginTime = item.Ctime
		info.EndTime = item.Ctime
		info.Gtype = item.GameId
		info.RoomId = strconv.FormatInt(int64(item.GameId), 10)
		info.DeskId = item.MerchantOrderNo
		info.Players = item.UserId
		info.FBet = fmt.Sprintf("%.2f", Chip2Float(item.Amount))
		info.FPlayerIds = item.UserId

		// 外接返奖信息
		if uidReward, ok := nsqRewardTable[item.RoundId]; ok {
			if reward, ok := uidReward[item.UserId]; ok {
				info.EndTime = reward.Ctime
				info.FWin = fmt.Sprintf("%.2f", Chip2Float(reward.Amount))
				info.FScore = fmt.Sprintf("%.2f", Chip2Float(reward.Amount))
			}
		}
		list = append(list, *info)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].BeginTime > list[j].BeginTime
	})
	return list, totalinfo, count, err
}

// 游戏类型是否外接游戏
func IsExternalGame(gtype int32) bool {
	return gtype > 100
}

func (this *playerService) GetDetailListCk(page, pageSize int, params map[string]any) (details []*entity.DetailList, stats *entity.TotalDetail, total int64, err error) {
	stats = &entity.TotalDetail{}

	sql1 := `
		SELECT %s FROM (
			SELECT id,userid,begin_time,end_time,gtype,bet_amount,
				settle_score,before_score,after_score,cash_ming_tax,bonus_ming_tax,cash_an_tax,bonus_an_tax,win_type,
				room_id,players,rm_game_over_reason,change_card_type,control_type,player_factor,
				rmc_ok,rmc_ctype,rmc_roi_id,rmc_drop_id,rmc_control_effect,rmc_hierarchy,rmc_draw_num_player,rmc_draw_num_robot,rmc_not_draw_robot1st,rmc_not_draw_player1st,rmc_control_robot_draw,rmc_control_robot_draw_round,
				is_charge,is_strategy,lhd_strategy_id,up_strategy_id,crash_strategy_type,ab_strategy_id,cp_strategy_id,rb_strategy_id,player_stage_id,tp_model,is_tp_model3story_plus,
				tp_prxd_active,tp_prxd_high_card_rounds,tp_ljsb_active,tp_ljsb_jl_type,tp_ljsb_py,tp_ljsb_ps,tp_ljsb_pd,tp_ljsb_pt,tp_gcyx_active,tp_gcyx_rp,tp_gcyx_bp,
				mines_mines,mines_step,mines_multiple,mines_auto_mines,mines_force
			FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) %s
				UNION ALL
			SELECT round_id id,user_id userid,MIN(ctime) begin_time,MAX(ctime) end_time,game_id gtype,SUM(amount) bet_amount,
				0 settle_score,0 before_score,0 after_score,0 cash_ming_tax,0 bonus_ming_tax,0 cash_an_tax,0 bonus_an_tax,0 win_type,
				'' room_id,user_id players,0 rm_game_over_reason,0 change_card_type,0 control_type,0 player_factor,
				0 rmc_ok,0 rmc_ctype,'' rmc_roi_id,'' rmc_drop_id,0 rmc_control_effect,0 rmc_hierarchy,0 rmc_draw_num_player,0 rmc_draw_num_robot,0 rmc_not_draw_robot1st,0 rmc_not_draw_player1st,0 rmc_control_robot_draw,0 rmc_control_robot_draw_round,
				0 is_charge,0 is_strategy,0 lhd_strategy_id,0 up_strategy_id,[] crash_strategy_type,0 ab_strategy_id,0 cp_strategy_id,0 rb_strategy_id,0 player_stage_id,0 tp_model,0 is_tp_model3story_plus,
				0 tp_prxd_active,0 tp_prxd_high_card_rounds,0 tp_ljsb_active,0 tp_ljsb_jl_type,0 tp_ljsb_py,0 tp_ljsb_ps,0 tp_ljsb_pd,0 tp_ljsb_pt,0 tp_gcyx_active,0 tp_gcyx_rp,0 tp_gcyx_bp,
				0 mines_mines,0 mines_step,0 mines_multiple,0 mines_auto_mines,0 mines_force
			FROM game.col_nsq_log_external_bet t2 FINAL WHERE amount != 0 %s GROUP BY round_id, user_id, game_id
		) t1 %s
	`
	// 拼接条件
	where1, where2 := "", ""
	var args1, args2 []any
	// 查询
	if user_id, ok := params["user_id"]; ok {
		where1 += " and t1.userid in ?"
		args1 = append(args1, user_id)
		where2 += " and t2.user_id in ?"
		args2 = append(args2, user_id)
	}
	if round_id, ok := params["round_id"]; ok {
		where1 += " and t1.id = ?"
		args1 = append(args1, round_id)
		where2 += " and t2.round_id = ?"
		args2 = append(args2, round_id)
	}
	if desk_id, ok := params["desk_id"]; ok {
		where1 += " and t1.desk_id = ?"
		args1 = append(args1, desk_id)
		where2 += " and 1 = 0" // 外接没有桌子ID
	}
	// 游戏
	if gtype, ok := params["gtype"]; ok {
		where1 += " and t1.gtype = ?"
		args1 = append(args1, gtype)
		where2 += " and t2.game_id = ?"
		args2 = append(args2, gtype)
	}
	if room_id, ok := params["room_id"]; ok {
		where1 += " and t1.room_id = ?"
		args1 = append(args1, room_id)
		where2 += " and 1 = 0" // 外接没有房间
	}

	// 时间
	startTime, startOK := params["startTime"]
	endTime, endOK := params["endTime"]
	if startOK && endOK {
		where1 += " and t1.begin_time between ? and ?"
		args1 = append(args1, startTime, endTime)
		where2 += " and t2.ctime between ? and ?"
		args2 = append(args2, startTime, endTime)
	} else if startOK {
		where1 += " and t1.begin_time >= ?"
		args1 = append(args1, startTime)
		where2 += " and t2.ctime >= ?"
		args2 = append(args2, startTime)
	} else if endOK {
		where1 += " and t1.begin_time <= ?"
		args1 = append(args1, endTime)
		where2 += " and t2.ctime <= ?"
		args2 = append(args2, endTime)
	}

	// 查总数
	sql_count := fmt.Sprintf(sql1, "count(*) c", where1, where2, "")
	args_count := append(args1, args2...)
	err = ck.Select(&total, sql_count, args_count...)
	if err != nil || total == 0 {
		return
	}

	selects := "*"
	order := " ORDER BY begin_time DESC LIMIT ?, ?"
	sql_list := fmt.Sprintf(sql1, selects, where1, where2, order)
	args_list := append(args1, args2...)
	offset, limit := PageCalc(page, pageSize)
	args_list = append(args_list, offset, limit)
	err = ck.Select(&details, sql_list, args_list...)
	if err != nil || len(details) == 0 {
		return
	}

	// 外接返奖
	var waterIds []string
	for _, dtl := range details {
		if IsExternalGame(dtl.Gtype) {
			// dtl.BetAmount = dtl.
			waterIds = append(waterIds, dtl.WaterId)
		}
	}
	if len(waterIds) > 0 {
		var rewards []*ck.NsqLogExternalReward
		// rewardQ := ck.DB().Model(&ck.NsqLogExternalReward{}).Table("col_nsq_log_external_reward final")
		// rewardQ.Where("round_id in ?", waterIds)
		// rewardQ.Where("amount <> 0")
		// err = rewardQ.Find(&rewards).Error
		err = ck.Select(&rewards, `
			select t3.round_id, t3.user_id, sum(t3.amount) amount, max(t3.ctime) ctime
			from col_nsq_log_external_reward t3 final 
			where t3.round_id in ? and t3.amount != 0 group by t3.round_id, t3.user_id
		`, waterIds)
		if err != nil {
			return
		}
		if len(rewards) > 0 {
			externalRewardTable := make(map[string]map[string]*ck.NsqLogExternalReward)
			for _, reward := range rewards {
				uidReward, ok := externalRewardTable[reward.RoundId]
				if !ok {
					uidReward = make(map[string]*ck.NsqLogExternalReward)
					externalRewardTable[reward.RoundId] = uidReward
				}
				uidReward[reward.UserId] = reward
			}
			for _, dtl := range details {
				if IsExternalGame(dtl.Gtype) {
					dtl.SettleScore = 0 - dtl.BetAmount
					// 外接返奖信息
					if uidReward, ok := externalRewardTable[dtl.WaterId]; ok {
						if reward, ok := uidReward[dtl.UserId]; ok {
							dtl.EndTime = reward.Ctime
							dtl.SettleScore = reward.RewardAmount - dtl.BetAmount
							dtl.FWin = fmt.Sprintf("%.2f", Chip2Float(reward.RewardAmount))
							dtl.FScore = fmt.Sprintf("%.2f", Chip2Float(dtl.SettleScore))
						}
					}
					dtl.WinType = 2
					if dtl.SettleScore > 0 {
						dtl.WinType = 1
					} else if dtl.SettleScore == 0 {
						dtl.WinType = 3
					}
				}
			}
		}
	}

	// 条件对局 玩家胜局:0 玩家负局:0
	// statsQ := `
	// 	SELECT win_type, count(*) total, SUM(settle_score) score_sum, AVG(settle_score) score_avg, median(settle_score) score_median, topK(1)(settle_score) score_mode
	// 	FROM (
	// 		SELECT settle_score, win_type
	// 		FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) %s
	// 		UNION ALL

	// 		SELECT SUM(s3.amount) - SUM(s2.amount) settle_score, (CASE WHEN settle_score > 0 THEN 1 WHEN settle_score == 0 THEN 3 ELSE 2 END) win_type
	// 		FROM (SELECT t2.round_id, t2.user_id, SUM(t2.amount) amount FROM game.col_nsq_log_external_bet t2 FINAL WHERE t2.amount != 0 %s GROUP BY round_id, user_id) s2
	// 		LEFT JOIN (SELECT t2.round_id, t2.user_id, SUM(t2.amount) amount FROM game.col_nsq_log_external_reward t2 FINAL WHERE t2.amount != 0 %s GROUP BY round_id, user_id) s3
	// 			ON s2.round_id = s3.round_id AND s2.user_id = s3.user_id
	// 		GROUP BY s2.round_id, s2.user_id
	// 	) s1 GROUP BY win_type
	// `
	statsQ := `
		SELECT win_type, count(*) total, SUM(score) score_sum, AVG(score) score_avg, median(score) score_median, topK(1)(score) score_mode
		FROM (
			SELECT score, win_type
			FROM game.col_detail t1 FINAL WHERE robot = 0 and win_type in (1,2,3) %s
			UNION ALL

			SELECT SUM(s3.amount) - SUM(s2.amount) score, (CASE WHEN score > 0 THEN 1 WHEN score == 0 THEN 3 ELSE 2 END) win_type
			FROM (SELECT t2.round_id, t2.user_id, SUM(t2.amount) amount FROM game.col_nsq_log_external_bet t2 FINAL WHERE t2.amount != 0 %s GROUP BY round_id, user_id) s2
			LEFT JOIN (SELECT t2.round_id, t2.user_id, SUM(t2.amount) amount FROM game.col_nsq_log_external_reward t2 FINAL WHERE t2.amount != 0 %s GROUP BY round_id, user_id) s3
				ON s2.round_id = s3.round_id AND s2.user_id = s3.user_id
			GROUP BY s2.round_id, s2.user_id
		) s1 GROUP BY win_type
	`
	statsQ = fmt.Sprintf(statsQ, where1, where2, where2)
	statsArgs := append(args1, args2...)
	statsArgs = append(statsArgs, args2...)
	var statsR []map[string]any
	err = ck.Select(&statsR, statsQ, statsArgs...)
	if err != nil {
		return
	}
	for _, r := range statsR {
		win_type := utils.ToInt64(r["win_type"])
		total := utils.ToInt64(r["total"])
		score_sum := utils.ToInt64(r["score_sum"])
		score_avg := utils.ToFloat64(r["score_avg"])
		score_median := utils.ToFloat64(r["score_median"])
		score_mode, ok := r["score_mode"].([]int64)
		if !ok {
			score_mode = []int64{0}
		}

		stats.Number += total
		switch win_type {
		case 1:
			stats.WinNumber = total
			stats.Win = score_sum
			stats.WinAvg = score_avg / 100
			stats.WinMedian = score_median / 100
			if len(score_mode) > 0 {
				stats.WinMode = fmt.Sprintf("%.2f", ComputeFloat(score_mode[0], 100))
			}
		case 2:
			stats.LoseNumber = total
			stats.Lose = score_sum
			stats.LoseAvg = score_avg / 100
			stats.LoseMedian = score_median / 100
			if len(score_mode) > 0 {
				stats.LoseMode = fmt.Sprintf("%.2f", ComputeFloat(score_mode[0], 100))
			}
		case 3:
			stats.TieNumber = total
		}
	}
	stats.Revenue = ComputeFloat(stats.Win+stats.Lose, 100)
	return
}

// 中位数
func median(data []int64) float64 {
	sort.Slice(data, func(i, j int) bool {
		return data[i] < data[j]
	})

	n := len(data)
	if n == 0 {
		return 0
	}

	mid := n / 2
	if n%2 == 0 {
		return float64(data[mid-1]+data[mid]) / 2.0
	}

	return float64(data[mid])
}

// 众数
func mode(data []int64) []int64 {
	freq := make(map[int64]int)

	for _, v := range data {
		freq[v]++
	}

	var mode []int64
	maxFreq := 0

	for k, v := range freq {
		if v > maxFreq {
			maxFreq = v
			mode = []int64{k}
		} else if v == maxFreq {
			mode = append(mode, k)
		}
	}

	return mode
}

// 获取总数
func (this *playerService) GetDetailTotal(m bson.M) (int64, error) {
	return int64(Count(Details, m)), nil
}

// 根据用户ID查询该玩家游戏局数
func (this *playerService) GetDetailTotalByUser_old(userid string) ([]entity.UserGames, error) {
	// return int64(Count(Details, m)), nil
	var list []entity.UserGames
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"players": bson.M{
					"$regex":   fmt.Sprintf("\\b%s\\b", userid),
					"$options": "i",
				},
			},
		},
		{
			"$group": bson.M{
				"_id": "$gtype",
				"num": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	result := []bson.M{}
	pipe := Details.Pipe(pipeline)
	err := pipe.All(&result)
	// recordList := map[int]string{
	// 	1: "TP",
	// 	2: "DRAGON TIGER",
	// 	3: "7UPDOWN",
	// 	4: "RUMMY",
	// 	5: "AK47",
	// 	6: "JOKER",
	// 	7: "CRASH",
	// }
	for _, item := range result {
		info := new(entity.UserGames)
		id := item["_id"].(int)
		// for k, v := range recordList {
		// 	if id == int32(k) {
		// 		info.Name = v
		// 	}
		// }
		num := item["num"].(int)
		info.Id = int32(id)
		info.Number = int32(num)
		list = append(list, *info)
	}
	return list, err
}

func (this *playerService) GetDetailTotalByUser(userid string) (*entity.UserGameInfo, error) {
	info := new(entity.UserGameInfo)
	info.Id = userid
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"userid": userid,
			},
		},
		{
			"$group": bson.M{
				"_id": "$gtype",
				"num": bson.M{
					"$sum": "$number",
				},
			},
		},
	}
	result := []bson.M{}
	pipe := UserGameDatas.Pipe(pipeline)
	err := pipe.All(&result)
	for _, item := range result {
		id := item["_id"].(int64)
		num := item["num"].(int64)
		switch id {
		case 1:
			info.TPNumber = int32(num)
		case 2:
			info.LHDNumber = int32(num)
		case 3:
			info.UPNumber = int32(num)
		case 4:
			info.RMNumber = int32(num)
		case 5:
			info.AKNumber = int32(num)
		case 6:
			info.JOKERNumber = int32(num)
		case 7:
			info.CRASHNumber = int32(num)
		case 8:
			info.ABNumber = int32(num)
		case 9:
			info.CPNumber = int32(num)
		case 10:
			info.FJNumber = int32(num)
		case 11:
			info.RBNumber = int32(num)
		case 12:
			info.RMTwoNumber = int32(num)
		case 13:
			info.TP2Number = int32(num)
		case 600101:
			info.G_F_Number = int32(num)
		case 600002:
			info.L_N_Number = int32(num)
		case 600073:
			info.G_G_Number = int32(num)
		case 600022:
			info.F_O_Number = int32(num)
		case 600039:
			info.S_W_Number = int32(num)
		case 600120:
			info.F_O_Number = int32(num)
		case 600054:
			info.W_B_S_Number = int32(num)
		case 600104:
			info.R_O_A_Number = int32(num)
		case 600037:
			info.F_T_Number = int32(num)
		case 600028:
			info.D_O_S_M_Number = int32(num)
		case 600041:
			info.L_O_P_Number = int32(num)
		case 600098:
			info.C_W_Number = int32(num)
		case 600025:
			info.S_S_Number = int32(num)
		case 600093:
			info.A_R_Number = int32(num)
		case 600004:
			info.J_K_Number = int32(num)
		case 600108:
			info.W_B_Number = int32(num)
		case 600009:
			info.SUP_S_Number = int32(num)
		case 600012:
			info.C_N_Number = int32(num)
		case 600119:
			info.Gal_G_Number = int32(num)
		case 600110:
			info.W_O_T_Q_Number = int32(num)
		case 600029:
			info.H_T_O_D_C_Number = int32(num)
		case 600081:
			info.D_H_Number = int32(num)
		case 600086:
			info.L_RICH_Number = int32(num)
		case 600099:
			info.E_B_O_M_Number = int32(num)
		case 600102:
			info.D_O_M_Number = int32(num)
		case 600117:
			info.C_B_Number = int32(num)
		case 600103:
			info.Gal_G_Number = int32(num)
		case 600607:
			info.Auto_R_Number = int32(num)
		case 600572:
			info.L_B_Number = int32(num)
		case 600526:
			info.L_R_Number = int32(num)
		case 600513:
			info.S_S_B_Number = int32(num)
		case 600594:
			info.D_T_Number = int32(num)
		case 600583:
			info.D_C_Number = int32(num)
		case 600642:
			info.FAN_T_Number = int32(num)
		case 600536:
			info.G_W_B_Number = int32(num)
		case 600631:
			info.B_B_Number = int32(num)
		}
	}
	return info, err
}

func (this *playerService) GetDetailTotalByUser2(userid string) (*entity.UserGameInfo, error) {
	info := new(entity.UserGameInfo)
	info.Id = userid
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"players": bson.M{
					"$regex":   fmt.Sprintf("\\b%s\\b", userid),
					"$options": "i",
				},
			},
		},
		{
			"$group": bson.M{
				"_id": "$gtype",
				"num": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	result := []bson.M{}
	pipe := Details.Pipe(pipeline)
	err := pipe.All(&result)
	for _, item := range result {
		id := item["_id"].(int)
		num := item["num"].(int)
		switch id {
		case 1:
			info.TPNumber = int32(num)
		case 2:
			info.LHDNumber = int32(num)
		case 3:
			info.UPNumber = int32(num)
		case 4:
			info.RMNumber = int32(num)
		case 5:
			info.AKNumber = int32(num)
		case 6:
			info.JOKERNumber = int32(num)
		case 7:
			info.CRASHNumber = int32(num)
		case 8:
			info.ABNumber = int32(num)
		case 9:
			info.CPNumber = int32(num)
		case 10:
			info.FJNumber = int32(num)
		}
	}

	// 外接游戏局数
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"user_id": userid,
				"amount":  bson.M{"$ne": 0},
			},
		},
		{
			"$group": bson.M{
				"_id":   "$game_id",
				"count": bson.M{"$addToSet": "$round_id"},
			},
		},
	}
	result1 := []bson.M{}
	pipe1 := NsqLogExternalBets.Pipe(pipeline1)
	err = pipe1.All(&result1)
	for _, item := range result1 {
		id := item["_id"].(int)
		arr_num := item["count"].([]interface{})
		map_num := make(map[string]bool, 0)
		for _, v := range arr_num {
			rid := v.(string)
			map_num[rid] = true
		}
		num := len(map_num)
		switch id {
		case 600101:
			info.G_F_Number = int32(num)
		case 600002:
			info.L_N_Number = int32(num)
		case 600073:
			info.G_G_Number = int32(num)
		case 600022:
			info.F_O_Number = int32(num)
		case 600039:
			info.S_W_Number = int32(num)
		case 600120:
			info.F_O_Number = int32(num)
		case 600054:
			info.W_B_S_Number = int32(num)
		case 600104:
			info.R_O_A_Number = int32(num)
		case 600037:
			info.F_T_Number = int32(num)
		case 600028:
			info.D_O_S_M_Number = int32(num)
		case 600041:
			info.L_O_P_Number = int32(num)
		case 600098:
			info.C_W_Number = int32(num)
		case 600025:
			info.S_S_Number = int32(num)
		case 600093:
			info.A_R_Number = int32(num)
		case 600004:
			info.J_K_Number = int32(num)
		case 600108:
			info.W_B_Number = int32(num)
		case 600009:
			info.SUP_S_Number = int32(num)
		case 600012:
			info.C_N_Number = int32(num)
		case 600119:
			info.Gal_G_Number = int32(num)
		case 600110:
			info.W_O_T_Q_Number = int32(num)
		case 600029:
			info.H_T_O_D_C_Number = int32(num)
		case 600081:
			info.D_H_Number = int32(num)
		case 600086:
			info.L_RICH_Number = int32(num)
		case 600099:
			info.E_B_O_M_Number = int32(num)
		case 600102:
			info.D_O_M_Number = int32(num)
		case 600117:
			info.C_B_Number = int32(num)
		case 600103:
			info.Gal_G_Number = int32(num)
		case 600607:
			info.Auto_R_Number = int32(num)
		case 600572:
			info.L_B_Number = int32(num)
		case 600526:
			info.L_R_Number = int32(num)
		case 600513:
			info.S_S_B_Number = int32(num)
		case 600594:
			info.D_T_Number = int32(num)
		case 600583:
			info.D_C_Number = int32(num)
		case 600642:
			info.FAN_T_Number = int32(num)
		case 600536:
			info.G_W_B_Number = int32(num)
		case 600631:
			info.B_B_Number = int32(num)
		}
	}
	return info, err
}

// 根据ID获取对局信息
func (this *playerService) GetDetail(id string) (*entity.Detail, error) {
	details := new(entity.Detail)
	Get(Details, id, details)
	if details.WaterId == "" {
		return details, errors.New("对局不存在")
	}
	details = this.chipList5(details)
	return details, nil
}

// 根据句号返回外接对局详情
func (this *playerService) GetExternalDetail(id string) (*entity.ExternalDetail, error) {
	details := new(entity.ExternalDetail)
	// Get(NsqLogExternalBets, id, details)
	m := bson.M{}
	m["round_id"] = id
	NsqLogExternalBets.Find(m).One(details)
	if details.RoundId == "" {
		return details, errors.New("对局不存在")
	}
	var rewardlist []entity.NsqReward
	m["user_id"] = details.UserId
	NsqLogExternalRewards.Find(m).All(&rewardlist)
	details.RewardDetail = rewardlist
	details = this.chipList20(details)
	return details, nil
}

func (this *playerService) chipList5(info *entity.Detail) *entity.Detail {
	switch info.ChangeCardType {
	case 1:
		info.ChangeCardTypeName = "开局换牌"
	case 2:
		info.ChangeCardTypeName = "局中换牌"
	default:
		info.ChangeCardTypeName = "没有"
	}
	switch info.ControlType {
	case 1:
		info.ControlTypeName = "当前赢分"
	case 2:
		info.ControlTypeName = "房间系数"
	default:
		info.ControlTypeName = "无"
	}
	info.FWinScore = Chip2Float(int64(info.WinScore))
	info.FWinScore1 = Chip2Float(int64(info.WinScore1))
	info.FWinScore2 = Chip2Float(int64(info.WinScore2))
	info.FChargeMoney = Chip2Float(int64(info.ChargeMoney))
	switch info.Gtype {
	case 1, 13:
		// TP
		for j, e := range info.TPDetail {
			e.FScore = Chip2Float(int64(e.Score))
			e.FBeforeScore = Chip2Float(int64(e.BeforeScore))
			e.FAfterScore = Chip2Float(int64(e.AfterScore))
			e.FBeforeCash = Chip2Float(int64(e.BeforeCash))
			e.FAfterCash = Chip2Float(int64(e.AfterCash))
			e.FBeforeBonus = Chip2Float(int64(e.BeforeBonus))
			e.FAfterBonus = Chip2Float(int64(e.AfterBonus))
			e.FBet = Chip2Float(int64(e.Bet))
			e.FBottom = Chip2Float(int64(e.Bottom))
			e.FCashStock = Chip2Float(int64(e.CashStock))
			e.FBonusStock = Chip2Float(int64(e.BonusStock))
			e.FCashMingTax = Chip2Float(int64(e.CashMingTax))
			e.FBonusMingTax = Chip2Float(int64(e.BonusMingTax))
			e.FCashAnTax = Chip2Float(int64(e.CashAnTax))
			e.FBonusAnTax = Chip2Float(int64(e.BonusAnTax))
			if e.NewBieProbeId > 0 {
				info.NewBieProbeId = e.NewBieProbeId
				info.NewBieProbeRound = e.NewBieProbeRound
			}

			info.TPDetail[j] = e
		}
	case 2:
		// 龙虎斗
		info.FPlayerFactor = Chip2Float(int64(info.PlayerFactor))
		if info.LHDetail != nil {
			info.LHDetail.FBets = Chip2Float(info.LHDetail.Bets)
			info.LHDetail.FPlayerWin = Chip2Float(info.LHDetail.PlayerWin)
			for j, e := range info.LHDetail.UserDetail {
				e.FDragon = Chip2Float(e.Dragon)
				e.FTiger = Chip2Float(e.Tiger)
				e.FTie = Chip2Float(e.Tie)
				e.FWin = Chip2Float(e.Win)
				e.FBeforeScore = Chip2Float(e.BeforeScore)
				e.FAfterScore = Chip2Float(e.AfterScore)
				e.FBeforeCash = Chip2Float(e.BeforeCash)
				e.FAfterCash = Chip2Float(e.AfterCash)
				e.FBeforeBonus = Chip2Float(e.BeforeBonus)
				e.FAfterBonus = Chip2Float(e.AfterBonus)
				e.FCashStock = Chip2Float(e.CashStock)
				e.FBonusStock = Chip2Float(e.BonusStock)
				e.FCashMingTax = Chip2Float(e.CashMingTax)
				e.FBonusMingTax = Chip2Float(e.BonusMingTax)
				e.FCashAnTax = Chip2Float(e.CashAnTax)
				e.FBonusAnTax = Chip2Float(e.BonusAnTax)
				info.LHDetail.UserDetail[j] = e
			}
		}
	case 3:
		// 7updown
		info.FPlayerFactor = Chip2Float(int64(info.PlayerFactor))
		if info.UPDetail != nil {
			info.UPDetail.FBets = Chip2Float(info.UPDetail.Bets)
			info.UPDetail.FPlayerWin = Chip2Float(info.UPDetail.PlayerWin)

			// 暴击位置
			if info.UPDetail.Crit7up && len(info.UPDetail.CritOdds) > 0 {
				for seat, odd := range info.UPDetail.CritOdds {
					fSeat := fmt.Sprint(seat)
					switch seat {
					case 0:
						fSeat = "小"
					case 1:
						fSeat = "大"
					}
					info.UPDetail.FCritOdds = append(info.UPDetail.FCritOdds, entity.FLHCritOdd{
						Seat:     seat,
						FSeat:    fSeat,
						Multiple: fmt.Sprintf("x%d", odd),
					})
				}
				sort.Slice(info.UPDetail.FCritOdds, func(i, j int) bool {
					return info.UPDetail.FCritOdds[i].Seat < info.UPDetail.FCritOdds[j].Seat
				})
			}
			// 默认/暴击返奖倍数
			odds := map[uint32]int32{0: 2, 1: 2, 2: 27, 3: 13, 4: 9, 5: 7, 6: 6, 7: 5, 8: 6, 9: 7, 10: 9, 11: 13, 12: 27}
			critOdds := info.UPDetail.CritOdds
			if critOdds == nil {
				critOdds = make(map[uint32]int32)
			}
			for j, e := range info.UPDetail.UserDetail {
				e.FDragon = Chip2Float(e.Dragon)
				e.FTiger = Chip2Float(e.Tiger)
				e.FTie = Chip2Float(e.Tie)
				e.FWin = Chip2Float(e.Win)
				e.FBeforeScore = Chip2Float(e.BeforeScore)
				e.FAfterScore = Chip2Float(e.AfterScore)
				e.FBeforeCash = Chip2Float(e.BeforeCash)
				e.FAfterCash = Chip2Float(e.AfterCash)
				e.FBeforeBonus = Chip2Float(e.BeforeBonus)
				e.FAfterBonus = Chip2Float(e.AfterBonus)
				e.FCashStock = Chip2Float(e.CashStock)
				e.FBonusStock = Chip2Float(e.BonusStock)
				e.FCashMingTax = Chip2Float(e.CashMingTax)
				e.FBonusMingTax = Chip2Float(e.BonusMingTax)
				e.FCashAnTax = Chip2Float(e.CashAnTax)
				e.FBonusAnTax = Chip2Float(e.BonusAnTax)
				e.Crit7up = info.UPDetail.Crit7up
				if info.UPDetail.Crit7up && len(e.SeatBets) > 0 {
					for seat, bets := range e.SeatBets {
						fSeat := fmt.Sprint(seat)
						switch seat {
						case 0:
							fSeat = "小"
							e.FDragon = Chip2Float(bets)
						case 1:
							fSeat = "大"
							e.FTiger = Chip2Float(bets)
						case 7:
							e.FTie = Chip2Float(bets)
						}
						var crit bool
						multiple := odds[seat]
						if m, ok := critOdds[seat]; ok {
							crit = true
							multiple = m
						}
						hit := seat == uint32(info.UPDetail.Winner) || seat == uint32(info.UPDetail.PointValue)
						e.FSeatBets = append(e.FSeatBets, entity.FLHSeatBets{
							Seat:     seat,
							FSeat:    fSeat,
							Bets:     fmt.Sprintf("%.2f", Chip2Float(bets)),
							Hit:      hit,
							Crit:     hit && crit,
							Multiple: fmt.Sprintf("x%d", multiple),
						})
					}
					sort.Slice(e.FSeatBets, func(i, j int) bool {
						return e.FSeatBets[i].Seat < e.FSeatBets[j].Seat
					})
				}
				info.UPDetail.UserDetail[j] = e
			}
		}
	case 4:
		// Rummy
		if info.RMDetail != nil {
			for j, e := range info.RMDetail {
				e.FScore = Chip2Float(int64(e.Score))
				e.FBeforeScore = Chip2Float(int64(e.BeforeScore))
				e.FAfterScore = Chip2Float(int64(e.AfterScore))
				e.FBeforeCash = Chip2Float(int64(e.BeforeCash))
				e.FAfterCash = Chip2Float(int64(e.AfterCash))
				e.FBeforeBonus = Chip2Float(int64(e.BeforeBonus))
				e.FAfterBonus = Chip2Float(int64(e.AfterBonus))
				e.FCashStock = Chip2Float(int64(e.CashStock))
				e.FBonusStock = Chip2Float(int64(e.BonusStock))
				e.FCashMingTax = Chip2Float(int64(e.CashMingTax))
				e.FBonusMingTax = Chip2Float(int64(e.BonusMingTax))
				e.FCashAnTax = Chip2Float(int64(e.CashAnTax))
				e.FBonusAnTax = Chip2Float(int64(e.BonusAnTax))
				info.RMDetail[j] = e
			}
		}
	case 5:
		// AK47
		if info.AK47Detail != nil {
			for j, e := range info.AK47Detail {
				e.FScore = Chip2Float(int64(e.Score))
				e.FBeforeScore = Chip2Float(int64(e.BeforeScore))
				e.FAfterScore = Chip2Float(int64(e.AfterScore))
				e.FBeforeCash = Chip2Float(int64(e.BeforeCash))
				e.FAfterCash = Chip2Float(int64(e.AfterCash))
				e.FBeforeBonus = Chip2Float(int64(e.BeforeBonus))
				e.FAfterBonus = Chip2Float(int64(e.AfterBonus))
				e.FBet = Chip2Float(int64(e.Bet))
				e.FBottom = Chip2Float(int64(e.Bottom))
				e.FCashStock = Chip2Float(int64(e.CashStock))
				e.FBonusStock = Chip2Float(int64(e.BonusStock))
				e.FCashMingTax = Chip2Float(int64(e.CashMingTax))
				e.FBonusMingTax = Chip2Float(int64(e.BonusMingTax))
				e.FCashAnTax = Chip2Float(int64(e.CashAnTax))
				e.FBonusAnTax = Chip2Float(int64(e.BonusAnTax))
				info.AK47Detail[j] = e
			}
		}
	case 6:
		// Joker
		if info.JOKERDetail != nil {
			for j, e := range info.JOKERDetail {
				e.FScore = Chip2Float(int64(e.Score))
				e.FBeforeScore = Chip2Float(int64(e.BeforeScore))
				e.FAfterScore = Chip2Float(int64(e.AfterScore))
				e.FBeforeCash = Chip2Float(int64(e.BeforeCash))
				e.FAfterCash = Chip2Float(int64(e.AfterCash))
				e.FBeforeBonus = Chip2Float(int64(e.BeforeBonus))
				e.FAfterBonus = Chip2Float(int64(e.AfterBonus))
				e.FBet = Chip2Float(int64(e.Bet))
				e.FBottom = Chip2Float(int64(e.Bottom))
				e.FCashStock = Chip2Float(int64(e.CashStock))
				e.FBonusStock = Chip2Float(int64(e.BonusStock))
				e.FCashMingTax = Chip2Float(int64(e.CashMingTax))
				e.FBonusMingTax = Chip2Float(int64(e.BonusMingTax))
				e.FCashAnTax = Chip2Float(int64(e.CashAnTax))
				e.FBonusAnTax = Chip2Float(int64(e.BonusAnTax))
				info.JOKERDetail[j] = e
			}
		}

	case 7, 10:
		// CRASH
		info.FPlayerFactor = Chip2Float(int64(info.PlayerFactor))
		if info.CRASHDetail != nil {
			info.CRASHDetail.FBets = Chip2Float(info.CRASHDetail.Bets)
			info.CRASHDetail.FPlayerWin = Chip2Float(info.CRASHDetail.PlayerLose)
			isStrategy := false
			if len(info.CRASHDetail.StrategyType) > 0 {
				isStrategy = true
			}
			info.CRASHDetail.IsStrategy = isStrategy
			for j, e := range info.CRASHDetail.UserDetail {
				e.FBet = Chip2Float(e.Bet)
				e.FBet0 = Chip2Float(e.Bet0)
				e.FBet1 = Chip2Float(e.Bet1)
				e.FWin = Chip2Float(e.Win)
				e.FWin0 = Chip2Float(e.Win0)
				e.FWin1 = Chip2Float(e.Win1)
				e.FBeforeScore = Chip2Float(e.BeforeScore)
				e.FAfterScore = Chip2Float(e.AfterScore)
				e.FBeforeCash = Chip2Float(e.BeforeCash)
				e.FAfterCash = Chip2Float(e.AfterCash)
				e.FBeforeBonus = Chip2Float(e.BeforeBonus)
				e.FAfterBonus = Chip2Float(e.AfterBonus)
				e.FCashMingTax = Chip2Float(e.CashMingTax)
				e.FBonusMingTax = Chip2Float(e.BonusMingTax)
				e.FCashAnTax = Chip2Float(e.CashAnTax)
				e.FBonusAnTax = Chip2Float(e.BonusAnTax)
				info.CRASHDetail.UserDetail[j] = e
			}
			users := info.CRASHDetail.UserDetail
			sort.Slice(users, func(i, j int) bool {
				r1, r2 := utils.CaseElse(users[i].Robot, 1, 0), utils.CaseElse(users[j].Robot, 1, 0)
				if r1 < r2 {
					return true
				} else if r1 > r2 {
					return false
				}
				return users[i].Bet > users[j].Bet
			})
			info.CRASHDetail.UserDetail = users
		}
	case 8:
		// AB
		info.FPlayerFactor = Chip2Float(int64(info.PlayerFactor))
		if info.ABDetail != nil {
			info.ABDetail.FBets = Chip2Float(info.ABDetail.Bets)
			info.ABDetail.FPlayerWin = Chip2Float(info.ABDetail.PlayerWin)
			info.ABDetail.FCanWinScore = Chip2Float(info.ABDetail.CanWinScore)
			if info.ABDetail.Winner == 1 {
				info.ABDetail.WinnerStr = "ANDAR"
			} else if info.ABDetail.Winner == 2 {
				info.ABDetail.WinnerStr = "BAHAR"
			}
			if info.ABDetail.SideWinner == 3 {
				info.ABDetail.SideWinnerStr = "1-5"
			} else if info.ABDetail.SideWinner == 4 {
				info.ABDetail.SideWinnerStr = "6-10"
			} else if info.ABDetail.SideWinner == 5 {
				info.ABDetail.SideWinnerStr = "11-15"
			} else if info.ABDetail.SideWinner == 6 {
				info.ABDetail.SideWinnerStr = "16-25"
			} else if info.ABDetail.SideWinner == 7 {
				info.ABDetail.SideWinnerStr = "26-30"
			} else if info.ABDetail.SideWinner == 8 {
				info.ABDetail.SideWinnerStr = "31-35"
			} else if info.ABDetail.SideWinner == 9 {
				info.ABDetail.SideWinnerStr = "36-40"
			} else if info.ABDetail.SideWinner == 10 {
				info.ABDetail.SideWinnerStr = "41以上"
			}
			for j, e := range info.ABDetail.UserDetail {
				e.FWin = Chip2Float(e.Win)
				e.FBeforeScore = Chip2Float(e.BeforeScore)
				e.FAfterScore = Chip2Float(e.AfterScore)
				e.FBeforeCash = Chip2Float(e.BeforeCash)
				e.FAfterCash = Chip2Float(e.AfterCash)
				e.FBeforeBonus = Chip2Float(e.BeforeBonus)
				e.FAfterBonus = Chip2Float(e.AfterBonus)
				e.FCashMingTax = Chip2Float(e.CashMingTax)
				e.FBonusMingTax = Chip2Float(e.BonusMingTax)
				e.FCashAnTax = Chip2Float(e.CashAnTax)
				e.FBonusAnTax = Chip2Float(e.BonusAnTax)
				betInfo := new(entity.ABSeatBets)
				if e.SeatBets != nil {
					for i, item := range e.SeatBets {
						fmt.Print(i)
						switch i {
						case "1":
							betInfo.Seat1 = Chip2Float(item)
						case "2":
							betInfo.Seat2 = Chip2Float(item)
						case "3":
							betInfo.Seat3 = Chip2Float(item)
						case "4":
							betInfo.Seat4 = Chip2Float(item)
						case "5":
							betInfo.Seat5 = Chip2Float(item)
						case "6":
							betInfo.Seat6 = Chip2Float(item)
						case "7":
							betInfo.Seat7 = Chip2Float(item)
						case "8":
							betInfo.Seat8 = Chip2Float(item)
						case "9":
							betInfo.Seat9 = Chip2Float(item)
						case "10":
							betInfo.Seat10 = Chip2Float(item)
						}
						fmt.Print(item)
					}
				}
				e.ABInfo = betInfo
				info.ABDetail.UserDetail[j] = e
			}
		}
	case 9:
		// 彩票
		info.FPlayerFactor = Chip2Float(int64(info.PlayerFactor))
		if info.CPDetail != nil {
			info.CPDetail.FBets = Chip2Float(info.CPDetail.Bets)
			info.CPDetail.FPlayerWin = Chip2Float(info.CPDetail.PlayerWin)
			switch info.CPDetail.CardType {
			case 1:
				info.CPDetail.CardTypeStr = "高牌"
			case 2:
				info.CPDetail.CardTypeStr = "对子"
			case 3:
				info.CPDetail.CardTypeStr = "同花"
			case 4:
				info.CPDetail.CardTypeStr = "顺子"
			case 5:
				info.CPDetail.CardTypeStr = "同花顺"
			case 6:
				info.CPDetail.CardTypeStr = "豹子"
			}
			info.CPDetail.FCanWinScore = Chip2Float(info.CPDetail.CanWinScore)
			for j, e := range info.CPDetail.UserDetail {
				e.FWin = Chip2Float(e.Win)
				e.FBeforeScore = Chip2Float(e.BeforeScore)
				e.FAfterScore = Chip2Float(e.AfterScore)
				e.FBeforeCash = Chip2Float(e.BeforeCash)
				e.FAfterCash = Chip2Float(e.AfterCash)
				e.FBeforeBonus = Chip2Float(e.BeforeBonus)
				e.FAfterBonus = Chip2Float(e.AfterBonus)
				e.FCashMingTax = Chip2Float(e.CashMingTax)
				e.FBonusMingTax = Chip2Float(e.BonusMingTax)
				e.FCashAnTax = Chip2Float(e.CashAnTax)
				e.FBonusAnTax = Chip2Float(e.BonusAnTax)
				betInfo := new(entity.CPSeatBets)
				if e.SeatBets != nil {
					for i, item := range e.SeatBets {
						fmt.Print(i)
						switch i {
						case "1":
							betInfo.Seat1 = Chip2Float(item)
						case "2":
							betInfo.Seat2 = Chip2Float(item)
						case "3":
							betInfo.Seat3 = Chip2Float(item)
						case "4":
							betInfo.Seat4 = Chip2Float(item)
						case "5":
							betInfo.Seat5 = Chip2Float(item)
						case "6":
							betInfo.Seat6 = Chip2Float(item)
						}
						fmt.Print(item)
					}
				}
				e.CPInfo = betInfo
				info.CPDetail.UserDetail[j] = e
			}
		}
	// case 10:
	// 	info.FPlayerFactor = Chip2Float(int64(info.PlayerFactor))
	// 	if info.CRASHDetail != nil {
	// 		info.CRASHDetail.FBets = Chip2Float(info.CRASHDetail.Bets)
	// 		info.CRASHDetail.FPlayerWin = Chip2Float(info.CRASHDetail.PlayerLose)
	// 		isStrategy := false
	// 		if len(info.CRASHDetail.StrategyType) > 0 {
	// 			isStrategy = true
	// 		}
	// 		info.CRASHDetail.IsStrategy = isStrategy
	// 		for j, e := range info.CRASHDetail.UserDetail {
	// 			e.FBet = Chip2Float(e.Bet)
	// 			e.FWin = Chip2Float(e.Win)
	// 			e.FBeforeScore = Chip2Float(e.BeforeScore)
	// 			e.FAfterScore = Chip2Float(e.AfterScore)
	// 			e.FBeforeCash = Chip2Float(e.BeforeCash)
	// 			e.FAfterCash = Chip2Float(e.AfterCash)
	// 			e.FBeforeBonus = Chip2Float(e.BeforeBonus)
	// 			e.FAfterBonus = Chip2Float(e.AfterBonus)
	// 			e.FCashMingTax = Chip2Float(e.CashMingTax)
	// 			e.FBonusMingTax = Chip2Float(e.BonusMingTax)
	// 			e.FCashAnTax = Chip2Float(e.CashAnTax)
	// 			e.FBonusAnTax = Chip2Float(e.BonusAnTax)
	// 			info.CRASHDetail.UserDetail[j] = e
	// 		}
	// 	}
	case 11:
		// 红黑
		info.FPlayerFactor = Chip2Float(int64(info.PlayerFactor))
		if info.RBDetail != nil {
			for k, c := range info.RBDetail.CardType {
				strType := ""
				switch c {
				case 1:
					strType = "高牌"
				case 2:
					strType = "小对子"
				case 3:
					strType = "大对子"
				case 4:
					strType = "同花"
				case 5:
					strType = "顺子"
				case 6:
					strType = "同花顺"
				case 7:
					strType = "豹子"
				}
				if k == 0 {
					info.RBDetail.CardType1 = strType
				}
				if k == 1 {
					info.RBDetail.CardType2 = strType
				}
			}
			strWinner := ""
			for _, c := range info.RBDetail.Winner {
				if strWinner != "" {
					strWinner += " | "
				}
				switch c {
				case 0:
					strWinner += "幸运一击"
				case 1:
					strWinner += "红"
				case 2:
					strWinner += "黑"
				}
			}
			info.RBDetail.WinnerStr = strWinner
			info.RBDetail.FBets = Chip2Float(info.RBDetail.Bets)
			info.RBDetail.FPlayerWin = Chip2Float(info.RBDetail.PlayerWin)
			info.RBDetail.FCanWinScore = Chip2Float(info.RBDetail.CanWinScore)
			// if info.RBDetail.Winner == 1 {
			// 	info.RBDetail.WinnerStr = "ANDAR"
			// } else if info.RBDetail.Winner == 2 {
			// 	info.RBDetail.WinnerStr = "BAHAR"
			// }
			// if info.RBDetail.SideWinner == 3 {
			// 	info.RBDetail.SideWinnerStr = "1-5"
			// } else if info.RBDetail.SideWinner == 4 {
			// 	info.RBDetail.SideWinnerStr = "6-10"
			// } else if info.RBDetail.SideWinner == 5 {
			// 	info.RBDetail.SideWinnerStr = "11-15"
			// } else if info.RBDetail.SideWinner == 6 {
			// 	info.RBDetail.SideWinnerStr = "16-25"
			// } else if info.RBDetail.SideWinner == 7 {
			// 	info.RBDetail.SideWinnerStr = "26-30"
			// } else if info.RBDetail.SideWinner == 8 {
			// 	info.RBDetail.SideWinnerStr = "31-35"
			// } else if info.RBDetail.SideWinner == 9 {
			// 	info.RBDetail.SideWinnerStr = "36-40"
			// } else if info.RBDetail.SideWinner == 10 {
			// 	info.RBDetail.SideWinnerStr = "41以上"
			// }
			for j, e := range info.RBDetail.UserDetail {
				e.FWin = Chip2Float(e.Win)
				e.FBeforeScore = Chip2Float(e.BeforeScore)
				e.FAfterScore = Chip2Float(e.AfterScore)
				e.FBeforeCash = Chip2Float(e.BeforeCash)
				e.FAfterCash = Chip2Float(e.AfterCash)
				e.FBeforeBonus = Chip2Float(e.BeforeBonus)
				e.FAfterBonus = Chip2Float(e.AfterBonus)
				e.FCashMingTax = Chip2Float(e.CashMingTax)
				e.FBonusMingTax = Chip2Float(e.BonusMingTax)
				e.FCashAnTax = Chip2Float(e.CashAnTax)
				e.FBonusAnTax = Chip2Float(e.BonusAnTax)
				betInfo := new(entity.ABSeatBets)
				if e.SeatBets != nil {
					for i, item := range e.SeatBets {
						fmt.Print(i)
						switch i {
						case "0":
							betInfo.Seat1 = Chip2Float(item)
						case "1":
							betInfo.Seat2 = Chip2Float(item)
						case "2":
							betInfo.Seat3 = Chip2Float(item)
						}
						fmt.Print(item)
					}
				}
				e.ABInfo = betInfo
				info.RBDetail.UserDetail[j] = e
			}
		}
	case 12:
		// Rummy双人
		if info.RMDetail != nil {
			for j, e := range info.RMDetail {
				e.FScore = Chip2Float(int64(e.Score))
				e.FBeforeScore = Chip2Float(int64(e.BeforeScore))
				e.FAfterScore = Chip2Float(int64(e.AfterScore))
				e.FBeforeCash = Chip2Float(int64(e.BeforeCash))
				e.FAfterCash = Chip2Float(int64(e.AfterCash))
				e.FBeforeBonus = Chip2Float(int64(e.BeforeBonus))
				e.FAfterBonus = Chip2Float(int64(e.AfterBonus))
				e.FCashStock = Chip2Float(int64(e.CashStock))
				e.FBonusStock = Chip2Float(int64(e.BonusStock))
				e.FCashMingTax = Chip2Float(int64(e.CashMingTax))
				e.FBonusMingTax = Chip2Float(int64(e.BonusMingTax))
				e.FCashAnTax = Chip2Float(int64(e.CashAnTax))
				e.FBonusAnTax = Chip2Float(int64(e.BonusAnTax))
				info.RMDetail[j] = e
			}
		}
	case 14:
		if info.MinesDetail != nil {
			info.MinesDetail.FBets = Chip2Float(info.MinesDetail.Bets)
			info.MinesDetail.FScore = Chip2Float(info.MinesDetail.Score)
			info.MinesDetail.FBeforeScore = Chip2Float(info.MinesDetail.BeforeScore)
			info.MinesDetail.FAfterScore = Chip2Float(info.MinesDetail.AfterScore)
			info.MinesDetail.FBeforeCash = Chip2Float(info.MinesDetail.BeforeCash)
			info.MinesDetail.FAfterCash = Chip2Float(info.MinesDetail.AfterCash)

			for i, pit := range info.MinesDetail.MinesPits {
				step := 0
				for j, stepPit := range info.MinesDetail.StepPits {
					if i == int(stepPit) {
						step = j + 1
						break
					}
				}
				info.MinesDetail.FMinesPits = append(info.MinesDetail.FMinesPits, entity.FMinesPit{
					Pit:     pit,
					Step:    int32(step),
					StepPit: pit == 1 || pit == 2,
				})
			}
		}
	}
	return info

}

func (this *playerService) chipList20(info *entity.ExternalDetail) *entity.ExternalDetail {
	if len(info.RewardDetail) > 0 {
		for k, v := range info.RewardDetail {
			c, _ := ConvertToIndiaTime(v.Ctime)
			v.STime = c
			v.FAmount = Chip2Float(v.Amount)
			info.RewardDetail[k] = v
		}
	}
	info.FAmount = Chip2Float(info.Amount)

	// 游戏列表
	for k, v := range GtypeNameMap {
		if k == int(info.GameId) {
			info.GameName = v
		}
	}
	return info
}

// 转换为分展示
func (this *playerService) chipList4(list []entity.Detail) []entity.Detail {
	for _, v := range list {
		c, _ := ConvertToIndiaTime(v.BeginTime)
		v.STime = c
		c1, _ := ConvertToIndiaTime(v.EndTime)
		v.ETime = c1

		// for j, e := range v.TPDetail {
		// 	e.FScore = Chip2Float(int64(e.Score))
		// 	e.FBeforeScore = Chip2Float(int64(e.BeforeScore))
		// 	e.FAfterScore = Chip2Float(int64(e.AfterScore))
		// 	e.FBeforeCash = Chip2Float(int64(e.BeforeCash))
		// 	e.FAfterCash = Chip2Float(int64(e.AfterCash))
		// 	e.FBeforeBonus = Chip2Float(int64(e.BeforeBonus))
		// 	e.FAfterBonus = Chip2Float(int64(e.AfterBonus))
		// 	e.FBet = Chip2Float(int64(e.Bet))
		// 	e.FBottom = Chip2Float(int64(e.Bottom))
		// 	e.FCashStock = Chip2Float(int64(e.CashStock))
		// 	e.FBonusStock = Chip2Float(int64(e.BonusStock))
		// 	e.FCashMingTax = Chip2Float(int64(e.CashMingTax))
		// 	e.FBonusMingTax = Chip2Float(int64(e.BonusMingTax))
		// 	e.FCashAnTax = Chip2Float(int64(e.CashAnTax))
		// 	e.FBonusAnTax = Chip2Float(int64(e.BonusAnTax))
		// 	v.TPDetail[j] = e
		// }
		// list[k] = v
	}
	return list
}

// 添加VIP
func (this *playerService) AddVip(shop *entity.Vip) error {
	shop.Ctime = bson.Now()
	if !Insert(Vips, shop) {
		return errors.New("写入失败:" + shop.Id)
	}
	return nil
}

// 获取商品
func (this *playerService) GetVip(id string) (*entity.Vip, error) {
	shop := new(entity.Vip)
	Get(Vips, id, shop)
	if shop.Id == "" {
		return shop, errors.New("商品不存在")
	}
	return shop, nil
}

// 移除商品
func (this *playerService) DelVip(id string) error {
	if Delete(Vips, bson.M{"_id": id}) {
		return nil
	}
	return errors.New("移除失败")
}

// 获取列表
func (this *playerService) GetVipList(page, pageSize int, m bson.M) ([]entity.Vip, error) {
	var list []entity.Vip
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	err := Vips.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	return list, err
}

// 获取总数
func (this *playerService) GetVipListTotal(m bson.M) (int64, error) {
	return int64(Count(Vips, m)), nil
}

// 获取列表
func (this *playerService) GetEnvList(page, pageSize int, m bson.M) ([]entity.Env, error) {
	var list []entity.Env
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	err := Envs.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	return list, err
}

// 获取总数
func (this *playerService) GetEnvListTotal(m bson.M) (int64, error) {
	return int64(Count(Envs, m)), nil
}

// 添加VIP
func (this *playerService) AddEnv(shop *entity.Env) error {
	if !Insert(Envs, shop) {
		return errors.New("写入失败:" + shop.Key)
	}
	return nil
}

// 获取商品
func (this *playerService) GetEnv(id string) (*entity.Env, error) {
	shop := new(entity.Env)
	Get(Envs, id, shop)
	if shop.Key == "" {
		return shop, errors.New("商品不存在")
	}
	return shop, nil
}

// 移除商品
func (this *playerService) DelEnv(id string) error {
	if Delete(Envs, bson.M{"_id": id}) {
		return nil
	}
	return errors.New("移除失败")
}

/*
	银行卡黑名单
*/
// 获取银行卡黑名单总数
func (this *playerService) GetCardBlacklistTotal(m bson.M) (int64, error) {
	return int64(Count(CardBlacklists, m)), nil
}

// 转换为分展示
func (this *playerService) chipList8(list []entity.CardBlacklist) []entity.CardBlacklist {
	for k, v := range list {
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		list[k] = v
	}
	return list
}

func (this *playerService) GetCardBlacklist(page, pageSize int, m bson.M) ([]entity.CardBlacklist, error) {
	var list []entity.CardBlacklist
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	err := CardBlacklists.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList8(list)
	return list, err
}

// 添加银行卡号黑名单
func (this *playerService) AddCardBlacklist(info *entity.CardBlacklist) error {
	if info.Card == "" {
		return errors.New("银行卡号不能为空")
	}
	info.Ctime = bson.Now()
	if !Insert(CardBlacklists, info) {
		return errors.New("写入失败:" + info.Card)
	}
	return nil
}

// 删除银行卡号黑名单
func (this *playerService) DelCard(id string) error {
	if id == "" {
		return errors.New("银行卡号不能为空")
	}
	m := bson.M{"_id": id}
	if Delete(CardBlacklists, m) {
		return nil
	}
	return errors.New("更新失败")
}

// 获取银行卡号黑名单
func (this *playerService) GetCard(id string) (*entity.CardBlacklist, error) {
	info := new(entity.CardBlacklist)
	Get(CardBlacklists, id, info)
	if info.Card == "" {
		return info, errors.New("银行卡号不存在")
	}
	return info, nil
}

/*
	设备码黑名单
*/
// 获取设备码黑名单总数
func (this *playerService) GetEquipmentBlacklistTotal(m bson.M) (int64, error) {
	return int64(Count(EquipmentBlacklists, m)), nil
}

func (this *playerService) GetEquipmentBlacklist(page, pageSize int, m bson.M) ([]entity.EquipmentBlacklist, error) {
	var list []entity.EquipmentBlacklist
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	err := EquipmentBlacklists.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList7(list)
	return list, err
}

// 转换为分展示
func (this *playerService) chipList7(list []entity.EquipmentBlacklist) []entity.EquipmentBlacklist {
	for k, v := range list {
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		list[k] = v
	}
	return list
}

// 添加设备码黑名单
func (this *playerService) AddEquipmentBlacklist(info *entity.EquipmentBlacklist) error {
	if info.Code == "" {
		return errors.New("设备码不能为空")
	}
	info.Ctime = bson.Now()
	if !Insert(EquipmentBlacklists, info) {
		return errors.New("写入失败:" + info.Code)
	}
	return nil
}

// 删除银行卡号黑名单
func (this *playerService) DelEquipment(id string) error {
	if id == "" {
		return errors.New("设备码不能为空")
	}
	m := bson.M{"_id": id}
	if Delete(EquipmentBlacklists, m) {
		return nil
	}
	return errors.New("更新失败")
}

// 获取银行卡号黑名单
func (this *playerService) GetEquipment(id string) (*entity.EquipmentBlacklist, error) {
	info := new(entity.EquipmentBlacklist)
	Get(EquipmentBlacklists, id, info)
	if info.Code == "" {
		return info, errors.New("设备码不存在")
	}
	return info, nil
}

// 获取对局曲线数据
func (this *playerService) GetGameCurve(userid string, startStr, endStr string) ([]entity.Detail, error) {
	list := make([]entity.Detail, 0)
	// s := fmt.Sprintf("%s 00:00:00", startStr)
	// startTime := utils.Str2Time(s).Unix()
	// e := fmt.Sprintf("%s 23:59:59", endStr)
	// endTime := utils.Str2Time(e).Unix()
	m := FindByDate1(startStr, endStr, "begin_time", "begin_time")
	if userid != "" {
		m["players"] = bson.M{
			"$regex":   fmt.Sprintf("\\b%s\\b", userid),
			"$options": "i",
		}
	}
	Details.Find(m).All(&list)
	// pipeline := []bson.M{
	// 	{
	// 		"$match": m,
	// 	},
	// 	{
	// 		"$group": bson.M{
	// 			"_id": "$userid",
	// 			"Count": bson.M{
	// 				"$sum": 1,
	// 			},
	// 			"Score": bson.M{
	// 				"$sum": "$score",
	// 			},
	// 		},
	// 	},
	// }

	// Get(Details, id, info)
	// if info.Code == "" {
	// 	return info, errors.New("设备码不存在")
	// }
	return list, nil
}

/*
VB流水
*/

/*金币流水*/
// 获取列表
func (this *playerService) GetVbWaterList(page, pageSize int, m bson.M) ([]entity.LogOutDiamond, error) {
	var list []entity.LogOutDiamond
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	err := LogVBDiamonds.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	//转换为分展示
	list = this.chipList28(list)
	return list, err
}

// 获取总数
func (this *playerService) GetVbWaterTotal(m bson.M) (int64, error) {
	return int64(Count(LogVBDiamonds, m)), nil
}

// 转换为分展示
func (this *playerService) chipList28(list []entity.LogOutDiamond) []entity.LogOutDiamond {
	for k, v := range list {
		s, _ := ConvertToIndiaTime(v.Ctime)
		v.ShowTime = s
		v.FBeforeOutCash = Chip2Float(v.BeforeOutCash)
		v.FAfterOutCash = Chip2Float(v.AfterOutCash)
		v.FBeforeCash = Chip2Float(v.BeforeCash)
		v.FAfterCash = Chip2Float(v.AfterCash)
		list[k] = v
	}
	return list
}

// 获取服务器在线玩家数据
func calculateData() []entity.OnlineUser {
	result, _ := GmRequest(pb.WebOnlineUser, pb.CONFIG_UPSERT, pb.NULL)
	beego.Trace("result: ", result)
	list, flag := ConvertToUserSlice(result)
	if flag {
		return list
	} else {
		return make([]entity.OnlineUser, 0)
	}
}

var (
	// 在线用户缓存数据
	cacheOnline cache.Cache
)

// 获取在线用户列表
func (this *playerService) ListOnlineUsers() []entity.OnlineUser {
	// 尝试从缓存中获取数据
	userlist := make([]entity.OnlineUser, 0)
	if cacheOnline != nil {
		data := cacheOnline.Get("cached_data")
		flag := false
		if data != nil {
			userlist, flag = ConvertToUserSlice(data)
		} else {
			// 如果缓存中没有数据，则进行数据计算和缓存操作
			calculatedData := calculateData()
			// 将计算得到的数据存入缓存

			cacheOnline.Put("cached_data", calculatedData, 3600)
			userlist = calculatedData
			flag = true
		}
		if !flag {
			userlist = make([]entity.OnlineUser, 0)
		}
	} else {
		// 初始化缓存适配器
		cacheOnline, _ = cache.NewCache("memory", `{"interval":3600}`)

		// 如果缓存中没有数据，则进行数据计算和缓存操作
		calculatedData := calculateData()
		// 将计算得到的数据存入缓存

		cacheOnline.Put("cached_data", calculatedData, 3600)
		userlist = calculatedData
	}

	return userlist
}

func (this *playerService) IsUsersOnline(userids []string) (onlines map[string]bool) {
	onlines = make(map[string]bool)
	for _, userid := range userids {
		onlines[userid] = false
	}

	userlist := this.ListOnlineUsers()

	for _, user := range userlist {
		if _, ok := onlines[user.Id]; ok {
			onlines[user.Id] = true
		}
	}
	return
}
