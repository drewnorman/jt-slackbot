package events

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/drewnorman/jt-slackbot/core/internal/slack"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

// An AppMentionHandler processes app
// mention events.
type AppMentionHandler struct {
	logger          *zap.Logger
	slackHttpClient *slack.HttpClient
}

// AppMentionHandlerParameters describe
// how to create a new AppMentionHandler.
type AppMentionHandlerParameters struct {
	Logger          *zap.Logger
	SlackHttpClient *slack.HttpClient
}

// An appMentionEvent defines the
// attributes of an app mention event.
type appMentionEvent struct {
	appUserId string
	channelId string
	text      string
}

// NewAppMentionHandler returns a new
// instance of AppMentionHandler
// according to the given parameters.
func NewAppMentionHandler(
	params *AppMentionHandlerParameters,
) (*AppMentionHandler, error) {
	if params.Logger == nil {
		return nil, errors.New("missing logger")
	}
	if params.SlackHttpClient == nil {
		return nil, errors.New("missing slack http client")
	}
	return &AppMentionHandler{
		logger:          params.Logger,
		slackHttpClient: params.SlackHttpClient,
	}, nil
}

// Process processes the given event data
// and tries to respond appropriately.
func (handler *AppMentionHandler) Process(
	eventData map[string]interface{},
) error {
	event, err := eventFromData(eventData)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(
		map[string]string{
			"message": event.text,
		},
	)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		"http://localhost:5000/converse",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}

	decoded := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&decoded)
	if err != nil {
		return err
	}

	reply, ok := decoded["reply"].(string)
	if !ok {
		return errors.New("failed to find reply in response")
	}

	err = handler.slackHttpClient.SendMessageToChannel(
		reply,
		event.channelId,
	)
	if err != nil {
		return err
	}

	return nil
}

// eventFromData returns a new appMentionEvent
// from the given event data or an error if
// any of the necessary data is missing.
func eventFromData(
	data map[string]interface{},
) (*appMentionEvent, error) {
	authorizations, ok := data["authorizations"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to determine authorizations from data %v", data)
	}
	authorization, ok := authorizations[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to determine authorization from data %v", data)
	}
	appUserId, ok := authorization["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to determine user id from data %v", data)
	}
	eventData, ok := data["event"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to determine event data from data %v", data)
	}
	channelId, ok := eventData["channel"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to determine channel from event data %v", eventData)
	}
	text, ok := eventData["text"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to determine text from event data %v", eventData)
	}
	text = strings.ReplaceAll(text, "<@"+appUserId+">", "")
	return &appMentionEvent{
		appUserId,
		channelId,
		text,
	}, nil
}
