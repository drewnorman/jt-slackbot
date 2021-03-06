package events

import (
	"bytes"
	"container/list"
	"encoding/json"
	"errors"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/drewnorman/jt-slackbot/internal/slack"
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

func fakeSlackHttpClient(
	t *testing.T,
	data map[string]interface{},
) *slack.HttpClient {
	t.Helper()

	bodyJson, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	httpClient := fakeHttpClient(
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
	)

	slackHttpClient, err := slack.NewHttpClient(
		&slack.HttpClientParameters{
			ApiUrl:     gofakeit.URL(),
			AppToken:   gofakeit.UUID(),
			BotToken:   gofakeit.UUID(),
			HttpClient: httpClient,
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	return slackHttpClient
}

type genericAppMentionHandler struct {
	Processed bool
	process   func(eventData map[string]interface{}) error
}

func (handler *genericAppMentionHandler) Process(
	eventData map[string]interface{},
) error {
	err := handler.process(eventData)
	if err != nil {
		return err
	}
	handler.Processed = true
	return nil
}

func fakeAppMentionHandler(
	process func(eventData map[string]interface{}) error,
) *genericAppMentionHandler {
	return &genericAppMentionHandler{
		Processed: false,
		process:   process,
	}
}

func TestNewHandler(t *testing.T) {
	type args struct {
		params *Parameters
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ReturnsNewHandler",
			args: args{
				params: &Parameters{
					SlackHttpClient: fakeSlackHttpClient(
						t,
						map[string]interface{}{},
					),
				},
			},
			wantErr: false,
		},
		{
			name: "MissingHttpClient",
			args: args{
				params: &Parameters{
					SlackHttpClient: nil,
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
	type args struct {
		events                           chan map[string]interface{}
		complete                         chan struct{}
		fakeEvents                       []map[string]interface{}
		checkIfAppMentionProcessed       bool
		appMentionShouldNotHaveProcessed bool
		appMentionHandler                *genericAppMentionHandler
		leaveEventsChannelOpenAfterWrite bool
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
							"type": "app_mention",
						},
					},
				},
				checkIfAppMentionProcessed: true,
				appMentionHandler: fakeAppMentionHandler(
					func(eventData map[string]interface{}) error {
						return nil
					},
				),
			},
		},
		{
			name: "ReturnsWithAppMentionError",
			args: args{
				events:   make(chan map[string]interface{}),
				complete: make(chan struct{}),
				fakeEvents: []map[string]interface{}{
					{
						"event_id": gofakeit.UUID(),
						"event": map[string]interface{}{
							"type": "app_mention",
						},
					},
				},
				leaveEventsChannelOpenAfterWrite: true,
				checkIfAppMentionProcessed:       true,
				appMentionShouldNotHaveProcessed: true,
				appMentionHandler: fakeAppMentionHandler(
					func(eventData map[string]interface{}) error {
						return errors.New("fake app mention event handler error")
					},
				),
			},
		},
		{
			name: "IgnoresEventsWithoutId",
			args: args{
				events:   make(chan map[string]interface{}),
				complete: make(chan struct{}),
				fakeEvents: []map[string]interface{}{
					{},
					{},
				},
			},
		},
		{
			name: "IgnoresAlreadyProcessedEvents",
			args: args{
				events:   make(chan map[string]interface{}),
				complete: make(chan struct{}),
				fakeEvents: []map[string]interface{}{
					{
						"event_id": "1111-2222-3333-4444",
						"event": map[string]interface{}{
							"type": "app_mention",
						},
					},
					{
						"event_id": "1111-2222-3333-4444",
						"event": map[string]interface{}{
							"type": "app_mention",
						},
					},
					{
						"event_id": gofakeit.UUID(),
						"event": map[string]interface{}{
							"type": "app_mention",
						},
					},
				},
				checkIfAppMentionProcessed: true,
				appMentionHandler: fakeAppMentionHandler(
					func(eventData map[string]interface{}) error {
						return nil
					},
				),
			},
		},
		{
			name: "IgnoresEventsMissingEventData",
			args: args{
				events:   make(chan map[string]interface{}),
				complete: make(chan struct{}),
				fakeEvents: []map[string]interface{}{
					{
						"event_id": gofakeit.UUID(),
					},
				},
			},
		},
		{
			name: "IgnoresUnrecognizedEvents",
			args: args{
				events:   make(chan map[string]interface{}),
				complete: make(chan struct{}),
				fakeEvents: []map[string]interface{}{
					{
						"event_id": gofakeit.UUID(),
						"event": map[string]interface{}{
							"type": "",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &Handler{
				processedQueue:    list.New(),
				appMentionHandler: tt.args.appMentionHandler,
			}

			go handler.Process(tt.args.events, tt.args.complete)

			for _, event := range tt.args.fakeEvents {
				tt.args.events <- event
			}
			if !tt.args.leaveEventsChannelOpenAfterWrite {
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

			if tt.args.checkIfAppMentionProcessed {
				if tt.args.appMentionHandler.Processed == tt.args.appMentionShouldNotHaveProcessed {
					t.Errorf(
						"Process() = %v, want %v",
						tt.args.appMentionHandler.Processed,
						tt.args.appMentionShouldNotHaveProcessed,
					)
				}
			}
		})
	}
}
