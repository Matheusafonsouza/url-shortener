package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"golang-urlshortener/internal/shorturl"

	"github.com/gin-gonic/gin"
)

type service interface {
	CreateShortURL(ctx context.Context, input shorturl.CreateInput) (shorturl.URLMapping, error)
	GetStats(ctx context.Context, code string) (shorturl.URLMapping, error)
}

type Handlers struct {
	service service
	baseURL string
}

type createShortURLRequest struct {
	URL      string `json:"url" binding:"required"`
	Alias    string `json:"alias"`
	TTLHours *int   `json:"ttl_hours"`
}

type createShortURLResponse struct {
	Code        string  `json:"code"`
	ShortURL    string  `json:"short_url"`
	OriginalURL string  `json:"original_url"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
}

type statsResponse struct {
	Code           string  `json:"code"`
	OriginalURL    string  `json:"original_url"`
	ClickCount     int64   `json:"click_count"`
	CreatedAt      string  `json:"created_at"`
	ExpiresAt      *string `json:"expires_at,omitempty"`
	LastAccessedAt *string `json:"last_accessed_at,omitempty"`
}

func NewHandlers(service service, baseURL string) *Handlers {
	return &Handlers{
		service: service,
		baseURL: strings.TrimSuffix(baseURL, "/"),
	}
}

func (h *Handlers) CreateShortURL(c *gin.Context) {
	var req createShortURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	mapping, err := h.service.CreateShortURL(c.Request.Context(), shorturl.CreateInput{
		OriginalURL: req.URL,
		Alias:       req.Alias,
		TTLHours:    req.TTLHours,
	})
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	response := createShortURLResponse{
		Code:        mapping.Code,
		ShortURL:    h.baseURL + "/" + mapping.Code,
		OriginalURL: mapping.OriginalURL,
	}
	if mapping.ExpiresAt != nil {
		formatted := mapping.ExpiresAt.UTC().Format(time.RFC3339)
		response.ExpiresAt = &formatted
	}

	c.JSON(http.StatusCreated, response)
}

func (h *Handlers) GetStats(c *gin.Context) {
	code := c.Param("code")
	mapping, err := h.service.GetStats(c.Request.Context(), code)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	response := statsResponse{
		Code:        mapping.Code,
		OriginalURL: mapping.OriginalURL,
		ClickCount:  mapping.ClickCount,
		CreatedAt:   mapping.CreatedAt.UTC().Format(time.RFC3339),
	}

	if mapping.ExpiresAt != nil {
		formatted := mapping.ExpiresAt.UTC().Format(time.RFC3339)
		response.ExpiresAt = &formatted
	}

	if mapping.LastAccessedAt != nil {
		formatted := mapping.LastAccessedAt.UTC().Format(time.RFC3339)
		response.LastAccessedAt = &formatted
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handlers) respondWithError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, shorturl.ErrInvalidURL),
		errors.Is(err, shorturl.ErrInvalidAlias),
		errors.Is(err, shorturl.ErrReservedAlias),
		errors.Is(err, shorturl.ErrInvalidTTL):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, shorturl.ErrCodeConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, shorturl.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, shorturl.ErrExpired):
		c.JSON(http.StatusGone, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
