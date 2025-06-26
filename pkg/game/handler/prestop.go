package handler

import (
	"goserver/pkg/glog"
	"net/http"

	ini "gopkg.in/ini.v1"
)

func StartPrestopCallback(cfg *ini.File, port string, callback func()) {
	probe := cfg.Section("env").Key("prestop").MustBool(false)
	if !probe {
		return
	}

	http.HandleFunc("/prestop", func(w http.ResponseWriter, r *http.Request) {
		glog.Infof("prestop callback 收到请求")
		callback()
		glog.Infof("prestop callback 执行完毕")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	go func() {
		glog.Infof("prestop callback 开始监听 %s 端口", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			glog.Errorf("prestop callback 服务出错: %v", err)
		}
	}()
}
