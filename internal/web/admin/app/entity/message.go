package entity

import "time"

type ChatLog struct {
	Receiver string    `bson:"receiver" json:"receiver"` //接收者userid
	Content  string    `bson:"content" json:"content"`   //内容
	Title    string    `bson:"title" json:"title"`       //标题
	Ctime    time.Time `bson:"ctime" json:"ctime"`       //创建时间
	Sender   string    `bson:"sender" json:"sender"`     //发送者userid
	Name     string    `bson:"name" json:"name"`         //发送者姓名
}

// 后台聊天日志
type FeedBackLog struct {
	Userid         string    `bson:"_id" json:"id"`                   //id
	Nickname       string    `bson:"-"`                               // 用户昵称
	BundleId       string    `bson:"-"`                               // 渠道
	VipLv          int64     `bson:"-"`                               // vip等级
	ChargeType     string    `bson:"-"`                               // 玩家类型
	OnlineStatus   bool      `bson:"-"`                               // 在线状态
	ReplyStatus    int       `json:"replyStatus" bson:"reply_status"` // 沟通状态 1:未查看；2:未回复； 3：已回复
	LastTime       time.Time `bson:"-"`                               // 最后消息时间
	LastMsg        string    `bson:"-"`                               // 最近消息
	Utime          time.Time `json:"utime" bson:"utime"`              // 最后操作时间
	AssignUser     string    `json:"assignUser" bson:"assign_user"`   // 分配的客服
	AssignUserName string    `bson:"-"`                               // 分配的客服名称
	Logs           []ChatLog `bson:"log" json:"log"`                  // 记录
}

// 公告
type Notice struct {
	Id       string    `bson:"_id"`
	Rtype    int       `bson:"rtype"`    //0:跑马灯 1:系统公告;2:停服公告
	Language int       `bson:"language"` //1:英语
	Num      int       `bson:"num"`      //发送频率(分)
	Del      int       `bson:"del"`      //是否移除
	Content  string    `bson:"content"`  //广播内容
	Stime    time.Time `bson:"stime"`    //开始时间
	Etime    time.Time `bson:"etime"`    //结束时间
	Ctime    time.Time `bson:"ctime"`    //创建时间
}

type Banner struct {
	Id        string    `bson:"_id"`
	Rtype     int       `bson:"rtype"`                   // 跳转类型
	RtypeName string    `bson:"-"`                       // 跳转类型
	SortId    int       `json:"sortId" bson:"sort_id"`   // 排序ID
	ImgName   string    `json:"imgName" bson:"img_name"` // 图片名
	Url       string    `json:"url" bson:"url"`          // 图片地址
	AUrl      string    `json:"aUrl" bson:"a_url"`       // 跳转地址
	Ctime     time.Time `bson:"ctime"`                   // 创建时间
}

type BonusBanner struct {
	Id        string    `bson:"_id"`
	Rtype     int       `bson:"rtype"`                   // 跳转类型
	RtypeName string    `bson:"-"`                       // 跳转类型
	SortId    int       `json:"sortId" bson:"sort_id"`   // 排序ID
	ImgName   string    `json:"imgName" bson:"img_name"` // 图片名
	Url       string    `json:"url" bson:"url"`          // 图片地址
	AUrl      string    `json:"aUrl" bson:"a_url"`       // 跳转地址
	Ctime     time.Time `bson:"ctime"`                   // 创建时间
}

// 修改客服邮箱、tg、WhatsApp等
type ModifyCustomer struct {
	Id       string    `bson:"_id" json:"id"`
	Mail     string    `json:"mail" bson:"mail"`          // 邮箱
	Telegram string    `json:"telegram" bson:"telegram"`  // Telegram
	WhatsApp string    `json:"whatsApp" bson:"whats_app"` // WhatsApp
	AreaCode string    `bson:"-"`                         // 区号
	Utime    time.Time `json:"utime" bson:"utime"`        // 创建时间
}

// 修改客服功能配置
type CustomerServiceConfig struct {
	Id      string    `bson:"_id" json:"id"`           // id
	Switch  int       `json:"switch" bson:"switch"`    // 设置开关 0：不分单；1：顺序分单
	Users   []string  `json:"users" bson:"users"`      // 分单用户 10:00-19:30 后台用户id
	Users2  []string  `json:"users2" bson:"users2"`    // 分单用户 19:30-02:30 后台用户id
	Users3  []string  `json:"users3" bson:"users3"`    // 分单用户 02:30-10:00 后台用户id
	IsReset bool      `json:"isReset" bson:"is_reset"` // 重置排序
	Ctime   time.Time `bson:"ctime"`                   // 创建时间
}

// 客服消息分配记录日志
type ChatAssignLog struct {
	Id      string    `bson:"_id" json:"id"`           // id
	Userid  string    `json:"userid" bson:"userid"`    // 玩家ID
	CUserid string    `json:"cUserid" bson:"c_userid"` // 客服ID
	CSUsers []string  `json:"csUsers" bson:"cs_users"` // 可分配客服id
	Ctime   time.Time `bson:"ctime"`                   // 创建时间
}

// CdKey配置
type CdKeyConfig struct {
	Key            string    `json:"key" bson:"_id"`                        // CDKEY
	Remark         string    `json:"remark" bson:"remark"`                  // 功能备注
	Diamond        int64     `bson:"diamond" json:"diamond"`                // 彩金
	GiveDiamond    int64     `json:"giveDiamond" bson:"give_diamond"`       // 赠送彩金
	OutDiamond     int64     `json:"outDiamond" bson:"out_diamond"`         // 可提现彩金
	Bouns          int64     `json:"bouns" bson:"bouns"`                    // 罐子BOUNS
	UserId         string    `json:"userId" bson:"user_id"`                 // 指定用户ID
	OperatorId     string    `json:"operatorId" bson:"operator_id"`         // 操作人ID
	OperatorName   string    `json:"operatorName" bson:"operator_name"`     // 操作人名称
	ReceivePeoples int32     `json:"receivePeoples" bson:"receive_peoples"` // 领取人数
	Ctime          time.Time `bson:"ctime"`                                 // 创建时间
	Etime          time.Time `bson:"etime"`                                 // 结束时间
	Utime          time.Time `bson:"utime"`                                 // 修改时间
	FDiamond       float64   `bson:"-"`                                     // 彩金
	FGiveDiamond   float64   `bson:"-"`                                     // 赠送彩金
	FOutDiamond    float64   `bson:"-"`                                     // 可提现彩金
	FBouns         float64   `bson:"-"`                                     // 罐子BOUNS
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

// 用户自定义头像上传记录
type UploadUserHead struct {
	Id       string `bson:"_id" json:"id"`           // id
	UserId   string `json:"userId" bson:"user_id"`   // 用户ID
	IsAudit  int64  `json:"isAudit" bson:"is_audit"` // 审核状态 0:待审核 1：通过；2：拒绝
	Url      string `json:"url" bson:"url"`          // 头像地址
	Auditor  string `json:"auditor" bson:"auditor"`  // 审核人
	ATime    int64  `json:"aTime" bson:"a_time"`     // 审核时间
	Ctime    int64  `json:"ctime" bson:"ctime"`      // 上传时间(毫秒时间戳)
	Nickname string `json:"nickname" bson:"-"`       // 用户昵称
	BundleId string `json:"bundleId" bson:"-"`       // 渠道
	S_ATime  string `json:"sATime" bson:"-"`
	S_Ctime  string `json:"sCtime" bson:"-"`
}

// 自定义头像
type UserCustomPhoto struct {
	Id     string `bson:"_id" json:"id"`        //
	Result int    `json:"result" bson:"result"` // 审核结果
}
