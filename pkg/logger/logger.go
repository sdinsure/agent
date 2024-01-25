package logger

import "context"

type Logger interface {
	Info(fmtStr string, values ...interface{})
	Warn(fmtStr string, values ...interface{})
	Error(fmtStr string, values ...interface{})
	Fatal(fmtStr string, values ...interface{})

	Infox(ctx context.Context, fmtStr string, values ...interface{})
	Warnx(ctx context.Context, fmtStr string, values ...interface{})
	Errorx(ctx context.Context, fmtStr string, values ...interface{})
}
