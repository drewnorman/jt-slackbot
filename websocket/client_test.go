package websocket

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func fakeWebsocketServer(
	handler http.HandlerFunc,
) (*httptest.Server, string) {
	server := httptest.NewServer(handler)
	wssUrl := upgradeHttpUrl(server.Listener.Addr().String())
	return server, wssUrl
}

func upgradeHttpUrl(httpUrl string) string {
	url := strings.TrimPrefix(httpUrl, "https://")
	url = strings.TrimPrefix(url, "http://")
	return "ws://" + url
}

func TestClient_Connect(t *testing.T) {
	fakeServer, wssUrl := fakeWebsocketServer(
		func(w http.ResponseWriter, r *http.Request) {
			_, err := wsUpgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Errorf("Connect() error = %v, wantErr %v", err, false)
			}

		},
	)
	defer fakeServer.Close()

	type args struct {
		wssUrl string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Connects",
			args: args{
				wssUrl: wssUrl,
			},
			wantErr: false,
		},
		{
			name: "MissingWssUrl",
			args: args{
				wssUrl: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient()
			err := client.Connect(tt.args.wssUrl)
			if (err != nil) != tt.wantErr {
				t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_Close(t *testing.T) {
	fakeServer, wssUrl := fakeWebsocketServer(
		func(w http.ResponseWriter, r *http.Request) {
			_, err := wsUpgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Errorf("Connect() error = %v, wantErr %v", err, false)
			}
		},
	)
	defer fakeServer.Close()

	type args struct {
		complete      chan struct{}
		closeComplete bool
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "ClosesWithComplete",
			args: args{
				complete:      make(chan struct{}),
				closeComplete: true,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "ClosesAfterTimeout",
			args: args{
				complete:      make(chan struct{}),
				closeComplete: false,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient()
			err := client.Connect(wssUrl)
			if err != nil {
				t.Errorf("Close() error = %v, wantErr %v", err, false)
			}

			if tt.args.closeComplete {
				close(tt.args.complete)
			} else {
				defer close(tt.args.complete)
			}

			timedOut, err := client.Close(tt.args.complete, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
			if timedOut != tt.want {
				t.Errorf("Close() got = %v, want %v", timedOut, tt.want)
			}
		})
	}
}

func TestClient_Listen(t *testing.T) {
	var conn *websocket.Conn
	fakeServer, wssUrl := fakeWebsocketServer(
		func(w http.ResponseWriter, r *http.Request) {
			var err error
			conn, err = wsUpgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Errorf("Connect() error = %v, wantErr %v", err, false)
			}
		},
	)
	defer fakeServer.Close()

	type args struct {
		events          chan map[string]interface{}
		fakeEvents      []map[string]interface{}
		invalidResponse bool
		closeAfterWrite bool
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "Listens",
			args: args{
				events: make(chan map[string]interface{}),
				fakeEvents: []map[string]interface{}{
					{
						"type":        "events_api",
						"envelope_id": gofakeit.UUID(),
						"payload":     map[string]interface{}{},
					},
				},
				invalidResponse: false,
				closeAfterWrite: false,
			},
		},
		{
			name: "InvalidJsonResponse",
			args: args{
				events:          make(chan map[string]interface{}),
				fakeEvents:      nil,
				invalidResponse: true,
				closeAfterWrite: true,
			},
		},
		{
			name: "MissingType",
			args: args{
				events: make(chan map[string]interface{}),
				fakeEvents: []map[string]interface{}{
					{
						"envelope_id": gofakeit.UUID(),
						"payload":     map[string]interface{}{},
					},
				},
				invalidResponse: false,
				closeAfterWrite: true,
			},
		},
		{
			name: "MissingEnvelopeId",
			args: args{
				events: make(chan map[string]interface{}),
				fakeEvents: []map[string]interface{}{
					{
						"type":    "events_api",
						"payload": map[string]interface{}{},
					},
				},
				invalidResponse: false,
				closeAfterWrite: true,
			},
		},
		{
			name: "MissingPayload",
			args: args{
				events: make(chan map[string]interface{}),
				fakeEvents: []map[string]interface{}{
					{
						"type":        "events_api",
						"envelope_id": gofakeit.UUID(),
					},
				},
				invalidResponse: false,
				closeAfterWrite: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient()
			err := client.Connect(wssUrl)
			if err != nil {
				t.Errorf("Close() error = %v, wantErr %v", err, false)
			}

			go client.Listen(tt.args.events)

			if tt.args.invalidResponse {
				err = conn.WriteMessage(websocket.TextMessage, []byte(""))
			} else {
				for _, event := range tt.args.fakeEvents {
					err = conn.WriteJSON(event)
					if err != nil {
						t.Errorf("Connect() error = %v, wantErr %v", err, false)
					}
				}
			}

			if tt.args.closeAfterWrite {
				conn.Close()
			}

			var event map[string]interface{}
			select {
			case event = <-tt.args.events:
			case <-time.After(10 * time.Millisecond):
				t.Errorf("Close() got = %v, want %v", event, tt.want)
			}
		})
	}
}
