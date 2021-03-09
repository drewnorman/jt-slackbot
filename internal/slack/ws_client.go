package slack

import (
	"errors"
	"github.com/Jeffail/gabs/v2"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"time"
)

type WsClient struct {
	logger     *zap.Logger
	connection *websocket.Conn
}

type WsClientParameters struct {
	Logger *zap.Logger
}

func NewWsClient(
	params WsClientParameters,
) (*WsClient, error) {
	if params.Logger == nil {
		return nil, errors.New("missing logger")
	}
	return &WsClient{
		logger: params.Logger,
	}, nil
}

func (client *WsClient) Connect(
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

func (client *WsClient) Close(
	complete chan struct{},
	timeout time.Duration,
) (bool, error) {
	client.logger.Debug("sending close message to wss")
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
		client.logger.Debug("sent close message to wss")
	case <-time.After(timeout):
		client.logger.Debug("timed out sending close message to wss")
		timedOut = true
	}
	return timedOut, nil
}

func (client *WsClient) Disconnect() error {
	client.logger.Debug("closing ws client connection")
	return client.connection.Close()
}

func (client *WsClient) Listen(
	events chan map[string]interface{},
) {
	defer close(events)
	for {
		_, message, err := client.connection.ReadMessage()

		if err != nil {
			client.logger.Error(
				"failed to read ws message",
				zap.String("err", err.Error()),
			)
			return
		}

		decoded, err := gabs.ParseJSON(message)
		if err != nil {
			client.logger.Warn(
				"failed to parse ws message",
				zap.String("err", err.Error()),
			)
			continue
		}

		messageType, ok := decoded.Path("type").Data().(string)
		if !ok {
			client.logger.Warn("failed to determine ws message type")
			continue
		} else {
			switch messageType {
			case "events_api":
			}
		}
		client.logger.Debug("received message of type event")

		envelopeId, ok := decoded.Path("envelope_id").Data().(string)
		if !ok {
			client.logger.Warn("failed to determine envelope id")
			continue
		}

		err = client.connection.WriteJSON(map[string]interface{}{
			"envelope_id": envelopeId,
		})
		if err != nil {
			client.logger.Warn("failed to acknowledge message")
			continue
		}
		client.logger.Debug("acknowledged message")

		payload := decoded.Path("payload").Data()
		if payload == nil {
			client.logger.Warn("failed to determine message payload")
			continue
		}
		event := decoded.Path("payload").Data().(map[string]interface{})

		client.logger.Debug("sending event for processing")
		events <- event
	}
}
