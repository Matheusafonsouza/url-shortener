package shorturl

import "context"

type Repository interface {
	Create(ctx context.Context, mapping URLMapping) error
	GetByCode(ctx context.Context, code string) (URLMapping, error)
	IncrementClickCount(ctx context.Context, code string) error
}
