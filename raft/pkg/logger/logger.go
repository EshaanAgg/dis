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
)

type Logger struct {
	indent string
	color  string
	level  LogLevel
}

func NewLogger(indent string, index int64, lvl *LogLevel) *Logger {
	level := Debug
	if lvl != nil {
		level = *lvl
	}

	colorIdx := int(index) % len(allColours)
	color := allColours[colorIdx]
	return &Logger{
		indent: indent,
		color:  color,
		level:  level,
	}
}

// Printf prints the formatted log with indentation and color
func (l *Logger) printf(format string, lvl LogLevel, args ...any) {
	if lvl < l.level {
		return
	}

	col := colors[l.color]
	if col == "" {
		col = colors["reset"]
	}

	message := fmt.Sprintf(format, args...)
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		if line != "" {
			fmt.Printf("%s%s%s%s\n", col, l.indent, line, colors["reset"])
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
