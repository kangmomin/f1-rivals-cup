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
	ErrNewsNotFound = errors.New("news not found")
)

// NewsRepository handles news database operations
type NewsRepository struct {
	db *database.DB
}

// NewNewsRepository creates a new NewsRepository
func NewNewsRepository(db *database.DB) *NewsRepository {
	return &NewsRepository{db: db}
}

// Create creates a new news article
func (r *NewsRepository) Create(ctx context.Context, news *model.News) error {
	query := `
		INSERT INTO news (league_id, author_id, title, content, is_published, published_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.db.Pool.QueryRowContext(ctx, query,
		news.LeagueID,
		news.AuthorID,
		news.Title,
		news.Content,
		news.IsPublished,
		news.PublishedAt,
	).Scan(&news.ID, &news.CreatedAt, &news.UpdatedAt)

	return err
}

// GetByID retrieves a news article by ID
func (r *NewsRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.News, error) {
	query := `
		SELECT n.id, n.league_id, n.author_id, n.title, n.content, n.is_published, n.published_at, n.created_at, n.updated_at, u.nickname
		FROM news n
		JOIN users u ON n.author_id = u.id
		WHERE n.id = $1
	`

	news := &model.News{}
	err := r.db.Pool.QueryRowContext(ctx, query, id).Scan(
		&news.ID,
		&news.LeagueID,
		&news.AuthorID,
		&news.Title,
		&news.Content,
		&news.IsPublished,
		&news.PublishedAt,
		&news.CreatedAt,
		&news.UpdatedAt,
		&news.AuthorNickname,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNewsNotFound
		}
		return nil, err
	}

	return news, nil
}

// ListByLeague retrieves a paginated list of news for a league (published only)
func (r *NewsRepository) ListByLeague(ctx context.Context, leagueID uuid.UUID, page, limit int, publishedOnly bool) ([]*model.News, int, error) {
	offset := (page - 1) * limit

	// Count total
	countQuery := `SELECT COUNT(*) FROM news WHERE league_id = $1`
	if publishedOnly {
		countQuery += ` AND is_published = true`
	}

	var total int
	if err := r.db.Pool.QueryRowContext(ctx, countQuery, leagueID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get news
	query := `
		SELECT n.id, n.league_id, n.author_id, n.title, n.content, n.is_published, n.published_at, n.created_at, n.updated_at, u.nickname
		FROM news n
		JOIN users u ON n.author_id = u.id
		WHERE n.league_id = $1
	`
	if publishedOnly {
		query += ` AND n.is_published = true`
	}
	query += `
		ORDER BY n.published_at DESC NULLS LAST, n.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, leagueID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var newsList []*model.News
	for rows.Next() {
		news := &model.News{}
		if err := rows.Scan(
			&news.ID,
			&news.LeagueID,
			&news.AuthorID,
			&news.Title,
			&news.Content,
			&news.IsPublished,
			&news.PublishedAt,
			&news.CreatedAt,
			&news.UpdatedAt,
			&news.AuthorNickname,
		); err != nil {
			return nil, 0, err
		}
		newsList = append(newsList, news)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return newsList, total, nil
}

// Update updates a news article
func (r *NewsRepository) Update(ctx context.Context, news *model.News) error {
	query := `
		UPDATE news
		SET title = $1, content = $2, updated_at = NOW()
		WHERE id = $3
	`

	result, err := r.db.Pool.ExecContext(ctx, query,
		news.Title,
		news.Content,
		news.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNewsNotFound
	}

	return nil
}

// Publish publishes or unpublishes a news article
func (r *NewsRepository) Publish(ctx context.Context, id uuid.UUID, publish bool) error {
	var query string
	if publish {
		query = `
			UPDATE news
			SET is_published = true, published_at = NOW(), updated_at = NOW()
			WHERE id = $1
		`
	} else {
		query = `
			UPDATE news
			SET is_published = false, published_at = NULL, updated_at = NOW()
			WHERE id = $1
		`
	}

	result, err := r.db.Pool.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNewsNotFound
	}

	return nil
}

// Delete deletes a news article
func (r *NewsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM news WHERE id = $1`

	result, err := r.db.Pool.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNewsNotFound
	}

	return nil
}
