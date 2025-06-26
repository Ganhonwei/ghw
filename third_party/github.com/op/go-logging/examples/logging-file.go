package main

import (
	"flag"
	"os"
	"time"

	logging "go-logging"
)

var log = logging.MustGetLogger("logging")

var format1 = logging.MustStringFormatter(
	`[%{level:.4s}] [%{time:2006-01-02 15:04:05.000}] [%{shortfile}] %{shortfunc} %{pid} %{message}`,
)

var format2 = logging.MustStringFormatter(
	`%{color}[%{level:.4s}] [%{time:15:04:05.000}] [%{shortfile}] %{pid}%{color:reset} %{message}`,
)

type Password string

func (p Password) Redacted() interface{} {
	return logging.Redact(string(p))
}

func main() {
	flag.Parse()
	//now := time.Now()
	//file, _, err := create("log", now)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//backend1 := logging.NewLogBackend(file, "", 0)
	backend1 := logging.NewLogBackendFile(os.Stderr, "", 0)
	//backend2 := logging.NewLogBackend(os.Stderr, "", 0)

	backend1Formatter := logging.NewBackendFormatter(backend1, format1)
	backend1Leveled := logging.AddModuleLevel(backend1Formatter)
	backend1Leveled.SetLevel(logging.INFO, "")

	//backend2Formatter := logging.NewBackendFormatter(backend2, format2)
	//backend2Leveled := logging.AddModuleLevel(backend2Formatter)
	//backend2Leveled.SetLevel(logging.INFO, "")

	//logging.SetBackend(backend1Leveled, backend2Leveled)
	logging.SetBackend(backend1Leveled)

	for i := 0; i <= 1000; i++ {
		log.Debugf("debug %s", Password("secret"))
		log.Info("info")
		log.Notice("notice")
		log.Warning("warning")
		log.Error("err")
		log.Critical("crit")
		//<-time.After(200 * time.Millisecond)
	}

	<-time.After(10 * time.Second)
}
