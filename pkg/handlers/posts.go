package handlers

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"html/template"
	"io"
	"net/http"
	"redditclone/pkg/post"
	"redditclone/pkg/session"
	"redditclone/pkg/user"
	"time"
)

type PostsHandler struct {
	Tmpl        *template.Template
	PostsRepo   post.PostsRepo
	SessionRepo session.SessionsRepo
	Logger      *zap.SugaredLogger
}

func WriteResponse(w http.ResponseWriter, body any) error {
	resJSON, err := json.Marshal(body)
	if err != nil {
		return errors.New(`incorrect JSON err`)
	}
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(resJSON)
	if err != nil {
		return errors.New(`write err`)
	}
	return nil
}

func (h *PostsHandler) AllPosts(w http.ResponseWriter, _ *http.Request) {
	elems, err := h.PostsRepo.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = WriteResponse(w, elems)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PostsHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			http.Error(w, "body close err", http.StatusInternalServerError)
		}
	}(r.Body)
	var newPost post.Post
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body err", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &newPost)
	if err != nil {
		http.Error(w, "unmarshal err", http.StatusBadRequest)
		return
	}
	sess, err := session.SessFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	item := h.PostsRepo.AddPost(user.User{ID: sess.UserID, Username: sess.Username},
		newPost, post.RandStringRunes(), time.Now())
	w.Header().Add("Content-Type", "application/json")
	err = WriteResponse(w, item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PostsHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var resPost *post.Post
	err := h.PostsRepo.GetPost(vars["postID"], &resPost)
	if err != nil {
		http.Error(w, `DB err`, http.StatusInternalServerError)
		return
	}
	err = WriteResponse(w, resPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PostsHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	items, err := h.PostsRepo.GetCategory(vars["category"])
	if err != nil {
		http.Error(w, `DB err`, http.StatusInternalServerError)
		return
	}
	err = WriteResponse(w, items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PostsHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			http.Error(w, "body close err", http.StatusInternalServerError)
		}
	}(r.Body)
	vars := mux.Vars(r)

	bodyComment := struct {
		Comment string `json:"comment"`
	}{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body err", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &bodyComment)
	if err != nil {
		http.Error(w, "unmarshal err", http.StatusBadRequest)
		return
	}

	sess, err := session.SessFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	var resPost *post.Post
	err = h.PostsRepo.AddComment(vars["postID"], bodyComment.Comment, time.Now(),
		user.User{ID: sess.UserID, Username: sess.Username}, post.RandStringRunes(), &resPost)
	if err != nil {
		http.Error(w, `DB err`, http.StatusInternalServerError)
		return
	}
	err = WriteResponse(w, resPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PostsHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var resPost *post.Post
	err := h.PostsRepo.DeleteComment(vars["postID"], vars["commentID"], &resPost)
	if err != nil {
		http.Error(w, `DB err`, http.StatusInternalServerError)
		return
	}
	err = WriteResponse(w, resPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PostsHandler) Upvote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	sess, err := session.SessFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var resPost *post.Post
	err = h.PostsRepo.UpvotePost(vars["postID"], user.User{ID: sess.UserID, Username: sess.Username}, &resPost)
	if err != nil {
		http.Error(w, `DB err`, http.StatusInternalServerError)
		return
	}
	err = WriteResponse(w, resPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PostsHandler) Downvote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	sess, err := session.SessFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var resPost *post.Post
	err = h.PostsRepo.DownvotePost(vars["postID"], user.User{ID: sess.UserID, Username: sess.Username}, &resPost)
	if err != nil {
		http.Error(w, `DB err`, http.StatusInternalServerError)
		return
	}
	err = WriteResponse(w, resPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PostsHandler) Unvote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	sess, err := session.SessFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var resPost *post.Post
	err = h.PostsRepo.UnvotePost(vars["postID"], user.User{ID: sess.UserID, Username: sess.Username}, &resPost)
	if err != nil {
		http.Error(w, `DB err`, http.StatusInternalServerError)
		return
	}
	err = WriteResponse(w, resPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PostsHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	err := h.PostsRepo.DeletePost(vars["postID"])
	if err != nil {
		http.Error(w, `DB err`, http.StatusInternalServerError)
		return
	}
	err = WriteResponse(w, map[string]interface{}{
		"message": "success",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PostsHandler) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	items, err := h.PostsRepo.GetUserPosts(vars["username"])
	if err != nil {
		http.Error(w, `DB err`, http.StatusInternalServerError)
		return
	}
	err = WriteResponse(w, items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
