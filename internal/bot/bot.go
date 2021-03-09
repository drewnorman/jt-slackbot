package bot

import (
	"errors"
	"fmt"
	"github.com/drewnorman/jt-slackbot/internal/events"
	"github.com/drewnorman/jt-slackbot/internal/slack"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"time"
)

type Bot struct {
	logger             *zap.Logger
	apiUrl             string
	appToken           string
	botToken           string
	maxConnectAttempts int
	debugWssReconnects bool
	httpClient         *slack.HttpClient
	wsClient           *slack.WsClient
	handler            *events.Handler
	interrupt          chan os.Signal
}

type Parameters struct {
	Logger             *zap.Logger
	ApiUrl             string
	AppToken           string
	BotToken           string
	MaxConnectAttempts int
	DebugWssReconnects bool
}

func New(params *Parameters) (*Bot, error) {
	if params.Logger == nil {
		return nil, errors.New("missing logger")
	}
	if params.ApiUrl == "" {
		return nil, errors.New("missing api url")
	}
	if params.AppToken == "" {
		return nil, errors.New("missing app token")
	}
	if params.BotToken == "" {
		return nil, errors.New("missing bot token")
	}

	maxConnectAttempts := 3
	debugWssReconnects := false
	if params.MaxConnectAttempts != maxConnectAttempts {
		maxConnectAttempts = params.MaxConnectAttempts
	}
	if params.DebugWssReconnects {
		debugWssReconnects = true
	}

	bot := &Bot{
		logger:             params.Logger,
		apiUrl:             params.ApiUrl,
		appToken:           params.AppToken,
		botToken:           params.BotToken,
		maxConnectAttempts: maxConnectAttempts,
		debugWssReconnects: debugWssReconnects,
	}

	httpClient, err := slack.NewHttpClient(
		&slack.HttpClientParameters{
			Logger:   bot.logger,
			ApiUrl:   bot.apiUrl,
			AppToken: bot.appToken,
			BotToken: bot.botToken,
		},
	)
	if err != nil {
		return nil, err
	}
	bot.httpClient = httpClient

	bot.interrupt = make(chan os.Signal, 1)
	signal.Notify(bot.interrupt, os.Interrupt)

	return bot, nil
}

func (bot *Bot) AttemptToConnect() error {
	wssUrl := ""
	attemptsLeft := bot.maxConnectAttempts
	for {
		var err error
		bot.logger.Debug("requesting slack wss url")
		wssUrl, err = bot.httpClient.RequestWssUrl(
			bot.debugWssReconnects,
		)
		if err != nil {
			bot.logger.Warn(
				"failed requesting slack wss url",
				zap.String("err", err.Error()),
			)
			attemptsLeft -= 1
			if attemptsLeft > 0 {
				bot.logger.Debug(
					"retrying slack wss url request",
					zap.Int("attemptsLeft", attemptsLeft),
				)
				continue
			}
			bot.logger.Debug("failed to retrieve slack wss url")
			break
		}
		break
	}
	if wssUrl == "" {
		return fmt.Errorf(
			"failed to retrieve slack wss url after %d attempts",
			bot.maxConnectAttempts,
		)
	}
	bot.logger.Debug(
		"retrieved slack wss url",
		zap.String("wssUrl", wssUrl),
	)

	var err error
	bot.logger.Debug("creating new slack ws client")
	bot.wsClient, err = slack.NewWsClient(
		slack.WsClientParameters{
			Logger: bot.logger,
		},
	)
	if err != nil {
		return err
	}
	bot.logger.Debug("created new slack ws client")

	attemptsLeft = bot.maxConnectAttempts
	for attemptsLeft > 0 {
		bot.logger.Debug("connecting to slack wss")
		err := bot.wsClient.Connect(wssUrl)
		if err != nil {
			bot.logger.Warn(
				"failed connecting to slack wss",
				zap.String("err", err.Error()),
			)
			attemptsLeft -= 1
			if attemptsLeft == 0 {
				return fmt.Errorf(
					"failed to connect to slack wss after %d attempts",
					bot.maxConnectAttempts,
				)
			}
			bot.logger.Debug(
				"retrying slack wss connection",
				zap.Int("attemptsLeft", attemptsLeft),
			)
			continue
		}
		break
	}

	return nil
}

func (bot *Bot) PrepareWorkspace() error {
	channels, err := bot.httpClient.PublicChannels()
	if err != nil {
		return err
	}
	bot.logger.Debug("retrieved public channels for workspace")

	for _, channel := range channels {
		channelId, ok := channel.(map[string]interface{})["id"].(string)
		if !ok {
			bot.logger.Warn("failed to determine channel id")
			continue
		}
		err = bot.httpClient.JoinChannel(channelId)
		if err != nil {
			bot.logger.Warn(
				"failed to join channel",
				zap.String("channelId", channelId),
			)
			continue
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func (bot *Bot) Start() (bool, error) {
	var err error
	bot.logger.Debug("creating events handler")
	bot.handler, err = events.NewHandler(&events.Parameters{
		SlackHttpClient: bot.httpClient,
	})
	if err != nil {
		return false, err
	}
	bot.logger.Debug("created events handler")

	eventsStream := make(chan map[string]interface{})
	complete := make(chan struct{})

	bot.logger.Debug("starting event listening and handling")
	go bot.wsClient.Listen(eventsStream)
	go bot.handler.Process(eventsStream, complete)
	bot.logger.Debug("started event listening and handling")

	restart := true
	for {
		select {
		case <-complete:
			bot.logger.Info("processing completed")
			break
		case <-bot.interrupt:
			bot.logger.Info("received interrupt signal")
			restart = false
			break
		default:
			continue
		}
		break
	}

	bot.logger.Debug("closing ws client")
	_, err = bot.wsClient.Close(complete, 1*time.Second)
	if err != nil {
		return false, err
	}
	bot.logger.Debug("closed ws client")

	bot.logger.Debug("disconnecting from wss")
	return restart, bot.wsClient.Disconnect()
}
