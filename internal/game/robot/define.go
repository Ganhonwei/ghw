package robot

import (
	"goserver/gen/pb"
	"goserver/gen/tb"
	"goserver/pkg/game/config"
	"sync/atomic"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
)

var (
	manEnNames   = make([]string, 0)
	manInNames   = make([]string, 0)
	womanEnNames = make([]string, 0)
	womanInNames = make([]string, 0)
	manPhotos    = make([]string, 0)
	womanPhotos  = make([]string, 0)
)

// 延时消息队列
type DelayQuee struct {
	msg       interface{} // 消息
	sendTime  int64       // 延时时间
	delayCall func() bool // 回调
}

type DeskData struct {
	NowChip      uint32      //本轮剩余下注筹码
	NowSeat      uint32      //本轮位置
	nextTime     int64       //下次下注时间
	lastSendTime int64       //上次发包时间
	msgQuee      []DelayQuee //消息队列
	emojiQuee    []DelayQuee //表情队列

	seat uint32 //位置

	// 临时数据
	round        uint32      //玩牌局数
	sits         uint32      //尝试坐下次数
	bits         uint32      //尝试下注次数
	bitNum       uint32      //尝试下注数量
	timer        uint32      //在线时间
	ready        bool        //准备状态
	see          bool        //是否看牌
	alive        bool        //是否在游戏中
	cards        []uint32    //手牌
	ccards       []uint32    //变牌
	sortCards    [][]uint32  //sort牌
	wildCard     uint32      //万能牌
	qiCard       uint32      //弃牌堆
	qiCards      []uint32    //rm弃牌堆
	faceUpCard   []uint32    //rm弃牌堆出现过的牌,摸起后不移除
	preSeat      uint32      //上家
	state        int32       //状态
	seats        []uint32    //其他玩家座位
	core         bool        //核心人机
	seeAndPack   bool        //看后即弃
	identity     string      //身份
	chargeInGame bool        //本回合人机假装充值
	prevAction   int         // 记录上一个动作
	prevRAction  *pb.RAction // 记录上一个指定动作

	robotType       int32 //机器人类别
	count           int32 //参与人数
	cardType        int32 //牌型编号
	isMax           bool  //胜负关系
	afterRefusePack bool  //被拒后波胆

	seeingInOther atomic.Bool // 其他玩家回合看牌中(延迟消息)

	emoji         tb.EmojiRobotEmojiConfigRecord //人机表情
	jhstrategy    *pb.JHCoinRobotStrategyNtf     //tp人机策略
	jokerstrategy *pb.JOKERCoinRobotStrategyNtf  //joker人机策略
	ak47strategy  *pb.AK47CoinRobotStrategyNtf   //ak47人机策略
	lh            lhdRobot                       //lhd临时数据
	up            upRobot                        //7up临时数据
	crash         crashRobot                     //crash临时数据
	ab            abRobot                        //andarbahar临时数据
	cp            cpRobot                        //彩票临时数据
	plane         planeRobot                     //crash临时数据
}

// 龙虎斗数据
type lhdRobot struct {
	timer       int32                  //下注间隔
	waitTime    int32                  //等待时间
	state       int32                  //游戏状态
	maxBetCount int32                  //离开的下注轮数
	betCount    int32                  //当前的下注轮数
	lhstrategy  *pb.LHRobotStrategyNtf // lhd人机策略
}

// 7up数据
type upRobot struct {
	timer       int32                  //下注间隔
	waitTime    int32                  //等待时间
	state       int32                  //游戏状态
	maxBetCount int32                  //离开的下注轮数
	betCount    int32                  //当前的下注轮数
	lhstrategy  *pb.UPRobotStrategyNtf //7up人机策略
}

type crashRobot struct {
	timer        int32                     //下注间隔
	waitTime     int32                     //等待时间
	boom         bool                      //是否爆炸
	state        int32                     //游戏状态
	maxBetCount  int32                     //离开的下注轮数
	betCount     int32                     //当前的下注轮数
	back         bool                      //是否逃离
	cashstrategy *pb.CRASHRobotStrategyNtf //crash人机策略
}

type planeRobot struct {
	timer        int32                     //下注间隔
	waitTime     int32                     //等待时间
	boom         bool                      //是否爆炸
	state        int32                     //游戏状态
	maxBetCount  int32                     //离开的下注轮数
	betCount     int32                     //当前的下注轮数
	back         bool                      //是否逃离
	cashstrategy *pb.PLANERobotStrategyNtf //人机策略
}

// andarbahar数据
type abRobot struct {
	timer       int32                  //下注间隔
	waitTime    int32                  //等待时间
	state       int32                  //游戏状态
	maxBetCount int32                  //离开的下注轮数
	betCount    int32                  //当前的下注轮数
	nowChip     map[uint32]uint32      //本轮剩余下注筹码
	abstrategy  *pb.ABRobotStrategyNtf //ab人机策略
}

// 彩票数据
type cpRobot struct {
	timer       int32                  //下注间隔
	waitTime    int32                  //等待时间
	state       int32                  //游戏状态
	maxBetCount int32                  //离开的下注轮数
	betCount    int32                  //当前的下注轮数
	cpstrategy  *pb.CPRobotStrategyNtf //彩票人机策略
}

type RobotData struct {
	Rtype  int32
	Ltype  int32
	Gameid string
	Roomid string
	EnvBet int32
	Min    int32
	Max    int32
	Gtype  int32
	Emoji  []byte
}

// 召唤人机队列
type CallRobot struct {
	Msg      RobotData
	CallTime int64 // 进入房间时间
}

func InitRobotNames() {
	file, err := excelize.OpenFile("config/name.xlsx")
	if err != nil {
		panic(err)
	}
	var err1, err2, err3 error
	manEnNames, manInNames, err1 = config.GetRobotNameByExcel("man", file)
	if err1 != nil {
		panic(err1)
	}
	womanEnNames, womanInNames, err2 = config.GetRobotNameByExcel("woman", file)
	if err2 != nil {
		panic(err2)
	}
	manPhotos, womanPhotos, err3 = config.GetRobotNameByExcel("photo", file)
	if err3 != nil {
		panic(err3)
	}
}
