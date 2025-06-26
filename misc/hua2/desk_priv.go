package main

import "goserver/gen/pb"

// isPrivRoom 是否是私人房
func (t *Desk) isPrivRoom() bool {
	return t.Rtype == int32(pb.ROOM_TYPE1)
}

// isPrivFunRoom 是否私人房真金模式
func (t *Desk) isPrivCashRoom() bool {
	return t.Rtype == int32(pb.ROOM_TYPE1) && t.Gmode == 0
}

// isPrivFunRoom 是否私人房娱乐模式
func (t *Desk) isPrivFunRoom() bool {
	return t.isPrivRoom() && t.Gmode == 1
}
