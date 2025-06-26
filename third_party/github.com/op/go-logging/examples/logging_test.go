package main

import "testing"

func TestLog(t *testing.T) {
	defer Flush()
	Init()
	//go run logging.go -log_dir=. -stderrthreshold="INFO"
	log.Debugf("debug %s", Password("secret"))
	log.Info("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("err")
	log.Critical("crit")
}
