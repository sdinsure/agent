package storagepostgres

import (
	logger "github.com/sdinsure/agent/pkg/logger"
)

func newGlogWriter(log logger.Logger) *GlogWriter {
	return &GlogWriter{log: log}
}

// GlogWriter implements gorm.logger.Writer interface
type GlogWriter struct {
	log logger.Logger
}

func (g *GlogWriter) Printf(format string, values ...interface{}) {
	g.log.Info(format, values...)
}
