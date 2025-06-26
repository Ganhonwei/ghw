package online

import (
	"context"
	"encoding/json"
	"goserver/gen/pb"
	"goserver/pkg/data/mq"
	"goserver/pkg/glog"
)

func requestOnlineUserMinuteReq(req *mq.RequestEmptyArgs) (ret *pb.OnlineUserMinuteRsp, err error) {
	userIds, err := rdb.HKeys(context.Background(), ONLINE_USER_KEY).Result()
	if err != nil {
		glog.Error("requestOnlineUserMinuteReq", "error", err)
		return
	}

	return &pb.OnlineUserMinuteRsp{
		UserIds: userIds,
	}, nil
}

func requestOnlineUserListReq(req *mq.RequestEmptyArgs) (ret *pb.OnlineUserListRsp, err error) {
	userList, err := rdb.HGetAll(context.Background(), ONLINE_USER_KEY).Result()
	if err != nil {
		glog.Error("requestOnlineUserListReq", "error", err)
		return
	}

	list := make([]*pb.OnlineUserList, 0)
	for _, u := range userList {
		onlineUser := &pb.OnlineUserList{}
		err := json.Unmarshal([]byte(u), &onlineUser)
		if err != nil {
			glog.Error("requestOnlineUserListReq", "error", err)
			continue
		}
		list = append(list, onlineUser)
	}

	return &pb.OnlineUserListRsp{
		UserList: list,
	}, nil

}
