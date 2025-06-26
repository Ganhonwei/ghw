package config

import (
	"goserver/pkg/data"
	"sync"
)

// 分享活动
var EmojiMap *sync.Map

// InitEmoji 启动初始化
func InitEmoji() {
	EmojiMap = new(sync.Map)
	l := data.GetEmojiList()
	for _, v := range l {
		SetEmoji(v)
	}
}

// InitEmoji2 启动初始化
func InitEmoji2() {
	EmojiMap = new(sync.Map)
}

// SetEmoji 添加新的分享数据
func SetEmoji(v data.Emoji) {
	EmojiMap.Store(v.Id, v)
}

func GetEmoji(id int32) data.Emoji {
	if s, ok := EmojiMap.Load(id); ok {
		if emoji, ok := s.(data.Emoji); ok {
			return emoji
		}
		return data.Emoji{}
	}
	return data.Emoji{}
}

func GetEmojiMap() map[int32]data.Emoji {
	emojis := make(map[int32]data.Emoji)
	EmojiMap.Range(func(key, value any) bool {
		emojis[key.(int32)] = value.(data.Emoji)
		return true
	})
	return emojis
}
