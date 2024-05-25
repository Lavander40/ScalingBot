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

type Response struct{
	Limit float32 `json:"limit"`
}

func NewServer(storage storage.Storage, client *telegram.Client) *Server {
	return &Server{
		tg: client,
		storage: storage,
	}
}

func (s *Server) Start() (err error) {
	http.HandleFunc("/", s.alertHandler)
    return http.ListenAndServe(":6060", nil)
}

func (s *Server) alertHandler(w http.ResponseWriter, r *http.Request) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

	path := r.URL.Path
	fmt.Println(body)

	if path == "/webhook" {
    	s.HandleWebhook(body)
	}

	if path == "/variables" {
		limit := s.HandleVar(body)
		resp := Response{
			Limit: limit,
		}
		json.NewEncoder(w).Encode(resp)
	}
}

func (s *Server) HandleWebhook(req []byte) {
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
		msg := fmt.Sprintf("Внимание облако %s перегружено, значение предела %s%% преодолено, последнее значение метрики: %f", jsonReq.Title, jsonReq.Message, jsonReq.Alerts[0].Values.A)
		s.tg.SendMessage(user, msg)
	}
}

func (s *Server) HandleVar(req []byte) float32 {
	return 0.5
}