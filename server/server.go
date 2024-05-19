package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j/log"
)

type Server struct {}

type CloudRequest struct {
	Receiver string `json:"receiver"`
	Status   string `json:"status"`
	Alerts   []struct {
		Status string `json:"status"`
		Labels struct {
			Alertname     string `json:"alertname"`
			GrafanaFolder string `json:"grafana_folder"`
			Policy        string `json:"policy"`
		} `json:"labels"`
		Annotations struct {
		} `json:"annotations"`
		GeneratorURL string    `json:"generatorURL"`
		Fingerprint  string    `json:"fingerprint"`
		SilenceURL   string    `json:"silenceURL"`
		DashboardURL string    `json:"dashboardURL"`
		PanelURL     string    `json:"panelURL"`
		Values       struct {
			A float64 `json:"A"`
			C int     `json:"C"`
		} `json:"values"`
		ValueString string `json:"valueString"`
	} `json:"alerts"`
	GroupLabels struct {
		Alertname     string `json:"alertname"`
		GrafanaFolder string `json:"grafana_folder"`
	} `json:"groupLabels"`
	CommonLabels struct {
		Alertname     string `json:"alertname"`
		GrafanaFolder string `json:"grafana_folder"`
		Policy        string `json:"policy"`
	} `json:"commonLabels"`
	CommonAnnotations struct {
	} `json:"commonAnnotations"`
	GroupKey        string `json:"groupKey"`
	Title           string `json:"title"`
	State           string `json:"state"`
	Message         string `json:"message"`
}

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

    HandleRequest(body)
}

func HandleRequest(req []byte) {
	var jsonReq CloudRequest
	err := json.Unmarshal(req, &jsonReq)
	if err != nil {
		log.ERROR(err)
	}
	fmt.Println(jsonReq.Message)
}