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
	tg      *telegram.Client
	storage storage.Storage
}

type CloudRequest struct {
	Receiver string `json:"receiver"`
	Status   string `json:"status"`
	Alerts   []struct {
		Values struct {
			B float32 `json:"B0"`
		} `json:"values"`
	} `json:"alerts"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

type Response struct {
	Limit float32 `json:"limit"`
}

func NewServer(storage storage.Storage, client *telegram.Client) *Server {
	return &Server{
		tg:      client,
		storage: storage,
	}
}

func (s *Server) Start() (err error) {
	http.HandleFunc("/webhook", s.HandleWebhook)
	http.HandleFunc("/variables", s.HandleVar)
	return http.ListenAndServe(":6060", nil)
}

func (s *Server) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var jsonReq CloudRequest
	if err := json.Unmarshal(body, &jsonReq); err != nil {
		log.Fatal(err)
	}

	userList, err := s.storage.GetUserByCloud(jsonReq.Title)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", string(body))	

	if jsonReq.Status == "firing" {
		for _, user := range userList {
			msg := fmt.Sprintf("Внимание облако %s перегружено, значение предела %s%% преодолено, последнее значение метрики: %f", jsonReq.Title, jsonReq.Message, jsonReq.Alerts[0].Values.B * 100)
			s.tg.SendMessage(user, msg)
		}
	}

	if jsonReq.Status == "resolved" {
		for _, user := range userList {
			msg := fmt.Sprintf("Облако %s больше не нагружено, значение метрик находится ниже установленного предела %s%%", jsonReq.Title, jsonReq.Message)
			s.tg.SendMessage(user, msg)
		}
	}
}

func (s *Server) HandleVar(w http.ResponseWriter, r *http.Request) {
// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	defer r.Body.Close()

	resp := Response{
		Limit: 0.5,
	}
	json.NewEncoder(w).Encode(resp)
}
