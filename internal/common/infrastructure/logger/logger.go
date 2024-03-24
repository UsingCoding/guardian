package logger

import (
	"time"

	"github.com/sirupsen/logrus"
)

type Fields logrus.Fields

type Logger interface {
	WithFields(Fields) Logger

	Info(...interface{})
	Infof(format string, args ...interface{})
	Error(error, ...interface{})
	Errorf(err error, format string, args ...interface{})
}

const appNameKey = "appID"

type Config struct {
	AppID string
}

func NewLogger(config Config) Logger {
	impl := logrus.New()
	impl.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap:        fieldMap,
	})

	return &logger{
		FieldLogger: impl.WithField(appNameKey, config.AppID),
	}
}

type logger struct {
	logrus.FieldLogger
}

var fieldMap = logrus.FieldMap{
	logrus.FieldKeyTime: "@timestamp",
	logrus.FieldKeyMsg:  "message",
}

func (l *logger) WithFields(fields Fields) Logger {
	return &logger{l.FieldLogger.WithFields(logrus.Fields(fields))}
}

func (l *logger) Error(err error, args ...interface{}) {
	l.FieldLogger.WithError(err).Error(args...)
}

func (l *logger) Errorf(err error, format string, args ...interface{}) {
	l.FieldLogger.WithError(err).Errorf(format, args...)
}

func (l *logger) FatalError(err error, args ...interface{}) {
	l.FieldLogger.WithError(err).Fatal(args...)
}
