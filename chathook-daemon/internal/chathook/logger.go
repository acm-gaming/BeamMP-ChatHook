package chathook

import "github.com/charmbracelet/log"

func ApplyLogLevel(logger *log.Logger, level string) {
	switch level {
	case "debug":
		logger.SetLevel(log.DebugLevel)
	case "warn", "warning":
		logger.SetLevel(log.WarnLevel)
	case "error":
		logger.SetLevel(log.ErrorLevel)
	default:
		logger.SetLevel(log.InfoLevel)
	}
}
