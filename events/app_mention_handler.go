package events

import (
	"errors"
	"fmt"
	"github.com/drewnorman/jt-slackbot/slack"
	"github.com/jdkato/prose/v2"
	"strings"
)

type AppMentionHandler struct {
	SlackHttpClient *slack.HttpClient
}

type AppMentionHandlerParameters struct {
	SlackHttpClient *slack.HttpClient
}

type appMentionEvent struct {
	appUserId string
	channelId string
	text      string
}

func NewAppMentionHandler(
	params *AppMentionHandlerParameters,
) (*AppMentionHandler, error) {
	if params.SlackHttpClient == nil {
		return nil, errors.New("missing slack http client")
	}
	return &AppMentionHandler{
		SlackHttpClient: params.SlackHttpClient,
	}, nil
}

func (handler *AppMentionHandler) Process(
	eventData map[string]interface{},
) error {
	event, err := eventFromData(eventData)
	if err != nil {
		return err
	}

	doc, err := prose.NewDocument(
		event.text,
		prose.WithExtraction(false),
		prose.WithSegmentation(false),
	)
	if err != nil {
		return err
	}

	for _, tok := range doc.Tokens() {
		err = handler.SlackHttpClient.SendMessageToChannel(
			tok.Text,
			event.channelId,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

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
