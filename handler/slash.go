package handler

import (
	"strings"
	"encoding/json"
	"log"
	"os"
	"net/http"
)

const (
	USAGE_STRING = "Usage: /hayaoki [kiken|cancel|list] [month/day[-month/day]]"
)

type ResponceMessage struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

type SlashHandler struct {
	Token string
	logger *log.Logger
}

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
		s.responceMsg(w, "Prease specify the command.\n" + USAGE_STRING, "ephemeral")
		return
	} else if len(cmds) < 2 {
		switch cmds[0] {
		case "kiken":
			s.responceMsg(w, "Valid message.\n", "ephemeral")
		case "list":
			s.responceMsg(w, "Valid message.\n", "ephemeral")
		case "delete":
			s.responceMsg(w, "Valid message.\n", "ephemeral")
		default:
			s.responceMsg(w, "Invalid message.\n" + USAGE_STRING, "ephemeral")
		}

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