package logger

import (
    "os"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// 全局 Logger 实例
var Log *zap.Logger

// Config 日志配置
type Config struct {
    Level      string // debug, info, warn, error
    Encoding   string // console, json
    OutputPath string // 日志文件路径，空表示只输出到控制台
}

// Init 初始化日志
func Init(cfg Config) error {
    // 日志级别
    var level zapcore.Level
    switch cfg.Level {
    case "debug":
        level = zapcore.DebugLevel
    case "info":
        level = zapcore.InfoLevel
    case "warn":
        level = zapcore.WarnLevel
    case "error":
        level = zapcore.ErrorLevel
    default:
        level = zapcore.InfoLevel
    }

    // 编码器配置
    encoderConfig := zapcore.EncoderConfig{
        TimeKey:        "time",
        LevelKey:       "level",
        NameKey:        "logger",
        CallerKey:      "caller",
        FunctionKey:    zapcore.OmitKey,
        MessageKey:     "msg",
        StacktraceKey:  "stacktrace",
        LineEnding:     zapcore.DefaultLineEnding,
        EncodeLevel:    zapcore.CapitalLevelEncoder,
        EncodeTime:     zapcore.ISO8601TimeEncoder,
        EncodeDuration: zapcore.SecondsDurationEncoder,
        EncodeCaller:   zapcore.ShortCallerEncoder,
    }

    // 编码器
    var encoder zapcore.Encoder
    if cfg.Encoding == "json" {
        encoder = zapcore.NewJSONEncoder(encoderConfig)
    } else {
        encoder = zapcore.NewConsoleEncoder(encoderConfig)
    }

    // 输出目标
    cores := []zapcore.Core{
        zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level),
    }

    // 文件输出
    if cfg.OutputPath != "" {
        file, err := os.OpenFile(cfg.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if err != nil {
            return err
        }
        cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(file), level))
    }

    // 合并输出
    core := zapcore.NewTee(cores...)

    // 创建 Logger
    Log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

    return nil
}

// Sync 刷新日志缓冲区
func Sync() {
    if Log != nil {
        _ = Log.Sync()
    }
}

// 直接使用 zap 方法，不再封装
// 调用方式：logger.Log.Info("msg", zap.String("key", "value"))