package zlog

import (
	"fmt"
	"goserver/pkg/utils"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

func Init(outdir, outfile string) {
	// 创建文件导出目录
	if !utils.FileExists(outdir) {
		if e1 := os.Mkdir(outdir, os.ModePerm); e1 != nil {
			fmt.Println("mkdir export dir error: ", e1)
		}
	}

	encoder := getEncoder()
	writeSyncer := getLogWriter(outdir, outfile)
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)

	// AddCaller 添加将调用函数信息记录到日志中的功能。
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	logger.Info("zlog init")
	sugarLogger := logger.Sugar()
	Logger = sugarLogger
}

func getLogWriter(outdir, outfile string) zapcore.WriteSyncer {
	// file, _ := os.Create(outdir + "/" + outfile)
	// 追加文件
	file, err := os.OpenFile(outdir+"/"+outfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("open log file error: ", outdir, "/", outfile, err)
	}
	return zapcore.AddSync(file)
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // 修改时间编码器

	// 在日志文件中使用大写字母记录日志级别
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// NewConsoleEncoder 打印更符合人们观察的方式
	return zapcore.NewConsoleEncoder(encoderConfig)
}
