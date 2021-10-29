package logger

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

type Logger struct{}

func New(file string, level Level) (*Logger, error) {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		return nil, fmt.Errorf("cant open file for logs: %w", err)
	}

	log.SetOutput(f)
	log.SetLevel(log.Level(level))

	return &Logger{}, nil
}

func (l Logger) Info(msg string) {
	log.Info(msg)
}

func (l Logger) Warning(msg string) {
	log.Warning(msg)
}

func (l Logger) WarningWithFields(msg string, fields map[string]interface{}) {
	log.WithFields(fields).Warning(msg)
}

func (l Logger) Error(msg string) {
	log.Error(msg)
}
