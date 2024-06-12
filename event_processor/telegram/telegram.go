package telegram

import (
	"errors"
	"scaling-bot/client/telegram"
	"scaling-bot/scaler"
	"scaling-bot/storage"
	ep "scaling-bot/event_processor"
)

type Processor struct {
	tg      *telegram.Client
	offset  int
	storage storage.Storage
	scaler  scaler.Scaler
}

type Meta struct {
	ChatId   int
	Username string
	MessageId int
}

func New(client *telegram.Client, storage storage.Storage, scaler scaler.Scaler) *Processor {
	return &Processor{
		tg:      client,
		offset:  0,
		storage: storage,
		scaler: scaler,
	}
}

func (p *Processor) Fetch(limit int) ([]ep.Event, error) {
	updates, err := p.tg.Updates(p.offset, limit)
	if err != nil {
		return nil, err
	}

	if len(updates) == 0 {
		return nil, nil
	}

	res := make([]ep.Event, 0, len(updates))

	for _, u := range updates {
		res = append(res, event(u))
	}

	p.offset = updates[len(updates)-1].Id + 1

	return res, nil
}

func (p *Processor) Process(event ep.Event) error {
	switch event.Type {
	case ep.Message:
		return p.processMessage(event)
	default:
		return errors.New("unknown event type")
	}
}

func (p *Processor) processMessage(event ep.Event) error {
	meta, err := fetchMeta(event)
	if err != nil {
		return err
	}
	if err := p.doCmd(event.Text, meta.ChatId, meta.Username, meta.MessageId); err != nil {
		return err
	}

	return nil
}

func fetchMeta(event ep.Event) (Meta, error) {
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, errors.New("can't fetch meta")
	}

	return res, nil
}

func event(update telegram.Update) ep.Event {
	uType := fetchType(update)

	res := ep.Event{
		Type: uType,
		Text: fetchText(update),
	}

	if uType == ep.Message {
		res.Meta = Meta{
			ChatId:   update.Message.Chat.Id,
			Username: update.Message.From.Username,
			MessageId: update.Message.MessageId,
		}
	}

	return res
}

func fetchType(update telegram.Update) ep.Type {
	if update.Message == nil {
		return ep.Unknown
	}
	return ep.Message
}

func fetchText(update telegram.Update) string {
	if update.Message == nil {
		return ""
	}
	return update.Message.Text
}
