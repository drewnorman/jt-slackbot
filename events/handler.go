package events

import (
	"container/list"
	"errors"
	"github.com/drewnorman/jt-slackbot/slack"
)

type Handler struct {
	processedQueue    *list.List
	appMentionHandler eventHandler
}

type Parameters struct {
	SlackHttpClient *slack.HttpClient
}

type eventHandler interface {
	Process(eventData map[string]interface{}) error
}

const processedQueueMaxLength = 5

func NewHandler(params *Parameters) (*Handler, error) {
	if params.SlackHttpClient == nil {
		return nil, errors.New("missing http client")
	}
	appMentionHandler, err := NewAppMentionHandler(&AppMentionHandlerParameters{
		SlackHttpClient: params.SlackHttpClient,
	})
	if err != nil {
		return nil, err
	}
	return &Handler{
		processedQueue:    list.New(),
		appMentionHandler: appMentionHandler,
	}, nil
}

func (handler *Handler) Process(
	events chan map[string]interface{},
	complete chan struct{},
) {
	defer close(complete)
	for event := range events {
		eventId, ok := event["event_id"].(string)
		if !ok {
			continue
		}
		if handler.hasAlreadyProcessed(eventId) {
			continue
		}

		eventData, ok := (event["event"]).(map[string]interface{})
		if !ok {
			continue
		}

		handler.processed(eventId)

		switch eventData["type"] {
		case "app_mention":
			err := handler.appMentionHandler.Process(event)
			if err != nil {
				break
			}
			continue
		default:
			continue
		}
		break
	}
}

func (handler *Handler) hasAlreadyProcessed(
	eventId string,
) bool {
	for id := handler.processedQueue.Front(); id != nil; id = id.Next() {
		if eventId == id.Value {
			return true
		}
	}
	return false
}

func (handler *Handler) processed(
	eventId string,
) {
	handler.processedQueue.PushFront(eventId)
	if handler.processedQueue.Len() > processedQueueMaxLength {
		handler.processedQueue.Remove(handler.processedQueue.Back())
	}
}
