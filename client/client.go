package client

// client responsible for getting updates from users and sending message to them
type Client interface {
	Updates(int, int) ([]interface{}, error)
	SendMessage(int, string, string) error // user_id, msg
	DeleteMessage(int, int) error
	//doRequest(url.Values, string) ([]byte, error)  // values method
}
