package config

import "github.com/charmbracelet/log"

type logAdapter struct {
	log.Logger
}

func (l logAdapter) Fatal(args ...interface{}) {
	if len(args) > 1 {
		l.Logger.Fatal(args[0], args[1:]...)
	} else {
		l.Logger.Fatal(args[0])
	}
}

func (l logAdapter) Fatalf(format string, args ...interface{}) {
	l.Logger.Fatalf(format, args...)
}

func (l logAdapter) Print(args ...interface{}) {
	if len(args) > 1 {
		l.Logger.Debug(args[0], args[1:]...)
	} else {
		l.Logger.Debug(args[0])
	}
}

func (l logAdapter) Printf(format string, args ...interface{}) {
	l.Logger.Debugf(format, args...)
}

func (l logAdapter) Println(args ...interface{}) {
	if len(args) > 1 {
		l.Logger.Debug(args[0], args[1:]...)
	} else {
		l.Logger.Debug(args[0])
	}
}
