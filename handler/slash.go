package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/odaira09403/hayaoki_bot/sheets"
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
	Token          string
	SpleadSheet    *sheets.SpreadSheet
	SlashToBotChan chan string
	logger         *log.Logger
}

// NewSlashHandler create SlashHandler instance.
func NewSlashHandler(token, googleSecretPath string, slashToBotChan chan string) *SlashHandler {
	ss, err := sheets.NewSpreadSheet(googleSecretPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	return &SlashHandler{Token: token, SpleadSheet: ss, SlashToBotChan: slashToBotChan}
}

// Run runs SlashHandler.
func (s *SlashHandler) Run() int {
	s.logger = log.New(os.Stderr, "[slash]\t", log.LstdFlags)
	s.logger.Println("Start SlashHandler.")
	http.HandleFunc("/command", s.handler)
	s.logger.Fatal(http.ListenAndServe(":3000", nil))
	return 1
}

func (s *SlashHandler) handler(w http.ResponseWriter, r *http.Request) {
	s.logger.Println("Recieive message.")

	if r.PostFormValue("token") != s.Token {
		s.logger.Println("Invalid token.")
		return
	}

	inputMsg := r.PostFormValue("text")
	userName := r.PostFormValue("user_name")
	if inputMsg == "" {
		err := s.hayaoki(userName, w)
		if err != nil {
			s.responceMsg(w, err.Error(), "ephemeral")
		}
		return
	}

	cmds := strings.Split(inputMsg, " ")

	switch cmds[0] {
	case "kiken":
		if len(cmds) == 1 {
			if err := s.kiken(userName, "", w); err != nil {
				s.responceMsg(w, err.Error(), "ephemeral")
			}
		} else if len(cmds) == 2 {
			if err := s.kiken(userName, cmds[1], w); err != nil {
				s.responceMsg(w, err.Error(), "ephemeral")
			}
		} else {
			s.responceMsg(w, "Invalid format.\n"+UsageString, "ephemeral")
		}
	case "list":
		s.responceMsg(w, "Valid format.\n", "ephemeral")
	case "delete":
		s.responceMsg(w, "Valid format.\n", "ephemeral")
	default:
		s.responceMsg(w, "Invalid format.\n"+UsageString, "ephemeral")
	}
}

func (s *SlashHandler) responceMsg(w http.ResponseWriter, text string, messageType string) {
	rStruct := ResponceMessage{Text: text, ResponseType: messageType}
	// rStruct := ResponceMessage{Text: "test", ResponseType: "in_channel"}

	body, err := json.Marshal(rStruct)
	if err != nil {
		s.logger.Println(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (s *SlashHandler) hayaoki(user string, w http.ResponseWriter) error {
	// Time limitation
	now := time.Now()
	limit := time.Date(now.Year(), now.Month(), now.Day(), 8, 35, 0, 0, now.Location())
	if now.Hour() < 6 || now.After(limit) {
		return errors.New("Please type /hayaoki between 6:00 and 8:35")
	}

	// Append the sheet date if the date is not today.
	date, err := s.SpleadSheet.Hayaoki.GetLastDate()
	if err != nil {
		return err
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, date.Location())
	if date == nil || !date.Equal(today) {
		s.logger.Println("Last date is not today.")
		err := s.SpleadSheet.Hayaoki.AddNewDate()
		if err != nil {
			return err
		}
	}

	// Append the sheet user if the user who send the command is not exist.
	exist, err := s.SpleadSheet.Hayaoki.UserExists(user)
	if err != nil {
		return err
	}
	if !exist {
		err := s.SpleadSheet.Hayaoki.AddNewUser(user)
		if err != nil {
			return err
		}
	}

	// Set hayaoki flag.
	err = s.SpleadSheet.Hayaoki.SetHayaokiFlag(user)
	if err != nil {
		return err
	}

	s.SlashToBotChan <- user + "さんが早起きに成功しました。"
	s.responceMsg(w, "Hayaoki accepted!", "ephemeral")
	return nil
}

func (s *SlashHandler) kiken(user string, dateStr string, w http.ResponseWriter) error {
	// Append the sheet user if the user who send the command is not exist.
	exist, err := s.SpleadSheet.Kiken.UserExists(user)
	if err != nil {
		return err
	}
	if !exist {
		err := s.SpleadSheet.Kiken.AddNewUser(user)
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
	if err := s.SpleadSheet.Kiken.AddDate(user, dates); err != nil {
		return err
	}

	if len(dates) == 1 {
		s.SlashToBotChan <- user + "さんがに" + dates[0].Format("01月02日") + "に棄権します。"
		s.responceMsg(w, "Kiken accepted!\n Date: "+dates[0].Format("2006/01/02"), "ephemeral")
	} else if len(dates) == 2 {
		if dates[0].After(dates[1]) {
			s.responceMsg(w, "Invalid range.\n Date: "+dates[0].Format("2006/01/02")+"-"+dates[1].Format("2006/01/02"), "ephemeral")
		}
		s.SlashToBotChan <- user + "さんがに" + dates[0].Format("01月02日") + "から" + dates[0].Format("01月02日") + "の間棄権します。"
		s.responceMsg(w, "Kiken accepted!\n Date: "+dates[0].Format("2006/01/02")+"-"+dates[1].Format("2006/01/02"), "ephemeral")
	} else {
		s.responceMsg(w, "Invalid format.\n"+UsageString, "ephemeral")
	}
	return nil
}
