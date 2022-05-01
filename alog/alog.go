// Package alog provides async logging with goroutine.
package alog

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *log.Logger

func TraceLog(exchange, msg string) {
	fileName := fmt.Sprintf("bs_%s.log", exchange)
	fp, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	logger = log.New(fp, "", log.Ldate|log.Ltime)
	logger.SetOutput(&lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    1, // megabytes after which new file is created
		MaxBackups: 2, // number of backups
		MaxAge:     3, //days
	})
	logger.Println(msg)
}
