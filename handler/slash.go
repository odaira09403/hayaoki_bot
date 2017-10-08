package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
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
	Token  string
	logger *log.Logger
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
	cmds := strings.Split(inputMsg, " ")
	if len(cmds) < 1 {
		s.responceMsg(w, "Prease specify the command.\n"+UsageString, "ephemeral")
		return
	}

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
