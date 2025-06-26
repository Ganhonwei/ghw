package service

import (
	"fmt"

	"goserver/internal/web/admin/app/entity"

	"github.com/globalsign/mgo/bson"
)

// 系统动态
type actionService struct{}

// 添加记录
func (this *actionService) Add(action, actor, objectType string, objectId string, extra string, remark string) bool {
	act := new(entity.Action)
	act.Action = action
	act.Actor = actor
	act.ObjectType = objectType
	act.ObjectId = objectId
	act.Extra = extra
	act.CreateTime = bson.Now()
	act.Message = remark
	act.Id = bson.NewObjectId().Hex()
	Insert(Actions, act)
	return true
}

// 登录动态
func (this *actionService) Login(userName string, userId string, ip string) {
	this.Add("login", userName, "user", userId, ip, "")
}

// 退出登录
func (this *actionService) Logout(userName string, userId string, ip string) {
	this.Add("logout", userName, "user", userId, ip, "")
}

// 更新个人信息
func (this *actionService) UpdateProfile(userName string, userId string) {
	this.Add("update_profile", userName, "user", userId, "", "")
}

// 更新个人钻石,otype操作类型,oid操作目标id,extra操作的数量
func (this *actionService) UpdateDiamond(userName, otype, oid, extra string, remark string) {
	this.Add("update_diamond", userName, otype, oid, extra, remark)
}
func (this *actionService) UpdateNumber(userName, otype, oid, extra string, remark string) {
	this.Add("update_number", userName, otype, oid, extra, remark)
}
func (this *actionService) UpdateExpend(userName, otype, oid, extra string, remark string) {
	this.Add("update_expend", userName, otype, oid, extra, remark)
}
func (this *actionService) UpdateChip(userName, otype, oid, extra string, remark string) {
	this.Add("update_chip", userName, otype, oid, extra, remark)
}

// 注册动态
func (this *actionService) Regist(userName string, agent string, ip string) {
	this.Add("regist", userName, "agent", agent, ip, "")
}

// 角色动态
func (this *actionService) AddRole(userName string, roleName string) {
	this.Add("add_role", userName, "role_name", roleName, "", "")
}
func (this *actionService) DelRole(userName string, roleId string) {
	this.Add("del_role", userName, "role_id", roleId, "", "")
}
func (this *actionService) UpdateRole(userName string, roleName string) {
	this.Add("update_role", userName, "role_name", roleName, "", "")
}
func (this *actionService) PermRole(userName string, roleName string) {
	this.Add("perm_role", userName, "role_name", roleName, "", "")
}

// 用户动态
func (this *actionService) AddUser(userName string, user_name string) {
	this.Add("add_user", userName, "user_name", user_name, "", "")
}
func (this *actionService) UpdateUser(userName string, id string) {
	this.Add("update_user", userName, "user_id", id, "", "")
}
func (this *actionService) DelUser(userName string, id string) {
	this.Add("del_user", userName, "user_id", id, "", "")
}

// 代理动态
func (this *actionService) AddAgency(userName string, phone, agency string) {
	this.Add("add_agency", userName, "phone", phone, agency, "")
}
func (this *actionService) UpdateBuild(userName string, userid, agent string) {
	this.Add("build_agency", userName, "userid", userid, agent, "")
}
func (this *actionService) AddApplyCash(userName string, money string) {
	this.Add("apply_cash", userName, "money", money, "", "")
}
func (this *actionService) ExtractApplyCash(userName string, orderid string) {
	this.Add("extract_cash", userName, "orderid", orderid, "", "")
}
func (this *actionService) UpdateAgency(userName string, agent, rate string) {
	this.Add("update_agency", userName, "rate", rate, agent, "")
}

// 公告动态
func (this *actionService) AddNotice(userName string, notice_id string) {
	this.Add("add_notice", userName, "notice_id", notice_id, "", "")
}
func (this *actionService) UpdateNotice(userName string, notice_id string) {
	this.Add("update_notice", userName, "notice_id", notice_id, "", "")
}
func (this *actionService) Notice(userName string, notice_id string) {
	this.Add("notice", userName, "notice_id", notice_id, "", "")
}
func (this *actionService) DelNotice(userName string, notice_id string) {
	this.Add("del_notice", userName, "notice_id", notice_id, "", "")
}

// 商城动态
func (this *actionService) AddShop(userName string, shop_id string, remark string) {
	this.Add("add_shop", userName, "shop_id", shop_id, "", remark)
}
func (this *actionService) UpdateShop(userName string, shop_id string, remark string) {
	this.Add("update_shop", userName, "shop_id", shop_id, "", remark)
}
func (this *actionService) Shop(userName string, shop_id string, remark string) {
	this.Add("shop", userName, "shop_id", shop_id, "", remark)
}
func (this *actionService) DelShop(userName string, shop_id string, remark string) {
	this.Add("del_shop", userName, "shop_id", shop_id, "", remark)
}

// 商城动态
func (this *actionService) AddGame(userName string, game_id string) {
	this.Add("add_game", userName, "game_id", game_id, "", "")
}
func (this *actionService) Game(userName string, game_id string) {
	this.Add("game", userName, "game_id", game_id, "", "")
}
func (this *actionService) DelGame(userName string, game_id string) {
	this.Add("del_game", userName, "game_id", game_id, "", "")
}
func (this *actionService) EditGame(userName string, game_id string) {
	this.Add("edit_game", userName, "game_id", game_id, "", "")
}

// 获取动态列表
func (this *actionService) GetList(page, pageSize int, m bson.M) ([]entity.Action, error) {
	var list []entity.Action
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "create_time", false)
	err := Actions.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	if err == nil {
		num := len(list)
		for i := 0; i < num; i++ {
			this.format(&list[i])
		}
	}
	return list, err
}

// 获取动态列表总数
func (this *actionService) GetTotal(m bson.M) (int64, error) {
	return int64(Count(Actions, m)), nil
}

// 格式化
func (this *actionService) format(action *entity.Action) {
	c, _ := ConvertToIndiaTime(action.CreateTime.Unix())
	action.CreateTime = c
	switch action.Action {
	case "login":
		action.ActionName = fmt.Sprintf("%s 登录系统，IP:%s", action.Actor, action.Extra)
	case "logout":
		action.ActionName = fmt.Sprintf("%s 退出系统", action.Actor)
	case "update_profile":
		action.ActionName = fmt.Sprintf("%s 更新了个人资料", action.Actor)
	case "create_task":
		action.ActionName = fmt.Sprintf("%s 创建了编号为 %s 的发布单", action.Actor, action.ObjectId)
	case "regist":
		action.ActionName = fmt.Sprintf("%s 注册成功，IP:%s", action.Actor, action.Extra)
	case "add_role":
		action.ActionName = fmt.Sprintf("%s 添加角色，角色名称为%s", action.Actor, action.ObjectId)
	case "del_role":
		action.ActionName = fmt.Sprintf("%s 删除角色，角色ID为%s", action.Actor, action.ObjectId)
	case "update_role":
		action.ActionName = fmt.Sprintf("%s 更新角色，角色名称为%s", action.Actor, action.ObjectId)
	case "perm_role":
		action.ActionName = fmt.Sprintf("%s 更新角色权限，角色名称为%s", action.Actor, action.ObjectId)
	case "add_user":
		action.ActionName = fmt.Sprintf("%s 添加账号，账号名称为%s", action.Actor, action.ObjectId)
	case "update_user":
		action.ActionName = fmt.Sprintf("%s 更新账号，账号ID：%s", action.Actor, action.ObjectId)
	case "del_user":
		action.ActionName = fmt.Sprintf("%s 删除账号，账号ID：%s", action.Actor, action.ObjectId)
	case "edit_share_superior":
		action.ActionName = fmt.Sprintf("%s 分享活动-上下级设置，修改账号ID：%s的上级", action.Actor, action.ObjectId)
	case "update_channel":
		action.ActionName = fmt.Sprintf("%s 渠道设置，修改渠道名称：%s的数据", action.Actor, action.Extra)
	case "add_channel":
		action.ActionName = fmt.Sprintf("%s 渠道设置，新增渠道名称：%s的数据", action.Actor, action.Extra)
	case "del_channel":
		action.ActionName = fmt.Sprintf("%s 渠道设置，删除渠道名称：%s的数据", action.Actor, action.Extra)
	case "stock_cash_edit":
		action.ActionName = fmt.Sprintf("%s 修改彩金库存，房间ID：%s", action.Actor, action.ObjectId)
	case "stock_bonus_edit":
		action.ActionName = fmt.Sprintf("%s 修改奖励金库存，房间ID：%s", action.Actor, action.ObjectId)
	case "update_point_control":
		action.ActionName = fmt.Sprintf("%s 设置点控，玩家ID：%s", action.Actor, action.ObjectId)
	case "set_system_icon_config":
		action.ActionName = fmt.Sprintf("%s 系统设置-游戏图标", action.Actor)
	case "set_game_icon_config":
		action.ActionName = fmt.Sprintf("%s 系统设置-游戏图标", action.Actor)
	case "set_pay_channel":
		action.ActionName = fmt.Sprintf("%s 财务管理-通道配置", action.Actor)
	case "set_system_controls_config":
		action.ActionName = fmt.Sprintf("%s 系统设置-控制设置", action.Actor)
	case "send_feed_back":
		action.ActionName = fmt.Sprintf("%s 客服消息回复", action.Actor)
	case "del_banner":
		action.ActionName = fmt.Sprintf("%s 删除Banner", action.Actor)
	case "del_cdkkey":
		action.ActionName = fmt.Sprintf("%s 删除%sCDK", action.Actor, action.ObjectId)
	case "pay_callback":
		action.ActionName = fmt.Sprintf("%s 充值记录-支付手动回调，订单ID：%s", action.Actor, action.ObjectId)
	case "withdraw_audit":
		action.ActionName = fmt.Sprintf("%s 提现记录-审核通过，订单ID：%s", action.Actor, action.ObjectId)
	case "withdraw_frozen":
		action.ActionName = fmt.Sprintf("%s 提现记录-提现冻结，订单ID：%s", action.Actor, action.ObjectId)
	case "withdraw_return":
		action.ActionName = fmt.Sprintf("%s 提现记录-提现退回，订单ID：%s", action.Actor, action.ObjectId)
	case "withdraw_ban":
		action.ActionName = fmt.Sprintf("%s 提现记录-封号，订单ID：%s", action.Actor, action.ObjectId)
	case "withdraw_tag":
		action.ActionName = fmt.Sprintf("%s 提现记录-提现标记，订单ID：%s", action.Actor, action.ObjectId)
	case "withdraw_transfer_order":
		action.ActionName = fmt.Sprintf("%s 提现记录-提现转单，订单ID：%s", action.Actor, action.ObjectId)
	case "add_set_withdraw":
		action.ActionName = fmt.Sprintf("%s 添加提现配置数据", action.Actor)
	case "update_set_withdraw":
		action.ActionName = fmt.Sprintf("%s 修改提现配置数据", action.Actor)
	case "del_set_withdraw":
		action.ActionName = fmt.Sprintf("%s 删除提现配置数据", action.Actor)
	case "add_pay_blacklist":
		action.ActionName = fmt.Sprintf("%s 添加支付黑名单，手机号码：%s", action.Actor, action.ObjectId)
	case "del_pay_blacklist":
		action.ActionName = fmt.Sprintf("%s 删除支付黑名单", action.Actor)
	case "add_pay_blacklist_upload":
		action.ActionName = fmt.Sprintf("%s 批量添加支付黑名单", action.Actor)
	case "ServerWhite_add":
		action.ActionName = fmt.Sprintf("%s 添加开服IP限制，IP：%s", action.Actor, action.ObjectId)
	case "ServerWhite_del":
		action.ActionName = fmt.Sprintf("%s 删除开服IP限制，IP：%s", action.Actor, action.ObjectId)
	case "IPwhite_add":
		action.ActionName = fmt.Sprintf("%s 添加IP限制，IP：%s", action.Actor, action.ObjectId)
	case "IPwhite_del":
		action.ActionName = fmt.Sprintf("%s 删除IP限制，IP：%s", action.Actor, action.ObjectId)
	case "CardBlacklist_add":
		action.ActionName = fmt.Sprintf("%s 添加银行卡黑名单，银行卡号：%s", action.Actor, action.ObjectId)
	case "CardBlacklist_del":
		action.ActionName = fmt.Sprintf("%s 删除银行卡黑名单，银行卡号：%s", action.Actor, action.ObjectId)
	case "Equipment_Blacklist_add":
		action.ActionName = fmt.Sprintf("%s 添加设备码黑名单那，设备码：%s", action.Actor, action.ObjectId)
	case "Equipment_Blacklist_del":
		action.ActionName = fmt.Sprintf("%s 删除设备码黑名单那，设备码：%s", action.Actor, action.ObjectId)
	case "Upload_Config":
		strStatus := "失败"
		if action.Extra == "0" {
			strStatus = "成功"
		}
		action.ActionName = fmt.Sprintf("%s 上传配置文件%s，上传文件类型：%s", action.Actor, strStatus, action.ObjectId)
	case "update_diamond":
		strOpt := ""
		if action.ObjectType == "63" {
			strOpt = "充值彩金"
		}
		if action.ObjectType == "64" {
			strOpt = "修改彩金"
		}
		if action.ObjectType == "65" {
			strOpt = "充值奖励金"
		}
		if action.ObjectType == "66" {
			strOpt = "修改奖励金"
		}
		if action.ObjectType == "65" {
			strOpt = "增加可提现金币"
		}
		action.ActionName = fmt.Sprintf("%s %s，用户ID：%s，金额：%s分", action.Actor, strOpt, action.ObjectId, action.Extra)
	case "add_notice":
		action.ActionName = fmt.Sprintf("%s 新增公告消息", action.Actor)
	case "update_notice":
		action.ActionName = fmt.Sprintf("%s 修改公告消息", action.Actor)
	case "notice":
		action.ActionName = fmt.Sprintf("%s 公告消息", action.Actor)
	case "del_notice":
		action.ActionName = fmt.Sprintf("%s 删除公告消息", action.Actor)
	case "add_shop":
		action.ActionName = fmt.Sprintf("%s 新增充值配置", action.Actor)
	case "update_shop":
		action.ActionName = fmt.Sprintf("%s 修改充值配置", action.Actor)
	case "shop":
		action.ActionName = fmt.Sprintf("%s 充值配置", action.Actor)
	case "del_shop":
		action.ActionName = fmt.Sprintf("%s 删除充值配置", action.Actor)
	case "add_game":
		action.ActionName = fmt.Sprintf("%s 新增游戏房间", action.Actor)
	case "edit_game":
		action.ActionName = fmt.Sprintf("%s 修改游戏房间", action.Actor)
	case "game":
		action.ActionName = fmt.Sprintf("%s 游戏房间", action.Actor)
	case "del_game":
		action.ActionName = fmt.Sprintf("%s 删除游戏房间", action.Actor)
	case "NoviceStock_cash_edit":
		action.ActionName = fmt.Sprintf("%s 新手库存修改彩金库存，游戏ID：%s", action.Actor, action.ObjectId)
	case "Upload_Config_Notify":
		action.ActionName = fmt.Sprintf("%s 重新加载游戏配置文件", action.Actor)
	}
}
