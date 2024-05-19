package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"scaling-bot/client/telegram"
	"scaling-bot/storage"
)

type Server struct {
	tg *telegram.Client
	storage storage.Storage
}

type CloudRequest struct {
	Receiver string `json:"receiver"`
	Status   string `json:"status"`
	Alerts   []struct {
		Values       struct {
			A float64 `json:"A"`
			C int     `json:"C"`
		} `json:"values"`
	} `json:"alerts"`
	Title           string `json:"title"`
	Message         string `json:"message"`
}

func NewServer(storage storage.Storage, client *telegram.Client) *Server {
	return &Server{
		storage: storage,
	}
}

func (s *Server) Start() (err error) {
	http.HandleFunc("/webhook", s.alertHandler)
    return http.ListenAndServe(":6060", nil)
}

func (s *Server) alertHandler(w http.ResponseWriter, r *http.Request) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Failed to read request body", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    s.HandleRequest(body)
}

func (s *Server) HandleRequest(req []byte) {
	var jsonReq CloudRequest
	err := json.Unmarshal(req, &jsonReq)
	if err != nil {
		log.Fatal(err)
	}

	userList, err := s.storage.GetUserByCloud(jsonReq.Title)
	if err != nil {
		log.Fatal(err)
	}

	for _, user := range userList {
		s.tg.SendMessage(user, fmt.Sprintf("Внимание облако %s перегружено, значение предела 50% преодолено, последнее значение метрики: %f", jsonReq.Title, jsonReq.Message, jsonReq.Alerts[0].Values.A))
	}
}