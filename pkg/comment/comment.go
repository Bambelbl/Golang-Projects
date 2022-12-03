package comment

import (
	"redditclone/pkg/user"
	"time"
)

type Comment struct {
	ID      string    `json:"id" bson:"_id"`
	Author  user.User `json:"author" bson:"author"`
	Body    string    `json:"body" bson:"body"`
	Created time.Time `json:"created" bson:"created"`
}
