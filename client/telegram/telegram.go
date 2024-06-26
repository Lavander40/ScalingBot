package telegram

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

const (
	getUpdateMethod     = "getUpdates"
	sendMessageMethod   = "sendMessage"
	deleteMessageMethod = "deleteMessage"
)

type Client struct {
	host     string
	basePath string
	client   http.Client
}

func New(host string, token string) *Client {
	return &Client{
		host:     host,
		basePath: newBasePath(token),
		client:   http.Client{},
	}
}

func newBasePath(token string) string {
	return "bot" + token
}

func (c *Client) Updates(offset int, limit int) ([]Update, error) {
	q := url.Values{}
	q.Add("offset", strconv.Itoa(offset))
	q.Add("limit", strconv.Itoa(limit))

	data, err := c.doRequest(q, getUpdateMethod, http.MethodGet)
	if err != nil {
		return nil, err
	}

	var res UpdatesResp
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	return res.Result, err
}

func (c *Client) SendMessage(chatId int, text string) error {
	q := url.Values{}
	q.Add("chat_id", strconv.Itoa(chatId))
	q.Add("text", text)

	_, err := c.doRequest(q, sendMessageMethod, http.MethodGet)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteMessage(chatId int, messageId int) error {
	q := url.Values{}
	q.Add("chat_id", strconv.Itoa(chatId))
	q.Add("message_id", strconv.Itoa(messageId))

	time.Sleep(10 * time.Second)
	_, err := c.doRequest(q, deleteMessageMethod, http.MethodPost)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) doRequest(query url.Values, tgMethod string, httpMethod string) ([]byte, error) {
	u := url.URL{
		Scheme: "https",
		Host:   c.host,
		Path:   path.Join(c.basePath, tgMethod),
	}

	req, err := http.NewRequest(httpMethod, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = query.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
