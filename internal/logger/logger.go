package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go-devops/internal/config"
)

var Logger *logrus.Logger

// 初始化日志
func Init() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("加载配置失败，使用默认日志配置: %v\n", err)
		initWithDefaults()
		return
	}

	Logger = logrus.New()

	// 设置日志级别
	level := parseLogLevel(cfg.Logging.Level)
	Logger.SetLevel(level)

	// 设置日志格式
	Logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		PrettyPrint:     false,
	})

	// 根据配置设置输出
	if cfg.Logging.ToFile {
		// 创建日志目录
		logDir := strings.TrimSuffix(cfg.Logging.FilePath, "/")
		if logDir == "" {
			logDir = "logs"
		}
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Printf("创建日志目录失败: %v\n", err)
			Logger.SetOutput(os.Stdout)
		} else {
			// 创建日志文件
			logFile := filepath.Join(logDir, fmt.Sprintf("app-%s.log", time.Now().Format("2006-01-02")))
			file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				fmt.Printf("打开日志文件失败: %v\n", err)
				Logger.SetOutput(os.Stdout)
			} else {
				// 同时输出到文件和控制台
				multiWriter := io.MultiWriter(os.Stdout, file)
				Logger.SetOutput(multiWriter)
			}
		}
	} else {
		// 只输出到控制台
		Logger.SetOutput(os.Stdout)
	}

	Logger.Info("日志系统初始化成功")
}

// 使用默认配置初始化日志
func initWithDefaults() {
	Logger = logrus.New()
	Logger.SetLevel(logrus.InfoLevel)
	Logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		PrettyPrint:     false,
	})
	Logger.SetOutput(os.Stdout)
}

// 解析日志级别
func parseLogLevel(level string) logrus.Level {
	switch strings.ToLower(level) {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn", "warning":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	default:
		return logrus.InfoLevel
	}
}

// 记录信息日志
func Info(args ...interface{}) {
	Logger.Info(args...)
}

// 记录信息日志（格式化）
func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

// 记录警告日志
func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

// 记录警告日志（格式化）
func Warnf(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}

// 记录错误日志
func Error(args ...interface{}) {
	Logger.Error(args...)
}

// 记录错误日志（格式化）
func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

// 记录调试日志
func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

// 记录调试日志（格式化）
func Debugf(format string, args ...interface{}) {
	Logger.Debugf(format, args...)
}

// 记录致命错误日志
func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

// 记录致命错误日志（格式化）
func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}

// 带字段的日志记录
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Logger.WithFields(fields)
}

// 记录HTTP请求日志
func LogRequest(method, path, clientIP string, statusCode int, latency time.Duration, userAgent string) {
	Logger.WithFields(logrus.Fields{
		"method":      method,
		"path":        path,
		"client_ip":   clientIP,
		"status_code": statusCode,
		"latency":     latency.String(),
		"user_agent":  userAgent,
		"type":        "http_request",
	}).Info("HTTP请求")
}

// 记录用户操作日志
func LogUserAction(userID uint, username, action, resource string, success bool, details string) {
	Logger.WithFields(logrus.Fields{
		"user_id":  userID,
		"username": username,
		"action":   action,
		"resource": resource,
		"success":  success,
		"details":  details,
		"type":     "user_action",
	}).Info("用户操作")
}

// 记录数据库操作日志
func LogDBOperation(operation, table string, success bool, error string) {
	Logger.WithFields(logrus.Fields{
		"operation": operation,
		"table":     table,
		"success":   success,
		"error":     error,
		"type":      "db_operation",
	}).Info("数据库操作")
}

// 记录系统事件日志
func LogSystemEvent(event, component string, details interface{}) {
	Logger.WithFields(logrus.Fields{
		"event":     event,
		"component": component,
		"details":   details,
		"type":      "system_event",
	}).Info("系统事件")
}
