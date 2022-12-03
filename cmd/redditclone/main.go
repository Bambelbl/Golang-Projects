package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"redditclone/pkg/handlers"
	"redditclone/pkg/middleware"
	"redditclone/pkg/post"
	"redditclone/pkg/post/mongoapi"
	"redditclone/pkg/session"
	"redditclone/pkg/user"
)

func main() {

	dsn := "root:love@tcp(localhost:3306)/golang?"
	dsn += "charset=utf8"
	dsn += "&interpolateParams=true&parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	db.SetMaxOpenConns(10)
	err = db.Ping()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongoapi.Connect(context.TODO(), clientOptions)
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	collPostRepo := client.Database("golang").Collection("posts")

	tmp, err := template.ParseGlob("../../static/html/*")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	sessionRepo := session.NewMySQLRepo(db)
	userRepo := user.NewMySQLRepo(db)
	postRepo := post.NewMongoRepo(collPostRepo)
	templates := template.Must(tmp, err)
	zapLogger, err := zap.NewProduction()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer func(zapLogger *zap.Logger) {
		err = zapLogger.Sync()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}(zapLogger)
	logger := zapLogger.Sugar()

	userHandler := &handlers.UserHandler{
		Tmpl:        templates,
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
		Logger:      logger,
	}

	postHandler := &handlers.PostsHandler{
		Tmpl:        templates,
		PostsRepo:   postRepo,
		SessionRepo: sessionRepo,
		Logger:      logger,
	}

	api := mux.NewRouter()
	api.HandleFunc("/register", userHandler.Register).Methods("POST")
	api.HandleFunc("/login", userHandler.Login).Methods("POST")
	api.HandleFunc("/posts/", postHandler.AllPosts).Methods("GET")
	api.Handle("/posts",
		middleware.CheckAuth(sessionRepo, http.HandlerFunc(postHandler.CreatePost))).Methods("POST")
	api.HandleFunc("/post/{postID:[A-Za-z0-9]+}", postHandler.GetPost).Methods("GET")
	api.HandleFunc("/posts/{category:[A-Za-z]+}", postHandler.GetCategory).Methods("GET")
	api.Handle("/post/{postID:[A-Za-z0-9]+}",
		middleware.CheckAuth(sessionRepo, http.HandlerFunc(postHandler.CreateComment))).Methods("POST")
	api.Handle("/post/{postID:[A-Za-z0-9]+}/{commentID:[A-Za-z0-9]+}",
		middleware.CheckAuth(sessionRepo, http.HandlerFunc(postHandler.DeleteComment))).Methods("DELETE")
	api.Handle("/post/{postID:[A-Za-z0-9]+}/upvote",
		middleware.CheckAuth(sessionRepo, http.HandlerFunc(postHandler.Upvote))).Methods("GET")
	api.Handle("/post/{postID:[A-Za-z0-9]+}/downvote",
		middleware.CheckAuth(sessionRepo, http.HandlerFunc(postHandler.Downvote))).Methods("GET")
	api.Handle("/post/{postID:[A-Za-z0-9]+}/unvote",
		middleware.CheckAuth(sessionRepo, http.HandlerFunc(postHandler.Unvote))).Methods("GET")
	api.Handle("/post/{postID:[A-Za-z0-9]+}",
		middleware.CheckAuth(sessionRepo, http.HandlerFunc(postHandler.DeletePost))).Methods("DELETE")
	api.HandleFunc("/user/{username:[A-Za-z0-9_]+}", postHandler.GetUserPosts).Methods("GET")

	r := mux.NewRouter()
	r.PathPrefix("/api/").Handler(http.StripPrefix("/api", api))
	r.NotFoundHandler = http.HandlerFunc(userHandler.Index)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		http.FileServer(http.Dir("../../static/"))))

	router := middleware.AccessLog(logger, r)
	router = middleware.Panic(router)
	err = http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println(err.Error())
	}
}
