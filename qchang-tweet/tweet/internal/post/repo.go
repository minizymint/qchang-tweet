package post

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type repo struct {
	conn *pgx.Conn
}

func NewRepository(conn *pgx.Conn) *repo {
	return &repo{conn: conn}
}

func (r *repo) CreatePost(ctx context.Context, post *Post) error {
	_, err := r.conn.Exec(ctx, "INSERT INTO posts (id, owner_id, content, likes, created_at) VALUES ($1, $2, $3, $4, $5)",
		post.ID, post.OwnerID, post.Content, post.Likes, post.CreatedAt)

	return err
}

func (r *repo) UpdatePost(ctx context.Context, postID, userID uuid.UUID, content string) error {
	_, err := r.conn.Exec(ctx, "UPDATE posts SET content = $1, updated_at = now() WHERE id = $2 AND owner_id = $3", content, postID, userID)

	return err
}

func (r *repo) GetPost(ctx context.Context, postID, userID uuid.UUID) (*Post, error) {
	row := r.conn.QueryRow(ctx, "SELECT id, owner_id, content, likes, created_at, updated_at FROM posts WHERE id = $1 AND owner_id = $2", postID, userID)
	post := &Post{}
	err := row.Scan(&post.ID, &post.OwnerID, &post.Content, &post.Likes, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (r *repo) GetPosts(ctx context.Context, userID uuid.UUID, page int, limit int) ([]*Post, error) {
	offset := (page - 1) * limit
	rows, err := r.conn.Query(ctx, `
        SELECT id, owner_id, content, likes, created_at, updated_at
        FROM posts
        WHERE owner_id = $1
        ORDER BY created_at DESC
        OFFSET $2 LIMIT $3
    `, userID, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		post := &Post{}
		err := rows.Scan(&post.ID, &post.OwnerID, &post.Content, &post.Likes, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *repo) DeletePost(ctx context.Context, postID, userID uuid.UUID) error {
	_, err := r.conn.Exec(ctx, "DELETE FROM posts WHERE id = $1 AND owner_id = $2", postID, userID)

	return err
}

func (r *repo) GetPostsByUserIDs(ctx context.Context, userIDs []uuid.UUID) ([]*Post, error) {
	rows, err := r.conn.Query(ctx, "SELECT id, owner_id, content, likes, created_at, updated_at FROM posts WHERE owner_id = ANY($1)", userIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		post := &Post{}
		err := rows.Scan(&post.ID, &post.OwnerID, &post.Content, &post.Likes, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func (r *repo) AddLike(ctx context.Context, postID uuid.UUID) error {
	_, err := r.conn.Exec(ctx, "UPDATE posts SET likes = likes + 1 WHERE id = $1", postID)

	return err
}

func (r *repo) Exists(ctx context.Context, postID uuid.UUID) (bool, error) {
	var exists bool
	err := r.conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)", postID).Scan(&exists)

	return exists, err
}
