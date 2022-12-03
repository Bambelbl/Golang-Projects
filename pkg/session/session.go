package session

import (
	"context"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"redditclone/pkg/user"
	"time"
)

type Session struct {
	ID       string
	UserID   string
	Username string
	Expires  time.Time
}

func NewSession(newUser user.User) (*Session, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": jwt.MapClaims{
			"username": newUser.Username,
			"id":       newUser.ID,
		},
		"iat": time.Now().Unix(),
		"exp": time.Now().AddDate(0, 0, 7).Unix(),
	})
	tokenString, err := token.SignedString([]byte("kakoy-to ptikol"))
	if err != nil {
		fmt.Println("session")
		return nil, err
	}
	return &Session{
		ID:       tokenString,
		UserID:   newUser.ID,
		Username: newUser.Username,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	}, nil
}

var sessionKey = "sessionKey"

func SessFromContext(ctx context.Context) (*Session, error) {
	sess, ok := ctx.Value(sessionKey).(*Session)
	if !ok || sess == nil {
		return nil, errors.New("no session found")
	}
	if sess.Expires.Before(time.Now()) {
		return nil, errors.New("session is expired")
	}
	return sess, nil
}

func ContextWithSession(ctx context.Context, sess *Session) context.Context {
	return context.WithValue(ctx, sessionKey, sess)
}

// SessionsRepo go install github.com/golang/mock/mockgen@v1.6.0
//
//go:generate mockgen -source=session.go -destination=repo_mock.go -package=session SessionsRepo
type SessionsRepo interface {
	Check(string) (*Session, error)
	Create(user.User) (string, error)
}
