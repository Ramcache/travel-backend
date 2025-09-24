package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func New(env string) *zap.SugaredLogger {
	var level zapcore.Level
	if env == "prod" {
		level = zapcore.InfoLevel
	} else {
		level = zapcore.DebugLevel
	}

	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    50,
		MaxBackups: 7,
		MaxAge:     30,
		Compress:   true,
	})

	consoleWriter := zapcore.Lock(os.Stdout)

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), fileWriter, level),
		zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()), consoleWriter, level),
	)

	log := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return log.Sugar()
}
