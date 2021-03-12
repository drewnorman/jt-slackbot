package logging

import (
	"errors"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"testing"
)

func TestNewLogger(t *testing.T) {
	type args struct {
		params        *LoggerParameters
		expectedLevel zapcore.Level
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "ReturnsInfoLoggerByDefault",
			args: args{
				params:        &LoggerParameters{},
				expectedLevel: zapcore.InfoLevel,
			},
		},
		{
			name: "ReturnsLoggerWithLevel",
			args: args{
				params: &LoggerParameters{
					Level: zapcore.ErrorLevel,
				},
				expectedLevel: zapcore.ErrorLevel,
			},
		},
		{
			name: "MultipleWriters",
			args: args{
				params: &LoggerParameters{
					Level: zapcore.ErrorLevel,
					Writers: []io.Writer{
						os.Stdout,
						os.Stderr,
					},
				},
				expectedLevel: zapcore.ErrorLevel,
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := NewLogger(*tt.args.params)
				if !got.Core().Enabled(tt.args.expectedLevel) {
					t.Errorf(
						"NewLogger() error = %v, wantErr %v",
						errors.New("expected level not enabled"),
						false,
					)
				}
			},
		)
	}
}
