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
	ERROR
)

func (l *Logger) Error(fmtstr string, s ...interface{}) {
	if l.Level <= ERROR {
		str := fmt.Sprintf(fmtstr, s...)
		fmt.Fprintln(os.Stderr, "ERROR", str)
	}
}

func (l *Logger) Debug(fmtstr string, s ...interface{}) {
	if l.Level <= DEBUG {
		str := fmt.Sprintf(fmtstr, s...)
		fmt.Fprintln(os.Stderr, str)
	}
}

func (l *Logger) Info(fmtstr string, s ...interface{}) {
	if l.Level <= INFO {
		str := fmt.Sprintf(fmtstr, s...)
		fmt.Fprintln(os.Stderr, str)
	}
}
