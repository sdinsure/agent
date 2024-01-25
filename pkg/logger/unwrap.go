package logger

import "go.uber.org/zap"

func GetUnderlyingZapLoggerOrDie(l Logger) *zap.Logger {
	_, isLoggerImpl := l.(*loggerImpl)
	if !isLoggerImpl {
		panic("not loggerImpl type")
	}
	return l.(*loggerImpl).Logger
}
