package user

import (
	"database/sql"
	"errors"
	"math/rand"
)

var (
	ErrNoUser    = errors.New("no user found")
	ErrBadPass   = errors.New("invald password")
	ErrUserExist = errors.New("user already exist")
	ErrInternal  = errors.New("internal error")
)

type UsersMySQLRepository struct {
	DB *sql.DB
}

func NewMySQLRepo(db *sql.DB) *UsersMySQLRepository {
	return &UsersMySQLRepository{DB: db}
}

func (repo *UsersMySQLRepository) Authorize(login, pass string) (*User, error) {
	user := &User{}
	err := repo.DB.
		QueryRow("SELECT id, username, pass FROM users WHERE username = ?", login).
		Scan(&user.ID, &user.Username, &user.password)
	if err == sql.ErrNoRows {
		return nil, ErrNoUser
	} else if err != nil {
		return nil, ErrInternal
	}
	if user.password != pass {
		return nil, ErrBadPass
	}
	return user, nil
}

var (
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func RandStringRunes() string {
	b := make([]rune, 24)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (repo *UsersMySQLRepository) AddUser(id, login, pass string) error {
	user := &User{}
	err := repo.DB.
		QueryRow("SELECT id, username, pass FROM users WHERE username = ?", login).
		Scan(&user.ID, &user.Username, &user.password)
	switch err {
	case sql.ErrNoRows:
		_, err := repo.DB.Exec(
			"INSERT INTO users (`id`, `username`, `pass`) VALUES (?, ?, ?)",
			id,
			login,
			pass)
		if err != nil {
			return ErrInternal
		}
		return nil
	case nil:
		return ErrUserExist
	default:
		return ErrInternal
	}
}
