package shorturl

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	createFn          func(ctx context.Context, mapping URLMapping) error
	getByCodeFn       func(ctx context.Context, code string) (URLMapping, error)
	incrementClicksFn func(ctx context.Context, code string) error
}

func (m *mockRepository) Create(ctx context.Context, mapping URLMapping) error {
	if m.createFn == nil {
		return nil
	}

	return m.createFn(ctx, mapping)
}

func (m *mockRepository) GetByCode(ctx context.Context, code string) (URLMapping, error) {
	if m.getByCodeFn == nil {
		return URLMapping{}, ErrNotFound
	}

	return m.getByCodeFn(ctx, code)
}

func (m *mockRepository) IncrementClickCount(ctx context.Context, code string) error {
	if m.incrementClicksFn == nil {
		return nil
	}

	return m.incrementClicksFn(ctx, code)
}

func TestCreateShortURL_DefaultTTL(t *testing.T) {
	var saved URLMapping
	repo := &mockRepository{
		createFn: func(_ context.Context, mapping URLMapping) error {
			saved = mapping
			return nil
		},
	}

	service := NewService(repo, 720*time.Hour)
	now := time.Date(2026, 3, 6, 10, 0, 0, 0, time.UTC)
	service.nowFunc = func() time.Time { return now }

	_, err := service.CreateShortURL(context.Background(), CreateInput{OriginalURL: "https://example.com"})
	if err != nil {
		t.Fatalf("CreateShortURL returned error: %v", err)
	}

	if saved.ExpiresAt == nil {
		t.Fatal("expected expires_at to be set")
	}

	if !saved.ExpiresAt.Equal(now.Add(720 * time.Hour)) {
		t.Fatalf("unexpected expires_at: got %v", saved.ExpiresAt)
	}
}

func TestCreateShortURL_AliasConflict(t *testing.T) {
	repo := &mockRepository{
		createFn: func(_ context.Context, _ URLMapping) error {
			return ErrCodeConflict
		},
	}

	service := NewService(repo, 720*time.Hour)
	_, err := service.CreateShortURL(context.Background(), CreateInput{
		OriginalURL: "https://example.com",
		Alias:       "myalias",
	})

	if !errors.Is(err, ErrCodeConflict) {
		t.Fatalf("expected ErrCodeConflict, got %v", err)
	}
}

func TestResolveShortCode_Expired(t *testing.T) {
	now := time.Date(2026, 3, 6, 10, 0, 0, 0, time.UTC)
	expiredAt := now.Add(-1 * time.Hour)

	repo := &mockRepository{
		getByCodeFn: func(_ context.Context, _ string) (URLMapping, error) {
			return URLMapping{
				Code:        "abc12345",
				OriginalURL: "https://example.com",
				ExpiresAt:   &expiredAt,
			}, nil
		},
	}

	service := NewService(repo, 720*time.Hour)
	service.nowFunc = func() time.Time { return now }

	_, err := service.ResolveShortCode(context.Background(), "abc12345")
	if !errors.Is(err, ErrExpired) {
		t.Fatalf("expected ErrExpired, got %v", err)
	}
}
