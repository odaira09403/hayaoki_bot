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
	Token       string
	SpleadSheet *sheets.SpreadSheet
	logger      *log.Logger
}

// NewSlashHandler create SlashHandler instance.
func NewSlashHandler(token, googleSecretPath string) *SlashHandler {
	ss, err := sheets.NewSpreadSheet(googleSecretPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	return &SlashHandler{Token: token, SpleadSheet: ss}
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
	if inputMsg == "" {
		err := s.hayaoki(r.PostFormValue("user_name"), w)
		if err != nil {
			s.responceMsg(w, err.Error(), "ephemeral")
		}
		return
	}

	cmds := strings.Split(inputMsg, " ")

	switch cmds[0] {
	case "kiken":
		if len(cmds) == 2 {
			err := s.kiken(cmds[1], w)
			if err != nil {
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
	exists, err := s.SpleadSheet.Hayaoki.UserExists(user)
	if err != nil {
		return err
	}
	if !exists {
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

	s.responceMsg(w, "Hayaoki accepted!", "ephemeral")
	return nil
}

func (s *SlashHandler) kiken(dateStr string, w http.ResponseWriter) error {
	dates := strings.Split(dateStr, "-")
	if len(dates) > 2 {
		s.responceMsg(w, "Invalid format.\n"+UsageString, "ephemeral")
	} else if len(dates) == 1 {
		t, err := time.Parse(InputDateFormat, dates[0])
		if err != nil {
			return err
		}
		s.responceMsg(w, "Kiken accepted!\n Date: "+t.Format("2006/01/02"), "ephemeral")
	} else {
		t1, err := time.Parse(InputDateFormat, dates[0])
		if err != nil {
			return err
		}
		t2, err := time.Parse(InputDateFormat, dates[1])
		if err != nil {
			return err
		}
		s.responceMsg(w, "Kiken accepted!\n Date: "+t1.Format("2006/01/02")+"-"+t2.Format("2006/01/02"), "ephemeral")
	}
	return nil
}
