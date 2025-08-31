package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"main/internal/models"
	"time"

	"github.com/google/uuid"
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
		logger: logger.With(slog.String("component", "valkey_store")),
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

func (vk *ValkeyStore) Get(ctx context.Context, sessID uuid.UUID) (*models.Session, error) {
	key := fmt.Sprintf("session:%s", sessID.String())

	getCmd := vk.client.B().Get().Key(key).Build()
	expireCmd := vk.client.B().Expire().Key(key).Seconds(int64(defaultSessionTTL.Seconds())).Build()

	results := vk.client.DoMulti(ctx, getCmd, expireCmd)
	if results == nil {
		vk.logger.Error("pipeline failed: no results")
		return nil, errors.New("pipeline failed: no results returned")
	}
	if len(results) != 2 {
		vk.logger.Error("pipeline returned unexpected result count", slog.Int("count", len(results)))
		return nil, fmt.Errorf("pipeline unexpected result count: %d", len(results))
	}

	getResp := results[0]
	if err := getResp.Error(); err != nil {
		if errors.Is(err, valkey.Nil) {
			vk.logger.Warn("session not found", slog.String("key", key))
			return nil, ErrSessionNotFound
		}
		vk.logger.Error("failed to fetch session", slog.String("error", err.Error()))
		return nil, err
	}

	var sess models.Session
	if err := getResp.DecodeJSON(&sess); err != nil {
		vk.logger.Error("failed to decode session JSON", slog.String("error", err.Error()))
		return nil, err
	}

	expResp := results[1]
	return &sess, expResp.Error()
}

func (vk *ValkeyStore) Refresh(ctx context.Context, sessID uuid.UUID) error {
	key := fmt.Sprintf("session:%s", sessID.String())
	cmd := vk.client.B().Expire().Key(key).Seconds(int64(defaultSessionTTL.Seconds())).Build()
	if err := vk.client.Do(ctx, cmd).Error(); err != nil {
		if errors.Is(err, valkey.Nil) {
			vk.logger.Warn("session not found", slog.String("key", key))
			return ErrSessionNotFound
		}
		return err
	}
	return nil
}
