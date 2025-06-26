package handler

import (
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"net/http"

	ini "gopkg.in/ini.v1"
)

func StartReadinessProbe(cfg *ini.File, port string) {
	probe := cfg.Section("env").Key("probe").MustBool(false)
	if !probe {
		return
	}

	http.HandleFunc("/readiness", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/reload-config", func(w http.ResponseWriter, r *http.Request) {
		err := table.LoadTables()
		if err != nil {
			glog.Errorf("reload config err %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		glog.Infof("reload config success")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	go func() {
		glog.Infof("就绪探针开始监听 %s 端口", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			glog.Errorf("就绪探针服务出错: %v", err)
		}
	}()
}
