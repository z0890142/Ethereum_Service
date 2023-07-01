package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var defaultLogger *zap.Logger

func init() {
	l, _ := New(Options{
		Level:   "INFO",
		Outputs: []string{"stdout"},
	})

	SetLogger(l)
}

// Options for logger
type Options struct {
	Level   string
	Outputs []string
}

// New a logger from option
func New(o Options) (*zap.Logger, error) {
	lvl := new(zapcore.Level)
	if err := lvl.Set(o.Level); err != nil {
		return nil, err
	}

	conf := zap.Config{
		Level:         zap.NewAtomicLevelAt(*lvl),
		Development:   false,
		DisableCaller: true,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "log_time",
			LevelKey:       "level",
			CallerKey:      "log_caller",
			MessageKey:     "message",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeDuration: zapcore.MillisDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		},
		OutputPaths:      o.Outputs,
		ErrorOutputPaths: o.Outputs,
	}

	return conf.Build(zap.AddCaller())
}

// SetLogger set logger
func SetLogger(l *zap.Logger) {
	defaultLogger = l
}

// GetLogger get logger
func GetLogger() *zap.Logger {
	return defaultLogger
}

// Debug level log
func Debug(msg string) {
	defer defaultLogger.Sync()
	defaultLogger.Debug(msg)
}

// Info level log
func Info(msg string) {
	defer defaultLogger.Sync()
	defaultLogger.Info(msg)
}

// Warn level log
func Warn(msg string) {
	defer defaultLogger.Sync()
	defaultLogger.Warn(msg)
}

// Error level log
func Error(msg string) {
	defer defaultLogger.Sync()
	defaultLogger.Error(msg)
}

// Errorf level log with format
func Errorf(format string, args ...interface{}) {
	defer defaultLogger.Sync()
	defaultLogger.Error(fmt.Sprintf(format, args...))
}

// Panic level log
func Panic(msg string) {
	defer defaultLogger.Sync()
	defaultLogger.Panic(msg)
}

// With log with extra fields
func With(args ...interface{}) *zap.Logger {
	return defaultLogger.Sugar().With(args...).Desugar()
}

// LoadExtra log with extra fields in map
func LoadExtra(extras map[string]interface{}) *zap.Logger {
	args := make([]interface{}, len(extras)*2)
	i := 0
	for k, v := range extras {
		args[i] = k
		args[i+1] = v
		i += 2
	}
	return With(args...)
}
