package telegram

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	ep "scaling-bot/event_processor"
	"scaling-bot/storage"
	"strconv"
	"strings"
	"time"
)

const (
	StartCmd = "/start"
	HelpCmd  = "/help"
	TokenCmd = "/token"
	AddCmd   = "/add"
	RmCmd    = "/rm"
	LimitCmd = "/limit"
	LastCmd  = "/last"
	StatusCmd = "/status"
)

func (p *Processor) doCmd(text string, chatId int, userName string, messageId int) error {
	text = strings.TrimSpace(text)

	log.Printf("run commant %s, by %s", text, userName)

	if isConfig(text) {
		conf := strings.Split(text[strings.LastIndex(text, " ") + 1 :], ":")

		credentials := storage.Credentials{
			UserId:   chatId,
			CloudId:  conf[0],
			AuthToken: "Bearer " + conf[1],
		}

		err := p.scaler.CheckAuth(credentials, chatId)
		if err != nil {
			return p.tg.SendMessage(chatId, ep.No_connect)
		}
		err = p.storage.SetCred(credentials)
		if err != nil {
			return err
		}

		_ = p.tg.DeleteMessage(chatId, messageId)
		
		return p.tg.SendMessage(chatId, ep.Connect_msg)
	}

	switch text {
	case StartCmd:
		return p.tg.SendMessage(chatId, ep.Welcome_msg)
	case HelpCmd:
		return p.tg.SendMessage(chatId, ep.Help_msg)
	case TokenCmd:
		return p.tg.SendMessage(chatId, ep.No_amount)
	}

	// проверка на наличие даннвых подлключения к облаку
	credentials, err := p.storage.GetCred(chatId)
	if err != nil {
		return p.tg.SendMessage(chatId, ep.Unidentified_msg)
	}

	if isAmount(text) {
		amount, _ := strconv.Atoi(text[strings.LastIndex(text, " ")+1:])
		if amount > 10 { amount = 10 }
		res, err := p.getLast(credentials, amount)
		if err != nil {
			return p.tg.SendMessage(chatId, ep.Fail_msg)
		}
		return p.tg.SendMessage(credentials.UserId, res)
	}

	if isLimit(text) {
		return p.setLimit(text, credentials, userName)
	}

	switch text {
	case AddCmd:
		return p.changeInstance(credentials, userName, 1)
	case RmCmd:
		return p.changeInstance(credentials, userName, -1)
	case StatusCmd:
		return p.getStatus(credentials)
	case LastCmd:
		res, err := p.getLast(credentials, 1)
		if err != nil {
			return p.tg.SendMessage(chatId, ep.Fail_msg)
		}
		return p.tg.SendMessage(chatId, res)
	case LimitCmd, TokenCmd:
		return p.tg.SendMessage(chatId, ep.No_amount)
	default:
		return p.tg.SendMessage(chatId, ep.Unknown_msg)
	}
}

func isConfig(text string) bool {
	match, _ := regexp.MatchString("^/token [a-zA-Z0-9_-]+:t1.[A-Z0-9a-z_-]+[=]{0,2}.[A-Z0-9a-z_-]{86}[=]{0,2}$", text)
	return match
}

func isAmount(text string) bool {
	match, _ := regexp.MatchString("^/last ?[0-9]+$", text)
	return match
}

func isLimit(text string) bool {
	match, _ := regexp.MatchString("^/limit ?[0-9]+$", text)
	return match
}

func (p *Processor) changeInstance(credentials storage.Credentials, userName string, amount int) error {
	call := &storage.Action{
		CloudId:    credentials.CloudId,
		Type:      1,
		Amount:    amount,
		UserName:  userName,
		CreatedAt: time.Now(),
	}

	err := p.scaler.ApplyAction(credentials, *call)
	if errors.Is(err, storage.ErrOutOfLimit) {
		_ = p.tg.SendMessage(credentials.UserId, ep.Not_enough)
		return err
	}
	if err != nil {
		_ = p.tg.SendMessage(credentials.UserId, ep.Fail_msg)
		return err
	}

	if err := p.storage.SaveAction(call); err != nil {
		_ = p.tg.SendMessage(credentials.UserId, ep.Fail_msg)
		return err
	}

	return p.tg.SendMessage(credentials.UserId, ep.Wait_msg)
}

func (p *Processor) getStatus(credentials storage.Credentials) error {
	msg, err := p.scaler.GetStatus(credentials)
	if err != nil {
		msg = "Связь с облаком не установлена, проверьте срок жизни токена"
	}
	return p.tg.SendMessage(credentials.UserId, msg)
}

func (p *Processor) setLimit(text string, credentials storage.Credentials, userName string) error {
	amount, _ := strconv.Atoi(text[strings.LastIndex(text, " ")+1:])
	if amount < 1 || amount > 100 {
		return p.tg.SendMessage(credentials.UserId, ep.Impossible_percent_msg)
	}

	call := &storage.Action{
		CloudId:    credentials.CloudId,
		Type:      2,
		Amount:    amount,
		UserName:  userName,
		CreatedAt: time.Now(),
	}

	if err := p.storage.SaveAction(call); err != nil {
		_ = p.tg.SendMessage(credentials.UserId, ep.Fail_msg)
		return err
	}

	return p.tg.SendMessage(credentials.UserId, ep.Sucess_msg)
}

func (p *Processor) getLast(credentials storage.Credentials, amount int) (string, error) {
	calls, err := p.storage.GetActions(credentials.CloudId, amount)
	if errors.Is(err, storage.ErrEmpty) {
		return ep.No_found_msg, nil
	}
	if err != nil {
		return "", err
	}

	var res string
	for _, call := range calls {
		switch call.Type {
		case 1:
			if call.Amount > 0 {
				res += fmt.Sprintf("Был добавлен экземпляр приложения\nПользователь: %s\nВремя: %v\n\n", call.UserName, call.CreatedAt)
			} else {
				res += fmt.Sprintf("Был удалён экземпляр приложения\nПользователь: %s\nВремя: %v\n\n", call.UserName, call.CreatedAt)
			}
		case 2:
			res += fmt.Sprintf("Установлен лимит загрузки ОЗУ (%%): %d\nПользователь: %s\nВремя: %v\n\n", call.Amount, call.UserName, call.CreatedAt)
		default:
			return "", storage.ErrUnknownType
		}
	}

	return res, nil
}
