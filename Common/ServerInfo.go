package Common

import (
	"strings"

	"github.com/gookit/slog"
)

type DBInfo struct {
	Name            string
	IP              string
	Port            int16
	User            string
	Password        string
	ConcurTaskCount int
	MaxOpenConns    int
	MaxIdleConns    int
}

type LogConfigInfo struct {
	//Error, Warn, Notice, Info, Debug, Trace
	LogLevel        string
	LogFilePathName string
	FileRotateSize  uint64
}

func ConvertLogLevel(str string) slog.Level {
	strLower := strings.ToLower(str)
	switch strLower {
	case "error":
		return slog.ErrorLevel
	case "warn":
		return slog.WarnLevel
	case "notice":
		return slog.NoticeLevel
	case "info":
		return slog.InfoLevel
	case "debug":
		return slog.DebugLevel
	case "trace":
		return slog.TraceLevel
	}

	return slog.DebugLevel
}
