package telegram

type UpdatesResp struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type Update struct {
	Id      int              `json:"update_id"`
	Message *IncomingMessage `json:"message"`
}

type IncomingMessage struct {
	Text string `json:"text"`
	Chat Chat   `json:"chat"`
	From From   `json:"from"`
}

type Chat struct {
	Id int `json:"id"`
}

type From struct {
	Username string `json:"username"`
}
