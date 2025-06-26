package entity

// 手动发放优惠券
type ManualCoupon struct {
	Id                 string   `json:"id" bson:"_id"`
	SendSwitch         bool     `json:"sendSwitch" bson:"send_switch"`                 // 开关
	MinRecharge        int64    `json:"minRecharge" bson:"min_recharge"`               // 最低充值
	Discount           int64    `json:"discount" bson:"discount"`                      // 减免金额
	ValidityTime       int32    `json:"validityTime" bson:"validity_time"`             // 有效期
	Num                int32    `json:"num" bson:"num"`                                // 人均张数
	OrderAvgRange      []int64  `json:"orderAvgRange" bson:"order_avg_range"`          // 订单平均金额
	ProfitabilityRatio []int64  `json:"profitabilityRatio" bson:"profitability_ratio"` // 盈利比例
	RecharClassify     []int    `json:"recharClassify" bson:"rechar_classify"`         // 充值分类
	AcctType           []int    `json:"acctType" bson:"acct_type"`                     // 账户类型
	UserTag            []int    `json:"userTag" bson:"user_tag"`                       // 用户标签
	UserIds            []string `json:"userIds" bson:"user_ids"`                       // 用户ID
	SendTime           int64    `json:"sendTime" bson:"send_time"`                     // 发送时间
	UsedTime           int64    `json:"usedTime" bson:"used_time"`                     // 用完时间
	Ctime              int64    `json:"ctime" bson:"ctime"`                            // 创建时间
	SendCount          int64    `json:"sendCount" bson:"send_count"`                   // 发送数量
	SendUserCount      int64    `json:"sendUserCount" bson:"send_user_count"`          // 发送用户数量
	SeeCount           int64    `json:"seeCount" bson:"-"`                             // 看到数量
	SeeUserCount       int64    `json:"seeUserCount" bson:"see_user_count"`            // 查看用户数量
	UsedCount          int64    `json:"usedCount" bson:"used_count"`                   // 已使用数量
	UsedUserCount      int64    `json:"usedUserCount" bson:"used_user_count"`          // 已使用用户数量
	UsedRatio          string   `json:"usedRatio" bson:"-"`                            // 送达张数使用率
	UsedUserRatio      string   `json:"usedUserRatio" bson:"-"`                        // 送达用户使用率
	SeeRatio           string   `json:"seeRatio" bson:"-"`                             // 看到率
	SeeUserRatio       string   `json:"seeUserRatio" bson:"-"`                         // 看到用户率
	GiveRatio          string   `json:"giveRatio" bson:"-"`                            // 赠送率
	PayAvg             string   `json:"payAvg" bson:"-"`                               // 均单价
	ProfitabilityStr   string   `json:"profitabilityStr" bson:"-"`                     // 盈利比例
	RecharClassifyStr  string   `json:"recharClassifyStr" bson:"-"`                    // 充值分类
	AcctTypeStr        string   `json:"acctTypeStr" bson:"-"`                          // 账户类型
	UserTagStr         string   `json:"userTagStr" bson:"-"`                           // 用户标签
	SendTimeStr        string   `json:"sendTimeStr" bson:"-"`                          // 发送时间
	CtimeStr           string   `json:"ctimeStr" bson:"-"`                             // 创建时间
	UsedTimeStr        string   `json:"usedTimeStr" bson:"-"`                          // 用完时间
	UserIdStr          string   `json:"userIdStr" bson:"-"`                            // 用户ids
	Tag                int      `json:"Tag" bson:"-"`                                  // 用户标签
}
