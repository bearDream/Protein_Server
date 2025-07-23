package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel 定义日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Logger 全局日志结构体
type Logger struct {
	level LogLevel
}

var (
	// GlobalLogger 全局日志实例
	GlobalLogger *Logger
)

// 初始化全局日志
func init() {
	GlobalLogger = NewLogger(INFO)
}

// NewLogger 创建新的日志实例
func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level: level,
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// getLevelString 获取日志级别字符串
func (l *Logger) getLevelString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// formatMessage 格式化日志消息
func (l *Logger) formatMessage(level LogLevel, message string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("[%s] [%s] %s", timestamp, l.getLevelString(level), message)
}

// shouldLog 判断是否应该输出日志
func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.level
}

// Debug 输出调试日志
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.shouldLog(DEBUG) {
		message := fmt.Sprintf(format, args...)
		log.Println(l.formatMessage(DEBUG, message))
	}
}

// Info 输出信息日志
func (l *Logger) Info(format string, args ...interface{}) {
	if l.shouldLog(INFO) {
		message := fmt.Sprintf(format, args...)
		log.Println(l.formatMessage(INFO, message))
	}
}

// Warn 输出警告日志
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.shouldLog(WARN) {
		message := fmt.Sprintf(format, args...)
		log.Println(l.formatMessage(WARN, message))
	}
}

// Error 输出错误日志
func (l *Logger) Error(format string, args ...interface{}) {
	if l.shouldLog(ERROR) {
		message := fmt.Sprintf(format, args...)
		log.Println(l.formatMessage(ERROR, message))
	}
}

// Fatal 输出致命错误日志并退出程序
func (l *Logger) Fatal(format string, args ...interface{}) {
	if l.shouldLog(FATAL) {
		message := fmt.Sprintf(format, args...)
		log.Println(l.formatMessage(FATAL, message))
		os.Exit(1)
	}
}

// 全局函数，方便直接调用
func Debug(format string, args ...interface{}) {
	GlobalLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	GlobalLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	GlobalLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	GlobalLogger.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	GlobalLogger.Fatal(format, args...)
}

// SetLogLevel 设置全局日志级别
func SetLogLevel(level LogLevel) {
	GlobalLogger.SetLevel(level)
} 