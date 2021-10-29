package logger

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

type Level log.Level

var ErrInvalidLevel = errors.New("invalid level")

// levels for logging.
const (
	LevelInfo  = "INFO"
	LevelError = "ERROR"
	LevelFatal = "FATAL"
)

var levels = map[string]log.Level{
	LevelInfo:  log.InfoLevel,
	LevelError: log.ErrorLevel,
	LevelFatal: log.FatalLevel,
}

func NewLevel(name string) (Level, error) {
	level, ok := levels[name]

	if !ok {
		return 0, ErrInvalidLevel
	}

	return Level(level), nil
}
