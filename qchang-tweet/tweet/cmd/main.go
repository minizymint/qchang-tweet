package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"tweet.io/config"
	"tweet.io/internal/auth"
	"tweet.io/internal/feed"
	"tweet.io/internal/follow"
	"tweet.io/internal/post"
	"tweet.io/internal/post/comment"
	"tweet.io/internal/post/like"
	"tweet.io/internal/user"
)

const (
	wait time.Duration = time.Second * 5
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	cfg := config.New()

	conn, err := pgx.Connect(context.Background(), cfg.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	userRepo := user.NewRepository(conn)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	authService := auth.NewService(userService, cfg.AuthSecretKey)
	authHandler := auth.NewHandler(authService)

	commentRepo := comment.NewRepository(conn)
	likeRepo := like.NewRepository(conn)

	postRepo := post.NewRepository(conn)
	postService := post.NewService(postRepo, commentRepo, likeRepo)
	postHandler := post.NewHandler(postService)

	followRepo := follow.NewRepository(conn)
	followService := follow.NewService(followRepo, userService)
	followHandler := follow.NewHandler(followService)

	feedService := feed.NewService(followService, postService)
	feedHandler := feed.NewHandler(feedService)

	r := mux.NewRouter()

	r.HandleFunc("/v1/users/register", userHandler.CreateUser).Methods(http.MethodPost)
	r.HandleFunc("/v1/auth/login", authHandler.Login).Methods(http.MethodPost)

	userRouter := r.NewRoute().Subrouter()
	userRouter.Use(auth.AuthMiddleware(cfg.AuthSecretKey))
	userRouter.HandleFunc("/v1/users/profile", userHandler.GetProfile).Methods(http.MethodGet)

	postRouter := r.NewRoute().Subrouter()
	postRouter.Use(auth.AuthMiddleware(cfg.AuthSecretKey))
	postRouter.HandleFunc("/v1/posts", postHandler.CreatePost).Methods(http.MethodPost)
	postRouter.HandleFunc("/v1/posts/{post_id}", postHandler.GetPost).Methods(http.MethodGet)
	postRouter.HandleFunc("/v1/posts", postHandler.GetPosts).Methods(http.MethodGet)
	postRouter.HandleFunc("/v1/posts/{post_id}", postHandler.UpdatePost).Methods(http.MethodPatch)
	postRouter.HandleFunc("/v1/posts/{post_id}", postHandler.DeletePost).Methods(http.MethodDelete)
	postRouter.HandleFunc("/v1/posts/{post_id}/likes", postHandler.AddLike).Methods(http.MethodGet)

	commentRouter := r.NewRoute().Subrouter()
	commentRouter.Use(auth.AuthMiddleware(cfg.AuthSecretKey))
	commentRouter.HandleFunc("/v1/posts/{post_id}/comments", postHandler.CreateComment).Methods(http.MethodPost)
	commentRouter.HandleFunc("/v1/posts/{post_id}/comments/{comment_id}", postHandler.UpdateComment).Methods(http.MethodPatch)
	commentRouter.HandleFunc("/v1/posts/{post_id}/comments/{comment_id}", postHandler.DeleteComment).Methods(http.MethodDelete)
	commentRouter.HandleFunc("/v1/posts/{post_id}/comments/{comment_id}", postHandler.GetComment).Methods(http.MethodGet)
	commentRouter.HandleFunc("/v1/posts/{post_id}/comments", postHandler.GetComments).Methods(http.MethodGet)

	followRouter := r.NewRoute().Subrouter()
	followRouter.Use(auth.AuthMiddleware(cfg.AuthSecretKey))
	followRouter.HandleFunc("/v1/follows", followHandler.CreateFollow).Methods(http.MethodPost)

	feedRouter := r.NewRoute().Subrouter()
	feedRouter.Use(auth.AuthMiddleware(cfg.AuthSecretKey))
	feedRouter.HandleFunc("/v1/feed", feedHandler.GetFeed).Methods(http.MethodGet)

	srv := &http.Server{
		Addr:         "0.0.0.0:8080",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	log.Println("Server started on port 8080")

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	srv.Shutdown(ctx)

	log.Println("shutting down")
	os.Exit(0)
}
