package scaler

import "scaling-bot/storage"

type Scaler interface {
	CheckAuth(string, int) error
	ApplyAction(storage.Credentials, storage.Action) error
}