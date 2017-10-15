package main

import (
	"flag"
	"os"
	"time"

	"github.com/odaira09403/hayaoki_bot/handler"
)

const location = "Asia/Tokyo"

func main() {
	botToken := flag.String("slack-bot-token", "", "Slack bot api token.")
	slashToken := flag.String("slack-slash-token", "", "Slack slash command api token.")
	flag.Parse()

	// Init timezone
	loc, err := time.LoadLocation(location)
	if err != nil {
		loc = time.FixedZone(location, 9*60*60)
	}
	time.Local = loc

	slashToBotChan := make(chan string)

	botHandler := handler.NewBotHandler(*botToken, slashToBotChan)
	go func() {
		os.Exit(botHandler.Run())
	}()

	slashHandler := handler.NewSlashHandler(*slashToken, "./google_client_secret.json", slashToBotChan)
	go func() {
		os.Exit(slashHandler.Run())
	}()

	// Lock process for hanlers.
	var lock chan int
	<-lock
}
