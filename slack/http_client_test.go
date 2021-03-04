package slack

import (
	"bytes"
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

type roundTripHandler func(req *http.Request) *http.Response

func (handler roundTripHandler) RoundTrip(
	req *http.Request,
) (*http.Response, error) {
	return handler(req), nil
}

func fakeHttpClient(
	handler roundTripHandler,
) *http.Client {
	return &http.Client{
		Transport: handler,
	}
}

func defaultFakeHttpClient(
	data map[string]interface{},
) (*http.Client, error) {
	bodyJson, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return fakeHttpClient(
		func(req *http.Request) *http.Response {
			header := http.Header{}
			header.Add("Content-Type", "application/json")
			return &http.Response{
				StatusCode: 200,
				Header:     header,
				Body: ioutil.NopCloser(
					bytes.NewBufferString(string(bodyJson)),
				),
			}
		},
	), nil
}

func TestNewClient(t *testing.T) {
	type args struct {
		params *HttpClientParameters
	}

	apiUrl := gofakeit.URL()
	appToken := gofakeit.UUID()
	botToken := gofakeit.UUID()

	tests := []struct {
		name    string
		args    args
		want    *HttpClient
		wantErr bool
	}{
		{
			name: "ReturnsNewClient",
			args: args{
				params: &HttpClientParameters{
					ApiUrl:   apiUrl,
					AppToken: appToken,
					BotToken: botToken,
				},
			},
			wantErr: false,
		},
		{
			name: "MissingApiUrl",
			args: args{
				params: &HttpClientParameters{
					ApiUrl:   "",
					AppToken: appToken,
					BotToken: botToken,
				},
			},
			wantErr: true,
		},
		{
			name: "MissingAppToken",
			args: args{
				params: &HttpClientParameters{
					ApiUrl:   apiUrl,
					AppToken: "",
					BotToken: botToken,
				},
			},
			wantErr: true,
		},
		{
			name: "MissingBotToken",
			args: args{
				params: &HttpClientParameters{
					ApiUrl:   apiUrl,
					AppToken: appToken,
					BotToken: "",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewHttpClient(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHttpClient() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_RequestWssUrl(t *testing.T) {
	type args struct {
		debugWssReconnects bool
	}

	apiUrl := gofakeit.URL()

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "ReturnsDefaultUrl",
			args: args{
				debugWssReconnects: false,
			},
			want: apiUrl,
		},
		{
			name: "ReturnsDebugUrl",
			args: args{
				debugWssReconnects: true,
			},
			want: apiUrl + "&debug_reconnects=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpClient, err := defaultFakeHttpClient(
				map[string]interface{}{
					"url": apiUrl,
				},
			)
			if err != nil {
				t.Errorf("RequestWssUrl() error = %v, wantErr %v", err, false)
				return
			}
			client := &HttpClient{
				httpClient: httpClient,
			}
			url, err := client.RequestWssUrl(tt.args.debugWssReconnects)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequestWssUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if url != tt.want {
				t.Errorf("RequestWssUrl() got = %s, want %s", url, tt.want)
			}
		})
	}
}

func TestClient_JoinChannel(t *testing.T) {
	type args struct {
		channelId string
	}

	channelId := gofakeit.UUID()
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "JoinsChannel",
			args: args{
				channelId: channelId,
			},
			wantErr: false,
		},
		{
			name: "MissingChannelId",
			args: args{
				channelId: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpClient, err := defaultFakeHttpClient(
				map[string]interface{}{
					"ok": true,
				},
			)
			if err != nil {
				t.Errorf("JoinChannel() error = %v, wantErr %v", err, false)
				return
			}
			client := &HttpClient{
				httpClient: httpClient,
			}
			err = client.JoinChannel(tt.args.channelId)
			if (err != nil) != tt.wantErr {
				t.Errorf("JoinChannel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_PublicChannels(t *testing.T) {
	channels := []map[string]interface{}{
		{
			"id": gofakeit.UUID(),
		},
		{
			"id": gofakeit.UUID(),
		},
	}

	tests := []struct {
		name    string
		want    []map[string]interface{}
		wantErr bool
	}{
		{
			name:    "PublicChannels",
			want:    channels,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpClient, err := defaultFakeHttpClient(
				map[string]interface{}{
					"channels": channels,
				},
			)
			if err != nil {
				t.Errorf("JoinChannel() error = %v, wantErr %v", err, false)
				return
			}
			client := &HttpClient{
				httpClient: httpClient,
			}
			channels, err := client.PublicChannels()
			if (err != nil) != tt.wantErr {
				t.Errorf("JoinPublicChannels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for index, channel := range channels {
				if !reflect.DeepEqual(channel, tt.want[index]) {
					t.Errorf("JoinPublicChannels() = %v, want %v", channels, tt.want)
				}
			}
		})
	}
}

func TestClient_SendMessageToChannel(t *testing.T) {
	type args struct {
		message         string
		channelId       string
		invalidResponse bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "SendsMessageToChannel",
			args: args{
				message:         gofakeit.LoremIpsumSentence(5),
				channelId:       gofakeit.UUID(),
				invalidResponse: false,
			},
			wantErr: false,
		},
		{
			name: "MissingMessage",
			args: args{
				message:         "",
				channelId:       gofakeit.UUID(),
				invalidResponse: false,
			},
			wantErr: true,
		},
		{
			name: "MissingChannel",
			args: args{
				message:         gofakeit.LoremIpsumSentence(5),
				channelId:       "",
				invalidResponse: false,
			},
			wantErr: true,
		},
		{
			name: "UnexpectedJsonResponse",
			args: args{
				message:         gofakeit.LoremIpsumSentence(5),
				channelId:       gofakeit.UUID(),
				invalidResponse: true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var httpClient *http.Client
			if tt.args.invalidResponse {
				httpClient = fakeHttpClient(
					func(req *http.Request) *http.Response {
						header := http.Header{}
						header.Add("Content-Type", "application/json")
						return &http.Response{
							StatusCode: 200,
							Header:     header,
							Body: ioutil.NopCloser(
								bytes.NewBufferString(""),
							),
						}
					},
				)
			} else {
				var err error
				httpClient, err = defaultFakeHttpClient(
					map[string]interface{}{
						"ok": true,
					},
				)
				if err != nil {
					t.Errorf("SendMessageToChannel() error = %v, wantErr %v", err, false)
					return
				}
			}

			client := &HttpClient{
				httpClient: httpClient,
			}
			err := client.SendMessageToChannel(tt.args.message, tt.args.channelId)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendMessageToChannel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
