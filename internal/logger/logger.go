package logger

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// InitLogger 初始化日志
func InitLogger(dataPath string, production bool) {
	// 设置日志级别
	if production {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// 设置日志格式
	if production {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// 创建日志目录
	logDir := filepath.Join(dataPath, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		logrus.Warn("创建日志目录失败:", err)
		return
	}

	// 创建日志文件
	logFile := filepath.Join(logDir, "filecodebox.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Warn("打开日志文件失败:", err)
		return
	}

	// 设置输出到文件和控制台
	if production {
		logrus.SetOutput(file)
	} else {
		// 开发模式输出到控制台和文件
		logrus.SetOutput(os.Stdout)
		// 可以添加文件输出的hook
	}
}
