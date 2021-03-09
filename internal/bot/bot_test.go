package bot

import (
	"github.com/brianvoe/gofakeit/v6"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"testing"
)

func fakeZapLogger() *zap.Logger {
	return zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(
				zap.NewProductionEncoderConfig(),
			),
			zapcore.AddSync(
				os.NewFile(0, os.DevNull),
			),
			zap.FatalLevel,
		),
	)
}

func TestNew(t *testing.T) {
	type args struct {
		params *Parameters
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ReturnsNewBot",
			args: args{
				params: &Parameters{
					Logger:   fakeZapLogger(),
					ApiUrl:   gofakeit.URL(),
					AppToken: gofakeit.UUID(),
					BotToken: gofakeit.UUID(),
				},
			},
			wantErr: false,
		},
		{
			name: "MissingLogger",
			args: args{
				params: &Parameters{
					ApiUrl:   gofakeit.URL(),
					AppToken: gofakeit.UUID(),
					BotToken: gofakeit.UUID(),
				},
			},
			wantErr: true,
		},
		{
			name: "MissingApiUrl",
			args: args{
				params: &Parameters{
					Logger:   fakeZapLogger(),
					AppToken: gofakeit.UUID(),
					BotToken: gofakeit.UUID(),
				},
			},
			wantErr: true,
		},
		{
			name: "MissingAppToken",
			args: args{
				params: &Parameters{
					Logger:   fakeZapLogger(),
					ApiUrl:   gofakeit.URL(),
					BotToken: gofakeit.UUID(),
				},
			},
			wantErr: true,
		},
		{
			name: "MissingBotToken",
			args: args{
				params: &Parameters{
					Logger:   fakeZapLogger(),
					ApiUrl:   gofakeit.URL(),
					AppToken: gofakeit.UUID(),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
