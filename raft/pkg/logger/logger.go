package logger

import (
	"fmt"
	"strings"
)

type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Error
	Quiet // To be used to silence all the logs from a node
)

type Logger struct {
	indent string
	level  LogLevel
}

func NewLogger(indent string, index int64, lvl *LogLevel) *Logger {
	level := Debug
	if lvl != nil {
		level = *lvl
	}

	return &Logger{
		indent: indent,
		level:  level,
	}
}

// Printf prints the formatted log with indentation
func (l *Logger) printf(format string, lvl LogLevel, args ...any) {
	if lvl < l.level {
		return
	}

	message := fmt.Sprintf(format, args...)
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		if line != "" {
			fmt.Printf("%s%s\n", l.indent, line)
		}
	}
}

func (l *Logger) Debug(format string, args ...any) {
	l.printf(format, Debug, args...)
}

func (l *Logger) Info(format string, args ...any) {
	l.printf(format, Info, args...)
}

func (l *Logger) Error(format string, args ...any) {
	l.printf(format, Error, args...)
}
