package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"redditclone/pkg/comment"
	"redditclone/pkg/post"
	"redditclone/pkg/session"
	"redditclone/pkg/user"
	"redditclone/pkg/vote"
	"strings"
	"testing"
	"time"
)

func TestPostsHandler_AllPosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	st := post.NewMockPostsRepo(ctrl)

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	service := &PostsHandler{
		PostsRepo: st,
		Logger:    zap.NewNop().Sugar(),
		Tmpl:      template.Must(template.ParseGlob("../../static/html/*")),
	}

	// GetAll error
	st.EXPECT().GetAll().Return([]*post.Post{}, errors.New("oaoaoao"))

	req := httptest.NewRequest("GET", "/posts/", nil)
	w := httptest.NewRecorder()

	service.AllPosts(w, req)

	resp := w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("expected resp status 500, got %d", resp.StatusCode)
		return
	}

	// Correct
	resPosts := []*post.Post{{ID: "1"}, {ID: "2"}, {ID: "3"}}

	st.EXPECT().GetAll().Return(resPosts, nil)
	req = httptest.NewRequest("POST", "/posts/", nil)
	w = httptest.NewRecorder()

	service.AllPosts(w, req)

	resp = w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("expected resp status 200, got %d", resp.StatusCode)
		return
	}

	var respPosts []post.Post
	body, err := io.ReadAll(w.Body)
	if err != nil {
		http.Error(w, "read body err", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &respPosts)
	if err != nil {
		http.Error(w, "unmarshal err", http.StatusBadRequest)
		return
	}

	if len(respPosts) != len(resPosts) {
		t.Errorf("incorrect result: want len of respPosts = 3, have: %d", len(respPosts))
	}
	if respPosts[0].ID != "1" {
		t.Errorf("incorrect result: want 1-st element ID of resPosts = 1, have: %s", respPosts[0].ID)
	}
}

func TestPostsHandler_CreatePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	st := post.NewMockPostsRepo(ctrl)

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	service := &PostsHandler{
		PostsRepo: st,
		Logger:    zap.NewNop().Sugar(),
		Tmpl:      template.Must(template.ParseGlob("../../static/html/*")),
	}

	author := user.User{
		ID:       "1",
		Username: "mem",
	}

	newPostID := post.RandStringRunes()
	timeCreated := time.Now()
	newPost := post.Post{
		Category:         "sufferings",
		Comments:         &[]comment.Comment{},
		Title:            "reddit",
		Type:             "text",
		Text:             "helpmepls",
		URL:              "",
		Author:           author,
		Created:          timeCreated,
		ID:               newPostID,
		Score:            1,
		UpvotePercentage: 100,
		Views:            0,
		Votes:            &[]vote.Vote{{UserID: author.ID, Vote: 1}},
	}

	// Unmarshal error
	req := httptest.NewRequest("POST", "/posts", strings.NewReader("mem"))
	w := httptest.NewRecorder()

	service.CreatePost(w, req)

	resp := w.Result()
	if resp.StatusCode != 400 {
		t.Errorf("expected resp status 400, got %d", resp.StatusCode)
		return
	}

	// ReadAll err
	req = httptest.NewRequest("POST", "/posts", errReader(0))
	w = httptest.NewRecorder()

	service.CreatePost(w, req)
	resp = w.Result()
	if resp.StatusCode != 400 {
		t.Errorf("expected resp status 400, got %d", resp.StatusCode)
		return
	}

	// Correct
	body, err := json.Marshal(newPost)
	if err != nil {
		fmt.Println(err.Error())
	}

	st.EXPECT().AddPost(author, gomock.Any(), gomock.Any(), gomock.Any()).Return(&newPost)
	req = httptest.NewRequest("POST", "/posts", bytes.NewReader(body))
	w = httptest.NewRecorder()
	sess := session.Session{
		ID:       post.RandStringRunes(),
		UserID:   author.ID,
		Username: author.Username,
		Expires:  time.Now().Add(90 * 24 * time.Hour),
	}
	ctx := session.ContextWithSession(context.TODO(), &sess)
	service.CreatePost(w, req.WithContext(ctx))
	resp = w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("expected resp status 200, got %d", resp.StatusCode)
		return
	}
	var respPost post.Post
	body, err = io.ReadAll(w.Body)
	if err != nil {
		http.Error(w, "read body err", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &respPost)
	if err != nil {
		http.Error(w, "unmarshal err", http.StatusBadRequest)
		return
	}

	if respPost.ID != newPostID {
		t.Errorf("incorrect result: want 1-st element ID of resPosts = 1, have: %s", respPost.ID)
	}

	// Session err
	req = httptest.NewRequest("POST", "/posts", bytes.NewReader(body))
	w = httptest.NewRecorder()
	service.CreatePost(w, req)
	resp = w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("expected resp status 500, got %d", resp.StatusCode)
		return
	}
}

func TestPostsHandler_GetPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	st := post.NewMockPostsRepo(ctrl)

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	service := &PostsHandler{
		PostsRepo: st,
		Logger:    zap.NewNop().Sugar(),
		Tmpl:      template.Must(template.ParseGlob("../../static/html/*")),
	}

	getPost := post.Post{ID: "1"}
	// Correct GetPost
	st.EXPECT().GetPost(gomock.Any(), gomock.Any()).SetArg(1, &getPost)
	req := httptest.NewRequest("GET", "/post/", nil)
	w := httptest.NewRecorder()

	service.GetPost(w, req)
	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("expected resp status 200, got %d", resp.StatusCode)
		return
	}
	var respPost post.Post
	body, err := io.ReadAll(w.Body)
	if err != nil {
		http.Error(w, "read body err", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &respPost)
	if err != nil {
		http.Error(w, "unmarshal err", http.StatusBadRequest)
		return
	}

	if respPost.ID != "1" {
		t.Errorf("incorrect result: want 1-st element ID of resPosts = 1, have: %s", respPost.ID)
	}

	// Err GetPost
	st.EXPECT().GetPost(gomock.Any(), gomock.Any()).Return(errors.New("kakoy-to prikol"))
	req = httptest.NewRequest("GET", "/post/", nil)
	w = httptest.NewRecorder()

	service.GetPost(w, req)
	resp = w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("expected resp status 500, got %d", resp.StatusCode)
		return
	}
}

func TestPostsHandler_GetCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	st := post.NewMockPostsRepo(ctrl)

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	service := &PostsHandler{
		PostsRepo: st,
		Logger:    zap.NewNop().Sugar(),
		Tmpl:      template.Must(template.ParseGlob("../../static/html/*")),
	}

	// Correct GetCategory

	resPosts := []*post.Post{{ID: "1"}, {ID: "2"}, {ID: "3"}}
	st.EXPECT().GetCategory(gomock.Any()).Return(resPosts, nil)
	req := httptest.NewRequest("GET", "/posts/", nil)
	w := httptest.NewRecorder()

	service.GetCategory(w, req)
	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("expected resp status 200, got %d", resp.StatusCode)
		return
	}
	var respPosts []post.Post
	body, err := io.ReadAll(w.Body)
	if err != nil {
		http.Error(w, "read body err", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &respPosts)
	if err != nil {
		http.Error(w, "unmarshal err", http.StatusBadRequest)
		return
	}

	if len(respPosts) != len(resPosts) {
		t.Errorf("incorrect result: want len of respPosts = 3, have: %d", len(respPosts))
	}
	if respPosts[0].ID != "1" {
		t.Errorf("incorrect result: want 1-st element ID of resPosts = 1, have: %s", respPosts[0].ID)
	}

	// Err GetCategory
	st.EXPECT().GetCategory(gomock.Any()).Return(nil, errors.New("kakoy-to prikol"))
	req = httptest.NewRequest("GET", "/posts/", nil)
	w = httptest.NewRecorder()

	service.GetCategory(w, req)
	resp = w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("expected resp status 500, got %d", resp.StatusCode)
		return
	}
}

func TestPostsHandler_CreateComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	st := post.NewMockPostsRepo(ctrl)

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	service := &PostsHandler{
		PostsRepo: st,
		Logger:    zap.NewNop().Sugar(),
		Tmpl:      template.Must(template.ParseGlob("../../static/html/*")),
	}

	author := user.User{
		ID:       "1",
		Username: "mem",
	}

	// Unmarshal error
	req := httptest.NewRequest("POST", "/post/", strings.NewReader("mem"))
	w := httptest.NewRecorder()

	service.CreateComment(w, req)

	resp := w.Result()
	if resp.StatusCode != 400 {
		t.Errorf("expected resp status 400, got %d", resp.StatusCode)
		return
	}

	// ReadAll err
	req = httptest.NewRequest("POST", "/post/", errReader(0))
	w = httptest.NewRecorder()

	service.CreateComment(w, req)
	resp = w.Result()
	if resp.StatusCode != 400 {
		t.Errorf("expected resp status 400, got %d", resp.StatusCode)
		return
	}

	// Correct AddComment
	newComment := struct {
		Comment string `json:"comment"`
	}{Comment: "defrgthyuj"}
	body, err := json.Marshal(newComment)
	if err != nil {
		fmt.Println(err.Error())
	}

	st.EXPECT().AddComment(gomock.Any(), newComment.Comment,
		gomock.Any(), author, gomock.Any(), gomock.Any()).Return(nil).
		SetArg(5, &post.Post{
			Comments: &[]comment.Comment{{ID: "1", Body: newComment.Comment}}})

	req = httptest.NewRequest("POST", "/post/", bytes.NewReader(body))
	w = httptest.NewRecorder()

	sess := session.Session{
		ID:       post.RandStringRunes(),
		UserID:   author.ID,
		Username: author.Username,
		Expires:  time.Now().Add(90 * 24 * time.Hour),
	}
	ctx := session.ContextWithSession(context.TODO(), &sess)
	service.CreateComment(w, req.WithContext(ctx))

	resp = w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("expected resp status 200, got %d", resp.StatusCode)
		return
	}

	var respPost post.Post
	bodyResp, err := io.ReadAll(w.Body)
	if err != nil {
		http.Error(w, "read body err", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bodyResp, &respPost)
	if err != nil {
		http.Error(w, "unmarshal err", http.StatusBadRequest)
		return
	}

	if (*respPost.Comments)[0].Body != newComment.Comment {
		t.Errorf("incorrect result: want text of comment \"defrgthyuj\", have: %s",
			(*respPost.Comments)[0].Body)
	}

	// SessionErr
	req = httptest.NewRequest("POST", "/post/", bytes.NewReader(body))
	w = httptest.NewRecorder()
	service.CreateComment(w, req)
	resp = w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("expected resp status 500, got %d", resp.StatusCode)
		return
	}

	// Err AddComment
	st.EXPECT().AddComment(gomock.Any(), newComment.Comment,
		gomock.Any(), author, gomock.Any(), gomock.Any()).Return(errors.New("kakoy-to prikol"))
	req = httptest.NewRequest("POST", "/post/", bytes.NewReader(body))
	w = httptest.NewRecorder()

	service.CreateComment(w, req.WithContext(ctx))
	resp = w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("expected resp status 500, got %d", resp.StatusCode)
		return
	}
}

func TestPostsHandler_DeleteComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	st := post.NewMockPostsRepo(ctrl)

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	service := &PostsHandler{
		PostsRepo: st,
		Logger:    zap.NewNop().Sugar(),
		Tmpl:      template.Must(template.ParseGlob("../../static/html/*")),
	}

	// Err DeleteComment
	st.EXPECT().DeleteComment(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("kakoy-to prikol"))
	req := httptest.NewRequest("POST", "/post/", nil)
	w := httptest.NewRecorder()

	service.DeleteComment(w, req)
	resp := w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("expected resp status 500, got %d", resp.StatusCode)
		return
	}

	// Correct DeleteComment
	st.EXPECT().DeleteComment(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).
		SetArg(2, &post.Post{ID: "1"})
	req = httptest.NewRequest("POST", "/post/", nil)
	w = httptest.NewRecorder()

	service.DeleteComment(w, req)
	resp = w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("expected resp status 200, got %d", resp.StatusCode)
		return
	}

	var respPost post.Post
	bodyResp, err := io.ReadAll(w.Body)
	if err != nil {
		http.Error(w, "read body err", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bodyResp, &respPost)
	if err != nil {
		http.Error(w, "unmarshal err", http.StatusBadRequest)
		return
	}

	if respPost.ID != "1" {
		t.Errorf("incorrect result: want ID of respPost: 1, have: %s", respPost.ID)
	}
}

func TestPostsHandler_Upvote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	st := post.NewMockPostsRepo(ctrl)

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	service := &PostsHandler{
		PostsRepo: st,
		Logger:    zap.NewNop().Sugar(),
		Tmpl:      template.Must(template.ParseGlob("../../static/html/*")),
	}
	author := user.User{
		ID:       "1",
		Username: "mem",
	}
	sess := session.Session{
		ID:       post.RandStringRunes(),
		UserID:   author.ID,
		Username: author.Username,
		Expires:  time.Now().Add(90 * 24 * time.Hour),
	}
	ctx := session.ContextWithSession(context.TODO(), &sess)

	// Err UpvotePost
	st.EXPECT().UpvotePost(gomock.Any(), author, gomock.Any()).Return(errors.New("kakoy-to prikol"))
	req := httptest.NewRequest("POST", "/post/", nil)
	w := httptest.NewRecorder()

	service.Upvote(w, req.WithContext(ctx))
	resp := w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("expected resp status 500, got %d", resp.StatusCode)
		return
	}

	// Correct UpvotePost
	st.EXPECT().UpvotePost(gomock.Any(), author, gomock.Any()).Return(nil).
		SetArg(2, &post.Post{ID: "1"})
	req = httptest.NewRequest("POST", "/post/", nil)
	w = httptest.NewRecorder()

	service.Upvote(w, req.WithContext(ctx))
	resp = w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("expected resp status 200, got %d", resp.StatusCode)
		return
	}

	var respPost post.Post
	bodyResp, err := io.ReadAll(w.Body)
	if err != nil {
		http.Error(w, "read body err", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bodyResp, &respPost)
	if err != nil {
		http.Error(w, "unmarshal err", http.StatusBadRequest)
		return
	}

	if respPost.ID != "1" {
		t.Errorf("incorrect result: want ID of respPost: 1, have: %s", respPost.ID)
	}

	// Session err
	req = httptest.NewRequest("POST", "/post/", nil)
	w = httptest.NewRecorder()

	service.Upvote(w, req)
	resp = w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("expected resp status 500, got %d", resp.StatusCode)
		return
	}

}

func TestPostsHandler_Downvote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	st := post.NewMockPostsRepo(ctrl)

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	service := &PostsHandler{
		PostsRepo: st,
		Logger:    zap.NewNop().Sugar(),
		Tmpl:      template.Must(template.ParseGlob("../../static/html/*")),
	}
	author := user.User{
		ID:       "1",
		Username: "mem",
	}
	sess := session.Session{
		ID:       post.RandStringRunes(),
		UserID:   author.ID,
		Username: author.Username,
		Expires:  time.Now().Add(90 * 24 * time.Hour),
	}
	ctx := session.ContextWithSession(context.TODO(), &sess)

	// Err DownvotePost
	st.EXPECT().DownvotePost(gomock.Any(), author, gomock.Any()).Return(errors.New("kakoy-to prikol"))
	req := httptest.NewRequest("POST", "/post/", nil)
	w := httptest.NewRecorder()

	service.Downvote(w, req.WithContext(ctx))
	resp := w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("expected resp status 500, got %d", resp.StatusCode)
		return
	}

	// Correct DownvotePost
	st.EXPECT().DownvotePost(gomock.Any(), author, gomock.Any()).Return(nil).
		SetArg(2, &post.Post{ID: "1"})
	req = httptest.NewRequest("POST", "/post/", nil)
	w = httptest.NewRecorder()

	service.Downvote(w, req.WithContext(ctx))
	resp = w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("expected resp status 200, got %d", resp.StatusCode)
		return
	}
	var respPost post.Post
	bodyResp, err := io.ReadAll(w.Body)
	if err != nil {
		http.Error(w, "read body err", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bodyResp, &respPost)
	if err != nil {
		http.Error(w, "unmarshal err", http.StatusBadRequest)
		return
	}

	if respPost.ID != "1" {
		t.Errorf("incorrect result: want ID of respPost: 1, have: %s", respPost.ID)
	}

	// Session err
	req = httptest.NewRequest("POST", "/post/", nil)
	w = httptest.NewRecorder()

	service.Downvote(w, req)
	resp = w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("expected resp status 500, got %d", resp.StatusCode)
		return
	}
}

func TestPostsHandler_Unvote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	st := post.NewMockPostsRepo(ctrl)

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	service := &PostsHandler{
		PostsRepo: st,
		Logger:    zap.NewNop().Sugar(),
		Tmpl:      template.Must(template.ParseGlob("../../static/html/*")),
	}
	author := user.User{
		ID:       "1",
		Username: "mem",
	}
	sess := session.Session{
		ID:       post.RandStringRunes(),
		UserID:   author.ID,
		Username: author.Username,
		Expires:  time.Now().Add(90 * 24 * time.Hour),
	}
	ctx := session.ContextWithSession(context.TODO(), &sess)

	// Err UnvotePost
	st.EXPECT().UnvotePost(gomock.Any(), author, gomock.Any()).Return(errors.New("kakoy-to prikol"))
	req := httptest.NewRequest("POST", "/post/", nil)
	w := httptest.NewRecorder()

	service.Unvote(w, req.WithContext(ctx))
	resp := w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("expected resp status 500, got %d", resp.StatusCode)
		return
	}

	// Correct UnvotePost
	st.EXPECT().UnvotePost(gomock.Any(), author, gomock.Any()).Return(nil).
		SetArg(2, &post.Post{ID: "1"})
	req = httptest.NewRequest("POST", "/post/", nil)
	w = httptest.NewRecorder()

	service.Unvote(w, req.WithContext(ctx))
	resp = w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("expected resp status 200, got %d", resp.StatusCode)
		return
	}

	var respPost post.Post
	bodyResp, err := io.ReadAll(w.Body)
	if err != nil {
		http.Error(w, "read body err", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bodyResp, &respPost)
	if err != nil {
		http.Error(w, "unmarshal err", http.StatusBadRequest)
		return
	}

	if respPost.ID != "1" {
		t.Errorf("incorrect result: want ID of respPost: 1, have: %s", respPost.ID)
	}

	// Session err
	req = httptest.NewRequest("POST", "/post/", nil)
	w = httptest.NewRecorder()

	service.Unvote(w, req)
	resp = w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("expected resp status 500, got %d", resp.StatusCode)
		return
	}

}

func TestPostsHandler_DeletePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	st := post.NewMockPostsRepo(ctrl)

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	service := &PostsHandler{
		PostsRepo: st,
		Logger:    zap.NewNop().Sugar(),
		Tmpl:      template.Must(template.ParseGlob("../../static/html/*")),
	}

	// Err DeletePost
	st.EXPECT().DeletePost(gomock.Any()).Return(errors.New("kakoy-to prikol"))
	req := httptest.NewRequest("POST", "/post/", nil)
	w := httptest.NewRecorder()

	service.DeletePost(w, req)
	resp := w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("expected resp status 500, got %d", resp.StatusCode)
		return
	}

	// Correct DeletePost
	st.EXPECT().DeletePost(gomock.Any()).Return(nil)
	req = httptest.NewRequest("POST", "/post/", nil)
	w = httptest.NewRecorder()

	service.DeletePost(w, req)
	resp = w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("expected resp status 200, got %d", resp.StatusCode)
		return
	}
	res := struct {
		Message string `json:"message"`
	}{}
	bodyResp, err := io.ReadAll(w.Body)
	if err != nil {
		http.Error(w, "read body err", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bodyResp, &res)
	if err != nil {
		http.Error(w, "unmarshal err", http.StatusBadRequest)
		return
	}

	if res.Message != "success" {
		t.Errorf("incorrect result: want \"success\", have: %s", res.Message)
	}
}

func TestPostsHandler_GetUserPosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	st := post.NewMockPostsRepo(ctrl)

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	service := &PostsHandler{
		PostsRepo: st,
		Logger:    zap.NewNop().Sugar(),
		Tmpl:      template.Must(template.ParseGlob("../../static/html/*")),
	}

	// Err GetUserPosts
	st.EXPECT().GetUserPosts(gomock.Any()).Return(nil, errors.New("kakoy-to prikol"))
	req := httptest.NewRequest("POST", "/post/", nil)
	w := httptest.NewRecorder()

	service.GetUserPosts(w, req)
	resp := w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("expected resp status 500, got %d", resp.StatusCode)
		return
	}

	// Correct GetUserPosts
	resPosts := []*post.Post{{ID: "1"}, {ID: "2"}, {ID: "3"}}
	st.EXPECT().GetUserPosts(gomock.Any()).Return(resPosts, nil)
	req = httptest.NewRequest("POST", "/post/", nil)
	w = httptest.NewRecorder()

	service.GetUserPosts(w, req)
	resp = w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("expected resp status 200, got %d", resp.StatusCode)
		return
	}
	var respPosts []post.Post
	body, err := io.ReadAll(w.Body)
	if err != nil {
		http.Error(w, "read body err", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &respPosts)
	if err != nil {
		http.Error(w, "unmarshal err", http.StatusBadRequest)
		return
	}

	if len(respPosts) != len(resPosts) {
		t.Errorf("incorrect result: want len of respPosts = 3, have: %d", len(respPosts))
	}
	if respPosts[0].ID != "1" {
		t.Errorf("incorrect result: want 1-st element ID of resPosts = 1, have: %s", respPosts[0].ID)
	}
}
