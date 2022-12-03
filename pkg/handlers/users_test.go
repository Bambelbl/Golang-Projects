package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"html/template"
	"net/http/httptest"
	"redditclone/pkg/session"
	"redditclone/pkg/user"
	"strings"
	"testing"
)

func TestUserHandler_Index(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	st := user.NewMockUsersRepo(ctrl)

	// Correct
	service := &UserHandler{
		UserRepo: st,
		Logger:   zap.NewNop().Sugar(),
		Tmpl:     template.Must(template.ParseGlob("../../static/html/*")),
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	service.Index(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("expected resp status 200, got %d", resp.StatusCode)
		return
	}

	// Tmpl error
	service.Tmpl = template.Must(template.ParseGlob("../../static/js/*"))
	req = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()

	service.Index(w, req)

	resp = w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("expected resp status 500, got %d", resp.StatusCode)
		return
	}
}

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

func (errReader) Close() error {
	return errors.New("zachem_ya_eto_delau")
}

func TestUserHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	st := user.NewMockUsersRepo(ctrl)
	sess := session.NewMockSessionsRepo(ctrl)

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	dbSess, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer dbSess.Close()

	service := &UserHandler{
		UserRepo:    st,
		SessionRepo: sess,
		Logger:      zap.NewNop().Sugar(),
		Tmpl:        template.Must(template.ParseGlob("../../static/html/*")),
	}

	// Unmarshal error
	req := httptest.NewRequest("POST", "/register", strings.NewReader("mem"))
	w := httptest.NewRecorder()

	service.Register(w, req)

	resp := w.Result()
	if resp.StatusCode != 400 {
		t.Errorf("expected resp status 400, got %d", resp.StatusCode)
		return
	}

	// ErrUserExist
	newUsername := "mem"
	newUserPass := "12345678"

	st.EXPECT().AddUser(gomock.Any(), newUsername, newUserPass).Return(user.ErrUserExist)

	body, err := json.Marshal(map[string]interface{}{
		"username": newUsername,
		"password": newUserPass,
	})
	if err != nil {
		fmt.Println("marshal err")
		return
	}

	req = httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	w = httptest.NewRecorder()

	service.Register(w, req)
	resp = w.Result()
	if resp.StatusCode != 422 {
		t.Errorf("expected resp status 422, got %d", resp.StatusCode)
		return
	}

	// AddUser err
	st.EXPECT().AddUser(gomock.Any(), newUsername, newUserPass).Return(errors.New("kakoy-to prikol"))

	body, err = json.Marshal(map[string]interface{}{
		"username": newUsername,
		"password": newUserPass,
	})
	if err != nil {
		fmt.Println("marshal err")
		return
	}

	req = httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	w = httptest.NewRecorder()

	service.Register(w, req)
	resp = w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("expected resp status 500, got %d", resp.StatusCode)
		return
	}

	// Authorize ErrNoUser
	newUserID := "1"
	resultUser := &user.User{
		ID:       newUserID,
		Username: newUsername,
	}
	st.EXPECT().AddUser(gomock.Any(), newUsername, newUserPass).Return(nil)
	st.EXPECT().Authorize(newUsername, newUserPass).Return(resultUser, user.ErrNoUser)

	body, err = json.Marshal(map[string]interface{}{
		"username": newUsername,
		"password": newUserPass,
	})
	if err != nil {
		fmt.Println("marshal err")
		return
	}

	req = httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	w = httptest.NewRecorder()

	service.Register(w, req)
	resp = w.Result()
	if resp.StatusCode != 400 {
		t.Errorf("expected resp status 400, got %d", resp.StatusCode)
		return
	}

	// Authorize ErrBadPass
	st.EXPECT().AddUser(gomock.Any(), newUsername, newUserPass).Return(nil)
	st.EXPECT().Authorize(newUsername, newUserPass).Return(resultUser, user.ErrBadPass)

	body, err = json.Marshal(map[string]interface{}{
		"username": newUsername,
		"password": newUserPass,
	})
	if err != nil {
		fmt.Println("marshal err")
		return
	}

	req = httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	w = httptest.NewRecorder()

	service.Register(w, req)
	resp = w.Result()
	if resp.StatusCode != 400 {
		t.Errorf("expected resp status 400, got %d", resp.StatusCode)
		return
	}

	// SessionCreate err
	st.EXPECT().AddUser(gomock.Any(), newUsername, newUserPass).Return(nil)
	st.EXPECT().Authorize(newUsername, newUserPass).Return(resultUser, nil)
	sess.EXPECT().Create(*resultUser).Return("", errors.New("kakoy-to prikol"))

	body, err = json.Marshal(map[string]interface{}{
		"username": newUsername,
		"password": newUserPass,
	})
	if err != nil {
		fmt.Println("marshal err")
		return
	}

	req = httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	w = httptest.NewRecorder()

	service.Register(w, req)
	resp = w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("expected resp status 500, got %d", resp.StatusCode)
		return
	}

	// Correct
	st.EXPECT().AddUser(gomock.Any(), newUsername, newUserPass).Return(nil)
	st.EXPECT().Authorize(newUsername, newUserPass).Return(resultUser, nil)
	sess.EXPECT().Create(*resultUser).Return("kektoken", nil)

	body, err = json.Marshal(map[string]interface{}{
		"username": newUsername,
		"password": newUserPass,
	})
	if err != nil {
		fmt.Println("marshal err")
		return
	}

	req = httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	w = httptest.NewRecorder()

	service.Register(w, req)
	resp = w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("expected resp status 200, got %d", resp.StatusCode)
		return
	}

	// ReadAll err
	req = httptest.NewRequest("POST", "/register", errReader(0))
	w = httptest.NewRecorder()

	service.Register(w, req)
	resp = w.Result()
	if resp.StatusCode != 400 {
		t.Errorf("expected resp status 400, got %d", resp.StatusCode)
		return
	}

}

func TestUserHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	st := user.NewMockUsersRepo(ctrl)
	sess := session.NewMockSessionsRepo(ctrl)

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	dbSess, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer dbSess.Close()

	service := &UserHandler{
		UserRepo:    st,
		SessionRepo: sess,
		Logger:      zap.NewNop().Sugar(),
		Tmpl:        template.Must(template.ParseGlob("../../static/html/*")),
	}

	newUserID := "1"
	newUsername := "mem"
	newUserPass := "12345678"

	resultUser := &user.User{
		ID:       newUserID,
		Username: newUsername,
	}

	// Authorize ErrBadPass
	st.EXPECT().Authorize(newUsername, newUserPass).Return(resultUser, user.ErrBadPass)

	body, err := json.Marshal(map[string]interface{}{
		"username": newUsername,
		"password": newUserPass,
	})
	if err != nil {
		fmt.Println("marshal err")
		return
	}

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	service.Login(w, req)
	resp := w.Result()
	if resp.StatusCode != 401 {
		t.Errorf("expected resp status 401, got %d", resp.StatusCode)
		return
	}

	// SessionCreate err
	st.EXPECT().Authorize(newUsername, newUserPass).Return(resultUser, nil)
	sess.EXPECT().Create(*resultUser).Return("", errors.New("kakoy-to prikol"))

	body, err = json.Marshal(map[string]interface{}{
		"username": newUsername,
		"password": newUserPass,
	})
	if err != nil {
		fmt.Println("marshal err")
		return
	}

	req = httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	w = httptest.NewRecorder()

	service.Login(w, req)
	resp = w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("expected resp status 500, got %d", resp.StatusCode)
		return
	}

	// Correct
	st.EXPECT().Authorize(newUsername, newUserPass).Return(resultUser, nil)
	sess.EXPECT().Create(*resultUser).Return("kektoken", nil)

	body, err = json.Marshal(map[string]interface{}{
		"username": newUsername,
		"password": newUserPass,
	})
	if err != nil {
		fmt.Println("marshal err")
		return
	}

	req = httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	w = httptest.NewRecorder()

	service.Login(w, req)
	resp = w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("expected resp status 200, got %d", resp.StatusCode)
		return
	}

	// Unmarshal error
	req = httptest.NewRequest("POST", "/login", strings.NewReader("mem"))
	w = httptest.NewRecorder()

	service.Login(w, req)

	resp = w.Result()
	if resp.StatusCode != 400 {
		t.Errorf("expected resp status 400, got %d", resp.StatusCode)
		return
	}

	// ReadAll err
	req = httptest.NewRequest("POST", "/login", errReader(0))
	w = httptest.NewRecorder()

	service.Login(w, req)
	resp = w.Result()
	if resp.StatusCode != 400 {
		t.Errorf("expected resp status 400, got %d", resp.StatusCode)
		return
	}

}
