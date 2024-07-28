package post

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"tweet.io/internal/post/comment"
	"tweet.io/internal/post/like"
)

var (
	ErrPostNotFound      = errors.New("post not found")
	ErrCommentNotFound   = errors.New("comment not found")
	ErrLikeAlreadyExists = errors.New("like already exists")
)

type CreatePostServiceParams struct {
	OwnerID uuid.UUID
	Content string
}

type UpdatePostServiceParams struct {
	OwnerID uuid.UUID
	Content string
}

type CreateCommentServiceParams struct {
	PostID  uuid.UUID
	OwnerID uuid.UUID
	Content string
}

type UpdateCommentServiceParams struct {
	ID      uuid.UUID
	PostID  uuid.UUID
	OwnerID uuid.UUID
	Content string
}

type GetCommentServiceParams struct {
	PostID  uuid.UUID
	OwnerID uuid.UUID
}

type GetCommentsServiceParams struct {
	PostID  uuid.UUID
	OwnerID uuid.UUID
}

type repository interface {
	CreatePost(ctx context.Context, post *Post) error
	UpdatePost(ctx context.Context, postID, userID uuid.UUID, content string) error
	GetPost(ctx context.Context, postID, userID uuid.UUID) (*Post, error)
	GetPosts(ctx context.Context, userID uuid.UUID, page, limit int) ([]*Post, error)
	DeletePost(ctx context.Context, postID, userID uuid.UUID) error
	GetPostsByUserIDs(ctx context.Context, userIDs []uuid.UUID) ([]*Post, error)
	AddLike(ctx context.Context, postID uuid.UUID) error
	Exists(ctx context.Context, postID uuid.UUID) (bool, error)
}

type commentRepository interface {
	Create(ctx context.Context, comment *comment.Comment) error
	Update(ctx context.Context, id, postID, ownerID uuid.UUID, content string) error
	GetComment(ctx context.Context, id, postID, ownerID uuid.UUID) (*comment.Comment, error)
	GetComments(ctx context.Context, postID, ownerID uuid.UUID, page, limit int) ([]*comment.Comment, error)
	Delete(ctx context.Context, id, postID, ownerID uuid.UUID) error
}

type likeRepository interface {
	Create(ctx context.Context, like *like.Like) error
	Exists(ctx context.Context, postID, ownerID uuid.UUID) (bool, error)
}

type Service struct {
	repo        repository
	commentRepo commentRepository
	likeRepo    likeRepository
}

func NewService(repo repository, commentRepo commentRepository, likeRepo likeRepository) *Service {
	return &Service{
		repo:        repo,
		commentRepo: commentRepo,
		likeRepo:    likeRepo,
	}
}

func (s *Service) CreatePost(ctx context.Context, params *CreatePostServiceParams) (*Post, error) {
	post, err := NewPost(params.OwnerID, params.Content)
	if err != nil {
		return nil, err
	}

	err = s.repo.CreatePost(ctx, post)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *Service) UpdatePost(ctx context.Context, postID, userID uuid.UUID, params *UpdatePostServiceParams) (*Post, error) {
	exists, err := s.repo.Exists(ctx, postID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, ErrPostNotFound
	}

	err = s.repo.UpdatePost(ctx, postID, userID, params.Content)
	if err != nil {
		return nil, err
	}

	post, err := s.repo.GetPost(ctx, postID, userID)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *Service) GetPost(ctx context.Context, postID, userID uuid.UUID) (*Post, error) {
	exists, err := s.repo.Exists(ctx, postID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, ErrPostNotFound
	}

	post, err := s.repo.GetPost(ctx, postID, userID)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *Service) GetPosts(ctx context.Context, userID uuid.UUID, page int, limit int) ([]*Post, error) {
	posts, err := s.repo.GetPosts(ctx, userID, page, limit)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (s *Service) DeletePost(ctx context.Context, postID, userID uuid.UUID) error {
	exists, err := s.repo.Exists(ctx, postID)
	if err != nil {
		return err
	}

	if !exists {
		return ErrPostNotFound
	}

	err = s.repo.DeletePost(ctx, postID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) CreateComment(ctx context.Context, params *CreateCommentServiceParams) (*comment.Comment, error) {
	exists, err := s.repo.Exists(ctx, params.PostID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, ErrPostNotFound
	}

	comment, err := comment.NewComment(params.PostID, params.OwnerID, params.Content)
	if err != nil {
		return nil, err
	}

	err = s.commentRepo.Create(ctx, comment)
	if err != nil {
		return nil, err
	}

	// Create notification message
	notification := map[string]string{
		"type":    "comment",
		"postId":  params.PostID.String(),
		"userId":  params.OwnerID.String(),
		"message": "User " + params.OwnerID.String() + " commented on your post",
	}
	notificationMessage, _ := json.Marshal(notification)

	// Publish notification
	if err := publishNotification(notificationMessage); err != nil {
		log.Printf("Failed to publish notification: %v", err)
	}

	return comment, nil
}

func (s *Service) UpdateComment(ctx context.Context, params *UpdateCommentServiceParams) (*comment.Comment, error) {
	exists, err := s.repo.Exists(ctx, params.PostID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, ErrPostNotFound
	}

	err = s.commentRepo.Update(ctx, params.ID, params.PostID, params.OwnerID, params.Content)
	if err != nil {
		return nil, err
	}

	comment, err := s.commentRepo.GetComment(ctx, params.ID, params.PostID, params.OwnerID)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *Service) GetComment(ctx context.Context, id uuid.UUID, params *GetCommentServiceParams) (*comment.Comment, error) {
	exists, err := s.repo.Exists(ctx, params.PostID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, ErrPostNotFound
	}

	comment, err := s.commentRepo.GetComment(ctx, id, params.PostID, params.OwnerID)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *Service) GetComments(ctx context.Context, page, limit int, params *GetCommentsServiceParams) ([]*comment.Comment, error) {
	exists, err := s.repo.Exists(ctx, params.PostID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, ErrPostNotFound
	}

	comments, err := s.commentRepo.GetComments(ctx, params.PostID, params.OwnerID, page, limit)
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func (s *Service) DeleteComment(ctx context.Context, id, postID, ownerID uuid.UUID) error {
	exists, err := s.repo.Exists(ctx, postID)
	if err != nil {
		return err
	}

	if !exists {
		return ErrPostNotFound
	}

	_, err = s.commentRepo.GetComment(ctx, id, postID, ownerID)
	if err != nil {
		return ErrCommentNotFound
	}

	err = s.commentRepo.Delete(ctx, id, postID, ownerID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) AddLike(ctx context.Context, postID, userID uuid.UUID) error {
	exists, err := s.repo.Exists(ctx, postID)
	if err != nil {
		return err
	}

	if !exists {
		return ErrPostNotFound
	}

	exists, err = s.likeRepo.Exists(ctx, postID, userID)
	if err != nil {
		return err
	}

	if exists {
		return ErrLikeAlreadyExists
	}

	like, err := like.NewLike(postID, userID)
	if err != nil {
		return err
	}

	// create a like record
	err = s.likeRepo.Create(ctx, like)
	if err != nil {
		return err
	}

	// increment the post's like count
	err = s.repo.AddLike(ctx, postID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetPostsByUserIDs(ctx context.Context, userIDs []uuid.UUID) ([]*Post, error) {
	posts, err := s.repo.GetPostsByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func publishNotification(message []byte) error {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"notifications",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %v", err)
	}

	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %v", err)
	}

	return nil
}
