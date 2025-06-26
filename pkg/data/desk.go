package data

import (
	"fmt"
	"math/rand"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// DeskData 房间基础数据
type DeskData struct {
	Rid      string `json:"rid"`       //房间ID
	Unique   string `json:"unique"`    //唯一ID
	DeskType int32  `json:"desk_type"` //房间类型
	Gtype    int32  `json:"gtype"`     //游戏类型
	Rtype    int32  `json:"rtype"`     //房间类型
	Dtype    int32  `json:"dtype"`     //桌子类型
	Ltype    int32  `json:"ltype"`     //彩票类型/房间等级类型
	Rname    string `json:"rname"`     //房间名字
	Count    uint32 `json:"count"`     //牌局人数限制
	Cost     uint32 `json:"cost"`      //抽佣百分比/创建消耗
	Ante     uint32 `json:"ante"`      //底分
	Vip      uint32 `json:"vip"`       //vip限制
	Chip     uint32 `json:"chip"`      //chip限制/进入房间限制
	//
	Deal  bool   `json:"deal"`  //房间是否可以上庄
	Carry uint32 `json:"carry"` //上庄携带限制
	Down  uint32 `json:"down"`  //下庄携带限制
	Top   uint32 `json:"top"`   //下庄最高携带限制
	//
	Sit   uint32 `json:"sit"`   //房间内坐下限制
	Ctime uint32 `json:"ctime"` //创建时间
	//
	Code    string `json:"code"`    //房间邀请码
	Cid     string `json:"cid"`     //房间创建人
	Expire  int64  `json:"expire"`  //牌局设定的过期时间
	Round   uint32 `json:"round"`   //牌局数
	Payment uint32 `json:"payment"` //付费方式1=AA or 0=房主支付
	//
	Minimum int64 `json:"minimum"` //房间最低限制  离场限制
	Maximum int64 `json:"maximum"` //房间最高限制  入场限制
	//
	Pub      bool   `json:"pub"`      //公开显示
	Mode     uint32 `json:"mode"`     //模式，0普通，1疯狂
	Multiple uint32 `json:"multiple"` //倍数，0低，1中，2高
	Game     Game
	RoomId   int32 `json:"room_id"` // 房间id
	//
	Gmode int32 // 0真金,1娱乐模式
	// Rounds int32 // 私人房总牌局数
}

// DeskRole 牌桌玩家数据
type DeskRole struct {
	//数据
	*User
	//进程ID
	Pid *actor.PID
	//离线状态
	Offline bool
	// 私人房离线超时踢出时间
	PrivOfflineTimeout int64
	//位置
	Seat uint32
	//彩金占比
	CashPro int32
	//比例
	Ratio float64
	//百人连续不下注轮数
	NoBetTimes int32
	//超时次数
	TimeoutCount int
	//drop次数
	DropCount int
}

// DeskSeat 牌桌位置数据
type DeskSeat struct {
	Userid               string     //玩家id
	Ready                bool       //是否准备
	Ready2               bool       //是否发牌动画播放完成
	Watch                bool       //是否观战
	BeDealer             uint32     //1抢庄,2不抢,0等待抢
	DealerN              uint32     //抢庄倍数
	Bet                  int64      //下注数量
	Card                 uint32     //定庄牌
	Cards                []uint32   //手牌
	WinRate              float64    //胜率
	ShowCards            []uint32   //亮牌
	ChangeCards          []uint32   //变牌
	ShowChangeCards      []uint32   //亮变牌
	SortCards            [][]uint32 //摆牌
	Finish               uint32     //最后finish的牌
	Sort                 bool       //轮到自己操作时是否sort了
	Winner               bool       //赢家
	Declare              bool       //声明牌
	Power                uint32     //牌力
	Vote                 uint32     //投票解散1同意,2反对
	Niu                  bool       //提交操作
	Voice                uint32     //投票语音房间1同意,2反对
	Core                 bool       //核心人机
	SeeAndPack           bool       //看后即弃
	Identity             string     //身份 a主机 b主机 c僚机 d僚机
	Timeout              bool       //是否超时
	Drop                 bool       //是否drop
	Again                uint32     //投票再来一局1同意,2反对
	ActTimes             int32      //本局操作轮数
	DrawedNeed           int        //rm控制概率生效时玩家已摸到需要的牌张数
	Actions              []int      //玩家操作记录
	Touchs               []int      //摸牌操作记录
	StakeRateYi          int32      //stake概率(亿)
	MaxMultiple          float64    //极限倍数
	SeeRate              int32      //对局中增加的看牌概率
	BiggerThanPlayer     bool       //人机牌比玩家大
	RobotType            int        //人机类型 1左,2中,3右
	RobotTypeChangeTimes int        // 人机类型切换次数
	LeaveRateInited      bool       // 换桌概率初始化
	LeaveRate            int32      // 换桌概率
	TryLeave             bool       // 本局判断过离开
	LeaveWin             bool       // 是否赢家
	SeatRounds           int32      // 该位置连续打局数
	SideShowBeRejected   bool       // 比牌被拒绝过
}

// DeskGame 私人局牌桌当局数据
type DeskGame struct {
	GameId     string            //对局号
	Round      uint32            //私人房间打牌局数
	Cards      []uint32          //没摸起的海底牌
	LaiCards   []uint32          //癞子牌
	QiCards    []uint32          //弃牌堆
	WildCard   uint32            //万能牌
	BetNum     int64             //当前局下注总数/总底池
	CashBets   int64             //彩金池
	CoinBets   int64             //奖励金池
	Dealer     string            //庄家
	DealerSeat uint32            //庄家的座位
	FakeSeats  map[uint32]string //假位置数据
	//
	VoiceSeat  uint32  //语音房间投票发起者座位号
	VoiceTime  int64   //语音房间结束投票时间
	BeginTime  int64   //对局开始时间
	FinaFactor float64 // 最终系数
	NowStock   int64   // 当前库存
}

// DeskAct 金花数据, TODO 按轮记数
type DeskAct struct {
	//当前底注
	ActAnte int64
	//当前操作位置
	ActSeat uint32
	//当前操作值
	ActState int32
	//当前跟注额
	ActCallNum int64
	//当前最小加注额
	ActRaiseNum int64
	//第几轮
	ActTimes int32
	////当前操作的底池
	//ActPot int
	////底池,可多个池
	// ActPots map[int]*DeskPot
	//位置操作
	ActSeats map[uint32]*ActStatus
	//位置充值次数
	ActRechargeTimes map[uint32]int
}

// ActStatus 位置操作状态
type ActStatus struct {
	See   bool  //是否看牌
	Alive bool  //是否还在
	Lose  bool  //是否输了
	Pack  bool  //是否弃牌
	Dirty bool  //比牌污染
	Bet   int64 //上次操作下注
	Drop  bool  //是否drop
	ZhaHu bool  //诈胡
	Chaal bool  //是否chaal或chaal*2
	//Allin bool //all in
	////本轮下注额
	ActNum int64
	//得分
	ActScore int64
	//库存
	Stock *ActStock
	// 操作记录
	ActHistory []int32
}

type ActStock struct {
	CashStock  int64 //彩金库存
	BonusStock int64 //奖励金库存
	CashMing   int64 //彩金明税
	BonusMing  int64 //奖励金明税
	CashAn     int64 //彩金暗税
	BonusAn    int64 //奖励金暗税
	CashGive   int64 //赠送金回收
	Robot      bool  //机器人跳过
}

////DeskPot 底池
//type DeskPot struct {
//	//参与底池位置
//	PotSeats []uint32
//	//底池金额
//	PotNum int64
//}

// DeskPriv 私人局牌桌当局数据
type DeskPriv struct {
	VoteSeat        uint32   //投票发起者座位号
	VoteTime        int64    //结束投票时间
	VoteStartTime   int64    //开始投票时间
	Votes           []uint32 // 解散投票记录
	LaunchVoteTimes uint32   // 发起投票次数
	//
	PrivPlayer map[string][3]string //参加私人局的用户[nickname,photo]
	PrivScore  map[string]int64     //私人局用户战绩积分
	Joins      map[string]uint32    //私人局用户参与次数
	// 胜利局数、失败局数、总成绩。总输赢
	PrivWins   map[string]uint32 // 玩家胜利局数
	PrivLoses  map[string]uint32 // 玩家失败局数
	Again      uint32            // 1是2否
	AgainSeat  uint32            // 再来一局发起者座位号
	AgainTime  int64             // 结束再来一局投票时间
	AgainVotes []uint32          // 再来一局投票记录
	AgainRound int32             // 再来一局次数
	// 7天对局回放

	// andarbahar
	*ABDeskPriv
}

// AB私人房
type ABDeskPriv struct {
	Joker           uint32           // 本回合joker牌
	AndarCards      []uint32         // Andar开的牌
	BaharCards      []uint32         // Bahar开牌
	Bets            map[string]int64 //userid:num, 玩家下注金额
	AndarBets       map[string]int64 //userid:num, 玩家andar下注金额
	BaharBets       map[string]int64 //userid:num, 玩家bahar下注金额
	ABScore         map[string]int64 // 输赢分
	ABRoundBetTimes map[string]int32 // round,user,seat = times 玩家当前下注回合位置下注次数
	// AndarSeatRoleBets  map[string]Currency // andar下注详细
	// BarharSeatRoleBets map[string]Currency // barhar下注详细
	CardRoundTime int64  // 本轮翻牌开始时间
	CardRound     uint32 // 本局翻牌轮数
	Winner        uint32 // 赢家:1andar 2bahar
}

// DeskFree 百人场牌桌当局数据
type DeskFree struct {
	CarryInit    int64                 //庄家初始的携带
	Carry        int64                 //庄家的携带
	DealerNum    uint32                //做庄次数
	BetTime      int                   //下注时长
	Dealers      map[string]int64      //上庄列表,userid: carry
	Cards        map[uint32][]uint32   //手牌
	Power        map[uint32]uint32     //牌力
	Bets         map[string]int64      //userid:num, 位置1玩家下注金额
	SeatBets     map[uint32]int64      //seat:num, 位置总下注金额
	SeatCashBets map[uint32]int64      //seat:num, 位置彩金总下注金额
	SeatBetLogs  map[string][]Currency //userid:bets, 位置下注记录
	//位置下注详细
	SeatRoleBets map[uint32]map[string]int64
	//结果 seat:num,seat=(1,2,3,4,5),倍数
	Multiple map[uint32]int64
	Score1   map[uint32]int64 //位置(1-5)输赢总量
	Score2   map[string]int64 //每个闲家输赢总量
	//位置(1-5)上每个玩家输赢
	Score3 map[uint32]map[string]int64
	//Trend 输赢趋势
	Trends []*FreeTrend
	//上局赢家
	Winers []*FreeWiner
	//结束下庄
	DealerDown bool
	//赠送金库存
	GiveStock int64
	//新手模式
	*FreeDeskNewbiew

	UserFactorMap map[string]float64 // 玩家系数
	//龙虎and7up
	*LHDeskFree
	LHHistory    []int32          // 输赢记录20条
	UPHistory    []*pb.UPHistory  // 输赢记录100条
	NewbiewRobot map[string]*User // 假人
	//退可提现彩金
	BackOutDiamond map[string]int64

	//crash
	*CRASHDeskFree

	//andarbahar
	*ABDeskFree
	ABWinnerSeat []uint32 // 1.andar 2.bahar  100条
	ABJoker      []uint32 // joker牌

	//彩票
	*LotteryDesk
	CPHistory    []uint32    // 历史记录
	CPJackpot    int64       // 奖池
	CPBigWinners []CPJackpot // 开奖记录

	// 红黑
	*RBDeskFree
	RBHistory []RBHistory // 红黑历史
}

type LHDeskFree struct {
	// lhd
	LHCards         []uint32                       // 开牌
	LHCheat         map[string]int32               // 防刷水局数
	LHSeatRoleBets  map[uint32]map[string]Currency // 位置下注详细
	LHRobotSeat     map[string]uint32              // 人机位置
	ScoreMap        map[string]Currency            // 输赢分
	OddsMap         map[uint32]int32               // 赔付倍数
	CritOddsMap     map[uint32]int32               // 暴击赔付倍数
	LHRobotBets     map[string]int64               // 人机下注
	LHObserver      []string                       // 观察者
	LHDStrategys    []StrategyInfo                 // 本局策略
	TriggerStrategy int                            // 生效的策略
	LHDetail        LHDetail                       // 龙虎详情
	UPDetail        UPDetail                       // 7up详情
}

type CRASHDeskFree struct {
	CrashBoom             bool                     //已经爆炸
	CrashJackpot          int64                    //奖池
	CrashBets             map[string]Currency      //位置0爆点下注
	CrashBets1            map[string]Currency      //位置1爆点下注
	CrashNextBets         map[string]Currency      //位置0下一轮爆点下注
	CrashNextBets1        map[string]Currency      //位置1下一轮爆点下注
	CRASHHistory          []int32                  //历史爆炸倍数
	CRASHMulpitle         int32                    //当前倍数
	CRASHRand             *rand.Rand               //本局随机数
	CRASHBoomMulpitle     int32                    //爆炸倍数
	CRASHTakeoffTime      time.Time                //起飞时间
	CRASHBoomTime         int64                    //爆炸时间(毫秒)
	CRASHLose             int64                    //已逃脱赔付
	CRASHRealStock        int64                    //房间彩金库存
	CRASHFactor           float64                  //最终系数
	CRASHCashStock        int64                    //变化彩金库存
	CRASHCoinStock        int64                    //变化奖励金库存
	CRASHMCashTax         int64                    //彩金明税
	CRASHMCoinTax         int64                    //奖励金明税
	CRASHACashTax         int64                    //彩金暗税
	CRASHACoinTax         int64                    //彩金暗税
	CRASHFirstPartIn      bool                     //首次参与
	CRASHDetail           Detail                   //对局详情
	CRASHLeadTime         int                      //爆炸后等待时长
	CRASHScoreMap         map[string]Currency      //位置0输赢分
	CRASHScoreMap1        map[string]Currency      //位置1输赢分
	CRASHTaxMap           map[string]int64         //赢分交税
	CRASHPlayerJackpot    map[string]int64         //玩家中奖分数
	CrashAutoLeave        map[string]int32         //位置0自动撤离玩家
	CrashAutoLeave1       map[string]int32         //位置1自动撤离玩家
	CRASHBack             map[string]int32         //位置0成功撤离玩家->撤离倍数
	CRASHBack1            map[string]int32         //位置1成功撤离玩家->撤离倍数
	CRASHControlMulpitle  int32                    //控制倍数(测试)
	CRASHObserver         []string                 //观察者
	CRASHStrategys        []StrategyInfo           //本局触发策略
	CRASHTriggerStrategys map[int]StrategyInfo     //本局生效策略
	BetRank               []*pb.CRASHBetRankUser   //前100名打码用户
	LastRecord            *pb.CrashLastRoundRecord //前100名打码用户
}

type StrategyInfo struct {
	Id     int   // 策略id
	Weight int   // 优先级
	Seat   []int // 可开位置

	// lhd
	Disturb bool // 心想事成：是否扰动
}

type ABDeskFree struct {
	ACards         []uint32                       // Andar开的牌
	BCards         []uint32                       // Bahar开牌
	DrawTime       int                            // 开牌时长
	ABSeatRoleBets map[uint32]map[string]Currency // 位置下注详细
	ABFixedPoker   map[int]uint32                 // 固定牌
	ABScoreMap     map[string]Currency            // 输赢分
	OddsMap        map[uint32]int                 // 赔付倍数
	Joker          uint32                         // joker牌
	Winner         uint32                         // 赢家
	SideWinner     uint32                         // 边注赢家
	CardType       int                            // 牌型id
	ABDetail       ABDetail                       // 对局详情
	ABStrategys    []StrategyInfo                 // 本局策略
	ABObserver     []string                       // 观察者
}

type LotteryDesk struct {
	LCards            []uint32                       // 开的牌
	LSeatRoleBets     map[uint32]map[string]Currency // 位置下注详细
	LScoreMap         map[string]Currency            // 输赢分
	CPRobotBets       map[string]int64               // 人机下注
	CPRobotSeat       map[string]uint32              // 人机位置
	CPStrategys       []StrategyInfo                 // 本局策略
	CPTriggerStrategy int                            // 生效的策略
	LDetail           CPDetail                       // 彩票详情
	CPObserver        []string                       // 观察者
}

type RBDeskFree struct {
	// 红黑
	RBCards           [][]uint32                     // 开牌
	RBSeatRoleBets    map[uint32]map[string]Currency // 位置下注详细
	RBRobotSeat       map[string]uint32              // 人机下注位置
	RBRobotBets       map[string]int64               // 人机下注
	RBScoreMap        map[string]Currency            // 输赢分
	RBObserver        []string                       // 观察者
	RBStrategys       []StrategyInfo                 // 本局策略
	RBTriggerStrategy int                            // 生效的策略
	RBDetail          RBDetail                       // 红黑详情
}

type RBHistory struct {
	CardType uint32 // 牌型
	Winner   uint32 // 赢的位置
}

// FreeTrend 输赢趋势, //true 赢 false 输
type FreeTrend struct {
	Seat2 bool //天
	Seat3 bool //地
	Seat4 bool //玄
	Seat5 bool //黄
}

// FreeWiner 上局赢家
type FreeWiner struct {
	Userid   string
	Nickname string
	Photo    string
	Coin     int64 //赢利数量
}

// 彩票大赢家
type CPJackpot struct {
	GameId   string   // 对局id
	Userid   string   // userid
	Nickname string   // 名称
	Photo    string   // 头像
	Bet      int64    // 下注
	Get      int64    // 获得奖励
	Draw     int64    // 中奖
	Ctime    int64    // 时间
	Cards    []uint32 // 牌值
	People   int      // 赢分人数
}

// DeskBase 房间牌桌数据
type DeskBase struct {
	//基础数据
	*DeskData
	//进程ID
	Pid *actor.PID
	//房间人数
	Number uint32
	//真人玩家
	RealUser []string
}

// 百人新手模式
type FreeDeskNewbiew struct {
	BetQeuee []NewbiewFreeBetQeuee // 新手模式下注队列
	// BetLastTime int64                 // 上次处理时间
}

// 百人新手模式人机下注队列
type NewbiewFreeBetQeuee struct {
	Robotid string // 人机id
	BetTime int64  // 下注时间
	Chip    int64  // 下注额
	Roomid  string // 房间id
	Seat    uint32 // 下注位置
}

// 局内充值
type DeskRecharge struct {
	Price uint32 //价格
	Give  uint32 //赠送
	Desc  string //描述
}

// =====================================支付接口============================================
// 判断能不能买
func (s *DeskRecharge) CanOrder(user *User, id string, ctype int32) bool {
	return true
}

// 充值之后
func (s *DeskRecharge) RechargeAfter(user *User, id string, ctype int32, save bool) (*pb.PaySuccessNtf, error) {
	glog.Infof("user %s recharge success, shop:%s", user.Userid, id)
	rsp := new(pb.PaySuccessNtf)
	rsp.Rtype = pb.GAMERECHARGE
	return rsp, nil
}

// 完善订单
func (s *DeskRecharge) BuildPayOrder(user *User, order *PayOrder, game bool) {
	order.Amount = s.Price
	order.Score = s.Price
	order.ShopName = s.Desc
	order.ShopType = GAME_RECHARGE

	order.OtherPresent = fmt.Sprintf("%d", s.Give)
	// if user.VBBank >= int64(s.Give) {
	// }
}

func (s *DeskRecharge) RechageLType() int32 {
	return int32(pb.LOG_TYPE80)
}
