package logger

import (
	"fmt"
	"strings"
)

type Logger struct {
	Indent string
	Color  string
}

func NewLogger(indent string, index int64) *Logger {
	colorIdx := int(index) % len(allColours)
	color := allColours[colorIdx]
	return &Logger{
		Indent: indent,
		Color:  color,
	}
}

// Printf prints the formatted log with indentation and color
func (l *Logger) Printf(format string, args ...any) {
	col := colors[l.Color]
	if col == "" {
		col = colors["reset"]
	}

	message := fmt.Sprintf(format, args...)
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		if line != "" {
			fmt.Printf("%s%s%s%s\n", col, l.Indent, line, colors["reset"])
		}
	}
}
