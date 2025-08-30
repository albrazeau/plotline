package store

import (
	"context"
	"main/internal/models"
)

type Store interface {
	Save(context.Context, *models.Session) error
	Close()
}
