package sqlite

import (
	"context"
	"database/sql"
	"scaling-bot/storage"
	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db  *sql.DB
	ctx context.Context
}

func New(ctx context.Context, path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &Storage{db: db, ctx: ctx}, nil
}

func (s Storage) Init() error {
	q := `CREATE TABLE IF NOT EXISTS 
	calls (call_id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, cloud_id INTEGER, type INTEGER, amount INTEGER, user_name TEXT, created_at TIMESTAMP);
	CREATE TABLE IF NOT EXISTS 
	credentials (user_id INTEGER PRIMARY KEY, api_token TEXT, cloud_id TEXT);`
	
	_, err := s.db.ExecContext(s.ctx, q)
	if err != nil {
		return err
	}

	return nil
}

func (s Storage) SaveAction(a *storage.Action) error {
	q := `INSERT INTO calls (cloud_id, type, amount, user_name, created_at) VALUES (?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(s.ctx, q, a.CloudId, a.Type, a.Amount, a.UserName, a.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (s Storage) GetActions(groupId string, amount int) ([]*storage.Action, error) {
	q := `SELECT type, amount, user_name, created_at FROM calls WHERE cloud_id = ? ORDER BY created_at ASC LIMIT ?`

	rows, err := s.db.QueryContext(s.ctx, q, groupId, amount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calls []*storage.Action
	for rows.Next() {
		var call storage.Action
		if err := rows.Scan(&call.Type, &call.Amount, &call.UserName, &call.CreatedAt); err != nil {
			return nil, err
		}
		calls = append(calls, &call)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(calls) == 0 {
		return nil, storage.ErrEmpty
	}

	return calls, nil
}

func (s Storage) SetCred(c storage.Credentials) error {
	q := `INSERT OR IGNORE INTO credentials (user_id, api_token, cloud_id) VALUES (?, ?, ?); UPDATE credentials SET api_token = ?, cloud_id = ? WHERE user_id = ?`

	_, err := s.db.ExecContext(s.ctx, q, c.UserId, c.AuthToken, c.CloudId, c.AuthToken, c.CloudId, c.UserId)
	if err != nil {
		return err
	}

	return nil
}

func (s Storage) GetCred(chatId int) (storage.Credentials, error) {
	c := storage.Credentials{
		UserId: chatId,
	}

	q := `SELECT api_token, cloud_id FROM credentials WHERE user_id = ?`

	err := s.db.QueryRowContext(s.ctx, q, chatId).Scan(&c.AuthToken, &c.CloudId)
	if err != nil {
		return c, err
	}

	return c, nil
}

func (s Storage) GetUserByCloud(cloudId string) ([]int, error) {
	var users []int
	q := `SELECT user_id FROM credentials WHERE cloud_id = ?`

	rows, err := s.db.QueryContext(s.ctx, q, cloudId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user int
		if err := rows.Scan(&user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
