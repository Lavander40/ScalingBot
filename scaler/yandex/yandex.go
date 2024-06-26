package yandex

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"scaling-bot/storage"
	"strconv"
)

type Scaler struct{}

type GetResponse struct {
	ScalePolicy struct {
		FixedScale struct {
			Size string `json:"size"`
		} `json:"fixedScale"`
	} `json:"scalePolicy"`
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type PatchResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func New() *Scaler {
	return &Scaler{}
}

func (s *Scaler) CheckAuth(credentials storage.Credentials, chatId int) error {
	req, err := http.NewRequest("GET", "https://mks.api.cloud.yandex.net/managed-kubernetes/v1/nodeGroups/"+credentials.CloudId, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", credentials.AuthToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return storage.ErrCheckConnect
	}

	return nil
}

func (s *Scaler) getAmount(credentials storage.Credentials) (int, error) {
	req, err := http.NewRequest("GET", "https://mks.api.cloud.yandex.net/managed-kubernetes/v1/nodeGroups/" + credentials.CloudId, nil)
	if err != nil {
		return 0, err
	}

	// set content-type header to JSON
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", credentials.AuthToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var target GetResponse
	if err := json.NewDecoder(resp.Body).Decode(&target); err != nil {
		return 0, err
	}

	size, _ := strconv.Atoi(target.ScalePolicy.FixedScale.Size)

	return size, nil
}

func (s *Scaler) GetStatus(credentials storage.Credentials) (string, error) {
	req, err := http.NewRequest("GET", "https://mks.api.cloud.yandex.net/managed-kubernetes/v1/nodeGroups/" + credentials.CloudId, nil)
	if err != nil {
		return "", err
	}

	// set content-type header to JSON
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", credentials.AuthToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(strconv.Itoa(resp.StatusCode))
	}

	var target GetResponse
	if err := json.NewDecoder(resp.Body).Decode(&target); err != nil {
		return "", err
	}

	return fmt.Sprintf("Подключение к облаку %s присутствует\nИдентификатор кластера: %s\nЧисло узлов в кластере: %s\nСостояние системы: %s\n", target.Name, target.ID, target.ScalePolicy.FixedScale.Size, target.Status), nil
}

func (s *Scaler) ApplyAction(credentials storage.Credentials, call storage.Action) error {
	size, err := s.getAmount(credentials)
	if err != nil {
		return err
	}
	if size == 1 && call.Amount == -1 {
		return storage.ErrOutOfLimit
	}

	payload := `{
		"updateMask": "scalePolicy.fixedScale.size",
		"scalePolicy": {
			"fixedScale": {
				"size": "` + strconv.Itoa(size+call.Amount) + `"
			}
		}
	}`

	req, err := http.NewRequest("PATCH", "https://mks.api.cloud.yandex.net/managed-kubernetes/v1/nodeGroups/" + credentials.CloudId, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		panic(err)
	}

	// set content-type header to JSON
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", credentials.AuthToken)

	// create HTTP client and execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var target PatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&target); err != nil {
		return err
	}
	if target.Error.Message != "" {
		return fmt.Errorf("error in resp: %s", target.Error.Message)
	}

	return nil
}
