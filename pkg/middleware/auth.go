package middleware

import (
	"net/http"
	"redditclone/pkg/session"
	"strings"
)

func CheckAuth(sm *session.SessionsMySQLRepository, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inToken := strings.Split(r.Header.Get("Authorization"), " ")[1]
		sess, err := sm.Check(inToken)
		if err != nil {
			http.Error(w, `not auth`, http.StatusUnauthorized)
		}
		ctx := session.ContextWithSession(r.Context(), sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
