package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	loggerflags "github.com/sdinsure/agent/pkg/logger/flags"
)

var (
	_ Logger = &loggerImpl{}
)

type loggerImpl struct {
	*zap.Logger
}

func (l *loggerImpl) Flush() error {
	return l.Logger.Sync()
}

func (l *loggerImpl) Info(fmtStr string, values ...interface{}) {
	l.Logger.Info(fmt.Sprintf(fmtStr, values...))
}

func (l *loggerImpl) Warn(fmtStr string, values ...interface{}) {
	l.Logger.Warn(fmt.Sprintf(fmtStr, values...))
}

func (l *loggerImpl) Error(fmtStr string, values ...interface{}) {
	l.Logger.Error(fmt.Sprintf(fmtStr, values...))
}

func (l *loggerImpl) Fatal(fmtStr string, values ...interface{}) {
	fmt.Printf("[FATAL] %s", fmt.Sprintf(fmtStr, values...))
	l.Logger.Fatal(fmt.Sprintf(fmtStr, values...))
}

func (l *loggerImpl) attachCtx(ctx context.Context) *loggerImpl {
	fields := []zapcore.Field{}
	tags := grpc_ctxtags.Extract(ctx)
	for k, v := range tags.Values() {
		fields = append(fields, zap.Any(k, v))
	}
	return &loggerImpl{Logger: l.Logger.With(fields...)}
}

func (l *loggerImpl) Infox(ctx context.Context, fmtStr string, values ...interface{}) {
	l.attachCtx(ctx).Info(fmtStr, values...)
}

func (l *loggerImpl) Warnx(ctx context.Context, fmtStr string, values ...interface{}) {
	l.attachCtx(ctx).Warn(fmtStr, values...)
}

func (l *loggerImpl) Errorx(ctx context.Context, fmtStr string, values ...interface{}) {
	l.attachCtx(ctx).Error(fmtStr, values...)
}

type LogTag interface {
	name() string
}

type AppLog struct{}

func (a AppLog) name() string {
	return "app.log"
}

func NewLogName(tag LogTag) string {
	name := fmt.Sprintf("%s.%s.%s.%s.%d",
		program,
		host,
		userName,
		tag.name(),
		pid)
	return name
}

func NewLogger() *loggerImpl {
	encConfig := zap.NewProductionEncoderConfig()
	encConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	infoLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.InfoLevel
	})
	errorFatalLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.ErrorLevel || level == zapcore.FatalLevel
	})
	core := zapcore.NewTee(
		// write info to stdout
		// and stderr/fatal to stderr
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encConfig),
			zapcore.Lock(os.Stdout),
			infoLevel,
		),
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encConfig),
			zapcore.Lock(os.Stderr),
			errorFatalLevel,
		),
		// also write all logs to applog file
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encConfig),
			zapcore.AddSync(&lumberjack.Logger{
				Filename:   filepath.Join(loggerflags.GetLogDir(), NewLogName(AppLog{})),
				MaxSize:    500, // megabytes
				MaxBackups: 3,
				MaxAge:     28, // days
			}),
			zap.InfoLevel,
		),
	)
	zapLogger := zap.New(core, zap.AddCallerSkip(2))
	return &loggerImpl{
		Logger: zapLogger,
	}
}
