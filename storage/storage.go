package storage

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"time"
)

var (
	ErrEmpty        = errors.New("no saved logs")
	ErrOutOfLimit   = errors.New("amount is out of set limit")
	ErrUnknownType  = errors.New("unknown type of log")
	ErrCheckConnect = errors.New("connect is unsuccessful")
)

type Storage interface {
	SaveAction(*Action) error
	GetActions(string, int) ([]*Action, error) // cloud_is, amount
	SetCred(Credentials) error
	GetCred(int) (Credentials, error) // user_id
	GetUserByCloud(string) ([]int, error)
}

type Action struct {
	CloudId   string
	UserName  string
	Type      int // for response parsing purposes
	Amount    int
	CreatedAt time.Time
}

type Credentials struct {
	UserId    int
	CloudId   string
	AuthToken string
}

func (a *Action) Hash() (string, error) {
	h := sha1.New()

	if _, err := io.WriteString(h, a.CreatedAt.String()); err != nil {
		return "", err
	}

	if _, err := io.WriteString(h, a.UserName); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
