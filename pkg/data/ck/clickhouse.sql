-- optimize table game.col_user final;
-- optimize table game.col_user_finance final;
-- optimize table game.col_trade_record final;
-- optimize table game.col_withdraw_record final;
-- optimize table game.col_withdraw_record_transfer_order_detail final;
-- optimize table game.col_log_water final;
-- optimize table game.col_detail final;
-- optimize table game.col_log_login final;

create database game;

create table game.col_user (
	userid String comment '用户id',
	ctime DateTime64(3) comment '注册时间',
	ver Int64 comment 'version'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(ctime)
primary key userid
order by (userid)
SETTINGS index_granularity = 8192 -- 每8192行才生成一条索引 
comment '用户表';

create table game.col_user_finance (
	userid String comment '用户id',
	ctime DateTime64(3) comment '注册时间',
	ver Int64 comment 'version'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(ctime)
primary key userid
order by (userid)
SETTINGS index_granularity = 8192 -- 每8192行才生成一条索引 
comment '用户账户金额表';

create table game.col_trade_record (
	order_id String comment '订单ID',
	userid String comment '用户id',
	ctime DateTime64(3) comment '订单时间',
	ver Int64 comment 'version'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(ctime)
primary key (userid, order_id)
order by (userid, order_id)
SETTINGS index_granularity = 8192 -- 每8192行才生成一条索引 
comment '交易订单';

create table game.col_withdraw_record (
	order_id String comment '订单ID',
	userid String comment '用户id',
	ctime DateTime64(3) comment '订单时间',
	ver Int64 comment 'version'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(ctime)
primary key (userid, order_id)
order by (userid, order_id)
SETTINGS index_granularity = 8192 -- 每8192行才生成一条索引 
comment '提现订单';

create table game.col_withdraw_record_transfer_order_detail (
	order_id String comment '订单ID',
	userid String comment '用户id',
	channel_id UInt32 comment '用户id',
	ctime Int64 comment '订单时间',
	ver Int64 comment 'version'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(toDateTime(ctime))
primary key (order_id, userid, channel_id)
order by (order_id, userid, channel_id)
SETTINGS index_granularity = 8192 -- 每8192行才生成一条索引 
comment '提现订单转单详情';

create table game.col_log_water (
	id String comment 'id',
	userid String comment '账户ID',
	ctime DateTime64(3) comment '订单时间',
	ver Int64 comment 'version'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(ctime)
primary key (userid)
order by (userid, id)
SETTINGS index_granularity = 8192 -- 每8192行才生成一条索引 
comment '流水日志';

create table game.col_detail (
	id String comment 'id',
	userid String comment '用户ID,拆players',
	begin_time Int64 comment '开始时间',
	ver Int64 comment 'version'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(toDateTime(begin_time))
primary key (id, userid)
order by (id, userid, begin_time)
SETTINGS index_granularity = 8192 -- 每8192行才生成一条索引 
comment '对局详情表';

create table game.col_log_login (
	userid String comment '用户ID',
	login_time DateTime64(3) comment 'login Time',
	ver Int64 comment 'version'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(login_time)
primary key (userid, login_time)
order by (userid, login_time)
SETTINGS index_granularity = 8192 -- 每8192行才生成一条索引 
comment '登录日志';

create table game.col_nsq_log_external_bet (
	id String comment 'id',
	user_id String comment '用户ID',
	round_id String comment 'round_id',
	ctime Int64 comment 'ctime',
	ver Int64 comment 'version'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(toDateTime(ctime))
primary key (user_id, round_id, id)
order by (user_id, round_id, id)
SETTINGS index_granularity = 8192 -- 每8192行才生成一条索引 
comment '外接下注';

create table game.col_nsq_log_external_reward (
	id String comment 'id',
	user_id String comment '用户ID',
	round_id String comment 'round_id',
	ctime Int64 comment 'ctime',
	ver Int64 comment 'version'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(toDateTime(ctime))
primary key (user_id, round_id, id)
order by (user_id, round_id, id)
SETTINGS index_granularity = 8192 -- 每8192行才生成一条索引 
comment '外接返奖';

create table game.col_nsq_log_external_cancel (
	id String comment 'id',
	user_id String comment '用户ID',
	round_id String comment 'round_id',
	ctime Int64 comment 'ctime',
	ver Int64 comment 'version'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(toDateTime(ctime))
primary key (user_id, round_id, id)
order by (user_id, round_id, id)
SETTINGS index_granularity = 8192 -- 每8192行才生成一条索引 
comment '外接取消';

create table game.col_activity_betrank_prize (
	id String comment 'id',
	userid String comment '用户ID',
	rank_type Int32 comment '排行榜类型: 1日榜 2周榜 3月榜',
	etime Int64 comment 'etime',
	ver Int64 comment 'version'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(toDateTime(etime))
primary key (rank_type, etime, userid, id)
order by (rank_type, etime, userid, id)
SETTINGS index_granularity = 8192 -- 每8192行才生成一条索引 
comment '打码排行榜活动';

create table game.col_activity_turn_draw_log (
	id String comment 'id',
	turn_stime Int64 comment 'turn_stime',
	userid String comment '用户ID',
	ver Int64 comment 'version'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(toDateTime(turn_stime))
primary key (turn_stime, userid, id)
order by (turn_stime, userid, id)
SETTINGS index_granularity = 8192
comment '转盘活动抽奖记录';

create table game.col_activity_turn_prize_log (
	id String comment 'id',
	turn_stime Int64 comment 'turn_stime',
	userid String comment '用户ID',
	ver Int64 comment 'version'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(toDateTime(turn_stime))
primary key (turn_stime, userid, id)
order by (turn_stime, userid, id)
SETTINGS index_granularity = 8192
comment '转盘活动领奖记录';

create table game.col_launch_invite_log (
	id String comment 'id',
	userid String comment '用户ID',
	ver Int64 comment 'version',
	ctime Int64 comment '点击邀请时间戳毫秒'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(toDateTime(ctime/1000))
primary key (ctime, userid, id)
order by (ctime, userid, id)
SETTINGS index_granularity = 8192
comment '发起邀请日志';


create table game.col_activity_share_income_record (
	ver Int64 comment 'version',
	id String comment 'id',
	userid String comment '用户ID',
	super_id String comment '上级id',
	lv Int32 comment '属于几级代理',
	team_lv Int32 comment '获得奖励的上级团队等级',
	itype Int32 comment '1.打码奖励,2人数人头奖励,3.累计任务人头奖励,4.受邀者人头奖励',
	amount Int64 comment '返佣奖励金额(毫:1分=10厘=100毫)',
	amount_type Int32 comment '返佣奖励类型: 1:bonus,2:cash,3:withdrawable',
	water_id String comment '打码对局id/人头订单id',
	gtype Int32 comment '打码游戏类型',
	bets Int64 comment '打码量',
	idate String comment '日期字符串 2025-01-01',
	ctime Int64 comment '创建时间戳毫秒'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(toDateTime(ctime/1000))
primary key (ctime, super_id, itype, idate, id)
order by (ctime, super_id, itype, idate, id)
SETTINGS index_granularity = 4096
comment '代理打码/人头奖励记录表';

create table game.col_activity_share_income_record_tack_log (
	ver Int64 comment 'version',
	id String comment 'id: userid-date-itype',
	super_id String comment '用户ID',
	amount Int64 comment '领取奖励金额(分)',
	amount_type Int32 comment '返佣奖励类型: 1:bonus,2:cash,3:withdrawable',
	idate String comment '日期字符串 2025-01-01',
	itype Int32 comment '1.打码奖励,2人数人头奖励,3.累计任务人头奖励,4.受邀者人头奖励',
	ctime Int64 comment '创建时间戳毫秒'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(toDateTime(ctime/1000))
primary key (super_id, idate, itype, id)
order by (super_id, idate, itype, id)
SETTINGS index_granularity = 4096
comment '代理打码/人头奖励领取记录表';

create table game.col_log_vbdiamond (
	id String comment 'id',
	userid String comment '账户ID',
	reason Int32 comment '账变原因',
	vb Int64 comment '账变',
	ctime Int64 comment '时间(秒)',
	ver Int64 comment 'version'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(toDateTime(ctime))
primary key (ctime, userid, reason, id)
order by (ctime, userid, reason, id)
SETTINGS index_granularity = 8192
comment 'vbbank变化日志';

create table game.col_activity_volatility_subsidy (
	ver Int64 comment 'version',
	id String comment 'id',
	userid String comment '用户ID',
	sdate String comment '日期字符串 2025-01-01',
	subsidy Int64 comment '补贴金额',
	subsidy_type Int32 comment '补贴金额类型: 1:bonus,2:cash,3:withdrawable',
	regist_area Int32 comment 'abc类',
	first_time Int64 comment '首次领取时间戳毫秒',
	ctime Int64 comment '创建时间戳毫秒'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(toDateTime(ctime/1000))
primary key (ctime, userid, id)
order by (ctime, userid, id)
SETTINGS index_granularity = 4096
comment '波动返水记录';

create table game.col_log_online_users (
	id Int64 comment 'id',
	ctime Int64 comment '时间(秒)',
	ver Int64 comment 'version',
	day_time UInt32 comment '年月日 20250422',
	minute_time UInt32 comment '时分 1611',
	users UInt32,
	pay_users UInt32,
	online_users UInt32,
	online_pay_users UInt32,
	offline_users UInt32,
	offline_pay_users UInt32
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(toDateTime(ctime))
primary key (day_time, minute_time)
order by (day_time, minute_time)
SETTINGS index_granularity = 8192
comment '每分钟在线人数日志';

create table game.col_log_outdiamond (
	id String comment 'id',
	userid String comment '账户ID',
	reason Int32 comment '账变原因',
	out_cash Int64 comment '账变',
	ctime Int64 comment '时间(秒)',
	ver Int64 comment 'version'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(toDateTime(ctime))
primary key (ctime, userid, reason, id)
order by (ctime, userid, reason, id)
SETTINGS index_granularity = 8192
comment '可提现金变化日志';

create table game.col_user_coupon (
	id String comment 'id',
	user_id String comment '用户id',
	coupon_id String comment '优惠券id',
	is_auto Bool comment '是否自动',
	amount Int64 comment '优惠金额',
	min_recharge Int64 comment '最低充值',
	status Bool comment '是否使用',
	over_time Int64 comment '过期时间',
	ctime Int64 comment '创建时间',
	utime Int64 comment '使用时间',
	ver Int64 comment 'version'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(toDateTime(ctime))
primary key (coupon_id, user_id, id)
order by (coupon_id, user_id, id)
SETTINGS index_granularity = 8192
comment '用户优惠券';

create table game.col_customer_chat_sessions (
	id Int64 comment 'id',
	id2 String comment '对话id2',
	userid String comment '玩家id',
	customer String comment '分配客服id',
	score Int32 comment '玩家评分1-5, 0未评分',
	ctime Int64 comment '对话创建时间毫秒',
	etime Int64 comment '对话结束时间毫秒',
	seat_time Int64 comment '对话分配时间毫秒',
	read_time Int64 comment '首次已读时间毫秒',
	reply_time Int64 comment '首次回复时间毫秒',
	score_time Int64 comment '评分时间毫秒',
	question_type Int64 comment '问题类型: 1.充值 2.提现 3.故障或疑问 4.其他',
	resolved Bool comment '解决状态',
	remark String comment '备注',
	ver Int64 comment 'version'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(toDateTime(ctime))
primary key (userid, id)
order by (userid, id)
SETTINGS index_granularity = 8192
comment '客服聊天会话信息';

create table game.col_customer_chat_messages (
	id Int64 comment '消息id',
	id2 String comment '消息id2',
	session_id Int64 comment '对话id',
	session_id2 String comment '对话id2',
	userid String comment '玩家id',
	reply Bool comment '是否是客服回复',
	sender String comment '发送者userid',
	ctype Int32 comment '消息内容类型: 0.文本 1.图片 2.视频',
	content String comment '消息内容',
	`read` Bool comment '是否已读',
	ctime Int64 comment '创建时间毫秒',
	rtime Int64 comment '读消息时间毫秒',
	reader String comment '接收者userid',
	`filename` String comment 'ctype=1/2时文件名',
	ver Int64 comment 'version'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(toDateTime(ctime))
primary key (userid, session_id, id)
order by (userid, session_id, id)
SETTINGS index_granularity = 8192
comment '客服聊天记录';

create table game.col_customer_seat_log (
	id Int64 comment 'id',
	customer String comment '客服id',
	complete Bool comment '是否包含打开/关闭完整记录',
	open_time Int64 comment '坐席打开时间毫秒',
	close_time Int64 comment '坐席关闭时间毫秒',
	ver Int64 comment 'version'
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(toDateTime(open_time))
primary key (customer, id)
order by (customer, id)
SETTINGS index_granularity = 8192
comment '客服坐席操作记录';
