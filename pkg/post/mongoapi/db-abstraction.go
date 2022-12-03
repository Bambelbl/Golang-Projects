package mongoapi

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type DatabaseAPI interface {
	Collection(name string, opts ...*options.CollectionOptions) CollectionAPI
}

type ClientAPI interface {
	Database(name string, opts ...*options.DatabaseOptions) DatabaseAPI
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context, rp *readpref.ReadPref) error
}

type CollectionAPI interface {
	FindOne(ctx context.Context, filter interface{},
		opts ...*options.FindOneOptions) SingleResultAPI
	ReplaceOne(ctx context.Context, filter interface{},
		replacement interface{}, opts ...*options.ReplaceOptions) (UpdateResultAPI, error)
	InsertOne(ctx context.Context, document interface{},
		opts ...*options.InsertOneOptions) (InsertOneResultAPI, error)
	DeleteOne(ctx context.Context, filter interface{},
		opts ...*options.DeleteOptions) (DeleteResultAPI, error)
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (CursorAPI, error)
}

type SingleResultAPI interface {
	Decode(v interface{}) error
}

type CursorAPI interface {
	Decode(v interface{}) error
	Next(ctx context.Context) bool
	Close(ctx context.Context) error
	Err() error
}

type UpdateResultAPI interface{}

type InsertOneResultAPI interface{}

type DeleteResultAPI interface{}

type mongoClient struct {
	cl *mongo.Client
}
type mongoDatabase struct {
	db *mongo.Database
}

type mongoCollection struct {
	coll *mongo.Collection
}

type mongoSingleResult struct {
	sr *mongo.SingleResult
}

type mongoCursor struct {
	c *mongo.Cursor
}

type mongoUpdateResult struct {
	u *mongo.UpdateResult
}

type mongoInsertOne struct {
	i *mongo.InsertOneResult
}

type mongoDeleteResult struct {
	d *mongo.DeleteResult
}

func Connect(ctx context.Context, opts ...*options.ClientOptions) (ClientAPI, error) {
	c, err := mongo.Connect(ctx, opts...)
	return &mongoClient{cl: c}, err
}

func (mc *mongoClient) Ping(ctx context.Context, rp *readpref.ReadPref) error {
	return mc.cl.Ping(ctx, rp)
}

func (mc *mongoClient) Disconnect(ctx context.Context) error {
	return mc.cl.Disconnect(ctx)
}

func (mc *mongoClient) Database(name string, opts ...*options.DatabaseOptions) DatabaseAPI {
	return &mongoDatabase{mc.cl.Database(name, opts...)}
}

func (md *mongoDatabase) Collection(name string, opts ...*options.CollectionOptions) CollectionAPI {
	collection := md.db.Collection(name, opts...)
	return &mongoCollection{coll: collection}
}

func (m mongoCursor) Decode(v interface{}) error {
	return m.c.Decode(v)
}

func (m mongoCursor) Next(ctx context.Context) bool {
	return m.c.Next(ctx)
}

func (m mongoCursor) Close(ctx context.Context) error {
	return m.c.Close(ctx)
}

func (m mongoCursor) Err() error {
	return m.c.Err()
}

func (mc *mongoCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (CursorAPI, error) {
	cur, err := mc.coll.Find(ctx, filter, opts...)
	return &mongoCursor{c: cur}, err
}

func (mc *mongoCollection) FindOne(ctx context.Context, filter interface{},
	opts ...*options.FindOneOptions) SingleResultAPI {
	singleResult := mc.coll.FindOne(ctx, filter, opts...)
	return &mongoSingleResult{sr: singleResult}
}

func (mc *mongoCollection) ReplaceOne(ctx context.Context, filter interface{},
	replacement interface{}, opts ...*options.ReplaceOptions) (UpdateResultAPI, error) {
	upd, err := mc.coll.ReplaceOne(ctx, filter, replacement, opts...)
	return &mongoUpdateResult{u: upd}, err
}

func (mc *mongoCollection) InsertOne(ctx context.Context, document interface{},
	opts ...*options.InsertOneOptions) (InsertOneResultAPI, error) {
	i, err := mc.coll.InsertOne(ctx, document, opts...)
	return &mongoInsertOne{i: i}, err
}

func (mc *mongoCollection) DeleteOne(ctx context.Context, filter interface{},
	opts ...*options.DeleteOptions) (DeleteResultAPI, error) {
	del, err := mc.coll.DeleteOne(ctx, filter, opts...)
	return &mongoDeleteResult{d: del}, err
}

func (sr *mongoSingleResult) Decode(v interface{}) error {
	return sr.sr.Decode(v)
}
