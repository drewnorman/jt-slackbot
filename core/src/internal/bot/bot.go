package bot

import (
	"errors"
	"fmt"
	"github.com/drewnorman/jt-slackbot/core/internal/events"
	"github.com/drewnorman/jt-slackbot/core/internal/slack"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// A Bot manages a Slack WebSocket connection and
// processes events until failure or interrupt.
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

// Parameters describe the configuration for
// a new Bot.
type Parameters struct {
	Logger             *zap.Logger
	ApiUrl             string
	AppToken           string
	BotToken           string
	MaxConnectAttempts int
	DebugWssReconnects bool
}

// defaultMaxConnectAttempts determines the
// number of times to retry any stage of the
// Slack WebSocket connection process
const defaultMaxConnectAttempts = 3

// defaultEventProcessingTimeout defines the
// duration of time to wait for event processing
// to complete before stopping the bot entirely
const defaultEventProcessingTimeout = 3 * time.Second

// New returns a new instance of Bot according
// to the given parameters.
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

	maxConnectAttempts := defaultMaxConnectAttempts
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
	signal.Notify(
		bot.interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	return bot, nil
}

// Run connects to Slack, prepares the workspace,
// and executes the main sequence until it
// encounters an error or is explicitly told to
// stop and not restart
func (bot *Bot) Run() error {
	restart := true
	var err error
	for restart {
		bot.logger.Info("preparing workspace")
		err = bot.prepareWorkspace()
		if err != nil {
			return err
		}
		bot.logger.Info("prepared workspace")

		bot.logger.Info("connecting to slack")
		err = bot.attemptToConnect()
		if err != nil {
			return err
		}
		bot.logger.Info("connected to slack")

		bot.logger.Info("executing main sequence")
		restart, err = bot.executeMainSequence()
		if err != nil {
			return err
		}
		bot.logger.Info("stopped main sequence")

		if restart {
			bot.logger.Info("reconnecting to slack")
		}
	}
	return err
}

// attemptToConnect requests a Slack WebSocket URL
// and attempts to connect with it, retrying until
// the max attempts specified for the Bot have
// been reached.
func (bot *Bot) attemptToConnect() error {
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

// prepareWorkspace retrieves all public channels
// for the workspace and tries to join them one
// at a time.
func (bot *Bot) prepareWorkspace() error {
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

// executeMainSequence creates an event handler and begins
// concurrent listening and processing of
// Slack events.
func (bot *Bot) executeMainSequence() (bool, error) {
	var err error
	bot.logger.Debug("creating events handler")
	bot.handler, err = events.NewHandler(
		&events.Parameters{
			Logger:          bot.logger,
			SlackHttpClient: bot.httpClient,
		},
	)
	if err != nil {
		return false, err
	}
	bot.logger.Debug("created events handler")

	eventsStream := make(chan map[string]interface{})
	processingComplete := make(chan struct{})

	bot.logger.Debug("starting event listening and handling")
	go bot.wsClient.Listen(eventsStream)
	go bot.handler.Process(eventsStream, processingComplete)
	bot.logger.Debug("started event listening and handling")

	restart := true
	select {
	case <-bot.interrupt:
		bot.logger.Info("received interrupt signal")
		restart = false
	case <-processingComplete:
		bot.logger.Info("event handling completed")
	}

	bot.logger.Debug("closing ws client")
	_, err = bot.wsClient.Close(
		processingComplete,
		defaultEventProcessingTimeout,
	)
	if err != nil {
		return false, err
	}
	bot.logger.Debug("closed ws client")

	bot.logger.Debug("disconnecting from wss")
	return restart, bot.wsClient.Disconnect()
}
