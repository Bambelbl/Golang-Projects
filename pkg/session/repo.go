package session

import (
	"database/sql"
	"errors"
	"redditclone/pkg/user"
	"time"
)

type SessionsMySQLRepository struct {
	DB *sql.DB
}

func NewMySQLRepo(db *sql.DB) *SessionsMySQLRepository {
	return &SessionsMySQLRepository{DB: db}
}

func (sm *SessionsMySQLRepository) Check(id string) (*Session, error) {
	sess := &Session{}
	err := sm.DB.
		QueryRow("SELECT id, userid, username, expires FROM sessions WHERE id = ?", id).
		Scan(&sess.ID, &sess.UserID, &sess.Username, &sess.Expires)
	if err != nil {
		return nil, err
	}
	if sess.Expires.Before(time.Now()) {
		return nil, errors.New("session expired")
	}
	return sess, nil
}

func (sm *SessionsMySQLRepository) Create(newUser user.User) (string, error) {
	sess, err := NewSession(newUser)
	if err != nil {
		return "", errors.New(`new session err`)
	}
	_, err = sm.DB.Exec(
		"INSERT INTO sessions (`id`, `userid`, `username`, `expires`) VALUES (?, ?, ?, ?)",
		sess.ID,
		sess.UserID,
		sess.Username,
		sess.Expires,
	)
	if err != nil {
		return "", errors.New(`db err`)
	}
	return sess.ID, nil
}
