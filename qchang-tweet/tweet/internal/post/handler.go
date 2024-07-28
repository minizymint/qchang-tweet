package post

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"tweet.io/internal/platform/reqctx"
	"tweet.io/internal/platform/response"
	"tweet.io/internal/post/comment"
)

type CreatePostHandlerParams struct {
	Content string `json:"content"`
}

type UpdatePostHandlerParams struct {
	Content string `json:"content"`
}

type CreatePostHandlerResponse struct {
	ID        uuid.UUID  `json:"id"`
	OwnerID   uuid.UUID  `json:"owner_id"`
	Content   string     `json:"content"`
	Likes     int        `json:"likes"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type UpdatePostHandlerResponse struct {
	ID        uuid.UUID  `json:"id"`
	OwnerID   uuid.UUID  `json:"owner_id"`
	Content   string     `json:"content"`
	Likes     int        `json:"likes"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type GetPostHandlerResponse struct {
	ID        uuid.UUID  `json:"id"`
	OwnerID   uuid.UUID  `json:"owner_id"`
	Content   string     `json:"content"`
	Likes     int        `json:"likes"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type GetPostsHandlerResponse struct {
	Posts []*GetPostHandlerResponse `json:"posts"`
}

type CreateCommentHandlerParams struct {
	Content string `json:"content"`
}

type UpdateCommentHandlerParams struct {
	Content string `json:"content"`
}

type CreateCommentHandlerResponse struct {
	ID        uuid.UUID  `json:"id"`
	PostID    uuid.UUID  `json:"post_id"`
	OwnerID   uuid.UUID  `json:"owner_id"`
	Content   string     `json:"content"`
	Likes     int        `json:"likes"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type UpdateCommentHandlerResponse struct {
	ID        uuid.UUID  `json:"id"`
	PostID    uuid.UUID  `json:"post_id"`
	OwnerID   uuid.UUID  `json:"owner_id"`
	Content   string     `json:"content"`
	Likes     int        `json:"likes"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type GetCommentHandlerResponse struct {
	ID        uuid.UUID  `json:"id"`
	PostID    uuid.UUID  `json:"post_id"`
	OwnerID   uuid.UUID  `json:"owner_id"`
	Content   string     `json:"content"`
	Likes     int        `json:"likes"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type GetCommentsHandlerResponse struct {
	Comments []*GetCommentHandlerResponse `json:"comments"`
}

type service interface {
	CreatePost(ctx context.Context, params *CreatePostServiceParams) (*Post, error)
	UpdatePost(ctx context.Context, postID, ownerID uuid.UUID, params *UpdatePostServiceParams) (*Post, error)
	GetPost(ctx context.Context, postID, ownerID uuid.UUID) (*Post, error)
	GetPosts(ctx context.Context, ownerID uuid.UUID, page int, limit int) ([]*Post, error)
	DeletePost(ctx context.Context, postID, ownerID uuid.UUID) error
	AddLike(ctx context.Context, postID, ownerID uuid.UUID) error
	CreateComment(ctx context.Context, params *CreateCommentServiceParams) (*comment.Comment, error)
	UpdateComment(ctx context.Context, params *UpdateCommentServiceParams) (*comment.Comment, error)
	GetComment(ctx context.Context, id uuid.UUID, params *GetCommentServiceParams) (*comment.Comment, error)
	GetComments(ctx context.Context, page, limit int, params *GetCommentsServiceParams) ([]*comment.Comment, error)
	DeleteComment(ctx context.Context, commentID, postID, ownerID uuid.UUID) error
}

type handler struct {
	service service
}

func NewHandler(service service) *handler {
	return &handler{service: service}
}

func (h *handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var params CreatePostHandlerParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	userID, ok := reqctx.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusInternalServerError, errors.New("user not found"))
		return
	}

	post, err := h.service.CreatePost(r.Context(), &CreatePostServiceParams{
		OwnerID: userID,
		Content: params.Content,
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err)
		return
	}

	resp := CreatePostHandlerResponse{
		ID:        post.ID,
		OwnerID:   post.OwnerID,
		Content:   post.Content,
		Likes:     post.Likes,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}

	response.Success(w, http.StatusOK, resp)
}

func (h *handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	var params UpdatePostHandlerParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	postID, err := uuid.Parse(mux.Vars(r)["post_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	userID, ok := reqctx.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusInternalServerError, errors.New("user not found"))
		return
	}

	post, err := h.service.UpdatePost(r.Context(), postID, userID, &UpdatePostServiceParams{
		Content: params.Content,
	})
	if err != nil {
		switch err {
		case ErrPostNotFound:
			response.Error(w, http.StatusNotFound, err)
		default:
			response.Error(w, http.StatusInternalServerError, err)
		}
		return
	}

	resp := UpdatePostHandlerResponse{
		ID:        post.ID,
		OwnerID:   post.OwnerID,
		Content:   post.Content,
		Likes:     post.Likes,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}

	response.Success(w, http.StatusOK, resp)
}

func (h *handler) GetPost(w http.ResponseWriter, r *http.Request) {
	postID, err := uuid.Parse(mux.Vars(r)["post_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	userID, ok := reqctx.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusInternalServerError, errors.New("user not found"))
		return
	}

	post, err := h.service.GetPost(r.Context(), postID, userID)
	if err != nil {
		switch err {
		case ErrPostNotFound:
			response.Error(w, http.StatusNotFound, err)
		default:
			response.Error(w, http.StatusInternalServerError, err)
		}
		return
	}

	resp := UpdatePostHandlerResponse{
		ID:        post.ID,
		OwnerID:   post.OwnerID,
		Content:   post.Content,
		Likes:     post.Likes,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}

	response.Success(w, http.StatusOK, resp)
}

func (h *handler) GetPosts(w http.ResponseWriter, r *http.Request) {
	userID, ok := reqctx.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusInternalServerError, errors.New("user not found"))
		return
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 {
			limit = l
		}
	}

	posts, err := h.service.GetPosts(r.Context(), userID, page, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err)
		return
	}

	resp := &GetPostsHandlerResponse{}
	for _, p := range posts {
		resp.Posts = append(resp.Posts, &GetPostHandlerResponse{
			ID:        p.ID,
			OwnerID:   p.OwnerID,
			Content:   p.Content,
			Likes:     p.Likes,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		})
	}

	response.Success(w, http.StatusOK, resp)
}

func (h *handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	postID, err := uuid.Parse(mux.Vars(r)["post_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	userID, ok := reqctx.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusInternalServerError, errors.New("user not found"))
		return
	}

	err = h.service.DeletePost(r.Context(), postID, userID)
	if err != nil {
		switch err {
		case ErrPostNotFound:
			response.Error(w, http.StatusNotFound, err)
		default:
			response.Error(w, http.StatusInternalServerError, err)
		}
		return
	}

	response.Success(w, http.StatusOK, "successfully deleted post")
}

func (h *handler) AddLike(w http.ResponseWriter, r *http.Request) {
	postID, err := uuid.Parse(mux.Vars(r)["post_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	userID, ok := reqctx.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusInternalServerError, errors.New("user not found"))
		return
	}

	err = h.service.AddLike(r.Context(), postID, userID)
	if err != nil {
		switch err {
		case ErrLikeAlreadyExists:
			response.Error(w, http.StatusBadRequest, err)
		case ErrPostNotFound:
			response.Error(w, http.StatusNotFound, err)
		default:
			response.Error(w, http.StatusBadRequest, err)
		}
		return
	}

	response.Success(w, http.StatusCreated, "successfully liked post")
}

func (h *handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	postID, err := uuid.Parse(mux.Vars(r)["post_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	var params CreateCommentHandlerParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	userID, ok := reqctx.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusInternalServerError, errors.New("user not found"))
		return
	}

	newComment, err := h.service.CreateComment(r.Context(), &CreateCommentServiceParams{
		PostID:  postID,
		OwnerID: userID,
		Content: params.Content,
	})
	if err != nil {
		switch err {
		case comment.ErrPostEmpty, comment.ErrEmptyContent, comment.ErrOwnerEmpty:
			response.Error(w, http.StatusBadRequest, err)
		default:
			response.Error(w, http.StatusInternalServerError, err)
		}

		return
	}

	resp := &CreateCommentHandlerResponse{
		ID:        newComment.ID,
		PostID:    newComment.PostID,
		OwnerID:   newComment.OwnerID,
		Content:   newComment.Content,
		Likes:     newComment.Likes,
		CreatedAt: newComment.CreatedAt,
	}

	response.Success(w, http.StatusOK, resp)
}

func (h *handler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	postID, err := uuid.Parse(mux.Vars(r)["post_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	commentID, err := uuid.Parse(mux.Vars(r)["comment_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	var params UpdateCommentHandlerParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	userID, ok := reqctx.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusInternalServerError, errors.New("user not found"))
		return
	}

	comment, err := h.service.UpdateComment(r.Context(), &UpdateCommentServiceParams{
		ID:      commentID,
		PostID:  postID,
		OwnerID: userID,
		Content: params.Content,
	})

	if err != nil {
		switch err {
		case ErrCommentNotFound, ErrPostNotFound:
			response.Error(w, http.StatusNotFound, err)
		default:
			response.Error(w, http.StatusInternalServerError, err)
		}
		return
	}

	resp := &UpdateCommentHandlerResponse{
		ID:        comment.ID,
		PostID:    comment.PostID,
		OwnerID:   comment.OwnerID,
		Content:   comment.Content,
		Likes:     comment.Likes,
		CreatedAt: comment.CreatedAt,
	}

	response.Success(w, http.StatusOK, resp)
}

func (h *handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	postID, err := uuid.Parse(mux.Vars(r)["post_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	commentID, err := uuid.Parse(mux.Vars(r)["comment_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	userID, ok := reqctx.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusInternalServerError, errors.New("user not found"))
		return
	}

	err = h.service.DeleteComment(r.Context(), commentID, postID, userID)
	if err != nil {
		switch err {
		case ErrCommentNotFound, ErrPostNotFound:
			response.Error(w, http.StatusNotFound, err)
		default:
			response.Error(w, http.StatusInternalServerError, err)
		}
		return
	}

	response.Success(w, http.StatusOK, "successfully deleted comment")
}

func (h *handler) GetComment(w http.ResponseWriter, r *http.Request) {
	postID, err := uuid.Parse(mux.Vars(r)["post_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	commentID, err := uuid.Parse(mux.Vars(r)["comment_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	userID, ok := reqctx.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusInternalServerError, errors.New("user not found"))
		return
	}

	comment, err := h.service.GetComment(r.Context(), commentID, &GetCommentServiceParams{
		PostID:  postID,
		OwnerID: userID,
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err)
		return
	}

	resp := &GetCommentHandlerResponse{
		ID:        comment.ID,
		PostID:    comment.PostID,
		OwnerID:   comment.OwnerID,
		Content:   comment.Content,
		Likes:     comment.Likes,
		CreatedAt: comment.CreatedAt,
	}

	response.Success(w, http.StatusOK, resp)
}

func (h *handler) GetComments(w http.ResponseWriter, r *http.Request) {
	postID, err := uuid.Parse(mux.Vars(r)["post_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	userID, ok := reqctx.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusInternalServerError, errors.New("user not found"))
		return
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	comments, err := h.service.GetComments(r.Context(), page, limit, &GetCommentsServiceParams{
		PostID:  postID,
		OwnerID: userID,
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err)
		return
	}

	resp := &GetCommentsHandlerResponse{}
	for _, p := range comments {
		resp.Comments = append(resp.Comments, &GetCommentHandlerResponse{
			ID:        p.ID,
			PostID:    p.PostID,
			OwnerID:   p.OwnerID,
			Content:   p.Content,
			Likes:     p.Likes,
			CreatedAt: p.CreatedAt,
		})
	}

	response.Success(w, http.StatusOK, resp)
}
