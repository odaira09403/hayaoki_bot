package handler

import (
	"log"
	"os"
	"strings"

	"github.com/nlopes/slack"
)

// BotHandler handles SlaskBot message.
type BotHandler struct {
	Token          string
	SlashToBotChan chan string
	logger         *log.Logger
}

// NewBotHandler creates BotHandler instance.
func NewBotHandler(token string, slashToBotChan chan string) *BotHandler {
	return &BotHandler{Token: token, SlashToBotChan: slashToBotChan}
}

// Run runs BotHandler.
func (h *BotHandler) Run() int {
	h.logger = log.New(os.Stderr, "[bot]\t", log.LstdFlags)
	h.logger.Println("Start BotHandler.")
	api := slack.New(h.Token)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.HelloEvent:
				h.logger.Print("Hello Event")

			case *slack.MessageEvent:
				h.handleMessage(ev, rtm)

			case *slack.InvalidAuthEvent:
				h.logger.Print("Invalid credentials")
				return 1
			}

		case msg := <-h.SlashToBotChan:
			h.logger.Println("Message from slash command:", msg)
			rtm.SendMessage(rtm.NewOutgoingMessage(msg, "C7G1P683H"))
		}
	}
}

func (h *BotHandler) handleMessage(ev *slack.MessageEvent, rtm *slack.RTM) {
	h.logger.Printf("Message: %v\n", ev.Msg.Text)
	h.logger.Println("UserId:", rtm.GetInfo().User.ID)
	h.logger.Println("Channel:", ev.Channel)

	cmds := strings.Split(ev.Text, " ")
	// Identify mention for bot.
	if len(cmds) > 0 {
		if cmds[0] == "<@"+rtm.GetInfo().User.ID+">" {
			cmds = cmds[1:]
		} else {
			// Ignore public channel message.
			if rtm.GetInfo().GetChannelByID(ev.Channel) != nil {
				return
			}
		}
	}

	user, err := rtm.GetUserInfo(ev.Msg.User)
	if err != nil {
		rtm.SendMessage(rtm.NewOutgoingMessage(err.Error(), ev.Channel))
	}
	botName, err := rtm.GetUserInfo(rtm.GetInfo().User.ID)
	if err != nil {
		rtm.SendMessage(rtm.NewOutgoingMessage(err.Error(), ev.Channel))
	}

	// Ignore message from myself.
	if user.ID == botName.ID {
		h.logger.Println("Message from myself.")
		return
	}

	rtm.SendMessage(rtm.NewOutgoingMessage("Hi "+user.Name+". I'm "+botName.Name+".", ev.Channel))
	rtm.SendMessage(rtm.NewOutgoingMessage("I will regularly send you hayaoki information. ", ev.Channel))
	return
}
