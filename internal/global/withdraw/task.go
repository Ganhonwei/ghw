package withdraw

import (
	"goserver/pkg/glog"

	"github.com/robfig/cron/v3"
)

var c *cron.Cron

func StartAutoTransferChecker() {
	c = cron.New(cron.WithSeconds())
	c.AddFunc("0 0/5 * * * *", func() {
		funcData(checkTransfer)
	}) // 转单
	c.AddFunc("5 0 0 * * *", func() {
		funcData(ResetTransferCash)
	}) // 重置补贴金
	c.Start()
}

func funcData(f func() error) {
	defer func() {
		if err := recover(); err != nil {
			glog.Errorf("funcData error: %v", err)
		}
	}()
	err := f()
	if err != nil {
		glog.Errorf(err.Error())
	}
}
