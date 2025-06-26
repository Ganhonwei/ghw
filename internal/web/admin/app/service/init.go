package service

var (
	UserService   *userService   // 用户服务
	RoleService   *roleService   // 角色服务
	MailService   *mailService   // 邮件服务
	ActionService *actionService // 系统动态
	SystemService *systemService
	//---
	PlayerService *playerService // 玩家管理
	LoggerService *loggerService // 日志管理
	AgencyService *agencyService // 代理管理
	ChartsService *chartsService // 统计图表

	// CC Add
	PayService         *payService         // 支付服务
	GameService        *gameService        // 游戏
	MessageService     *messageService     // 消息
	StatisticsService  *statisticsService  // 统计管理
	Statistics2Service *statistics2Service // 统计管理
	GameStatsService   *gameStatsService   // 游戏统计
	ActivityService    *activityService    // 活动管理
	ChannelService     *channelService     // 渠道管理
	UploadService      *uploadService      // 上传
)

func initService() {
	UserService = &userService{}
	RoleService = &roleService{}
	MailService = &mailService{}
	ActionService = &actionService{}
	SystemService = &systemService{}
	PlayerService = &playerService{}
	LoggerService = &loggerService{}
	AgencyService = &agencyService{}
	ChartsService = &chartsService{}

	PayService = &payService{}
	GameService = &gameService{}
	MessageService = &messageService{}
	StatisticsService = &statisticsService{}
	Statistics2Service = &statistics2Service{}
	GameStatsService = &gameStatsService{}
	ActivityService = &activityService{}
	ChannelService = &channelService{}
	UploadService = &uploadService{}
}

var GtypeNameMap = map[int]string{
	1:      "TP",
	2:      "DRAGON TIGER",
	3:      "7UPDOWN",
	4:      "RUMMY",
	5:      "AK47",
	6:      "JOKER",
	7:      "CRASH",
	8:      "ANDARBAHAR",
	9:      "彩票",
	10:     "飞机",
	11:     "红黑大战",
	12:     "RUMMY双人",
	13:     "TP2",
	14:     "MINES",
	15:     "FORTUNE_GEMS2",
	16:     "FORTUNE_GEMS",
	600101: "Ganesha Fortune",
	600002: "Lucky Neko",
	600073: "Ganesha Gold",
	600022: "Fortune Ox",
	600039: "Speed Winner",
	600120: "Treasures of Aztec",
	600054: "Wild Bounty Showdown",
	600104: "Rise of Apollo",
	600037: "Fortune Tiger",
	600028: "Destiny of Sun & Moon",
	600041: "Legend of Perseus",
	600098: "CaiShen Wins",
	600025: "Songkran Splash",
	600093: "Asgardian Rising",
	600004: "Jurassic Kingdom",
	600108: "Wild Bandito",
	600009: "Supermarket Spree",
	600012: "Cocktail Nights",
	600119: "Galactic Gems",
	600110: "Ways of the Qilin",
	600029: "Honey Trap of Diao Chan",
	600081: "Dragon Hatch",
	600086: "Leprechaun Riches",
	600099: "Egypt's Book of Mystery",
	600102: "Dreams of Macau",
	600117: "Queen of Bounty",
	600103: "Candy Bonanza",
	600607: "Auto-Roulette",
	600572: "Lightning Blackjack",
	600526: "Lightning Roulette",
	600513: "Super Sic Bo",
	600594: "Dragon Tiger",
	600583: "Dream Catcher",
	600642: "Fan Tan",
	600536: "Golden Wealth Baccarat",
	600631: "Bac Bo",

	600276: "Fortune Gems",
	600369: "Fortune Gems 2",
	600402: "Fortune Gems 3",
	600332: "Money Coming",
	602287: "Crazy Time",
	600312: "Jackpot Fishing",
	600247: "Crazy777",
	602687: "Money Coming Expand Bets",
	600326: "Super Ace",
	604137: "Lightning Roulette",
	602344: "Funky Time",
	600339: "Ocean King Jackpot",
	604278: "Fortune Roulette",
	// 600101: "Ganesha Fortune",
	// 600009: "Supermarket Spree",
	600325: "Charge Buffalo",
	// 600120: "Treasures of Aztec",
	600374: "Fortune Dragon",
	600019: "Fortune Rabbit",
	604266: "Lucky Jaguar",
	600076: "Double Fortune",
	600761: "3 Coin Treasures",
	// 600583: "Dream Catcher",
	600561: "Super Andar Bahar",
	600766: "Crazy Hunter",
	600362: "Boxing King",

	604093: "Crazy Pachinko",

	606132: "Coin Tree",
	600082: "Vampire's Charm",
	600116: "Wild Fireworks",
	600322: "Fortune Monkey",
	600296: "Jungle King",
	606138: "Fruity Wheel",
	600399: "Jackpot Joker",
	600307: "Mega Ace",
	600302: "Fa Fa Fa",
	201001: "Golden Bank",
	600318: "World Cup",
	600100: "Mahjong Ways 2",
	// 600004: "Jurassic Kingdom",
	606141: "Go For Champion",
	600271: "Golden Empire",
	600218: "Ali Baba",
	600092: "Mahjong Ways",
	600311: "Bubble Beauty",
	600095: "Fortune Mouse",
	606139: "Treasure Quest",
	606140: "Fortune Gems Scratch",
	600304: "SevenSevenSeven",
	600036: "Butterfly Blossom",
	600366: "Golden Joker",
	600071: "Jungle Delight",
	600080: "Captain's Bounty",
	600396: "Devil Fire 2",
	606123: "Golden Bank 2",
	606122: "Safari Mystery",
	600255: "Lucky Goldbricks",
	600013: "Guardians of Ice & Fire",
	600052: "Forge of Wealth",
	600278: "Book of Gold",
	600764: "Potion Wizard",
	600053: "Wild Coaster",
	600269: "Medusa",
	600308: "Samba",
	604297: "Lucky Doggy",
	600397: "Zeus",
	600301: "war of dragons",
}

var GtypeFactoryMap = map[int]string{
	1:      "BigWin",
	2:      "BigWin",
	3:      "BigWin",
	4:      "BigWin",
	5:      "BigWin",
	6:      "BigWin",
	7:      "BigWin",
	8:      "BigWin",
	9:      "BigWin",
	10:     "BigWin",
	11:     "BigWin",
	12:     "BigWin",
	13:     "BigWin",
	14:     "BigWin",
	15:     "BigWin",
	600101: "PG",
	600002: "PG",
	600073: "PG",
	600022: "PG",
	600039: "PG",
	600120: "PG",
	600054: "PG",
	600104: "PG",
	600037: "PG",
	600028: "PG",
	600041: "PG",
	600098: "PG",
	600025: "PG",
	600093: "PG",
	600004: "PG",
	600108: "PG",
	600009: "PG",
	600012: "PG",
	600119: "PG",
	600110: "PG",
	600029: "PG",
	600081: "PG",
	600086: "PG",
	600099: "PG",
	600102: "PG",
	600117: "PG",
	600103: "PG",
	600607: "Evolution",
	600572: "Evolution",
	600526: "--",
	600513: "Evolution",
	600594: "Evolution",
	600583: "Evolution",
	600642: "Evolution",
	600536: "Evolution",
	600631: "Evolution",

	600276: "JiLi",
	600369: "JiLi",
	600402: "JiLi",
	600332: "JiLi",
	602287: "Evolution",
	600312: "JiLi",
	600247: "JiLi",
	602687: "JiLi",
	600326: "JiLi",
	604137: "Evolution",
	602344: "Evolution",
	600339: "JiLi",
	604278: "JiLi",
	// 600101: "Ganesha Fortune",
	// 600009: "Supermarket Spree",
	600325: "JiLi",
	// 600120: "Treasures of Aztec",
	600374: "PG",
	600019: "PG",
	604266: "JiLi",
	600076: "PG",
	600761: "JiLi",
	// 600583: "Dream Catcher",
	600561: "Evolution",
	600766: "JiLi",
	600362: "JiLi",

	606132: "JiLi",
	600082: "PG",
	600116: "PG",
	600322: "JiLi",
	600296: "JiLi",
	606138: "JiLi",
	600399: "JiLi",
	600307: "JiLi",
	600302: "JiLi",
	201001: "Gbet-JiLi",
	600318: "JiLi",
	600100: "PG",
	// 600004: "PG",
	606141: "JiLi",
	600271: "JiLi",
	600218: "JiLi",
	600092: "PG",
	600311: "JiLi",
	600095: "PG",
	606139: "JiLi",
	606140: "JiLi",
	600304: "JiLi",
	600036: "PG",
	600366: "JiLi",
	600071: "PG",
	600080: "PG",
	600396: "JiLi",
	606123: "JiLi",
	606122: "JiLi",
	600255: "JiLi",
	600013: "PG",
	600052: "PG",
	600278: "JiLi",
	600764: "JiLi",
	600053: "PG",
	600269: "JiLi",
	600308: "JiLi",
	604297: "JiLi",
	600397: "JiLi",
	600301: "JiLi",
}

var (
	SlotsGameids = []int{
		600101, 600002, 600073, 600022, 600039, 600120, 600054, 600104, 600037, 600028, 600041, 600098, 600025, 600093, 600004, 600108, 600009, 600012, 600119, 600110, 600029, 600081, 600086, 600099, 600102, 600117, 600103,
		600276, 600369, 600402, 600332, 600247, 602687, 600326, 600325, 600374, 600019, 604266, 600076, 600761, 600362,
		600339, // 未知游戏

		606132, 600082, 600116, 600322, 600296, 606138, 600399, 600307, 600302, 201001, 600318, 600100, 600004, 606141, 600271, 600218, 600092, 600311, 600095, 606139, 606140, 600304, 600036, 600366, 600071, 600080, 600396, 606123, 606122, 600255, 600013, 600052, 600278, 600764, 600053, 600269, 600308, 604297, 600397, 600301,
		15,
	}
	ZrsxGameids = []int{
		600607, 600572, 600526, 600513, 600594, 600583, 600642, 600536, 600631,
		602287, 604137, 602344, 600561,
		604278, // 百人非真人
		600312, // 捕鱼
		600766, // 捕鱼
		604278, // 未知游戏
	}

	// pve百人/pvp对战/pve单机
	// pve百人就是玩家与平台对战的百人场，比如crash、比如真人里面的龙虎、sicbo这类
	PVEBaiRenGameids = []int{2, 3, 7, 8, 10, 11,
		600607, // 转盘
		600572, 600526, 600513, 600594, 600583, 600642, 600536, 600631,
		602287, 604137, 602344, 600561,
		604278,
	}
	// pvp对战就是tp、rummy这种
	PVPDuiZhanGameids = []int{1, 4, 5, 6, 12, 13}
	// pve单机就是slots这种
	PVEDanJiGameids = []int{
		9, 14, 15,
		// slots
		600101, 600002, 600073, 600022, 600039, 600120, 600054, 600104, 600037, 600028, 600041, 600098, 600025, 600093, 600004, 600108, 600009, 600012, 600119, 600110, 600029, 600081, 600086, 600099, 600102, 600117, 600103,
		600276, 600369, 600402, 600332, 600247, 602687, 600326, 600325, 600374, 600019, 604266, 600076, 600761, 600362,
		// 捕鱼
		600312, 600766,
		600339, // 未知游戏

		606132, 600082, 600116, 600322, 600296, 606138, 600399, 600307, 600302, 201001, 600318, 600100, 600004, 606141, 600271, 600218, 600092, 600311, 600095, 606139, 606140, 600304, 600036, 600366, 600071, 600080, 600396, 606123, 606122, 600255, 600013, 600052, 600278, 600764, 600053, 600269, 600308, 604297, 600397, 600301,
	}
)

// 流水日志类型
var (
	// reason: 0全部,1充值,2提现,3游戏返奖,4下注,5奖金福利,7撤销下注
	LTypeReason1 = []int32{57, 58, 59, 74, 80, 84, 86, 96, 119, 129, 139, 152}                                   // 1.充值
	LTypeReason2 = []int32{13, 63, 136, 137, 138}                                                                // 2.提现
	LTypeReason3 = []int32{45, 67, 68, 70, 71, 73, 88, 91, 95, 101, 107, 111, 113, 115, 117, 128, 131, 150, 154} // 3.游戏返奖
	LTypeReason4 = []int32{5, 75, 76, 79, 87, 90, 94, 100, 106, 110, 112, 114, 116, 127, 130, 148, 149, 153}     // 4.下注
	LTypeReason6 = []int32{69, 72, 89, 92, 109}                                                                  // 6.扣税,抽水
	LTypeReason7 = []int32{146, 147, 151}                                                                        // 7.撤销下注
	LTypeReason8 = []int32{83, 64, 65}                                                                           // 8.后台加彩金, 103 转为正常玩家
)
