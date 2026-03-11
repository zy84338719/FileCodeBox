package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

type Config struct {
	Level      string
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

func Init(cfg *Config) error {
	level := getLevel(cfg.Level)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var core zapcore.Core
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)

	if cfg.Filename != "" {
		file, err := os.OpenFile(cfg.Filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		fileCore := zapcore.NewCore(encoder, zapcore.AddSync(file), level)
		core = zapcore.NewTee(consoleCore, fileCore)
	} else {
		core = consoleCore
	}

	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return nil
}

func getLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}

func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

func With(fields ...zap.Field) *zap.Logger {
	return Logger.With(fields...)
}
