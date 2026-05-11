package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	infoLog  *log.Logger
	errorLog *log.Logger
)

func init() {
	// 创建日志目录
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal("Failed to create log directory:", err)
	}

	// 创建当天的日志文件
	today := time.Now().Format("2006-01-02")
	logFile, err := os.OpenFile(fmt.Sprintf("logs/%s.log", today), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to create log file:", err)
	}

	infoLog = log.New(logFile, "INFO: ", log.Ldate|log.Ltime)
	errorLog = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime)
}

// Info 记录信息日志	
func Info(msg string) {
	fmt.Println(msg)
	infoLog.Println(msg)
}

// Infof 记录格式化信息日志
func Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(msg)
	infoLog.Println(msg)
}

// Error 记录错误日志
func Error(msg string) {
	fmt.Println("ERROR:", msg)
	errorLog.Println(msg)
}

// Errorf 记录格式化错误日志
func Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println("ERROR:", msg)
	errorLog.Println(msg)
}