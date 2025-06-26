package handler

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
)

// PackMinesRoom 打包房间信息
func PackMinesRoom(d *data.DeskData) *pb.MinesRoomData {
	return &pb.MinesRoomData{
		Roomid: d.Rid,   //牌局id
		Gtype:  d.Gtype, //game type
		Rtype:  d.Rtype, //room type
		Dtype:  d.Dtype, //desk type
		Ltype:  d.Ltype, //level type
		Rname:  d.Rname, //room name
	}
}
