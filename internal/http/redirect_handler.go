package http

import (
	"context"
	"errors"
	"net/http"

	"golang-urlshortener/internal/shorturl"

	"github.com/gin-gonic/gin"
)

type redirectService interface {
	ResolveShortCode(ctx context.Context, code string) (shorturl.URLMapping, error)
}

type RedirectHandler struct {
	service redirectService
}

func NewRedirectHandler(service redirectService) *RedirectHandler {
	return &RedirectHandler{service: service}
}

func (h *RedirectHandler) Redirect(c *gin.Context) {
	code := c.Param("code")
	mapping, err := h.service.ResolveShortCode(c.Request.Context(), code)
	if err != nil {
		switch {
		case errors.Is(err, shorturl.ErrNotFound):
			c.String(http.StatusNotFound, "short URL not found")
		case errors.Is(err, shorturl.ErrExpired):
			c.String(http.StatusGone, "short URL expired")
		default:
			c.String(http.StatusInternalServerError, "internal server error")
		}

		return
	}

	c.Redirect(http.StatusFound, mapping.OriginalURL)
}
