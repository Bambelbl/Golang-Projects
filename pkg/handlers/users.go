package handlers

import (
	"encoding/json"
	"go.uber.org/zap"
	"html/template"
	"io"
	"net/http"
	"redditclone/pkg/session"
	"redditclone/pkg/user"
)

type UserHandler struct {
	Tmpl        *template.Template
	UserRepo    user.UsersRepo
	SessionRepo session.SessionsRepo
	Logger      *zap.SugaredLogger
	SigString   string
}

func (h *UserHandler) Index(w http.ResponseWriter, _ *http.Request) {
	err := h.Tmpl.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, `Template errror`, http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			http.Error(w, "body close err", http.StatusInternalServerError)
		}
	}(r.Body)
	var newUser = struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body err", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &newUser)
	if err != nil {
		http.Error(w, "cant unpack payload", http.StatusBadRequest)
		return
	}

	err = h.UserRepo.AddUser(user.RandStringRunes(), newUser.Username, newUser.Password)
	w.Header().Add("Content-Type", "application/json")
	switch err {
	case user.ErrUserExist:
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	case nil:
		u, err := h.UserRepo.Authorize(newUser.Username, newUser.Password)
		if err == user.ErrNoUser {
			http.Error(w, `no user`, http.StatusBadRequest)
			return
		}
		if err == user.ErrBadPass {
			http.Error(w, `bad pass`, http.StatusBadRequest)
			return
		}
		token, err := h.SessionRepo.Create(*u)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp, err := json.Marshal(map[string]interface{}{
			"token": token,
		})
		if err != nil {
			http.Error(w, `marshal err`, http.StatusInternalServerError)
		}
		_, err = w.Write(resp)
		if err != nil {
			http.Error(w, `Write err`, http.StatusInternalServerError)
		}
	default:
		http.Error(w, `AddUser err`, http.StatusInternalServerError)
	}
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			http.Error(w, "body close err", http.StatusInternalServerError)
		}
	}(r.Body)
	type User struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var loginUser User
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body err", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &loginUser)
	if err != nil {
		http.Error(w, "cant unpack payload", http.StatusBadRequest)
		return
	}
	u, err := h.UserRepo.Authorize(loginUser.Username, loginUser.Password)
	if err != nil {
		http.Error(w, `bad pass`, http.StatusUnauthorized)
		return
	}
	token, err := h.SessionRepo.Create(*u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(map[string]interface{}{
		"token": token,
	})
	if err != nil {
		http.Error(w, `marshal err`, http.StatusInternalServerError)
	}
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, `Write err`, http.StatusInternalServerError)
	}
}
