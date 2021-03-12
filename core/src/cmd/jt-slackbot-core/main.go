package main

import (
	"github.com/drewnorman/jt-slackbot/core/internal/bot"
	"github.com/drewnorman/jt-slackbot/core/internal/configuration"
	"github.com/drewnorman/jt-slackbot/core/internal/logging"
	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"os"
)

// maxLogFileSizeInMb, maxLogBackups, and
// maxLogAgeInDays determine the settings
// for log file management
const (
	maxLogFileSizeInMb = 10
	maxLogBackups      = 3
	maxLogAgeInDays    = 28
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
				&lumberjack.Logger{
					Filename:   "/var/log/jt-slackbot-core/main.log",
					MaxSize:    maxLogFileSizeInMb,
					MaxBackups: maxLogBackups,
					MaxAge:     maxLogAgeInDays,
				},
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

	logger.Info("starting bot")
	err = slackBot.Run()
	if err != nil {
		logger.Error(
			"error during bot execution",
			zap.String("err", err.Error()),
		)
	}
	logger.Info("stopped bot")

	os.Exit(0)
}
