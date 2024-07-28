package user

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"tweet.io/internal/platform/reqctx"
	"tweet.io/internal/platform/response"
)

type CreateUserRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Firstname   string `json:"firstname"`
	Lastname    string `json:"lastname"`
	Displayname string `json:"displayname"`
}

type CreateUserResponse struct {
	ID string `json:"id"`
}

type GetUserProfileResponse struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	Firstname   string     `json:"firstname"`
	Lastname    string     `json:"lastname"`
	Displayname string     `json:"displayname"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type service interface {
	CreateUser(ctx context.Context, params *CreateServiceParams) (*User, error)
	Authenticate(ctx context.Context, email string, password string) (*User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*User, error)
}

type handler struct {
	svc service
}

func NewHandler(svc service) *handler {
	return &handler{
		svc: svc,
	}
}

func (h *handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var params CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	user, err := h.svc.CreateUser(r.Context(), &CreateServiceParams{
		Email:       params.Email,
		Password:    params.Password,
		Firstname:   params.Firstname,
		Lastname:    params.Lastname,
		Displayname: params.Displayname,
	})

	if err != nil {
		switch err {
		case
			ErrEmailAlreadyExists,
			ErrInvalidEmail,
			ErrInvalidPassword:
			response.Error(w, http.StatusBadRequest, err)
		default:
			response.Error(w, http.StatusInternalServerError, err)
		}
		return
	}

	resp := CreateUserResponse{
		ID: user.ID.String(),
	}

	response.Success(w, http.StatusCreated, resp)
}

func (h *handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := reqctx.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusInternalServerError, errors.New("user not found"))
		return
	}

	user, err := h.svc.GetUser(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err)
		return
	}

	resp := GetUserProfileResponse{
		ID:          user.ID.String(),
		Email:       user.Email,
		Firstname:   user.Firstname,
		Lastname:    user.Lastname,
		Displayname: user.Displayname,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	response.Success(w, http.StatusOK, resp)
}
