package main

import (
	"fmt"
	"os"
)

type Logger struct {
	Level int
}

const (
	DEBUG = iota
	INFO
	WARN
	ERROR
)

func (l *Logger) Error(s ...interface{}) {
	if l.Level <= ERROR {
		str := fmt.Sprint(s...)
		fmt.Fprintln(os.Stderr, "ERROR", str)
	}
}

func (l *Logger) Debug(s ...interface{}) {
	if l.Level <= DEBUG {
		str := fmt.Sprint(s...)
		fmt.Fprintln(os.Stderr, "DEBUG", str)
	}
}

func (l *Logger) Info(s ...interface{}) {
	if l.Level <= INFO {
		str := fmt.Sprint(s...)
		fmt.Fprintln(os.Stderr, "INFO", str)
	}
}
