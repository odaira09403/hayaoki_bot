package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"google.golang.org/appengine/log"
	"google.golang.org/appengine"
	"context"
	"google.golang.org/appengine/datastore"
	"github.com/odaira09403/hayaoki_bot/sheets"
	"github.com/nlopes/slack"
	"google.golang.org/appengine/urlfetch"
	"errors"
)

const (
	// UsageString is usualy used as a response of invalid format.
	UsageString = "Usage: /hayaoki [kiken|cancel|list] [month/day[-month/day]]"
	// InputDateFormat is date format of input message.
	InputDateFormat = "1/2"
)

// ResponceMessage is format of slash command message.
type ResponceMessage struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

// SlashHandler handles slash message.
type SlashHandler struct {
	Ctx            context.Context
	BotClient      *slack.Client
	PostPram       slack.PostMessageParameters
	HayaokiChannel string
	SpreadSheet    *sheets.SpreadSheet
}

// NewSlashHandler create SlashHandler instance.
func NewSlashHandler(channelID string) *SlashHandler {
	return &SlashHandler{HayaokiChannel: channelID}
}

// Run runs SlashHandler.
func (s *SlashHandler) Run() {
	http.HandleFunc("/command", s.handler)
}

type SlackToken struct {
	Value string `datastore:"value"`
}

func (s *SlashHandler) handler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	s.Ctx = ctx
	log.Infof(ctx, "Receive message.")

	// Get slash command token.
	key := datastore.NewKey(ctx, "slack", "slash_token", 0, nil)
	slackToken := new(SlackToken)
	if err := datastore.Get(ctx, key, slackToken); err != nil {
		log.Errorf(ctx, err.Error())
		return
	}

	// Check access token.
	if r.PostFormValue("token") != slackToken.Value {
		log.Infof(ctx, "Invalid token.")
		return
	}

	// New spread sheet instance.
	var err error = nil
	s.SpreadSheet, err = sheets.NewSpreadSheet(ctx)
	if err != nil {
		log.Errorf(ctx, err.Error())
		s.responseMsg(w, err.Error(), "ephemeral")
		return
	}

	// Get bot token.
	key = datastore.NewKey(ctx, "slack", "bot_token", 0, nil)
	slackToken = new(SlackToken)
	if err := datastore.Get(ctx, key, slackToken); err != nil {
		log.Errorf(ctx, err.Error())
		return
	}

	// New RTM instance for the reply.
	slack.SetHTTPClient(urlfetch.Client(ctx))
	s.BotClient = slack.New(slackToken.Value)
	s.PostPram = slack.PostMessageParameters{
		Username:    "hayaoki_bot",
		AsUser:      true,
		IconURL:     "https://avatars.slack-edge.com/2017-10-07/252871472931_f189dd4ee78316f6cd13_72.png",
	}

	inputMsg := r.PostFormValue("text")
	userName := r.PostFormValue("user_name")
	if inputMsg == "" {
		err := s.hayaoki(userName, w)
		if err != nil {
			log.Errorf(ctx, err.Error())
			s.responseMsg(w, err.Error(), "ephemeral")
		}
		return
	}

	cmds := strings.Split(inputMsg, " ")

	switch cmds[0] {
	case "kiken":
		if len(cmds) == 1 {
			if err := s.kiken(userName, "", w); err != nil {
				s.responseMsg(w, err.Error(), "ephemeral")
			}
		} else if len(cmds) == 2 {
			if err := s.kiken(userName, cmds[1], w); err != nil {
				s.responseMsg(w, err.Error(), "ephemeral")
			}
		} else {
			s.responseMsg(w, "Invalid format.\n"+UsageString, "ephemeral")
		}
	case "list":
		s.responseMsg(w, "Valid format. But it is not implemented.\n", "ephemeral")
	case "delete":
		s.responseMsg(w, "Valid format. But it is not implemented.\n", "ephemeral")
	default:
		s.responseMsg(w, "Invalid format.\n"+UsageString, "ephemeral")
	}
}

func (s *SlashHandler) responseMsg(w http.ResponseWriter, text string, messageType string) {
	rStruct := ResponceMessage{Text: text, ResponseType: messageType}

	body, err := json.Marshal(rStruct)
	if err != nil {
		log.Errorf(s.Ctx, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (s *SlashHandler) hayaoki(user string, w http.ResponseWriter) error {
	// Time limitation
	now := time.Now()
	limit := time.Date(now.Year(), now.Month(), now.Day(), 8, 5, 0, 0, now.Location())
	if now.Hour() < 6 || now.After(limit) {
		return errors.New("Please type /hayaoki between 6:00 and 8:05")
	}

	// Append the sheet date if the date is not today.
	date, err := s.SpreadSheet.Hayaoki.GetLastDate()
	if err != nil {
		return err
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, date.Location())
	if date == nil || !date.Equal(today) {
		log.Infof(s.Ctx, "Last date is not today.")
		err := s.SpreadSheet.Hayaoki.AddNewDate()
		if err != nil {
			return err
		}
	}

	// Append the sheet user if the user who send the command is not exist.
	exist, err := s.SpreadSheet.Hayaoki.UserExists(user)
	if err != nil {
		return err
	}
	if !exist {
		err := s.SpreadSheet.Hayaoki.AddNewUser(user)
		if err != nil {
			return err
		}
	}

	// Set hayaoki flag.
	err = s.SpreadSheet.Hayaoki.SetHayaokiFlag(now, user)
	if err != nil {
		return err
	}

	s.BotClient.PostMessage(s.HayaokiChannel, user + "さんが早起きに成功しました。", s.PostPram)
	s.responseMsg(w, "Hayaoki accepted!", "ephemeral")
	return nil
}

func (s *SlashHandler) kiken(user string, dateStr string, w http.ResponseWriter) error {
	// Append the sheet user if the user who send the command is not exist.
	exist, err := s.SpreadSheet.Kiken.UserExists(user)
	if err != nil {
		return err
	}
	if !exist {
		err := s.SpreadSheet.Kiken.AddNewUser(user)
		if err != nil {
			return err
		}
	}

	// Convert no year date string to date.
	now := time.Now()
	dates := []time.Time{}
	if dateStr != "" {
		dateStrs := strings.Split(dateStr, "-")
		for _, str := range dateStrs {
			t, err := time.Parse(InputDateFormat, str)
			if err != nil {
				return err
			}
			date := time.Date(now.Year(), t.Month(), t.Day(), 7, 30, 0, 0, now.Location())
			if now.After(date) {
				date = date.AddDate(1, 0, 0)
			}
			dates = append(dates, date)
		}
	} else {
		dates = append(dates, time.Now().Add((24-7)*time.Hour-30*time.Minute))
	}

	// Add new date.
	if err := s.SpreadSheet.Kiken.AddDate(user, dates); err != nil {
		return err
	}

	if len(dates) == 1 {
		if _, _, err := s.BotClient.PostMessage(
			s.HayaokiChannel,
			user + "さんがに" + dates[0].Format("01月02日") + "に棄権します。",
			s.PostPram); err != nil {
			return err
		}
		s.responseMsg(w, "Kiken accepted!\n Date: "+dates[0].Format("2006/01/02"), "ephemeral")
	} else if len(dates) == 2 {
		if dates[0].After(dates[1]) {
			s.responseMsg(w, "Invalid range.\n Date: "+dates[0].Format("2006/01/02")+"-"+dates[1].Format("2006/01/02"), "ephemeral")
		}
		if _, _, err := s.BotClient.PostMessage(
			s.HayaokiChannel,
			user + "さんがに" + dates[0].Format("01月02日") + "から" + dates[1].Format("01月02日") + "の間棄権します。",
			s.PostPram); err != nil {
				return err
		}
		s.responseMsg(w, "Kiken accepted!\n Date: "+dates[0].Format("2006/01/02")+"-"+dates[1].Format("2006/01/02"), "ephemeral")
	} else {
		s.responseMsg(w, "Invalid format.\n"+UsageString, "ephemeral")
	}
	return nil
}
