package mylog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"time"
)

var zapLogger *zap.Logger
var zapSugarLogger *zap.SugaredLogger
var encoderConfig = zapcore.EncoderConfig{
	TimeKey:        "timer",
	LevelKey:       "level",
	NameKey:        "name",
	CallerKey:      "caller",
	MessageKey:     "message",
	StacktraceKey:  "stacktrace",
	LineEnding:     "\n",
	EncodeLevel:    zapcore.CapitalColorLevelEncoder,
	EncodeTime:     encodeTime, //zapcore.ISO8601TimeEncoder,
	EncodeDuration: zapcore.StringDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

var fileOut = &lumberjack.Logger{
	Filename:   "./logs/mylog.log", // log path
	MaxSize:    100,                // log file size, M
	MaxBackups: 30,                 // backups num
	MaxAge:     7,                  // log save days
	LocalTime:  true,
	Compress:   false,
}

func init() {
	var err error
	zapConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development:       false,
		DisableStacktrace: true,
		Encoding:          "console", //"json",
		EncoderConfig:     encoderConfig,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stdout"},
		DisableCaller:     false,
	}
	zapLogger, err = zapConfig.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
	zapSugarLogger = zapLogger.Sugar()
}

func encodeTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.999999999"))
}

func InitMyLog(out *lumberjack.Logger) {
	// log cutting
	if out != nil {
		fileOut = out
	}
	// zap log
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(fileOut),
			zapcore.AddSync(os.Stdout),
		),
		zap.NewAtomicLevelAt(zapcore.DebugLevel),
	)
	// log
	caller := zap.AddCaller()
	zapLogger = zap.New(core, caller, zap.AddCallerSkip(1))
	zapSugarLogger = zapLogger.Sugar()
}
