package main

import (
	"flag"
	"os"

	"github.com/odaira09403/hayaoki_bot/handler"
)

func main() {
	botToken := flag.String("slack-bot-token", "", "Slack bot api token.")
	slashToken := flag.String("slack-slash-token", "", "Slack slash command api token.")
	flag.Parse()

	botHandler := handler.NewBotHandler(*botToken)
	go func() {
		os.Exit(botHandler.Run())
	}()

	slashHandler := handler.NewSlashHandler(*slashToken, "./google_client_secret.json")
	go func() {
		os.Exit(slashHandler.Run())
	}()

	// Lock process for hanlers.
	var lock chan int
	<-lock
}
