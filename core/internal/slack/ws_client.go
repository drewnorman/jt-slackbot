package slack

import (
	"errors"
	"github.com/Jeffail/gabs/v2"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"time"
)

// A slack.WsClient listens for Slack events and
// acknowledges them before transferring them
// for processing.
type WsClient struct {
	logger     *zap.Logger
	connection *websocket.Conn
}

// slack.WsClientParameters describe how to
// create a new slack.WsClient.
type WsClientParameters struct {
	Logger *zap.Logger
}

// NewWsClient returns a new slack.WsClient
// according to the given parameters.
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

// Connect dials the given WebSocket server URL
// and stores the resulting connection.
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

// Close writes a close message to the connection
// to allow for a graceful disconnection.
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

// Disconnect closes the connection.
func (client *WsClient) Disconnect() error {
	client.logger.Debug("closing ws client connection")
	return client.connection.Close()
}

// Listen receives Slack events and acknowledges
// them before sending them into the events channel.
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
			case "hello":
				client.logger.Info("received greeting from slack")
				continue
			default:
				client.logger.Warn(
					"unrecognized message type",
					zap.String("messageType", messageType),
				)
				continue
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
