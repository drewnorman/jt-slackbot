package configuration

import (
	"errors"
	"github.com/joho/godotenv"
	"go.uber.org/zap/zapcore"
	"os"
	"strconv"
)

// EnvLoader loads the environment files
// associated with the given filenames.
type EnvLoader func(filenames ...string) (err error)

// A Configuration is a collection of settings
// for the application.
type Configuration struct {
	ApiUrl             string
	AppToken           string
	BotToken           string
	MaxConnectAttempts int
	DebugWssReconnects bool
	LogLevel           zapcore.Level
	loadEnvironment    EnvLoader
}

// NewConfiguration returns a new instance of
// Configuration specifying the gotdotenv
// library should load the environment variables.
func NewConfiguration() *Configuration {
	return &Configuration{
		loadEnvironment: godotenv.Load,
	}
}

// Load prepares the Configuration from
// the environment, returning an error if the
// required Configuration fields could not
// be set.
func (config *Configuration) Load() error {
	exists := false
	if err := config.loadEnvironment(); err != nil {
		return err
	}

	config.ApiUrl, exists = os.LookupEnv("SLACK_API_URL")
	if !exists {
		return errors.New("missing slack api url")
	}

	config.AppToken, exists = os.LookupEnv("SLACK_APP_TOKEN")
	if !exists {
		return errors.New("missing slack app token")
	}

	config.BotToken, exists = os.LookupEnv("SLACK_BOT_TOKEN")
	if !exists {
		return errors.New("missing slack bot token")
	}

	retryMax, exists := os.LookupEnv("MAX_CONNECT_ATTEMPTS")
	if !exists {
		config.MaxConnectAttempts = 3
	} else {
		var err error
		config.MaxConnectAttempts, err = strconv.Atoi(retryMax)
		if err != nil {
			return err
		}
	}

	debugWssReconnects, exists := os.LookupEnv("DEBUG_WEBSOCKET_RECONNECTS")
	if !exists {
		config.DebugWssReconnects = false
	} else {
		config.DebugWssReconnects = debugWssReconnects == "true"
	}

	logLevel, exists := os.LookupEnv("LOG_LEVEL")
	if !exists {
		config.LogLevel = zapcore.InfoLevel
	} else {
		switch logLevel {
		case "debug":
			config.LogLevel = zapcore.DebugLevel
		case "info":
			config.LogLevel = zapcore.InfoLevel
		case "warn":
			config.LogLevel = zapcore.WarnLevel
		case "error":
			config.LogLevel = zapcore.ErrorLevel
		default:
			return errors.New("unrecognized log level")
		}
	}

	return nil
}
