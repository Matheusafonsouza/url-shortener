package web

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang-urlshortener/internal/shorturl"

	"github.com/gin-gonic/gin"
)

type service interface {
	CreateShortURL(ctx context.Context, input shorturl.CreateInput) (shorturl.URLMapping, error)
}

type Handlers struct {
	service service
	baseURL string
}

func NewHandlers(service service, baseURL string) *Handlers {
	return &Handlers{
		service: service,
		baseURL: strings.TrimSuffix(baseURL, "/"),
	}
}

func (h *Handlers) Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"error": "",
	})
}

func (h *Handlers) CreateShortURL(c *gin.Context) {
	originalURL := c.PostForm("url")
	alias := c.PostForm("alias")
	var ttlHours *int

	rawTTL := strings.TrimSpace(c.PostForm("ttl_hours"))
	if rawTTL != "" {
		parsed, err := strconv.Atoi(rawTTL)
		if err != nil {
			c.HTML(http.StatusBadRequest, "index.html", gin.H{"error": "ttl_hours must be a number"})
			return
		}

		ttlHours = &parsed
	}

	mapping, err := h.service.CreateShortURL(c.Request.Context(), shorturl.CreateInput{
		OriginalURL: originalURL,
		Alias:       alias,
		TTLHours:    ttlHours,
	})
	if err != nil {
		status := http.StatusBadRequest
		if !errors.Is(err, shorturl.ErrInvalidURL) &&
			!errors.Is(err, shorturl.ErrInvalidAlias) &&
			!errors.Is(err, shorturl.ErrReservedAlias) &&
			!errors.Is(err, shorturl.ErrInvalidTTL) &&
			!errors.Is(err, shorturl.ErrCodeConflict) {
			status = http.StatusInternalServerError
		}

		c.HTML(status, "index.html", gin.H{"error": err.Error()})
		return
	}

	var expiresAt string
	if mapping.ExpiresAt != nil {
		expiresAt = mapping.ExpiresAt.UTC().Format(time.RFC3339)
	}

	c.HTML(http.StatusCreated, "result.html", gin.H{
		"code":         mapping.Code,
		"original_url": mapping.OriginalURL,
		"short_url":    h.baseURL + "/" + mapping.Code,
		"expires_at":   expiresAt,
	})
}
