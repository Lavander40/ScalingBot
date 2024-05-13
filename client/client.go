package client

import "net/url"

// client responsible for getting updates from users and sending message to them
type Client interface {
	Updates(int, int) ([]interface{}, error)
	SendMessage(int, string) error // user_id, msg
	doRequest(url.Values, string) ([]byte, error)  // values method
}
