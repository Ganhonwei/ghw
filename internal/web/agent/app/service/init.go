package service

var (
	UserService       *userService // 用户服务
	RoleService       *roleService // 角色服务
	SystemService     *systemService
	ActionService     *actionService     // 系统动态
	StatisticsService *statisticsService // 统计管理
	PayService        *payService        // 支付服务
	ChannelService    *channelService    // 渠道管理
	PlayerService     *playerService     // 玩家
)

func initService() {
	UserService = &userService{}
	RoleService = &roleService{}
	SystemService = &systemService{}
	ActionService = &actionService{}
	PayService = &payService{}
	StatisticsService = &statisticsService{}
	ChannelService = &channelService{}
	PlayerService = &playerService{}
}
