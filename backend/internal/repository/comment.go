package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/f1-rivals-cup/backend/internal/database"
	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/google/uuid"
)

var (
	ErrCommentNotFound = errors.New("comment not found")
	// ErrNewsNotFound is defined in news.go
)

// CommentRepository handles comment database operations
type CommentRepository struct {
	db *database.DB
}

// NewCommentRepository creates a new CommentRepository
func NewCommentRepository(db *database.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create creates a new comment
func (r *CommentRepository) Create(ctx context.Context, comment *model.NewsComment) error {
	query := `
		INSERT INTO news_comments (news_id, author_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := r.db.Pool.QueryRowContext(ctx, query,
		comment.NewsID,
		comment.AuthorID,
		comment.Content,
	).Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)

	if err != nil {
		// Check for foreign key violation (news doesn't exist)
		if err.Error() == `pq: insert or update on table "news_comments" violates foreign key constraint "news_comments_news_id_fkey"` {
			return ErrNewsNotFound
		}
		return err
	}

	return nil
}

// GetByID retrieves a comment by ID
func (r *CommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.NewsComment, error) {
	query := `
		SELECT nc.id, nc.news_id, nc.author_id, nc.content, nc.created_at, nc.updated_at, u.nickname
		FROM news_comments nc
		JOIN users u ON nc.author_id = u.id
		WHERE nc.id = $1
	`

	comment := &model.NewsComment{}
	err := r.db.Pool.QueryRowContext(ctx, query, id).Scan(
		&comment.ID,
		&comment.NewsID,
		&comment.AuthorID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
		&comment.AuthorNickname,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}

	return comment, nil
}

// ListByNews retrieves comments for a news article with pagination
func (r *CommentRepository) ListByNews(ctx context.Context, newsID uuid.UUID, page, limit int) ([]*model.NewsComment, int, error) {
	offset := (page - 1) * limit

	// Count total comments
	countQuery := `SELECT COUNT(*) FROM news_comments WHERE news_id = $1`
	var total int
	if err := r.db.Pool.QueryRowContext(ctx, countQuery, newsID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get comments with author nickname
	query := `
		SELECT nc.id, nc.news_id, nc.author_id, nc.content, nc.created_at, nc.updated_at, u.nickname
		FROM news_comments nc
		JOIN users u ON nc.author_id = u.id
		WHERE nc.news_id = $1
		ORDER BY nc.created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, newsID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var comments []*model.NewsComment
	for rows.Next() {
		c := &model.NewsComment{}
		if err := rows.Scan(
			&c.ID,
			&c.NewsID,
			&c.AuthorID,
			&c.Content,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.AuthorNickname,
		); err != nil {
			return nil, 0, err
		}
		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// Update updates a comment's content
func (r *CommentRepository) Update(ctx context.Context, comment *model.NewsComment) error {
	query := `
		UPDATE news_comments
		SET content = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING updated_at
	`

	err := r.db.Pool.QueryRowContext(ctx, query, comment.Content, comment.ID).Scan(&comment.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCommentNotFound
		}
		return err
	}

	return nil
}

// Delete removes a comment
func (r *CommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM news_comments WHERE id = $1`

	result, err := r.db.Pool.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrCommentNotFound
	}

	return nil
}
