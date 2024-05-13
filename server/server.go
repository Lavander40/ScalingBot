package server

import (
	"fmt"
	"io"
	"net/http"
)

type Server struct {}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Start() (err error) {
	http.HandleFunc("/webhook", alertHandler)
    return http.ListenAndServe(":6060", nil)
}

func alertHandler(w http.ResponseWriter, r *http.Request) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Failed to read request body", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    // Обработка тела запроса (JSON) с данными об уведомлении
    fmt.Println("Received alert:")
    fmt.Println(string(body))

    // В этом месте вы можете сделать что-то с данными об уведомлении, например, отправить их в лог или обработать как-то еще
}