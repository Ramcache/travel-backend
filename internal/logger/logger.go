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

	// Настраиваем lumberjack для ротации файлов // потом в config
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/app.log", // можно вынести в конфиг
		MaxSize:    50,             // мегабайты
		MaxBackups: 7,              // количество файлов
		MaxAge:     30,             // дни хранения
		Compress:   true,           // сжатие старых логов
	})

	consoleWriter := zapcore.Lock(os.Stdout)

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), fileWriter, level),
		zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()), consoleWriter, level),
	)

	log := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return log.Sugar()
}
