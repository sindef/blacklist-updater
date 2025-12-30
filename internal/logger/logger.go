package logger

import (
	"fmt"
	"os"
	"time"
)

type Logger struct {
	debug bool
}

func New(debug bool) *Logger {
	return &Logger{debug: debug}
}

func (l *Logger) Info(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[ERROR] "+format+"\n", args...)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if l.debug {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

func (l *Logger) Timestamp() string {
	return time.Now().Format(time.RFC3339)
}
