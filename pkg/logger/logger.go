package logger

import (
	"log"
	"os"
)

type Logger struct {
	logInfo *log.Logger
	logWarn *log.Logger
	logErr  *log.Logger
}

func New() *Logger {
	flags := log.LstdFlags | log.Lshortfile
	file, _ := os.OpenFile("logs/info.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	logInfo := log.New(file, "INFO:  ", flags)
	logWarn := log.New(file, "WARN:  ", flags)
	logErr := log.New(file, "ERR:  ", flags)

	return &Logger{
		logInfo: logInfo,
		logWarn: logWarn,
		logErr:  logErr,
	}
}

func (l *Logger) Info(v ...interface{}) {
	l.logInfo.Println(v...)
}
func (l *Logger) Warn(v ...interface{}) {
	l.logWarn.Println(v...)
}
func (l *Logger) Err(v ...interface{}) {
	l.logErr.Println(v...)
}
