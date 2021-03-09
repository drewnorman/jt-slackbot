package main

import (
	"github.com/drewnorman/jt-slackbot/internal/bot"
	"github.com/drewnorman/jt-slackbot/internal/configuration"
	"github.com/drewnorman/jt-slackbot/internal/logging"
	"go.uber.org/zap"
	"io"
	"log"
	"os"
)

// main loads a configuration from which it creates
// a bot that runs until failure or signal interrupt
func main() {
	config := configuration.NewConfiguration()
	err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %s", err)
	}

	logger := logging.NewLogger(
		logging.LoggerParameters{
			Level: config.LogLevel,
			Writers: []io.Writer{
				os.Stdout,
			},
		},
	)

	logger.Info("creating new bot")
	slackBot, err := bot.New(&bot.Parameters{
		Logger:             logger,
		ApiUrl:             config.ApiUrl,
		AppToken:           config.AppToken,
		BotToken:           config.BotToken,
		MaxConnectAttempts: config.MaxConnectAttempts,
		DebugWssReconnects: config.DebugWssReconnects,
	})
	if err != nil {
		logger.Error(
			"failed to create bot",
			zap.String("err", err.Error()),
		)
		os.Exit(1)
	}

	restart := true
	for restart {
		logger.Info("connecting to slack")
		err = slackBot.AttemptToConnect()
		if err != nil {
			logger.Fatal(
				"failed to connect",
				zap.String("err", err.Error()),
			)
		}
		logger.Info("connected to slack")

		logger.Info("preparing workspace")
		err = slackBot.PrepareWorkspace()
		if err != nil {
			logger.Fatal(
				"failed to prepare workspace",
				zap.String("err", err.Error()),
			)
		}
		logger.Info("prepared workspace")

		logger.Info("starting bot")
		restart, err = slackBot.Start()
		if err != nil {
			logger.Fatal(
				"failed during bot execution",
				zap.String("err", err.Error()),
			)
		}
		logger.Info("stopped bot")

		logger.Info("reconnecting to slack")
	}

	err = logger.Sync()
	if err != nil {
		log.Fatalf("failed to flush logs: %s", err)
	}

	os.Exit(0)
}
