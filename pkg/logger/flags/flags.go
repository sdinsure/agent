package loggerflags

import (
	"flag"
	"os"
)

var (
	logDir = new(string)
)

// logDirs contains a list of candidates of dir for log

func InitFlags() {
	logDir = flag.String("log_dir", "", "If non-empty, write log files in this directory")
}

func GetLogDir() string {
	if *logDir != "" {
		return *logDir
	}
	return os.TempDir()
}
