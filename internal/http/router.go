package http

import (
	"html/template"
	"net/http"

	"golang-urlshortener/internal/http/api"
	"golang-urlshortener/internal/http/middleware"
	"golang-urlshortener/internal/http/web"

	"github.com/gin-gonic/gin"
)

func NewRouter(
	apiHandlers *api.Handlers,
	webHandlers *web.Handlers,
	redirectHandler gin.HandlerFunc,
	templates *template.Template,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())
	router.SetHTMLTemplate(templates)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/", webHandlers.Index)
	router.POST("/shorten", webHandlers.CreateShortURL)

	apiV1 := router.Group("/api/v1")
	{
		apiV1.POST("/shorten", apiHandlers.CreateShortURL)
		apiV1.GET("/urls/:code/stats", apiHandlers.GetStats)
	}

	router.GET("/:code", redirectHandler)

	return router
}
