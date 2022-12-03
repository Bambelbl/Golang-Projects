package post

import (
	"redditclone/pkg/comment"
	"redditclone/pkg/user"
	"redditclone/pkg/vote"
	"time"
)

type Post struct {
	Score            int                `json:"score" bson:"score"`
	Views            int                `json:"views" bson:"views"`
	Type             string             `json:"type" bson:"type"`
	Title            string             `json:"title" bson:"title"`
	Author           user.User          `json:"author" bson:"author"`
	Category         string             `json:"category" bson:"category"`
	Text             string             `json:"text,omitempty" bson:"text"`
	URL              string             `json:"url,omitempty" bson:"url"`
	Votes            *[]vote.Vote       `json:"votes" bson:"votes"`
	Comments         *[]comment.Comment `json:"comments" bson:"comments"`
	Created          time.Time          `json:"created" bson:"created"`
	UpvotePercentage int                `json:"upvotePercentage" bson:"upvotePercentage"`
	ID               string             `json:"id" bson:"_id"`
}

//go:generate mockgen -source=post.go -destination=repo_mock.go -package=post PostsRepo

type PostsRepo interface {
	GetAll() ([]*Post, error)
	AddPost(author user.User, reqPost Post, newPostID string, timeCreated time.Time) *Post
	GetPost(id string, post **Post) error
	GetCategory(category string) ([]*Post, error)
	AddComment(id string, newComment string, timeCreated time.Time, author user.User, newCimmentID string, post **Post) error
	DeleteComment(postID string, commentID string, post **Post) error
	UpvotePost(postID string, author user.User, post **Post) error
	DownvotePost(postID string, author user.User, post **Post) error
	UnvotePost(postID string, author user.User, post **Post) error
	DeletePost(postID string) error
	GetUserPosts(username string) ([]*Post, error)
}
