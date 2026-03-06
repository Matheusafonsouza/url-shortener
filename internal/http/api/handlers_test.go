package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang-urlshortener/internal/shorturl"

	"github.com/gin-gonic/gin"
)

type mockService struct {
	createFn func(ctx context.Context, input shorturl.CreateInput) (shorturl.URLMapping, error)
	statsFn  func(ctx context.Context, code string) (shorturl.URLMapping, error)
}

func (m *mockService) CreateShortURL(ctx context.Context, input shorturl.CreateInput) (shorturl.URLMapping, error) {
	return m.createFn(ctx, input)
}

func (m *mockService) GetStats(ctx context.Context, code string) (shorturl.URLMapping, error) {
	return m.statsFn(ctx, code)
}

func TestCreateShortURL_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expiresAt := time.Date(2026, 4, 5, 10, 0, 0, 0, time.UTC)
	handlers := NewHandlers(&mockService{
		createFn: func(_ context.Context, _ shorturl.CreateInput) (shorturl.URLMapping, error) {
			return shorturl.URLMapping{
				Code:        "abc12345",
				OriginalURL: "https://example.com",
				ExpiresAt:   &expiresAt,
			}, nil
		},
		statsFn: func(_ context.Context, _ string) (shorturl.URLMapping, error) {
			return shorturl.URLMapping{}, nil
		},
	}, "http://localhost:8080")

	router := gin.New()
	router.POST("/api/v1/shorten", handlers.CreateShortURL)

	body, _ := json.Marshal(map[string]any{"url": "https://example.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.Code)
	}
}

func TestCreateShortURL_Conflict(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handlers := NewHandlers(&mockService{
		createFn: func(_ context.Context, _ shorturl.CreateInput) (shorturl.URLMapping, error) {
			return shorturl.URLMapping{}, shorturl.ErrCodeConflict
		},
		statsFn: func(_ context.Context, _ string) (shorturl.URLMapping, error) {
			return shorturl.URLMapping{}, nil
		},
	}, "http://localhost:8080")

	router := gin.New()
	router.POST("/api/v1/shorten", handlers.CreateShortURL)

	body, _ := json.Marshal(map[string]any{"url": "https://example.com", "alias": "myalias"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", resp.Code)
	}
}
