package main

import (
	"github.com/brianvoe/gofakeit/v6"
	"os"
	"strconv"
	"testing"
)

func TestLoadConfiguration(t *testing.T) {
	type args struct {
		environment map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "LoadsConfiguration",
			args: args{
				environment: map[string]string{
					"SLACK_API_URL":              gofakeit.URL(),
					"SLACK_APP_TOKEN":            gofakeit.UUID(),
					"SLACK_BOT_TOKEN":            gofakeit.UUID(),
					"MAX_CONNECT_ATTEMPTS":       strconv.Itoa(gofakeit.Number(0, 4)),
					"DEBUG_WEBSOCKET_RECONNECTS": "false",
				},
			},
			wantErr: false,
		},
		{
			name: "MissingSlackApiUrl",
			args: args{
				environment: map[string]string{
					"SLACK_APP_TOKEN":            gofakeit.UUID(),
					"SLACK_BOT_TOKEN":            gofakeit.UUID(),
					"MAX_CONNECT_ATTEMPTS":       strconv.Itoa(gofakeit.Number(0, 4)),
					"DEBUG_WEBSOCKET_RECONNECTS": "false",
				},
			},
			wantErr: true,
		},
		{
			name: "MissingSlackAppToken",
			args: args{
				environment: map[string]string{
					"SLACK_API_URL":              gofakeit.URL(),
					"SLACK_BOT_TOKEN":            gofakeit.UUID(),
					"MAX_CONNECT_ATTEMPTS":       strconv.Itoa(gofakeit.Number(0, 4)),
					"DEBUG_WEBSOCKET_RECONNECTS": "false",
				},
			},
			wantErr: true,
		},
		{
			name: "MissingSlackBotToken",
			args: args{
				environment: map[string]string{
					"SLACK_API_URL":              gofakeit.URL(),
					"SLACK_APP_TOKEN":            gofakeit.UUID(),
					"MAX_CONNECT_ATTEMPTS":       strconv.Itoa(gofakeit.Number(0, 4)),
					"DEBUG_WEBSOCKET_RECONNECTS": "false",
				},
			},
			wantErr: true,
		},
		{
			name: "MissingMaxConnectAttempts",
			args: args{
				environment: map[string]string{
					"SLACK_API_URL":              gofakeit.URL(),
					"SLACK_BOT_TOKEN":            gofakeit.UUID(),
					"SLACK_APP_TOKEN":            gofakeit.UUID(),
					"DEBUG_WEBSOCKET_RECONNECTS": "false",
				},
			},
			wantErr: false,
		},
		{
			name: "MissingDebugWebsocketReconnects",
			args: args{
				environment: map[string]string{
					"SLACK_API_URL":   gofakeit.URL(),
					"SLACK_BOT_TOKEN": gofakeit.UUID(),
					"SLACK_APP_TOKEN": gofakeit.UUID(),
				},
			},
			wantErr: false,
		},
		{
			name: "DebugWebsocketReconnects",
			args: args{
				environment: map[string]string{
					"SLACK_API_URL":              gofakeit.URL(),
					"SLACK_BOT_TOKEN":            gofakeit.UUID(),
					"SLACK_APP_TOKEN":            gofakeit.UUID(),
					"DEBUG_WEBSOCKET_RECONNECTS": "true",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()

			for key, value := range tt.args.environment {
				err := os.Setenv(key, value)
				if err != nil {
					t.Errorf("LoadConfiguration() error = %v, wantErr %v", err, false)
				}
			}

			config := &Configuration{
				loadEnvironment: func(filenames ...string) (err error) {
					return nil
				},
			}

			err := config.Load()
			if err != nil {
				if !tt.wantErr {
					t.Errorf("LoadConfiguration() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if tt.args.environment["SLACK_API_URL"] != "" && config.apiUrl != tt.args.environment["SLACK_API_URL"] {
				t.Errorf("LoadConfiguration() = %v, want %v", config.apiUrl, tt.args.environment["SLACK_API_URL"])
			}

			if tt.args.environment["SLACK_APP_TOKEN"] != "" && config.appToken != tt.args.environment["SLACK_APP_TOKEN"] {
				t.Errorf("LoadConfiguration() = %v, want %v", config.appToken, tt.args.environment["SLACK_APP_TOKEN"])
			}

			if tt.args.environment["SLACK_BOT_TOKEN"] != "" && config.botToken != tt.args.environment["SLACK_BOT_TOKEN"] {
				t.Errorf("LoadConfiguration() = %v, want %v", config.botToken, tt.args.environment["SLACK_BOT_TOKEN"])
			}

			if tt.args.environment["MAX_CONNECT_ATTEMPTS"] != "" && strconv.Itoa(config.maxConnectAttempts) != tt.args.environment["MAX_CONNECT_ATTEMPTS"] {
				t.Errorf("LoadConfiguration() = %v, want %v", config.maxConnectAttempts, tt.args.environment["MAX_CONNECT_ATTEMPTS"])
			}

			if tt.args.environment["DEBUG_WEBSOCKET_RECONNECTS"] != "" && strconv.FormatBool(config.debugWssReconnects) != tt.args.environment["DEBUG_WEBSOCKET_RECONNECTS"] {
				t.Errorf("LoadConfiguration() = %v, want %v", config.debugWssReconnects, tt.args.environment["DEBUG_WEBSOCKET_RECONNECTS"])
			}
		})
	}
}
