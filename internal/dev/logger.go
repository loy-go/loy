package dev

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
)

// ANSI color escape codes for interleaving terminal logs.
var colors = []string{
	"\033[36m", // Cyan
	"\033[32m", // Green
	"\033[33m", // Yellow
	"\033[35m", // Magenta
	"\033[34m", // Blue
}

const colorReset = "\033[0m"

// PrefixedLogger manages colorized, thread-safe line prefixing for multiple concurrent processes.
type PrefixedLogger struct {
	out        io.Writer
	mu         sync.Mutex
	colorIndex int
	tagColors  map[string]string
	noColor    bool
}

// NewPrefixedLogger creates a new PrefixedLogger.
func NewPrefixedLogger(out io.Writer, noColor bool) *PrefixedLogger {
	return &PrefixedLogger{
		out:       out,
		tagColors: make(map[string]string),
		noColor:   noColor,
	}
}

// getColor resolves a distinct color for a given task tag.
func (l *PrefixedLogger) getColor(tag string) string {
	if l.noColor {
		return ""
	}
	if c, ok := l.tagColors[tag]; ok {
		return c
	}
	color := colors[l.colorIndex%len(colors)]
	l.colorIndex++
	l.tagColors[tag] = color
	return color
}

// LogLine prints a single line with tag prefix.
func (l *PrefixedLogger) LogLine(tag, text string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	color := l.getColor(tag)
	if color != "" {
		_, _ = fmt.Fprintf(l.out, "%s[%s]%s %s\n", color, tag, colorReset, text)
	} else {
		_, _ = fmt.Fprintf(l.out, "[%s] %s\n", tag, text)
	}
}

// PipeStream copies an io.Reader line-by-line through the prefixed logger until EOF.
func (l *PrefixedLogger) PipeStream(tag string, r io.Reader) {
	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			trimmed := strings.TrimRight(line, "\r\n")
			l.LogLine(tag, trimmed)
		}
		if err != nil {
			if err != io.EOF {
				l.LogLine(tag, fmt.Sprintf("Error reading stream: %v", err))
			}
			return
		}
	}
}
