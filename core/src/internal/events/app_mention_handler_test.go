package events

import (
	"testing"
)

func TestNewAppMentionHandler(t *testing.T) {
	type args struct {
		params *AppMentionHandlerParameters
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ReturnsNewHandler",
			args: args{
				params: &AppMentionHandlerParameters{
					Logger: fakeZapLogger(),
					SlackHttpClient: fakeSlackHttpClient(
						t,
						map[string]interface{}{},
					),
				},
			},
			wantErr: false,
		},
		{
			name: "MissingLogger",
			args: args{
				params: &AppMentionHandlerParameters{
					Logger: nil,
					SlackHttpClient: fakeSlackHttpClient(
						t,
						map[string]interface{}{},
					),
				},
			},
			wantErr: true,
		},
		{
			name: "MissingSlackHttpClient",
			args: args{
				params: &AppMentionHandlerParameters{
					Logger:          fakeZapLogger(),
					SlackHttpClient: nil,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAppMentionHandler(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"NewAppMentionHandler() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}
