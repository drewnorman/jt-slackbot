package main

import (
	"github.com/drewnorman/jt-slackbot/bot"
	"log"
	"os"
)

func main() {
	config := NewConfiguration()
	err := config.Load()
	if err != nil {
		log.Printf("Error loading configuration: %s", err)
		os.Exit(1)
	}

	slackBot, err := bot.New(&bot.Parameters{
		ApiUrl:             config.apiUrl,
		AppToken:           config.appToken,
		BotToken:           config.botToken,
		MaxConnectAttempts: config.maxConnectAttempts,
		DebugWssReconnects: config.debugWssReconnects,
	})
	if err != nil {
		log.Printf("Error creating bot: %s", err)
		os.Exit(1)
	}

	restart := true
	for restart {
		err = slackBot.AttemptToConnect()
		if err != nil {
			log.Printf("Error attempting to connect: %s", err)
			os.Exit(1)
		}

		err = slackBot.PrepareWorkspace()
		if err != nil {
			log.Printf("Error preparing workspace: %s", err)
			os.Exit(1)
		}

		restart, err = slackBot.Start()
		if err != nil {
			log.Printf("Failed with error: %s", err)
			os.Exit(1)
		}
	}

	os.Exit(0)
}
