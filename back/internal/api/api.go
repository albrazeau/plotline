package api

import (
	"errors"
	"log/slog"
	"main/internal/llm"
	"main/internal/session"
	"main/internal/session/store"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type API struct {
	logger  *slog.Logger
	session *session.Manager
	llm     llm.LLM
}

func New(logger *slog.Logger, session *session.Manager, llm llm.LLM) *API {
	return &API{
		logger:  logger.With(slog.String("component", "api")),
		session: session,
		llm:     llm,
	}
}

func (a *API) RegisterRoutes(router *gin.Engine) {
	// middlewares
	router.Use(a.recovery, a.requestLogger)

	// routes
	router.GET("/healthcheck", a.healthCheck)

	// v1 api
	v1Routes := router.Group("/api/v1")
	v1Routes.GET("/models", a.models)
	v1Routes.POST("/session", a.startSession)
	v1Routes.GET("/session/:id", a.getSession)
	v1Routes.PATCH("/session/:id", a.refreshSession)
}

func (a *API) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "healthy"})
}

func (a *API) models(c *gin.Context) {
	models, err := a.llm.Models(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to list models", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models)
}

func (a *API) startSession(c *gin.Context) {

	var body StartSessionRequest
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request body", "error": err.Error()})
		return
	}

	models, err := a.llm.Models(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to list models", "error": err.Error()})
		return
	}

	if !slices.Contains(models, body.Model) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request body", "error": "unknown model", "available_models": models})
		return
	}

	sess, err := a.session.Create(c.Request.Context(), body.Model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to start a session", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sess)
}

func (a *API) getSession(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request", "error": "missing id path parameter"})
		return
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request; invalid session id", "error": err.Error()})
		return
	}

	sess, err := a.session.Get(c.Request.Context(), uid)
	if err != nil {
		if errors.Is(err, store.ErrSessionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to get session", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sess)
}

func (a *API) refreshSession(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request", "error": "missing id path parameter"})
		return
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request; invalid session id", "error": err.Error()})
		return
	}

	if err := a.session.Refresh(c.Request.Context(), uid); err != nil {
		if errors.Is(err, store.ErrSessionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to refresh session", "error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
