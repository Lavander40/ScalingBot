package scaler

import "scaling-bot/storage"

type Scaler interface {
	CheckAuth(storage.Credentials, int) error
	ApplyAction(storage.Credentials, storage.Action) error
	GetStatus(storage.Credentials) (string, error)
}