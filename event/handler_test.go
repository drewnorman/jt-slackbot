package event

import (
	"bytes"
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	slackhttp "github.com/drewnorman/jt-slackbot/http"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
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

func fakeClient(
	data map[string]interface{},
) (*slackhttp.Client, error) {
	httpClient, err := defaultFakeHttpClient(data)
	if err != nil {
		return nil, err
	}
	return slackhttp.NewClient(&slackhttp.ClientParameters{
		ApiUrl:     gofakeit.URL(),
		AppToken:   gofakeit.UUID(),
		BotToken:   gofakeit.UUID(),
		HttpClient: httpClient,
	})
}

func TestNewHandler(t *testing.T) {
	client, err := fakeClient(map[string]interface{}{})
	if err != nil {
		t.Errorf("NewHandler() error = %v, wantErr %v", err, false)
	}

	type args struct {
		params *HandlerParameters
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ReturnsNewHandler",
			args: args{
				params: &HandlerParameters{
					HttpClient: client,
				},
			},
			wantErr: false,
		},
		{
			name: "MissingHttpClient",
			args: args{
				params: &HandlerParameters{
					HttpClient: nil,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewHandler(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandler_Process(t *testing.T) {
	client, err := fakeClient(map[string]interface{}{})
	if err != nil {
		t.Errorf("Process() error = %v, wantErr %v", err, false)
	}

	type args struct {
		events          chan map[string]interface{}
		complete        chan struct{}
		fakeEvents      []map[string]interface{}
		closeAfterWrite bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "ProcessesAppMention",
			args: args{
				events:   make(chan map[string]interface{}),
				complete: make(chan struct{}),
				fakeEvents: []map[string]interface{}{
					{
						"event_id": gofakeit.UUID(),
						"event": map[string]interface{}{
							"type":    "app_mention",
							"channel": gofakeit.UUID(),
						},
					},
				},
				closeAfterWrite: false,
			},
		},
		{
			name: "ProcessesUnrecognizedEvents",
			args: args{
				events:   make(chan map[string]interface{}),
				complete: make(chan struct{}),
				fakeEvents: []map[string]interface{}{
					{
						"event_id": gofakeit.UUID(),
						"event": map[string]interface{}{
							"type":    "",
							"channel": gofakeit.UUID(),
						},
					},
				},
				closeAfterWrite: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := NewHandler(&HandlerParameters{
				HttpClient: client,
			})
			if err != nil {
				t.Errorf("Process() error = %v, wantErr %v", err, false)
			}

			go handler.Process(tt.args.events, tt.args.complete)

			for _, event := range tt.args.fakeEvents {
				tt.args.events <- event
			}

			if tt.args.closeAfterWrite {
				close(tt.args.events)
			}

			timedOut := false
			select {
			case <-tt.args.complete:
			case <-time.After(50 * time.Millisecond):
				timedOut = true
			}

			if timedOut {
				t.Errorf("Process() = %v, want %v", timedOut, false)
			}
		})
	}
}
