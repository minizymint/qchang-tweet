package comment

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrNotFound = errors.New("follow not found")
)

type repo struct {
	conn *pgx.Conn
}

func NewRepository(conn *pgx.Conn) *repo {
	return &repo{conn: conn}
}

func (r *repo) Create(ctx context.Context, comment *Comment) error {
	_, err := r.conn.Exec(ctx, "INSERT INTO comments (id, post_id, owner_id, content, likes, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		comment.ID, comment.PostID, comment.OwnerID, comment.Content, 0, comment.CreatedAt)

	return err
}

func (r *repo) Update(ctx context.Context, id, postID, ownerID uuid.UUID, content string) error {
	_, err := r.conn.Exec(ctx, "UPDATE comments SET content = $1, updated_at = now() WHERE id = $2 AND post_id = $3 AND owner_id = $4", content, id, postID, ownerID)

	return err
}

func (r *repo) GetComment(ctx context.Context, id, postID, ownerID uuid.UUID) (*Comment, error) {
	row := r.conn.QueryRow(ctx, "SELECT id, post_id, owner_id, content, created_at, updated_at FROM comments WHERE id = $1 AND post_id = $2 AND owner_id = $3", id, postID, ownerID)
	comment := &Comment{}
	err := row.Scan(&comment.ID, &comment.PostID, &comment.OwnerID, &comment.Content, &comment.CreatedAt, &comment.UpdateAt)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

func (r *repo) GetComments(ctx context.Context, postID, ownerID uuid.UUID, page, limit int) ([]*Comment, error) {
	offset := (page - 1) * limit
	query := `
		SELECT id, post_id, owner_id, content, created_at, updated_at
		FROM comments
		WHERE post_id = $1 AND owner_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.conn.Query(ctx, query, postID, ownerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.OwnerID, &comment.Content, &comment.CreatedAt, &comment.UpdateAt); err != nil {
			return nil, err
		}
		comments = append(comments, &comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *repo) Delete(ctx context.Context, id, postID, ownerID uuid.UUID) error {
	_, err := r.conn.Exec(ctx, "DELETE FROM comments WHERE id = $1 AND post_id = $2 AND owner_id = $3", id, postID, ownerID)

	return err
}
