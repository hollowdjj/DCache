package DCache

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	enableLogger = false
	logger       *logrus.Logger
)

func GetLogger() *logrus.Logger {
	if !enableLogger {
		panic("logger not enabled")
	}
	return logger
}

func EnableLogger(path string) {
	enableLogger = true
	logger = logrus.New()
	//打印函数
	logger.SetReportCaller(true)
	//设置日志格式
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	//设置日志拆分
	logger.SetOutput(&lumberjack.Logger{
		Filename:   path,
		MaxSize:    50,   //单个日志文件最大50MB
		MaxBackups: 1,    //1个备份
		LocalTime:  true, //使用本地时间
		MaxAge:     3,    //只保留3天前的日志
	})
}
