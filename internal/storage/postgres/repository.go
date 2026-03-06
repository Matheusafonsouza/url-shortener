package postgres

import (
	"context"
	"errors"

	"golang-urlshortener/internal/shorturl"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, mapping shorturl.URLMapping) error {
	const query = `
		INSERT INTO short_urls (code, original_url, created_at, expires_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(ctx, query, mapping.Code, mapping.OriginalURL, mapping.CreatedAt, mapping.ExpiresAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return shorturl.ErrCodeConflict
		}

		return err
	}

	return nil
}

func (r *Repository) GetByCode(ctx context.Context, code string) (shorturl.URLMapping, error) {
	const query = `
		SELECT code, original_url, created_at, expires_at, click_count, last_accessed_at
		FROM short_urls
		WHERE code = $1
	`

	var mapping shorturl.URLMapping
	if err := r.pool.QueryRow(ctx, query, code).Scan(
		&mapping.Code,
		&mapping.OriginalURL,
		&mapping.CreatedAt,
		&mapping.ExpiresAt,
		&mapping.ClickCount,
		&mapping.LastAccessedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return shorturl.URLMapping{}, shorturl.ErrNotFound
		}

		return shorturl.URLMapping{}, err
	}

	return mapping, nil
}

func (r *Repository) IncrementClickCount(ctx context.Context, code string) error {
	const query = `
		UPDATE short_urls
		SET click_count = click_count + 1,
			last_accessed_at = NOW()
		WHERE code = $1
	`

	result, err := r.pool.Exec(ctx, query, code)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return shorturl.ErrNotFound
	}

	return nil
}
