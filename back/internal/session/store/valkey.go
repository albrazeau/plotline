package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"main/internal/models"
	"time"

	"github.com/valkey-io/valkey-go"
)

var _ Store = &ValkeyStore{}

const defaultSessionTTL = 30 * time.Minute

type ValkeyStore struct {
	client valkey.Client
	logger *slog.Logger
}

// NewValkeyStore initializes a store backed by valkey
func NewValkeyStore(ctx context.Context, logger *slog.Logger, option valkey.ClientOption) (*ValkeyStore, error) {
	client, err := valkey.NewClient(option)
	if err != nil {
		return nil, err
	}

	cmd := client.B().Ping().Build()
	err = client.Do(ctx, cmd).Error()
	if err != nil {
		return nil, err
	}

	return &ValkeyStore{
		client: client,
		logger: logger,
	}, nil
}

func (vk *ValkeyStore) Close() {
	vk.client.Close()
}

func setJSON[T any](ctx context.Context, client valkey.Client, key string, value T, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	cmd := client.B().Set().Key(key).Value(string(data)).Ex(ttl).Build()
	return client.Do(ctx, cmd).Error()
}

func (vk *ValkeyStore) Save(ctx context.Context, sess *models.Session) error {
	key := fmt.Sprintf("session:%s", sess.ID.String())
	err := setJSON(ctx, vk.client, key, sess, defaultSessionTTL)
	if err != nil {
		vk.logger.Error("failed to save session to valkey", slog.String("error", err.Error()))
		return err
	}
	return nil
}
