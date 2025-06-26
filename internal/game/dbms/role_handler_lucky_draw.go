package dbms

import (
	"context"
	"fmt"
	"goserver/gen/pb"
	"goserver/gen/tb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myredis"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"strconv"
	"sync"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/redis/go-redis/v9"
	"gopkg.in/mgo.v2/bson"
)

const (
	DrawNumberCharge int64  = 50000 // 充500一个号
	DrawNumberMax    int32  = 999   // 开奖抽奖号
	DrawNumberKey    string = "luckydraw:curnumber"
)

var (
	drawRound            int64                           = 1 // 当前抽奖轮数
	drawNumber           int32                           = 1 // 当前抽奖号
	drawUsers            map[string][]int32                  // userid -> 抽奖号s
	firstNumbers         map[int32]*data.LuckyDrawNumber     //一等奖号码
	drawLock             *sync.RWMutex
	robotTimeRanges      [4][2]time.Time
	robotTriggerTime     int64
	prev20RoundGives     []*data.LuckyDrawGive        // 前20轮开奖结果
	prev20RoundDrawUsers map[int64]map[string][]int32 // 前20轮玩家号码
)

type LuckDrawGive struct {
	RewordId int
	Userid   string
	Number   string // 中奖号码
	Reword   int64  // 奖励彩金
	Round    int64  // 开奖轮数
}

// 剩余中奖数
type LuckDrawRest struct {
	RewordId int
	Rest     int32 // 剩余中奖数
	Reword   int64 // 奖励彩金
}

// initLuckyDraw 启动时初始化
func (a *RoleActor) initLuckyDraw() {
	drawUsers = make(map[string][]int32)
	firstNumbers = make(map[int32]*data.LuckyDrawNumber)
	prev20RoundDrawUsers = make(map[int64]map[string][]int32)
	drawLock = &sync.RWMutex{}

	// 2:00:00-6:59:59 7:00:00-12:59:59	13:00:00-18:59:59 19:00:00-1:59:59
	robotTimeRanges = [4][2]time.Time{
		{utils.StrToTime("2024-01-01 02:00:00"), utils.StrToTime("2024-01-01 06:59:59")},
		{utils.StrToTime("2024-01-01 07:00:00"), utils.StrToTime("2024-01-01 12:59:59")},
		{utils.StrToTime("2024-01-01 13:00:00"), utils.StrToTime("2024-01-01 18:59:59")},
		{utils.StrToTime("2024-01-01 19:00:00"), utils.StrToTime("2024-01-02 01:59:59")},
	}

	// read db
	curRound, err := data.GetLuckyDrawCurRound()
	if err != nil {
		glog.Errorf("小米手机活动初始化错误1: %v", err)
		return
	}
	if curRound == 0 {
		luckyDrawRoundInit()
		return
	}
	drawRound = curRound

	// 读取上一轮开奖记录
	reloadPrev20Draw(curRound)

	numbers := data.GetDrawNumbersByRound(curRound)
	if len(numbers) == 0 {
		luckyDrawRoundInit()
		return
	}
	var firsts []int32
	for _, number := range numbers {
		// 人机一等奖劵
		if number.Robot {
			firstNumbers[number.Number] = number
			firsts = append(firsts, number.Number)
			continue
		}
		// 玩家奖券
		drawUsers[number.Userid] = append(drawUsers[number.Userid], number.Number)
	}

	// 当前抽奖号 drawNumber
	drawNumberLoadFromDB := func() {
		// 从数据库加载最后一个记录玩家的抽奖号
		maxNumber, err := data.GetRoundDrawNumberMax(curRound, firsts...)
		if err != nil {
			glog.Errorf("lucky draw init error2: %v", err)
			return
		}
		drawNumber = maxNumber.Number
	}
	// 从 redis 读最新的抽奖号
	num1, err := myredis.Redis().Get(context.Background(), DrawNumberKey).Result()
	if err != nil {
		if err != redis.Nil {
			glog.Errorf("lucky draw init error3: %v", err)
		}
		drawNumberLoadFromDB()
		return
	}
	num2, err := strconv.Atoi(num1)
	if err != nil {
		glog.Errorf("lucky draw init error4: %v", err)
		drawNumberLoadFromDB()
		return
	}
	drawNumber = int32(num2)
}

// 新一期活动初始化
func luckyDrawRoundInit() {
	drawNumber = 1
	drawUsers = make(map[string][]int32)
	firstNumbers = make(map[int32]*data.LuckyDrawNumber)
	rewards := table.GetTables().XiaoMiRewardTable.GetDataList()
	// 把一等奖分配给人机
	for _, r := range rewards {
		if r.Id == 1 {
			for i := 0; i < int(r.Count); i++ {
				number := int32(utils.RandMN(1, int(DrawNumberMax)))
				numLog := &data.LuckyDrawNumber{
					Id:     bson.NewObjectId().String(),
					Round:  drawRound,
					Number: number,
					Userid: generateUserId(uint32(drawRound)),
					Robot:  true,
					Ctime:  time.Now(),
				}
				if !numLog.Save() {
					glog.Errorf("一等奖券记录保存失败: %+v", numLog)
				}
				firstNumbers[number] = numLog
			}
		}
	}

	// 读取上一轮开奖记录进行追加
	reloadPrevDraw(drawRound)
}

func reloadPrevDraw(curRound int64) {
	if curRound < 2 {
		return
	}
	prevRound := curRound - 1
	gives := data.GetDrawGivesByRound(prevRound)
	prev20RoundGives = append(prev20RoundGives, gives...)

	numbers := data.GetDrawNumbersByRound(prevRound)
	roundDrawUsers := make(map[string][]int32)
	for _, num := range numbers {
		roundDrawUsers[num.Userid] = append(roundDrawUsers[num.Userid], num.Number)
	}
	prev20RoundDrawUsers[curRound-1] = roundDrawUsers

	// 清除内存20轮前的数据
	if len(prev20RoundGives) == 0 {
		return
	}
	if 20 <= prev20RoundGives[len(prev20RoundGives)-1].Round-prev20RoundGives[0].Round {
		cleanRound := prev20RoundGives[0].Round
		delete(prev20RoundDrawUsers, cleanRound)
		newRounds := make([]*data.LuckyDrawGive, 0, len(prev20RoundGives))
		for _, give := range prev20RoundGives {
			if give.Round != cleanRound {
				newRounds = append(newRounds, give)
			}
		}
		prev20RoundGives = newRounds
	}
}

// 读取前20轮开奖记录
func reloadPrev20Draw(curRound int64) {
	if curRound < 2 {
		return
	}
	roundN := curRound - 1
	roundM := roundN - 20
	if roundM < 1 {
		roundM = 1
	}
	prev20RoundGives = data.GetDrawGivesByRoundMN(roundM, roundN)
	if len(prev20RoundGives) == 0 {
		return
	}
	numbers := data.GetDrawNumbersByRoundMN(roundM, roundN)

	roundUserNumbers := make(map[int64]map[string][]int32)
	var round int64
	for _, num := range numbers {
		if num.Round != round {
			round = num.Round
			roundUserNumbers[round] = make(map[string][]int32)
		}
		roundUserNumbers[round][num.Userid] = append(roundUserNumbers[round][num.Userid], num.Number)
	}
	prev20RoundDrawUsers = roundUserNumbers
}

// 发送邮件
func (a *RoleActor) sendLuckyDrawMail(userid string, amount int64, numbers []int32) {
	user := a.getUserById(userid)
	if user == nil {
		glog.Errorf("sendRankWithdrawMail error: user %s not found", userid)
		return
	}
	numbersStr := ""
	for i, num := range numbers {
		numbersStr += fmt.Sprintf("%03d", num)
		if i != len(numbers)-1 {
			numbersStr += ","
		}
	}

	if role, ok := a.roles[userid]; ok {
		msg := &pb.LuckyDrawNumberMail{Amount: amount, Numbers: numbersStr, Mtype: 1}
		role.Pid.Tell(msg)
	} else {
		// 离线
		chat := data.ChatLog{
			Uid:      bson.NewObjectId().String(),
			Receiver: user.Userid,
			Title:    "System Message",
			Content:  handler.BuildLuckyDraw(user.Nickname, numbersStr, amount/100),
			Ctime:    utils.LocalTime(),
			Sender:   "-1",
			Name:     "System",
		}
		user.FeedBackLogMap[chat.Uid] = chat
		user.UpdateFeedBack()
	}
}

// 发送邮件
func (a *RoleActor) sendLuckyDrawWinMail(userid string, amount int64, level, round int) {
	user := a.getUserById(userid)
	if user == nil {
		glog.Errorf("sendLuckyDrawWinMail error: user %s not found", userid)
		return
	}
	if role, ok := a.roles[userid]; ok {
		msg := &pb.LuckyDrawNumberMail{
			Amount: amount,
			Mtype:  2,
			Level:  int32(level),
			Round:  int32(round),
		}
		role.Pid.Tell(msg)
	} else {
		// 离线
		chat := data.ChatLog{
			Uid:      bson.NewObjectId().String(),
			Receiver: user.Userid,
			Title:    "System Message",
			Content:  handler.BuildLuckyDrawWin(user.Nickname, level, round, amount/100),
			Ctime:    utils.LocalTime(),
			Sender:   "-1",
			Name:     "System",
		}
		user.FeedBackLogMap[chat.Uid] = chat
		user.UpdateFeedBack()
	}
}

// 分配抽奖号
func (a *RoleActor) allocateDrawNumber(save bool, userid string, chargeMoney int64) {
	count := chargeMoney / DrawNumberCharge
	drawLock.Lock() // todo 优化
	defer drawLock.Unlock()

	var drawedNumbers []int32
	for i := 0; i < int(count); i++ {
		if userid != "" {
			drawUsers[userid] = append(drawUsers[userid], drawNumber)
			glog.Debugf("lucky draw number: %d-%d, %s", drawRound, drawNumber, userid)
		}
		if save {
			if userid == "" {
				userid = generateUserId(uint32(drawRound))
			} else {
				drawedNumbers = append(drawedNumbers, drawNumber)
			}
			numLog := &data.LuckyDrawNumber{
				Id:     bson.NewObjectId().String(),
				Round:  drawRound,
				Number: drawNumber,
				Userid: userid,
				Robot:  false,
				Ctime:  time.Now(),
			}
			if !numLog.Save() {
				glog.Errorf("玩家抽奖号保存失败: %+v", numLog)
			}
		}
		if drawNumber%50 == 0 { // 50个打印下进度
			glog.Debugf("lucky draw number: %d-%d, %s", drawRound, drawNumber, userid)
		}

		// drawNumber 保存
		nextDrawNumber := drawNumber
		for i := 0; i <= int(DrawNumberMax); i++ {
			nextDrawNumber++
			if _, ok := firstNumbers[nextDrawNumber]; !ok {
				break
			}
		}
		if nextDrawNumber > DrawNumberMax {
			a.luckyDrawRoundOver()
		} else {
			drawNumber = nextDrawNumber
			myredis.Redis().Set(context.Background(), DrawNumberKey, fmt.Sprintf("%d", drawNumber), 0)
		}
	}

	// 发邮件
	if len(drawedNumbers) > 0 {
		go a.sendLuckyDrawMail(userid, chargeMoney, drawedNumbers)
	}
}

// 开奖
func (a *RoleActor) luckyDrawRoundOver() {
	drawRound++

	rewords := config.GetLuckyDrawRewards()
	rewordRests := make(map[int]*LuckDrawRest, len(rewords))
	for _, r := range rewords {
		rewordRests[r.Id] = &LuckDrawRest{
			RewordId: r.Id,
			Rest:     r.Limit,
			Reword:   r.Reward,
		}
		if r.Id == 1 {
			rewordRests[r.Id].Rest = 0
		}
	}

	var gives []*LuckDrawGive
	for userid, numbers := range drawUsers {
	userLoop:
		for _, number := range numbers {
			// 从五等奖往上判断
			for i := len(rewords) - 1; i >= 0; i-- {
				reword := rewords[i]
				if rest, ok := rewordRests[reword.Id]; !ok || rest.Rest <= 0 {
					continue
				}
				if reword.Rate == 0 {
					continue
				}
				if !utils.RandWan(reword.Rate) {
					continue
				}
				// 中奖
				rewordRests[reword.Id].Rest-- // 中奖人数减一
				gives = append(gives, &LuckDrawGive{
					RewordId: reword.Id,
					Userid:   userid,
					Number:   fmt.Sprintf("%03d", number),
					Round:    drawRound - 1,
					Reword:   reword.Reward,
				})
				break userLoop
			}
		}
	}

	a.luckyDrawGive(gives, rewordRests)
}

// 派奖
func (a *RoleActor) luckyDrawGive(userGives []*LuckDrawGive, rewordRests map[int]*LuckDrawRest) {
	// 发奖
	for _, give := range userGives {
		giveLog := &data.LuckyDrawGive{
			Id:       bson.NewObjectId().String(),
			Round:    give.Round,
			RewordId: int32(give.RewordId),
			Reword:   give.Reword,
			Userid:   give.Userid,
			Number:   give.Number,
			Robot:    false,
			Ctime:    time.Now(),
		}
		if !giveLog.Save() {
			glog.Errorf("中奖记录保存失败: %+v", giveLog)
		} else {
			// 发送中奖邮件
			go a.sendLuckyDrawWinMail(give.Userid, give.Reword, give.RewordId, int(give.Round))
		}
	}
	// 保存一等奖记录
	for _, number := range firstNumbers {
		log := &data.LuckyDrawGive{
			Id:       bson.NewObjectId().String(),
			Round:    drawRound - 1,
			RewordId: 1,
			Reword:   0,
			Userid:   number.Userid,
			Number:   fmt.Sprintf("%d", number.Number),
			Robot:    true,
			Ctime:    time.Now(),
		}
		if !log.Save() {
			glog.Errorf("一等奖记录保存失败: %+v", log)
		}
	}

	// 剩下奖励生成机器人中奖, 与玩家和一等奖不重号
	drawedNumbers := make(map[int32]bool)
	for _, numbers := range drawUsers {
		for _, number := range numbers {
			drawedNumbers[number] = true
		}
	}
	for number := range firstNumbers {
		drawedNumbers[number] = true
	}

	for _, rest := range rewordRests {
		if rest.Rest < 1 {
			continue
		}
		for i := 0; i < int(rest.Rest); i++ {
			// 生成彩票号
			var number int32
			for i := 0; i < 10000; i++ {
				num := int32(utils.RandMN(1, int(DrawNumberMax)))
				if !drawedNumbers[num] {
					number = num
					break
				}
			}
			if number == 0 {
				var num int32 = 1
				for ; num <= DrawNumberMax; num++ {
					if !drawedNumbers[num] {
						number = num
						break
					}
				}
			}
			if number == 0 {
				glog.Errorf("彩票号生成失败: %+v", rewordRests)
				continue
			}
			// 保存机器人中奖记录
			log := &data.LuckyDrawGive{
				Id:       bson.NewObjectId().String(),
				Round:    drawRound - 1,
				RewordId: int32(rest.RewordId),
				Reword:   rest.Reword,
				Number:   fmt.Sprintf("%d", number),
				Robot:    true,
				Ctime:    time.Now(),
			}
			log.Userid = generateUserId(uint32(log.Round))
			if !log.Save() {
				glog.Errorf("机器人中奖记录保存失败: %+v", log)
			}
		}
	}

	// 重置奖券
	luckyDrawRoundInit()
}

// 定时放号
func (a *RoleActor) luckyDrawTime() {
	now := time.Now()
	if now.Unix() < robotTriggerTime {
		return
	}
	// robots := config.GetLuckyDrawRobots()
	robots := table.GetTables().XiaoMiRobotTable.GetDataList()
	if len(robots) == 0 {
		return
	}

	// 放一个号
	go a.allocateDrawNumber(false, "", DrawNumberCharge)

	// 计算下次触发时间
	// 匹配放号范围
	var drawRobot *tb.XmXiaoMiRobotRecord
	for _, robot := range robots {
		if drawNumber >= robot.Counts[0] && drawNumber <= robot.Counts[1] {
			drawRobot = robot
			break
		}
	}
	if drawRobot == nil {
		glog.Errorf("小米手机活动定时放号错误: %v, %v", drawNumber, robots)
		return
	}

	// 匹配当前时间段
	day := 1
	if now.Hour() < 2 {
		day = 2
	}
	var period int
	now = utils.StrToTime(fmt.Sprintf("2024-01-%02d %02d:%02d:%02d", day, now.Hour(), now.Minute(), now.Second()))
	for i := 0; i < len(robotTimeRanges); i++ {
		if now.After(robotTimeRanges[i][0]) && now.Before(robotTimeRanges[i][1]) {
			period = i
			break
		}
	}
	var timePeriod []int32
	switch period {
	case 0:
		timePeriod = drawRobot.TimePeriod1
	case 1:
		timePeriod = drawRobot.TimePeriod2
	case 2:
		timePeriod = drawRobot.TimePeriod3
	case 3:
		timePeriod = drawRobot.TimePeriod4
	}
	nextSeconds := utils.RandMN(int(timePeriod[0]), int(timePeriod[1]))
	robotTriggerTime = time.Now().Unix() + int64(nextSeconds)
}

func generateUserId(channelId uint32) string {
	return strconv.FormatInt(handler.GenerateOrderId(channelId), 10)
}

// 玩家充值放号
func (a *RoleActor) LuckyDrawNumber(ctx actor.Context) {
	arg := ctx.Message().(*pb.LuckyDrawNumber)
	glog.Debugf("LuckyDrawNumber %#v", arg)
	if arg.ChargeMoney < DrawNumberCharge {
		return
	}
	go a.allocateDrawNumber(true, arg.Userid, arg.ChargeMoney)
}

// 活动页请求
func (a *RoleActor) LuckyDrawReq(ctx actor.Context) {
	drawLock.RLock()
	defer drawLock.RUnlock()

	arg := ctx.Message().(*pb.LuckyDrawReq)
	glog.Debugf("LuckyDrawReq %#v", arg)
	rsp := &pb.LuckyDrawRsp{}
	// rule rewords
	rewords := config.GetLuckyDrawRewards()
	for _, reword := range rewords {
		r := []int32{reword.Limit, int32(reword.Reward)}
		switch reword.Id {
		case 2:
			rsp.Reword2 = r
		case 3:
			rsp.Reword3 = r
		case 4:
			rsp.Reword4 = r
		case 5:
			rsp.Reword5 = r
		}
	}
	// current round
	rsp.CurRound = int32(drawRound)
	rsp.CurNumber = int32(drawNumber) + int32(len(firstNumbers))
	if rsp.CurNumber > DrawNumberMax {
		rsp.CurNumber = DrawNumberMax
	}
	rsp.MaxNumber = DrawNumberMax
	if numbers, ok := drawUsers[arg.Userid]; ok {
		rsp.YourNumbers = numbers
	}

	// prev round
	prevRound := drawRound - 1
	rsp.PrevRound = int32(prevRound)
	// 上一轮玩家号码
	if numbers, ok := prev20RoundDrawUsers[prevRound]; ok {
		rsp.PrevYourNumbers = numbers[arg.Userid]
	}
	// 上一轮开奖记录
	var prevGives []*data.LuckyDrawGive
	if drawRound > 1 {
		for i := len(prev20RoundGives) - 1; i >= 0; i-- {
			if prev20RoundGives[i].Round == prevRound {
				prevGives = append(prevGives, prev20RoundGives[i])
			}
			if prev20RoundGives[i].Round != prevRound && len(prevGives) > 0 {
				break
			}
		}
	}
	if len(prevGives) == 0 {
		ctx.Respond(rsp)
		return
	}
	for _, give := range prevGives {
		number, err := strconv.Atoi(give.Number)
		if err != nil {
			glog.Errorf("prev number parse error: %s", give.Number)
			continue
		}
		if give.Userid == arg.Userid {
			rsp.PrevReword = 1
			if give.Taked {
				rsp.PrevReword = 2
			}
		}
		switch give.RewordId {
		case 1:
			rsp.PrevNumbers1 = append(rsp.PrevNumbers1, int32(number))
		case 2:
			rsp.PrevNumbers2 = append(rsp.PrevNumbers2, int32(number))
		case 3:
			rsp.PrevNumbers3 = append(rsp.PrevNumbers3, int32(number))
		case 4:
			rsp.PrevNumbers4 = append(rsp.PrevNumbers4, int32(number))
		case 5:
			rsp.PrevNumbers5 = append(rsp.PrevNumbers5, int32(number))
		}
	}
	ctx.Respond(rsp)
}

// 中奖历史记录
func (a *RoleActor) LuckyDrawHistoryReq(ctx actor.Context) {
	drawLock.RLock()
	defer drawLock.RUnlock()

	arg := ctx.Message().(*pb.LuckyDrawHistoryReq)
	glog.Debugf("LuckyDrawHistoryReq %#v", arg)
	rsp := &pb.LuckyDrawHistoryRsp{}
	if drawRound < 2 {
		ctx.Respond(rsp)
		return
	}
	roundN := drawRound - 1
	roundM := roundN - 20
	if roundM < 1 {
		roundM = 1
	}

	var round int64
	var history *pb.LuckyDrawHistory
	for i := len(prev20RoundGives) - 1; i >= 0; i-- {
		give := prev20RoundGives[i]
		if give.Round != round {
			round = give.Round
			history = &pb.LuckyDrawHistory{Round: round}
			rsp.History = append(rsp.History, history)
			// 该轮玩家中奖号码
			if userNumbers, ok := prev20RoundDrawUsers[round]; ok {
				history.YourNumbers = userNumbers[arg.Userid]
			}
		}

		// 该轮玩家是否中奖
		if give.Userid == arg.Userid {
			history.Reword = 1
			if give.Taked {
				history.Reword = 2
			}
		}

		// 该轮开奖结果
		number, err := strconv.Atoi(give.Number)
		if err != nil {
			glog.Errorf("prev number parse error: %s", give.Number)
			continue
		}
		switch give.RewordId {
		case 1:
			history.Numbers1 = append(history.Numbers1, int32(number))
		case 2:
			history.Numbers2 = append(history.Numbers2, int32(number))
		case 3:
			history.Numbers3 = append(history.Numbers3, int32(number))
		case 4:
			history.Numbers4 = append(history.Numbers4, int32(number))
		case 5:
			history.Numbers5 = append(history.Numbers5, int32(number))
		}
	}
	ctx.Respond(rsp)
}

// 领奖
func (a *RoleActor) LuckyDrawTake(ctx actor.Context) {
	drawLock.RLock()
	defer drawLock.RUnlock()
	arg := ctx.Message().(*pb.LuckyDrawTake)
	glog.Debugf("LuckyDrawTake %#v", arg)
	rsp := &pb.LuckyDrawTaked{}

	for _, give := range prev20RoundGives {
		if give.Round != arg.Round {
			continue
		}
		if give.Userid != arg.Userid || give.Robot {
			continue
		}
		if give.Taked {
			rsp.Error = pb.AlreadyPrize
			ctx.Respond(rsp)
			return
		}
		give.Taked = true
		go give.UpdateTaked()

		rsp.RewordAmount = give.Reword
		rsp.Round = give.Round
		rsp.RewordId = give.RewordId
		rsp.Number = give.Number
		ctx.Respond(rsp)
		return
	}
	rsp.Error = pb.AwardFaild
	ctx.Respond(rsp)
}
