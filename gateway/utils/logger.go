package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var (
	logger      *log.Logger
	logFile     *os.File
	logLevel    LogLevel
	logFilePath string
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// InitLogger 初始化日志记录器
func InitLogger(logPath string, logLev string) error {

	// 创建日志目录
	logDir := logPath
	if logDir == "" {
		logDir = "logs"
	}
	if logLev == "" {
		logLev = "INFO"
	}
	logLevel = getLogLevelFromString(logLev)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// 创建日志文件
	logFileName := fmt.Sprintf("app_%s.log", time.Now().Format("2006-01-02"))
	logFilePath = filepath.Join(logDir, logFileName)

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	// 创建日志记录器
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger = log.New(multiWriter, "", 0)

	Info("Logger initialized successfully")
	return nil
}

// CloseLogger 关闭日志记录器
func CloseLogger() {
	if logFile != nil {
		Info("Closing logger")
		logFile.Close()
	}
}

// getLogLevelFromString 将字符串转换为日志级别
func getLogLevelFromString(level string) LogLevel {
	switch level {
	case "DEBUG":
		return DEBUG
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

// logWithLevel 根据日志级别记录日志
func logWithLevel(level LogLevel, format string, v ...interface{}) {
	if logger == nil || level < logLevel {
		return
	}

	// 获取调用者信息
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	} else {
		// 简化文件路径，只保留文件名
		file = filepath.Base(file)
	}

	// 格式化日志消息
	message := fmt.Sprintf(format, v...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] [%s] [%s:%d] %s", timestamp, level.String(), file, line, message)

	logger.Println(logEntry)
}

// Debug 记录调试日志
func Debug(format string, v ...interface{}) {
	logWithLevel(DEBUG, format, v...)
}

// Info 记录信息日志
func Info(format string, v ...interface{}) {
	logWithLevel(INFO, format, v...)
}

// Warn 记录警告日志
func Warn(format string, v ...interface{}) {
	logWithLevel(WARN, format, v...)
}

// Error 记录错误日志
func Error(format string, v ...interface{}) {
	logWithLevel(ERROR, format, v...)
}
