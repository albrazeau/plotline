package store

import (
	"context"

	"github.com/valkey-io/valkey-go"
)

var _ Store = &ValkeyStore{}

type ValkeyStore struct {
	client valkey.Client
}

// NewValkeyStore initializes a store backed by valkey
func NewValkeyStore(ctx context.Context, option valkey.ClientOption) (*ValkeyStore, error) {
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
	}, nil
}

func (vk *ValkeyStore) Close() {
	vk.client.Close()
}
