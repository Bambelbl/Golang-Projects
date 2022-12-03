package post

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/exp/slices"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"redditclone/pkg/comment"
	"redditclone/pkg/post/mongoapi"
	"redditclone/pkg/user"
	"redditclone/pkg/vote"
	"time"
)

var (
	ErrNoPost    = errors.New("no post found")
	ErrNoComment = errors.New("no comment found")
	ErrInternal  = errors.New("internal error")
)

type PostsMongoRepository struct {
	Col mongoapi.CollectionAPI
}

func NewMongoRepo(col mongoapi.CollectionAPI) PostsRepo {
	return &PostsMongoRepository{Col: col}
}

func (repo *PostsMongoRepository) GetAll() ([]*Post, error) {
	var res = make([]*Post, 0)
	cur, err := repo.Col.Find(context.TODO(), bson.M{}, options.Find().SetSort(bson.M{"score": -1}))
	if err != nil {
		return nil, ErrInternal
	}
	for cur.Next(context.TODO()) {
		var post Post
		err = cur.Decode(&post)
		if err != nil {
			return nil, ErrInternal
		}
		res = append(res, &post)
	}
	if err = cur.Err(); err != nil {
		return nil, ErrInternal
	}
	err = cur.Close(context.TODO())
	if err != nil {
		return nil, ErrInternal
	}
	return res, nil
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

func (repo *PostsMongoRepository) AddPost(author user.User, reqPost Post,
	newPostID string, timeCreated time.Time) *Post {
	newPost := Post{

		Category:         reqPost.Category,
		Comments:         &[]comment.Comment{},
		Title:            reqPost.Title,
		Type:             reqPost.Type,
		URL:              reqPost.URL,
		Text:             reqPost.Text,
		Author:           author,
		Created:          timeCreated,
		ID:               newPostID,
		Score:            1,
		UpvotePercentage: 100,
		Views:            0,
		Votes:            &[]vote.Vote{{UserID: author.ID, Vote: 1}},
	}
	_, err := repo.Col.InsertOne(context.TODO(), newPost)
	if err != nil {
		return nil
	}
	return &newPost
}

func (repo *PostsMongoRepository) GetPost(postID string, post **Post) error {
	err := repo.Col.FindOne(context.TODO(), bson.M{"_id": postID}).Decode(post)
	if err == mongo.ErrNoDocuments {
		*post = nil
		return ErrNoPost
	} else if err != nil {
		*post = nil
		return ErrInternal
	}
	(*post).Views++
	_, err = repo.Col.ReplaceOne(context.TODO(), bson.M{"_id": postID}, post)
	if err != nil {
		*post = nil
		return ErrInternal
	}
	return nil
}

func (repo *PostsMongoRepository) GetCategory(category string) ([]*Post, error) {

	var res = make([]*Post, 0)
	cur, err := repo.Col.Find(context.TODO(), bson.M{"category": category}, options.Find().SetSort(bson.M{"score": -1}))
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		var post Post
		err = cur.Decode(&post)
		if err != nil {
			return nil, err
		}
		res = append(res, &post)
	}
	if err = cur.Err(); err != nil {
		return nil, err
	}

	err = cur.Close(context.TODO())
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (repo *PostsMongoRepository) AddComment(postID string, newComment string,
	timeCreated time.Time, author user.User, newCommentID string, post **Post) error {
	err := repo.Col.FindOne(context.TODO(), bson.M{"_id": postID}).Decode(post)
	if err == mongo.ErrNoDocuments {
		*post = nil
		return ErrNoPost
	} else if err != nil {
		*post = nil
		return ErrInternal
	}
	*(*post).Comments = append(*(*post).Comments, comment.Comment{
		Author:  author,
		Body:    newComment,
		Created: timeCreated,
		ID:      newCommentID,
	})
	_, err = repo.Col.ReplaceOne(context.TODO(), bson.M{"_id": postID}, post)
	if err != nil {
		*post = nil
		return ErrInternal
	}
	return nil
}

func (repo *PostsMongoRepository) DeleteComment(postID string, commentID string, post **Post) error {
	err := repo.Col.FindOne(context.TODO(), bson.M{"_id": postID}).Decode(post)
	if err == mongo.ErrNoDocuments {
		*post = nil
		return ErrNoPost
	} else if err != nil {
		*post = nil
		return ErrInternal
	}
	idxComment := slices.IndexFunc(*(*post).Comments, func(comment comment.Comment) bool {
		return comment.ID == commentID
	})
	if idxComment != -1 {
		*(*post).Comments = slices.Delete(*(*post).Comments, idxComment, idxComment+1)
	} else {
		*post = nil
		return ErrNoComment
	}
	_, err = repo.Col.ReplaceOne(context.TODO(), bson.M{"_id": postID}, post)
	if err != nil {
		*post = nil
		return ErrInternal
	}
	return nil
}

func (repo *PostsMongoRepository) UpvotePost(postID string, author user.User, post **Post) error {
	err := repo.Col.FindOne(context.TODO(), bson.M{"_id": postID}).Decode(post)
	if err == mongo.ErrNoDocuments {
		*post = nil
		return ErrNoPost
	} else if err != nil {
		*post = nil
		return ErrInternal
	}
	idxVote := slices.IndexFunc(*((*post).Votes), func(vote vote.Vote) bool {
		return vote.UserID == author.ID
	})
	if idxVote == -1 {
		*(*post).Votes = append(*(*post).Votes, vote.Vote{
			UserID: author.ID,
			Vote:   1,
		})
		(*post).Score++
	} else {
		(*(*post).Votes)[idxVote].Vote = 1
		(*post).Score += 2
	}
	countUpvoteUser := 0
	for _, item := range *(*post).Votes {
		if item.Vote == 1 {
			countUpvoteUser++
		}
	}
	(*post).UpvotePercentage = (countUpvoteUser * 100) / len(*(*post).Votes)
	_, err = repo.Col.ReplaceOne(context.TODO(), bson.M{"_id": postID}, post)
	if err != nil {
		*post = nil
		return ErrInternal
	}
	return nil
}
func (repo *PostsMongoRepository) DownvotePost(postID string, author user.User, post **Post) error {
	err := repo.Col.FindOne(context.TODO(), bson.M{"_id": postID}).Decode(post)
	if err == mongo.ErrNoDocuments {
		*post = nil
		return ErrNoPost
	} else if err != nil {
		*post = nil
		return ErrInternal
	}
	idxVote := slices.IndexFunc(*(*post).Votes, func(vote vote.Vote) bool {
		return vote.UserID == author.ID
	})
	if idxVote == -1 {
		*(*post).Votes = append(*(*post).Votes, vote.Vote{
			UserID: author.ID,
			Vote:   -1,
		})
		(*post).Score--
	} else {
		(*(*post).Votes)[idxVote].Vote = -1
		(*post).Score -= 2
	}
	countUpvoteUser := 0
	for _, item := range *(*post).Votes {
		if item.Vote == 1 {
			countUpvoteUser++
		}
	}
	(*post).UpvotePercentage = (countUpvoteUser * 100) / len(*(*post).Votes)
	_, err = repo.Col.ReplaceOne(context.TODO(), bson.M{"_id": postID}, post)
	if err != nil {
		*post = nil
		return ErrInternal
	}
	return nil
}

func (repo *PostsMongoRepository) UnvotePost(postID string, author user.User, post **Post) error {
	err := repo.Col.FindOne(context.TODO(), bson.M{"_id": postID}).Decode(post)
	if err == mongo.ErrNoDocuments {
		*post = nil
		return ErrNoPost
	} else if err != nil {
		*post = nil
		return ErrInternal
	}
	idxVote := slices.IndexFunc(*(*post).Votes, func(vote vote.Vote) bool {
		return vote.UserID == author.ID
	})
	if (*(*post).Votes)[idxVote].Vote == 1 {
		(*post).Score--
	} else {
		(*post).Score++
	}
	countUpvoteUser := 0
	for _, item := range *(*post).Votes {
		if item.Vote == 1 {
			countUpvoteUser++
		}
	}
	(*post).UpvotePercentage = (countUpvoteUser * 100) / len(*(*post).Votes)
	*(*post).Votes = slices.Delete(*(*post).Votes, idxVote, idxVote+1)
	_, err = repo.Col.ReplaceOne(context.TODO(), bson.M{"_id": postID}, post)
	if err != nil {
		*post = nil
		return ErrInternal
	}
	return nil
}

func (repo *PostsMongoRepository) DeletePost(postID string) error {
	_, err := repo.Col.DeleteOne(context.TODO(), bson.M{"_id": postID})
	if err != nil {
		return ErrInternal
	}
	return nil
}

func (repo *PostsMongoRepository) GetUserPosts(username string) ([]*Post, error) {
	var res = make([]*Post, 0)
	cur, err := repo.Col.Find(context.TODO(), bson.M{"author.username": username})
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		var post Post
		err = cur.Decode(&post)
		if err != nil {
			return nil, err
		}
		res = append(res, &post)
	}
	if err = cur.Err(); err != nil {
		return nil, err
	}

	err = cur.Close(context.TODO())
	if err != nil {
		return nil, err
	}
	slices.SortStableFunc(res, func(lhs *Post, rhs *Post) bool {
		return lhs.Created.After(rhs.Created)
	})
	return res, nil
}
