package post

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
	"redditclone/pkg/comment"
	"redditclone/pkg/user"
	"redditclone/pkg/vote"
	"time"

	"redditclone/pkg/post/mongoapi"
	"redditclone/pkg/post/mongoapi/mocks"
	"testing"
)

func TestGetAll(t *testing.T) {
	collectionAPI := mongoapi.CollectionAPI(&mocks.CollectionAPI{})
	curHelperCorrect := mongoapi.CursorAPI(&mocks.CursorAPI{})

	// correct empty
	collectionAPI.(*mocks.CollectionAPI).
		On("Find", context.TODO(), bson.M{}, options.Find().SetSort(bson.M{"score": -1})).
		Return(curHelperCorrect, nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(false).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Err").
		Return(nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Close", context.TODO()).
		Return(nil).Once()

	repo := NewMongoRepo(collectionAPI)
	posts, err := repo.GetAll()
	assert.Empty(t, posts)
	assert.NoError(t, err)

	// Find err
	collectionAPI.(*mocks.CollectionAPI).
		On("Find", context.TODO(), bson.M{}, options.Find().SetSort(bson.M{"score": -1})).
		Return(curHelperCorrect, ErrInternal).Once()

	posts, err = repo.GetAll()
	assert.Empty(t, posts)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}

	// Correct not empty res
	collectionAPI.(*mocks.CollectionAPI).
		On("Find", context.TODO(), bson.M{}, options.Find().SetSort(bson.M{"score": -1})).
		Return(curHelperCorrect, nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(true).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Decode", mock.AnythingOfType("*post.Post")).
		Return(func(newPost interface{}) error {
			newPost.(*Post).Score = 10
			newPost.(*Post).ID = "1"
			return nil
		}).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(true).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Decode", mock.AnythingOfType("*post.Post")).
		Return(func(newPost interface{}) error {
			newPost.(*Post).Score = 2
			newPost.(*Post).ID = "2"
			return nil
		}).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(false).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Err").
		Return(nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Close", context.TODO()).
		Return(nil).Once()

	posts, err = repo.GetAll()
	assert.Equal(t, 2, len(posts))
	assert.Equal(t, "1", posts[0].ID)
	assert.Equal(t, "2", posts[1].ID)
	assert.NoError(t, err)

	// Decode err
	collectionAPI.(*mocks.CollectionAPI).
		On("Find", context.TODO(), bson.M{}, options.Find().SetSort(bson.M{"score": -1})).
		Return(curHelperCorrect, nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(true).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Decode", &Post{}).
		Return(ErrInternal).Once()

	posts, err = repo.GetAll()
	assert.Empty(t, posts)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}

	// Err err
	collectionAPI.(*mocks.CollectionAPI).
		On("Find", context.TODO(), bson.M{}, options.Find().SetSort(bson.M{"score": -1})).
		Return(curHelperCorrect, nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(true).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Decode", &Post{}).
		Return(nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(false).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Err").
		Return(ErrInternal).Once()

	posts, err = repo.GetAll()
	assert.Empty(t, posts)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}

	// Close err
	collectionAPI.(*mocks.CollectionAPI).
		On("Find", context.TODO(), bson.M{}, options.Find().SetSort(bson.M{"score": -1})).
		Return(curHelperCorrect, nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(true).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Decode", &Post{}).
		Return(nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(false).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Err").
		Return(nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Close", context.TODO()).
		Return(ErrInternal).Once()

	posts, err = repo.GetAll()
	assert.Empty(t, posts)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}
}

func TestAddPost(t *testing.T) {
	collectionAPI := mongoapi.CollectionAPI(&mocks.CollectionAPI{})

	reqPost := Post{
		Title: "reddit",
		Type:  "text",
		Text:  "helpmepls",
		URL:   "",
	}

	author := user.User{
		Username: "mem",
		ID:       "sw234rt56",
	}
	newPostID := RandStringRunes()
	timeCreated := time.Now()
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

	// Correct
	collectionAPI.(*mocks.CollectionAPI).
		On("InsertOne", context.TODO(), newPost).
		Return(&newPost, nil).Once()

	repo := NewMongoRepo(collectionAPI)
	post := repo.AddPost(author, reqPost, newPostID, timeCreated)
	assert.NotEmpty(t, post)

	// InsertOne err
	collectionAPI.(*mocks.CollectionAPI).
		On("InsertOne", context.TODO(), newPost).
		Return(nil, ErrInternal).Once()

	post = repo.AddPost(author, reqPost, newPostID, timeCreated)
	assert.Empty(t, post)
}

func TestGetPost(t *testing.T) {
	var collectionAPI mongoapi.CollectionAPI
	var singleResultAPI mongoapi.SingleResultAPI
	collectionAPI = &mocks.CollectionAPI{}
	singleResultAPI = &mocks.SingleResultAPI{}

	postID := RandStringRunes()
	author := user.User{
		Username: "mem",
		ID:       "sw234rt56",
	}
	getPost := Post{
		Category:         "sufferings",
		Comments:         &[]comment.Comment{},
		Title:            "reddit",
		Type:             "text",
		Text:             "helpmepls",
		URL:              "",
		Author:           author,
		Created:          time.Now(),
		ID:               postID,
		Score:            1,
		UpvotePercentage: 100,
		Views:            0,
		Votes:            &[]vote.Vote{{UserID: author.ID, Vote: 1}},
	}
	postFromDB := &getPost

	// Correct
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(nil).Once()

	collectionAPI.(*mocks.CollectionAPI).
		On("ReplaceOne", context.TODO(), bson.M{"_id": postID}, &postFromDB).
		Return(nil, nil).Once()

	repo := NewMongoRepo(collectionAPI)
	err := repo.GetPost(postID, &postFromDB)
	assert.NoError(t, err)

	postFromDB = &getPost
	// ErrNoDocuments
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(mongo.ErrNoDocuments).Once()

	err = repo.GetPost(postID, &postFromDB)
	assert.Empty(t, postFromDB)
	if assert.Error(t, err) {
		assert.Equal(t, ErrNoPost, err)
	}

	postFromDB = &getPost
	// Err Internal Decode
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(errors.New("kakoy-to prikol")).Once()

	err = repo.GetPost(postID, &postFromDB)
	assert.Empty(t, postFromDB)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}

	postFromDB = &getPost
	// ReplaceOne err
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(nil).Once()

	collectionAPI.(*mocks.CollectionAPI).
		On("ReplaceOne", context.TODO(), bson.M{"_id": postID}, &postFromDB).
		Return(nil, errors.New("kakoy-to prikol")).Once()

	err = repo.GetPost(postID, &postFromDB)
	assert.Empty(t, postFromDB)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}
}

func TestGetCategory(t *testing.T) {

	var collectionAPI mongoapi.CollectionAPI
	var curHelperCorrect mongoapi.CursorAPI
	collectionAPI = &mocks.CollectionAPI{}
	curHelperCorrect = &mocks.CursorAPI{}

	// correct empty
	collectionAPI.(*mocks.CollectionAPI).
		On("Find", context.TODO(), bson.M{"category": "mem"}, options.Find().SetSort(bson.M{"score": -1})).
		Return(curHelperCorrect, nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(false).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Err").
		Return(nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Close", context.TODO()).
		Return(nil).Once()

	repo := NewMongoRepo(collectionAPI)
	posts, err := repo.GetCategory("mem")
	assert.Empty(t, posts)
	assert.NoError(t, err)

	// Find err
	collectionAPI.(*mocks.CollectionAPI).
		On("Find", context.TODO(), bson.M{"category": "mem"}, options.Find().SetSort(bson.M{"score": -1})).
		Return(curHelperCorrect, ErrInternal).Once()

	posts, err = repo.GetCategory("mem")
	assert.Empty(t, posts)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}

	// Correct not empty res
	collectionAPI.(*mocks.CollectionAPI).
		On("Find", context.TODO(), bson.M{"category": "mem"}, options.Find().SetSort(bson.M{"score": -1})).
		Return(curHelperCorrect, nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(true).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Decode", mock.AnythingOfType("*post.Post")).
		Return(func(newPost interface{}) error {
			newPost.(*Post).Score = 10
			newPost.(*Post).ID = "1"
			return nil
		}).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(true).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Decode", mock.AnythingOfType("*post.Post")).
		Return(func(newPost interface{}) error {
			newPost.(*Post).Score = 2
			newPost.(*Post).ID = "2"
			return nil
		}).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(false).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Err").
		Return(nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Close", context.TODO()).
		Return(nil).Once()

	posts, err = repo.GetCategory("mem")
	assert.Equal(t, 2, len(posts))
	assert.Equal(t, "1", posts[0].ID)
	assert.Equal(t, "2", posts[1].ID)
	assert.NoError(t, err)

	// Decode err
	collectionAPI.(*mocks.CollectionAPI).
		On("Find", context.TODO(), bson.M{"category": "mem"}, options.Find().SetSort(bson.M{"score": -1})).
		Return(curHelperCorrect, nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(true).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Decode", &Post{}).
		Return(ErrInternal).Once()

	posts, err = repo.GetCategory("mem")
	assert.Empty(t, posts)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}

	// Err err
	collectionAPI.(*mocks.CollectionAPI).
		On("Find", context.TODO(), bson.M{"category": "mem"}, options.Find().SetSort(bson.M{"score": -1})).
		Return(curHelperCorrect, nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(true).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Decode", &Post{}).
		Return(nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(false).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Err").
		Return(ErrInternal).Once()

	posts, err = repo.GetCategory("mem")
	assert.Empty(t, posts)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}

	// Close err
	collectionAPI.(*mocks.CollectionAPI).
		On("Find", context.TODO(), bson.M{"category": "mem"}, options.Find().SetSort(bson.M{"score": -1})).
		Return(curHelperCorrect, nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(true).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Decode", &Post{}).
		Return(nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(false).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Err").
		Return(nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Close", context.TODO()).
		Return(ErrInternal).Once()

	posts, err = repo.GetCategory("mem")
	assert.Empty(t, posts)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}
}

func TestAddComment(t *testing.T) {
	var collectionAPI mongoapi.CollectionAPI
	var singleResultAPI mongoapi.SingleResultAPI
	collectionAPI = &mocks.CollectionAPI{}
	singleResultAPI = &mocks.SingleResultAPI{}

	postID := RandStringRunes()
	author := user.User{
		Username: "mem",
		ID:       "sw234rt56",
	}

	getPost := Post{
		Category:         "sufferings",
		Comments:         &[]comment.Comment{},
		Title:            "reddit",
		Type:             "text",
		Text:             "helpmepls",
		URL:              "",
		Author:           author,
		Created:          time.Now(),
		ID:               postID,
		Score:            1,
		UpvotePercentage: 100,
		Views:            0,
		Votes:            &[]vote.Vote{{UserID: author.ID, Vote: 1}},
	}

	timeCreated := time.Now()
	newCommentID := RandStringRunes()
	postFromDB := &getPost

	// Correct
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(nil).Once()

	collectionAPI.(*mocks.CollectionAPI).
		On("ReplaceOne", context.TODO(), bson.M{"_id": postID}, &postFromDB).
		Return(nil, nil).Once()

	repo := NewMongoRepo(collectionAPI)
	err := repo.AddComment(postID, "mem", timeCreated, author, newCommentID, &postFromDB)
	assert.NoError(t, err)

	postFromDB = &getPost
	// ErrNoDocuments
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(mongo.ErrNoDocuments).Once()

	err = repo.AddComment(postID, "mem", timeCreated, author, newCommentID, &postFromDB)
	assert.Empty(t, postFromDB)
	if assert.Error(t, err) {
		assert.Equal(t, ErrNoPost, err)
	}

	postFromDB = &getPost
	// Err Internal Decode
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(errors.New("kakoy-to prikol")).Once()

	err = repo.AddComment(postID, "mem", timeCreated, author, newCommentID, &postFromDB)
	assert.Empty(t, postFromDB)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}

	postFromDB = &getPost
	// ReplaceOne err
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(nil).Once()

	collectionAPI.(*mocks.CollectionAPI).
		On("ReplaceOne", context.TODO(), bson.M{"_id": postID}, &postFromDB).
		Return(nil, errors.New("kakoy-to prikol")).Once()

	err = repo.AddComment(postID, "mem", timeCreated, author, newCommentID, &postFromDB)
	assert.Empty(t, postFromDB)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}
}

func TestDeleteComment(t *testing.T) {
	var collectionAPI mongoapi.CollectionAPI
	var singleResultAPI mongoapi.SingleResultAPI
	collectionAPI = &mocks.CollectionAPI{}
	singleResultAPI = &mocks.SingleResultAPI{}

	postID := RandStringRunes()
	author := user.User{
		Username: "mem",
		ID:       "sw234rt56",
	}

	getPost := Post{
		Category: "sufferings",
		Comments: &[]comment.Comment{
			{ID: "1",
				Author:  author,
				Body:    "wefgb",
				Created: time.Now()}},
		Title:            "reddit",
		Type:             "text",
		Text:             "helpmepls",
		URL:              "",
		Author:           author,
		Created:          time.Now(),
		ID:               postID,
		Score:            1,
		UpvotePercentage: 100,
		Views:            0,
		Votes:            &[]vote.Vote{{UserID: author.ID, Vote: 1}},
	}

	postFromDB := &getPost

	// Correct
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(nil).Once()

	collectionAPI.(*mocks.CollectionAPI).
		On("ReplaceOne", context.TODO(), bson.M{"_id": postID}, &postFromDB).
		Return(nil, nil).Once()

	repo := NewMongoRepo(collectionAPI)
	err := repo.DeleteComment(postID, "1", &postFromDB)
	assert.NoError(t, err)

	postFromDB = &getPost

	// ErrNoComment
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(nil).Once()

	err = repo.DeleteComment(postID, "2", &postFromDB)
	assert.Empty(t, postFromDB)
	if assert.Error(t, err) {
		assert.Equal(t, ErrNoComment, err)
	}

	postFromDB = &getPost

	// ErrNoDocuments
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(mongo.ErrNoDocuments).Once()

	err = repo.DeleteComment(postID, "1", &postFromDB)
	assert.Empty(t, postFromDB)
	if assert.Error(t, err) {
		assert.Equal(t, ErrNoPost, err)
	}

	postFromDB = &getPost
	// Err Internal Decode
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(errors.New("kakoy-to prikol")).Once()

	err = repo.DeleteComment(postID, "1", &postFromDB)
	assert.Empty(t, postFromDB)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}

	postFromDB = &Post{
		Category: "sufferings",
		Comments: &[]comment.Comment{
			{ID: "1",
				Author:  author,
				Body:    "wefgb",
				Created: time.Now()}},
		Title:            "reddit",
		Type:             "text",
		Text:             "helpmepls",
		URL:              "",
		Author:           author,
		Created:          time.Now(),
		ID:               postID,
		Score:            1,
		UpvotePercentage: 100,
		Views:            0,
		Votes:            &[]vote.Vote{{UserID: author.ID, Vote: 1}},
	}

	// ReplaceOne err
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(nil).Once()

	collectionAPI.(*mocks.CollectionAPI).
		On("ReplaceOne", context.TODO(), bson.M{"_id": postID}, &postFromDB).
		Return(nil, errors.New("kakoy-to prikol")).Once()

	err = repo.DeleteComment(postID, "1", &postFromDB)
	assert.Empty(t, postFromDB)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}
}

func TestUpvotePost(t *testing.T) {
	var collectionAPI mongoapi.CollectionAPI
	var singleResultAPI mongoapi.SingleResultAPI
	collectionAPI = &mocks.CollectionAPI{}
	singleResultAPI = &mocks.SingleResultAPI{}

	postID := RandStringRunes()
	author := user.User{
		Username: "mem",
		ID:       "sw234rt56",
	}

	getPost := Post{
		Category:         "sufferings",
		Comments:         &[]comment.Comment{},
		Title:            "reddit",
		Type:             "text",
		Text:             "helpmepls",
		URL:              "",
		Author:           author,
		Created:          time.Now(),
		ID:               postID,
		Score:            1,
		UpvotePercentage: 100,
		Views:            0,
		Votes:            &[]vote.Vote{{UserID: author.ID, Vote: 1}},
	}
	postFromDB := &getPost

	// Correct
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(nil).Once()

	collectionAPI.(*mocks.CollectionAPI).
		On("ReplaceOne", context.TODO(), bson.M{"_id": postID}, &postFromDB).
		Return(nil, nil).Once()

	repo := NewMongoRepo(collectionAPI)
	err := repo.UpvotePost(postID, author, &postFromDB)
	assert.NoError(t, err)

	postFromDB = &getPost
	// ErrNoDocuments
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(mongo.ErrNoDocuments).Once()

	err = repo.UpvotePost(postID, author, &postFromDB)
	assert.Empty(t, postFromDB)
	if assert.Error(t, err) {
		assert.Equal(t, ErrNoPost, err)
	}

	postFromDB = &getPost
	// Err Internal Decode
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(errors.New("kakoy-to prikol")).Once()

	err = repo.UpvotePost(postID, author, &postFromDB)
	assert.Empty(t, postFromDB)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}

	postFromDB = &getPost
	// ReplaceOne err
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(nil).Once()

	collectionAPI.(*mocks.CollectionAPI).
		On("ReplaceOne", context.TODO(), bson.M{"_id": postID}, &postFromDB).
		Return(nil, errors.New("kakoy-to prikol")).Once()

	err = repo.UpvotePost(postID, author, &postFromDB)
	assert.Empty(t, postFromDB)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}

	postFromDB = &getPost
	// No Vote
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(nil).Once()

	collectionAPI.(*mocks.CollectionAPI).
		On("ReplaceOne", context.TODO(), bson.M{"_id": postID}, &postFromDB).
		Return(nil, nil).Once()

	err = repo.UpvotePost(postID, user.User{ID: "5"}, &postFromDB)
	assert.NoError(t, err)
}

func TestDownvotePost(t *testing.T) {
	var collectionAPI mongoapi.CollectionAPI
	var singleResultAPI mongoapi.SingleResultAPI
	collectionAPI = &mocks.CollectionAPI{}
	singleResultAPI = &mocks.SingleResultAPI{}

	postID := RandStringRunes()
	author := user.User{
		Username: "mem",
		ID:       "sw234rt56",
	}

	getPost := Post{
		Category:         "sufferings",
		Comments:         &[]comment.Comment{},
		Title:            "reddit",
		Type:             "text",
		Text:             "helpmepls",
		URL:              "",
		Author:           author,
		Created:          time.Now(),
		ID:               postID,
		Score:            1,
		UpvotePercentage: 100,
		Views:            0,
		Votes: &[]vote.Vote{{UserID: author.ID, Vote: 1},
			{UserID: "2", Vote: 1}},
	}
	postFromDB := &getPost

	// Correct
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(nil).Once()

	collectionAPI.(*mocks.CollectionAPI).
		On("ReplaceOne", context.TODO(), bson.M{"_id": postID}, &postFromDB).
		Return(nil, nil).Once()

	repo := NewMongoRepo(collectionAPI)
	err := repo.DownvotePost(postID, author, &postFromDB)
	assert.NoError(t, err)

	postFromDB = &getPost
	// ErrNoDocuments
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(mongo.ErrNoDocuments).Once()

	err = repo.DownvotePost(postID, author, &postFromDB)
	assert.Empty(t, postFromDB)
	if assert.Error(t, err) {
		assert.Equal(t, ErrNoPost, err)
	}

	postFromDB = &getPost
	// Err Internal Decode
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(errors.New("kakoy-to prikol")).Once()

	err = repo.DownvotePost(postID, author, &postFromDB)
	assert.Empty(t, postFromDB)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}

	postFromDB = &getPost
	// ReplaceOne err
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(nil).Once()

	collectionAPI.(*mocks.CollectionAPI).
		On("ReplaceOne", context.TODO(), bson.M{"_id": postID}, &postFromDB).
		Return(nil, errors.New("kakoy-to prikol")).Once()

	err = repo.DownvotePost(postID, author, &postFromDB)
	assert.Empty(t, postFromDB)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}

	postFromDB = &getPost
	// No Vote
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(nil).Once()

	collectionAPI.(*mocks.CollectionAPI).
		On("ReplaceOne", context.TODO(), bson.M{"_id": postID}, &postFromDB).
		Return(nil, nil).Once()

	err = repo.DownvotePost(postID, user.User{ID: "5"}, &postFromDB)
	assert.NoError(t, err)
}

func TestUnvotePost(t *testing.T) {
	var collectionAPI mongoapi.CollectionAPI
	var singleResultAPI mongoapi.SingleResultAPI
	collectionAPI = &mocks.CollectionAPI{}
	singleResultAPI = &mocks.SingleResultAPI{}

	postID := RandStringRunes()
	author := user.User{
		Username: "mem",
		ID:       "sw234rt56",
	}

	getPost := Post{
		Category:         "sufferings",
		Comments:         &[]comment.Comment{},
		Title:            "reddit",
		Type:             "text",
		Text:             "helpmepls",
		URL:              "",
		Author:           author,
		Created:          time.Now(),
		ID:               postID,
		Score:            1,
		UpvotePercentage: 100,
		Views:            0,
		Votes:            &[]vote.Vote{{UserID: author.ID, Vote: 1}},
	}
	postFromDB := &getPost

	// Correct score --
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(nil).Once()

	collectionAPI.(*mocks.CollectionAPI).
		On("ReplaceOne", context.TODO(), bson.M{"_id": postID}, &postFromDB).
		Return(nil, nil).Once()

	repo := NewMongoRepo(collectionAPI)
	err := repo.UnvotePost(postID, author, &postFromDB)
	assert.NoError(t, err)

	getPost.Votes = &[]vote.Vote{{UserID: author.ID, Vote: -1}}
	postFromDB = &getPost
	// Correct score++
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(nil).Once()

	collectionAPI.(*mocks.CollectionAPI).
		On("ReplaceOne", context.TODO(), bson.M{"_id": postID}, &postFromDB).
		Return(nil, nil).Once()

	err = repo.UnvotePost(postID, author, &postFromDB)
	assert.NoError(t, err)

	postFromDB = &getPost
	// ErrNoDocuments
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(mongo.ErrNoDocuments).Once()

	err = repo.UnvotePost(postID, author, &postFromDB)
	assert.Empty(t, postFromDB)
	if assert.Error(t, err) {
		assert.Equal(t, ErrNoPost, err)
	}

	postFromDB = &getPost
	// Err Internal Decode
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(errors.New("kakoy-to prikol")).Once()

	err = repo.UnvotePost(postID, author, &postFromDB)
	assert.Empty(t, postFromDB)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}

	getPost.Votes = &[]vote.Vote{{UserID: author.ID, Vote: 1}}
	postFromDB = &getPost
	// ReplaceOne err
	collectionAPI.(*mocks.CollectionAPI).
		On("FindOne", context.TODO(), bson.M{"_id": postID}).
		Return(singleResultAPI, nil).Once()

	singleResultAPI.(*mocks.SingleResultAPI).
		On("Decode", &postFromDB).
		Return(nil).Once()

	collectionAPI.(*mocks.CollectionAPI).
		On("ReplaceOne", context.TODO(), bson.M{"_id": postID}, &postFromDB).
		Return(nil, errors.New("kakoy-to prikol")).Once()

	err = repo.UnvotePost(postID, author, &postFromDB)
	assert.Empty(t, postFromDB)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}
}

func TestDeletePost(t *testing.T) {
	collectionAPI := mongoapi.CollectionAPI(&mocks.CollectionAPI{})

	postID := RandStringRunes()

	// Correct
	collectionAPI.(*mocks.CollectionAPI).
		On("DeleteOne", context.TODO(), bson.M{"_id": postID}).
		Return(nil, nil).Once()

	repo := NewMongoRepo(collectionAPI)
	err := repo.DeletePost(postID)
	assert.NoError(t, err)

	// DeleteOne err
	collectionAPI.(*mocks.CollectionAPI).
		On("DeleteOne", context.TODO(), bson.M{"_id": postID}).
		Return(nil, errors.New("kakoy-to prikol")).Once()

	repo = NewMongoRepo(collectionAPI)
	err = repo.DeletePost(postID)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}
}

func TestGetUserPosts(t *testing.T) {
	var collectionAPI mongoapi.CollectionAPI
	var curHelperCorrect mongoapi.CursorAPI

	collectionAPI = &mocks.CollectionAPI{}
	curHelperCorrect = &mocks.CursorAPI{}

	username := "mem"
	// correct empty
	collectionAPI.(*mocks.CollectionAPI).
		On("Find", context.TODO(), bson.M{"author.username": username}).
		Return(curHelperCorrect, nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(false).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Err").
		Return(nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Close", context.TODO()).
		Return(nil).Once()

	repo := NewMongoRepo(collectionAPI)
	posts, err := repo.GetUserPosts(username)
	assert.Empty(t, posts)
	assert.NoError(t, err)

	// Find err
	collectionAPI.(*mocks.CollectionAPI).
		On("Find", context.TODO(), bson.M{"author.username": username}).
		Return(curHelperCorrect, ErrInternal).Once()

	posts, err = repo.GetUserPosts(username)
	assert.Empty(t, posts)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}

	// Correct not empty res
	collectionAPI.(*mocks.CollectionAPI).
		On("Find", context.TODO(), bson.M{"author.username": username}).
		Return(curHelperCorrect, nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(true).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Decode", mock.AnythingOfType("*post.Post")).
		Return(func(newPost interface{}) error {
			newPost.(*Post).Created = time.Now().Add(60 * 24 * time.Hour)
			newPost.(*Post).ID = "1"
			return nil
		}).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(true).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Decode", mock.AnythingOfType("*post.Post")).
		Return(func(newPost interface{}) error {
			newPost.(*Post).Created = time.Now().Add(15 * 24 * time.Hour)
			newPost.(*Post).ID = "3"
			return nil
		}).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(true).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Decode", mock.AnythingOfType("*post.Post")).
		Return(func(newPost interface{}) error {
			newPost.(*Post).Created = time.Now().Add(30 * 24 * time.Hour)
			newPost.(*Post).ID = "2"
			return nil
		}).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(false).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Err").
		Return(nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Close", context.TODO()).
		Return(nil).Once()

	posts, err = repo.GetUserPosts(username)
	assert.Equal(t, 3, len(posts))
	assert.Equal(t, "1", posts[0].ID)
	assert.Equal(t, "2", posts[1].ID)
	assert.Equal(t, "3", posts[2].ID)
	assert.NoError(t, err)

	// Decode err
	collectionAPI.(*mocks.CollectionAPI).
		On("Find", context.TODO(), bson.M{"author.username": username}).
		Return(curHelperCorrect, nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(true).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Decode", &Post{}).
		Return(ErrInternal).Once()

	posts, err = repo.GetUserPosts(username)
	assert.Empty(t, posts)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}

	// Err err
	collectionAPI.(*mocks.CollectionAPI).
		On("Find", context.TODO(), bson.M{"author.username": username}).
		Return(curHelperCorrect, nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(true).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Decode", &Post{}).
		Return(nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(false).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Err").
		Return(ErrInternal).Once()

	posts, err = repo.GetUserPosts(username)
	assert.Empty(t, posts)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}

	// Close err
	collectionAPI.(*mocks.CollectionAPI).
		On("Find", context.TODO(), bson.M{"author.username": username}).
		Return(curHelperCorrect, nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(true).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Decode", &Post{}).
		Return(nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Next", context.TODO()).
		Return(false).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Err").
		Return(nil).Once()

	curHelperCorrect.(*mocks.CursorAPI).
		On("Close", context.TODO()).
		Return(ErrInternal).Once()

	posts, err = repo.GetUserPosts(username)
	assert.Empty(t, posts)
	if assert.Error(t, err) {
		assert.Equal(t, ErrInternal, err)
	}
}
