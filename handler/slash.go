package handler

import (
	"encoding/json"
	"log"
	"os"
	"net/http"
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
	r.PostFormValue("text")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	rStruct := ResponceMessage{Text: "test", ResponseType: "ephemeral"}
	// rStruct := ResponceMessage{Text: "test", ResponseType: "in_channel"}

	body, err := json.Marshal(rStruct)
	if err != nil {
		s.logger.Println(err.Error())
		return
	}

    w.Write(body)
}