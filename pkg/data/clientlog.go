package data

type ClientLog struct {
	DeviceId string
	Time     string
	Content  string
}

func (t *ClientLog) Save() bool {
	return Insert(ClientLogs, t)
}
