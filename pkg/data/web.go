package data

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

const (
	ITYPE1 uint32 = 1 //钻石
	ITYPE2 uint32 = 2 //金币
	ITYPE3 uint32 = 3 //VIP
	// 操作类型
	Pass     = 1 // 通过
	Delivery = 5 // 派单
	Back     = 2 // 回退
	Freeze   = 3 // 冻结
	Transfer = 4 // 转单
)

type RespErr struct {
	ErrCode int    `json:"errcode"` //错误码
	ErrMsg  string `json:"errmsg"`  //错误信息
	Result  string `json:"result"`  //正常时返回信息
}

// 货币变更请求
type ReqMsg struct {
	Userid string `json:"userid"` //角色ID
	Rtype  int    `json:"rtype"`  //类型
	Itemid uint32 `json:"itemid"` //物品,1钻石,2金币,3VIP
	Amount int32  `json:"amount"` //数量
}

// 在线状态请求
type ReqOnlineStatusMsg struct {
	Userid []string `json:"userid"` //角色ID
}

type RespOnlineStatusMsg struct {
	Userid map[string]int `json:"userid"` //角色ID,1在线
}

// 发布公告
type ReqNoticeMsg struct {
	Id      string    `json:"id"`
	Rtype   int       `json:"rtype"`    //类型
	Atype   uint32    `json:"atype"`    //分包类型
	Acttype int       `json:"act_type"` //操作类型
	Top     int       `json:"top"`      //置顶
	Num     int       `json:"num"`      //广播次数
	Del     int       `json:"del"`      //是否移除
	Content string    `json:"content"`  //广播内容
	Etime   time.Time `json:"etime"`    //过期时间
	Ctime   time.Time `json:"ctime"`    //创建时间
}

// 绑定请求(修改绑定)
type ReqBuildMsg struct {
	Userid string `json:"userid"` //角色ID
	Agent  string `json:"agent"`  //代理ID
}

// 代理商赠送给下级用户
type ReqGiveDiamondMsg struct {
	Userid string `json:"userid"` //角色ID
	Agent  string `json:"agent"`  //代理商
	Rtype  int    `json:"rtype"`  //类型
	Itemid uint32 `json:"itemid"` //物品,1钻石,2金币
	Amount int32  `json:"amount"` //数量
}

// 设置手牌
type ReqSetHandsMsg struct {
	Userid string   `json:"userid"` //代理商
	Round  uint32   `json:"round"`  //牌局
	Hands  []uint32 `json:"hands"`  //手牌
}

// 获取设置手牌
type ReqGetHandsMsg struct {
	Userid string `json:"userid"` //代理商
}

type RespGetHandsMsg struct {
	List []Rounds `json:"list"` //手牌列表
}

type Rounds struct {
	Round uint32   `json:"round"` //牌局
	Hands []uint32 `json:"hands"` //手牌
}

// 发布商品
type ReqShopMsg struct {
	Id     string    `json:"id"`     //购买ID
	Atype  uint32    `json:"atype"`  //分包类型
	Status int       `json:"status"` //物品状态,1=热卖
	Propid int       `json:"propid"` //兑换的物品,1=钻石
	Payway int       `json:"payway"` //支付方式,1=RMB
	Number uint32    `json:"number"` //兑换的数量
	Price  uint32    `json:"price"`  //支付价格
	Name   string    `json:"name"`   //物品名字
	Info   string    `json:"info"`   //物品信息
	Del    int       `json:"del"`    //是否移除
	Etime  time.Time `json:"etime"`  //过期时间
	Ctime  time.Time `json:"ctime"`  //创建时间
}

// 设置变量
// key      value
// regist_diamond    注册赠送钻石
// regist_coin       注册赠送金币
// build             绑定赠送
// first_pay_multi   首充送n倍
// first_pay_coin    首充送金币
// relieve           救济金次数
// prizedraw         转盘抽奖次数
// bankrupt_coin     破产金额
// relieve_coin      救济金额
type ReqEnvMsg struct {
	Key   string `json:"key"`   //key
	Value int32  `json:"value"` //value
}

type ReqGetEnvMsg struct {
	Key string `json:"key"` //key
}

type ReqDelEnvMsg struct {
	Key string `json:"key"` //key
}

// 设置变量
type RespEnvMsg struct {
	List []ReqEnvMsg `json:"list"` //list
}

// 发布抽奖
type ReqPrizeMsg struct {
	Id     string    `json:"id"`     //id
	Rate   uint32    `json:"rate"`   //概率
	Rtype  int       `json:"rtype"`  //类型,1钻石,2金币
	Amount int32     `json:"amount"` //数量
	Del    int       `json:"del"`    //是否移除
	Ctime  time.Time `json:"ctime"`  //创建时间
}

// 发布宝箱
type ReqBoxMsg struct {
	Id       string    `json:"id"`       //id
	Duration uint32    `json:"duration"` //时间(秒)
	Rtype    int       `json:"rtype"`    //类型,1钻石,2金币
	Amount   int32     `json:"amount"`   //数量
	Del      int       `json:"del"`      //是否移除
	Ctime    time.Time `json:"ctime"`    //创建时间
}

// 房间数据
type ReqRoomMsg struct {
	Userid string `json:"userid"` //角色ID
	Rtype  int    `json:"rtype"`  //0打印数据,1离开房间,2解散房间
}

type RespRoomMsg struct {
	Userid   string `json:"userid"`    //角色ID
	DeskData string `json:"desk_data"` //代理ID
}

// 经典
type ReqClassicMsg struct {
	Id      string    `json:"id"`      //id
	Ptype   int       `json:"ptype"`   //玩法类型1看牌抢庄,3通比牛牛4牛牛坐庄
	Rtype   int       `json:"rtype"`   //房间类型1初级,2中级,3高级,4大师
	Ante    uint32    `json:"ante"`    //房间底分
	Minimum uint32    `json:"minimum"` //房间最低
	Maximum uint32    `json:"maximum"` //房间最高0表示没限制
	Del     int       `json:"del"`     //是否移除
	Ctime   time.Time `json:"ctime"`   //创建时间
}

// vip
type ReqVipMsg struct {
	Id     string    `json:"id"`     //ID
	Level  int       `json:"level"`  //等级
	Number uint32    `json:"number"` //等级充值金额数量限制
	Pay    uint32    `json:"pay"`    //充值赠送百分比5=赠送充值的5%
	Prize  uint32    `json:"prize"`  //赠送抽奖次数
	Kick   int       `json:"kick"`   //经典场可踢人次数
	Del    int       `json:"del"`    //是否移除
	Ctime  time.Time `json:"ctime"`  //创建时间
}

// dan
type ReqDanMsg struct {
	Id     string     `json:"_id"`    //序号
	Name   string     `json:"name"`   //名称
	Dan    int        `json:"dan"`    //段0-6
	Ante   uint32     `json:"ante"`   //隐藏底分
	Number uint32     `json:"number"` //预计局数
	Coin   uint32     `json:"coin"`   //最低限制金币
	Di     uint32     `json:"di"`     //房间底分
	Level  []ReqStars `json:"level"`  //星级,积分
	Ctime  time.Time  `json:"ctime"`  //创建时间
	Del    int        `json:"del"`    //是否移除
}

type ReqStars struct {
	Stars  int   `json:"stars"`  //星
	Points int32 `json:"points"` //积分
}

// task
type ReqTaskMsg struct {
	Id      string    `json:"id"`      //任务序号
	Name    string    `json:"name"`    //任务名称
	Diamond uint32    `json:"diamond"` //钻石奖励
	Coin    uint32    `json:"coin"`    //金币奖励
	Ctime   time.Time `json:"ctime"`   //创建时间
	Del     int       `json:"del"`     //是否移除
}

// 疯狂投注
type ReqBettingMsg struct {
	Seat uint32  `json:"seat"` //位置
	Odds float32 `json:"odds"` //赔率
}

// 全民刮奖
type ReqLotteryMsg struct {
	Times   uint32 `json:"times"`   //次数
	Diamond uint32 `json:"diamond"` //钻石
}

// 玩家黑白名单
type BlackList struct {
	Userid string `json:"userid"` //用户id
	Status int    `json:"status"` //状态
}

// ip白名单
type IPwhite struct {
	IP    string    `bson:"_id" json:"id"`      // IP
	Ctime time.Time `bson:"ctime" json:"ctime"` // 添加时间
}

// 在线玩家
type OnlineUser struct {
	Id         string `json:"id"`          //用户id
	Nanme      string `json:"name"`        //呢称
	Channel    string `json:"channel"`     //渠道
	GameId     string `json:"game_id"`     //游戏id
	RoomId     string `json:"room_id"`     //房间
	Asset      int64  `json:"asset"`       //总资产
	ShowAsset  int64  `json:"show_asset"`  //显示资产
	Cash       int64  `json:"cash"`        //彩金
	Bouns      int64  `json:"bouns"`       //游戏金
	OtherAsset int64  `json:"other_asset"` //其他资产
	Recharge   uint32 `json:"recharge"`    //已充值
	CashOut    uint32 `json:"cash_out"`    //已提现
	WinScore   int64  `json:"win_score"`   //当前赢分
}

// 修改玩家数据
type ModifyUserData struct {
	UserId      string `json:"userId"`     //用户id
	Photo       string `json:"photo"`      //头像
	NickName    string `json:"nickname"`   //名称
	Phone       string `json:"phone"`      //电话
	BankName    string `json:"bankName"`   //银行名称
	BankNumber  string `json:"bankNumber"` //银行卡号
	Ifsc        string `json:"ifsc"`       //ifsc
	BundleId    string `json:"bundleId"`
	OsVersion   string `json:"osVersion"`
	AfId        string `json:"afId"`
	MediaSource string `json:"mediaSource"`
	AfKey       string `json:"afKey"`
	AppId       string `json:"appId"`
	RefGameId   string `json:"refGameId"`
	RefPkgName  string `json:"refPkgName"`
}

// 点控
type PointControl struct {
	UserId string `json:"userid"` //用户id
	Switch bool   `json:"switch"` //开关
	Factor int32  `json:"factor"` //系数
	Score  int64  `json:"score"`  //分数
}

// 修改上级
type ShareSuperior struct {
	UserId   string `json:"userid"`   //用户id
	Superior string `json:"superior"` //上级id
}

// 充值号码限制
type RechargeLimit struct {
	Phone    string    `bson:"_id" json:"id"`            //支付手机号
	Operator string    `json:"operator" bson:"operator"` // 操作人
	Ctime    time.Time `bson:"ctime" json:"ctime"`       // 添加时间
}

type DeviceIdLimit struct {
	Id       string    `bson:"_id" json:"id"`            // 设备码
	Operator string    `json:"operator" bson:"operator"` // 操作人
	Ctime    time.Time `bson:"ctime" json:"ctime"`       // 添加时间
}

// GetIpWhiteList ip白名单列表
func GetIpWhiteList() []IPwhite {
	var list []IPwhite
	ListByQ(IPwhites, bson.M{}, &list)
	return list
}

// GetPhoneBlackList 充值黑名单
func GetPhoneBlackList() []RechargeLimit {
	var list []RechargeLimit
	ListByQ(PayBlackLists, bson.M{}, &list)
	return list
}

// GetPhoneBlackList 充值黑名单
func GetDeviceIdBlackList() []DeviceIdLimit {
	var list []DeviceIdLimit
	ListByQ(EquipmentBlacklists, bson.M{}, &list)
	return list
}

// 大厅轮播图
type Banner struct {
	Id      string    `json:"-" bson:"_id"`
	SortId  int       `json:"sortId" bson:"sort_id"`   // 排序ID
	ImgName string    `json:"imgName" bson:"img_name"` // 图片名
	Rtype   int       `json:"rtype" bson:"rtype"`      // 跳转类型
	Url     string    `json:"url" bson:"url"`          // 图片地址
	Link    string    `json:"link" bson:"a_url"`       // 跳转链接
	Ctime   time.Time `json:"-" bson:"ctime"`          // 创建时间
}

// GetHallBannerList 大厅轮播图
func GetHallBannerList() []Banner {
	var list []Banner
	ListByQ(Banners, bson.M{}, &list)
	return list
}

// bonus中心
type BonusBanner struct {
	Id      string    `json:"-" bson:"_id"`
	SortId  int       `json:"sortId" bson:"sort_id"`   // 排序ID
	ImgName string    `json:"imgName" bson:"img_name"` // 图片名
	Rtype   int       `json:"rtype" bson:"rtype"`      // 跳转类型
	Url     string    `json:"url" bson:"url"`          // 图片地址
	Link    string    `json:"link" bson:"a_url"`       // 跳转链接
	Ctime   time.Time `json:"-" bson:"ctime"`          // 创建时间
}

// BonusBanner bonus中心
func GetBonusBannerList() []BonusBanner {
	var list []BonusBanner
	ListByQ(BonusBanners, bson.M{}, &list)
	return list
}

// GetServerWhiteList ip白名单列表
func GetServerWhiteList() []IPwhite {
	var list []IPwhite
	ListByQ(ServerWhites, bson.M{}, &list)
	return list
}

// 赠送可提现金
type GiveWithdrawCash struct {
	Userid string `json:"userid" bson:"_id"` // 玩家id
	Cash   int64  `json:"cash" bson:"cash"`  // 添加的可提现金
}

// 修改客服邮箱、tg、WhatsApp等
type ModifyCustomer struct {
	Id       string    `json:"id" bson:"_id"`
	Mail     string    `json:"mail" bson:"mail"`
	Telegram string    `json:"telegram" bson:"telegram"`
	WhatsApp string    `json:"whatsApp" bson:"whats_app"`
	Utime    time.Time `json:"utime" bson:"utime"`
}

// GetCustomerAddress 客服地址
func GetCustomerAddressList() []ModifyCustomer {
	var list []ModifyCustomer
	ListByQ(CustomerAdresses, bson.M{}, &list)
	return list
}

// 修改分享活动配置
type ModifyShare struct {
	Id        string    `bson:"_id" json:"id"`
	Youtobe   string    `json:"youtobe" bson:"youtobe"`     // youtobe
	Ins       string    `json:"ins" bson:"ins"`             // ins
	Facebook  string    `json:"facebook" bson:"facebook"`   // facebook
	Telegram  string    `json:"telegram" bson:"telegram"`   // Telegram
	CashPrize string    `json:"cashPrize" bson:"cashPrize"` // 显示奖金券金额
	Utime     time.Time `json:"utime" bson:"utime"`         // 创建时间
	WhatsApp  string    `json:"whatsapp" bson:"whatsapp"`   // whatsApp
	X         string    `json:"x" bson:"x"`                 // x(Twitter)
}

// GetModifyShareWayList 分享配置
func GetModifyShareWayList() []ModifyShare {
	var list []ModifyShare
	ListByQ(ShareWays, bson.M{}, &list)
	return list
}

// 自定义头像
type UserCustomPhoto struct {
	Id     string `bson:"_id" json:"id"`        //
	Result int    `json:"result" bson:"result"` // 审核结果
}

type ModifyNewbieStock struct {
	Gtype int32 `json:"gtype" bson:"gtype"` // 新手库存
	Stock int64 `json:"stock" bson:"stock"` // 值
}

type ModifyGameStock struct {
	Gtype int32 `json:"gtype" bson:"gtype"` // 游戏类型
	Stock int64 `json:"stock" bson:"stock"` // 游戏库存
}
