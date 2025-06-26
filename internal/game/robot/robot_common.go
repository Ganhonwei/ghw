package robot

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/gen/tb"
	"goserver/pkg/utils"
)

func IsValid(ec tb.EmojiRobotEmojiConfigRecord) bool {
	if ec.Id == 0 {
		return false
	} else {
		return true
	}
}

func getGameName(gtype int32) string {
	switch gtype {
	case int32(pb.HUA):
		return cfg.Section("game.hua").Name()
	case int32(pb.LHD):
		return cfg.Section("game.lhd").Name()
	case int32(pb.SEVEN):
		return cfg.Section("game.7up").Name()
	case int32(pb.RUMMY):
		return cfg.Section("game.rummy").Name()
	case int32(pb.AK47):
		return cfg.Section("game.ak47").Name()
	case int32(pb.JOKER):
		return cfg.Section("game.joker").Name()
	case int32(pb.CRASH):
		return cfg.Section("game.crash").Name()
	case int32(pb.ABAR):
		return cfg.Section("game.andarbahar").Name()
	case int32(pb.LOTTERY):
		return cfg.Section("game.lottery").Name()
	case int32(pb.PLANE):
		return cfg.Section("game.plane").Name()
	case int32(pb.RUMMY2):
		return cfg.Section("game.rummy_2").Name()
	case int32(pb.HUA2):
		return cfg.Section("game.hua_2").Name()
	}
	return ""
}

func GetUserBase(vipLv int, custom bool) (string, string, uint32) {
	var photoRange int32 = 30
	if vipLv >= 8 {
		photoRange = 52
	} else if vipLv >= 4 {
		photoRange = 42
	}

	name := fmt.Sprintf("Player%d", utils.RandInt32N(8999999)+1000000)
	photo := fmt.Sprintf("%d", utils.RandInt32N(photoRange)+1)
	var sex uint32 = uint32(utils.RandInt32N(2) + 1)
	if utils.RandWan(5000) {
		// 50%概率使用配置的名字
		india := false
		if utils.RandWan(2000) {
			// 剩下的人中20%使用印度名字
			india = true
		}
		if utils.RandWan(6000) {
			// 剩下的人中60%中是男性
			sex = 1
			i := utils.RandInt32N(int32(len(manPhotos)))
			photo = manPhotos[i]
			if india && len(manInNames) > 0 {
				index := utils.RandInt32N(int32(len(manInNames)))
				name = manInNames[index]
				// manInNames = append(manInNames[:index], manInNames[index+1:]...)
			} else {
				index := utils.RandInt32N(int32(len(manEnNames)))
				name = manEnNames[index]
				// manEnNames = append(manEnNames[:index], manEnNames[index+1:]...)
			}
		} else {
			sex = 2
			// 女性
			i := utils.RandInt32N(int32(len(womanPhotos)))
			photo = womanPhotos[i]
			if india {
				index := utils.RandInt32N(int32(len(womanInNames)))
				name = womanInNames[index]
				// womanInNames = append(womanInNames[:index], womanInNames[index+1:]...)
			} else {
				index := utils.RandInt32N(int32(len(womanEnNames)))
				name = womanEnNames[index]
				// womanEnNames = append(womanEnNames[:index], womanEnNames[index+1:]...)
			}
		}
	}

	// 80%概率使用自定义头像
	if utils.RandWan(8000) && custom && vipLv > 0 {
		if utils.RandWan(9000) {
			// 剩下的人中60%中是男性
			info := getHead("man")
			if info != nil {
				name = info.Name
				photo = fmt.Sprintf("https://cdn.g2qdh.com/avatar/man/%d.jpg", info.Id)
			}
		} else {
			info := getHead("woman")
			if info != nil {
				name = info.Name
				photo = fmt.Sprintf("https://cdn.g2qdh.com/avatar/woman/%d.jpg", info.Id)
			}
		}
	}
	return name, photo, sex
}

func GetContentAndDelay(prob int32, weight []int32, delay []int32) (content string, ret_delay int) {
	// if true {
	if utils.RandWan(int32(prob)) {
		var choices []utils.Choice
		for i, v := range weight {
			choices = append(choices, utils.Choice{Weight: int(v), Item: i})
		}
		_c, _ := utils.WeightedChoice(choices)
		i := _c.Item.(int)
		content = fmt.Sprintf("%d", i+1)
		ret_delay = utils.RandMN(int(delay[0]), int(delay[1]))
	}
	return
}

func GetContentAndDelay2(prob []int32, weight []int32, delay []int32, win, lose uint32) (to uint32, content string, ret_delay int) {

	i, _ := utils.ChoiceInt32Index(prob)
	switch i {
	case 0: //不发
		return
	case 1: //发赢
		to = win
		face_idx, _ := utils.ChoiceInt32Index(weight)
		content = fmt.Sprintf("%d", face_idx+1)
		ret_delay = utils.RandMN(int(delay[0]), int(delay[1]))
	case 2: //发输
		to = lose
		face_idx, _ := utils.ChoiceInt32Index(weight)
		content = fmt.Sprintf("%d", face_idx+1)
		ret_delay = utils.RandMN(int(delay[0]), int(delay[1]))
	}
	return
}

// 游戏开始
func (r *RoleActor) EmojiGameStart(to uint32) {
	if !IsValid(r.emoji) {
		return
	}

	content, delay := GetContentAndDelay(r.emoji.GameStartProb, r.emoji.GameStartWeight, r.emoji.GameStartDelay)
	r.SendEmoji(to, content, delay)
}

func (c *RoleActor) EmojiWaitLong(to uint32) {
	if !IsValid(c.emoji) {
		return
	}
	content, delay := GetContentAndDelay(c.emoji.WaitLongProb, c.emoji.WaitLongWeight, c.emoji.WaitLongDelay)
	c.SendEmoji(to, content, delay)
}

// 上桌发送
func (c *RoleActor) EmojiDeskSend(to uint32) {
	if !IsValid(c.emoji) {
		return
	}
	content, delay := GetContentAndDelay(c.emoji.SendProb, c.emoji.SendWeight, c.emoji.SendDelay)
	c.SendEmoji(to, content, delay)
}

// 回复表情
func (r *RoleActor) EmojiReply(to uint32) {
	if !IsValid(r.emoji) {
		return
	}
	content, delay := GetContentAndDelay(r.emoji.ReplyProb, r.emoji.ReplyWeight, r.emoji.ReplyDelay)
	r.SendEmoji(to, content, delay)
}

// 其他玩家看牌后
func (r *RoleActor) EmojiOtherSee(to uint32) {
	if !IsValid(r.emoji) {
		return
	}
	content, delay := GetContentAndDelay(r.emoji.OtherSeeProb, r.emoji.OtherSeeWeight, r.emoji.OtherSeeDelay)
	r.SendEmoji(to, content, delay)
}

// 其他玩家加注
func (r *RoleActor) EmojiOtherRaise(to uint32) {
	if !IsValid(r.emoji) {
		return
	}
	content, delay := GetContentAndDelay(r.emoji.OtherRaiseProb, r.emoji.OtherRaiseWeight, r.emoji.OtherRaiseDelay)
	r.SendEmoji(to, content, delay)
}

// 其他玩家比牌
func (r *RoleActor) EmojiOtherBi(win, lose uint32) {
	if !IsValid(r.emoji) {
		return
	}
	to, content, delay := GetContentAndDelay2(r.emoji.OtherBiProb, r.emoji.OtherBiWeight, r.emoji.OtherBiDelay, win, lose)
	r.SendEmoji(to, content, delay)
}

// 自己比牌赢
func (r *RoleActor) EmojiSelfBiWin(lose uint32) {
	if !IsValid(r.emoji) {
		return
	}
	content, delay := GetContentAndDelay(r.emoji.SelfBiWinProb, r.emoji.SelfBiWinWeight, r.emoji.SelfBiWinDelay)
	r.SendEmoji(lose, content, delay)
}

// 摸上家牌
func (r *RoleActor) EmojiMoPreCard(seat uint32) {
	if !IsValid(r.emoji) {
		return
	}
	content, delay := GetContentAndDelay(r.emoji.MoPreCardProb, r.emoji.MoPreCardWeight, r.emoji.MoPreCardDelay)
	r.SendEmoji(seat, content, delay)
}

// 其他玩家出牌
func (r *RoleActor) EmojiOtherChuCard(seat uint32) {
	if !IsValid(r.emoji) {
		return
	}
	content, delay := GetContentAndDelay(r.emoji.OtherChuCardProb, r.emoji.OtherChuCardWeight, r.emoji.OtherChuCardDelay)
	r.SendEmoji(seat, content, delay)
}

// 发送表情
func (r *RoleActor) SendEmoji(to uint32, content string, delay int) {
	if to == 0 || content == "" || delay == 0 {
		return
	}
	c2s := new(pb.ChatTextReq)
	c2s.To = to
	c2s.Content = content
	r.SendEmojiDefer(c2s, int64(delay))
}
