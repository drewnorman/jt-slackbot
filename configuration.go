package main

import (
	"errors"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

type EnvLoader func(filenames ...string) (err error)

type Configuration struct {
	apiUrl             string
	appToken           string
	botToken           string
	maxConnectAttempts int
	debugWssReconnects bool
	loadEnvironment    EnvLoader
}

func NewConfiguration() *Configuration {
	return &Configuration{
		loadEnvironment: godotenv.Load,
	}
}

func (config *Configuration) Load() error {
	exists := false
	if err := config.loadEnvironment(); err != nil {
		return err
	}

	config.apiUrl, exists = os.LookupEnv("SLACK_API_URL")
	if !exists {
		return errors.New("missing slack api url")
	}

	config.appToken, exists = os.LookupEnv("SLACK_APP_TOKEN")
	if !exists {
		return errors.New("missing slack app token")
	}

	config.botToken, exists = os.LookupEnv("SLACK_BOT_TOKEN")
	if !exists {
		return errors.New("missing slack bot token")
	}

	retryMax, exists := os.LookupEnv("MAX_CONNECT_ATTEMPTS")
	if !exists {
		config.maxConnectAttempts = 3
	} else {
		var err error
		config.maxConnectAttempts, err = strconv.Atoi(retryMax)
		if err != nil {
			return err
		}
	}

	debugWssReconnects, exists := os.LookupEnv("DEBUG_WEBSOCKET_RECONNECTS")
	if !exists {
		config.debugWssReconnects = false
	} else {
		config.debugWssReconnects = debugWssReconnects == "true"
	}

	return nil
}
