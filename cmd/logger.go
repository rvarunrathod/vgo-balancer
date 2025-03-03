package main

import (
	"os"
	"sync"

	"go.uber.org/zap"

	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func CreateLogger(logPath string) *zap.Logger {
	var once sync.Once
	var logger *zap.Logger

	once.Do(func() {
		stdout := zapcore.AddSync(os.Stdout)

		file := zapcore.AddSync(&lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    200, // megabytes
			MaxBackups: 3,
			MaxAge:     30, // days
		})

		level := zap.NewAtomicLevelAt(zap.InfoLevel)

		productionCfg := zap.NewProductionEncoderConfig()
		productionCfg.TimeKey = "timestamp"
		productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

		developmentCfg := zap.NewDevelopmentEncoderConfig()
		developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

		var consoleEncoder zapcore.Encoder
		if os.Getenv("APP_ENV") == "dev" {
			consoleEncoder = zapcore.NewConsoleEncoder(developmentCfg)
		} else {
			consoleEncoder = zapcore.NewConsoleEncoder(productionCfg)
		}
		fileEncoder := zapcore.NewJSONEncoder(productionCfg)

		core := zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, stdout, level),
			zapcore.NewCore(fileEncoder, file, level),
		)
		logger = zap.New(core)
	})

	return logger
}
