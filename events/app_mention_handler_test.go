package events

import (
	"github.com/brianvoe/gofakeit/v6"
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
					SlackHttpClient: fakeSlackHttpClient(
						t,
						map[string]interface{}{},
					),
				},
			},
			wantErr: false,
		},
		{
			name: "MissingSlackHttpClient",
			args: args{
				params: &AppMentionHandlerParameters{
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

func TestAppMentionHandler_Process(t *testing.T) {
	type args struct {
		eventData map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ProcessesAppMentionEvent",
			args: args{
				eventData: map[string]interface{}{
					"authorizations": []interface{}{
						map[string]interface{}{
							"user_id": gofakeit.UUID(),
						},
					},
					"event": map[string]interface{}{
						"channel": gofakeit.UUID(),
						"text":    gofakeit.LoremIpsumSentence(20),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "IllFormattedEventData",
			args: args{
				eventData: map[string]interface{}{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &AppMentionHandler{
				SlackHttpClient: fakeSlackHttpClient(
					t,
					map[string]interface{}{
						"ok": true,
					},
				),
			}
			err := handler.Process(tt.args.eventData)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Process() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}
