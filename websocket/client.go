package websocket

import (
	"errors"
	"github.com/Jeffail/gabs/v2"
	"github.com/gorilla/websocket"
	"time"
)

type Client struct {
	connection *websocket.Conn
}

func NewClient() *Client {
	return &Client{}
}

func (client *Client) Connect(
	wssUrl string,
) error {
	if wssUrl == "" {
		return errors.New("missing wss url")
	}
	var err error
	client.connection, _, err = websocket.DefaultDialer.Dial(wssUrl, nil)
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) Close(
	complete chan struct{},
	timeout time.Duration,
) (bool, error) {
	err := client.connection.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(
			websocket.CloseNormalClosure,
			"",
		),
	)
	if err != nil {
		return false, err
	}

	timedOut := false
	select {
	case <-complete:
	case <-time.After(timeout):
		timedOut = true
	}
	return timedOut, nil
}

func (client *Client) Disconnect() error {
	return client.connection.Close()
}

func (client *Client) Listen(
	events chan map[string]interface{},
) {
	defer close(events)
	for {
		_, message, err := client.connection.ReadMessage()

		if err != nil {
			return
		}

		decoded, err := gabs.ParseJSON(message)
		if err != nil {
			continue
		}

		messageType, ok := decoded.Path("type").Data().(string)
		if !ok {
			continue
		} else {
			switch messageType {
			case "events_api":
			}
		}

		envelopeId, ok := decoded.Path("envelope_id").Data().(string)
		if !ok {
			continue
		}

		err = client.connection.WriteJSON(map[string]interface{}{
			"envelope_id": envelopeId,
		})
		if err != nil {
			continue
		}

		payload := decoded.Path("payload").Data()
		if payload == nil {
			continue
		}
		event := decoded.Path("payload").Data().(map[string]interface{})

		events <- event
	}
}
