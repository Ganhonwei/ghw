package config

import "goserver/pkg/data"

// 对战房基础配置
var PvpRoom *data.PvpRoom

// InitPvpRoom 启动初始化
func InitPvpRoom() {
	pvpRooms := data.GetPvpRoomList()
	if len(pvpRooms) == 0 {
		return
	}
	PvpRoom = &pvpRooms[0]
}

// ----------------------对战房配置----------------------------------
func SetPvpRoom(pvpRoom *data.PvpRoom) {
	PvpRoom = pvpRoom
}

func GetPvpRoom() *data.PvpRoom {
	return PvpRoom
}
