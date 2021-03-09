package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
)

// LoggerParameters specify how a new
// zap.Logger should be created.
type LoggerParameters struct {
	Level   zapcore.Level
	Writers []io.Writer
}

// NewLogger returns a new instance of
// zap.Logger according to the given
// parameters.
func NewLogger(params LoggerParameters) *zap.Logger {
	var writeSyncer zapcore.WriteSyncer

	if len(params.Writers) == 0 {
		writeSyncer = zapcore.AddSync(os.Stdout)
	} else {
		if len(params.Writers) == 1 {
			writeSyncer = zapcore.AddSync(
				params.Writers[0],
			)
		} else {
			var writeSyncers []zapcore.WriteSyncer
			for _, writer := range params.Writers {
				writeSyncers = append(
					writeSyncers,
					zapcore.AddSync(writer),
				)
			}
			writeSyncer = zapcore.NewMultiWriteSyncer(
				writeSyncers...,
			)
		}
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(
				encoderConfig,
			),
			writeSyncer,
			params.Level,
		),
	)
}
