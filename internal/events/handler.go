package events

import (
	"container/list"
	"errors"
	"github.com/drewnorman/jt-slackbot/internal/slack"
	"go.uber.org/zap"
)

type Handler struct {
	logger            *zap.Logger
	processedQueue    *list.List
	appMentionHandler eventHandler
}

type Parameters struct {
	Logger          *zap.Logger
	SlackHttpClient *slack.HttpClient
}

type eventHandler interface {
	Process(eventData map[string]interface{}) error
}

const processedQueueMaxLength = 5

func NewHandler(params *Parameters) (*Handler, error) {
	if params.Logger == nil {
		return nil, errors.New("missing logger")
	}
	if params.SlackHttpClient == nil {
		return nil, errors.New("missing http client")
	}
	appMentionHandler, err := NewAppMentionHandler(
		&AppMentionHandlerParameters{
			Logger:          params.Logger,
			SlackHttpClient: params.SlackHttpClient,
		},
	)
	if err != nil {
		return nil, err
	}
	return &Handler{
		logger:            params.Logger,
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
			handler.logger.Warn("failed to retrieve event id")
			continue
		}
		if handler.hasAlreadyProcessed(eventId) {
			handler.logger.Debug(
				"already processed event",
				zap.String("eventId", eventId),
			)
			continue
		}

		eventData, ok := (event["event"]).(map[string]interface{})
		if !ok {
			handler.logger.Warn("failed to retrieve event data")
			continue
		}

		handler.processed(eventId)

		switch eventData["type"] {
		case "app_mention":
			err := handler.appMentionHandler.Process(event)
			if err != nil {
				handler.logger.Error(
					"failed to process app mention event",
					zap.String("err", err.Error()),
					zap.String("eventId", eventId),
				)
				break
			}
			continue
		default:
			handler.logger.Debug(
				"skipping processing of unrecognized event",
				zap.String("eventId", eventId),
				zap.String("eventType", eventData["type"].(string)),
			)
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
