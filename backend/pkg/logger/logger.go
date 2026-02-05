package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger
var sugar *zap.SugaredLogger

// Config 日志配置
type Config struct {
	Level      string
	Format     string
	Output     string
	FilePath   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
}

// Init 初始化日志
func Init(cfg *Config) error {
	level := parseLevel(cfg.Level)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if cfg.Format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	var writeSyncer zapcore.WriteSyncer
	if cfg.Output == "file" && cfg.FilePath != "" {
		file, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		writeSyncer = zapcore.AddSync(file)
	} else {
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	core := zapcore.NewCore(encoder, writeSyncer, level)
	log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	sugar = log.Sugar()

	return nil
}

func parseLevel(levelStr string) zapcore.Level {
	switch levelStr {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// GetLogger 获取Logger实例
func GetLogger() *zap.Logger {
	if log == nil {
		Init(&Config{
			Level:  "info",
			Format: "console",
			Output: "stdout",
		})
	}
	return log
}

// GetSugar 获取SugaredLogger实例
func GetSugar() *zap.SugaredLogger {
	if sugar == nil {
		GetLogger()
	}
	return sugar
}

// Sync 同步日志
func Sync() error {
	if log != nil {
		return log.Sync()
	}
	return nil
}

// Debug 调试日志
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal 致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// WithFields 添加字段
func WithFields(fields ...zap.Field) *zap.Logger {
	return GetLogger().With(fields...)
}

// String 字符串字段
func String(key, val string) zap.Field {
	return zap.String(key, val)
}

// Int 整数字段
func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

// Int64 整数字段
func Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

// Float64 浮点数字段
func Float64(key string, val float64) zap.Field {
	return zap.Float64(key, val)
}

// Bool 布尔字段
func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

// Duration 时间间隔字段
func Duration(key string, val time.Duration) zap.Field {
	return zap.Duration(key, val)
}

// Time 时间字段
func Time(key string, val time.Time) zap.Field {
	return zap.Time(key, val)
}

// Any 任意类型字段
func Any(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}

// Err 错误字段
func Err(err error) zap.Field {
	return zap.Error(err)
}
