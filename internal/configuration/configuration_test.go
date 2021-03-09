package configuration

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
		{
			name: "MissingLogLevel",
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
			name: "DebugLogLevel",
			args: args{
				environment: map[string]string{
					"SLACK_API_URL":   gofakeit.URL(),
					"SLACK_BOT_TOKEN": gofakeit.UUID(),
					"SLACK_APP_TOKEN": gofakeit.UUID(),
					"LOG_LEVEL":       "debug",
				},
			},
			wantErr: false,
		},
		{
			name: "InfoLogLevel",
			args: args{
				environment: map[string]string{
					"SLACK_API_URL":   gofakeit.URL(),
					"SLACK_BOT_TOKEN": gofakeit.UUID(),
					"SLACK_APP_TOKEN": gofakeit.UUID(),
					"LOG_LEVEL":       "info",
				},
			},
			wantErr: false,
		},
		{
			name: "WarningErrorLevel",
			args: args{
				environment: map[string]string{
					"SLACK_API_URL":   gofakeit.URL(),
					"SLACK_BOT_TOKEN": gofakeit.UUID(),
					"SLACK_APP_TOKEN": gofakeit.UUID(),
					"LOG_LEVEL":       "warn",
				},
			},
			wantErr: false,
		},
		{
			name: "ErrorLogLevel",
			args: args{
				environment: map[string]string{
					"SLACK_API_URL":   gofakeit.URL(),
					"SLACK_BOT_TOKEN": gofakeit.UUID(),
					"SLACK_APP_TOKEN": gofakeit.UUID(),
					"LOG_LEVEL":       "error",
				},
			},
			wantErr: false,
		},
		{
			name: "UnrecognizedLogMode",
			args: args{
				environment: map[string]string{
					"SLACK_API_URL":   gofakeit.URL(),
					"SLACK_BOT_TOKEN": gofakeit.UUID(),
					"SLACK_APP_TOKEN": gofakeit.UUID(),
					"LOG_LEVEL":       "unknown",
				},
			},
			wantErr: true,
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

			if tt.args.environment["SLACK_API_URL"] != "" && config.ApiUrl != tt.args.environment["SLACK_API_URL"] {
				t.Errorf("LoadConfiguration() = %v, want %v", config.ApiUrl, tt.args.environment["SLACK_API_URL"])
			}

			if tt.args.environment["SLACK_APP_TOKEN"] != "" && config.AppToken != tt.args.environment["SLACK_APP_TOKEN"] {
				t.Errorf("LoadConfiguration() = %v, want %v", config.AppToken, tt.args.environment["SLACK_APP_TOKEN"])
			}

			if tt.args.environment["SLACK_BOT_TOKEN"] != "" && config.BotToken != tt.args.environment["SLACK_BOT_TOKEN"] {
				t.Errorf("LoadConfiguration() = %v, want %v", config.BotToken, tt.args.environment["SLACK_BOT_TOKEN"])
			}

			if tt.args.environment["MAX_CONNECT_ATTEMPTS"] != "" && strconv.Itoa(config.MaxConnectAttempts) != tt.args.environment["MAX_CONNECT_ATTEMPTS"] {
				t.Errorf("LoadConfiguration() = %v, want %v", config.MaxConnectAttempts, tt.args.environment["MAX_CONNECT_ATTEMPTS"])
			}

			if tt.args.environment["DEBUG_WEBSOCKET_RECONNECTS"] != "" && strconv.FormatBool(config.DebugWssReconnects) != tt.args.environment["DEBUG_WEBSOCKET_RECONNECTS"] {
				t.Errorf("LoadConfiguration() = %v, want %v", config.DebugWssReconnects, tt.args.environment["DEBUG_WEBSOCKET_RECONNECTS"])
			}

			if tt.args.environment["LOG_LEVEL"] != "" && config.LogLevel.String() != tt.args.environment["LOG_LEVEL"] {
				t.Errorf("LoadConfiguration() = %v, want %v", config.DebugWssReconnects, tt.args.environment["DEBUG_WEBSOCKET_RECONNECTS"])
			}
		})
	}
}
