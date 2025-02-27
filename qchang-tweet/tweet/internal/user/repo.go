package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrNotFound = errors.New("user not found")
)

type repo struct {
	conn *pgx.Conn
}

func NewRepository(conn *pgx.Conn) *repo {
	return &repo{conn: conn}
}

func (r *repo) CreateUser(ctx context.Context, user *User) error {
	_, err := r.conn.Exec(ctx, "INSERT INTO users (id, email, firstname, lastname, displayname, hashed_password, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		user.ID, user.Email, user.Firstname, user.Lastname, user.Displayname, user.HashedPassword, user.CreatedAt)

	return err
}

func (r *repo) FindByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}
	err := r.conn.QueryRow(ctx, "SELECT id, email, hashed_password FROM users WHERE email = $1", email).Scan(&user.ID, &user.Email, &user.HashedPassword)

	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}

	return user, err
}

func (r *repo) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	user := &User{}
	err := r.conn.QueryRow(ctx, "SELECT id, email, firstname, lastname, displayname, created_at, updated_at FROM users WHERE id = $1", id).
		Scan(&user.ID, &user.Email, &user.Firstname, &user.Lastname, &user.Displayname, &user.CreatedAt, &user.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}

	return user, err
}
