package event

import (
	"container/list"
	"errors"
	"github.com/drewnorman/jt-slackbot/slack"
)

type Handler struct {
	httpClient     *slack.HttpClient
	processedQueue *list.List
}

type HandlerParameters struct {
	HttpClient *slack.HttpClient
}

const processedQueueMaxLength = 5

func NewHandler(params *HandlerParameters) (*Handler, error) {
	if params.HttpClient == nil {
		return nil, errors.New("missing http client")
	}
	return &Handler{
		httpClient:     params.HttpClient,
		processedQueue: list.New(),
	}, nil
}

func (handler *Handler) Process(
	events chan map[string]interface{},
	complete chan struct{},
) {
	defer close(complete)
	for event := range events {
		eventId := event["event_id"].(string)

		if handler.hasAlreadyProcessed(eventId) {
			continue
		}

		eventData := (event["event"]).(map[string]interface{})
		handler.processed(eventId)

		switch eventData["type"] {
		case "app_mention":
			err := handler.processAppMentionEvent(event)
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

func (handler *Handler) processAppMentionEvent(
	event map[string]interface{},
) error {
	eventData := event["event"].(map[string]interface{})
	channelId := eventData["channel"].(string)
	err := handler.httpClient.SendMessageToChannel(
		"woof",
		channelId,
	)
	if err != nil {
		return err
	}
	return nil
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
