package api

import (
	"log/slog"
	"main/internal/llm"
	"main/internal/session"
	"net/http"

	"github.com/gin-gonic/gin"
)

type API struct {
	logger  *slog.Logger
	session *session.Session
	llm     llm.LLM
}

func New(logger *slog.Logger, session *session.Session, llm llm.LLM) *API {
	return &API{
		logger:  logger.With(slog.String("component", "api")),
		session: session,
		llm:     llm,
	}
}

func (a *API) RegisterRoutes(router gin.IRoutes) {
	// middlewares
	router.Use(a.recovery, a.requestLogger)

	// routes
	router.GET("/healthcheck", a.healthCheck)
}

func (a *API) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "healthy"})
}
