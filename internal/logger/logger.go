package logger

import (
	"log"
)

type Logger struct {
	debug bool
}

func New(debug bool) *Logger {
	return &Logger{debug: debug}
}

func (l *Logger) Info(format string, v ...any) {
	log.Printf("[INFO] "+format, v...)
}

func (l *Logger) Error(format string, v ...any) {
	log.Printf("[ERROR] "+format, v...)
}

func (l *Logger) Debug(format string, v ...any) {
	if l.debug {
		log.Printf("[DEBUG] "+format, v...)
	}
}
