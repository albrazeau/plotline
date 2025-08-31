package session

import (
	"context"
	"fmt"
	"log/slog"
	"main/internal/models"
	"main/internal/session/store"
	"time"

	"github.com/google/uuid"
)

type Manager struct {
	s      store.Store
	logger *slog.Logger
}

func New(logger *slog.Logger, s store.Store) *Manager {
	return &Manager{s: s, logger: logger}
}

func (mgr *Manager) Create(ctx context.Context, model string) (*models.Session, error) {
	id := uuid.New()

	sess := &models.Session{
		ID:        id,
		CreatedAt: time.Now().UTC(),
		Model:     model,
	}

	if err := mgr.s.Save(ctx, sess); err != nil {
		return nil, fmt.Errorf("unable to save session: %w", err)
	}

	return sess, nil
}

func (mgr *Manager) Save(ctx context.Context, sess *models.Session) error {
	return mgr.s.Save(ctx, sess)
}

func (mgr *Manager) Get(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	return mgr.s.Get(ctx, id)
}
