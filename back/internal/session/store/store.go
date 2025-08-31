package store

import (
	"context"
	"errors"
	"main/internal/models"

	"github.com/google/uuid"
)

var ErrSessionNotFound = errors.New("session not found")

type Store interface {
	Save(context.Context, *models.Session) error
	Get(ctx context.Context, id uuid.UUID) (*models.Session, error)
	Refresh(ctx context.Context, id uuid.UUID) error
	Close()
}
