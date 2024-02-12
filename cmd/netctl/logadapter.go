package main

import (
	"context"

	"github.com/charmbracelet/log"
	sqldblogger "github.com/simukti/sqldb-logger"
)

type logadapter struct {
	logger *log.Logger
}

func (l logadapter) Log(_ context.Context, level sqldblogger.Level, msg string, data map[string]interface{}) {
	keyvals := []interface{}{}
	for k, v := range data {
		keyvals = append(keyvals, k, v)
	}

	switch level {
	case sqldblogger.LevelError:
		l.logger.Error(msg, keyvals...)
	case sqldblogger.LevelInfo:
		l.logger.Info(msg, keyvals...)
	case sqldblogger.LevelDebug:
		l.logger.Debug(msg, keyvals...)
	case sqldblogger.LevelTrace:
		l.logger.Debug(msg, keyvals...)
	default:
		l.logger.Debug(msg, keyvals)
	}
}
